package service

import (
	"context"
	"encoding/json"
	"testing"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestUpdateResourceByType_Pod(t *testing.T) {
	// 1. 准备假客户端
	initialPod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:            "test-pod",
			Namespace:       "default",
			ResourceVersion: "1",
		},
	}
	clientset := fake.NewSimpleClientset(initialPod)

	// 2. 准备更新数据
	updatedPod := &v1.Pod{
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

func TestUpdateResourceByType_MissingResourceVersion(t *testing.T) {
	clientset := fake.NewSimpleClientset()
	pod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-pod",
			Namespace: "default",
		},
	}
	jsonBytes, _ := json.Marshal(pod)

	err := UpdateResourceByType(context.Background(), clientset, "pod", "default", "test-pod", jsonBytes)
	if err == nil {
		t.Fatal("Expected error for missing ResourceVersion, got nil")
	}
	if err.Error() != "missing required field: resourceVersion" {
		t.Errorf("Unexpected error message: %v", err)
	}
}

func TestUpdateResourceByType_UnsupportedType(t *testing.T) {
	clientset := fake.NewSimpleClientset()
	err := UpdateResourceByType(context.Background(), clientset, "unknown", "default", "test", []byte("{}"))
	if err == nil {
		t.Fatal("Expected error for unsupported type, got nil")
	}
}
