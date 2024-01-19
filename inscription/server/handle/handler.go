package handle

import (
	"context"
	"errors"
	"fmt"
	"github.com/btcsuite/btcd/rpcclient"
	"github.com/gin-gonic/gin"
	"github.com/inscription-c/insc/inscription/index/dao"
	"github.com/inscription-c/insc/inscription/log"
	"github.com/inscription-c/insc/internal/signal"
	"net/http"
	"os"
)

type Options struct {
	addr    string
	testnet bool
	engin   *gin.Engine
	db      *dao.DB
	cli     *rpcclient.Client
}

type Option func(*Options)

func WithAddr(addr string) func(*Options) {
	return func(options *Options) {
		options.addr = addr
	}
}

func WithEngin(g *gin.Engine) func(*Options) {
	return func(options *Options) {
		options.engin = g
	}
}

func WithDB(db *dao.DB) func(*Options) {
	return func(options *Options) {
		options.db = db
	}
}
func WithTestNet(testnet bool) func(*Options) {
	return func(options *Options) {
		options.testnet = testnet
	}
}

func WithClient(cli *rpcclient.Client) func(*Options) {
	return func(options *Options) {
		options.cli = cli
	}
}

type Handler struct {
	options *Options
}

func New(opts ...Option) (*Handler, error) {
	h := &Handler{}
	h.options = &Options{}
	for _, opt := range opts {
		opt(h.options)
	}
	if h.options.addr == "" {
		h.options.addr = ":8335"
		if h.options.testnet {
			h.options.addr = ":18335"
		}
	}
	if h.options.db == nil {
		return nil, fmt.Errorf("db is nil")
	}
	if h.options.engin == nil {
		h.options.engin = gin.New()
	}
	return h, nil
}

func (h *Handler) Run() error {
	h.InitRoute()
	srv := &http.Server{
		Addr:    h.options.addr,
		Handler: h.options.engin,
	}
	signal.AddInterruptHandler(func() {
		if err := srv.Shutdown(context.Background()); err != nil {
			log.Srv.Error("srv.Shutdown", "err", err)
		}
	})
	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Srv.Error("srv.ListenAndServe", "err", err)
			os.Exit(1)
		}
	}()
	return nil
}

func (h *Handler) InitRoute() {
	h.options.engin.GET("/inscription/:query", h.Inscription)
}
