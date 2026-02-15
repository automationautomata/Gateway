package middlewares

import (
	"gateway/server/interfaces"
	"net/http"
	"runtime/debug"
)

type Recover struct {
	log interfaces.Logger
}

func NewRecover(log interfaces.Logger) *Recover {
	return &Recover{log}
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
