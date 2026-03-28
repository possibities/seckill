package pkg

import (
	"sync"
	"testing"
)

func TestNewSnowflakeInvalidWorkerID(t *testing.T) {
	if _, err := NewSnowflake(-1); err == nil {
		t.Fatal("expected error for negative workerID")
	}
	if _, err := NewSnowflake(maxWorkerID + 1); err == nil {
		t.Fatal("expected error for overflow workerID")
	}
}

func TestSnowflakeIncreasing(t *testing.T) {
	s, err := NewSnowflake(1)
	if err != nil {
		t.Fatalf("new snowflake: %v", err)
	}

	last := int64(0)
	for i := 0; i < 1000; i++ {
		id, genErr := s.NextID()
		if genErr != nil {
			t.Fatalf("next id: %v", genErr)
		}
		if id <= last {
			t.Fatalf("id not increasing: current=%d last=%d", id, last)
		}
		last = id
	}
}

func TestSnowflakeConcurrentUnique(t *testing.T) {
	s, err := NewSnowflake(2)
	if err != nil {
		t.Fatalf("new snowflake: %v", err)
	}

	const n = 5000
	ids := make(chan int64, n)
	var wg sync.WaitGroup

	for i := 0; i < n; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			id, genErr := s.NextID()
			if genErr != nil {
				t.Errorf("next id: %v", genErr)
				return
			}
			ids <- id
		}()
	}

	wg.Wait()
	close(ids)

	seen := make(map[int64]struct{}, n)
	for id := range ids {
		if _, ok := seen[id]; ok {
			t.Fatalf("duplicate id found: %d", id)
		}
		seen[id] = struct{}{}
	}

	if len(seen) != n {
		t.Fatalf("unexpected size: got=%d want=%d", len(seen), n)
	}
}
