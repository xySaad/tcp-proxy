package model

import (
	"slices"
	"sync"
)

type Pool[T any] struct {
	value []T
	sync.Mutex
}

func (p *Pool[T]) Size() int {
	p.Lock()
	defer p.Unlock()
	return len(p.value)
}

func NewPool[T any](value []T) *Pool[T] {
	return &Pool[T]{
		value: value,
	}
}

func (p *Pool[T]) Add(item T) {
	p.Lock()
	defer p.Unlock()

	p.value = append(p.value, item)
}

func (p *Pool[T]) Pop() {
	p.Lock()
	defer p.Unlock()

	p.value = p.value[:len(p.value)-1]
}

func (p *Pool[T]) Find(callback func(T) bool) *T {
	p.Lock()
	defer p.Unlock()
	i := slices.IndexFunc(p.value, callback)
	if i < 0 {
		return nil
	}
	return &p.value[i]
}

func (p *Pool[T]) ForEach(callback func(*T)) {
	p.Lock()
	defer p.Unlock()

	for _, v := range p.value {
		callback(&v)
	}
}
