package handle

import "github.com/gin-contrib/pprof"

func (h *Handler) InitRouter() {
	if h.options.enablePProf {
		pprof.Register(h.Engine())
	}
	h.Engine().GET("/inscription/:query", h.Inscription)
	h.Engine().GET("/content/:inscriptionId", h.Content)
}
