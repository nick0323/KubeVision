package cache

import (
	"container/list"
	"context"
	"hash/fnv"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/nick0323/K8sVision/model"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	cacheSizeGauge = promauto.NewGauge(prometheus.GaugeOpts{
		Name: "kubevision_cache_size",
		Help: "Current cache size",
	})

	cacheHitsCounter = promauto.NewCounter(prometheus.CounterOpts{
		Name: "kubevision_cache_hits_total",
		Help: "Total number of cache hits",
	})

	cacheMissesCounter = promauto.NewCounter(prometheus.CounterOpts{
		Name: "kubevision_cache_misses_total",
		Help: "Total number of cache misses",
	})
)

const (
	DefaultMaxSize         = 1000
	DefaultTTL             = 5 * time.Minute
	DefaultCleanupInterval = 1 * time.Minute
	DefaultShardCount      = 16
)

type CacheItem[T any] struct {
	Value      T
	ExpireTime time.Time
}

type cacheEntry[T any] struct {
	key  string
	item *CacheItem[T]
}

type cacheShard[T any] struct {
	data    map[string]*list.Element
	lruList *list.List
	mutex   sync.RWMutex
}

type MemoryCache[T any] struct {
	shards  []*cacheShard[T]
	shardMask uint32
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
			ShardCount:      DefaultShardCount,
		}
	}

	shardCount := config.ShardCount
	if shardCount <= 0 {
		shardCount = DefaultShardCount
	}
	if shardCount&(shardCount-1) != 0 {
		shardCount = DefaultShardCount
	}

	shards := make([]*cacheShard[interface{}], shardCount)
	for i := range shards {
		shards[i] = &cacheShard[interface{}]{
			data:    make(map[string]*list.Element, config.MaxSize/shardCount+1),
			lruList: list.New(),
		}
	}

	ctx, cancel := context.WithCancel(context.Background())
	cache := &MemoryCache[interface{}]{
		shards:    shards,
		shardMask: uint32(shardCount - 1),
		maxSize:   config.MaxSize,
		ttl:       config.TTL,
		ctx:       ctx,
		cancel:    cancel,
	}

	if config.Enabled {
		go cache.cleanupWorker(config.CleanupInterval)
		go cache.metricsReporter(5 * time.Minute)
	}

	return cache
}

func (c *MemoryCache[T]) getShard(key string) *cacheShard[T] {
	h := fnv.New32a()
	h.Write([]byte(key))
	return c.shards[h.Sum32()&c.shardMask]
}

func (c *MemoryCache[T]) Set(key string, value T) {
	c.SetWithTTL(key, value, c.ttl)
}

func (c *MemoryCache[T]) SetWithTTL(key string, value T, ttl time.Duration) {
	s := c.getShard(key)
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if elem, ok := s.data[key]; ok {
		s.lruList.Remove(elem)
		delete(s.data, key)
	}

	for s.lruList.Len() >= c.maxSize/c.lenShards() {
		s.evictLRU()
	}

	entry := &cacheEntry[T]{
		key: key,
		item: &CacheItem[T]{
			Value:      value,
			ExpireTime: time.Now().Add(ttl),
		},
	}
	elem := s.lruList.PushFront(entry)
	s.data[key] = elem
}

func (c *MemoryCache[T]) lenShards() int {
	return len(c.shards)
}

func (c *MemoryCache[T]) Get(key string) (T, bool) {
	var zero T

	s := c.getShard(key)

	s.mutex.RLock()
	elem, exists := s.data[key]
	if !exists {
		s.mutex.RUnlock()
		c.misses.Add(1)
		cacheMissesCounter.Inc()
		return zero, false
	}

	entry := elem.Value.(*cacheEntry[T])

	if time.Now().After(entry.item.ExpireTime) {
		s.mutex.RUnlock()
		s.mutex.Lock()
		if elem2, ok := s.data[key]; ok {
			entry2 := elem2.Value.(*cacheEntry[T])
			if time.Now().After(entry2.item.ExpireTime) {
				s.lruList.Remove(elem2)
				delete(s.data, key)
			}
		}
		s.mutex.Unlock()
		c.misses.Add(1)
		cacheMissesCounter.Inc()
		return zero, false
	}

	s.mutex.RUnlock()
	s.mutex.Lock()
	if elem2, ok := s.data[key]; ok {
		s.lruList.MoveToFront(elem2)
	}
	s.mutex.Unlock()

	c.hits.Add(1)
	cacheHitsCounter.Inc()
	return entry.item.Value, true
}

func (c *MemoryCache[T]) Delete(key string) {
	s := c.getShard(key)
	s.mutex.Lock()
	defer s.mutex.Unlock()
	if elem, ok := s.data[key]; ok {
		s.lruList.Remove(elem)
		delete(s.data, key)
	}
}

func (c *MemoryCache[T]) DeleteByPrefix(prefix string) {
	var wg sync.WaitGroup
	for _, s := range c.shards {
		wg.Add(1)
		go func(shard *cacheShard[T]) {
			defer wg.Done()
			shard.mutex.Lock()
			for key, elem := range shard.data {
				if strings.HasPrefix(key, prefix) {
					shard.lruList.Remove(elem)
					delete(shard.data, key)
				}
			}
			shard.mutex.Unlock()
		}(s)
	}
	wg.Wait()
}

func (c *MemoryCache[T]) Clear() {
	for _, s := range c.shards {
		s.mutex.Lock()
		s.data = make(map[string]*list.Element)
		s.lruList = list.New()
		s.mutex.Unlock()
	}
}

func (c *MemoryCache[T]) Size() int {
	total := 0
	for _, s := range c.shards {
		s.mutex.RLock()
		total += s.lruList.Len()
		s.mutex.RUnlock()
	}
	return total
}

func (s *cacheShard[T]) evictLRU() {
	elem := s.lruList.Back()
	if elem == nil {
		return
	}
	entry := elem.Value.(*cacheEntry[T])
	delete(s.data, entry.key)
	s.lruList.Remove(elem)
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
	for _, s := range c.shards {
		s.mutex.Lock()
		now := time.Now()
		var next *list.Element
		for e := s.lruList.Front(); e != nil; e = next {
			next = e.Next()
			entry := e.Value.(*cacheEntry[T])
			if now.After(entry.item.ExpireTime) {
				delete(s.data, entry.key)
				s.lruList.Remove(e)
			}
		}
		s.mutex.Unlock()
	}
}

func (c *MemoryCache[T]) metricsReporter(interval time.Duration) {
	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			cacheSizeGauge.Set(float64(c.Size()))
		case <-c.ctx.Done():
			return
		}
	}
}

func (c *MemoryCache[T]) Close() {
	c.cancel()
	c.Clear()
}

func (c *MemoryCache[T]) GetStats() map[string]interface{} {
	size := c.Size()

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
		"shards":  len(c.shards),
		"hitRate": hitRate,
		"hits":    hits,
		"misses":  misses,
	}
}
