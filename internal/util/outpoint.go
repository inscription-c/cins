package util

import (
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
	"github.com/gogf/gf/v2/util/gconv"
	"github.com/inscription-c/insc/constants"
	"strings"
)

func StringToOutpoint(s string) *wire.OutPoint {
	s = strings.ToLower(s)
	if !constants.OutpointRegexp.MatchString(s) {
		return nil
	}
	insId := strings.Split(s, constants.OutpointDelimiter)
	h, _ := chainhash.NewHashFromStr(insId[0])
	return &wire.OutPoint{
		Hash:  *h,
		Index: gconv.Uint32(insId[1]),
	}
}
