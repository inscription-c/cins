package handle

import (
	"github.com/gin-gonic/gin"
	"github.com/gogf/gf/v2/util/gconv"
	"github.com/inscription-c/insc/internal/util"
	"net/http"
)

// Inscriptions is a handler function for handling inscriptions requests.
// It validates the request parameters and calls the doInscriptions function.
func (h *Handler) Inscriptions(ctx *gin.Context) {
	page := ctx.DefaultQuery("page", "1")
	if err := h.doInscriptions(ctx, page); err != nil {
		ctx.Status(http.StatusInternalServerError)
		return
	}
}

// doInscriptions is a helper function for handling inscriptions requests.
// It retrieves the inscriptions based on the provided page number and returns them in the response.
func (h *Handler) doInscriptions(ctx *gin.Context, pageStr string) error {
	// Convert the page number to an integer.
	page := gconv.Int(pageStr)
	// If the page number is less than or equal to 0, return a bad request status.
	if page <= 0 {
		ctx.Status(http.StatusBadRequest)
		return nil
	}
	// Set the page size.
	pageSize := 100

	// Retrieve the inscriptions for the specified page.
	inscriptions, err := h.DB().FindInscriptionsByPage(page, pageSize)
	if err != nil {
		return err
	}

	// Create a slice to hold the inscription IDs.
	inscriptionIds := make([]string, 0, len(inscriptions))
	// Iterate over the inscriptions and add their IDs to the slice.
	for _, v := range inscriptions {
		inscriptionIds = append(inscriptionIds, util.NewInscriptionId(v.Outpoint, v.Offset))
	}
	// Return the inscription IDs, page index, and a flag indicating if there are more inscriptions.
	ctx.JSON(http.StatusOK, gin.H{
		"page_index":   page,
		"more":         len(inscriptionIds) > pageSize,
		"inscriptions": inscriptionIds,
	})
	return nil
}
