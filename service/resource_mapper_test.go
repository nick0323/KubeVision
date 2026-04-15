package service

import (
	"testing"
	"time"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ==================== ParseCPU 测试 ====================

// TestParseCPU_Existing 测试已有的 CPU 解析功能（由 common_test.go 覆盖）
// 这里添加额外的边界情况测试
func TestParseCPU_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected float64
	}{
		{"negative nanos", "-500000000n", -0.5},
		{"negative millis", "-500m", -0.5},
		{"negative number", "-2", -2.0},
		{"zero nanos", "0n", 0},
		{"zero millis", "0m", 0},
		{"large value", "10000m", 10.0},
		{"small nanos", "1n", 1e-9},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseCPU(tt.input)
			if result != tt.expected {
				t.Errorf("ParseCPU(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

// ==================== ParseMemory 测试 ====================

// TestParseMemory_Existing 测试已有的内存解析功能（由 common_test.go 覆盖）
// 这里添加额外的边界情况测试
func TestParseMemory_EdgeCases(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected float64
	}{
		{"negative Ki", "-1048576Ki", -1.0},
		{"negative Mi", "-1024Mi", -1.0},
		{"negative Gi", "-2Gi", -2.0},
		{"zero bytes", "0", 0},
		{"large Gi", "100Gi", 100.0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseMemory(tt.input)
			if result != tt.expected {
				t.Errorf("ParseMemory(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

// ==================== GetWorkloadStatus 测试 ====================

func TestGetWorkloadStatus(t *testing.T) {
	tests := []struct {
		name     string
		ready    int32
		desired  int32
		expected string
	}{
		{"fully available", 3, 3, "Available"},
		{"partially available", 2, 3, "Partial"},
		{"unavailable", 0, 3, "Unavailable"},
		{"desired zero available", 0, 0, "Unavailable"},
		{"one ready one desired", 1, 1, "Available"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetWorkloadStatus(tt.ready, tt.desired)
			if result != tt.expected {
				t.Errorf("GetWorkloadStatus(%d, %d) = %s, want %s", tt.ready, tt.desired, result, tt.expected)
			}
		})
	}
}

// ==================== TruncateString 测试 ====================

func TestTruncateString(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		maxLen   int
		expected string
	}{
		{"no truncation needed", "hello", 10, "hello"},
		{"exact length", "hello", 5, "hello"},
		{"truncation needed", "hello world", 5, "hello..."},
		{"empty string", "", 5, ""},
		{"zero max length", "hello", 0, "..."},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := TruncateString(tt.input, tt.maxLen)
			if result != tt.expected {
				t.Errorf("TruncateString(%q, %d) = %q, want %q", tt.input, tt.maxLen, result, tt.expected)
			}
		})
	}
}

// ==================== CalculateAge 测试 ====================

func TestCalculateAge(t *testing.T) {
	now := metav1.Now()

	// 1 分钟前
	oneMinAgo := metav1.Time{Time: now.Time.Add(-1 * time.Minute)}
	result := CalculateAge(oneMinAgo)
	if result != "1m" {
		t.Errorf("CalculateAge(1 minute ago) = %s, want 1m", result)
	}

	// 1 小时前
	oneHourAgo := metav1.Time{Time: now.Time.Add(-1 * time.Hour)}
	result = CalculateAge(oneHourAgo)
	if result != "1h" {
		t.Errorf("CalculateAge(1 hour ago) = %s, want 1h", result)
	}

	// 1 天前
	oneDayAgo := metav1.Time{Time: now.Time.Add(-24 * time.Hour)}
	result = CalculateAge(oneDayAgo)
	if result != "1d" {
		t.Errorf("CalculateAge(1 day ago) = %s, want 1d", result)
	}

	// 刚刚
	result = CalculateAge(now)
	if result != "0s" {
		t.Errorf("CalculateAge(now) = %s, want 0s", result)
	}
}

// ==================== CalculatePodReady 测试 ====================

func TestCalculatePodReady(t *testing.T) {
	tests := []struct {
		name     string
		pod      v1.Pod
		expected string
	}{
		{
			name: "all containers ready",
			pod: v1.Pod{
				Status: v1.PodStatus{
					ContainerStatuses: []v1.ContainerStatus{
						{Name: "c1", Ready: true},
						{Name: "c2", Ready: true},
					},
				},
			},
			expected: "2/2",
		},
		{
			name: "partial containers ready",
			pod: v1.Pod{
				Status: v1.PodStatus{
					ContainerStatuses: []v1.ContainerStatus{
						{Name: "c1", Ready: true},
						{Name: "c2", Ready: false},
					},
				},
			},
			expected: "1/2",
		},
		{
			name: "no containers",
			pod: v1.Pod{
				Status: v1.PodStatus{
					ContainerStatuses: []v1.ContainerStatus{},
					Phase:             v1.PodPending,
				},
			},
			expected: "Pending",
		},
		{
			name: "no containers failed",
			pod: v1.Pod{
				Status: v1.PodStatus{
					ContainerStatuses: []v1.ContainerStatus{},
					Phase:             v1.PodFailed,
				},
			},
			expected: "Failed",
		},
		{
			name: "no containers succeeded",
			pod: v1.Pod{
				Status: v1.PodStatus{
					ContainerStatuses: []v1.ContainerStatus{},
					Phase:             v1.PodSucceeded,
				},
			},
			expected: "Succeeded",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CalculatePodReady(tt.pod)
			if result != tt.expected {
				t.Errorf("CalculatePodReady = %s, want %s", result, tt.expected)
			}
		})
	}
}

// ==================== CalculatePodRestarts 测试 ====================

func TestCalculatePodRestarts(t *testing.T) {
	tests := []struct {
		name     string
		pod      v1.Pod
		expected int32
	}{
		{
			name: "no restarts",
			pod: v1.Pod{
				Status: v1.PodStatus{
					ContainerStatuses: []v1.ContainerStatus{
						{Name: "c1", RestartCount: 0},
						{Name: "c2", RestartCount: 0},
					},
				},
			},
			expected: 0,
		},
		{
			name: "some restarts",
			pod: v1.Pod{
				Status: v1.PodStatus{
					ContainerStatuses: []v1.ContainerStatus{
						{Name: "c1", RestartCount: 3},
						{Name: "c2", RestartCount: 5},
					},
				},
			},
			expected: 8,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CalculatePodRestarts(tt.pod)
			if result != tt.expected {
				t.Errorf("CalculatePodRestarts = %d, want %d", result, tt.expected)
			}
		})
	}
}
