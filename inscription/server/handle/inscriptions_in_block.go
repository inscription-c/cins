package handle

import (
	"github.com/gin-gonic/gin"
	"github.com/gogf/gf/v2/util/gconv"
	"github.com/inscription-c/insc/internal/util"
	"net/http"
)

// InscriptionsInBlock is a handler function for handling inscriptions in block requests.
// It validates the request parameters and calls the doInscriptionsInBlock function.
func (h *Handler) InscriptionsInBlock(ctx *gin.Context) {
	height := ctx.DefaultQuery("height", "0")
	page := ctx.DefaultQuery("page", "1")
	if err := h.doInscriptionsInBlock(ctx, gconv.Int(height), gconv.Int(page)); err != nil {
		ctx.Status(http.StatusInternalServerError)
		return
	}
}

// doInscriptionsInBlock is a helper function for handling inscriptions in block requests.
// It retrieves the inscriptions in a specific block based on the provided height and page number and returns them in the response.
func (h *Handler) doInscriptionsInBlock(ctx *gin.Context, height, page int) error {
	// If the page number is less than or equal to 0, return a bad request status.
	if page <= 0 {
		ctx.Status(http.StatusBadRequest)
		return nil
	}

	// Set the page size.
	size := 100
	// Retrieve the inscriptions for the specified block and page.
	list, err := h.DB().FindInscriptionsInBlock(height, page, size)
	if err != nil {
		return err
	}
	// Check if there are more inscriptions than the page size.
	more := len(list) > size

	// Create a slice to hold the inscription IDs.
	inscriptionsIds := make([]string, 0, len(list))
	// Iterate over the inscriptions and add their IDs to the slice.
	for _, v := range list {
		inscriptionsIds = append(inscriptionsIds, util.NewInscriptionId(v.Outpoint, v.Offset))
	}

	// Return the block height, page index, a flag indicating if there are more inscriptions, and the inscription IDs.
	ctx.JSON(http.StatusOK, gin.H{
		"block_height": height,
		"page_index":   page,
		"more":         more,
		"inscriptions": inscriptionsIds,
	})
	return nil
}
