package handle

import (
	"fmt"
	"github.com/dustin/go-humanize"
	"github.com/gin-gonic/gin"
	"github.com/gogf/gf/v2/util/gconv"
	"github.com/inscription-c/insc/constants"
	"github.com/inscription-c/insc/inscription/server/handle/api"
	"github.com/inscription-c/insc/internal/util"
	"github.com/shopspring/decimal"
	"golang.org/x/sync/errgroup"
	"net/http"
)

type HomePageStatisticsResp struct {
	Inscriptions string `json:"inscriptions"`
	StoredData   string `json:"stored_data"`
	TotalFees    string `json:"total_fees"`
}

func (h *Handler) HomePageStatistics(ctx *gin.Context) {
	apiResp := &api.Resp{}
	if err := h.doHomePageStatistics(apiResp); err != nil {
		apiResp.ApiRespErr(api.CodeError500, err.Error())
	}
	ctx.JSON(http.StatusOK, apiResp)
}

func (h *Handler) doHomePageStatistics(apiResp *api.Resp) error {
	resp := &HomePageStatisticsResp{}

	errWg := &errgroup.Group{}
	errWg.Go(func() error {
		total, err := h.DB().InscriptionsNum()
		if err != nil {
			return err
		}
		resp.Inscriptions = util.NumberFormat(gconv.String(total))
		return nil
	})
	errWg.Go(func() error {
		storedData, err := h.DB().InscriptionsStoredData()
		if err != nil {
			return err
		}
		resp.StoredData = humanize.Bytes(storedData)
		return nil
	})
	errWg.Go(func() error {
		totalFees, err := h.DB().InscriptionsTotalFees()
		if err != nil {
			return err
		}
		btc := decimal.NewFromInt(int64(totalFees)).
			Div(decimal.NewFromInt(int64(constants.OneBtc)))
		resp.TotalFees = fmt.Sprintf("%s BTC", btc.String())
		return nil
	})
	apiResp.ApiRespOK(resp)
	return nil
}
