package util

import (
	"fmt"
	"github.com/inscription-c/insc/constants"
)

func NewInscriptionId(outpoint string, idx uint32) string {
	return fmt.Sprintf("%s%s%d", outpoint, constants.InscriptionIdDelimiter, idx)
}
