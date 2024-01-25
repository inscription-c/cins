package handle

import (
	"github.com/gin-gonic/gin"
	"github.com/gogf/gf/v2/util/gconv"
	"github.com/inscription-c/insc/constants"
	"golang.org/x/sync/errgroup"
	"net/http"
)

func (h *Handler) BRC20CTokens(ctx *gin.Context) {
	tk := ctx.Query("tk")
	page := ctx.DefaultQuery("page", "1")
	if tk == "" {
		ctx.Status(http.StatusBadRequest)
		return
	}
	if gconv.Int(page) < 1 {
		ctx.Status(http.StatusBadRequest)
		return
	}
	if err := h.doBRC20CTokens(ctx, tk, gconv.Int(page)); err != nil {
		ctx.Status(http.StatusInternalServerError)
		return
	}
}

func (h *Handler) doBRC20CTokens(ctx *gin.Context, tk string, page int) error {
	pageSize := 100
	list, err := h.DB().FindTokenPageByTicker(constants.ProtocolBRC20C, tk, constants.OperationDeploy, page, pageSize)
	if err != nil {
		return err
	}
	respSize := len(list)
	if respSize > pageSize {
		respSize = pageSize
	}

	currentNum := 10
	errWg := &errgroup.Group{}
	ch := make(chan int, currentNum)
	respList := make([]gin.H, respSize)

	errWg.Go(func() error {
		for idx := range respList {
			ch <- idx
		}
		close(ch)
		return nil
	})

	for i := 0; i < currentNum; i++ {
		errWg.Go(func() error {
			for idx := range ch {
				v := list[idx]
				resp, err := h.GetBRC20TokenInfo(v)
				if err != nil {
					return err
				}
				respList[idx] = resp
			}
			return nil
		})
	}
	if err := errWg.Wait(); err != nil {
		return err
	}

	ctx.JSON(http.StatusOK, gin.H{
		"page_index": page,
		"more":       len(list) > pageSize,
		"tokens":     respList,
	})
	return nil
}
