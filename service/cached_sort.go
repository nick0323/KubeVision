package service

import (
	"reflect"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/nick0323/K8sVision/cache"
	"go.uber.org/zap"
)

// CachedSortService 缓存排序服务
type CachedSortService struct {
	cacheManager *cache.CacheManager
	logger       *zap.Logger
	sortCache    sync.Map // map[string]*SortCacheEntry
}

// SortCacheEntry 排序缓存条目
type SortCacheEntry struct {
	Data      []interface{}
	SortBy    string
	SortOrder string
	Timestamp time.Time
	ExpiresAt time.Time
}

// NewCachedSortService 创建新的缓存排序服务
func NewCachedSortService(
	cacheManager *cache.CacheManager,
	logger *zap.Logger,
) *CachedSortService {
	return &CachedSortService{
		cacheManager: cacheManager,
		logger:       logger,
		sortCache:    sync.Map{},
	}
}

// GetSortedData 获取排序后的数据（带缓存）
func (s *CachedSortService) GetSortedData(
	resourceType cache.ResourceType,
	sortBy string,
	sortOrder string,
	limit int,
	offset int,
) ([]interface{}, int, error) {
	cacheKey := s.makeCacheKey(resourceType, sortBy, sortOrder)

	// 检查缓存
	if cached, exists := s.sortCache.Load(cacheKey); exists {
		entry := cached.(*SortCacheEntry)
		if time.Now().Before(entry.ExpiresAt) {
			// 缓存命中
			s.logger.Debug("Sort cache hit",
				zap.String("resource", string(resourceType)),
				zap.String("sortBy", sortBy),
				zap.String("sortOrder", sortOrder),
			)
			return s.paginateData(entry.Data, limit, offset), len(entry.Data), nil
		}
		// 缓存过期，删除
		s.sortCache.Delete(cacheKey)
	}

	// 缓存未命中，从缓存管理器获取数据并排序
	data := s.cacheManager.ListResources(resourceType)
	if data == nil {
		return nil, 0, nil
	}

	// 排序
	sortedData := s.sortData(data, sortBy, sortOrder)

	// 缓存结果（5 分钟过期）
	s.sortCache.Store(cacheKey, &SortCacheEntry{
		Data:      sortedData,
		SortBy:    sortBy,
		SortOrder: sortOrder,
		Timestamp: time.Now(),
		ExpiresAt: time.Now().Add(5 * time.Minute),
	})

	s.logger.Debug("Sort cache miss, data sorted and cached",
		zap.String("resource", string(resourceType)),
		zap.String("sortBy", sortBy),
		zap.Int("count", len(sortedData)),
	)

	return s.paginateData(sortedData, limit, offset), len(sortedData), nil
}

// InvalidateCache 使缓存失效
func (s *CachedSortService) InvalidateCache(resourceType cache.ResourceType) {
	cacheKeyPrefix := string(resourceType) + "|"

	s.sortCache.Range(func(key, value interface{}) bool {
		if keyStr, ok := key.(string); ok {
			if strings.HasPrefix(keyStr, cacheKeyPrefix) {
				s.sortCache.Delete(keyStr)
				s.logger.Debug("Sort cache invalidated",
					zap.String("key", keyStr),
				)
			}
		}
		return true
	})
}

// ClearCache 清空所有排序缓存
func (s *CachedSortService) ClearCache() {
	s.sortCache.Range(func(key, value interface{}) bool {
		s.sortCache.Delete(key)
		return true
	})
	s.logger.Info("All sort cache cleared")
}

// GetCacheStats 获取缓存统计
func (s *CachedSortService) GetCacheStats() map[string]interface{} {
	stats := make(map[string]interface{})
	count := 0

	s.sortCache.Range(func(key, value interface{}) bool {
		count++
		return true
	})

	stats["sortCacheCount"] = count
	return stats
}

// makeCacheKey 生成缓存键
func (s *CachedSortService) makeCacheKey(
	resourceType cache.ResourceType,
	sortBy string,
	sortOrder string,
) string {
	return string(resourceType) + "|" + sortBy + "|" + sortOrder
}

// sortData 对数据进行排序
func (s *CachedSortService) sortData(data []interface{}, sortBy string, sortOrder string) []interface{} {
	// 创建副本
	sorted := make([]interface{}, len(data))
	copy(sorted, data)

	// 使用 Go 标准库排序（快速排序）
	sort.Slice(sorted, func(i, j int) bool {
		valI := s.getFieldValue(sorted[i], sortBy)
		valJ := s.getFieldValue(sorted[j], sortBy)

		compare := strings.Compare(valI, valJ)
		if sortOrder == "desc" {
			compare = -compare
		}
		return compare < 0
	})

	return sorted
}

// paginateData 分页数据
func (s *CachedSortService) paginateData(data []interface{}, limit int, offset int) []interface{} {
	if offset >= len(data) || offset < 0 {
		return []interface{}{}
	}
	if limit <= 0 {
		return []interface{}{}
	}

	end := offset + limit
	if end > len(data) {
		end = len(data)
	}

	if offset >= end {
		return []interface{}{}
	}

	return data[offset:end]
}

// getFieldValue 获取字段值
func (s *CachedSortService) getFieldValue(item interface{}, fieldName string) string {
	if item == nil {
		return ""
	}

	// 尝试作为 map 处理
	if m, ok := item.(map[string]interface{}); ok {
		if val, exists := m[fieldName]; exists {
			return s.toString(val)
		}
	}

	// 使用反射获取结构体字段
	v := reflect.ValueOf(item)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if v.Kind() == reflect.Struct {
		field := v.FieldByName(fieldName)
		if field.IsValid() {
			return s.toString(field.Interface())
		}
		// 尝试小写首字母
		fieldNameLower := strings.ToLower(fieldName[:1]) + fieldName[1:]
		field = v.FieldByName(fieldNameLower)
		if field.IsValid() {
			return s.toString(field.Interface())
		}
	}

	return ""
}

// toString 转换为字符串
func (s *CachedSortService) toString(val interface{}) string {
	if val == nil {
		return ""
	}
	switch v := val.(type) {
	case string:
		return v
	case int, int32, int64:
		return strconv.FormatInt(reflect.ValueOf(v).Int(), 10)
	case float32, float64:
		return strconv.FormatFloat(reflect.ValueOf(v).Float(), 'f', -1, 64)
	case bool:
		if v {
			return "true"
		}
		return "false"
	default:
		return reflect.ValueOf(val).String()
	}
}
