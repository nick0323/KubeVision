package api

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	v1 "k8s.io/api/core/v1"
)

func TestValidateExecParams(t *testing.T) {
	tests := []struct {
		name      string
		namespace string
		podName   string
		wantErr   bool
		errMsg    string
	}{
		{"both empty", "", "", true, "invalid namespace format"},
		{"empty podName", "default", "", true, "missing pod name"},
		{"empty namespace", "", "my-pod", true, "invalid namespace format"},
		{"both valid", "default", "my-pod", false, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateExecParams(tt.namespace, tt.podName)
			if tt.wantErr {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), tt.errMsg)
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestParseExecCommand(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []string
		wantErr  bool
	}{
		{"empty string", "", []string{"/bin/sh"}, false},
		{"single command", "/bin/sh", []string{"/bin/sh"}, false},
		{"command with args", "/bin/bash -c ls", []string{"/bin/bash", "-c", "ls"}, false},
		{"multiple spaces", "/bin/sh   -c   echo", []string{"/bin/sh", "-c", "echo"}, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := parseExecCommand(tt.input)
			if tt.wantErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestIsValidNamespace(t *testing.T) {
	tests := []struct {
		name      string
		namespace string
		valid     bool
	}{
		{"empty", "", false},
		{"valid short", "default", true},
		{"valid with hyphen", "kube-system", true},
		{"valid with number", "ns2", true},
		{"valid alphanumeric", "my-namespace-123", true},
		{"invalid with uppercase", "Default", false},
		{"invalid with dot", "my.namespace", false},
		{"invalid with underscore", "my_namespace", false},
		{"too long", "a" + string(make([]byte, 253)), false},
		{"starts with hyphen", "-namespace", false},
		{"ends with hyphen", "namespace-", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isValidNamespace(tt.namespace)
			assert.Equal(t, tt.valid, result)
		})
	}
}

func TestHasContainer(t *testing.T) {
	pod := &v1.Pod{
		Spec: v1.PodSpec{
			Containers: []v1.Container{
				{Name: "nginx"},
				{Name: "sidecar"},
			},
			InitContainers: []v1.Container{
				{Name: "init-setup"},
			},
		},
	}

	tests := []struct {
		name     string
		pod      *v1.Pod
		container string
		exists   bool
	}{
		{"regular container", pod, "nginx", true},
		{"second container", pod, "sidecar", true},
		{"init container", pod, "init-setup", true},
		{"non-existent", pod, "not-found", false},
		{"empty returns true", pod, "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := hasContainer(tt.pod, tt.container)
			assert.Equal(t, tt.exists, result)
		})
	}
}

func TestCheckExecConnectionLimit(t *testing.T) {
	logger := zap.NewNop()

	InitWebSocketManager(100)
	err := checkExecConnectionLimit(logger, "testuser")
	assert.NoError(t, err)
	InitWebSocketManager(100)
}
