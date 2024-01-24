package handle

import (
	"github.com/gin-gonic/gin"
	"github.com/gogf/gf/v2/util/gconv"
	"github.com/inscription-c/insc/internal/util"
	"net/http"
)

func (h *Handler) Inscriptions(ctx *gin.Context) {
	page := ctx.DefaultQuery("page", "1")
	if err := h.doInscriptions(ctx, page); err != nil {
		ctx.Status(http.StatusInternalServerError)
		return
	}
}

func (h *Handler) doInscriptions(ctx *gin.Context, pageStr string) error {
	page := gconv.Int(pageStr)
	if page <= 0 {
		ctx.Status(http.StatusBadRequest)
		return nil
	}
	pageSize := 100

	inscriptions, err := h.DB().FindInscriptionsByPage(page, pageSize)
	if err != nil {
		return err
	}

	inscriptionIds := make([]string, 0, len(inscriptions))
	for _, v := range inscriptions {
		inscriptionIds = append(inscriptionIds, util.NewInscriptionId(v.Outpoint, v.Offset))
	}
	ctx.JSON(http.StatusOK, gin.H{
		"page_index":   page,
		"more":         len(inscriptionIds) > pageSize,
		"inscriptions": inscriptionIds,
	})
	return nil
}
