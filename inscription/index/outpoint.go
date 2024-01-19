package index

import (
	"github.com/gogf/gf/v2/util/gconv"
	"github.com/inscription-c/insc/constants"
)

func (tx *Tx) GetValueByOutpoint(outpoint string) (value int64, err error) {
	v, err := tx.Get(constants.BucketOutpointToValue, []byte(outpoint))
	if err != nil {
		return 0, err
	}
	value = gconv.Int64(string(v))
	return
}
