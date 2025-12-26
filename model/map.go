package model

import (
	"01proxy/model/mutex"
)

type Map[K comparable, V any] struct {
	v  map[K]V
	mu mutex.Mutex
}

func NewTunnelMap[K comparable, V any]() Map[K, V] {
	return Map[K, V]{
		v: make(map[K]V),
	}
}

func (tm *Map[K, V]) Set(key K, value V) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	tm.v[key] = value
}

func (tm *Map[K, V]) Get(key K) (V, bool) {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	v, ok := tm.v[key]

	return v, ok
}

func (tm *Map[K, V]) Delete(key K) {
	tm.mu.Lock()
	defer tm.mu.Unlock()

	delete(tm.v, key)
}

func (tm *Map[K, V]) Size() int {
	tm.mu.Lock()
	defer tm.mu.Unlock()
	return len(tm.v)
}
