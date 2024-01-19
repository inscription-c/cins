package index

import "github.com/inscription-c/insc/constants"

type Height struct {
	Height uint64
}

func (h *Height) Subsidy() uint64 {
	if h.Height < 33 {
		return (50 * constants.OneBtc) >> h.Height
	}
	return 0
}
