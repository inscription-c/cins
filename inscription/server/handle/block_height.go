package handle

import (
	"github.com/gin-gonic/gin"
	"github.com/gogf/gf/v2/util/gconv"
	"net/http"
)

// BlockHeight return latest block height
func (h *Handler) BlockHeight(ctx *gin.Context) {
	if err := h.doBlockHeight(ctx); err != nil {
		ctx.Status(http.StatusInternalServerError)
		return
	}
	ctx.Status(http.StatusOK)
}

func (h *Handler) doBlockHeight(ctx *gin.Context) error {
	height, err := h.DB().BlockHeight()
	if err != nil {
		return err
	}
	if _, err := ctx.Writer.WriteString(gconv.String(height)); err != nil {
		return err
	}
	return nil
}
