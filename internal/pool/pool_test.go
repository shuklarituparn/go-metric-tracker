package pool

import (
	"sync"
	"testing"
)

type TestStruct struct {
	ID   int
	Name string
	Tags []string
}

func (t *TestStruct) Reset() {
	if t == nil {
		return
	}
	t.ID = 0
	t.Name = ""
	t.Tags = t.Tags[:0]
}

func TestPool_GetReturnsNonNil(t *testing.T) {
	p := New(func() *TestStruct {
		return &TestStruct{Tags: make([]string, 0, 8)}
	})

	obj := p.Get()
	if obj == nil {
		t.Fatalf("expected non-nil object from pool")
	}
}

func TestPool_PutResetsObject(t *testing.T) {
	p := New(func() *TestStruct {
		return &TestStruct{Tags: make([]string, 0, 8)}
	})

	obj := p.Get()
	obj.ID = 123
	obj.Name = "hello"
	obj.Tags = append(obj.Tags, "a", "b", "c")

	p.Put(obj)

	obj2 := p.Get()

	if obj2.ID != 0 {
		t.Fatalf("expected ID reset to 0, got %d", obj2.ID)
	}
	if obj2.Name != "" {
		t.Fatalf("expected Name reset to empty, got %q", obj2.Name)
	}
	if len(obj2.Tags) != 0 {
		t.Fatalf("expected Tags length 0, got %d", len(obj2.Tags))
	}
	if cap(obj2.Tags) == 0 {
		t.Fatalf("expected Tags to retain capacity")
	}
}

func TestPool_ReusesObjects(t *testing.T) {
	p := New(func() *TestStruct { return &TestStruct{} })

	obj1 := p.Get()
	p.Put(obj1)

	obj2 := p.Get()

	_ = obj2
}

func TestPool_ConcurrentUsage(t *testing.T) {
	p := New(func() *TestStruct { return &TestStruct{} })

	const goroutines = 50
	const iters = 200

	var wg sync.WaitGroup
	wg.Add(goroutines)

	for i := 0; i < goroutines; i++ {
		go func(i int) {
			defer wg.Done()
			for j := 0; j < iters; j++ {
				obj := p.Get()
				obj.ID = i*1000 + j
				obj.Name = "x"
				p.Put(obj)
			}
		}(i)
	}

	wg.Wait()
}
