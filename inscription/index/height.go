package index

type Height struct {
	height uint32
}

func NewHeight(height uint32) *Height {
	return &Height{
		height: height,
	}
}

func (h *Height) N() uint32 {
	return h.height
}

func (h *Height) Subsidy() uint32 {
	return NewEpochFrom(h).Subsidy()
}
