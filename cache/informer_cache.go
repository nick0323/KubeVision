package cache

import (
	"context"
	"sync"
	"time"

	"go.uber.org/zap"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

// ResourceType 资源类型
type ResourceType string

const (
	ResourcePods       ResourceType = "pods"
	ResourceNodes      ResourceType = "nodes"
	ResourceNamespaces ResourceType = "namespaces"
)

// ResourceCache 资源缓存接口
type ResourceCache interface {
	List() []interface{}
	Get(name string) (interface{}, bool)
	Count() int
	IsSynced() bool
	Close()
	Start()
}

// InformerCache Informer 缓存
type InformerCache struct {
	informer cache.SharedIndexInformer
	stopCh   chan struct{}
	logger   *zap.Logger
}

// NewInformerCache 使用 SharedInformerFactory 创建 Informer 缓存
func NewInformerCache(
	ctx context.Context,
	clientset *kubernetes.Clientset,
	resourceType ResourceType,
	namespace string,
	logger *zap.Logger,
) (*InformerCache, error) {
	// 创建 SharedInformerFactory
	factory := informers.NewSharedInformerFactory(clientset, 0)

	var informer cache.SharedIndexInformer

	// 根据资源类型获取不同的 Informer
	switch resourceType {
	case ResourcePods:
		informer = factory.Core().V1().Pods().Informer()
	case ResourceNodes:
		informer = factory.Core().V1().Nodes().Informer()
	case ResourceNamespaces:
		informer = factory.Core().V1().Namespaces().Informer()
	default:
		return nil, nil
	}

	// 添加事件处理器（可选，用于调试）
	informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			logger.Debug("Resource added", zap.String("type", string(resourceType)))
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			logger.Debug("Resource updated", zap.String("type", string(resourceType)))
		},
		DeleteFunc: func(obj interface{}) {
			logger.Debug("Resource deleted", zap.String("type", string(resourceType)))
		},
	})

	informerCache := &InformerCache{
		informer: informer,
		stopCh:   make(chan struct{}),
		logger:   logger,
	}

	return informerCache, nil
}

// Start 启动 Informer
func (c *InformerCache) Start() {
	go c.informer.Run(c.stopCh)
	c.logger.Info("Informer cache started")
}

// List 列出所有资源
func (c *InformerCache) List() []interface{} {
	return c.informer.GetStore().List()
}

// Get 获取单个资源
func (c *InformerCache) Get(name string) (interface{}, bool) {
	item, exists, err := c.informer.GetStore().GetByKey(name)
	return item, exists && err == nil
}

// Count 获取资源数量
func (c *InformerCache) Count() int {
	return len(c.informer.GetStore().List())
}

// IsSynced 检查是否已同步
func (c *InformerCache) IsSynced() bool {
	return c.informer.HasSynced()
}

// Close 关闭缓存
func (c *InformerCache) Close() {
	close(c.stopCh)
	c.logger.Info("Informer cache stopped")
}

// CacheManager 缓存管理器
type CacheManager struct {
	caches map[ResourceType]ResourceCache
	mutex  sync.RWMutex
	logger *zap.Logger
}

// NewCacheManager 创建新的缓存管理器
func NewCacheManager(logger *zap.Logger) *CacheManager {
	return &CacheManager{
		caches: make(map[ResourceType]ResourceCache),
		logger: logger,
	}
}

// RegisterCache 注册资源缓存
func (m *CacheManager) RegisterCache(resourceType ResourceType, cache ResourceCache) {
	m.mutex.Lock()
	defer m.mutex.Unlock()
	m.caches[resourceType] = cache
	m.logger.Info("Resource cache registered", zap.String("resource", string(resourceType)))
}

// GetCache 获取资源缓存
func (m *CacheManager) GetCache(resourceType ResourceType) (ResourceCache, bool) {
	m.mutex.RLock()
	defer m.mutex.RUnlock()
	cache, exists := m.caches[resourceType]
	return cache, exists
}

// ListResources 列出资源
func (m *CacheManager) ListResources(resourceType ResourceType) []interface{} {
	cache, exists := m.GetCache(resourceType)
	if !exists {
		return nil
	}
	return cache.List()
}

// StartAll 启动所有缓存
func (m *CacheManager) StartAll() {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	for resourceType, cache := range m.caches {
		cache.Start()
		m.logger.Info("Starting cache", zap.String("resource", string(resourceType)))
	}
}

// WaitForCacheSync 等待所有缓存同步
func (m *CacheManager) WaitForCacheSync(timeout time.Duration) bool {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	ticker := time.NewTicker(100 * time.Millisecond)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return false
		case <-ticker.C:
			allSynced := true
			m.mutex.RLock()
			for _, cache := range m.caches {
				if !cache.IsSynced() {
					allSynced = false
					break
				}
			}
			m.mutex.RUnlock()

			if allSynced {
				m.logger.Info("All caches synced")
				return true
			}
		}
	}
}

// Close 关闭所有缓存
func (m *CacheManager) Close() {
	m.mutex.Lock()
	defer m.mutex.Unlock()

	for resourceType, cache := range m.caches {
		cache.Close()
		m.logger.Info("Cache closed", zap.String("resource", string(resourceType)))
	}
}

// GetCacheStats 获取缓存统计
func (m *CacheManager) GetCacheStats() map[string]int {
	m.mutex.RLock()
	defer m.mutex.RUnlock()

	stats := make(map[string]int)
	for resourceType, cache := range m.caches {
		stats[string(resourceType)] = cache.Count()
	}
	return stats
}
