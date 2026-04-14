package service

import (
	"math"
	"testing"
)

// TestParseCPU 测试 CPU 解析
func TestParseCPU(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected float64
	}{
		{"empty string", "", 0},
		{"nanos", "500000000n", 0.5},
		{"millis", "500m", 0.5},
		{"whole number", "2", 2.0},
		{"decimal", "1.5", 1.5},
		{"1000m", "1000m", 1.0},
		{"250m", "250m", 0.25},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseCPU(tt.input)
			if math.Abs(result-tt.expected) > 0.0001 {
				t.Errorf("ParseCPU(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

// TestParseMemory 测试内存解析
func TestParseMemory(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected float64
	}{
		{"empty string", "", 0},
		{"Ki", "1048576Ki", 1.0},      // 1 GiB
		{"Mi", "1024Mi", 1.0},          // 1 GiB
		{"Gi", "2Gi", 2.0},             // 2 GiB
		{"bytes", "1073741824", 1.0},   // 1 GiB in bytes
		{"512Mi", "512Mi", 0.5},        // 0.5 GiB
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := ParseMemory(tt.input)
			if math.Abs(result-tt.expected) > 0.01 {
				t.Errorf("ParseMemory(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}
