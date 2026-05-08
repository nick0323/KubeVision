package api

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
)

func TestBuildPodLogOptions(t *testing.T) {
	tests := []struct {
		name       string
		container  string
		timestamps string
		previous   string
		tailLines  string
		expected   *corev1.PodLogOptions
	}{
		{
			name:       "all empty",
			container:  "",
			timestamps: "",
			previous:   "",
			tailLines:  "",
			expected:   &corev1.PodLogOptions{Follow: true},
		},
		{
			name:       "with container and timestamps",
			container:  "nginx",
			timestamps: "true",
			previous:   "",
			tailLines:  "",
			expected: &corev1.PodLogOptions{
				Container:  "nginx",
				Follow:     true,
				Timestamps: true,
			},
		},
		{
			name:       "with previous and tailLines",
			container:  "",
			timestamps: "",
			previous:   "true",
			tailLines:  "100",
			expected: &corev1.PodLogOptions{
				Follow:   true,
				Previous: true,
				TailLines: int64Ptr(100),
			},
		},
		{
			name:       "all values set",
			container:  "app",
			timestamps: "true",
			previous:   "false",
			tailLines:  "50",
			expected: &corev1.PodLogOptions{
				Container:  "app",
				Follow:     true,
				Timestamps: true,
				TailLines:  int64Ptr(50),
			},
		},
		{
			name:       "timestamps false",
			container:  "",
			timestamps: "false",
			previous:   "",
			tailLines:  "",
			expected:   &corev1.PodLogOptions{Follow: true},
		},
		{
			name:       "invalid tailLines",
			container:  "",
			timestamps: "",
			previous:   "",
			tailLines:  "invalid",
			expected:   &corev1.PodLogOptions{Follow: true},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := buildPodLogOptions(tt.container, tt.timestamps, tt.previous, tt.tailLines)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCheckConnectionLimit(t *testing.T) {
	logger := zap.NewNop()

	InitWebSocketManager(100)
	err := checkConnectionLimit(logger)
	assert.NoError(t, err)

	InitWebSocketManager(100)
}

func TestCheckConnectionLimit_Unlimited(t *testing.T) {
	logger := zap.NewNop()
	InitWebSocketManager(-1)
	err := checkConnectionLimit(logger)
	assert.NoError(t, err)
	InitWebSocketManager(100)
}

func int64Ptr(i int) *int64 {
	v := int64(i)
	return &v
}
