package pool

import "sync"

type Resettable interface {
	Reset()
}

type Pool[T Resettable] struct {
	pool sync.Pool
}

func New[T Resettable](newFn func() T) *Pool[T] {
	p := &Pool[T]{}
	p.pool.New = func() any { return newFn() }
	return p
}

func (p *Pool[T]) Get() T {
	return p.pool.Get().(T)
}

func (p *Pool[T]) Put(v T) {
	v.Reset()
	p.pool.Put(v)
}
