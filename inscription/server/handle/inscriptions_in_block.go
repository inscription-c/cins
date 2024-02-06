package handle

import (
	"github.com/gin-gonic/gin"
	"github.com/gogf/gf/v2/util/gconv"
	"github.com/inscription-c/insc/inscription/index/model"
	"golang.org/x/sync/errgroup"
	"net/http"
)

func (h *Handler) InscriptionsInBlock(ctx *gin.Context) {
	height := ctx.Param("height")
	if height == "" {
		height = "0"
	}
	if err := h.doInscriptionsInBlock(ctx, gconv.Uint32(height)); err != nil {
		ctx.Status(http.StatusInternalServerError)
		return
	}
}

func (h *Handler) doInscriptionsInBlock(ctx *gin.Context, height uint32) error {
	var err error
	var latestHeight uint32
	var blockHash string
	var list []*model.OutPoint

	errWg := &errgroup.Group{}
	errWg.Go(func() error {
		blockHash, err = h.DB().BlockHash(height)
		if err != nil {
			return err
		}
		return nil
	})
	errWg.Go(func() error {
		list, err = h.DB().FindInscriptionsInBlock(height)
		if err != nil {
			return err
		}
		return nil
	})
	errWg.Go(func() error {
		latestHeight, err = h.DB().BlockHeight()
		if err != nil {
			return err
		}
		return nil
	})

	ctx.JSON(http.StatusOK, gin.H{
		"hash":         blockHash,
		"target":       height,
		"best_height":  latestHeight,
		"inscriptions": list,
	})
	return nil
}
