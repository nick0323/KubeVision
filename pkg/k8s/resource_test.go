package k8s

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestResourceTypeNormalize(t *testing.T) {
	tests := []struct {
		input    ResourceType
		expected ResourceType
	}{
		{"pod", ResourcePod},
		{"deployment", ResourceDeployment},
		{"statefulset", ResourceStatefulSet},
		{"pvc", ResourcePVC},
		{"pv", ResourcePV},
		{"sc", ResourceStorageClass},
		{"service", ResourceService},
		{"configmap", ResourceConfigMap},
		{"secret", ResourceSecret},
		{"ingress", ResourceIngress},
		{"job", ResourceJob},
		{"cronjob", ResourceCronJob},
		{"namespace", ResourceNamespace},
		{"node", ResourceNode},
		{"endpoint", ResourceEndpoint},
		{"event", ResourceEvent},

	}

	for _, tt := range tests {
		t.Run(string(tt.input), func(t *testing.T) {
			result := tt.input.Normalize()
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestResourceTypeNormalizeUnknown(t *testing.T) {
	result := ResourceType("unknown").Normalize()
	assert.Equal(t, ResourceType("unknown"), result)
}

func TestResourceTypeIsClusterScoped(t *testing.T) {
	cluster := []ResourceType{ResourcePV, ResourceStorageClass, ResourceNamespace, ResourceNode}
	namespaced := []ResourceType{ResourcePod, ResourceDeployment, ResourceStatefulSet, ResourceDaemonSet, ResourceService, ResourceConfigMap, ResourceSecret, ResourceIngress, ResourceJob, ResourceCronJob, ResourcePVC, ResourceEndpoint, ResourceEvent}

	for _, rt := range cluster {
		t.Run(string(rt), func(t *testing.T) {
			assert.True(t, rt.IsClusterScoped())
		})
	}

	for _, rt := range namespaced {
		t.Run(string(rt), func(t *testing.T) {
			assert.False(t, rt.IsClusterScoped())
		})
	}
}

func TestResourceTypeIsClusterScopedUnknown(t *testing.T) {
	assert.False(t, ResourceType("unknown").IsClusterScoped())
}

func TestGetKindByResourceType(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{"pod", "Pod"},
		{"deployment", "Deployment"},
		{"statefulset", "StatefulSet"},
		{"daemonset", "DaemonSet"},
		{"service", "Service"},
		{"configmap", "ConfigMap"},
		{"secret", "Secret"},
		{"ingress", "Ingress"},
		{"job", "Job"},
		{"cronjob", "CronJob"},
		{"persistentvolumeclaim", "PersistentVolumeClaim"},
		{"pvc", "PersistentVolumeClaim"},
		{"persistentvolume", "PersistentVolume"},
		{"pv", "PersistentVolume"},
		{"storageclass", "StorageClass"},
		{"sc", "StorageClass"},
		{"namespace", "Namespace"},
		{"node", "Node"},
		{"endpoint", "Endpoint"},
		{"event", "Event"},
		{"unknown", ""},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := GetKindByResourceType(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestNewGetters(t *testing.T) {
	getters := NewGetters(nil)
	assert.Len(t, getters, 17)

	expectedTypes := []ResourceType{
		ResourcePod, ResourceDeployment, ResourceStatefulSet, ResourceDaemonSet,
		ResourceService, ResourceConfigMap, ResourceSecret, ResourceIngress,
		ResourceJob, ResourceCronJob, ResourcePVC, ResourcePV,
		ResourceStorageClass, ResourceNamespace, ResourceNode, ResourceEndpoint,
		ResourceEvent,
	}

	for _, rt := range expectedTypes {
		t.Run(string(rt), func(t *testing.T) {
			getter, ok := getters[rt]
			assert.True(t, ok, "missing getter for %s", rt)
			assert.NotNil(t, getter)
		})
	}
}

func TestNewUpdaters(t *testing.T) {
	updaters := NewUpdaters(nil)
	assert.Len(t, updaters, 15)

	expectedTypes := []ResourceType{
		ResourcePod, ResourceDeployment, ResourceStatefulSet, ResourceDaemonSet,
		ResourceService, ResourceConfigMap, ResourceSecret, ResourceIngress,
		ResourceJob, ResourceCronJob, ResourcePVC, ResourcePV,
		ResourceStorageClass, ResourceNamespace, ResourceNode,
	}

	for _, rt := range expectedTypes {
		t.Run(string(rt), func(t *testing.T) {
			updater, ok := updaters[rt]
			assert.True(t, ok, "missing updater for %s", rt)
			assert.NotNil(t, updater)
		})
	}
}

func TestNewDeleters(t *testing.T) {
	deleters := NewDeleters(nil)
	assert.Len(t, deleters, 15)

	expectedTypes := []ResourceType{
		ResourcePod, ResourceDeployment, ResourceStatefulSet, ResourceDaemonSet,
		ResourceService, ResourceConfigMap, ResourceSecret, ResourceIngress,
		ResourceJob, ResourceCronJob, ResourcePVC, ResourcePV,
		ResourceStorageClass, ResourceNamespace, ResourceNode,
	}

	for _, rt := range expectedTypes {
		t.Run(string(rt), func(t *testing.T) {
			deleter, ok := deleters[rt]
			assert.True(t, ok, "missing deleter for %s", rt)
			assert.NotNil(t, deleter)
		})
	}
}

func TestNewCreators(t *testing.T) {
	creators := NewCreators(nil)
	assert.Len(t, creators, 15)

	expectedTypes := []ResourceType{
		ResourcePod, ResourceDeployment, ResourceStatefulSet, ResourceDaemonSet,
		ResourceService, ResourceConfigMap, ResourceSecret, ResourceIngress,
		ResourceJob, ResourceCronJob, ResourcePVC, ResourcePV,
		ResourceStorageClass, ResourceNamespace, ResourceNode,
	}

	for _, rt := range expectedTypes {
		t.Run(string(rt), func(t *testing.T) {
			creator, ok := creators[rt]
			assert.True(t, ok, "missing creator for %s", rt)
			assert.NotNil(t, creator)
		})
	}
}


