package cache

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/nick0323/K8sVision/model"
)

const (
	DefaultMaxSize         = 1000
	DefaultTTL             = 5 * time.Minute
	DefaultCleanupInterval = 1 * time.Minute
)

type CacheItem[T any] struct {
	Value      T
	ExpireTime time.Time
}

type MemoryCache[T any] struct {
	data    map[string]*CacheItem[T]
	mutex   sync.RWMutex
	maxSize int
	ttl     time.Duration
	hits    atomic.Int64
	misses  atomic.Int64
	ctx     context.Context
	cancel  context.CancelFunc
}

func NewMemoryCache(config *model.CacheConfig, logger interface{}) *MemoryCache[interface{}] {
	if config == nil {
		config = &model.CacheConfig{
			Enabled:         true,
			MaxSize:         DefaultMaxSize,
			TTL:             DefaultTTL,
			CleanupInterval: DefaultCleanupInterval,
		}
	}

	ctx, cancel := context.WithCancel(context.Background())
	cache := &MemoryCache[interface{}]{
		data:    make(map[string]*CacheItem[interface{}], config.MaxSize),
		maxSize: config.MaxSize,
		ttl:     config.TTL,
		ctx:     ctx,
		cancel:  cancel,
	}

	if config.Enabled {
		go cache.cleanupWorker(config.CleanupInterval)
	}

	return cache
}

func (c *MemoryCache[T]) Set(key string, value T) {
	c.SetWithTTL(key, value, c.ttl)
}

func (c *MemoryCache[T]) SetWithTTL(key string, value T, ttl time.Duration) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	if len(c.data) >= c.maxSize {
		c.evictOldest()
	}

	c.data[key] = &CacheItem[T]{
		Value:      value,
		ExpireTime: time.Now().Add(ttl),
	}
}

func (c *MemoryCache[T]) Get(key string) (T, bool) {
	var zero T

	c.mutex.RLock()
	item, exists := c.data[key]
	c.mutex.RUnlock()

	if !exists {
		c.misses.Add(1)
		return zero, false
	}

	if time.Now().After(item.ExpireTime) {
		c.mutex.Lock()
		delete(c.data, key)
		c.mutex.Unlock()
		c.misses.Add(1)
		return zero, false
	}

	c.hits.Add(1)
	return item.Value, true
}

func (c *MemoryCache[T]) Delete(key string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	delete(c.data, key)
}

func (c *MemoryCache[T]) Clear() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.data = make(map[string]*CacheItem[T])
}

func (c *MemoryCache[T]) Size() int {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return len(c.data)
}

func (c *MemoryCache[T]) evictOldest() {
	if len(c.data) == 0 {
		return
	}

	var oldestKey string
	oldestTime := time.Now()

	for key, item := range c.data {
		if item.ExpireTime.Before(oldestTime) {
			oldestKey = key
			oldestTime = item.ExpireTime
		}
	}

	if oldestKey != "" {
		delete(c.data, oldestKey)
	}
}

func (c *MemoryCache[T]) cleanupWorker(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			c.cleanup()
		case <-c.ctx.Done():
			return
		}
	}
}

func (c *MemoryCache[T]) cleanup() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	now := time.Now()
	for key, item := range c.data {
		if now.After(item.ExpireTime) {
			delete(c.data, key)
		}
	}
}

func (c *MemoryCache[T]) Close() {
	c.cancel()
	c.Clear()
}

func (c *MemoryCache[T]) GetStats() map[string]interface{} {
	c.mutex.RLock()
	size := len(c.data)
	c.mutex.RUnlock()

	hits := c.hits.Load()
	misses := c.misses.Load()
	total := hits + misses

	hitRate := float64(0)
	if total > 0 {
		hitRate = float64(hits) / float64(total) * 100
	}

	return map[string]interface{}{
		"size":    size,
		"maxSize": c.maxSize,
		"hitRate": hitRate,
		"hits":    hits,
		"misses":  misses,
	}
}
