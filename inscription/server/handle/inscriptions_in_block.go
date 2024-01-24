package handle

import (
	"github.com/gin-gonic/gin"
	"github.com/gogf/gf/v2/util/gconv"
	"github.com/inscription-c/insc/internal/util"
	"net/http"
)

func (h *Handler) InscriptionsInBlock(ctx *gin.Context) {
	height := ctx.DefaultQuery("height", "0")
	page := ctx.DefaultQuery("page", "1")
	if err := h.doInscriptionsInBlock(ctx, gconv.Int(height), gconv.Int(page)); err != nil {
		ctx.Status(http.StatusInternalServerError)
		return
	}
}

func (h *Handler) doInscriptionsInBlock(ctx *gin.Context, height, page int) error {
	if page <= 0 {
		ctx.Status(http.StatusBadRequest)
		return nil
	}

	size := 100
	list, err := h.DB().FindInscriptionsInBlock(height, page, size)
	if err != nil {
		return err
	}
	more := len(list) > size

	inscriptionsIds := make([]string, 0, len(list))
	for _, v := range list {
		inscriptionsIds = append(inscriptionsIds, util.NewInscriptionId(v.Outpoint, v.Offset))
	}

	ctx.JSON(http.StatusOK, gin.H{
		"block_height": height,
		"page_index":   page,
		"more":         more,
		"inscriptions": inscriptionsIds,
	})
	return nil
}
