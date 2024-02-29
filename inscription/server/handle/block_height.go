package handle

import (
	"github.com/gin-gonic/gin"
	"github.com/gogf/gf/v2/util/gconv"
	"net/http"
)

// BlockHeight return latest block height
func (h *Handler) BlockHeight(ctx *gin.Context) {
	if err := h.doBlockHeight(ctx); err != nil {
		ctx.String(http.StatusInternalServerError, err.Error())
		return
	}
}

func (h *Handler) doBlockHeight(ctx *gin.Context) error {
	height, err := h.DB().BlockHeight()
	if err != nil {
		return err
	}
	ctx.String(http.StatusOK, gconv.String(height))
	return nil
}
