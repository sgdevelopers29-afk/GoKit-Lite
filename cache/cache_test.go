package cache_test

import (
	"testing"

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
	tests := []struct {
		name     string
		key      string
		value    string
		search   string
		expected string
		found    bool
	}{
		{
			name:     "existing key",
			key:      "name",
			value:    "Shreyas",
			search:   "name",
			expected: "Shreyas",
			found:    true,
		},
		{
			name:     "missing key",
			key:      "name",
			value:    "Shreyas",
			search:   "age",
			expected: "",
			found:    false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := cache.New[string, string]()
			c.Set(tt.key, tt.value)

			val, ok := c.Get(tt.search)
			if ok != tt.found {
				t.Errorf("expected found %v, got %v", tt.found, ok)
			}
			if val != tt.expected {
				t.Errorf("expected value %q, got %q", tt.expected, val)
			}
		})
	}
}

func TestCache_Delete(t *testing.T) {
	c := cache.New[int, string]()
	c.Set(1, "one")
	c.Set(2, "two")

	c.Delete(1)

	_, ok := c.Get(1)
	if ok {
		t.Error("expected key 1 to be deleted")
	}

	val, ok := c.Get(2)
	if !ok || val != "two" {
		t.Error("expected key 2 to remain")
	}
}

func TestCache_Size(t *testing.T) {
	c := cache.New[string, int]()
	if c.Size() != 0 {
		t.Errorf("expected size 0, got %d", c.Size())
	}

	c.Set("a", 1)
	c.Set("b", 2)
	
	if c.Size() != 2 {
		t.Errorf("expected size 2, got %d", c.Size())
	}

	c.Set("a", 3) // overwrite shouldn't increase size
	if c.Size() != 2 {
		t.Errorf("expected size 2 after overwrite, got %d", c.Size())
	}
}

func TestCache_Clear(t *testing.T) {
	c := cache.New[string, string]()
	c.Set("k1", "v1")
	c.Set("k2", "v2")

	if c.Size() != 2 {
		t.Fatalf("expected size 2, got %d", c.Size())
	}

	c.Clear()

	if c.Size() != 0 {
		t.Errorf("expected size 0 after clear, got %d", c.Size())
	}

	_, ok := c.Get("k1")
	if ok {
		t.Error("expected k1 to be cleared")
	}
}

func TestCache_MultipleEntries(t *testing.T) {
	c := cache.New[int, int]()
	
	for i := 0; i < 100; i++ {
		c.Set(i, i*10)
	}

	if c.Size() != 100 {
		t.Errorf("expected size 100, got %d", c.Size())
	}

	for i := 0; i < 100; i++ {
		val, ok := c.Get(i)
		if !ok {
			t.Errorf("expected key %d to be found", i)
		}
		if val != i*10 {
			t.Errorf("expected %d for key %d, got %d", i*10, i, val)
		}
	}
}
