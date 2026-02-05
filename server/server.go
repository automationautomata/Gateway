package server

import (
	"fmt"
	"gateway/config"
	"gateway/server/interfaces"
	"net/http"
	"time"
)

func NewServer(cfg config.ServerConfig, root http.Handler, mw ...interfaces.Middleware) *http.Server {
	return &http.Server{
		Addr:         fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Handler:      chain(root, mw),
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 10 * time.Second,
	}
}

func chain(h http.Handler, mws []interfaces.Middleware) http.Handler {
	for i := len(mws) - 1; i >= 0; i-- {
		h = mws[i].Wrap(h)
	}
	return h
}
