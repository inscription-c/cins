package handle

import (
	"github.com/gin-gonic/gin"
	"github.com/gogf/gf/v2/util/gconv"
	"net/http"
)

// InscriptionsInBlockPage is a handler function for handling inscriptions in block requests.
// It validates the request parameters and calls the doInscriptionsInBlock function.
func (h *Handler) InscriptionsInBlockPage(ctx *gin.Context) {
	height := ctx.Param("height")
	if height == "" {
		height = "0"
	}
	page := ctx.Param("page")
	if page == "" {
		page = "1"
	}
	if err := h.doInscriptionsInBlockPage(ctx, gconv.Int(height), gconv.Int(page)); err != nil {
		ctx.String(http.StatusInternalServerError, err.Error())
		return
	}
}

// doInscriptionsInBlock is a helper function for handling inscriptions in block requests.
// It retrieves the inscriptions in a specific block based on the provided height and page number and returns them in the response.
func (h *Handler) doInscriptionsInBlockPage(ctx *gin.Context, height, page int) error {
	if page <= 0 {
		ctx.String(http.StatusBadRequest, "invalid page")
		return nil
	}

	size := 100
	// Retrieve the inscriptions for the specified block and page.
	list, err := h.DB().FindInscriptionsInBlockPage(height, page, size)
	if err != nil {
		return err
	}
	more := false
	if len(list) > size {
		more = true
		list = list[:size]
	}

	// Return the block height, page index, a flag indicating if there are more inscriptions, and the inscription IDs.
	ctx.JSON(http.StatusOK, gin.H{
		"block_height": height,
		"page_index":   page,
		"more":         more,
		"inscriptions": list,
	})
	return nil
}
