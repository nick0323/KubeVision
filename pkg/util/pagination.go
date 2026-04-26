package util

import (
	"fmt"
	"reflect"
	"sort"
	"strconv"
	"strings"

	"github.com/nick0323/K8sVision/model"
)

// Paginate 分页函数（泛型）
func Paginate[T any](list []T, offset, limit int) []T {
	if offset < 0 || limit <= 0 || offset >= len(list) {
		return []T{}
	}
	if end := offset + limit; end < len(list) {
		return list[offset:end]
	}
	return list[offset:]
}

// GenericSearchFilter 通用搜索过滤（泛型）
func GenericSearchFilter[T model.SearchableItem](items []T, search string) []T {
	if search == "" {
		return items
	}
	searchLower := strings.ToLower(search)
	filtered := make([]T, 0, len(items))
	for _, item := range items {
		if matchesSearch(item, searchLower) {
			filtered = append(filtered, item)
		}
	}
	return filtered
}

func matchesSearch(item model.SearchableItem, searchLower string) bool {
	for _, value := range item.GetSearchableFields() {
		if strings.Contains(strings.ToLower(value), searchLower) {
			return true
		}
	}
	return false
}

// SortItems 排序函数（泛型）
func SortItems[T any](items []T, sortBy, sortOrder string) []T {
	sorted := make([]T, len(items))
	copy(sorted, items)
	sort.Slice(sorted, func(i, j int) bool {
		valI := getFieldValue(sorted[i], sortBy)
		valJ := getFieldValue(sorted[j], sortBy)
		compare := compareValues(valI, valJ, sortBy)
		if sortOrder == "desc" {
			compare = -compare
		}
		return compare < 0
	})
	return sorted
}

func getFieldValue(item interface{}, fieldName string) string {
	if item == nil {
		return ""
	}
	if m, ok := item.(map[string]interface{}); ok {
		if val, exists := m[fieldName]; exists {
			return fmt.Sprintf("%v", val)
		}
	}
	v := reflect.ValueOf(item)
	if v.Kind() == reflect.Ptr {
		v = v.Elem()
	}
	if v.Kind() != reflect.Struct {
		return ""
	}
	for _, name := range []string{
		fieldName,
		strings.ToUpper(fieldName[:1]) + fieldName[1:],
		strings.ToLower(fieldName[:1]) + fieldName[1:],
	} {
		if field := v.FieldByName(name); field.IsValid() && field.CanInterface() {
			return fmt.Sprintf("%v", field.Interface())
		}
	}
	return ""
}

func compareValues(valA, valB, sortBy string) int {
	if sortBy == "age" || sortBy == "Age" {
		timeA := parseAgeToSeconds(valA)
		timeB := parseAgeToSeconds(valB)
		if timeA < timeB {
			return -1
		} else if timeA > timeB {
			return 1
		}
		return 0
	}
	return strings.Compare(valA, valB)
}

func parseAgeToSeconds(ageStr string) int64 {
	if ageStr == "" || ageStr == "-" {
		return 0
	}
	i := 0
	for i < len(ageStr) && ageStr[i] >= '0' && ageStr[i] <= '9' {
		i++
	}
	if i == 0 {
		return 0
	}
	num, _ := strconv.ParseInt(ageStr[:i], 10, 64)
	unit := ageStr[i:]
	switch unit {
	case "s":
		return num
	case "m":
		return num * 60
	case "h":
		return num * 3600
	case "d":
		return num * 86400
	default:
		return num
	}
}
