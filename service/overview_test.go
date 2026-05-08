package service

import (
	"context"
	"testing"
	"time"

	"github.com/nick0323/K8sVision/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/fake"
)

// MockKubernetesClientset 用于测试
type MockKubernetesClientset struct {
	*kubernetes.Clientset
	mock.Mock
}

func TestOverviewService_GetOverview(t *testing.T) {
	clientset := fake.NewSimpleClientset()

	// 添加测试数据
	node := &corev1.Node{
		ObjectMeta: metav1.ObjectMeta{Name: "test-node"},
		Status: corev1.NodeStatus{
			Conditions: []corev1.NodeCondition{
				{Type: corev1.NodeReady, Status: corev1.ConditionTrue},
			},
		},
	}
	_, err := clientset.CoreV1().Nodes().Create(context.Background(), node, metav1.CreateOptions{})
	assert.NoError(t, err)

	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{Name: "test-pod", Namespace: "default"},
		Spec: corev1.PodSpec{NodeName: "test-node"},
		Status: corev1.PodStatus{Phase: corev1.PodRunning},
	}
	_, err = clientset.CoreV1().Pods("default").Create(context.Background(), pod, metav1.CreateOptions{})
	assert.NoError(t, err)

	namespace := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{Name: "default"},
	}
	_, err = clientset.CoreV1().Namespaces().Create(context.Background(), namespace, metav1.CreateOptions{})
	assert.NoError(t, err)

	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{Name: "test-service", Namespace: "default"},
	}
	_, err = clientset.CoreV1().Services("default").Create(context.Background(), svc, metav1.CreateOptions{})
	assert.NoError(t, err)

	service := NewOverviewService(clientset)
	overview, err := service.GetOverview(context.Background())

	assert.NoError(t, err)
	assert.NotNil(t, overview)
	assert.Equal(t, 1, overview.NodeCount)
	assert.Equal(t, 1, overview.PodCount)
	assert.Equal(t, 1, overview.NamespaceCount)
	assert.Equal(t, 1, overview.ServiceCount)
}

func TestOverviewService_GetRecentEvents(t *testing.T) {
	clientset := fake.NewSimpleClientset()

	// 创建测试事件
	event := &corev1.Event{
		ObjectMeta: metav1.ObjectMeta{Name: "test-event", Namespace: "default"},
		InvolvedObject: corev1.ObjectReference{Kind: "Pod", Name: "test-pod"},
		Reason:        "Created",
		Message:       "Pod was created",
		LastTimestamp: metav1.NewTime(time.Now()),
	}
	_, err := clientset.CoreV1().Events("default").Create(context.Background(), event, metav1.CreateOptions{})
	assert.NoError(t, err)

	service := NewOverviewService(clientset)
	events, err := service.getRecentEvents(context.Background(), 1)

	assert.NoError(t, err)
	assert.NotNil(t, events)
	// 注意：fake clientset 对 events 支持有限，这里主要测试不 panic
}

func TestFormatTime(t *testing.T) {
	tests := []struct {
		name     string
		time     *metav1.Time
		expected string
	}{
		{"nil time", nil, ""},
		{"valid time", &metav1.Time{Time: time.Now()}, ""},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := model.FormatTime(tt.time)
			if tt.time == nil {
				assert.Equal(t, "", result)
			} else {
				assert.NotEmpty(t, result)
			}
		})
	}
}

func TestMin(t *testing.T) {
	tests := []struct {
		name     string
		a        int
		b        int
		expected int
	}{
		{"a smaller", 3, 5, 3},
		{"b smaller", 5, 3, 3},
		{"equal", 4, 4, 4},
		{"negative", -1, 1, -1},
		{"zero", 0, 5, 0},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := min(tt.a, tt.b)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestRound(t *testing.T) {
	tests := []struct {
		name     string
		val      float64
		expected float64
	}{
		{"positive", 3.14159, 3.1},
		{"negative", -3.14159, -3.1},
		{"zero", 0.0, 0.0},
		{"rounds up", 2.25, 2.3},
		{"rounds down", 2.24, 2.2},
		{"large number", 123456.789, 123456.8},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := round(tt.val)
			assert.Equal(t, tt.expected, result)
		})
	}
}
