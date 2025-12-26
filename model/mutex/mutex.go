package mutex

import (
	"01proxy/model/tags"
	"fmt"
	"path"
	"runtime"
	"sync"
	"sync/atomic"
)

type Mutex struct {
	mu     sync.Mutex
	nolock atomic.Bool
}

func (m *Mutex) Lock() {
	if m.nolock.Load() {
		return
	}
	m.mu.Lock()

	tags.Debug(func() {
		pc, file, line, _ := runtime.Caller(3)
		details := runtime.FuncForPC(pc)
		fmt.Printf("[debug] Lock from %s %s:%d\n", details.Name(), path.Base(file), line)
	})
}

func (m *Mutex) Unlock() {
	if m.nolock.Load() {
		return
	}
	m.mu.Unlock()

	tags.Debug(func() {
		pc, file, line, _ := runtime.Caller(2)
		details := runtime.FuncForPC(pc)
		fmt.Printf("[debug] Unlock from %s %s:%d\n", details.Name(), path.Base(file), line)
	})
}

func (m *Mutex) Bulk(callback func()) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.nolock.Store(true)
	callback()
	m.nolock.Store(false)
}

func (m *Mutex) TryLock() bool {
	return m.mu.TryLock()
}
