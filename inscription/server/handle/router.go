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
	h.Engine().GET("/brc20c/token/:tkid", h.BRC20CToken)
	h.Engine().GET("/brc20c/tokens/:tk/:page", h.BRC20CTokens)
	h.Engine().GET("/brc20c/mint-history/:tkidOrAddr/:page", h.BRC20CMintHistory)
	h.Engine().GET("/brc20c/holders/:tkid/:page", h.BRC20CHolders)
}
