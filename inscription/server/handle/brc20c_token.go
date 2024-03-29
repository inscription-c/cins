package handle

import (
	"github.com/gin-gonic/gin"
	"github.com/inscription-c/cins/constants"
	"github.com/inscription-c/cins/inscription/index/tables"
	"net/http"
)

// BRC20CToken is a handler function for handling BRC20C token requests.
// It validates the request parameters and calls the doBRC20CToken function.
func (h *Handler) BRC20CToken(ctx *gin.Context) {
	tkid := ctx.Param("tkid")
	if tkid == "" {
		ctx.String(http.StatusBadRequest, "invalid token id")
		return
	}
	if err := h.doBRC20CToken(ctx, tkid); err != nil {
		ctx.String(http.StatusInternalServerError, err.Error())
		return
	}
}

// doBRC20CToken is a helper function for handling BRC20C token requests.
// It retrieves the token information of a specific BRC20C token and returns them in the response.
func (h *Handler) doBRC20CToken(ctx *gin.Context, tkid string) error {
	inscriptionId := tables.StringToInscriptionId(tkid)
	token, err := h.DB().GetProtocolByInscriptionId(inscriptionId)
	if err != nil {
		return err
	}
	if token.Id == 0 {
		ctx.Status(http.StatusNotFound)
		return nil
	}
	if token.Protocol != constants.ProtocolCBRC20 {
		ctx.Status(http.StatusNotFound)
		return nil
	}

	//if token.Operator == constants.OperationMint {
	//	//inscriptionId := tables.StringToInscriptionId(token.TkID)
	//	token, err = h.DB().GetProtocolByInscriptionId(inscriptionId)
	//	if err != nil {
	//		return err
	//	}
	//	if token.Id == 0 {
	//		ctx.Status(http.StatusNotFound)
	//		return nil
	//	}
	//}

	resp, err := h.GetBRC20TokenInfo(&token)
	if err != nil {
		return err
	}
	ctx.JSON(http.StatusOK, resp)
	return nil
}

// GetBRC20TokenInfo retrieves the information of a specific BRC20C token.
// It takes a token of type *tables.Protocol as a parameter.
// It returns the token information and any error encountered.
func (h *Handler) GetBRC20TokenInfo(token *tables.Protocol) (gin.H, error) {
	//lock := &sync.Mutex{}
	resp := gin.H{
		"ticker_id":    token.InscriptionId,
		"ticker":       token.Ticker,
		"total_supply": token.Max,
	}

	//errWg := &errgroup.Group{}
	//errWg.Go(func() error {
	//	inscription, err := h.DB().GetInscriptionBySequenceNum(token.SequenceNum)
	//	if err != nil {
	//		return err
	//	}
	//	if inscription.Id == 0 {
	//		return errors.New("inscription not found")
	//	}
	//	lock.Lock()
	//	resp["c_ins_description"] = inscription.CInsDescription
	//	//resp["metadata"] = hex.EncodeToString(inscription.Metadata)
	//	lock.Unlock()
	//	return nil
	//})
	//
	//if err := errWg.Wait(); err != nil {
	//	return nil, err
	//}
	return resp, nil
}
