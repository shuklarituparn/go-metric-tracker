package pool

import "sync"

type Resettable interface {
	Reset()
}

type Pool[T any] struct {
	pool sync.Pool
	new  func() *T
}

func New[T any](newFunc func() *T) *Pool[T] {
	p := &Pool[T]{new: newFunc}
	p.pool.New = func() any { return newFunc() }
	return p
}

func (p *Pool[T]) Get() *T {
	return p.pool.Get().(*T)
}

func (p *Pool[T]) Put(obj *T) {
	if obj == nil {
		return
	}
	if resetter, ok := any(obj).(Resettable); ok {
		resetter.Reset()
	}
	p.pool.Put(obj)
}
