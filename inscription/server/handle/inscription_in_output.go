package handle

import (
	"github.com/btcsuite/btcd/txscript"
	"github.com/gin-gonic/gin"
	"github.com/inscription-c/insc/inscription/index/model"
	"github.com/inscription-c/insc/internal/util"
	"net/http"
)

func (h *Handler) InscriptionsInOutput(ctx *gin.Context) {
	output := ctx.Query("output")
	if output == "" {
		ctx.Status(http.StatusBadRequest)
		return
	}
	if err := h.doInscriptionsInOutput(ctx, output); err != nil {
		ctx.Status(http.StatusInternalServerError)
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
	_, address, _, err := txscript.ExtractPkScriptAddrs(txOut.PkScript, util.ActiveNet.Params)
	ctx.JSON(http.StatusOK, gin.H{
		"value":        txOut.Value,
		"script":       string(txOut.PkScript),
		"addresses":    address[0],
		"transaction":  output.Hash.String(),
		"inscriptions": inscriptions,
	})
	return nil
}
