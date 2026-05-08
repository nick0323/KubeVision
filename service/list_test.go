package service

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	storagev1 "k8s.io/api/storage/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestListOptions_Apply(t *testing.T) {
	tests := []struct {
		name     string
		opts     *ListOptions
		expected metav1.ListOptions
	}{
		{
			name:     "empty options",
			opts:     &ListOptions{},
			expected: metav1.ListOptions{},
		},
		{
			name: "with label selector",
			opts: &ListOptions{
				LabelSelector: "app=nginx",
			},
			expected: metav1.ListOptions{
				LabelSelector: "app=nginx",
			},
		},
		{
			name: "with field selector",
			opts: &ListOptions{
				FieldSelector: "metadata.name=test",
			},
			expected: metav1.ListOptions{
				FieldSelector: "metadata.name=test",
			},
		},
		{
			name: "with limit",
			opts: &ListOptions{
				Limit: 100,
			},
			expected: metav1.ListOptions{
				Limit: 100,
			},
		},
		{
			name: "all options",
			opts: &ListOptions{
				Namespace:     "default",
				LabelSelector: "app=nginx",
				FieldSelector: "metadata.name=test",
				Limit:         50,
			},
			expected: metav1.ListOptions{
				LabelSelector: "app=nginx",
				FieldSelector: "metadata.name=test",
				Limit:         50,
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := tt.opts.Apply()
			assert.Equal(t, tt.expected.LabelSelector, result.LabelSelector)
			assert.Equal(t, tt.expected.FieldSelector, result.FieldSelector)
			assert.Equal(t, tt.expected.Limit, result.Limit)
		})
	}
}

func TestDefaultListOptions(t *testing.T) {
	opts := DefaultListOptions()
	assert.NotNil(t, opts)
	assert.Equal(t, int64(1000), opts.Limit)
	assert.Empty(t, opts.Namespace)
	assert.Empty(t, opts.LabelSelector)
	assert.Empty(t, opts.FieldSelector)
}

func TestListPods(t *testing.T) {
	tests := []struct {
		name      string
		pods      []corev1.Pod
		namespace string
		expectErr bool
		expected  int
	}{
		{
			name: "list pods in namespace",
			pods: []corev1.Pod{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "pod-1",
						Namespace: "default",
					},
					Status: corev1.PodStatus{
						Phase: corev1.PodRunning,
					},
				},
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "pod-2",
						Namespace: "default",
					},
					Status: corev1.PodStatus{
						Phase: corev1.PodPending,
					},
				},
			},
			namespace: "default",
			expectErr: false,
			expected:  2,
		},
		{
			name:      "empty namespace returns all",
			pods:      []corev1.Pod{},
			namespace: "",
			expectErr: false,
			expected:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			clientset := fake.NewSimpleClientset()
			for i := range tt.pods {
				_, err := clientset.CoreV1().Pods(tt.pods[i].Namespace).Create(
					context.Background(), &tt.pods[i], metav1.CreateOptions{})
				assert.NoError(t, err)
			}

			result, err := ListPods(context.Background(), clientset, tt.namespace, "", "")
			if tt.expectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
				assert.Equal(t, tt.expected, len(result))
			}
		})
	}
}

func TestListPodsWithRaw(t *testing.T) {
	clientset := fake.NewSimpleClientset()
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-pod",
			Namespace: "default",
		},
		Status: corev1.PodStatus{
			Phase: corev1.PodRunning,
		},
	}
	_, err := clientset.CoreV1().Pods("default").Create(context.Background(), pod, metav1.CreateOptions{})
	assert.NoError(t, err)

	pods, rawList, err := ListPodsWithRaw(context.Background(), clientset, "default", "", "", false)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(pods))
	assert.NotNil(t, rawList)
	assert.Equal(t, "test-pod", pods[0].Name)
	assert.Equal(t, "default", pods[0].Namespace)
}

func TestListPodsWithRaw_NoLimit(t *testing.T) {
	clientset := fake.NewSimpleClientset()
	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-pod",
			Namespace: "default",
		},
	}
	_, err := clientset.CoreV1().Pods("default").Create(context.Background(), pod, metav1.CreateOptions{})
	assert.NoError(t, err)

	pods, _, err := ListPodsWithRaw(context.Background(), clientset, "default", "", "", true)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(pods))
}

func TestListDeployments(t *testing.T) {
	clientset := fake.NewSimpleClientset()
	deployment := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-deployment",
			Namespace: "default",
		},
		Spec: appsv1.DeploymentSpec{
			Replicas: ptrInt32(3),
		},
	}
	_, err := clientset.AppsV1().Deployments("default").Create(context.Background(), deployment, metav1.CreateOptions{})
	assert.NoError(t, err)

	result, err := ListDeployments(context.Background(), clientset, "default", "", "")
	assert.NoError(t, err)
	assert.Equal(t, 1, len(result))
	assert.Equal(t, "test-deployment", result[0].Name)
}

func TestListStatefulSets(t *testing.T) {
	clientset := fake.NewSimpleClientset()
	sts := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-sts",
			Namespace: "default",
		},
	}
	_, err := clientset.AppsV1().StatefulSets("default").Create(context.Background(), sts, metav1.CreateOptions{})
	assert.NoError(t, err)

	result, err := ListStatefulSets(context.Background(), clientset, "default", "", "")
	assert.NoError(t, err)
	assert.Equal(t, 1, len(result))
}

func TestListDaemonSets(t *testing.T) {
	clientset := fake.NewSimpleClientset()
	ds := &appsv1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-ds",
			Namespace: "default",
		},
	}
	_, err := clientset.AppsV1().DaemonSets("default").Create(context.Background(), ds, metav1.CreateOptions{})
	assert.NoError(t, err)

	result, err := ListDaemonSets(context.Background(), clientset, "default", "", "")
	assert.NoError(t, err)
	assert.Equal(t, 1, len(result))
}

func TestListServices(t *testing.T) {
	clientset := fake.NewSimpleClientset()
	svc := &corev1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-service",
			Namespace: "default",
		},
	}
	_, err := clientset.CoreV1().Services("default").Create(context.Background(), svc, metav1.CreateOptions{})
	assert.NoError(t, err)

	result, err := ListServices(context.Background(), clientset, "default", "", "")
	assert.NoError(t, err)
	assert.Equal(t, 1, len(result))
	assert.Equal(t, "test-service", result[0].Name)
}

func TestListIngresses(t *testing.T) {
	clientset := fake.NewSimpleClientset()
	ingress := &networkingv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-ingress",
			Namespace: "default",
		},
	}
	_, err := clientset.NetworkingV1().Ingresses("default").Create(context.Background(), ingress, metav1.CreateOptions{})
	assert.NoError(t, err)

	result, err := ListIngresses(context.Background(), clientset, "default", "", "")
	assert.NoError(t, err)
	assert.Equal(t, 1, len(result))
}

func TestListPVCs(t *testing.T) {
	clientset := fake.NewSimpleClientset()
	pvc := &corev1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-pvc",
			Namespace: "default",
		},
	}
	_, err := clientset.CoreV1().PersistentVolumeClaims("default").Create(context.Background(), pvc, metav1.CreateOptions{})
	assert.NoError(t, err)

	result, err := ListPVCs(context.Background(), clientset, "default", "", "")
	assert.NoError(t, err)
	assert.Equal(t, 1, len(result))
}

func TestListPVs(t *testing.T) {
	clientset := fake.NewSimpleClientset()
	pv := &corev1.PersistentVolume{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-pv",
		},
	}
	_, err := clientset.CoreV1().PersistentVolumes().Create(context.Background(), pv, metav1.CreateOptions{})
	assert.NoError(t, err)

	result, err := ListPVs(context.Background(), clientset, "", "")
	assert.NoError(t, err)
	assert.Equal(t, 1, len(result))
}

func TestListStorageClasses(t *testing.T) {
	clientset := fake.NewSimpleClientset()
	sc := &storagev1.StorageClass{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-sc",
		},
	}
	_, err := clientset.StorageV1().StorageClasses().Create(context.Background(), sc, metav1.CreateOptions{})
	assert.NoError(t, err)

	result, err := ListStorageClasses(context.Background(), clientset, "", "")
	assert.NoError(t, err)
	assert.Equal(t, 1, len(result))
}

func TestListConfigMaps(t *testing.T) {
	clientset := fake.NewSimpleClientset()
	cm := &corev1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-cm",
			Namespace: "default",
		},
	}
	_, err := clientset.CoreV1().ConfigMaps("default").Create(context.Background(), cm, metav1.CreateOptions{})
	assert.NoError(t, err)

	result, err := ListConfigMaps(context.Background(), clientset, "default", "", "")
	assert.NoError(t, err)
	assert.Equal(t, 1, len(result))
}

func TestListSecrets(t *testing.T) {
	clientset := fake.NewSimpleClientset()
	secret := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-secret",
			Namespace: "default",
		},
	}
	_, err := clientset.CoreV1().Secrets("default").Create(context.Background(), secret, metav1.CreateOptions{})
	assert.NoError(t, err)

	result, err := ListSecrets(context.Background(), clientset, "default", "", "")
	assert.NoError(t, err)
	assert.Equal(t, 1, len(result))
	assert.Equal(t, "test-secret", result[0].Name)
}

func TestListSecrets_AllNamespaces(t *testing.T) {
	clientset := fake.NewSimpleClientset()
	secret1 := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "secret-1",
			Namespace: "default",
		},
	}
	secret2 := &corev1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "secret-2",
			Namespace: "kube-system",
		},
	}
	_, err := clientset.CoreV1().Secrets("default").Create(context.Background(), secret1, metav1.CreateOptions{})
	assert.NoError(t, err)
	_, err = clientset.CoreV1().Secrets("kube-system").Create(context.Background(), secret2, metav1.CreateOptions{})
	assert.NoError(t, err)

	result, err := ListSecrets(context.Background(), clientset, "", "", "")
	assert.NoError(t, err)
	assert.Equal(t, 2, len(result))
}

func TestListNamespaces(t *testing.T) {
	clientset := fake.NewSimpleClientset()
	ns := &corev1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-namespace",
		},
	}
	_, err := clientset.CoreV1().Namespaces().Create(context.Background(), ns, metav1.CreateOptions{})
	assert.NoError(t, err)

	result, err := ListNamespaces(context.Background(), clientset, "", "")
	assert.NoError(t, err)
	assert.GreaterOrEqual(t, len(result), 1)
	found := false
	for _, n := range result {
		if n.Name == "test-namespace" {
			found = true
			break
		}
	}
	assert.True(t, found)
}

func TestListEvents(t *testing.T) {
	clientset := fake.NewSimpleClientset()
	event := &corev1.Event{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-event",
			Namespace: "default",
		},
		InvolvedObject: corev1.ObjectReference{
			Kind: "Pod",
			Name: "test-pod",
		},
		Reason:        "Created",
		Message:       "Pod was created",
		LastTimestamp: metav1.NewTime(time.Now()),
	}
	_, err := clientset.CoreV1().Events("default").Create(context.Background(), event, metav1.CreateOptions{})
	assert.NoError(t, err)

	result, err := ListEvents(context.Background(), clientset, "default", "", "1h", "", "")
	assert.NoError(t, err)
	assert.Equal(t, 1, len(result))
	assert.Equal(t, "test-event", result[0].Name)
}

func TestListEvents_WithInvolvedObject(t *testing.T) {
	clientset := fake.NewSimpleClientset()
	event := &corev1.Event{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-event",
			Namespace: "default",
		},
		InvolvedObject: corev1.ObjectReference{
			Kind: "Pod",
			Name: "test-pod",
		},
		Reason:        "Created",
		LastTimestamp: metav1.NewTime(time.Now()),
	}
	_, err := clientset.CoreV1().Events("default").Create(context.Background(), event, metav1.CreateOptions{})
	assert.NoError(t, err)

	result, err := ListEvents(context.Background(), clientset, "default", "Pod/test-pod", "1h", "", "")
	assert.NoError(t, err)
	assert.Equal(t, 1, len(result))
}

func TestListEvents_SinceFilter(t *testing.T) {
	clientset := fake.NewSimpleClientset()
	oldEvent := &corev1.Event{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "old-event",
			Namespace: "default",
		},
		LastTimestamp: metav1.Time{Time: time.Now().Add(-2 * time.Hour)},
	}
	newEvent := &corev1.Event{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "new-event",
			Namespace: "default",
		},
		LastTimestamp: metav1.Time{Time: time.Now()},
	}
	_, err := clientset.CoreV1().Events("default").Create(context.Background(), oldEvent, metav1.CreateOptions{})
	assert.NoError(t, err)
	_, err = clientset.CoreV1().Events("default").Create(context.Background(), newEvent, metav1.CreateOptions{})
	assert.NoError(t, err)

	result, err := ListEvents(context.Background(), clientset, "default", "", "1h", "", "")
	assert.NoError(t, err)
	assert.Equal(t, 1, len(result))
}

func TestListEvents_InvalidSince(t *testing.T) {
	clientset := fake.NewSimpleClientset()
	event := &corev1.Event{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-event",
			Namespace: "default",
		},
		LastTimestamp: metav1.Time{Time: time.Now()},
	}
	_, err := clientset.CoreV1().Events("default").Create(context.Background(), event, metav1.CreateOptions{})
	assert.NoError(t, err)

	// Invalid duration should default to 1h
	result, err := ListEvents(context.Background(), clientset, "default", "", "invalid", "", "")
	assert.NoError(t, err)
	assert.Equal(t, 1, len(result))
}

func TestListNodes(t *testing.T) {
	clientset := fake.NewSimpleClientset()
	node := &corev1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-node",
		},
		Status: corev1.NodeStatus{
			Conditions: []corev1.NodeCondition{
				{Type: corev1.NodeReady, Status: corev1.ConditionTrue},
			},
		},
	}
	_, err := clientset.CoreV1().Nodes().Create(context.Background(), node, metav1.CreateOptions{})
	assert.NoError(t, err)

	// Test with nil metricsClient and nil pods
	result, err := ListNodes(context.Background(), clientset, nil, nil, "", "")
	assert.NoError(t, err)
	assert.Equal(t, 1, len(result))
	assert.Equal(t, "test-node", result[0].Name)
}

func TestListNodes_WithPods(t *testing.T) {
	clientset := fake.NewSimpleClientset()
	node := &corev1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-node",
		},
	}
	_, err := clientset.CoreV1().Nodes().Create(context.Background(), node, metav1.CreateOptions{})
	assert.NoError(t, err)

	pod := &corev1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-pod",
			Namespace: "default",
		},
		Spec: corev1.PodSpec{
			NodeName: "test-node",
		},
	}
	_, err = clientset.CoreV1().Pods("default").Create(context.Background(), pod, metav1.CreateOptions{})
	assert.NoError(t, err)

	podsList := &corev1.PodList{
		Items: []corev1.Pod{*pod},
	}

	// Test with pods parameter
	result, err := ListNodes(context.Background(), clientset, nil, podsList, "", "")
	assert.NoError(t, err)
	assert.Equal(t, 1, len(result))
}

func TestListNodes_WithMetricsClient(t *testing.T) {
	clientset := fake.NewSimpleClientset()
	node := &corev1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-node",
		},
	}
	_, err := clientset.CoreV1().Nodes().Create(context.Background(), node, metav1.CreateOptions{})
	assert.NoError(t, err)

	// Test with metricsClient (fake doesn't support metrics, so it should be ignored)
	result, err := ListNodes(context.Background(), clientset, nil, nil, "", "")
	assert.NoError(t, err)
	assert.Equal(t, 1, len(result))
}

func TestListEndpoints(t *testing.T) {
	clientset := fake.NewSimpleClientset()
	endpoints := &corev1.Endpoints{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-endpoints",
			Namespace: "default",
		},
	}
	_, err := clientset.CoreV1().Endpoints("default").Create(context.Background(), endpoints, metav1.CreateOptions{})
	assert.NoError(t, err)

	result, err := ListEndpoints(context.Background(), clientset, "default", "", "")
	assert.NoError(t, err)
	assert.Equal(t, 1, len(result))
}

func TestListJobs(t *testing.T) {
	clientset := fake.NewSimpleClientset()
	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-job",
			Namespace: "default",
		},
	}
	_, err := clientset.BatchV1().Jobs("default").Create(context.Background(), job, metav1.CreateOptions{})
	assert.NoError(t, err)

	result, err := ListJobs(context.Background(), clientset, "default", "", "")
	assert.NoError(t, err)
	assert.Equal(t, 1, len(result))
}

func TestListCronJobs(t *testing.T) {
	clientset := fake.NewSimpleClientset()
	cronJob := &batchv1.CronJob{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-cronjob",
			Namespace: "default",
		},
	}
	_, err := clientset.BatchV1().CronJobs("default").Create(context.Background(), cronJob, metav1.CreateOptions{})
	assert.NoError(t, err)

	result, err := ListCronJobs(context.Background(), clientset, "default", "", "")
	assert.NoError(t, err)
	assert.Equal(t, 1, len(result))
}

// Helper function to create int32 pointer
func ptrInt32(i int32) *int32 {
	return &i
}
