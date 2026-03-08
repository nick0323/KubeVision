package cache

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/nick0323/K8sVision/model"
	"go.uber.org/zap"
)

type Manager struct {
	caches map[string]Cache
	mutex  sync.RWMutex
	logger *zap.Logger
	ctx    context.Context
	cancel context.CancelFunc
}

type Cache interface {
	Set(key string, value interface{})
	SetWithTTL(key string, value interface{}, ttl time.Duration)
	Get(key string) (interface{}, bool)
	Delete(key string)
	Clear()
	Size() int
	Keys() []string
	Close()
	GetStats() map[string]interface{}
}

func NewManager(config *model.CacheConfig, logger *zap.Logger) *Manager {
	ctx, cancel := context.WithCancel(context.Background())

	manager := &Manager{
		caches: make(map[string]Cache),
		logger: logger,
		ctx:    ctx,
		cancel: cancel,
	}

	// 初始化默认缓存（使用优化的 LRU 缓存）
	if config.Enabled {
		switch config.Type {
		case "memory", "lru":
			manager.caches["default"] = NewMemoryCacheLRU(config, logger)
		default:
			logger.Warn("不支持的缓存类型，使用 LRU 缓存", zap.String("type", config.Type))
			manager.caches["default"] = NewMemoryCacheLRU(config, logger)
		}
	}

	return manager
}

// GetCache 获取指定名称的缓存
func (m *Manager) GetCache(name string) (Cache, bool) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	cache, exists := m.caches[name]
	return cache, exists
}

// GetDefaultCache 获取默认缓存
func (m *Manager) GetDefaultCache() (Cache, bool) {
	return m.GetCache("default")
}

// CreateCache 创建新的缓存
func (m *Manager) CreateCache(name string, config *model.CacheConfig) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	if _, exists := m.caches[name]; exists {
		return fmt.Errorf("缓存已存在: %s", name)
	}

	var cache Cache
	switch config.Type {
	case "memory":
		cache = NewMemoryCache(config, m.logger)
	default:
		return fmt.Errorf("不支持的缓存类型: %s", config.Type)
	}

	m.caches[name] = cache
	m.logger.Info("缓存创建成功", zap.String("name", name), zap.String("type", config.Type))
	return nil
}

// DeleteCache 删除缓存
func (m *Manager) DeleteCache(name string) error {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	cache, exists := m.caches[name]
	if !exists {
		return fmt.Errorf("缓存不存在: %s", name)
	}

	cache.Close()
	delete(m.caches, name)
	m.logger.Info("缓存删除成功", zap.String("name", name))
	return nil
}

// Set 设置默认缓存
func (m *Manager) Set(key string, value interface{}) {
	if cache, exists := m.GetDefaultCache(); exists {
		cache.Set(key, value)
	}
}

// SetWithTTL 设置默认缓存并指定TTL
func (m *Manager) SetWithTTL(key string, value interface{}, ttl time.Duration) {
	if cache, exists := m.GetDefaultCache(); exists {
		cache.SetWithTTL(key, value, ttl)
	}
}

// Get 从默认缓存获取
func (m *Manager) Get(key string) (interface{}, bool) {
	if cache, exists := m.GetDefaultCache(); exists {
		return cache.Get(key)
	}
	return nil, false
}

// Delete 从默认缓存删除
func (m *Manager) Delete(key string) {
	if cache, exists := m.GetDefaultCache(); exists {
		cache.Delete(key)
	}
}

// Clear 清空默认缓存
func (m *Manager) Clear() {
	if cache, exists := m.GetDefaultCache(); exists {
		cache.Clear()
	}
}

// GetOrSet 获取缓存，如果不存在则设置
func (m *Manager) GetOrSet(key string, ttl time.Duration, loader func() (interface{}, error)) (interface{}, error) {
	// 尝试从缓存获取
	if value, exists := m.Get(key); exists {
		return value, nil
	}

	// 缓存未命中，加载数据
	value, err := loader()
	if err != nil {
		return nil, err
	}

	// 设置缓存
	m.SetWithTTL(key, value, ttl)
	return value, nil
}

// GetOrSetWithCache 从指定缓存获取或设置
func (m *Manager) GetOrSetWithCache(cacheName, key string, ttl time.Duration, loader func() (interface{}, error)) (interface{}, error) {
	cache, exists := m.GetCache(cacheName)
	if !exists {
		return nil, fmt.Errorf("缓存不存在: %s", cacheName)
	}

	// 尝试从缓存获取
	if value, exists := cache.Get(key); exists {
		return value, nil
	}

	// 缓存未命中，加载数据
	value, err := loader()
	if err != nil {
		return nil, err
	}

	// 设置缓存
	cache.SetWithTTL(key, value, ttl)
	return value, nil
}

// GetAllStats 获取所有缓存的统计信息
func (m *Manager) GetAllStats() map[string]interface{} {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	stats := make(map[string]interface{})
	for name, cache := range m.caches {
		stats[name] = cache.GetStats()
	}

	return stats
}

// Close 关闭所有缓存
func (m *Manager) Close() {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	for name, cache := range m.caches {
		cache.Close()
		m.logger.Info("缓存已关闭", zap.String("name", name))
	}

	m.caches = make(map[string]Cache)
	m.cancel()
}

// ListCaches 列出所有缓存名称
func (m *Manager) ListCaches() []string {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	names := make([]string, 0, len(m.caches))
	for name := range m.caches {
		names = append(names, name)
	}
	return names
}

// IsEnabled 检查缓存是否启用
func (m *Manager) IsEnabled() bool {
	_, exists := m.GetDefaultCache()
	return exists
}

// GetLogger 获取logger
func (m *Manager) GetLogger() *zap.Logger {
	return m.logger
}

// WarmupCache 缓存预热
func (m *Manager) WarmupCache(loader func() (map[string]interface{}, error)) error {
	if !m.IsEnabled() {
		return nil
	}

	cache, exists := m.GetDefaultCache()
	if !exists {
		return fmt.Errorf("默认缓存不存在")
	}

	// 加载数据
	data, err := loader()
	if err != nil {
		return fmt.Errorf("加载预热数据失败: %w", err)
	}

	// 批量设置缓存
	for key, value := range data {
		cache.Set(key, value)
	}

	m.logger.Info("缓存预热完成", zap.Int("count", len(data)))
	return nil
}

// GetCacheStats 获取缓存统计信息
func (m *Manager) GetCacheStats() map[string]interface{} {
	stats := make(map[string]interface{})

	m.mutex.RLock()
	defer m.mutex.RUnlock()

	for name, cache := range m.caches {
		stats[name] = cache.GetStats()
	}

	return stats
}

// ClearExpired 清理过期项
func (m *Manager) ClearExpired() {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	var totalCleaned int
	for _, cache := range m.caches {
		// 对于MemoryCache，我们可以手动触发清理
		if memCache, ok := cache.(*MemoryCache); ok {
			cleaned := memCache.ClearExpired()
			totalCleaned += cleaned
			// 移除Debug日志，避免噪声
		}
	}

	if totalCleaned > 0 {
		m.logger.Info("清理过期缓存项完成", zap.Int("total", totalCleaned))
	}
}
