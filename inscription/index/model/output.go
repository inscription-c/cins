package model

import (
	"fmt"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
	"github.com/gogf/gf/v2/util/gconv"
	"github.com/inscription-c/insc/constants"
	"strings"
)

type OutPoint struct {
	wire.OutPoint
}

func StringToOutpoint(s string) *OutPoint {
	s = strings.ToLower(s)
	if !constants.OutpointRegexp.MatchString(s) {
		return nil
	}
	insId := strings.Split(s, constants.OutpointDelimiter)
	h, _ := chainhash.NewHashFromStr(insId[0])
	return &OutPoint{
		OutPoint: wire.OutPoint{
			Hash:  *h,
			Index: gconv.Uint32(insId[1]),
		},
	}
}

func NewOutPoint(txId string, index uint32) *OutPoint {
	h, _ := chainhash.NewHashFromStr(txId)
	return &OutPoint{
		OutPoint: wire.OutPoint{
			Hash:  *h,
			Index: index,
		},
	}
}

func (o *OutPoint) String() string {
	return fmt.Sprintf("%s%s%d", o.Hash, constants.OutpointDelimiter, o.Index)
}

func (o *OutPoint) WireOutpoint() (*wire.OutPoint, error) {
	return wire.NewOutPointFromString(o.String())
}
