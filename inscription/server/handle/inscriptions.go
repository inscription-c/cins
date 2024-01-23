package handle

import (
	"github.com/gin-gonic/gin"
	"github.com/gogf/gf/v2/util/gconv"
	"github.com/inscription-c/insc/internal/util"
	"net/http"
)

func (h *Handler) Inscriptions(ctx *gin.Context) {
	page := ctx.DefaultQuery("page", "1")
	if page == "" {
		ctx.Status(http.StatusBadRequest)
		return
	}
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

	inscriptionIds, total, err := h.DB().FindInscriptionsByPage(page, pageSize)
	if err != nil {
		return err
	}

	inscriptionIdsResp := make([]string, 0, len(inscriptionIds))
	for _, v := range inscriptionIds {
		inscriptionIdsResp = append(inscriptionIdsResp, util.NewInscriptionId(v.Outpoint, v.Offset))
	}
	ctx.JSON(http.StatusOK, gin.H{
		"page_index":   page,
		"more":         total > int64(page*pageSize),
		"inscriptions": inscriptionIdsResp,
	})
	return nil
}
