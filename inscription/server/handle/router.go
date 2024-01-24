package handle

import "github.com/gin-contrib/pprof"

func (h *Handler) InitRouter() {
	if h.options.enablePProf {
		pprof.Register(h.Engine())
	}
	h.Engine().GET("/inscription/:query", h.Inscription)
	h.Engine().GET("/content/:inscriptionId", h.Content)
	h.Engine().GET("/inscriptions/:pages", h.Inscriptions)
	h.Engine().GET("/inscriptions/block/:height", h.InscriptionsInBlock)
	h.Engine().GET("/inscriptions/block/:height/:page", h.InscriptionsInBlock)
}
