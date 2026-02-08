package middlewares

import (
	"gateway/server/interfaces"
	"net"
	"net/http"
	"runtime/debug"
	"sync"
)

type Whitelist struct {
	clients sync.Map
}

func NewWhitelist(allowedClients ...string) *Whitelist {
	w := &Whitelist{}
	for _, key := range allowedClients {
		w.clients.Store(key, struct{}{})
	}
	return w
}

func (mw *Whitelist) Wrap(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		host, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			http.Error(w, "Invalid remote address", http.StatusBadRequest)
			return
		}
		if _, ok := mw.clients.Load(host); !ok {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	})
}

type Recover struct {
	log interfaces.Logger
}

func NewRecover(logger interfaces.Logger) *Recover {
	return &Recover{logger}
}

func (mw *Recover) Wrap(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		defer func() {
			if err := recover(); err != nil {
				mw.log.Error(r.Context(), "panic recovered", map[string]any{
					"error":  err,
					"method": r.Method,
					"path":   r.URL.Path,
					"remote": r.RemoteAddr,
					"stack":  string(debug.Stack()),
				})
				http.Error(w, "internal server error", http.StatusInternalServerError)
			}
		}()
		next.ServeHTTP(w, r)
	})
}
