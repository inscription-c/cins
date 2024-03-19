package handle

import (
	"errors"
	"github.com/btcsuite/btcd/txscript"
	"github.com/gin-gonic/gin"
	"github.com/inscription-c/cins/inscription/index/model"
	"github.com/inscription-c/cins/pkg/util"
	"net/http"
	"strings"
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
		errStr := strings.ToLower(err.Error())
		errExp := strings.ToLower("-5: No such mempool")
		if strings.Contains(errStr, errExp) {
			ctx.Status(http.StatusNotFound)
			return nil
		}
		return err
	}
	txOut := tx.MsgTx().TxOut[output.Index]

	var addressStr string
	pkScript, err := txscript.ParsePkScript(txOut.PkScript)
	if err != nil && !errors.Is(err, txscript.ErrUnsupportedScriptType) {
		return err
	}
	if err == nil {
		address, err := pkScript.Address(util.ActiveNet.Params)
		if err != nil {
			return err
		}
		addressStr = address.String()
	}

	scriptPk, err := txscript.DisasmString(txOut.PkScript)
	if err != nil {
		return err
	}

	ctx.JSON(http.StatusOK, gin.H{
		"value":         txOut.Value,
		"script_pubkey": scriptPk,
		"address":       addressStr,
		"transaction":   output.Hash.String(),
		"inscriptions":  inscriptions,
	})
	return nil
}
