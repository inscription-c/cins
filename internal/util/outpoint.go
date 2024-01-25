package util

import (
	"database/sql/driver"
	"errors"
	"fmt"
	"github.com/btcsuite/btcd/chaincfg/chainhash"
	"github.com/btcsuite/btcd/wire"
	"github.com/gogf/gf/v2/util/gconv"
	"github.com/inscription-c/insc/constants"
	"strings"
)

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

type OutPoint struct {
	wire.OutPoint
}

func (o *OutPoint) Scan(value interface{}) error {
	bytes, ok := value.([]byte)
	if !ok {
		return errors.New(fmt.Sprint("Failed to unmarshal JSONB value:", value))
	}
	outpoint := StringToOutpoint(string(bytes))
	*o = *outpoint
	return nil
}

func (o *OutPoint) Value() (driver.Value, error) {
	return []byte(o.String()), nil
}

func (o *OutPoint) MarshalJSON() ([]byte, error) {
	s := fmt.Sprintf("\"%s\"", o.InscriptionId().String())
	return []byte(s), nil
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

func (o *OutPoint) InscriptionId() *InscriptionId {
	return &InscriptionId{
		OutPoint: *o,
	}
}
