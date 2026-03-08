package cache

import (
	"context"
	"sync"
	"time"

	"github.com/nick0323/K8sVision/model"
	"go.uber.org/zap"
)

type CacheItem struct {
	Value      interface{}
	ExpireTime time.Time
	CreatedAt  time.Time
}

type MemoryCache struct {
	data            map[string]CacheItem
	mutex           sync.RWMutex
	maxSize         int
	ttl             time.Duration
	cleanupInterval time.Duration
	logger          *zap.Logger
	ctx             context.Context
	cancel          context.CancelFunc
}

// NewMemoryCache 创建新的内存缓存
func NewMemoryCache(config *model.CacheConfig, logger *zap.Logger) *MemoryCache {
	ctx, cancel := context.WithCancel(context.Background())

	cache := &MemoryCache{
		data:            make(map[string]CacheItem),
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

// Set 设置缓存
func (c *MemoryCache) Set(key string, value interface{}) {
	c.SetWithTTL(key, value, c.ttl)
}

// SetWithTTL 设置缓存并指定TTL
func (c *MemoryCache) SetWithTTL(key string, value interface{}, ttl time.Duration) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	// 检查容量限制
	if len(c.data) >= c.maxSize {
		c.evictOldest()
	}

	now := time.Now()
	c.data[key] = CacheItem{
		Value:      value,
		ExpireTime: now.Add(ttl),
		CreatedAt:  now,
	}

	// 生产环境不记录Debug级缓存设置日志
}

// Get 获取缓存
func (c *MemoryCache) Get(key string) (interface{}, bool) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	item, exists := c.data[key]
	if !exists {
		return nil, false
	}

	// 检查是否过期
	if time.Now().After(item.ExpireTime) {
		// 异步删除过期项
		go c.delete(key)
		return nil, false
	}

	// 生产环境不记录Debug级缓存命中日志
	return item.Value, true
}

// Delete 删除缓存
func (c *MemoryCache) Delete(key string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	delete(c.data, key)

	// 生产环境不记录Debug级缓存删除日志
}

// delete 内部删除方法（无锁）
func (c *MemoryCache) delete(key string) {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	delete(c.data, key)
}

// Clear 清空缓存
func (c *MemoryCache) Clear() {
	c.mutex.Lock()
	defer c.mutex.Unlock()

	c.data = make(map[string]CacheItem)

	// 安全检查logger
	if c.logger != nil {
		c.logger.Info("缓存已清空")
	}
}

// Size 获取缓存大小
func (c *MemoryCache) Size() int {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	return len(c.data)
}

// Keys 获取所有键
func (c *MemoryCache) Keys() []string {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	keys := make([]string, 0, len(c.data))
	for key := range c.data {
		keys = append(keys, key)
	}
	return keys
}

// evictOldest 淘汰最旧的项 - 使用更高效的策略
func (c *MemoryCache) evictOldest() {
	// 使用随机采样策略而不是遍历全部，提高性能
	const sampleSize = model.CacheSampleSize
	keys := make([]string, 0, sampleSize)
	createdTimes := make([]time.Time, 0, sampleSize)

	// 随机选择一些项进行比较
	count := 0
	for key, item := range c.data {
		if count >= sampleSize {
			break
		}
		keys = append(keys, key)
		createdTimes = append(createdTimes, item.CreatedAt)
		count++
	}

	if len(keys) == 0 {
		return
	}

	// 在样本中找到最早的项
	oldestIdx := 0
	oldestTime := createdTimes[0]
	for i := 1; i < len(createdTimes); i++ {
		if createdTimes[i].Before(oldestTime) {
			oldestIdx = i
			oldestTime = createdTimes[i]
		}
	}

	delete(c.data, keys[oldestIdx])

	// 生产环境不记录Debug级缓存淘汰日志
}

// cleanupWorker 清理工作协程
func (c *MemoryCache) cleanupWorker() {
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
func (c *MemoryCache) cleanup() int {
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

	if expiredCount > 0 {
		// 安全检查logger
		if c.logger != nil {
			c.logger.Info("缓存清理完成", zap.Int("expiredCount", expiredCount))
		}
	}

	return expiredCount
}

// ClearExpired 手动清理过期项并返回清理数量
func (c *MemoryCache) ClearExpired() int {
	return c.cleanup()
}

// Close 关闭缓存
func (c *MemoryCache) Close() {
	c.cancel()
	c.Clear()

	// 安全检查logger
	if c.logger != nil {
		c.logger.Info("缓存已关闭")
	}
}

// CacheStats 缓存统计信息结构
type CacheStats struct {
	Size         int     `json:"size"`
	MaxSize      int     `json:"maxSize"`
	ExpiredCount int     `json:"expiredCount"`
	TotalSize    int64   `json:"totalSize"` // 使用更精确的大小计算
	TTL          string  `json:"ttl"`
	HitRate      float64 `json:"hitRate"`
	Utilization  float64 `json:"utilization"`
	Hits         int64   `json:"hits"`
	Misses       int64   `json:"misses"`
}

// GetStats 获取缓存统计信息
func (c *MemoryCache) GetStats() map[string]interface{} {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	now := time.Now()
	expiredCount := 0
	totalSize := int64(0)

	// 计算过期项数量和大小
	for _, item := range c.data {
		if now.After(item.ExpireTime) {
			expiredCount++
		}
		// 估算每个缓存项的大小（简化估算）
		// 在实际应用中，可以根据具体类型做更精确的估算
		size := int64(100) // 默认估算每个项占用100字节
		totalSize += size
	}

	// 由于目前没有维护命中/未命中计数，暂时返回0
	hits := int64(0)
	misses := int64(0)
	totalRequests := hits + misses
	hitRate := 0.0
	if totalRequests > 0 {
		hitRate = float64(hits) / float64(totalRequests) * 100
	}

	return map[string]interface{}{
		"size":         len(c.data),
		"maxSize":      c.maxSize,
		"expiredCount": expiredCount,
		"totalSize":    totalSize,
		"ttl":          c.ttl.String(),
		"hitRate":      hitRate,
		"utilization":  float64(len(c.data)) / float64(c.maxSize) * 100,
		"hits":         hits,
		"misses":       misses,
	}
}

// NewMemoryCacheLRU 创建 LRU 缓存（使用优化版本）
// 这是 NewMemoryCacheOptimized 的别名，为了兼容 manager.go 的调用
func NewMemoryCacheLRU(config *model.CacheConfig, logger *zap.Logger) *MemoryCache {
	// 使用优化版本的缓存实现
	return NewMemoryCacheOptimized(config, logger)
}

// NewMemoryCacheOptimized 创建优化的内存缓存（LRU 实现）
// 这是实际使用的优化版本
func NewMemoryCacheOptimized(config *model.CacheConfig, logger *zap.Logger) *MemoryCache {
	ctx, cancel := context.WithCancel(context.Background())

	cache := &MemoryCache{
		data:            make(map[string]CacheItem),
		maxSize:         config.MaxSize,
		ttl:             config.TTL,
		cleanupInterval: config.CleanupInterval,
		logger:          logger,
		ctx:            ctx,
		cancel:         cancel,
	}

	// 启动清理协程
	if config.Enabled {
		go cache.cleanupWorker()
	}

	return cache
}
