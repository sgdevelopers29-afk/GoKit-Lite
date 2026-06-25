package cache_test

import (
	"sync"
	"testing"
	"time"

	"github.com/sgdevelopers29-afk/GoKit-Lite/cache"
)

func TestCache_New(t *testing.T) {
	c := cache.New[string, int]()
	if c == nil {
		t.Fatal("expected new cache instance, got nil")
	}
	if c.Size() != 0 {
		t.Errorf("expected size 0, got %d", c.Size())
	}
}

func TestCache_SetAndGet(t *testing.T) {
	c := cache.New[string, string]()

	// Permanent entry
	c.Set("permanent", "data")
	val, ok := c.Get("permanent")
	if !ok || val != "data" {
		t.Errorf("expected to get permanent data, got %v, %v", val, ok)
	}

	if !c.Has("permanent") {
		t.Errorf("expected cache to have permanent key")
	}

	// Missing key
	val, ok = c.Get("missing")
	if ok || val != "" {
		t.Errorf("expected not to find missing key, got %v, %v", val, ok)
	}
}

func TestCache_SetWithTTL_And_ExpiredEntry(t *testing.T) {
	c := cache.New[string, string]()

	c.SetWithTTL("temp", "volatile", 50*time.Millisecond)

	// Should exist immediately
	val, ok := c.Get("temp")
	if !ok || val != "volatile" {
		t.Errorf("expected to get temp data, got %v, %v", val, ok)
	}

	// Wait for expiration
	time.Sleep(100 * time.Millisecond)

	// Should be expired and automatically removed
	val, ok = c.Get("temp")
	if ok || val != "" {
		t.Errorf("expected temp to be expired, got %v, %v", val, ok)
	}
	
	if c.Has("temp") {
		t.Errorf("expected cache to not have temp key")
	}
}

func TestCache_CleanupWorker(t *testing.T) {
	c := cache.New[string, string]()
	
	// Start cleanup worker every 50ms
	c.StartCleanup(50 * time.Millisecond)
	
	// Ensure double-start doesn't block or panic
	c.StartCleanup(50 * time.Millisecond)

	c.SetWithTTL("k1", "v1", 20*time.Millisecond)
	c.SetWithTTL("k2", "v2", 150*time.Millisecond)

	// After 80ms, k1 should be cleaned up by worker, k2 should exist
	time.Sleep(80 * time.Millisecond)

	// We check size. Size includes expired if not cleaned up.
	// Because cleanup ran, k1 should be physically removed.
	if c.Size() != 1 {
		t.Errorf("expected size 1 after cleanup, got %d", c.Size())
	}

	c.StopCleanup()
	// Ensure double-stop doesn't panic
	c.StopCleanup()
}

func TestCache_DeleteAndClear(t *testing.T) {
	c := cache.New[int, string]()
	c.Set(1, "one")
	c.Set(2, "two")

	c.Delete(1)
	if c.Has(1) {
		t.Error("expected key 1 to be deleted")
	}

	if c.Size() != 1 {
		t.Errorf("expected size 1, got %d", c.Size())
	}

	c.Clear()
	if c.Size() != 0 {
		t.Errorf("expected size 0, got %d", c.Size())
	}
}

func TestCache_ConcurrentAccess(t *testing.T) {
	c := cache.New[int, int]()
	var wg sync.WaitGroup

	// Start cleanup worker to induce more concurrent map iterations
	c.StartCleanup(10 * time.Millisecond)
	defer c.StopCleanup()

	// 100 goroutines writing and reading
	for i := 0; i < 100; i++ {
		wg.Add(1)
		go func(val int) {
			defer wg.Done()
			
			// Concurrent Writes
			if val%2 == 0 {
				c.Set(val, val)
			} else {
				c.SetWithTTL(val, val, 20*time.Millisecond)
			}

			// Concurrent Reads
			_, _ = c.Get(val)
			_ = c.Has(val)
			_ = c.Size()
		}(i)
	}

	wg.Wait()

	// Perform random deletes concurrently
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(val int) {
			defer wg.Done()
			c.Delete(val)
		}(i)
	}

	wg.Wait()
}
