// Package pool provides a generic object pool for types with Reset() method.
package pool

import "sync"


type Resettable interface {
	Reset()
}

type Pool[T Resettable] struct {
	pool sync.Pool
	new  func() T
}

func New[T Resettable](newFunc func() T) *Pool[T] {
	return &Pool[T]{
		pool: sync.Pool{
			New: func() interface{} {
				return newFunc()
			},
		},
		new: newFunc,
	}
}

func (p *Pool[T]) Get() T {
	obj := p.pool.Get()
	if obj == nil {
		return p.new()
	}
	return obj.(T)
}

func (p *Pool[T]) Put(obj T) {
	obj.Reset()
	p.pool.Put(obj)
}