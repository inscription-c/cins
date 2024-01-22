package handle

func (h *Handler) InitRouter() {
	h.options.engin.GET("/inscription/:query", h.Inscription)
}
