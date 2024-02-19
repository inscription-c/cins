package handle

import (
	"github.com/gin-gonic/gin"
	"net/http"
)

// BlockClock return latest block clock
func (h *Handler) BlockClock(ctx *gin.Context) {
	if err := h.doBlockClock(ctx); err != nil {
		ctx.String(http.StatusInternalServerError, err.Error())
		return
	}
}

func (h *Handler) doBlockClock(ctx *gin.Context) error {
	height, header, err := h.DB().BlockHeader()
	if err != nil {
		return err
	}
	ctx.JSON(http.StatusOK, gin.H{
		"height": height,
		"hour":   header.Timestamp.Hour(),
		"minute": header.Timestamp.Minute(),
		"second": header.Timestamp.Second(),
	})
	return nil
}
