package cache

import (
	"container/list"
	"context"
	"strings"
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

type cacheEntry[T any] struct {
	key  string
	item *CacheItem[T]
}

type MemoryCache[T any] struct {
	data    map[string]*list.Element
	lruList *list.List
	mutex   sync.RWMutex
	maxSize int
	ttl     time.Duration
	hits    atomic.Int64
	misses  atomic.Int64
	ctx     context.Context
	cancel  context.CancelFunc
}

func NewMemoryCache(config *model.CacheConfig) *MemoryCache[interface{}] {
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
		data:    make(map[string]*list.Element, config.MaxSize),
		lruList: list.New(),
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

	if elem, ok := c.data[key]; ok {
		c.lruList.Remove(elem)
		delete(c.data, key)
	}

	for c.lruList.Len() >= c.maxSize {
		c.evictLRU()
	}

	entry := &cacheEntry[T]{
		key: key,
		item: &CacheItem[T]{
			Value:      value,
			ExpireTime: time.Now().Add(ttl),
		},
	}
	elem := c.lruList.PushFront(entry)
	c.data[key] = elem
}

func (c *MemoryCache[T]) Get(key string) (T, bool) {
	var zero T

	c.mutex.Lock()
	elem, exists := c.data[key]
	if !exists {
		c.mutex.Unlock()
		c.misses.Add(1)
		return zero, false
	}

	entry := elem.Value.(*cacheEntry[T])

	if time.Now().After(entry.item.ExpireTime) {
		c.lruList.Remove(elem)
		delete(c.data, key)
		c.mutex.Unlock()
		c.misses.Add(1)
		return zero, false
	}

	c.lruList.MoveToFront(elem)
	c.mutex.Unlock()

	c.hits.Add(1)
	return entry.item.Value, true
}

func (c *MemoryCache[T]) Delete(key string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	if elem, ok := c.data[key]; ok {
		c.lruList.Remove(elem)
		delete(c.data, key)
	}
}

func (c *MemoryCache[T]) DeleteByPrefix(prefix string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	for key, elem := range c.data {
		if strings.HasPrefix(key, prefix) {
			c.lruList.Remove(elem)
			delete(c.data, key)
		}
	}
}

func (c *MemoryCache[T]) Clear() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.data = make(map[string]*list.Element)
	c.lruList = list.New()
}

func (c *MemoryCache[T]) Size() int {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.lruList.Len()
}

func (c *MemoryCache[T]) evictLRU() {
	elem := c.lruList.Back()
	if elem == nil {
		return
	}
	entry := elem.Value.(*cacheEntry[T])
	delete(c.data, entry.key)
	c.lruList.Remove(elem)
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
	var next *list.Element
	for e := c.lruList.Front(); e != nil; e = next {
		next = e.Next()
		entry := e.Value.(*cacheEntry[T])
		if now.After(entry.item.ExpireTime) {
			delete(c.data, entry.key)
			c.lruList.Remove(e)
		}
	}
}

func (c *MemoryCache[T]) Close() {
	c.cancel()
	c.Clear()
}

func (c *MemoryCache[T]) GetStats() map[string]interface{} {
	c.mutex.RLock()
	size := c.lruList.Len()
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
