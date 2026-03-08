package service

import "github.com/nick0323/K8sVision/cache"

var (
	cacheManager     *cache.CacheManager
	cachedSortService *CachedSortService
)

// SetCacheManager 设置缓存管理器
func SetCacheManager(cm *cache.CacheManager) {
	cacheManager = cm
}

// GetCacheManager 获取缓存管理器
func GetCacheManager() *cache.CacheManager {
	return cacheManager
}

// SetCachedSortService 设置缓存排序服务
func SetCachedSortService(css *CachedSortService) {
	cachedSortService = css
}

// GetCachedSortService 获取缓存排序服务
func GetCachedSortService() *CachedSortService {
	return cachedSortService
}
