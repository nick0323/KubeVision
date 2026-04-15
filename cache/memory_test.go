package cache

import (
	"testing"
	"time"
)

// TestMemoryCache_SetGet 测试基本设置和获取
func TestMemoryCache_SetGet(t *testing.T) {
	cfg := MemoryCacheConfig{
		MaxSize:         100,
		TTL:             5 * time.Minute,
		CleanupInterval: 1 * time.Minute,
		Enabled:         true,
	}

	cache := NewMemoryCacheWithConfig[string](cfg)
	defer cache.Close()

	// 测试 Set 和 Get
	cache.Set("key1", "value1")
	val, found := cache.Get("key1")

	if !found {
		t.Error("Expected to find key1, but not found")
	}
	if val != "value1" {
		t.Errorf("Expected value1, got %v", val)
	}

	// 测试不存在的键
	_, found = cache.Get("nonexistent")
	if found {
		t.Error("Expected not to find nonexistent key, but found it")
	}
}

// TestMemoryCache_TTL 测试 TTL 过期
func TestMemoryCache_TTL(t *testing.T) {
	cfg := MemoryCacheConfig{
		MaxSize:         100,
		TTL:             50 * time.Millisecond,
		CleanupInterval: 1 * time.Minute,
		Enabled:         true,
	}

	cache := NewMemoryCacheWithConfig[string](cfg)
	defer cache.Close()

	cache.Set("key1", "value1")

	// 立即获取应该成功
	val, found := cache.Get("key1")
	if !found || val != "value1" {
		t.Error("Expected to find key1 immediately after set")
	}

	// 等待过期
	time.Sleep(100 * time.Millisecond)

	// 获取应该失败
	_, found = cache.Get("key1")
	if found {
		t.Error("Expected key1 to be expired")
	}
}

// TestMemoryCache_LRU 测试 LRU 淘汰
func TestMemoryCache_LRU(t *testing.T) {
	cfg := MemoryCacheConfig{
		MaxSize:         3,
		TTL:             5 * time.Minute,
		CleanupInterval: 1 * time.Minute,
		Enabled:         true,
	}

	cache := NewMemoryCacheWithConfig[string](cfg)
	defer cache.Close()

	// 填充缓存
	cache.Set("key1", "value1")
	time.Sleep(10 * time.Millisecond) // 确保时间戳不同
	cache.Set("key2", "value2")
	time.Sleep(10 * time.Millisecond)
	cache.Set("key3", "value3")

	// 访问 key1 使其成为最近使用的
	cache.Get("key1")

	// 打印当前状态帮助调试
	t.Logf("Size before adding key4: %d", cache.Size())
	t.Logf("Keys before adding key4: %v", cache.Keys())

	// 添加新项应该淘汰 key2（最久未访问）
	cache.Set("key4", "value4")

	t.Logf("Size after adding key4: %d", cache.Size())
	t.Logf("Keys after adding key4: %v", cache.Keys())

	_, found := cache.Get("key2")
	if found {
		t.Error("Expected key2 to be evicted")
	}

	// key1 和 key3 应该还在
	_, found = cache.Get("key1")
	if !found {
		t.Error("Expected key1 to still exist")
	}
	_, found = cache.Get("key3")
	if !found {
		t.Error("Expected key3 to still exist")
	}
	_, found = cache.Get("key4")
	if !found {
		t.Error("Expected key4 to exist")
	}
}

// TestMemoryCache_Delete 测试删除
func TestMemoryCache_Delete(t *testing.T) {
	cfg := MemoryCacheConfig{
		MaxSize:         100,
		TTL:             5 * time.Minute,
		CleanupInterval: 1 * time.Minute,
		Enabled:         true,
	}

	cache := NewMemoryCacheWithConfig[string](cfg)
	defer cache.Close()

	cache.Set("key1", "value1")
	cache.Delete("key1")

	_, found := cache.Get("key1")
	if found {
		t.Error("Expected key1 to be deleted")
	}
}

// TestMemoryCache_Clear 测试清空
func TestMemoryCache_Clear(t *testing.T) {
	cfg := MemoryCacheConfig{
		MaxSize:         100,
		TTL:             5 * time.Minute,
		CleanupInterval: 1 * time.Minute,
		Enabled:         true,
	}

	cache := NewMemoryCacheWithConfig[string](cfg)
	defer cache.Close()

	cache.Set("key1", "value1")
	cache.Set("key2", "value2")
	cache.Clear()

	if cache.Size() != 0 {
		t.Errorf("Expected size 0 after clear, got %d", cache.Size())
	}
}

// TestMemoryCache_Stats 测试统计
func TestMemoryCache_Stats(t *testing.T) {
	cfg := MemoryCacheConfig{
		MaxSize:         100,
		TTL:             5 * time.Minute,
		CleanupInterval: 1 * time.Minute,
		Enabled:         true,
	}

	cache := NewMemoryCacheWithConfig[string](cfg)
	defer cache.Close()

	cache.Set("key1", "value1")
	cache.Get("key1")        // hit
	cache.Get("key1")        // hit
	cache.Get("nonexistent") // miss

	stats := cache.GetStatsStruct()

	if stats.Hits != 2 {
		t.Errorf("Expected 2 hits, got %d", stats.Hits)
	}
	if stats.Misses != 1 {
		t.Errorf("Expected 1 miss, got %d", stats.Misses)
	}
	if stats.Size != 1 {
		t.Errorf("Expected size 1, got %d", stats.Size)
	}
}
