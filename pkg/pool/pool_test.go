package pool

import (
	"testing"
)

type TestStruct struct {
	ID    int
	Name  string
	Items []string
	Data  map[string]int
}

func (ts *TestStruct) Reset() {
	ts.ID = 0
	ts.Name = ""
	ts.Items = ts.Items[:0]
	clear(ts.Data)
}

func TestPool_GetPut(t *testing.T) {
	pool := New(func() *TestStruct {
		return &TestStruct{
			Items: make([]string, 0, 10),
			Data:  make(map[string]int),
		}
	})

	obj := pool.Get()
	if obj == nil {
		t.Fatal("Get returned nil")
	}

	obj.ID = 42
	obj.Name = "test"
	obj.Items = append(obj.Items, "item1", "item2")
	obj.Data["key"] = 100

	pool.Put(obj)

	obj2 := pool.Get()

	if obj2.ID != 0 {
		t.Errorf("Expected ID=0, got %d", obj2.ID)
	}
	if obj2.Name != "" {
		t.Errorf("Expected Name='', got '%s'", obj2.Name)
	}
	if len(obj2.Items) != 0 {
		t.Errorf("Expected len(Items)=0, got %d", len(obj2.Items))
	}
	if len(obj2.Data) != 0 {
		t.Errorf("Expected len(Data)=0, got %d", len(obj2.Data))
	}
}

func TestPool_PreservesCapacity(t *testing.T) {
	pool := New(func() *TestStruct {
		return &TestStruct{
			Items: make([]string, 0, 100),
		}
	})

	obj := pool.Get()
	obj.Items = append(obj.Items, "a", "b", "c")
	
	if cap(obj.Items) < 100 {
		t.Errorf("Expected capacity >= 100, got %d", cap(obj.Items))
	}

	pool.Put(obj)
	
	obj2 := pool.Get()
	
	if len(obj2.Items) != 0 {
		t.Errorf("Expected length 0, got %d", len(obj2.Items))
	}
	if cap(obj2.Items) < 100 {
		t.Errorf("Expected capacity >= 100, got %d", cap(obj2.Items))
	}
}

func TestPool_ResetHandlesNil(t *testing.T) {
	pool := New(func() *TestStruct {
		return &TestStruct{
			Items: make([]string, 0, 10),
			Data:  make(map[string]int),
		}
	})

	obj := pool.Get()
	obj.ID = 42
	obj.Name = "test"

	pool.Put(obj)

	obj2 := pool.Get()
	if obj2.ID != 0 || obj2.Name != "" {
		t.Error("Object was not properly reset")
	}
}

func TestPool_Concurrent(t *testing.T) {
	pool := New(func() *TestStruct {
		return &TestStruct{
			Items: make([]string, 0, 10),
			Data:  make(map[string]int),
		}
	})

	const goroutines = 100
	const operations = 1000

	done := make(chan bool)

	for i := 0; i < goroutines; i++ {
		go func(id int) {
			for j := 0; j < operations; j++ {
				obj := pool.Get()
				
				obj.ID = id
				obj.Name = "test"
				obj.Items = append(obj.Items, "item")
				obj.Data["key"] = j
				
				pool.Put(obj)
			}
			done <- true
		}(i)
	}

	for i := 0; i < goroutines; i++ {
		<-done
	}

	obj := pool.Get()
	if obj.ID != 0 || obj.Name != "" || len(obj.Items) != 0 || len(obj.Data) != 0 {
		t.Error("Object was not properly reset after concurrent usage")
	}
}

func BenchmarkPool_WithPool(b *testing.B) {
	pool := New(func() *TestStruct {
		return &TestStruct{
			Items: make([]string, 0, 100),
			Data:  make(map[string]int, 100),
		}
	})

	b.ResetTimer()
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			obj := pool.Get()
			obj.ID = 42
			obj.Name = "benchmark"
			for i := 0; i < 10; i++ {
				obj.Items = append(obj.Items, "item")
				obj.Data["key"] = i
			}
			pool.Put(obj)
		}
	})
}

func BenchmarkPool_WithoutPool(b *testing.B) {
	b.RunParallel(func(pb *testing.PB) {
		for pb.Next() {
			obj := &TestStruct{
				Items: make([]string, 0, 100),
				Data:  make(map[string]int, 100),
			}
			obj.ID = 42
			obj.Name = "benchmark"
			for i := 0; i < 10; i++ {
				obj.Items = append(obj.Items, "item")
				obj.Data["key"] = i
			}
		}
	})
}