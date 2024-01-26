package handle

import (
	"github.com/gin-gonic/gin"
	"github.com/gogf/gf/v2/util/gconv"
	"net/http"
)

// BlockHash return latest block hash, or return block hash by height.
func (h *Handler) BlockHash(ctx *gin.Context) {
	height := ctx.Param("height")
	if err := h.doBlockHash(ctx, height); err != nil {
		ctx.Status(http.StatusInternalServerError)
		return
	}
	ctx.Status(http.StatusOK)
}

func (h *Handler) doBlockHash(ctx *gin.Context, height string) error {
	var err error
	var blockHash string
	if height != "" {
		blockHash, err = h.DB().BlockHash(gconv.Uint32(height))
		if err != nil {
			return err
		}
	} else {
		blockHash, err = h.DB().BlockHash()
		if err != nil {
			return err
		}
	}
	if _, err := ctx.Writer.WriteString(blockHash); err != nil {
		return err
	}
	return nil
}
