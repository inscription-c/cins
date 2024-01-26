package handle

import (
	"github.com/gin-gonic/gin"
	"github.com/gogf/gf/v2/util/gconv"
	"github.com/inscription-c/insc/constants"
	"github.com/inscription-c/insc/internal/util"
	"net/http"
)

// BRC20CHolders is a handler function for handling BRC20C holders requests.
// It validates the request parameters and calls the doBRC20CHolders function.
func (h *Handler) BRC20CHolders(ctx *gin.Context) {
	tkid := ctx.Query("tkid")
	page := ctx.DefaultQuery("page", "1")
	if tkid == "" {
		ctx.Status(http.StatusBadRequest)
		return
	}
	if gconv.Int(page) < 1 {
		ctx.Status(http.StatusBadRequest)
		return
	}
	if err := h.doBRC20CHolders(ctx, tkid, gconv.Int(page)); err != nil {
		ctx.Status(http.StatusInternalServerError)
		return
	}
}

// doBRC20CHolders is a helper function for handling BRC20C holders requests.
// It retrieves the holders of a specific BRC20C token and returns them in the response.
func (h *Handler) doBRC20CHolders(ctx *gin.Context, tkid string, page int) error {
	pageSize := 100
	inscriptionId := util.StringToInscriptionId(tkid)
	if inscriptionId == nil {
		ctx.Status(http.StatusBadRequest)
		return nil
	}
	protocol, err := h.DB().GetProtocolByOutpoint(inscriptionId.OutPoint.String())
	if err != nil {
		return err
	}
	if protocol.Id == 0 || protocol.Protocol != constants.ProtocolBRC20C {
		ctx.Status(http.StatusNotFound)
		return nil
	}
	if protocol.Operator == constants.OperationMint {
		protocol, err = h.DB().GetProtocolByOutpoint(protocol.TkID.String())
		if err != nil {
			return err
		}
		if protocol.Id == 0 {
			ctx.Status(http.StatusNotFound)
			return nil
		}
	}

	list, err := h.DB().FindHoldersByTkId(protocol.Outpoint.String(), constants.ProtocolBRC20C, constants.OperationMint, page, pageSize)
	if err != nil {
		return err
	}
	more := false
	if len(list) > pageSize {
		more = true
		list = list[:pageSize]
	}

	// Respond with the holders list, page index and a flag indicating if there are more holders
	ctx.JSON(http.StatusOK, gin.H{
		"page_index": page,
		"more":       more,
		"holders":    list,
	})
	return nil
}
