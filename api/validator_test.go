package api

import (
	"testing"
)

func TestIsValidDNSName(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{"valid lowercase", "my-pod", true},
		{"valid with numbers", "pod-123", true},
		{"valid single char", "a", true},
		{"invalid starts with dash", "-pod", false},
		{"invalid ends with dash", "pod-", false},
		{"invalid uppercase", "MyPod", false},
		{"invalid special char", "pod_1", false},
		{"invalid empty", "", false},
		{"invalid too long", "a" + string(make([]byte, 253)), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isValidDNSName(tt.input); got != tt.want {
				t.Errorf("isValidDNSName(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestIsValidResourceName(t *testing.T) {
	tests := []struct {
		name  string
		input string
		want  bool
	}{
		{"valid name", "my-resource", true},
		{"invalid path traversal", "../etc/passwd", false},
		{"invalid backslash", "dir\\file", false},
		{"invalid DNS name", "-invalid", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := isValidResourceName(tt.input); got != tt.want {
				t.Errorf("isValidResourceName(%q) = %v, want %v", tt.input, got, tt.want)
			}
		})
	}
}

func TestValidateResourceParams(t *testing.T) {
	tests := []struct {
		name         string
		resourceType string
		namespace    string
		wantErr      bool
	}{
		{"valid namespaced", "pod", "default", false},
		{"valid cluster-scoped", "node", "", false},
		{"valid cluster-scoped pv", "pv", "", false},
		{"missing namespace for pod", "pod", "", true},
		{"unexpected namespace for node", "node", "default", true},
		{"valid plural forms", "deployments", "kube-system", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validateResourceParams(tt.resourceType, tt.namespace)
			if (err != nil) != tt.wantErr {
				t.Errorf("validateResourceParams(%q, %q) error = %v, wantErr %v", tt.resourceType, tt.namespace, err, tt.wantErr)
			}
		})
	}
}
