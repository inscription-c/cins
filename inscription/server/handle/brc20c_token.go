package handle

import (
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/inscription-c/insc/constants"
	"github.com/inscription-c/insc/inscription/index/tables"
	"github.com/inscription-c/insc/internal/util"
	"golang.org/x/sync/errgroup"
	"net/http"
	"sync"
)

func (h *Handler) BRC20CToken(ctx *gin.Context) {
	tkid := ctx.Query("tkid")
	if tkid == "" {
		ctx.Status(http.StatusBadRequest)
		return
	}
	if err := h.doBRC20CToken(ctx, tkid); err != nil {
		ctx.Status(http.StatusInternalServerError)
		return
	}
}

func (h *Handler) doBRC20CToken(ctx *gin.Context, tkid string) error {
	inscriptionId := util.StringToInscriptionId(tkid)
	token, err := h.DB().GetProtocolByOutpoint(inscriptionId.OutPoint.String())
	if err != nil {
		return err
	}
	if token.Id == 0 {
		ctx.Status(http.StatusNotFound)
		return nil
	}
	if token.Protocol != constants.ProtocolBRC20C {
		ctx.Status(http.StatusNotFound)
		return nil
	}

	if token.Operator == constants.OperationMint {
		token, err = h.DB().GetProtocolByOutpoint(token.TkID)
		if err != nil {
			return err
		}
		if token.Id == 0 {
			ctx.Status(http.StatusNotFound)
			return nil
		}
	}

	resp, err := h.GetBRC20TokenInfo(&token)
	if err != nil {
		return err
	}

	ctx.JSON(http.StatusOK, resp)
	return nil
}

func (h *Handler) GetBRC20TokenInfo(token *tables.Protocol) (gin.H, error) {
	lock := &sync.Mutex{}
	resp := gin.H{
		"ticker_id":    util.StringToOutpoint(token.Outpoint).InscriptionId().String(),
		"ticker":       token.Ticker,
		"total_supply": token.Max,
	}

	errWg := &errgroup.Group{}
	errWg.Go(func() error {
		inscription, err := h.DB().GetInscriptionBySequenceNum(token.SequenceNum)
		if err != nil {
			return err
		}
		if inscription.Id == 0 {
			return errors.New("inscription not found")
		}

		lock.Lock()
		resp["dst_chain"] = inscription.DstChain
		//resp["metadata"] = hex.EncodeToString(inscription.Metadata)
		lock.Unlock()
		return nil
	})

	errWg.Go(func() error {
		amount, err := h.DB().CountProtocolAmount(constants.ProtocolBRC20C, token.Ticker, constants.OperationMint, token.Outpoint)
		if err != nil {
			return err
		}
		lock.Lock()
		resp["circulating_supply"] = amount
		lock.Unlock()
		return nil
	})

	errWg.Go(func() error {
		holders, err := h.DB().CountToAddress(constants.ProtocolBRC20C, token.Ticker, constants.OperationMint, token.Outpoint)
		if err != nil {
			return err
		}
		lock.Lock()
		resp["holders"] = holders
		lock.Unlock()
		return nil
	})

	if err := errWg.Wait(); err != nil {
		return nil, err
	}
	return resp, nil
}
