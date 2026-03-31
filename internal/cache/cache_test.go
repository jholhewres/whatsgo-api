package cache

import (
	"sync"
	"testing"
	"time"
)

func TestSetAndGet(t *testing.T) {
	c := New[string, int](5 * time.Minute)

	c.Set("a", 1)
	val, ok := c.Get("a")
	if !ok {
		t.Fatal("expected key 'a' to exist")
	}
	if val != 1 {
		t.Fatalf("expected 1, got %d", val)
	}
}

func TestGetMissing(t *testing.T) {
	c := New[string, int](5 * time.Minute)

	_, ok := c.Get("missing")
	if ok {
		t.Fatal("expected key 'missing' to not exist")
	}
}

func TestExpiration(t *testing.T) {
	c := New[string, int](50 * time.Millisecond)

	c.Set("a", 1)
	time.Sleep(100 * time.Millisecond)

	_, ok := c.Get("a")
	if ok {
		t.Fatal("expected key 'a' to be expired")
	}
}

func TestSetWithTTL(t *testing.T) {
	c := New[string, int](5 * time.Minute)

	c.SetWithTTL("short", 1, 50*time.Millisecond)
	c.Set("long", 2)

	time.Sleep(100 * time.Millisecond)

	_, ok := c.Get("short")
	if ok {
		t.Fatal("expected 'short' to be expired")
	}

	val, ok := c.Get("long")
	if !ok || val != 2 {
		t.Fatal("expected 'long' to still exist with value 2")
	}
}

func TestInvalidate(t *testing.T) {
	c := New[string, int](5 * time.Minute)

	c.Set("a", 1)
	c.Invalidate("a")

	_, ok := c.Get("a")
	if ok {
		t.Fatal("expected key 'a' to be invalidated")
	}
}

func TestFlush(t *testing.T) {
	c := New[string, int](5 * time.Minute)

	c.Set("a", 1)
	c.Set("b", 2)
	c.Set("c", 3)
	c.Flush()

	for _, key := range []string{"a", "b", "c"} {
		if _, ok := c.Get(key); ok {
			t.Fatalf("expected key '%s' to be flushed", key)
		}
	}
}

func TestCleanup(t *testing.T) {
	c := New[string, int](50 * time.Millisecond)

	c.Set("expired", 1)
	c.SetWithTTL("alive", 2, 5*time.Minute)

	time.Sleep(100 * time.Millisecond)
	c.Cleanup()

	c.mu.RLock()
	_, expiredExists := c.entries["expired"]
	_, aliveExists := c.entries["alive"]
	c.mu.RUnlock()

	if expiredExists {
		t.Fatal("expected 'expired' to be cleaned up")
	}
	if !aliveExists {
		t.Fatal("expected 'alive' to survive cleanup")
	}
}

func TestOverwrite(t *testing.T) {
	c := New[string, int](5 * time.Minute)

	c.Set("a", 1)
	c.Set("a", 2)

	val, ok := c.Get("a")
	if !ok || val != 2 {
		t.Fatalf("expected 2, got %d", val)
	}
}

func TestConcurrentAccess(t *testing.T) {
	c := New[int, int](5 * time.Minute)
	var wg sync.WaitGroup

	// Writers
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			c.Set(n, n*2)
		}(i)
	}

	// Readers
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(n int) {
			defer wg.Done()
			c.Get(n)
		}(i)
	}

	// Cleanup
	wg.Add(1)
	go func() {
		defer wg.Done()
		c.Cleanup()
	}()

	wg.Wait()

	// Verify at least some values were written
	found := 0
	for i := 0; i < 100; i++ {
		if val, ok := c.Get(i); ok && val == i*2 {
			found++
		}
	}
	if found == 0 {
		t.Fatal("expected at least some values to be written")
	}
}

func TestInvalidateNonExistent(t *testing.T) {
	c := New[string, int](5 * time.Minute)
	// Should not panic
	c.Invalidate("nonexistent")
}
