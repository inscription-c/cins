package handle

func (h *Handler) InitRouter() {
	h.Engine().GET("/inscription/:query", h.Inscription)
}
