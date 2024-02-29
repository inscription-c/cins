package handle

import (
	"github.com/btcsuite/btcd/txscript"
	"github.com/gin-gonic/gin"
	"github.com/inscription-c/cins/inscription/index/model"
	"github.com/inscription-c/cins/internal/util"
	"net/http"
)

func (h *Handler) InscriptionsInOutput(ctx *gin.Context) {
	output := ctx.Param("output")
	if output == "" {
		ctx.String(http.StatusBadRequest, "invalid output")
		return
	}
	if err := h.doInscriptionsInOutput(ctx, output); err != nil {
		ctx.String(http.StatusInternalServerError, err.Error())
		return
	}
}

func (h *Handler) doInscriptionsInOutput(ctx *gin.Context, outputStr string) error {
	output := model.StringToOutpoint(outputStr)
	inscriptions, err := h.DB().GetInscriptionByOutpoint(output)
	if err != nil {
		return err
	}
	tx, err := h.RpcClient().GetRawTransaction(&output.Hash)
	if err != nil {
		return err
	}
	txOut := tx.MsgTx().TxOut[output.Index]
	_, addresses, num, err := txscript.ExtractPkScriptAddrs(txOut.PkScript, util.ActiveNet.Params)
	if err != nil {
		return err
	}
	address := ""
	if num > 0 {
		address = addresses[0].String()
	}

	scriptPk, err := txscript.DisasmString(txOut.PkScript)
	if err != nil {
		return err
	}

	ctx.JSON(http.StatusOK, gin.H{
		"value":         txOut.Value,
		"script_pubkey": scriptPk,
		"address":       address,
		"transaction":   output.Hash.String(),
		"inscriptions":  inscriptions,
	})
	return nil
}
