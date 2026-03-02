package server

import (
	"gateway/server/interfaces"
	"net/http"
	"sync"
)

type syncMap[K comparable, V any] struct {
	m *sync.Map
}

func newSyncMap[K comparable, V any]() *syncMap[K, V] {
	var m sync.Map
	return &syncMap[K, V]{&m}
}

func (r *syncMap[K, V]) add(k string, v V) {
	r.m.Store(k, v)
}

func (r *syncMap[K, V]) get(k K) (V, bool) {
	v, ok := r.m.Load(k)
	if !ok {
		var empty V
		return empty, false
	}
	return v.(V), true
}

func chain(h http.Handler, mws []interfaces.Middleware) http.Handler {
	for i := len(mws) - 1; i >= 0; i-- {
		h = mws[i].Wrap(h)
	}
	return h
}
