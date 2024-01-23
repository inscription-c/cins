package util

import (
	"fmt"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
	"github.com/gogf/gf/v2/util/gconv"
	"github.com/inscription-c/insc/constants"
	"strings"
)

func NewInscriptionId(outpoint string, idx uint32) string {
	return fmt.Sprintf("%s%s%d", outpoint, constants.InscriptionIdDelimiter, idx)
}

type InscriptionId struct {
	OutPoint
}

func (i *InscriptionId) String() string {
	return fmt.Sprintf("%s%s%d", i.Hash, constants.InscriptionIdDelimiter, i.Index)
}

func StringToInscriptionId(s string) *InscriptionId {
	s = strings.ToLower(strings.TrimSpace(s))
	if !constants.InscriptionIdRegexp.MatchString(s) {
		return nil
	}
	insId := strings.Split(s, constants.InscriptionIdDelimiter)
	h, _ := chainhash.NewHashFromStr(insId[0])
	return &InscriptionId{
		OutPoint: OutPoint{OutPoint: wire.OutPoint{
			Hash:  *h,
			Index: gconv.Uint32(insId[1]),
		}},
	}
}

func InscriptionIdToOutpoint(s string) *OutPoint {
	s = strings.ToLower(s)
	if !constants.InscriptionIdRegexp.MatchString(s) {
		return nil
	}
	insId := strings.Split(s, constants.InscriptionIdDelimiter)
	h, _ := chainhash.NewHashFromStr(insId[0])
	return &OutPoint{
		OutPoint: wire.OutPoint{
			Hash:  *h,
			Index: gconv.Uint32(insId[1]),
		},
	}
}
