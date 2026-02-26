package server

import (
	"fmt"
	"gateway/config"
	"gateway/server/interfaces"
	"net/http"
)

type Server struct {
	*http.Server
	Gateway *Gateway
}

type ServerOptions struct {
	Gateway *Gateway

	// могут быть nil
	Handlers    map[string]http.Handler
	Middlewares []interfaces.Middleware
}

func NewServer(cfg config.ServerConfig, opts ServerOptions) *Server {
	mux := http.NewServeMux()
	mux.Handle("/", opts.Gateway.Handler())

	if opts.Handlers != nil {
		for path, handler := range opts.Handlers {
			mux.Handle(path, handler)
		}
	}
	if opts.Middlewares != nil {
		chain(mux, opts.Middlewares)
	}

	return &Server{
		Gateway: opts.Gateway,
		Server: &http.Server{
			Addr:         fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
			ReadTimeout:  cfg.ReadTimeout,
			WriteTimeout: cfg.WriteTimeout,
			Handler:      mux,
		},
	}
}
