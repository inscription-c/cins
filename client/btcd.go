package client

import (
	"github.com/dotbitHQ/insc/constants"
	"github.com/shopspring/decimal"
)

func (c *Client) EstimateFee(nBlock int) (fee uint64, err error) {
	btcFee := new(float64)
	if err := c.SendRequest("estimatefee", btcFee, nBlock); err != nil {
		return 0, err
	}
	fee = uint64(decimal.NewFromFloat(*btcFee).Mul(decimal.NewFromInt(constants.OneBtc)).IntPart())
	return
}

func (c *Client) SendRawTransaction(rawTx string) (txHash string, err error) {
	if err := c.SendRequest("sendrawtransaction", &txHash, rawTx); err != nil {
		return "", err
	}
	return
}
