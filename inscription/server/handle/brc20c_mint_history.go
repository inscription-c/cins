package handle

import (
	"github.com/gin-gonic/gin"
	"github.com/gogf/gf/v2/util/gconv"
	"github.com/inscription-c/insc/constants"
	"github.com/inscription-c/insc/inscription/index/tables"
	"net/http"
)

// BRC20CMintHistory is a handler function for handling BRC20C mint history requests.
// It validates the request parameters and calls the doBRC20CMintHistory function.
func (h *Handler) BRC20CMintHistory(ctx *gin.Context) {
	tkidOrAddr := ctx.Query("tkidOrAddr")
	page := ctx.DefaultQuery("page", "1")
	if tkidOrAddr == "" {
		ctx.Status(http.StatusBadRequest)
		return
	}
	if gconv.Int(page) < 1 {
		ctx.Status(http.StatusBadRequest)
		return
	}
	if err := h.doBRC20CMintHistory(ctx, tkidOrAddr, gconv.Int(page)); err != nil {
		ctx.Status(http.StatusInternalServerError)
		return
	}
}

// doBRC20CMintHistory is a helper function for handling BRC20C mint history requests.
// It retrieves the mint history of a specific BRC20C token or address and returns them in the response.
func (h *Handler) doBRC20CMintHistory(ctx *gin.Context, tkidOrAddr string, page int) error {
	pageSize := 100
	var err error
	var res []*tables.ProtocolAmount
	tkid := tables.StringToInscriptionId(tkidOrAddr)
	if tkid == nil {
		address := tkidOrAddr
		res, err = h.DB().SumMintAmountByAddress(address, constants.ProtocolBRC20C, page, pageSize)
		if err != nil {
			return err
		}
		more := false
		if len(res) > pageSize {
			more = true
			res = res[:pageSize]
		}
		ctx.JSON(http.StatusOK, gin.H{
			"page_index": page,
			"more":       more,
			"amount":     res,
		})
		return nil
	}

	list, err := h.DB().FindMintHistoryByTkId(tkid.String(), constants.ProtocolBRC20C, constants.OperationMint, page, pageSize)
	if err != nil {
		return err
	}
	more := false
	if len(list) > pageSize {
		more = true
		list = list[:pageSize]
	}
	deploy, err := h.DB().GetInscriptionById(tkid)
	if err != nil {
		return err
	}

	ctx.JSON(http.StatusOK, gin.H{
		"chain":        deploy.DstChain,
		"ticker_id":    deploy.InscriptionId,
		"page_index":   page,
		"more":         more,
		"mint_history": list,
	})
	return nil
}