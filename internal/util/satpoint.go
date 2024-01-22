package util

import (
	"fmt"
	"github.com/inscription-c/insc/constants"
)

func NewSatPoint(outpoint string, sat uint64) string {
	return fmt.Sprintf("%s%s%d", outpoint, constants.OutpointDelimiter, sat)
}
