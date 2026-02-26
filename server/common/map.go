package common

import "sync"

type SyncMap[K comparable, V any] struct {
	m *sync.Map
}

func NewSyncMap[K comparable, V any]() *SyncMap[K, V] {
	var m sync.Map
	return &SyncMap[K, V]{&m}
}

func (r *SyncMap[K, V]) Add(k string, v V) {
	r.m.Store(k, v)
}

func (r *SyncMap[K, V]) Get(k K) (V, bool) {
	v, ok := r.m.Load(k)
	if !ok {
		var empty V
		return empty, false
	}
	return v.(V), true
}
