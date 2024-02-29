package handle

import (
	"github.com/gin-gonic/gin"
	"github.com/gogf/gf/v2/util/gconv"
	"net/http"
)

// Inscriptions is a handler function for handling inscriptions requests.
// It validates the request parameters and calls the doInscriptions function.
func (h *Handler) Inscriptions(ctx *gin.Context) {
	page := ctx.Param("page")
	if page == "" {
		page = "1"
	}
	if err := h.doInscriptions(ctx, page); err != nil {
		ctx.String(http.StatusInternalServerError, err.Error())
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
		ctx.String(http.StatusBadRequest, "invalid page")
		return nil
	}
	// Set the page size.
	pageSize := 100

	// Retrieve the inscriptions for the specified page.
	list, err := h.DB().FindInscriptionsByPage(page, pageSize)
	if err != nil {
		return err
	}
	more := false
	if len(list) > pageSize {
		more = true
		list = list[:pageSize]
	}

	ctx.JSON(http.StatusOK, gin.H{
		"page_index":   page,
		"more":         more,
		"inscriptions": list,
	})
	return nil
}
