package handle

import (
	sentrygin "github.com/getsentry/sentry-go/gin"
	"github.com/gin-contrib/pprof"
	"github.com/gin-gonic/gin"
	"github.com/inscription-c/cins/inscription/server/config"
	"github.com/inscription-c/cins/inscription/server/handle/middlewares"
)

func (h *Handler) InitRouter() {
	h.Engine().Use(gin.Recovery())
	if config.SrvCfg.Server.EnablePProf {
		pprof.Register(h.Engine())
	}
	if config.SrvCfg.Server.Prometheus {
		p := middlewares.NewPrometheus("gin")
		p.Use(h.Engine())
	}

	h.Engine().Use(middlewares.Cors(config.SrvCfg.Origins...))
	if config.SrvCfg.Sentry.Dsn != "" {
		h.Engine().Use(sentrygin.New(sentrygin.Options{
			Repanic: true,
		}))
	}
	h.Engine().Use(middlewares.Logger())

	// inscriptions
	h.Engine().GET("/inscription/:query", h.Inscription)
	h.Engine().GET("/content/:inscriptionId", h.Content)
	h.Engine().GET("/inscriptions/:pages", h.Inscriptions)
	h.Engine().GET("/inscriptions/block/:height/:page", h.InscriptionsInBlockPage)
	h.Engine().GET("/output/:output", h.InscriptionsInOutput)

	// cbrc20
	h.Engine().GET("/cbrc20/token/:tkid", h.BRC20CToken)
	h.Engine().GET("/cbrc20/tokens/:tk/:page", h.BRC20CTokens)

	// block
	h.Engine().GET("/blockhash", h.BlockHash)
	h.Engine().GET("/blockhash/:height", h.BlockHash)
	h.Engine().GET("/blockheight", h.BlockHeight)
	h.Engine().GET("/clock", h.BlockClock)
	h.Engine().GET("/block/:height", h.InscriptionsInBlock)

	r := h.Engine().Group("/r")
	r.GET("/blockheight", h.BlockHeight)
}
