package model

import (
	"01proxy/model/mutex"
	"slices"
)

type Pool[T any] struct {
	value []T
	mutex.Mutex
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

func (p *Pool[T]) Find(callback func(T) bool) T {
	p.Lock()
	defer p.Unlock()

	for _, v := range p.value {
		if callback(v) {
			return v
		}
	}

	var t T
	return t
}

func (p *Pool[T]) RemoveBy(predicate func(T) bool) {
	p.Lock()
	defer p.Unlock()

	for i, v := range p.value {
		if predicate(v) {
			p.value = slices.Delete(p.value, i, i+1)
			return
		}
	}

}

func (p *Pool[T]) Clear() []T {
	p.Lock()
	defer p.Unlock()

	items := p.value
	p.value = nil
	return items
}

func (p *Pool[T]) Value() []T {
	p.Lock()
	defer p.Unlock()

	return append([]T(nil), p.value...)
}
