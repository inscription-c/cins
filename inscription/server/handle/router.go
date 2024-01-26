package handle

import "github.com/gin-contrib/pprof"

func (h *Handler) InitRouter() {
	if h.options.enablePProf {
		pprof.Register(h.Engine())
	}
	// inscriptions
	h.Engine().GET("/inscription/:query", h.Inscription)
	h.Engine().GET("/content/:inscriptionId", h.Content)
	h.Engine().GET("/inscriptions/:pages", h.Inscriptions)
	h.Engine().GET("/inscriptions/block/:height/:page", h.InscriptionsInBlockPage)

	// brc20c
	h.Engine().GET("/brc20c/token/:tkid", h.BRC20CToken)
	h.Engine().GET("/brc20c/tokens/:tk/:page", h.BRC20CTokens)
	h.Engine().GET("/brc20c/mint-history/:tkidOrAddr/:page", h.BRC20CMintHistory)
	h.Engine().GET("/brc20c/holders/:tkid/:page", h.BRC20CHolders)

	// block
	h.Engine().GET("/blockhash", h.BlockHash)
	h.Engine().GET("/blockhash/:height", h.BlockHash)
	h.Engine().GET("/blockheight", h.BlockHeight)
	h.Engine().GET("/clock", h.BlockClock)
	h.Engine().GET("/block/:height", h.InscriptionsInBlock)
	h.Engine().GET("/output/:output", h.InscriptionsInOutput)
}
