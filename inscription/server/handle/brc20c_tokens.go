package handle

import (
	"github.com/gin-gonic/gin"
	"github.com/gogf/gf/v2/util/gconv"
	"github.com/inscription-c/insc/constants"
	"golang.org/x/sync/errgroup"
	"net/http"
)

// BRC20CTokens is a handler function for handling BRC20C tokens requests.
// It validates the request parameters and calls the doBRC20CTokens function.
func (h *Handler) BRC20CTokens(ctx *gin.Context) {
	tk := ctx.Param("tk")
	page := ctx.Param("page")
	if page == "" {
		page = "1"
	}
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

// doBRC20CTokens is a helper function for handling BRC20C tokens requests.
// It retrieves the tokens of a specific BRC20C token and returns them in the response.
func (h *Handler) doBRC20CTokens(ctx *gin.Context, tk string, page int) error {
	pageSize := 100
	list, err := h.DB().FindTokenPageByTicker(constants.ProtocolCBRC20, tk, constants.OperationDeploy, page, pageSize)
	if err != nil {
		return err
	}
	more := false
	if len(list) > pageSize {
		more = true
		list = list[:pageSize]
	}

	currentNum := 10
	errWg := &errgroup.Group{}
	ch := make(chan int, currentNum)
	respList := make([]gin.H, len(list))

	// This goroutine is responsible for sending indices to the channel for processing.
	errWg.Go(func() error {
		for idx := range respList {
			ch <- idx
		}
		close(ch)
		return nil
	})

	// These goroutines are responsible for processing the indices sent to the channel.
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

	// Respond with the tokens list, page index and a flag indicating if there are more tokens
	ctx.JSON(http.StatusOK, gin.H{
		"page_index": page,
		"more":       more,
		"tokens":     respList,
	})
	return nil
}
