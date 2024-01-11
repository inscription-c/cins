package client

import (
	"fmt"
	"github.com/btcsuite/btcd/btcjson"
	"github.com/dotbitHQ/insc/constants"
	"github.com/gogf/gf/v2/util/gconv"
	"github.com/shopspring/decimal"
	"regexp"
	"strings"
)

var (
	inscriptionIdDelimiter = "ic"
	outpointDelimiter      = ":"
	idRegexpContent        = `^[A-Za-z0-9]{64}%s\d+$`
	inscriptionIdRegexp    = regexp.MustCompile(fmt.Sprintf(idRegexpContent, inscriptionIdDelimiter))
	outpointRegexp         = regexp.MustCompile(fmt.Sprintf(idRegexpContent, outpointDelimiter))
)

type Response struct {
	Jsonrpc btcjson.RPCVersion `json:"jsonrpc"`
	Result  interface{}        `json:"result"`
	Error   *btcjson.RPCError  `json:"error"`
	ID      *interface{}       `json:"id"`
}

type Amount float64

func (a Amount) Sat() uint64 {
	return uint64(decimal.NewFromFloat(float64(a)).
		Mul(decimal.NewFromInt(constants.OneBtc)).IntPart())
}

type OutPoint struct {
	Txid string `json:"txid"`
	Vout uint64 `json:"vout"`
}

func (o *OutPoint) String() string {
	return fmt.Sprintf("%s%s%d", o.Txid, inscriptionIdDelimiter, o.Vout)
}

func (o *OutPoint) Outpoint() string {
	return fmt.Sprintf("%s%s%d", o.Txid, outpointDelimiter, o.Vout)
}

func StringToOutpoint(s string) *OutPoint {
	s = strings.ToLower(s)
	if !outpointRegexp.MatchString(s) {
		return nil
	}
	insId := strings.Split(s, outpointDelimiter)
	return &OutPoint{
		Txid: insId[0],
		Vout: gconv.Uint64(insId[1]),
	}
}

func InscriptionIdToOutpoint(s string) *OutPoint {
	s = strings.ToLower(s)
	if !inscriptionIdRegexp.MatchString(s) {
		return nil
	}
	insId := strings.Split(s, inscriptionIdDelimiter)
	return &OutPoint{
		Txid: insId[0],
		Vout: gconv.Uint64(insId[1]),
	}
}
