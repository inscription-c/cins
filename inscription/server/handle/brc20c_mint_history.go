package handle

import (
	"github.com/gin-gonic/gin"
	"github.com/gogf/gf/v2/util/gconv"
	"github.com/inscription-c/insc/constants"
	"github.com/inscription-c/insc/inscription/index/dao"
	"github.com/inscription-c/insc/internal/util"
	"net/http"
)

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

func (h *Handler) doBRC20CMintHistory(ctx *gin.Context, tkidOrAddr string, page int) error {
	pageSize := 100
	var err error
	var res []*dao.ProtocolAmount
	tkid := util.StringToInscriptionId(tkidOrAddr)
	if tkid == nil {
		address := tkidOrAddr
		res, err = h.DB().SumAmountByToAddress(constants.ProtocolBRC20C, address, page, pageSize)
		if err != nil {
			return err
		}
		if len(res) > pageSize {
			res = res[:pageSize]
		}
		ctx.JSON(http.StatusOK, gin.H{
			"page_index": page,
			"more":       len(res) > pageSize,
			"amount":     res,
		})
		return nil
	}

	list, err := h.DB().FindMintHistoryByTkId(tkid.OutPoint.String(), constants.ProtocolBRC20C, constants.OperationMint, page, pageSize)
	if err != nil {
		return err
	}
	more := false
	if len(list) > pageSize {
		more = true
		list = list[:pageSize]
	}
	deploy, err := h.DB().GetInscriptionById(tkid.OutPoint.String())
	if err != nil {
		return err
	}

	ctx.JSON(http.StatusOK, gin.H{
		"chain":        deploy.DstChain,
		"ticker_id":    deploy.Outpoint.InscriptionId().String(),
		"page_index":   page,
		"more":         more,
		"mint_history": list,
	})
	return nil
}
