package handle

import (
	"github.com/gin-gonic/gin"
	"github.com/gogf/gf/v2/util/gconv"
	"github.com/inscription-c/insc/constants"
	"github.com/inscription-c/insc/inscription/index/tables"
	"net/http"
)

// BRC20CHolders is a handler function for handling BRC20C holders requests.
// It validates the request parameters and calls the doBRC20CHolders function.
func (h *Handler) BRC20CHolders(ctx *gin.Context) {
	tkid := ctx.Param("tkid")
	page := ctx.Param("page")
	if page == "" {
		page = "1"
	}
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
	inscriptionId := tables.StringToInscriptionId(tkid)
	if inscriptionId == nil {
		ctx.Status(http.StatusBadRequest)
		return nil
	}
	protocol, err := h.DB().GetProtocolByInscriptionId(inscriptionId)
	if err != nil {
		return err
	}
	if protocol.Id == 0 || protocol.Protocol != constants.ProtocolCBRC20 {
		ctx.Status(http.StatusNotFound)
		return nil
	}
	if protocol.Operator == constants.OperationMint {
		//tkid = tables.StringToInscriptionId(protocol.TkID).String()
	}
	list, err := h.DB().FindHoldersByTkId(tkid, constants.ProtocolCBRC20, constants.OperationMint, page, pageSize)
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
