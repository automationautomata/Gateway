package server

import (
	"gateway/server/interfaces"
	"net/http"
)

func chain(h http.Handler, mws []interfaces.Middleware) http.Handler {
	for i := len(mws) - 1; i >= 0; i-- {
		h = mws[i].Wrap(h)
	}
	return h
}
