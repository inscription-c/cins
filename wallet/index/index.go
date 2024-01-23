package index

import (
	"context"
	"errors"
	"github.com/gin-gonic/gin"
	"github.com/inscription-c/insc/internal/net/multilistener"
	"github.com/inscription-c/insc/internal/signal"
	"github.com/inscription-c/insc/wallet/log"
	"net/http"
)

type Handler struct {
	*gin.Engine
}

func HandlerIndex(address string, multiListener *multilistener.Listener) {
	handler := &Handler{gin.New()}
	handler.InitRouter()
	srv := &http.Server{
		Addr:    address,
		Handler: handler,
	}
	signal.AddInterruptHandler(func() {
		_ = srv.Shutdown(context.Background())
	})
	go func() {
		if err := srv.Serve(multiListener); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Log.Error(err)
		}
	}()
}
