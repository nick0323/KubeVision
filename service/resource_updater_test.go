package service

import (
	"context"
	"encoding/json"
	"testing"

	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	storagev1 "k8s.io/api/storage/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestUpdateResourceByType_Pod(t *testing.T) {
	// 1. 准备假客户端
	initialPod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:            "test-pod",
			Namespace:       "default",
			ResourceVersion: "1",
		},
	}
	clientset := fake.NewSimpleClientset(initialPod)

	// 2. 准备更新数据
	updatedPod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:            "test-pod",
			Namespace:       "default",
			ResourceVersion: "1", // 必须提供 ResourceVersion
			Labels:          map[string]string{"updated": "true"},
		},
	}
	jsonBytes, _ := json.Marshal(updatedPod)

	// 3. 执行更新
	err := UpdateResourceByType(context.Background(), clientset, "pod", "default", "test-pod", jsonBytes)
	if err != nil {
		t.Fatalf("UpdateResourceByType() error = %v", err)
	}

	// 4. 验证结果
	pod, err := clientset.CoreV1().Pods("default").Get(context.Background(), "test-pod", metav1.GetOptions{})
	if err != nil {
		t.Fatalf("Failed to get pod: %v", err)
	}
	if pod.Labels["updated"] != "true" {
		t.Errorf("Expected label 'updated' to be 'true', got %v", pod.Labels["updated"])
	}
}

func TestUpdateResourceByType_ResourceNotFound(t *testing.T) {
	clientset := fake.NewSimpleClientset()
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-pod",
			Namespace: "default",
		},
	}
	jsonBytes, _ := json.Marshal(pod)

	err := UpdateResourceByType(context.Background(), clientset, "pod", "default", "test-pod", jsonBytes)
	if err == nil {
		t.Fatal("Expected error for non-existent resource, got nil")
	}
	assert.Contains(t, err.Error(), "not found")
}

func TestUpdateResourceByType_UnsupportedType(t *testing.T) {
	clientset := fake.NewSimpleClientset()
	err := UpdateResourceByType(context.Background(), clientset, "unknown", "default", "test", []byte("{}"))
	if err == nil {
		t.Fatal("Expected error for unsupported type, got nil")
	}
}

func TestResourceFactory(t *testing.T) {
	tests := []struct {
		name         string
		resourceType string
		expectedType interface{}
		expectErr    bool
	}{
		{"pod", "pod", &corev1.Pod{}, false},
		{"deployment", "deployment", &appsv1.Deployment{}, false},
		{"statefulset", "statefulset", &appsv1.StatefulSet{}, false},
		{"daemonset", "daemonset", &appsv1.DaemonSet{}, false},
		{"service", "service", &corev1.Service{}, false},
		{"configmap", "configmap", &corev1.ConfigMap{}, false},
		{"secret", "secret", &corev1.Secret{}, false},
		{"ingress", "ingress", &networkingv1.Ingress{}, false},
		{"job", "job", &batchv1.Job{}, false},
		{"cronjob", "cronjob", &batchv1.CronJob{}, false},
		{"pvc", "pvc", &corev1.PersistentVolumeClaim{}, false},
		{"pv", "pv", &corev1.PersistentVolume{}, false},
		{"storageclass", "storageclass", &storagev1.StorageClass{}, false},
		{"namespace", "namespace", &corev1.Namespace{}, false},
		{"node", "node", &corev1.Node{}, false},
		{"unsupported", "unknown", nil, true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := resourceFactory(tt.resourceType)
			if tt.expectErr {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.IsType(t, tt.expectedType, result)
			}
		})
	}
}

func TestGetResourceVersion(t *testing.T) {
	tests := []struct {
		name     string
		obj      interface{}
		expected string
	}{
		{
			name: "pod with resourceVersion",
			obj: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{
					ResourceVersion: "12345",
				},
			},
			expected: "12345",
		},
		{
			name: "pod without resourceVersion",
			obj: &corev1.Pod{
				ObjectMeta: metav1.ObjectMeta{},
			},
			expected: "",
		},
		{
			name:     "nil object",
			obj:      nil,
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getResourceVersion(tt.obj)
			assert.Equal(t, tt.expected, result)
		})
	}
}
