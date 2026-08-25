package cache

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func TestTTLBasicAndLazyExpiration(t *testing.T) {
	c := NewMemoryCache()

	// Set key with 50ms TTL
	err := c.SetWithTTL("tempKey", []byte("tempVal"), 50*time.Millisecond)
	if err != nil {
		t.Fatalf("unexpected error setting key: %v", err)
	}

	// Set key without TTL
	err = c.Set("permKey", []byte("permVal"))
	if err != nil {
		t.Fatalf("unexpected error setting key: %v", err)
	}

	// Immediately get tempKey
	val, err := c.Get("tempKey")
	if err != nil || string(val) != "tempVal" {
		t.Fatalf("expected tempVal, got %s (err: %v)", val, err)
	}

	// Sleep past expiration
	time.Sleep(70 * time.Millisecond)

	// Lazy expiration check
	_, err = c.Get("tempKey")
	if err != ErrKeyNotFound {
		t.Fatalf("expected ErrKeyNotFound for expired key, got %v", err)
	}

	// Permanent key should still exist
	val, err = c.Get("permKey")
	if err != nil || string(val) != "permVal" {
		t.Fatalf("expected permVal, got %s (err: %v)", val, err)
	}
}

func TestBackgroundCleanup(t *testing.T) {
	c := NewMemoryCache()

	// Start cleanup ticker every 20ms
	c.StartCleanup(20 * time.Millisecond)
	defer c.StopCleanup()

	c.SetWithTTL("k1", []byte("v1"), 30*time.Millisecond)
	c.SetWithTTL("k2", []byte("v2"), 30*time.Millisecond)
	c.Set("k3", []byte("v3"))

	if c.Len() != 3 {
		t.Fatalf("expected size 3, got %d", c.Len())
	}

	// Wait for background cleanup to run after expiration
	time.Sleep(80 * time.Millisecond)

	if c.Len() != 1 {
		t.Fatalf("expected size 1 after cleanup, got %d", c.Len())
	}

	val, err := c.Get("k3")
	if err != nil || string(val) != "v3" {
		t.Fatalf("expected k3 to persist, got %s (err %v)", val, err)
	}
}

func TestLRUEviction(t *testing.T) {
	c := NewMemoryCacheWithCapacity(3)

	c.Set("A", []byte("valA"))
	c.Set("B", []byte("valB"))
	c.Set("C", []byte("valC"))

	if c.Len() != 3 {
		t.Fatalf("expected capacity 3, got %d", c.Len())
	}

	// Access A so it becomes MRU (order: B, C, A) -> LRU is B
	c.Get("A")

	// Insert D -> should evict B
	c.Set("D", []byte("valD"))

	_, err := c.Get("B")
	if err != ErrKeyNotFound {
		t.Fatalf("expected B to be evicted, got err=%v", err)
	}

	// Check remaining keys A, C, D
	for _, k := range []string{"A", "C", "D"} {
		if _, err := c.Get(k); err != nil {
			t.Fatalf("expected key %s to exist", k)
		}
	}
}

func TestLRUUpdateExistingKey(t *testing.T) {
	c := NewMemoryCacheWithCapacity(3)

	c.Set("A", []byte("valA"))
	c.Set("B", []byte("valB"))
	c.Set("C", []byte("valC"))

	// Update A (order becomes B, C, A) -> LRU is B
	c.Set("A", []byte("valA2"))

	c.Set("D", []byte("valD")) // evicts B

	_, err := c.Get("B")
	if err != ErrKeyNotFound {
		t.Fatalf("expected B to be evicted, got err=%v", err)
	}

	val, err := c.Get("A")
	if err != nil || string(val) != "valA2" {
		t.Fatalf("expected valA2 for A, got %s (err %v)", val, err)
	}
}

func TestTTLAndLRUInteraction(t *testing.T) {
	c := NewMemoryCacheWithCapacity(3)

	// Set A with short TTL, B and C without TTL
	c.SetWithTTL("A", []byte("valA"), 40*time.Millisecond)
	c.Set("B", []byte("valB"))
	c.Set("C", []byte("valC"))

	// Sleep so A expires
	time.Sleep(60 * time.Millisecond)

	// A is expired. Insert D.
	c.Set("D", []byte("valD"))

	// A should not be returned
	_, err := c.Get("A")
	if err != ErrKeyNotFound {
		t.Fatalf("expected A to be expired/not found")
	}

	// B, C, D should all exist because A was replaced
	for _, k := range []string{"B", "C", "D"} {
		if _, err := c.Get(k); err != nil {
			t.Fatalf("expected key %s to exist", k)
		}
	}
}

func TestConcurrentTTLAndLRU(t *testing.T) {
	c := NewMemoryCacheWithCapacity(50)
	c.StartCleanup(10 * time.Millisecond)
	defer c.StopCleanup()

	var wg sync.WaitGroup
	const numGoroutines = 50
	const opsPerGoroutine = 100

	for i := 0; i < numGoroutines; i++ {
		wg.Add(1)
		go func(gID int) {
			defer wg.Done()
			for j := 0; j < opsPerGoroutine; j++ {
				key := fmt.Sprintf("key-%d", (gID*10+j)%100)
				val := fmt.Sprintf("val-%d", j)

				if j%3 == 0 {
					c.SetWithTTL(key, []byte(val), 20*time.Millisecond)
				} else if j%3 == 1 {
					c.Get(key)
				} else {
					c.Delete(key)
				}
			}
		}(i)
	}

	wg.Wait()
}
