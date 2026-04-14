package cache

import (
	"context"
	"sync"
	"sync/atomic"
	"time"

	"github.com/nick0323/K8sVision/model"
	"go.uber.org/zap"
)

// 缓存配置常量
const (
	DefaultMaxSize         = 1000
	DefaultTTL             = 5 * time.Minute
	DefaultCleanupInterval = 1 * time.Minute
)

// CacheItem 缓存项结构
type CacheItem[T any] struct {
	Value      T
	ExpireTime time.Time
	CreatedAt  time.Time
	LastAccess time.Time
}

// CacheStats 缓存统计信息
type CacheStats struct {
	Size          int     `json:"size"`
	MaxSize       int     `json:"maxSize"`
	ExpiredCount  int     `json:"expired_count"`
	TTL           string  `json:"ttl"`
	HitRate       float64 `json:"hit_rate"`
	Utilization   float64 `json:"utilization"`
	Hits          int64   `json:"hits"`
	Misses        int64   `json:"misses"`
	TotalRequests int64   `json:"total_requests"`
}

// MemoryCache 泛型内存缓存
type MemoryCache[T any] struct {
	data            map[string]CacheItem[T]
	mutex           sync.RWMutex
	maxSize         int
	ttl             time.Duration
	cleanupInterval time.Duration
	logger          *zap.Logger
	ctx             context.Context
	cancel          context.CancelFunc
	hits            atomic.Int64
	misses          atomic.Int64
}

// MemoryCacheConfig 缓存配置
type MemoryCacheConfig struct {
	MaxSize         int
	TTL             time.Duration
	CleanupInterval time.Duration
	Enabled         bool
	Logger          *zap.Logger
}

// NewMemoryCache 创建新的内存缓存
func NewMemoryCache(config *model.CacheConfig, logger *zap.Logger) *MemoryCache[interface{}] {
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
		data:            make(map[string]CacheItem[interface{}]),
		maxSize:         config.MaxSize,
		ttl:             config.TTL,
		cleanupInterval: config.CleanupInterval,
		logger:          logger,
		ctx:             ctx,
		cancel:          cancel,
	}

	// 启动清理协程
	if config.Enabled {
		go cache.cleanupWorker()
	}

	return cache
}

// NewMemoryCacheWithConfig 使用配置结构体创建缓存
func NewMemoryCacheWithConfig[T any](cfg MemoryCacheConfig) *MemoryCache[T] {
	if cfg.MaxSize <= 0 {
		cfg.MaxSize = DefaultMaxSize
	}
	if cfg.TTL <= 0 {
		cfg.TTL = DefaultTTL
	}
	if cfg.CleanupInterval <= 0 {
		cfg.CleanupInterval = DefaultCleanupInterval
	}

	ctx, cancel := context.WithCancel(context.Background())

	cache := &MemoryCache[T]{
		data:            make(map[string]CacheItem[T]),
		maxSize:         cfg.MaxSize,
		ttl:             cfg.TTL,
		cleanupInterval: cfg.CleanupInterval,
		logger:          cfg.Logger,
		ctx:             ctx,
		cancel:          cancel,
	}

	if cfg.Enabled {
		go cache.cleanupWorker()
	}

	return cache
}

// Set 设置缓存
func (c *MemoryCache[T]) Set(key string, value T) {
	c.SetWithTTL(key, value, c.ttl)
}

// SetWithTTL 设置缓存并指定 TTL
func (c *MemoryCache[T]) SetWithTTL(key string, value T, ttl time.Duration) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// 检查容量限制
	if len(c.data) >= c.maxSize {
		c.evictOldest()
	}

	now := time.Now()
	c.data[key] = CacheItem[T]{
		Value:      value,
		ExpireTime: now.Add(ttl),
		CreatedAt:  now,
		LastAccess: now,
	}
}

// Get 获取缓存
func (c *MemoryCache[T]) Get(key string) (T, bool) {
	var zero T

	c.mutex.RLock()
	defer c.mutex.RUnlock()

	item, exists := c.data[key]
	if !exists {
		c.misses.Add(1)
		return zero, false
	}

	// 检查是否过期
	if time.Now().After(item.ExpireTime) {
		c.misses.Add(1)
		// 异步删除过期项（限制并发数）
		go func() {
			select {
			case <-c.ctx.Done():
				return
			default:
				c.delete(key)
			}
		}()
		return zero, false
	}

	// 更新访问时间（用于 LRU）
	c.mutex.RUnlock()
	c.mutex.Lock()
	if item, exists := c.data[key]; exists {
		item.LastAccess = time.Now()
		c.data[key] = item
	}
	c.mutex.Unlock()
	c.mutex.RLock()

	c.hits.Add(1)
	return item.Value, true
}

// GetOrSet 获取或设置缓存
func (c *MemoryCache[T]) GetOrSet(key string, valueFunc func() (T, error)) (T, error) {
	// 先尝试获取
	if value, found := c.Get(key); found {
		return value, nil
	}

	// 获取失败，设置新值
	value, err := valueFunc()
	if err != nil {
		var zero T
		return zero, err
	}

	c.Set(key, value)
	return value, nil
}

// Delete 删除缓存
func (c *MemoryCache[T]) Delete(key string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	delete(c.data, key)
}

// delete 内部删除方法（无锁）
func (c *MemoryCache[T]) delete(key string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	delete(c.data, key)
}

// Clear 清空缓存
func (c *MemoryCache[T]) Clear() {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.data = make(map[string]CacheItem[T])
	if c.logger != nil {
		c.logger.Info("缓存已清空")
	}
}

// Size 获取缓存大小
func (c *MemoryCache[T]) Size() int {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return len(c.data)
}

// Keys 获取所有键
func (c *MemoryCache[T]) Keys() []string {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	keys := make([]string, 0, len(c.data))
	for key := range c.data {
		keys = append(keys, key)
	}
	return keys
}

// evictOldest LRU 淘汰：淘汰最近最少使用的项
func (c *MemoryCache[T]) evictOldest() {
	if len(c.data) == 0 {
		return
	}

	var oldestKey string
	var oldestTime time.Time

	// 找到最久未访问的项
	for key, item := range c.data {
		if oldestKey == "" || item.LastAccess.Before(oldestTime) {
			oldestKey = key
			oldestTime = item.LastAccess
		}
	}

	if oldestKey != "" {
		delete(c.data, oldestKey)
	}
}

// cleanupWorker 清理工作协程
func (c *MemoryCache[T]) cleanupWorker() {
	ticker := time.NewTicker(c.cleanupInterval)
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

// cleanup 清理过期项
func (c *MemoryCache[T]) cleanup() int {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	now := time.Now()
	expiredCount := 0

	for key, item := range c.data {
		if now.After(item.ExpireTime) {
			delete(c.data, key)
			expiredCount++
		}
	}

	if expiredCount > 0 && c.logger != nil {
		c.logger.Debug("缓存清理完成", zap.Int("expiredCount", expiredCount))
	}

	return expiredCount
}

// ClearExpired 手动清理过期项
func (c *MemoryCache[T]) ClearExpired() int {
	return c.cleanup()
}

// Close 关闭缓存
func (c *MemoryCache[T]) Close() {
	c.cancel()
	c.Clear()
	if c.logger != nil {
		c.logger.Info("缓存已关闭")
	}
}

// GetStats 获取缓存统计信息（返回 map 格式，实现 Cache 接口）
func (c *MemoryCache[T]) GetStats() map[string]interface{} {
	stats := c.GetStatsStruct()
	return map[string]interface{}{
		"size":          stats.Size,
		"maxSize":       stats.MaxSize,
		"expiredCount":  stats.ExpiredCount,
		"ttl":           stats.TTL,
		"hitRate":       stats.HitRate,
		"utilization":   stats.Utilization,
		"hits":          stats.Hits,
		"misses":        stats.Misses,
		"totalRequests": stats.TotalRequests,
	}
}

// GetStatsStruct 获取缓存统计信息（结构体格式）
func (c *MemoryCache[T]) GetStatsStruct() CacheStats {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	hits := c.hits.Load()
	misses := c.misses.Load()
	totalRequests := hits + misses

	hitRate := 0.0
	if totalRequests > 0 {
		hitRate = float64(hits) / float64(totalRequests) * 100
	}

	utilization := 0.0
	if c.maxSize > 0 {
		utilization = float64(len(c.data)) / float64(c.maxSize) * 100
	}

	// 计算过期项数量
	now := time.Now()
	expiredCount := 0
	for _, item := range c.data {
		if now.After(item.ExpireTime) {
			expiredCount++
		}
	}

	return CacheStats{
		Size:          len(c.data),
		MaxSize:       c.maxSize,
		ExpiredCount:  expiredCount,
		TTL:           c.ttl.String(),
		HitRate:       hitRate,
		Utilization:   utilization,
		Hits:          hits,
		Misses:        misses,
		TotalRequests: totalRequests,
	}
}
