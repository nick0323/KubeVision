package cache

import (
	"testing"
	"time"
)

func TestMemoryCache_SetAndGet(t *testing.T) {
	c := NewMemoryCache(nil, nil)
	defer c.Close()

	c.Set("key1", "value1")

	val, ok := c.Get("key1")
	if !ok {
		t.Fatal("expected key1 to exist")
	}
	if val != "value1" {
		t.Errorf("expected 'value1', got %v", val)
	}
}

func TestMemoryCache_GetNonExistent(t *testing.T) {
	c := NewMemoryCache(nil, nil)
	defer c.Close()

	_, ok := c.Get("nonexistent")
	if ok {
		t.Error("expected nonexistent key to not exist")
	}
}

func TestMemoryCache_Delete(t *testing.T) {
	c := NewMemoryCache(nil, nil)
	defer c.Close()

	c.Set("key1", "value1")
	c.Delete("key1")

	_, ok := c.Get("key1")
	if ok {
		t.Error("expected key1 to be deleted")
	}
}

func TestMemoryCache_Clear(t *testing.T) {
	c := NewMemoryCache(nil, nil)
	defer c.Close()

	c.Set("key1", "value1")
	c.Set("key2", "value2")
	c.Clear()

	if size := c.Size(); size != 0 {
		t.Errorf("expected size 0 after clear, got %d", size)
	}
}

func TestMemoryCache_Expiry(t *testing.T) {
	c := NewMemoryCache(nil, nil)
	defer c.Close()

	c.SetWithTTL("key1", "value1", 50*time.Millisecond)

	_, ok := c.Get("key1")
	if !ok {
		t.Fatal("expected key1 to exist immediately after set")
	}

	time.Sleep(100 * time.Millisecond)

	_, ok = c.Get("key1")
	if ok {
		t.Error("expected key1 to be expired")
	}
}

func TestMemoryCache_Size(t *testing.T) {
	c := NewMemoryCache(nil, nil)
	defer c.Close()

	if size := c.Size(); size != 0 {
		t.Errorf("expected size 0, got %d", size)
	}

	c.Set("key1", "value1")
	c.Set("key2", "value2")

	if size := c.Size(); size != 2 {
		t.Errorf("expected size 2, got %d", size)
	}
}

func TestMemoryCache_Stats(t *testing.T) {
	c := NewMemoryCache(nil, nil)
	defer c.Close()

	c.Set("key1", "value1")
	c.Get("key1")
	c.Get("nonexistent")

	stats := c.GetStats()
	hits := stats["hits"].(int64)
	misses := stats["misses"].(int64)

	if hits != 1 {
		t.Errorf("expected 1 hit, got %d", hits)
	}
	if misses != 1 {
		t.Errorf("expected 1 miss, got %d", misses)
	}
}
