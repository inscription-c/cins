package handle

import (
	"context"
	"errors"
	"fmt"
	"github.com/btcsuite/btcd/chaincfg"
	"github.com/btcsuite/btcd/rpcclient"
	"github.com/gin-gonic/gin"
	"github.com/inscription-c/insc/inscription/index/dao"
	"github.com/inscription-c/insc/inscription/log"
	"github.com/inscription-c/insc/internal/signal"
	"net/http"
	"os"
)

// Options is a struct that holds the configuration options for a Handler.
type Options struct {
	addr    string            // The address to bind the server to
	testnet bool              // Whether to use the testnet or not
	engin   *gin.Engine       // The gin engine for handling HTTP requests
	db      *dao.DB           // The database for storing data
	cli     *rpcclient.Client // The RPC client for interacting with the Bitcoin network
}

// Option is a function type that sets a specific option in an Options struct.
type Option func(*Options)

// WithAddr is a function that sets the address option for an Options struct.
// It takes a string representing the address and returns a function that sets the address option in the Options struct.
func WithAddr(addr string) func(*Options) {
	return func(options *Options) {
		options.addr = addr
	}
}

// WithEngin is a function that sets the gin engine option for an Options struct.
// It takes a pointer to a gin.Engine representing the gin engine and returns a function that sets the gin engine option in the Options struct.
func WithEngin(g *gin.Engine) func(*Options) {
	return func(options *Options) {
		options.engin = g
	}
}

// WithDB is a function that sets the database option for an Options struct.
// It takes a pointer to a dao.DB representing the database and returns a function that sets the database option in the Options struct.
func WithDB(db *dao.DB) func(*Options) {
	return func(options *Options) {
		options.db = db
	}
}

// WithTestNet is a function that sets the testnet option for an Options struct.
// It takes a boolean value representing whether to use the testnet or not and returns a function that sets the testnet option in the Options struct.
func WithTestNet(testnet bool) func(*Options) {
	return func(options *Options) {
		options.testnet = testnet
	}
}

// WithClient is a function that sets the rpcclient option for an Options struct.
// It takes a pointer to a rpcclient.Client representing the rpcclient and returns a function that sets the rpcclient option in the Options struct.
func WithClient(cli *rpcclient.Client) func(*Options) {
	return func(options *Options) {
		options.cli = cli
	}
}

// Handler is a struct that holds the options for handling requests.
type Handler struct {
	options *Options
}

// DB is a method that returns the database from the options of a Handler.
func (h *Handler) DB() *dao.DB {
	return h.options.db
}

// RpcClient is a method that returns the rpcclient from the options of a Handler.
func (h *Handler) RpcClient() *rpcclient.Client {
	return h.options.cli
}

// Engine is a method that returns the gin engine from the options of a Handler.
func (h *Handler) Engine() *gin.Engine {
	return h.options.engin
}

// GetChainParams is a method that returns the chain parameters based on the testnet option of a Handler.
// If the testnet option is set to true, it returns the TestNet3Params. Otherwise, it returns the MainNetParams.
func (h *Handler) GetChainParams() *chaincfg.Params {
	netParams := &chaincfg.MainNetParams
	if h.options.testnet {
		netParams = &chaincfg.TestNet3Params
	}
	return netParams
}

// New is a function that creates a new Handler with the given options.
// It takes a variadic number of Option functions and applies them to the options of the Handler.
// It returns a pointer to the newly created Handler and any error that occurred during the creation.
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

// Run is a method that starts the HTTP server of a Handler.
// It initializes the router, creates a new HTTP server and starts listening for requests.
// It also adds an interrupt handler that gracefully shuts down the server when an interrupt signal is received.
// It returns any error that occurred during the process.
func (h *Handler) Run() error {
	h.InitRouter()
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
