//go:build debug
// +build debug

package mutex

import (
	"fmt"
	"path"
	"runtime"
	"sync"
)

type Mutex struct {
	sync.Mutex
	nolock bool
}

func (m *Mutex) Lock() {
	if m.nolock {
		return
	}
	pc, file, line, _ := runtime.Caller(2)
	details := runtime.FuncForPC(pc)
	fmt.Printf("[debug] Lock from %s %s:%d\n", details.Name(), path.Base(file), line)
	m.Mutex.Lock()
}

func (m *Mutex) Unlock() {
	if m.nolock {
		return
	}
	pc, file, line, _ := runtime.Caller(2)
	details := runtime.FuncForPC(pc)
	fmt.Printf("[debug] Unlock from %s %s:%d\n", details.Name(), path.Base(file), line)
	m.Mutex.Unlock()
}

func (m *Mutex) Bulk(callback func()) {
	m.Lock()
	defer m.Unlock()
	m.nolock = true
	callback()
	m.nolock = false
}
