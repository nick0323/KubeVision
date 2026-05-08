package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/nick0323/K8sVision/model"
)

func TestPaginate(t *testing.T) {
	tests := []struct {
		name     string
		items    []int
		offset   int
		limit    int
		expected int
	}{
		{
			name:     "paginate first page",
			items:    []int{1, 2, 3, 4, 5},
			offset:   0,
			limit:    2,
			expected: 2,
		},
		{
			name:     "paginate second page",
			items:    []int{1, 2, 3, 4, 5},
			offset:   2,
			limit:    2,
			expected: 2,
		},
		{
			name:     "offset beyond length",
			items:    []int{1, 2, 3},
			offset:   5,
			limit:    2,
			expected: 0,
		},
		{
			name:     "limit larger than remaining",
			items:    []int{1, 2, 3},
			offset:   1,
			limit:    10,
			expected: 2,
		},
		{
			name:     "empty slice",
			items:    []int{},
			offset:   0,
			limit:    10,
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := Paginate(tt.items, tt.offset, tt.limit)
			assert.Equal(t, tt.expected, len(result))
		})
	}
}

func TestPaginate_NilInput(t *testing.T) {
	result := Paginate([]int{}, 0, 10)
	assert.NotNil(t, result)
	assert.Equal(t, 0, len(result))
}

func TestGenericSearchFilter(t *testing.T) {
	tests := []struct {
		name     string
		items    []model.SearchableItem
		search   string
		expected int
	}{
		{
			name: "search by name",
			items: []model.SearchableItem{
				model.Pod{Name: "pod-1", Namespace: "default"},
				model.Pod{Name: "pod-2", Namespace: "default"},
				model.Pod{Name: "deploy-1", Namespace: "prod"},
			},
			search:   "pod",
			expected: 2,
		},
		{
			name: "empty search returns all",
			items: []model.SearchableItem{
				model.Pod{Name: "pod-1", Namespace: "default"},
				model.Pod{Name: "pod-2", Namespace: "prod"},
			},
			search:   "",
			expected: 2,
		},
		{
			name:     "empty items",
			items:    []model.SearchableItem{},
			search:   "test",
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GenericSearchFilter(tt.items, tt.search)
			assert.Equal(t, tt.expected, len(result))
		})
	}
}

func TestGenericSearchFilter_NilInput(t *testing.T) {
	result := GenericSearchFilter([]model.SearchableItem{}, "test")
	assert.NotNil(t, result)
}

func TestSortItems(t *testing.T) {
	tests := []struct {
		name     string
		items    []model.Pod
		sortBy   string
		sortOrder string
		expected []string
	}{
		{
			name: "sort by name ascending",
			items: []model.Pod{
				{Name: "pod-b", Namespace: "default"},
				{Name: "pod-a", Namespace: "default"},
				{Name: "pod-c", Namespace: "default"},
			},
			sortBy:   "name",
			sortOrder: "asc",
			expected: []string{"pod-a", "pod-b", "pod-c"},
		},
		{
			name: "sort by name descending",
			items: []model.Pod{
				{Name: "pod-b", Namespace: "default"},
				{Name: "pod-a", Namespace: "default"},
				{Name: "pod-c", Namespace: "default"},
			},
			sortBy:   "name",
			sortOrder: "desc",
			expected: []string{"pod-c", "pod-b", "pod-a"},
		},
		{
			name: "empty items",
			items: []model.Pod{},
			sortBy: "name",
			sortOrder: "asc",
			expected: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SortItems(tt.items, tt.sortBy, tt.sortOrder)
			names := make([]string, len(result))
			for i, item := range result {
				names[i] = item.Name
			}
			assert.Equal(t, tt.expected, names)
		})
	}
}

func TestGetFieldValue(t *testing.T) {
	tests := []struct {
		name     string
		item     interface{}
		fieldName string
		expected string
	}{
		{
			name: "pod name",
			item: model.Pod{
				Name: "test-pod",
			},
			fieldName: "name",
			expected: "test-pod",
		},
		{
			name: "pod namespace",
			item: model.Pod{
				Name: "test-pod",
				Namespace: "default",
			},
			fieldName: "namespace",
			expected: "default",
		},
		{
			name: "nil item",
			item: nil,
			fieldName: "name",
			expected: "",
		},
		{
			name:     "non-struct item",
			item:     "string",
			fieldName: "name",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getFieldValue(tt.item, tt.fieldName)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCompareValues(t *testing.T) {
	tests := []struct {
		name     string
		valA     string
		valB     string
		sortBy   string
		expected int
	}{
		{
			name: "string comparison A < B",
			valA: "abc",
			valB: "def",
			sortBy: "name",
			expected: -1,
		},
		{
			name: "string comparison A > B",
			valA: "def",
			valB: "abc",
			sortBy: "name",
			expected: 1,
		},
		{
			name: "string comparison A = B",
			valA: "abc",
			valB: "abc",
			sortBy: "name",
			expected: 0,
		},
		{
			name: "age comparison A < B",
			valA: "1h",
			valB: "2h",
			sortBy: "age",
			expected: -1,
		},
		{
			name: "age comparison A > B",
			valA: "2h",
			valB: "1h",
			sortBy: "age",
			expected: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := compareValues(tt.valA, tt.valB, tt.sortBy)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestParseAgeToSeconds(t *testing.T) {
	tests := []struct {
		name     string
		ageStr   string
		expected int64
	}{
		{
			name: "seconds",
			ageStr: "30s",
			expected: 30,
		},
		{
			name: "minutes",
			ageStr: "5m",
			expected: 300,
		},
		{
			name: "hours",
			ageStr: "2h",
			expected: 7200,
		},
		{
			name: "days",
			ageStr: "3d",
			expected: 259200,
		},
		{
			name: "empty string",
			ageStr: "",
			expected: 0,
		},
		{
			name: "dash",
			ageStr: "-",
			expected: 0,
		},
		{
			name: "invalid format",
			ageStr: "invalid",
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := parseAgeToSeconds(tt.ageStr)
			assert.Equal(t, tt.expected, result)
		})
	}
}
