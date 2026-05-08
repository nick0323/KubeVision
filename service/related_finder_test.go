package service

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/fake"
)

func TestMatchesSelector(t *testing.T) {
	tests := []struct {
		name     string
		labels   map[string]string
		selector map[string]string
		want     bool
	}{
		{"empty selector", map[string]string{"app": "test"}, map[string]string{}, false},
		{"nil selector", map[string]string{"app": "test"}, nil, false},
		{"exact match", map[string]string{"app": "test"}, map[string]string{"app": "test"}, true},
		{"mismatch value", map[string]string{"app": "test"}, map[string]string{"app": "other"}, false},
		{"missing key", map[string]string{"app": "test"}, map[string]string{"tier": "web"}, false},
		{"partial match", map[string]string{"app": "test", "tier": "web"}, map[string]string{"app": "test"}, true},
		{"multiple selectors", map[string]string{"app": "test", "tier": "web"}, map[string]string{"app": "test", "tier": "web"}, true},
		{"nil labels", nil, map[string]string{"app": "test"}, false},
		{"both nil", nil, nil, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := matchesSelector(tt.labels, tt.selector)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestFindRelatedResources_Pod_OwnerRefs(t *testing.T) {
	clientset := fake.NewSimpleClientset()
	logger := zap.NewNop()

	pod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-pod",
			Namespace: "default",
			OwnerReferences: []metav1.OwnerReference{
				{Kind: "ReplicaSet", Name: "test-rs"},
			},
		},
		Spec: v1.PodSpec{
			NodeName: "worker-1",
			Volumes:  []v1.Volume{},
		},
	}

	results := FindRelatedResources(pod, "pod", "default", clientset, context.Background(), logger)
	assert.GreaterOrEqual(t, len(results), 1)

	found := false
	for _, r := range results {
		item, ok := r.(map[string]string)
		if ok && item["kind"] == "ReplicaSet" && item["name"] == "test-rs" {
			found = true
			assert.Equal(t, "owner", item["relation"])
		}
	}
	assert.True(t, found, "should find owner ReplicaSet")
}

func TestFindRelatedResources_Pod_Node(t *testing.T) {
	clientset := fake.NewSimpleClientset()
	logger := zap.NewNop()

	pod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-pod",
			Namespace: "default",
		},
		Spec: v1.PodSpec{
			NodeName: "worker-1",
			Volumes:  []v1.Volume{},
		},
	}

	results := FindRelatedResources(pod, "pod", "default", clientset, context.Background(), logger)

	found := false
	for _, r := range results {
		item, ok := r.(map[string]string)
		if ok && item["kind"] == "Node" && item["name"] == "worker-1" {
			found = true
			assert.Equal(t, "scheduledOn", item["relation"])
		}
	}
	assert.True(t, found, "should find Node worker-1")
}

func TestFindRelatedResources_Pod_VolumeRefs(t *testing.T) {
	clientset := fake.NewSimpleClientset()
	logger := zap.NewNop()

	pod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-pod",
			Namespace: "default",
		},
		Spec: v1.PodSpec{
			NodeName: "",
			Volumes: []v1.Volume{
				{Name: "cfg", VolumeSource: v1.VolumeSource{ConfigMap: &v1.ConfigMapVolumeSource{LocalObjectReference: v1.LocalObjectReference{Name: "my-config"}}}},
				{Name: "sec", VolumeSource: v1.VolumeSource{Secret: &v1.SecretVolumeSource{SecretName: "my-secret"}}},
				{Name: "data", VolumeSource: v1.VolumeSource{PersistentVolumeClaim: &v1.PersistentVolumeClaimVolumeSource{ClaimName: "my-pvc"}}},
			},
		},
	}

	results := FindRelatedResources(pod, "pod", "default", clientset, context.Background(), logger)
	assert.GreaterOrEqual(t, len(results), 3)

	types := make(map[string]string)
	for _, r := range results {
		item, ok := r.(map[string]string)
		if ok {
			types[item["kind"]] = item["name"]
		}
	}
	assert.Equal(t, "my-config", types["ConfigMap"])
	assert.Equal(t, "my-secret", types["Secret"])
	assert.Equal(t, "my-pvc", types["PersistentVolumeClaim"])
}

func TestFindRelatedResources_Pod_MaxLimit(t *testing.T) {
	clientset := fake.NewSimpleClientset()
	logger := zap.NewNop()

	owners := make([]metav1.OwnerReference, 150)
	for i := 0; i < 150; i++ {
		owners[i] = metav1.OwnerReference{Kind: "ReplicaSet", Name: "rs"}
	}

	pod := &v1.Pod{
		ObjectMeta: metav1.ObjectMeta{
			Name:            "test-pod",
			Namespace:       "default",
			OwnerReferences: owners,
		},
	}

	results := FindRelatedResources(pod, "pod", "default", clientset, context.Background(), logger)
	assert.LessOrEqual(t, len(results), maxRelatedResources)
}

func TestFindRelatedResources_Deployment(t *testing.T) {
	clientset := fake.NewSimpleClientset()
	logger := zap.NewNop()

	dep := &appsv1.Deployment{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-dep",
			Namespace: "default",
			Labels:    map[string]string{"app": "test"},
		},
		Spec: appsv1.DeploymentSpec{
			Selector: &metav1.LabelSelector{MatchLabels: map[string]string{"app": "test"}},
		},
	}

	results := FindRelatedResources(dep, "deployment", "default", clientset, context.Background(), logger)
	assert.NotNil(t, results)
}

func TestFindRelatedResources_Service(t *testing.T) {
	clientset := fake.NewSimpleClientset()
	logger := zap.NewNop()

	svc := &v1.Service{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-svc",
			Namespace: "default",
		},
		Spec: v1.ServiceSpec{
			Selector: map[string]string{"app": "test"},
		},
	}

	results := FindRelatedResources(svc, "service", "default", clientset, context.Background(), logger)
	assert.NotNil(t, results)
}

func TestFindRelatedResources_ConfigMap(t *testing.T) {
	clientset := fake.NewSimpleClientset()
	logger := zap.NewNop()

	cm := &v1.ConfigMap{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-cm",
			Namespace: "default",
		},
	}

	results := FindRelatedResources(cm, "configmap", "default", clientset, context.Background(), logger)
	assert.NotNil(t, results)
}

func TestFindRelatedResources_Secret(t *testing.T) {
	clientset := fake.NewSimpleClientset()
	logger := zap.NewNop()

	sec := &v1.Secret{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-secret",
			Namespace: "default",
		},
	}

	results := FindRelatedResources(sec, "secret", "default", clientset, context.Background(), logger)
	assert.NotNil(t, results)
}

func TestFindRelatedResources_PVC(t *testing.T) {
	clientset := fake.NewSimpleClientset()
	logger := zap.NewNop()

	pvc := &v1.PersistentVolumeClaim{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-pvc",
			Namespace: "default",
		},
		Spec: v1.PersistentVolumeClaimSpec{
			VolumeName: "test-pv",
		},
	}

	results := FindRelatedResources(pvc, "persistentvolumeclaim", "default", clientset, context.Background(), logger)
	assert.NotNil(t, results)
}

func TestFindRelatedResources_PV(t *testing.T) {
	clientset := fake.NewSimpleClientset()
	logger := zap.NewNop()

	pv := &v1.PersistentVolume{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-pv",
		},
		Spec: v1.PersistentVolumeSpec{
			ClaimRef: &v1.ObjectReference{Name: "test-pvc", Namespace: "default"},
		},
	}

	results := FindRelatedResources(pv, "persistentvolume", "default", clientset, context.Background(), logger)
	assert.NotNil(t, results)
}

func TestFindRelatedResources_Ingress(t *testing.T) {
	clientset := fake.NewSimpleClientset()
	logger := zap.NewNop()

	ing := &networkingv1.Ingress{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-ing",
			Namespace: "default",
		},
		Spec: networkingv1.IngressSpec{
			Rules: []networkingv1.IngressRule{
				{
					IngressRuleValue: networkingv1.IngressRuleValue{
						HTTP: &networkingv1.HTTPIngressRuleValue{
							Paths: []networkingv1.HTTPIngressPath{
								{
									Backend: networkingv1.IngressBackend{
										Service: &networkingv1.IngressServiceBackend{
											Name: "test-svc",
										},
									},
								},
							},
						},
					},
				},
			},
		},
	}

	results := FindRelatedResources(ing, "ingress", "default", clientset, context.Background(), logger)
	assert.NotNil(t, results)
}

func TestFindRelatedResources_StatefulSet(t *testing.T) {
	clientset := fake.NewSimpleClientset()
	logger := zap.NewNop()

	sts := &appsv1.StatefulSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-sts",
			Namespace: "default",
			Labels:    map[string]string{"app": "test"},
		},
		Spec: appsv1.StatefulSetSpec{
			VolumeClaimTemplates: []v1.PersistentVolumeClaim{
				{ObjectMeta: metav1.ObjectMeta{Name: "data-vol"}},
			},
		},
	}

	results := FindRelatedResources(sts, "statefulset", "default", clientset, context.Background(), logger)
	assert.NotNil(t, results)
}

func TestFindRelatedResources_DaemonSet(t *testing.T) {
	clientset := fake.NewSimpleClientset()
	logger := zap.NewNop()

	ds := &appsv1.DaemonSet{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-ds",
			Namespace: "default",
			Labels:    map[string]string{"app": "test"},
		},
	}

	results := FindRelatedResources(ds, "daemonset", "default", clientset, context.Background(), logger)
	assert.NotNil(t, results)
}

func TestFindRelatedResources_Job(t *testing.T) {
	clientset := fake.NewSimpleClientset()
	logger := zap.NewNop()

	job := &batchv1.Job{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-job",
			Namespace: "default",
			Labels:    map[string]string{"app": "test"},
		},
	}

	results := FindRelatedResources(job, "job", "default", clientset, context.Background(), logger)
	assert.NotNil(t, results)
}

func TestFindRelatedResources_CronJob(t *testing.T) {
	clientset := fake.NewSimpleClientset()
	logger := zap.NewNop()

	cj := &batchv1.CronJob{
		ObjectMeta: metav1.ObjectMeta{
			Name:      "test-cj",
			Namespace: "default",
		},
	}

	results := FindRelatedResources(cj, "cronjob", "default", clientset, context.Background(), logger)
	assert.NotNil(t, results)
}

func TestFindRelatedResources_Node(t *testing.T) {
	clientset := fake.NewSimpleClientset()
	logger := zap.NewNop()

	node := &v1.Node{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-node",
		},
	}

	results := FindRelatedResources(node, "node", "", clientset, context.Background(), logger)
	assert.NotNil(t, results)
}

func TestFindRelatedResources_Namespace(t *testing.T) {
	clientset := fake.NewSimpleClientset()
	logger := zap.NewNop()

	ns := &v1.Namespace{
		ObjectMeta: metav1.ObjectMeta{
			Name: "test-ns",
		},
	}

	results := FindRelatedResources(ns, "namespace", "", clientset, context.Background(), logger)
	assert.NotNil(t, results)
}

func TestFindRelatedResources_UnknownType(t *testing.T) {
	clientset := fake.NewSimpleClientset()
	logger := zap.NewNop()

	results := FindRelatedResources(nil, "unknown", "default", clientset, context.Background(), logger)
	assert.Empty(t, results)
}
