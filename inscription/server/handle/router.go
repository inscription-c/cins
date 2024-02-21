package handle

import (
	"github.com/gin-contrib/pprof"
	"github.com/inscription-c/insc/inscription/server/handle/middlewares"
)

func (h *Handler) InitRouter() {
	if h.options.enablePProf {
		pprof.Register(h.Engine())
	}
	h.Engine().Use(middlewares.Cors())
	// inscriptions
	h.Engine().GET("/inscription/:query", h.Inscription)
	h.Engine().GET("/content/:inscriptionId", h.Content)
	h.Engine().GET("/inscriptions/:pages", h.Inscriptions)
	h.Engine().GET("/inscriptions/block/:height/:page", h.InscriptionsInBlockPage)
	h.Engine().GET("/output/:output", h.InscriptionsInOutput)

	// cbrc20
	h.Engine().GET("/cbrc20/token/:tkid", h.BRC20CToken)
	h.Engine().GET("/cbrc20/tokens/:tk/:page", h.BRC20CTokens)
	//h.Engine().GET("/cbrc20/mint-history/:tkidOrAddr/:page", h.BRC20CMintHistory)
	//h.Engine().GET("/cbrc20/holders/:tkid/:page", h.BRC20CHolders)

	// block
	h.Engine().GET("/blockhash", h.BlockHash)
	h.Engine().GET("/blockhash/:height", h.BlockHash)
	h.Engine().GET("/blockheight", h.BlockHeight)
	h.Engine().GET("/clock", h.BlockClock)
	h.Engine().GET("/block/:height", h.InscriptionsInBlock)

	// scan
	scan := h.Engine().Group("/scan")
	scan.GET("/home/page/statistics", h.HomePageStatistics)
	scan.POST("/inscriptions", h.ScanInscriptions)
}
