package keymutex

import (
	"fmt"
	"sync"
)

type entry struct {
	sync.Mutex
	cnt int
}

type KeyMutex[T comparable] struct {
	mapMu sync.Mutex
	m     map[T]*entry
}

func New[T comparable]() *KeyMutex[T] {
	return &KeyMutex[T]{m: make(map[T]*entry)}
}

func (km *KeyMutex[T]) Lock(key T) {
	km.mapMu.Lock()
	e, ok := km.m[key]
	if !ok {
		e = &entry{}
		km.m[key] = e
	}
	e.cnt++
	km.mapMu.Unlock()

	e.Lock()
}

func (km *KeyMutex[T]) Unlock(key T) {
	km.mapMu.Lock()
	e, ok := km.m[key]
	if !ok {
		km.mapMu.Unlock()
		panic(fmt.Errorf("Unlock requested for key=%v but no entry found", key))
	}
	e.cnt--
	if e.cnt < 1 {
		delete(km.m, key)
	}
	km.mapMu.Unlock()

	e.Unlock()
}
