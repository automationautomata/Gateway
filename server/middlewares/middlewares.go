package middlewares

import (
	"gateway/internal/common"
	"net"
	"net/http"
)

type Whitelist struct {
	clients common.Set[string]
}

func NewWhitelist(allowedClients ...string) *Whitelist {
	return &Whitelist{common.NewSet(allowedClients...)}
}

func (mw *Whitelist) Wrap(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		host, _, err := net.SplitHostPort(r.RemoteAddr)
		if err != nil {
			http.Error(w, "Invalid remote address", http.StatusBadRequest)
			return
		}

		if !mw.clients.Has(host) {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}

		next.ServeHTTP(w, r)
	})
}
