package middlewares

import (
	"net"
	"net/http"
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
