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
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/api/resource"
)

func TestMapPods(t *testing.T) {
	tests := []struct {
		name     string
		pods     []corev1.Pod
		expected int
	}{
		{
			name: "single pod",
			pods: []corev1.Pod{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "pod-1",
						Namespace: "default",
						Labels:    map[string]string{"app": "nginx"},
					},
					Status: corev1.PodStatus{
						Phase: corev1.PodRunning,
						PodIP: "10.0.0.1",
						ContainerStatuses: []corev1.ContainerStatus{
							{Ready: true, RestartCount: 0, Name: "nginx"},
						},
					},
				},
			},
			expected: 1,
		},
		{
			name:     "empty pods",
			pods:     []corev1.Pod{},
			expected: 0,
		},
		{
			name: "pod with multiple containers",
			pods: []corev1.Pod{
				{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "multi-container-pod",
						Namespace: "default",
					},
					Status: corev1.PodStatus{
						Phase: corev1.PodRunning,
						ContainerStatuses: []corev1.ContainerStatus{
							{Ready: true, RestartCount: 1, Name: "container-1"},
							{Ready: false, RestartCount: 2, Name: "container-2"},
						},
					},
				},
			},
			expected: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := MapPods(tt.pods)
			assert.Equal(t, tt.expected, len(result))
			if len(result) > 0 {
				assert.NotEmpty(t, result[0].Name)
				assert.NotEmpty(t, result[0].Namespace)
			}
		})
	}
}

func TestMapDeployments(t *testing.T) {
	deployments := []appsv1.Deployment{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "deploy-1",
				Namespace: "default",
			},
			Spec: appsv1.DeploymentSpec{
				Replicas: ptrInt32(3),
			},
			Status: appsv1.DeploymentStatus{
				ReadyReplicas: 2,
			},
		},
	}

	result := MapDeployments(deployments)
	assert.Equal(t, 1, len(result))
	assert.Equal(t, "deploy-1", result[0].Name)
}

func TestMapStatefulSets(t *testing.T) {
	sts := []appsv1.StatefulSet{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "sts-1",
				Namespace: "default",
			},
			Spec: appsv1.StatefulSetSpec{
				Replicas: ptrInt32(2),
			},
			Status: appsv1.StatefulSetStatus{
				ReadyReplicas: 2,
			},
		},
	}

	result := MapStatefulSets(sts)
	assert.Equal(t, 1, len(result))
	assert.Equal(t, "sts-1", result[0].Name)
}

func TestMapNodes(t *testing.T) {
	nodes := []corev1.Node{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "node-1",
				Labels: map[string]string{
					"kubernetes.io/hostname": "node-1",
				},
			},
			Status: corev1.NodeStatus{
				Conditions: []corev1.NodeCondition{
					{Type: corev1.NodeReady, Status: corev1.ConditionTrue},
				},
			},
		},
	}

	result := MapNodes(nodes, nil, nil)
	assert.Equal(t, 1, len(result))
	assert.Equal(t, "node-1", result[0].Name)
}

func TestMapServices(t *testing.T) {
	services := []corev1.Service{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "svc-1",
				Namespace: "default",
			},
			Spec: corev1.ServiceSpec{
				Type: corev1.ServiceTypeClusterIP,
				Ports: []corev1.ServicePort{
					{Port: 80, Protocol: "TCP"},
				},
			},
		},
	}

	result := MapServices(services)
	assert.Equal(t, 1, len(result))
	assert.Equal(t, "svc-1", result[0].Name)
	assert.Equal(t, "ClusterIP", result[0].Type)
}

func TestMapIngresses(t *testing.T) {
	ingresses := []networkingv1.Ingress{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "ing-1",
				Namespace: "default",
			},
			Spec: networkingv1.IngressSpec{
				Rules: []networkingv1.IngressRule{
					{
						Host: "example.com",
						IngressRuleValue: networkingv1.IngressRuleValue{
							HTTP: &networkingv1.HTTPIngressRuleValue{
								Paths: []networkingv1.HTTPIngressPath{
									{Path: "/", PathType: pathTypePtr(networkingv1.PathTypePrefix)},
								},
							},
						},
					},
				},
			},
		},
	}

	result := MapIngresses(ingresses)
	assert.Equal(t, 1, len(result))
	assert.Equal(t, "ing-1", result[0].Name)
}

func TestMapPVCs(t *testing.T) {
	pvcs := []corev1.PersistentVolumeClaim{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "pvc-1",
				Namespace: "default",
			},
			Spec: corev1.PersistentVolumeClaimSpec{
				AccessModes: []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
			},
			Status: corev1.PersistentVolumeClaimStatus{
				Phase: corev1.ClaimBound,
			},
		},
	}

	result := MapPVCs(pvcs)
	assert.Equal(t, 1, len(result))
	assert.Equal(t, "pvc-1", result[0].Name)
}

func TestMapPVs(t *testing.T) {
	pvs := []corev1.PersistentVolume{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "pv-1",
			},
			Spec: corev1.PersistentVolumeSpec{
				Capacity: corev1.ResourceList{
					"storage": resource.MustParse("10Gi"),
				},
				AccessModes: []corev1.PersistentVolumeAccessMode{corev1.ReadWriteOnce},
			},
			Status: corev1.PersistentVolumeStatus{
				Phase: corev1.VolumeBound,
			},
		},
	}

	result := MapPVs(pvs)
	assert.Equal(t, 1, len(result))
	assert.Equal(t, "pv-1", result[0].Name)
}

func TestMapStorageClasses(t *testing.T) {
	reclaimDelete := corev1.PersistentVolumeReclaimDelete
	scs := []storagev1.StorageClass{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name: "standard",
			},
			Provisioner: "kubernetes.io/aws-ebs",
			ReclaimPolicy: &reclaimDelete,
		},
	}

	result := MapStorageClasses(scs)
	assert.Equal(t, 1, len(result))
	assert.Equal(t, "standard", result[0].Name)
}

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
	}

	result := MapNamespaces(namespaces)
	assert.Equal(t, 1, len(result))
	assert.Equal(t, "default", result[0].Name)
	assert.Equal(t, "Active", result[0].Status)
}

func TestMapEvents(t *testing.T) {
	events := []corev1.Event{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "event-1",
				Namespace: "default",
			},
			InvolvedObject: corev1.ObjectReference{
				Kind: "Pod",
				Name: "test-pod",
			},
			Reason:        "Created",
			Message:       "Pod was created",
			LastTimestamp: metav1.NewTime(time.Now()),
			Type:          "Normal",
		},
	}

	result := MapEvents(events)
	assert.Equal(t, 1, len(result))
	assert.Equal(t, "event-1", result[0].Name)
	assert.Equal(t, "Created", result[0].Reason)
}

func TestMapEndpoints(t *testing.T) {
	endpoints := []corev1.Endpoints{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "ep-1",
				Namespace: "default",
			},
			Subsets: []corev1.EndpointSubset{
				{
					Ports: []corev1.EndpointPort{
						{Port: 80, Name: "http"},
					},
				},
			},
		},
	}

	result := MapEndpoints(endpoints)
	assert.Equal(t, 1, len(result))
	assert.Equal(t, "ep-1", result[0].Name)
}

func TestMapJobs(t *testing.T) {
	jobs := []batchv1.Job{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "job-1",
				Namespace: "default",
			},
			Status: batchv1.JobStatus{
				Active:    1,
				Succeeded: 2,
				Failed:    0,
			},
		},
	}

	result := MapJobs(jobs)
	assert.Equal(t, 1, len(result))
	assert.Equal(t, "job-1", result[0].Name)
}

func TestMapCronJobs(t *testing.T) {
	cronJobs := []batchv1.CronJob{
		{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "cronjob-1",
				Namespace: "default",
			},
			Spec: batchv1.CronJobSpec{
				Schedule: "*/5 * * * *",
			},
		},
	}

	result := MapCronJobs(cronJobs)
	assert.Equal(t, 1, len(result))
	assert.Equal(t, "cronjob-1", result[0].Name)
}

func TestCalculateCPUUsage(t *testing.T) {
	tests := []struct {
		name       string
		capacity    *resource.Quantity
		usageStr   string
		expected   float64
	}{
		{
			name:     "normal usage",
			capacity:  resource.NewMilliQuantity(4000, resource.DecimalSI), // 4 cores
			usageStr: "2000m",                                  // 2 cores
			expected:  50.0,                                      // 50%
		},
		{
			name:     "zero capacity",
			capacity:  resource.NewMilliQuantity(0, resource.DecimalSI),
			usageStr: "1000m",
			expected:  0.0,
		},
		{
			name:     "empty usage",
			capacity:  resource.NewMilliQuantity(4000, resource.DecimalSI),
			usageStr: "",
			expected:  0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calculateCPUUsage(tt.capacity, tt.usageStr)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCalculateMemoryUsage(t *testing.T) {
	tests := []struct {
		name       string
		capacity    *resource.Quantity
		usageStr   string
		expected   float64
	}{
		{
			name:     "normal usage",
			capacity:  resource.NewQuantity(8*1024*1024*1024, resource.BinarySI), // 8Gi
			usageStr: "4294967296",                                         // 4Gi
			expected:  50.0,
		},
		{
			name:     "zero capacity",
			capacity:  resource.NewQuantity(0, resource.BinarySI),
			usageStr: "1073741824",
			expected:  0.0,
		},
		{
			name:     "empty usage",
			capacity:  resource.NewQuantity(8*1024*1024*1024, resource.BinarySI),
			usageStr: "",
			expected:  0.0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calculateMemoryUsage(tt.capacity, tt.usageStr)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetJobStatus(t *testing.T) {
	startTime := metav1.NewTime(time.Now())
	completions := int32(1)
	tests := []struct {
		name     string
		job      batchv1.Job
		expected string
	}{
		{
			name: "active job",
			job: batchv1.Job{
				Status: batchv1.JobStatus{
					Active:    1,
					StartTime: &startTime,
				},
			},
			expected: "Running",
		},
		{
			name: "successful job",
			job: batchv1.Job{
				Spec: batchv1.JobSpec{
					Completions: &completions,
				},
				Status: batchv1.JobStatus{
					Succeeded: 1,
				},
			},
			expected: "Completed",
		},
		{
			name: "failed job",
			job: batchv1.Job{
				Status: batchv1.JobStatus{
					Failed: 1,
				},
			},
			expected: "Failed",
		},
		{
			name: "pending job",
			job: batchv1.Job{
				Status: batchv1.JobStatus{},
			},
			expected: "Pending",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getJobStatus(tt.job)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetCronJobStatus(t *testing.T) {
	tests := []struct {
		name     string
		cj       batchv1.CronJob
		expected string
	}{
		{
			name: "suspended cronjob",
			cj: batchv1.CronJob{
				Spec: batchv1.CronJobSpec{
					Suspend: &[]bool{true}[0],
				},
			},
			expected: "Suspended",
		},
		{
			name: "active cronjob",
			cj: batchv1.CronJob{
				Spec: batchv1.CronJobSpec{
					Suspend: &[]bool{false}[0],
				},
			},
			expected: "Active",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getCronJobStatus(tt.cj)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetIngressClass(t *testing.T) {
	tests := []struct {
		name     string
		ing      networkingv1.Ingress
		expected string
	}{
		{
			name: "with class name",
			ing: networkingv1.Ingress{
				Spec: networkingv1.IngressSpec{
					IngressClassName: &[]string{"nginx"}[0],
				},
			},
			expected: "nginx",
		},
		{
			name: "without class name",
			ing: networkingv1.Ingress{},
			expected: "-",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getIngressClass(tt.ing)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetStorageClassName(t *testing.T) {
	tests := []struct {
		name           string
		pvc            corev1.PersistentVolumeClaim
		expected       string
	}{
		{
			name: "with storage class",
			pvc: corev1.PersistentVolumeClaim{
				Spec: corev1.PersistentVolumeClaimSpec{
					StorageClassName: &[]string{"standard"}[0],
				},
			},
			expected: "standard",
		},
		{
			name: "without storage class",
			pvc: corev1.PersistentVolumeClaim{
				Spec: corev1.PersistentVolumeClaimSpec{},
			},
			expected: "-",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getStorageClassName(tt.pvc)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetInternalIP(t *testing.T) {
	tests := []struct {
		name     string
		node      corev1.Node
		expected string
	}{
		{
			name: "with internal IP",
			node: corev1.Node{
				Status: corev1.NodeStatus{
					Addresses: []corev1.NodeAddress{
						{Type: corev1.NodeInternalIP, Address: "10.0.0.1"},
					},
				},
			},
			expected: "10.0.0.1",
		},
		{
			name: "without internal IP",
			node: corev1.Node{
				Status: corev1.NodeStatus{
					Addresses: []corev1.NodeAddress{
						{Type: corev1.NodeExternalIP, Address: "1.2.3.4"},
					},
				},
			},
			expected: "-",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getInternalIP(tt.node)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFormatEventLastSeen(t *testing.T) {
	tests := []struct {
		name     string
		event    corev1.Event
	}{
		{
			name: "with last timestamp",
			event: corev1.Event{
				LastTimestamp: metav1.NewTime(time.Now().Add(-5 * time.Minute)),
			},
		},
		{
			name: "with event time",
			event: corev1.Event{
				EventTime: metav1.MicroTime{Time: time.Now().Add(-10 * time.Minute)},
			},
		},
		{
			name: "with creation timestamp",
			event: corev1.Event{
				ObjectMeta: metav1.ObjectMeta{
					CreationTimestamp: metav1.NewTime(time.Now().Add(-15 * time.Minute)),
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatEventLastSeen(tt.event)
			assert.NotEmpty(t, result)
		})
	}
}

func TestGetNodeRoles(t *testing.T) {
	tests := []struct {
		name     string
		node     corev1.Node
		expected []string
	}{
		{
			name: "worker node",
			node: corev1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"node-role.kubernetes.io/worker": ""},
				},
			},
			expected: []string{"worker"},
		},
		{
			name: "control-plane node",
			node: corev1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"node-role.kubernetes.io/control-plane": ""},
				},
			},
			expected: []string{"control-plane"},
		},
		{
			name: "master node (legacy)",
			node: corev1.Node{
				ObjectMeta: metav1.ObjectMeta{
					Labels: map[string]string{"node-role.kubernetes.io/master": ""},
				},
			},
			expected: []string{"control-plane"},
		},
		{
			name: "no roles",
			node: corev1.Node{},
			expected: []string{"worker"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getNodeRoles(tt.node)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCountEndpointAddresses(t *testing.T) {
	tests := []struct {
		name     string
		ep       corev1.Endpoints
		expected int
	}{
		{
			name: "with hostname",
			ep: corev1.Endpoints{
				Subsets: []corev1.EndpointSubset{
					{Addresses: []corev1.EndpointAddress{{Hostname: "pod-1"}}},
				},
			},
			expected: 1,
		},
		{
			name: "with IP",
			ep: corev1.Endpoints{
				Subsets: []corev1.EndpointSubset{
					{Addresses: []corev1.EndpointAddress{{IP: "10.0.0.1"}}},
				},
			},
			expected: 1,
		},
		{
			name: "multiple addresses",
			ep: corev1.Endpoints{
				Subsets: []corev1.EndpointSubset{
					{Addresses: []corev1.EndpointAddress{
						{IP: "10.0.0.1"},
						{IP: "10.0.0.2"},
					}},
				},
			},
			expected: 2,
		},
		{
			name: "empty addresses",
			ep: corev1.Endpoints{},
			expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := countEndpointAddresses(tt.ep)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func ptrString(s string) *string {
	return &s
}

func TestGetWorkloadStatus(t *testing.T) {
	tests := []struct {
		name     string
		ready    int32
		desired  int32
		expected string
	}{
		{
			name:     "all ready",
			ready:    3,
			desired:  3,
			expected: "Available",
		},
		{
			name:     "partial ready",
			ready:    1,
			desired:  3,
			expected: "Partial",
		},
		{
			name:     "none ready",
			ready:    0,
			desired:  3,
			expected: "Unavailable",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GetWorkloadStatus(tt.ready, tt.desired)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCalculatePodReady(t *testing.T) {
	tests := []struct {
		name     string
		pod      corev1.Pod
		expected string
	}{
		{
			name: "pod ready",
			pod: corev1.Pod{
				Status: corev1.PodStatus{
					Phase: corev1.PodRunning,
					Conditions: []corev1.PodCondition{
						{Type: corev1.PodReady, Status: corev1.ConditionTrue},
					},
					ContainerStatuses: []corev1.ContainerStatus{
						{Ready: true, Name: "nginx"},
					},
				},
			},
			expected: "1/1",
		},
		{
			name: "pod not ready",
			pod: corev1.Pod{
				Status: corev1.PodStatus{
					Phase: corev1.PodPending,
					ContainerStatuses: []corev1.ContainerStatus{
						{Ready: false, Name: "nginx"},
					},
				},
			},
			expected: "0/1",
		},
		{
			name: "multiple containers partial ready",
			pod: corev1.Pod{
				Status: corev1.PodStatus{
					ContainerStatuses: []corev1.ContainerStatus{
						{Ready: true, Name: "c1"},
						{Ready: false, Name: "c2"},
					},
				},
			},
			expected: "1/2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CalculatePodReady(tt.pod)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCalculatePodRestarts(t *testing.T) {
	pod := corev1.Pod{
		Status: corev1.PodStatus{
			ContainerStatuses: []corev1.ContainerStatus{
				{RestartCount: 3},
				{RestartCount: 2},
			},
		},
	}

	result := CalculatePodRestarts(pod)
	assert.Equal(t, int32(5), result)
}

func TestCalculateAge(t *testing.T) {
	tests := []struct {
		name     string
		creation metav1.Time
	}{
		{
			name:     "recent creation",
			creation: metav1.NewTime(time.Now().Add(-5 * time.Minute)),
		},
		{
			name:     "old creation",
			creation: metav1.NewTime(time.Now().Add(-24 * 365 * time.Hour)),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CalculateAge(tt.creation)
			assert.NotEmpty(t, result)
		})
	}
}

func TestGetPodPhaseDisplay(t *testing.T) {
	tests := []struct {
		name     string
		phase    corev1.PodPhase
		expected string
	}{
		{"Running", corev1.PodRunning, "Running"},
		{"Pending", corev1.PodPending, "Pending"},
		{"Succeeded", corev1.PodSucceeded, "Succeeded"},
		{"Failed", corev1.PodFailed, "Failed"},
		{"Unknown", corev1.PodUnknown, "Unknown"},
		{"Custom", corev1.PodPhase("CustomPhase"), "CustomPhase"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getPodPhaseDisplay(tt.phase)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetServicePorts(t *testing.T) {
	tests := []struct {
		name     string
		svc      corev1.Service
		expected []string
	}{
		{
			name:     "no ports",
			svc:      corev1.Service{},
			expected: []string{"-"},
		},
		{
			name: "single port",
			svc: corev1.Service{
				Spec: corev1.ServiceSpec{
					Ports: []corev1.ServicePort{{Port: 80}},
				},
			},
			expected: []string{"80"},
		},
		{
			name: "multiple ports",
			svc: corev1.Service{
				Spec: corev1.ServiceSpec{
					Ports: []corev1.ServicePort{{Port: 80}, {Port: 443}, {Port: 8080}},
				},
			},
			expected: []string{"80", "443", "8080"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getServicePorts(tt.svc)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetEndpointPorts(t *testing.T) {
	tests := []struct {
		name     string
		ep       corev1.Endpoints
		expected []string
	}{
		{
			name:     "no subsets",
			ep:       corev1.Endpoints{},
			expected: []string{"-"},
		},
		{
			name: "single port",
			ep: corev1.Endpoints{
				Subsets: []corev1.EndpointSubset{
					{Ports: []corev1.EndpointPort{{Port: 80}}},
				},
			},
			expected: []string{"80"},
		},
		{
			name: "multiple ports across subsets",
			ep: corev1.Endpoints{
				Subsets: []corev1.EndpointSubset{
					{Ports: []corev1.EndpointPort{{Port: 80}, {Port: 443}}},
					{Ports: []corev1.EndpointPort{{Port: 8080}}},
				},
			},
			expected: []string{"80", "443", "8080"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getEndpointPorts(tt.ep)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetIngressHosts(t *testing.T) {
	tests := []struct {
		name     string
		ing      networkingv1.Ingress
		expected []string
	}{
		{
			name:     "no rules",
			ing:      networkingv1.Ingress{},
			expected: []string{"*"},
		},
		{
			name: "single host",
			ing: networkingv1.Ingress{
				Spec: networkingv1.IngressSpec{
					Rules: []networkingv1.IngressRule{
						{Host: "example.com"},
					},
				},
			},
			expected: []string{"example.com"},
		},
		{
			name: "multiple hosts",
			ing: networkingv1.Ingress{
				Spec: networkingv1.IngressSpec{
					Rules: []networkingv1.IngressRule{
						{Host: "example.com"},
						{Host: "api.example.com"},
						{Host: "web.example.com"},
					},
				},
			},
			expected: []string{"example.com", "api.example.com", "web.example.com"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getIngressHosts(tt.ing)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCalculateJobDuration(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name     string
		job      batchv1.Job
		expected string
	}{
		{
			name: "no start time",
			job:  batchv1.Job{},
			expected: "-",
		},
		{
			name: "completed job",
			job: batchv1.Job{
				Status: batchv1.JobStatus{
					StartTime:      &metav1.Time{Time: now.Add(-1 * time.Hour)},
					CompletionTime: &metav1.Time{Time: now},
				},
			},
			expected: "1h0m0s",
		},
		{
			name: "running job",
			job: batchv1.Job{
				Status: batchv1.JobStatus{
					StartTime: &metav1.Time{Time: now.Add(-30 * time.Minute)},
				},
			},
			expected: "30m0s",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calculateJobDuration(tt.job)
			if tt.expected == "-" {
				assert.Equal(t, tt.expected, result)
			} else {
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestGetNodeStatus(t *testing.T) {
	tests := []struct {
		name     string
		node     corev1.Node
		expected string
	}{
		{
			name: "ready node",
			node: corev1.Node{
				Status: corev1.NodeStatus{
					Conditions: []corev1.NodeCondition{
						{Type: corev1.NodeReady, Status: corev1.ConditionTrue},
					},
				},
			},
			expected: "Ready",
		},
		{
			name: "not ready node",
			node: corev1.Node{
				Status: corev1.NodeStatus{
					Conditions: []corev1.NodeCondition{
						{Type: corev1.NodeReady, Status: corev1.ConditionFalse},
					},
				},
			},
			expected: "NotReady",
		},
		{
			name: "no ready condition",
			node: corev1.Node{
				Status: corev1.NodeStatus{
					Conditions: []corev1.NodeCondition{
						{Type: "MemoryPressure", Status: corev1.ConditionFalse},
					},
				},
			},
			expected: "Unknown",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getNodeStatus(tt.node)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestIsDefaultStorageClass(t *testing.T) {
	tests := []struct {
		name     string
		sc       storagev1.StorageClass
		expected bool
	}{
		{
			name: "default storage class",
			sc: storagev1.StorageClass{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"storageclass.kubernetes.io/is-default-class": "true",
					},
				},
			},
			expected: true,
		},
		{
			name: "not default",
			sc: storagev1.StorageClass{
				ObjectMeta: metav1.ObjectMeta{
					Annotations: map[string]string{
						"storageclass.kubernetes.io/is-default-class": "false",
					},
				},
			},
			expected: false,
		},
		{
			name:     "no annotation",
			sc:       storagev1.StorageClass{},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isDefaultStorageClass(tt.sc)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetIngressPaths(t *testing.T) {
	tests := []struct {
		name     string
		ing      networkingv1.Ingress
		expected []string
	}{
		{
			name:     "no rules",
			ing:      networkingv1.Ingress{},
			expected: []string{"/"},
		},
		{
			name: "single path",
			ing: networkingv1.Ingress{
				Spec: networkingv1.IngressSpec{
					Rules: []networkingv1.IngressRule{
						{
							Host: "example.com",
							IngressRuleValue: networkingv1.IngressRuleValue{
								HTTP: &networkingv1.HTTPIngressRuleValue{
									Paths: []networkingv1.HTTPIngressPath{
										{Path: "/api"},
									},
								},
							},
						},
					},
				},
			},
			expected: []string{"/api"},
		},
		{
			name: "multiple paths",
			ing: networkingv1.Ingress{
				Spec: networkingv1.IngressSpec{
					Rules: []networkingv1.IngressRule{
						{
							IngressRuleValue: networkingv1.IngressRuleValue{
								HTTP: &networkingv1.HTTPIngressRuleValue{
									Paths: []networkingv1.HTTPIngressPath{
										{Path: "/api"},
										{Path: "/web"},
										{Path: "/admin"},
									},
								},
							},
						},
					},
				},
			},
			expected: []string{"/api", "/web", "/admin"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getIngressPaths(tt.ing)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetIngressServices(t *testing.T) {
	tests := []struct {
		name     string
		ing      networkingv1.Ingress
		expected []string
	}{
		{
			name:     "no rules",
			ing:      networkingv1.Ingress{},
			expected: []string{"-"},
		},
		{
			name: "single service",
			ing: networkingv1.Ingress{
				Spec: networkingv1.IngressSpec{
					Rules: []networkingv1.IngressRule{
						{
							IngressRuleValue: networkingv1.IngressRuleValue{
								HTTP: &networkingv1.HTTPIngressRuleValue{
									Paths: []networkingv1.HTTPIngressPath{
										{
											Backend: networkingv1.IngressBackend{
												Service: &networkingv1.IngressServiceBackend{
													Name: "api-service",
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			expected: []string{"api-service"},
		},
		{
			name: "multiple services",
			ing: networkingv1.Ingress{
				Spec: networkingv1.IngressSpec{
					Rules: []networkingv1.IngressRule{
						{
							IngressRuleValue: networkingv1.IngressRuleValue{
								HTTP: &networkingv1.HTTPIngressRuleValue{
									Paths: []networkingv1.HTTPIngressPath{
										{
											Backend: networkingv1.IngressBackend{
												Service: &networkingv1.IngressServiceBackend{
													Name: "api-service",
												},
											},
										},
										{
											Backend: networkingv1.IngressBackend{
												Service: &networkingv1.IngressServiceBackend{
													Name: "web-service",
												},
											},
										},
									},
								},
							},
						},
					},
				},
			},
			expected: []string{"api-service", "web-service"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getIngressServices(tt.ing)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFormatEventObject(t *testing.T) {
	tests := []struct {
		name     string
		obj      corev1.ObjectReference
		expected string
	}{
		{
			name:     "empty",
			obj:      corev1.ObjectReference{},
			expected: "-",
		},
		{
			name: "kind only",
			obj: corev1.ObjectReference{
				Kind: "Pod",
			},
			expected: "Pod",
		},
		{
			name: "name only",
			obj: corev1.ObjectReference{
				Name: "my-pod",
			},
			expected: "my-pod",
		},
		{
			name: "kind and name",
			obj: corev1.ObjectReference{
				Kind: "Pod",
				Name: "my-pod",
			},
			expected: "Pod/my-pod",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatEventObject(tt.obj)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestBuildNodePodCountMap(t *testing.T) {
	tests := []struct {
		name     string
		pods     *corev1.PodList
		expected map[string]int
	}{
		{
			name:     "nil pod list",
			pods:     nil,
			expected: map[string]int{},
		},
		{
			name: "single pod",
			pods: &corev1.PodList{
				Items: []corev1.Pod{
					{Spec: corev1.PodSpec{NodeName: "node-1"}},
				},
			},
			expected: map[string]int{"node-1": 1},
		},
		{
			name: "multiple pods same node",
			pods: &corev1.PodList{
				Items: []corev1.Pod{
					{Spec: corev1.PodSpec{NodeName: "node-1"}},
					{Spec: corev1.PodSpec{NodeName: "node-1"}},
					{Spec: corev1.PodSpec{NodeName: "node-2"}},
				},
			},
			expected: map[string]int{"node-1": 2, "node-2": 1},
		},
		{
			name: "pod without node",
			pods: &corev1.PodList{
				Items: []corev1.Pod{
					{Spec: corev1.PodSpec{NodeName: ""}},
				},
			},
			expected: map[string]int{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := buildNodePodCountMap(tt.pods)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestCalculateEventDuration(t *testing.T) {
	now := time.Now()
	tests := []struct {
		name     string
		event    corev1.Event
		expected string
	}{
		{
			name: "no first timestamp",
			event: corev1.Event{
				FirstTimestamp: metav1.Time{},
			},
			expected: "-",
		},
		{
			name: "with first and last timestamp",
			event: corev1.Event{
				FirstTimestamp: metav1.Time{Time: now.Add(-1 * time.Hour)},
				LastTimestamp:  metav1.Time{Time: now},
			},
			expected: "1h0m0s",
		},
		{
			name: "with event time",
			event: corev1.Event{
				FirstTimestamp: metav1.Time{Time: now.Add(-30 * time.Minute)},
				EventTime:      metav1.MicroTime{Time: now},
			},
			expected: "30m0s",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := calculateEventDuration(tt.event)
			if tt.expected == "-" {
				assert.Equal(t, tt.expected, result)
			} else {
				assert.Equal(t, tt.expected, result)
			}
		})
	}
}

func TestGetAccessMode(t *testing.T) {
	tests := []struct {
		name     string
		pvc      corev1.PersistentVolumeClaim
		expected string
	}{
		{
			name:     "no access modes",
			pvc:      corev1.PersistentVolumeClaim{},
			expected: "-",
		},
		{
			name: "single access mode",
			pvc: corev1.PersistentVolumeClaim{
				Spec: corev1.PersistentVolumeClaimSpec{
					AccessModes: []corev1.PersistentVolumeAccessMode{"ReadWriteOnce"},
				},
			},
			expected: "ReadWriteOnce",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getAccessMode(tt.pvc)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetClaimRef(t *testing.T) {
	tests := []struct {
		name     string
		pv       corev1.PersistentVolume
		expected string
	}{
		{
			name:     "no claim ref",
			pv:       corev1.PersistentVolume{},
			expected: "-",
		},
		{
			name: "with claim ref",
			pv: corev1.PersistentVolume{
				Spec: corev1.PersistentVolumeSpec{
					ClaimRef: &corev1.ObjectReference{
						Name:      "my-pvc",
						Namespace: "default",
					},
				},
			},
			expected: "default/my-pvc",
		},
		{
			name: "with claim ref no namespace",
			pv: corev1.PersistentVolume{
				Spec: corev1.PersistentVolumeSpec{
					ClaimRef: &corev1.ObjectReference{
						Name: "my-pvc",
					},
				},
			},
			expected: "default/my-pvc",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getClaimRef(tt.pv)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetPVAccessMode(t *testing.T) {
	tests := []struct {
		name     string
		pv       corev1.PersistentVolume
		expected string
	}{
		{
			name:     "no access modes",
			pv:       corev1.PersistentVolume{},
			expected: "-",
		},
		{
			name: "single access mode",
			pv: corev1.PersistentVolume{
				Spec: corev1.PersistentVolumeSpec{
					AccessModes: []corev1.PersistentVolumeAccessMode{"ReadWriteOnce"},
				},
			},
			expected: "ReadWriteOnce",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getPVAccessMode(tt.pv)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetReclaimPolicy(t *testing.T) {
	tests := []struct {
		name     string
		sc       storagev1.StorageClass
		expected string
	}{
		{
			name: "with reclaim policy",
			sc: storagev1.StorageClass{
				ReclaimPolicy: func() *corev1.PersistentVolumeReclaimPolicy {
					p := corev1.PersistentVolumeReclaimPolicy("Retain")
					return &p
				}(),
			},
			expected: "Retain",
		},
		{
			name:     "no reclaim policy",
			sc:       storagev1.StorageClass{},
			expected: "Delete",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getReclaimPolicy(tt.sc)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetReclaimPolicyStr(t *testing.T) {
	tests := []struct {
		name     string
		pv       corev1.PersistentVolume
		expected string
	}{
		{
			name: "with reclaim policy",
			pv: corev1.PersistentVolume{
				Spec: corev1.PersistentVolumeSpec{
					PersistentVolumeReclaimPolicy: "Retain",
				},
			},
			expected: "Retain",
		},
		{
			name:     "no reclaim policy",
			pv:       corev1.PersistentVolume{},
			expected: "Delete",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getReclaimPolicyStr(tt.pv)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestGetVolumeBindingMode(t *testing.T) {
	tests := []struct {
		name     string
		sc       storagev1.StorageClass
		expected string
	}{
		{
			name: "with volume binding mode",
			sc: storagev1.StorageClass{
				VolumeBindingMode: func() *storagev1.VolumeBindingMode {
					m := storagev1.VolumeBindingWaitForFirstConsumer
					return &m
				}(),
			},
			expected: "WaitForFirstConsumer",
		},
		{
			name:     "no volume binding mode",
			sc:       storagev1.StorageClass{},
			expected: "Immediate",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getVolumeBindingMode(tt.sc)
			assert.Equal(t, tt.expected, result)
		})
	}
}

// Helper functions
func pathTypePtr(pt networkingv1.PathType) *networkingv1.PathType {
	return &pt
}
