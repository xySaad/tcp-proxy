//go:build !debug
// +build !debug

package mutex

import (
	"sync"
)

type Mutex struct {
	mu     sync.Mutex
	nolock bool
}

func (m *Mutex) Lock() {
	if m.nolock {
		return
	}
	m.mu.Lock()
}

func (m *Mutex) Unlock() {
	if m.nolock {
		return
	}
	m.mu.Unlock()
}

func (m *Mutex) Bulk(callback func()) {
	m.Lock()
	defer m.Unlock()
	m.nolock = true
	callback()
	m.nolock = false
}

func (m *Mutex) TryLock() bool {
	return m.mu.TryLock()
}
