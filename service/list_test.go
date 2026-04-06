package service

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	storagev1 "k8s.io/api/storage/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ==================== MapPods 测试 ====================

func TestMapPods(t *testing.T) {
	pods := []corev1.Pod{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "test-pod",
				Namespace: "default",
			},
			Status: corev1.PodStatus{
				Phase: corev1.PodRunning,
				PodIP: "10.244.0.5",
				ContainerStatuses: []corev1.ContainerStatus{
					{
						Name:         "nginx",
						Ready:        true,
						RestartCount: 2,
					},
					{
						Name:         "sidecar",
						Ready:        false,
						RestartCount: 1,
					},
				},
			},
			Spec: corev1.PodSpec{
				NodeName: "node-1",
			},
		},
	}

	result := MapPods(pods, nil)

	assert.Len(t, result, 1)
	assert.Equal(t, "test-pod", result[0].Name)
	assert.Equal(t, "default", result[0].Namespace)
	assert.Equal(t, "Running", result[0].Status)
	assert.Equal(t, "1/2", result[0].Ready)
	assert.Equal(t, int32(3), result[0].Restarts)
	assert.Equal(t, "10.244.0.5", result[0].PodIP)
	assert.Equal(t, "node-1", result[0].NodeName)
}

func TestMapPods_EmptyList(t *testing.T) {
	result := MapPods([]corev1.Pod{}, nil)
	assert.Empty(t, result)
}

func TestMapPods_AllContainersReady(t *testing.T) {
	pods := []corev1.Pod{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "ready-pod",
				Namespace: "default",
			},
			Status: corev1.PodStatus{
				Phase: corev1.PodRunning,
				ContainerStatuses: []corev1.ContainerStatus{
					{Name: "c1", Ready: true, RestartCount: 0},
					{Name: "c2", Ready: true, RestartCount: 0},
				},
			},
			Spec: corev1.PodSpec{
				NodeName: "node-1",
			},
		},
	}

	result := MapPods(pods, nil)
	assert.Equal(t, "2/2", result[0].Ready)
	assert.Equal(t, int32(0), result[0].Restarts)
}

// ==================== MapDeployments 测试 ====================

func TestMapDeployments(t *testing.T) {
	replicas := int32(3)
	deployments := []appsv1.Deployment{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "nginx-deployment",
				Namespace: "default",
			},
			Spec: appsv1.DeploymentSpec{
				Replicas: &replicas,
			},
			Status: appsv1.DeploymentStatus{
				Replicas:          3,
				ReadyReplicas:     2,
				UpdatedReplicas:   3,
				AvailableReplicas: 2,
			},
		},
	}

	result := MapDeployments(deployments)

	assert.Len(t, result, 1)
	assert.Equal(t, "nginx-deployment", result[0].Name)
	assert.Equal(t, int32(2), result[0].ReadyReplicas)
	assert.Equal(t, int32(3), result[0].UpdatedReplicas)
	assert.Equal(t, int32(2), result[0].Available)
	assert.Equal(t, int32(3), result[0].Desired)
	assert.Equal(t, "Partial", result[0].Status)
}

func TestMapDeployments_Healthy(t *testing.T) {
	replicas := int32(3)
	deployments := []appsv1.Deployment{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "healthy-deployment",
				Namespace: "default",
			},
			Spec: appsv1.DeploymentSpec{
				Replicas: &replicas,
			},
			Status: appsv1.DeploymentStatus{
				Replicas:          3,
				ReadyReplicas:     3,
				UpdatedReplicas:   3,
				AvailableReplicas: 3,
			},
		},
	}

	result := MapDeployments(deployments)
	assert.Equal(t, "Available", result[0].Status)
}

func TestMapDeployments_NilReplicas(t *testing.T) {
	deployments := []appsv1.Deployment{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "default-replicas",
				Namespace: "default",
			},
			Spec: appsv1.DeploymentSpec{
				Replicas: nil,
			},
			Status: appsv1.DeploymentStatus{
				Replicas:          1,
				ReadyReplicas:     1,
				UpdatedReplicas:   1,
				AvailableReplicas: 1,
			},
		},
	}

	result := MapDeployments(deployments)
	assert.Equal(t, int32(1), result[0].Desired)
	assert.Equal(t, "Available", result[0].Status)
}

// ==================== MapServices 测试 ====================

func TestMapServices(t *testing.T) {
	services := []corev1.Service{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "nginx-svc",
				Namespace: "default",
			},
			Spec: corev1.ServiceSpec{
				Type:      corev1.ServiceTypeClusterIP,
				ClusterIP: "10.96.100.100",
				Ports: []corev1.ServicePort{
					{Name: "http", Port: 80},
					{Name: "https", Port: 443},
				},
			},
		},
	}

	result := MapServices(services)

	assert.Len(t, result, 1)
	assert.Equal(t, "nginx-svc", result[0].Name)
	assert.Equal(t, "ClusterIP", result[0].Type)
	assert.Equal(t, "10.96.100.100", result[0].ClusterIP)
	assert.Len(t, result[0].Ports, 2)
}

func TestMapServices_LoadBalancer(t *testing.T) {
	services := []corev1.Service{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "lb-svc",
				Namespace: "default",
			},
			Spec: corev1.ServiceSpec{
				Type: corev1.ServiceTypeLoadBalancer,
			},
			Status: corev1.ServiceStatus{
				LoadBalancer: corev1.LoadBalancerStatus{
					Ingress: []corev1.LoadBalancerIngress{
						{IP: "1.2.3.4"},
					},
				},
			},
		},
	}

	result := MapServices(services)
	assert.Equal(t, "LoadBalancer", result[0].Type)
}

// ==================== MapConfigMaps 测试 ====================

func TestMapConfigMaps(t *testing.T) {
	configMaps := []corev1.ConfigMap{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "nginx-config",
				Namespace: "default",
			},
			Data: map[string]string{
				"nginx.conf":   "server {...}",
				"mime.types":   "types {...}",
				"default.conf": "server {...}",
			},
		},
	}

	result := MapConfigMaps(configMaps)

	assert.Len(t, result, 1)
	assert.Equal(t, "nginx-config", result[0].Name)
	assert.Equal(t, 3, result[0].DataCount)
	assert.Len(t, result[0].Keys, 3)
	assert.Contains(t, result[0].Keys, "nginx.conf")
}

// ==================== MapSecrets 测试 ====================

func TestMapSecrets(t *testing.T) {
	secrets := []corev1.Secret{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "tls-secret",
				Namespace: "default",
			},
			Type: corev1.SecretTypeTLS,
			Data: map[string][]byte{
				"tls.crt": []byte("certificate"),
				"tls.key": []byte("key"),
			},
		},
	}

	result := MapSecrets(secrets)

	assert.Len(t, result, 1)
	assert.Equal(t, "tls-secret", result[0].Name)
	assert.Equal(t, "kubernetes.io/tls", result[0].Type)
	assert.Equal(t, 2, result[0].DataCount)
	assert.Contains(t, result[0].Keys, "tls.crt")
	assert.Contains(t, result[0].Keys, "tls.key")
}

// ==================== MapNodes 测试 ====================

func TestMapNodes(t *testing.T) {
	nodes := []corev1.Node{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "node-1",
			},
			Status: corev1.NodeStatus{
				Addresses: []corev1.NodeAddress{
					{Type: corev1.NodeInternalIP, Address: "192.168.1.10"},
				},
				Conditions: []corev1.NodeCondition{
					{Type: corev1.NodeReady, Status: corev1.ConditionTrue},
				},
				Capacity: corev1.ResourceList{
					corev1.ResourcePods: *resource.NewQuantity(110, resource.DecimalSI),
				},
			},
		},
	}

	result := MapNodes(nodes, nil, nil)

	assert.Len(t, result, 1)
	assert.Equal(t, "node-1", result[0].Name)
	assert.Equal(t, "192.168.1.10", result[0].IP)
	assert.Equal(t, "Ready", result[0].Status)
	assert.Contains(t, result[0].Role, "worker")
	assert.Equal(t, 110, result[0].PodsCapacity)
}

func TestMapNodes_ControlPlane(t *testing.T) {
	nodes := []corev1.Node{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "master-1",
				Labels: map[string]string{
					"node-role.kubernetes.io/control-plane": "",
					"node-role.kubernetes.io/master":        "",
				},
			},
			Status: corev1.NodeStatus{
				Addresses: []corev1.NodeAddress{
					{Type: corev1.NodeInternalIP, Address: "192.168.1.1"},
				},
				Conditions: []corev1.NodeCondition{
					{Type: corev1.NodeReady, Status: corev1.ConditionTrue},
				},
				Capacity: corev1.ResourceList{
					corev1.ResourcePods: *resource.NewQuantity(110, resource.DecimalSI),
				},
			},
		},
	}

	result := MapNodes(nodes, nil, nil)

	assert.Contains(t, result[0].Role, "control-plane")
}

func TestMapNodes_NotReady(t *testing.T) {
	nodes := []corev1.Node{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "node-down",
			},
			Status: corev1.NodeStatus{
				Conditions: []corev1.NodeCondition{
					{Type: corev1.NodeReady, Status: corev1.ConditionFalse},
				},
			},
		},
	}

	result := MapNodes(nodes, nil, nil)
	assert.Equal(t, "NotReady", result[0].Status)
}

// ==================== MapIngresses 测试 ====================

func TestMapIngresses(t *testing.T) {
	ingresses := []networkingv1.Ingress{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "nginx-ingress",
				Namespace: "default",
			},
			Spec: networkingv1.IngressSpec{
				IngressClassName: strPtr("nginx"),
				Rules: []networkingv1.IngressRule{
					{
						Host: "example.com",
						IngressRuleValue: networkingv1.IngressRuleValue{
							HTTP: &networkingv1.HTTPIngressRuleValue{
								Paths: []networkingv1.HTTPIngressPath{
									{
										Path: "/",
										Backend: networkingv1.IngressBackend{
											Service: &networkingv1.IngressServiceBackend{
												Name: "nginx-svc",
											},
										},
									},
								},
							},
						},
					},
				},
			},
			Status: networkingv1.IngressStatus{
				LoadBalancer: networkingv1.IngressLoadBalancerStatus{
					Ingress: []networkingv1.IngressLoadBalancerIngress{
						{IP: "1.2.3.4"},
					},
				},
			},
		},
	}

	result := MapIngresses(ingresses)

	assert.Len(t, result, 1)
	assert.Equal(t, "nginx-ingress", result[0].Name)
	assert.Equal(t, "nginx", result[0].Class)
	assert.Contains(t, result[0].Hosts, "example.com")
	assert.Equal(t, "Ready", result[0].Status)
}

// ==================== MapJobs 测试 ====================

func TestMapJobs_Completed(t *testing.T) {
	jobs := []batchv1.Job{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "backup-job",
				Namespace: "default",
			},
			Spec: batchv1.JobSpec{
				Completions: int32Ptr(1),
			},
			Status: batchv1.JobStatus{
				Succeeded: 1,
				Failed:    0,
				StartTime: &metav1.Time{Time: time.Now().Add(-10 * time.Minute)},
			},
		},
	}

	result := MapJobs(jobs)

	assert.Len(t, result, 1)
	assert.Equal(t, "backup-job", result[0].Name)
	assert.Equal(t, int32(1), result[0].Completions)
	assert.Equal(t, int32(1), result[0].Succeeded)
	assert.Equal(t, "Completed", result[0].Status)
}

func TestMapJobs_Failed(t *testing.T) {
	jobs := []batchv1.Job{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "failed-job",
				Namespace: "default",
			},
			Spec: batchv1.JobSpec{
				Completions: int32Ptr(1),
			},
			Status: batchv1.JobStatus{
				Succeeded: 0,
				Failed:    3,
			},
		},
	}

	result := MapJobs(jobs)
	assert.Equal(t, "Failed", result[0].Status)
}

// ==================== MapCronJobs 测试 ====================

func TestMapCronJobs(t *testing.T) {
	cronJobs := []batchv1.CronJob{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "daily-backup",
				Namespace: "default",
			},
			Spec: batchv1.CronJobSpec{
				Schedule: "0 2 * * *",
				Suspend:  boolPtr(false),
			},
			Status: batchv1.CronJobStatus{
				Active: []corev1.ObjectReference{},
			},
		},
	}

	result := MapCronJobs(cronJobs)

	assert.Len(t, result, 1)
	assert.Equal(t, "daily-backup", result[0].Name)
	assert.Equal(t, "0 2 * * *", result[0].Schedule)
	assert.False(t, result[0].Suspend)
	assert.Equal(t, 0, result[0].Active)
	assert.Equal(t, "Active", result[0].Status)
}

func TestMapCronJobs_Suspended(t *testing.T) {
	cronJobs := []batchv1.CronJob{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "suspended-job",
				Namespace: "default",
			},
			Spec: batchv1.CronJobSpec{
				Schedule: "0 2 * * *",
				Suspend:  boolPtr(true),
			},
		},
	}

	result := MapCronJobs(cronJobs)
	assert.Equal(t, "Suspended", result[0].Status)
}

// ==================== MapPVCs 测试 ====================

func TestMapPVCs_Bound(t *testing.T) {
	pvcs := []corev1.PersistentVolumeClaim{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "mysql-data",
				Namespace: "default",
			},
			Spec: corev1.PersistentVolumeClaimSpec{
				AccessModes: []corev1.PersistentVolumeAccessMode{
					corev1.ReadWriteOnce,
				},
				StorageClassName: strPtr("standard"),
				VolumeName:       "pvc-123456",
			},
			Status: corev1.PersistentVolumeClaimStatus{
				Phase: corev1.ClaimBound,
				Capacity: corev1.ResourceList{
					corev1.ResourceStorage: *resource.NewQuantity(10*1024*1024*1024, resource.BinarySI),
				},
			},
		},
	}

	result := MapPVCs(pvcs)

	assert.Len(t, result, 1)
	assert.Equal(t, "mysql-data", result[0].Name)
	assert.Equal(t, "Bound", result[0].Status)
	assert.Equal(t, "10Gi", result[0].Capacity)
	assert.Equal(t, "ReadWriteOnce", result[0].AccessMode)
	assert.Equal(t, "standard", result[0].StorageClass)
}

// ==================== MapPVs 测试 ====================

func TestMapPVs(t *testing.T) {
	pvs := []corev1.PersistentVolume{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "pv-001",
			},
			Spec: corev1.PersistentVolumeSpec{
				Capacity: corev1.ResourceList{
					corev1.ResourceStorage: *resource.NewQuantity(100*1024*1024*1024, resource.BinarySI),
				},
				AccessModes: []corev1.PersistentVolumeAccessMode{
					corev1.ReadWriteOnce,
				},
				StorageClassName:              "standard",
				PersistentVolumeReclaimPolicy: corev1.PersistentVolumeReclaimRetain,
				ClaimRef: &corev1.ObjectReference{
					Namespace: "default",
					Name:      "mysql-data",
				},
			},
			Status: corev1.PersistentVolumeStatus{
				Phase: corev1.VolumeBound,
			},
		},
	}

	result := MapPVs(pvs)

	assert.Len(t, result, 1)
	assert.Equal(t, "pv-001", result[0].Name)
	assert.Equal(t, "Bound", result[0].Status)
	assert.Equal(t, "100Gi", result[0].Capacity)
	assert.Equal(t, "ReadWriteOnce", result[0].AccessMode)
	assert.Equal(t, "default/mysql-data", result[0].ClaimRef)
	assert.Equal(t, "Retain", result[0].ReclaimPolicy)
}

// ==================== MapStorageClasses 测试 ====================

func TestMapStorageClasses(t *testing.T) {
	reclaimPolicy := corev1.PersistentVolumeReclaimRetain
	storageClasses := []storagev1.StorageClass{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "standard",
				Annotations: map[string]string{
					"storageclass.kubernetes.io/is-default-class": "true",
				},
			},
			Provisioner:       "kubernetes.io/no-provisioner",
			ReclaimPolicy:     &reclaimPolicy,
			VolumeBindingMode: bindingModePtr(storagev1.VolumeBindingWaitForFirstConsumer),
		},
	}

	result := MapStorageClasses(storageClasses)

	assert.Len(t, result, 1)
	assert.Equal(t, "standard", result[0].Name)
	assert.Equal(t, "kubernetes.io/no-provisioner", result[0].Provisioner)
	assert.Equal(t, "Retain", result[0].ReclaimPolicy)
	assert.Equal(t, "WaitForFirstConsumer", result[0].VolumeBindingMode)
	assert.True(t, result[0].IsDefault)
}

// ==================== MapNamespaces 测试 ====================

func TestMapNamespaces(t *testing.T) {
	namespaces := []corev1.Namespace{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "default",
			},
			Status: corev1.NamespaceStatus{
				Phase: corev1.NamespaceActive,
			},
		},
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "terminating-ns",
			},
			Status: corev1.NamespaceStatus{
				Phase: corev1.NamespaceTerminating,
			},
		},
	}

	result := MapNamespaces(namespaces)

	assert.Len(t, result, 2)
	assert.Equal(t, "Active", result[0].Status)
	assert.Equal(t, "Terminating", result[1].Status)
}

// ==================== MapEndpoints 测试 ====================

func TestMapEndpoints(t *testing.T) {
	endpoints := []corev1.Endpoints{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "nginx-svc",
				Namespace: "default",
			},
			Subsets: []corev1.EndpointSubset{
				{
					Addresses: []corev1.EndpointAddress{
						{IP: "10.244.0.5"},
						{IP: "10.244.0.6"},
						{IP: "10.244.0.7"},
					},
					Ports: []corev1.EndpointPort{
						{Name: "http", Port: 80},
						{Name: "https", Port: 443},
					},
				},
			},
		},
	}

	result := MapEndpoints(endpoints)

	assert.Len(t, result, 1)
	assert.Equal(t, "nginx-svc", result[0].Name)
	assert.Equal(t, 3, result[0].Addresses)
	assert.Len(t, result[0].Ports, 2)
}

// ==================== MapEvents 测试 ====================

func TestMapEvents(t *testing.T) {
	events := []corev1.Event{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "event-1",
				Namespace: "default",
			},
			Reason:  "Scheduled",
			Message: "Successfully assigned default/nginx-pod to node-1",
			Type:    corev1.EventTypeNormal,
			Count:   1,
			Source:  corev1.EventSource{Component: "default-scheduler"},
			InvolvedObject: corev1.ObjectReference{
				Kind:      "Pod",
				Name:      "nginx-pod",
				Namespace: "default",
			},
		},
	}

	result := MapEvents(events)

	assert.Len(t, result, 1)
	assert.Equal(t, "Scheduled", result[0].Reason)
	assert.Equal(t, "Normal", result[0].Type)
	assert.Equal(t, int32(1), result[0].Count)
	assert.Equal(t, "Pod/nginx-pod", result[0].Object)
	assert.Equal(t, "default-scheduler", result[0].Source)
}

// ==================== 辅助函数 ====================

func strPtr(s string) *string {
	return &s
}

func int32Ptr(i int32) *int32 {
	return &i
}

func boolPtr(b bool) *bool {
	return &b
}

func bindingModePtr(m storagev1.VolumeBindingMode) *storagev1.VolumeBindingMode {
	return &m
}
