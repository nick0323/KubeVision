package service

import (
	"context"
	"fmt"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// Describe 超时配置常量
const (
	DescribeTimeout = 10 * time.Second
	EventsTimeout   = 5 * time.Second
)

// DescribeResult 描述结果结构体（JSON 格式）
type DescribeResult struct {
	APIVersion string                 `json:"apiVersion"`
	Kind       string                 `json:"kind"`
	Metadata   metav1.ObjectMeta      `json:"metadata"`
	Spec       map[string]interface{} `json:"spec"`
	Status     map[string]interface{} `json:"status"`
	Events     []EventInfo            `json:"events,omitempty"`
}

// EventInfo 事件信息
type EventInfo struct {
	Type      string `json:"type"`
	Reason    string `json:"reason"`
	Message   string `json:"message"`
	Count     int32  `json:"count"`
	FirstTime string `json:"firstTime"`
	LastTime  string `json:"lastTime"`
	Source    string `json:"source"`
}

// DescribePod 获取 Pod 的详细信息（JSON 格式）
func DescribePod(ctx context.Context, clientset *kubernetes.Clientset, namespace, name string) (*DescribeResult, error) {
	// 添加超时控制
	ctx, cancel := context.WithTimeout(ctx, DescribeTimeout)
	defer cancel()

	pod, err := clientset.CoreV1().Pods(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get pod: %w", err)
	}

	// 获取相关事件（单独超时控制）
	eventsCtx, eventsCancel := context.WithTimeout(ctx, EventsTimeout)
	defer eventsCancel()
	events := getEventsForResource(eventsCtx, clientset, namespace, "Pod", name)

	// 构建结构化的 Pod 信息
	spec := map[string]interface{}{
		"containers":     pod.Spec.Containers,
		"initContainers": pod.Spec.InitContainers,
		"volumes":        pod.Spec.Volumes,
		"nodeName":       pod.Spec.NodeName,
		"serviceAccount": pod.Spec.ServiceAccountName,
		"restartPolicy":  pod.Spec.RestartPolicy,
		"dnsPolicy":      pod.Spec.DNSPolicy,
		"priority":       pod.Spec.Priority,
		"tolerations":    pod.Spec.Tolerations,
		"affinity":       pod.Spec.Affinity,
		"schedulerName":  pod.Spec.SchedulerName,
	}

	status := map[string]interface{}{
		"phase":                 pod.Status.Phase,
		"hostIP":                pod.Status.HostIP,
		"podIP":                 pod.Status.PodIP,
		"podIPs":                pod.Status.PodIPs,
		"conditions":            pod.Status.Conditions,
		"containerStatuses":     pod.Status.ContainerStatuses,
		"initContainerStatuses": pod.Status.InitContainerStatuses,
		"qosClass":              pod.Status.QOSClass,
		"startTime":             pod.Status.StartTime,
		"reason":                pod.Status.Reason,
		"message":               pod.Status.Message,
	}

	return &DescribeResult{
		APIVersion: "v1",
		Kind:       "Pod",
		Metadata:   pod.ObjectMeta,
		Spec:       spec,
		Status:     status,
		Events:     events,
	}, nil
}

// DescribeDeployment 获取 Deployment 的详细信息（JSON 格式）
func DescribeDeployment(ctx context.Context, clientset *kubernetes.Clientset, namespace, name string) (*DescribeResult, error) {
	// 添加超时控制
	ctx, cancel := context.WithTimeout(ctx, DescribeTimeout)
	defer cancel()

	deploy, err := clientset.AppsV1().Deployments(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get deployment: %w", err)
	}

	// 获取相关事件
	eventsCtx, eventsCancel := context.WithTimeout(ctx, EventsTimeout)
	defer eventsCancel()
	events := getEventsForResource(eventsCtx, clientset, namespace, "Deployment", name)

	spec := map[string]interface{}{
		"replicas":         deploy.Spec.Replicas,
		"selector":         deploy.Spec.Selector,
		"strategy":         deploy.Spec.Strategy,
		"template":         deploy.Spec.Template,
		"minReadySeconds":  deploy.Spec.MinReadySeconds,
		"revisionHistory":  deploy.Spec.RevisionHistoryLimit,
		"paused":           deploy.Spec.Paused,
		"progressDeadline": deploy.Spec.ProgressDeadlineSeconds,
	}

	status := map[string]interface{}{
		"replicas":            deploy.Status.Replicas,
		"readyReplicas":       deploy.Status.ReadyReplicas,
		"updatedReplicas":     deploy.Status.UpdatedReplicas,
		"availableReplicas":   deploy.Status.AvailableReplicas,
		"unavailableReplicas": deploy.Status.UnavailableReplicas,
		"conditions":          deploy.Status.Conditions,
		"observedGeneration":  deploy.Status.ObservedGeneration,
	}

	return &DescribeResult{
		APIVersion: "apps/v1",
		Kind:       "Deployment",
		Metadata:   deploy.ObjectMeta,
		Spec:       spec,
		Status:     status,
		Events:     events,
	}, nil
}

// DescribeService 获取 Service 的详细信息（JSON 格式）
func DescribeService(ctx context.Context, clientset *kubernetes.Clientset, namespace, name string) (*DescribeResult, error) {
	// 添加超时控制
	ctx, cancel := context.WithTimeout(ctx, DescribeTimeout)
	defer cancel()

	svc, err := clientset.CoreV1().Services(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get service: %w", err)
	}

	// 获取相关事件
	eventsCtx, eventsCancel := context.WithTimeout(ctx, EventsTimeout)
	defer eventsCancel()
	events := getEventsForResource(eventsCtx, clientset, namespace, "Service", name)

	spec := map[string]interface{}{
		"type":                  svc.Spec.Type,
		"clusterIP":             svc.Spec.ClusterIP,
		"clusterIPs":            svc.Spec.ClusterIPs,
		"externalIPs":           svc.Spec.ExternalIPs,
		"ports":                 svc.Spec.Ports,
		"selector":              svc.Spec.Selector,
		"sessionAffinity":       svc.Spec.SessionAffinity,
		"loadBalancerIP":        svc.Spec.LoadBalancerIP,
		"externalTrafficPolicy": svc.Spec.ExternalTrafficPolicy,
		"healthCheckNodePort":   svc.Spec.HealthCheckNodePort,
	}

	status := map[string]interface{}{
		"loadBalancer": svc.Status.LoadBalancer,
	}

	return &DescribeResult{
		APIVersion: "v1",
		Kind:       "Service",
		Metadata:   svc.ObjectMeta,
		Spec:       spec,
		Status:     status,
		Events:     events,
	}, nil
}

// DescribeNode 获取 Node 的详细信息
func DescribeNode(ctx context.Context, clientset *kubernetes.Clientset, name string) (*DescribeResult, error) {
	// 添加超时控制
	ctx, cancel := context.WithTimeout(ctx, 10*time.Second)
	defer cancel()

	node, err := clientset.CoreV1().Nodes().Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get node: %w", err)
	}

	// 获取相关事件
	eventsCtx, eventsCancel := context.WithTimeout(ctx, 5*time.Second)
	defer eventsCancel()
	events := getEventsForResource(eventsCtx, clientset, "", "Node", name)

	spec := map[string]interface{}{
		"taints":        node.Spec.Taints,
		"unschedulable": node.Spec.Unschedulable,
		"podCIDR":       node.Spec.PodCIDR,
		"podCIDRs":      node.Spec.PodCIDRs,
		"providerID":    node.Spec.ProviderID,
	}

	status := map[string]interface{}{
		"capacity":        node.Status.Capacity,
		"allocatable":     node.Status.Allocatable,
		"conditions":      node.Status.Conditions,
		"addresses":       node.Status.Addresses,
		"nodeInfo":        node.Status.NodeInfo,
		"images":          node.Status.Images,
		"daemonEndpoints": node.Status.DaemonEndpoints,
	}

	return &DescribeResult{
		APIVersion: "v1",
		Kind:       "Node",
		Metadata:   node.ObjectMeta,
		Spec:       spec,
		Status:     status,
		Events:     events,
	}, nil
}

// DescribeConfigMap 获取 ConfigMap 的详细信息（JSON 格式）
func DescribeConfigMap(ctx context.Context, clientset *kubernetes.Clientset, namespace, name string) (*DescribeResult, error) {
	cm, err := clientset.CoreV1().ConfigMaps(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get configmap: %w", err)
	}

	events := getEventsForResource(ctx, clientset, namespace, "ConfigMap", name)

	spec := map[string]interface{}{
		"data":       cm.Data,
		"binaryData": cm.BinaryData,
	}

	return &DescribeResult{
		APIVersion: "v1",
		Kind:       "ConfigMap",
		Metadata:   cm.ObjectMeta,
		Spec:       spec,
		Status:     nil,
		Events:     events,
	}, nil
}

// DescribeSecret 获取 Secret 的详细信息（JSON 格式）
func DescribeSecret(ctx context.Context, clientset *kubernetes.Clientset, namespace, name string) (*DescribeResult, error) {
	secret, err := clientset.CoreV1().Secrets(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get secret: %w", err)
	}

	events := getEventsForResource(ctx, clientset, namespace, "Secret", name)

	// 不返回实际的 secret 数据，只返回 key 列表
	dataKeys := make([]string, 0, len(secret.Data))
	for k := range secret.Data {
		dataKeys = append(dataKeys, k)
	}

	spec := map[string]interface{}{
		"type":       secret.Type,
		"dataKeys":   dataKeys, // 只返回 key 列表，不返回实际数据
		"stringData": secret.StringData,
	}

	return &DescribeResult{
		APIVersion: "v1",
		Kind:       "Secret",
		Metadata:   secret.ObjectMeta,
		Spec:       spec,
		Status:     nil,
		Events:     events,
	}, nil
}

// DescribePVC 获取 PersistentVolumeClaim 的详细信息（JSON 格式）
func DescribePVC(ctx context.Context, clientset *kubernetes.Clientset, namespace, name string) (*DescribeResult, error) {
	pvc, err := clientset.CoreV1().PersistentVolumeClaims(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get pvc: %w", err)
	}

	events := getEventsForResource(ctx, clientset, namespace, "PersistentVolumeClaim", name)

	spec := map[string]interface{}{
		"accessModes":      pvc.Spec.AccessModes,
		"resources":        pvc.Spec.Resources,
		"volumeName":       pvc.Spec.VolumeName,
		"storageClassName": pvc.Spec.StorageClassName,
		"volumeMode":       pvc.Spec.VolumeMode,
		"dataSource":       pvc.Spec.DataSource,
	}

	status := map[string]interface{}{
		"phase":       pvc.Status.Phase,
		"accessModes": pvc.Status.AccessModes,
		"capacity":    pvc.Status.Capacity,
		"conditions":  pvc.Status.Conditions,
	}

	return &DescribeResult{
		APIVersion: "v1",
		Kind:       "PersistentVolumeClaim",
		Metadata:   pvc.ObjectMeta,
		Spec:       spec,
		Status:     status,
		Events:     events,
	}, nil
}

// DescribePV 获取 PersistentVolume 的详细信息（JSON 格式）
func DescribePV(ctx context.Context, clientset *kubernetes.Clientset, name string) (*DescribeResult, error) {
	pv, err := clientset.CoreV1().PersistentVolumes().Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get pv: %w", err)
	}

	events := getEventsForResource(ctx, clientset, "", "PersistentVolume", name)

	spec := map[string]interface{}{
		"capacity":                      pv.Spec.Capacity,
		"accessModes":                   pv.Spec.AccessModes,
		"claimRef":                      pv.Spec.ClaimRef,
		"persistentVolumeReclaimPolicy": pv.Spec.PersistentVolumeReclaimPolicy,
		"storageClassName":              pv.Spec.StorageClassName,
		"volumeMode":                    pv.Spec.VolumeMode,
		"mountOptions":                  pv.Spec.MountOptions,
		"local":                         pv.Spec.Local,
		"hostPath":                      pv.Spec.HostPath,
		"nfs":                           pv.Spec.NFS,
		"iscsi":                         pv.Spec.ISCSI,
		"cephfs":                        pv.Spec.CephFS,
		"flexVolume":                    pv.Spec.FlexVolume,
	}

	status := map[string]interface{}{
		"phase": pv.Status.Phase,
	}

	return &DescribeResult{
		APIVersion: "v1",
		Kind:       "PersistentVolume",
		Metadata:   pv.ObjectMeta,
		Spec:       spec,
		Status:     status,
		Events:     events,
	}, nil
}

// DescribeStatefulSet 获取 StatefulSet 的详细信息（JSON 格式）
func DescribeStatefulSet(ctx context.Context, clientset *kubernetes.Clientset, namespace, name string) (*DescribeResult, error) {
	sts, err := clientset.AppsV1().StatefulSets(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get statefulset: %w", err)
	}

	events := getEventsForResource(ctx, clientset, namespace, "StatefulSet", name)

	spec := map[string]interface{}{
		"replicas":             sts.Spec.Replicas,
		"selector":             sts.Spec.Selector,
		"serviceName":          sts.Spec.ServiceName,
		"template":             sts.Spec.Template,
		"volumeClaimTemplates": sts.Spec.VolumeClaimTemplates,
		"podManagementPolicy":  sts.Spec.PodManagementPolicy,
		"updateStrategy":       sts.Spec.UpdateStrategy,
		"revisionHistoryLimit": sts.Spec.RevisionHistoryLimit,
	}

	status := map[string]interface{}{
		"replicas":          sts.Status.Replicas,
		"readyReplicas":     sts.Status.ReadyReplicas,
		"currentReplicas":   sts.Status.CurrentReplicas,
		"updatedReplicas":   sts.Status.UpdatedReplicas,
		"availableReplicas": sts.Status.AvailableReplicas,
		"conditions":        sts.Status.Conditions,
		"collisionCount":    sts.Status.CollisionCount,
	}

	return &DescribeResult{
		APIVersion: "apps/v1",
		Kind:       "StatefulSet",
		Metadata:   sts.ObjectMeta,
		Spec:       spec,
		Status:     status,
		Events:     events,
	}, nil
}

// DescribeDaemonSet 获取 DaemonSet 的详细信息（JSON 格式）
func DescribeDaemonSet(ctx context.Context, clientset *kubernetes.Clientset, namespace, name string) (*DescribeResult, error) {
	ds, err := clientset.AppsV1().DaemonSets(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get daemonset: %w", err)
	}

	events := getEventsForResource(ctx, clientset, namespace, "DaemonSet", name)

	spec := map[string]interface{}{
		"selector":             ds.Spec.Selector,
		"template":             ds.Spec.Template,
		"updateStrategy":       ds.Spec.UpdateStrategy,
		"minReadySeconds":      ds.Spec.MinReadySeconds,
		"revisionHistoryLimit": ds.Spec.RevisionHistoryLimit,
	}

	status := map[string]interface{}{
		"currentNumberScheduled": ds.Status.CurrentNumberScheduled,
		"numberMisscheduled":     ds.Status.NumberMisscheduled,
		"desiredNumberScheduled": ds.Status.DesiredNumberScheduled,
		"numberReady":            ds.Status.NumberReady,
		"numberAvailable":        ds.Status.NumberAvailable,
		"numberUnavailable":      ds.Status.NumberUnavailable,
		"collisionCount":         ds.Status.CollisionCount,
		"conditions":             ds.Status.Conditions,
	}

	return &DescribeResult{
		APIVersion: "apps/v1",
		Kind:       "DaemonSet",
		Metadata:   ds.ObjectMeta,
		Spec:       spec,
		Status:     status,
		Events:     events,
	}, nil
}

// DescribeJob 获取 Job 的详细信息（JSON 格式）
func DescribeJob(ctx context.Context, clientset *kubernetes.Clientset, namespace, name string) (*DescribeResult, error) {
	job, err := clientset.BatchV1().Jobs(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get job: %w", err)
	}

	events := getEventsForResource(ctx, clientset, namespace, "Job", name)

	spec := map[string]interface{}{
		"completions":             job.Spec.Completions,
		"parallelism":             job.Spec.Parallelism,
		"template":                job.Spec.Template,
		"backoffLimit":            job.Spec.BackoffLimit,
		"activeDeadlineSeconds":   job.Spec.ActiveDeadlineSeconds,
		"ttlSecondsAfterFinished": job.Spec.TTLSecondsAfterFinished,
		"completionMode":          job.Spec.CompletionMode,
		"suspend":                 job.Spec.Suspend,
	}

	status := map[string]interface{}{
		"startTime":      job.Status.StartTime,
		"completionTime": job.Status.CompletionTime,
		"active":         job.Status.Active,
		"succeeded":      job.Status.Succeeded,
		"failed":         job.Status.Failed,
		"conditions":     job.Status.Conditions,
	}

	return &DescribeResult{
		APIVersion: "batch/v1",
		Kind:       "Job",
		Metadata:   job.ObjectMeta,
		Spec:       spec,
		Status:     status,
		Events:     events,
	}, nil
}

// DescribeCronJob 获取 CronJob 的详细信息（JSON 格式）
func DescribeCronJob(ctx context.Context, clientset *kubernetes.Clientset, namespace, name string) (*DescribeResult, error) {
	cj, err := clientset.BatchV1().CronJobs(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get cronjob: %w", err)
	}

	events := getEventsForResource(ctx, clientset, namespace, "CronJob", name)

	spec := map[string]interface{}{
		"schedule":                   cj.Spec.Schedule,
		"timezone":                   cj.Spec.TimeZone,
		"concurrencyPolicy":          cj.Spec.ConcurrencyPolicy,
		"suspend":                    cj.Spec.Suspend,
		"jobTemplate":                cj.Spec.JobTemplate,
		"successfulJobsHistoryLimit": cj.Spec.SuccessfulJobsHistoryLimit,
		"failedJobsHistoryLimit":     cj.Spec.FailedJobsHistoryLimit,
	}

	status := map[string]interface{}{
		"lastScheduleTime":   cj.Status.LastScheduleTime,
		"lastSuccessfulTime": cj.Status.LastSuccessfulTime,
		"active":             cj.Status.Active,
	}

	return &DescribeResult{
		APIVersion: "batch/v1",
		Kind:       "CronJob",
		Metadata:   cj.ObjectMeta,
		Spec:       spec,
		Status:     status,
		Events:     events,
	}, nil
}

// DescribeIngress 获取 Ingress 的详细信息（JSON 格式）
func DescribeIngress(ctx context.Context, clientset *kubernetes.Clientset, namespace, name string) (*DescribeResult, error) {
	ing, err := clientset.NetworkingV1().Ingresses(namespace).Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get ingress: %w", err)
	}

	events := getEventsForResource(ctx, clientset, namespace, "Ingress", name)

	spec := map[string]interface{}{
		"ingressClassName": ing.Spec.IngressClassName,
		"defaultBackend":   ing.Spec.DefaultBackend,
		"tls":              ing.Spec.TLS,
		"rules":            ing.Spec.Rules,
	}

	status := map[string]interface{}{
		"loadBalancer": ing.Status.LoadBalancer,
	}

	return &DescribeResult{
		APIVersion: "networking.k8s.io/v1",
		Kind:       "Ingress",
		Metadata:   ing.ObjectMeta,
		Spec:       spec,
		Status:     status,
		Events:     events,
	}, nil
}

// DescribeStorageClass 获取 StorageClass 的详细信息（JSON 格式）
func DescribeStorageClass(ctx context.Context, clientset *kubernetes.Clientset, name string) (*DescribeResult, error) {
	sc, err := clientset.StorageV1().StorageClasses().Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get storageclass: %w", err)
	}

	events := getEventsForResource(ctx, clientset, "", "StorageClass", name)

	spec := map[string]interface{}{
		"provisioner":          sc.Provisioner,
		"parameters":           sc.Parameters,
		"reclaimPolicy":        sc.ReclaimPolicy,
		"mountOptions":         sc.MountOptions,
		"allowVolumeExpansion": sc.AllowVolumeExpansion,
		"volumeBindingMode":    sc.VolumeBindingMode,
	}

	return &DescribeResult{
		APIVersion: "storage.k8s.io/v1",
		Kind:       "StorageClass",
		Metadata:   sc.ObjectMeta,
		Spec:       spec,
		Status:     nil,
		Events:     events,
	}, nil
}

// DescribeNamespace 获取 Namespace 的详细信息（JSON 格式）
func DescribeNamespace(ctx context.Context, clientset *kubernetes.Clientset, name string) (*DescribeResult, error) {
	ns, err := clientset.CoreV1().Namespaces().Get(ctx, name, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get namespace: %w", err)
	}

	events := getEventsForResource(ctx, clientset, name, "Namespace", name)

	spec := map[string]interface{}{
		"finalizers": ns.Spec.Finalizers,
	}

	status := map[string]interface{}{
		"phase":      ns.Status.Phase,
		"conditions": ns.Status.Conditions,
	}

	return &DescribeResult{
		APIVersion: "v1",
		Kind:       "Namespace",
		Metadata:   ns.ObjectMeta,
		Spec:       spec,
		Status:     status,
		Events:     events,
	}, nil
}

// getEventsForResource 获取资源相关的事件
func getEventsForResource(ctx context.Context, clientset *kubernetes.Clientset, namespace, kind, name string) []EventInfo {
	eventList, err := clientset.CoreV1().Events(namespace).List(ctx, metav1.ListOptions{
		FieldSelector: fmt.Sprintf("involvedObject.name=%s,involvedObject.kind=%s", name, kind),
	})
	if err != nil || len(eventList.Items) == 0 {
		return nil
	}

	events := make([]EventInfo, 0, len(eventList.Items))
	for _, event := range eventList.Items {
		source := event.Source.Component
		if source == "" {
			source = event.ReportingController
		}

		events = append(events, EventInfo{
			Type:      event.Type,
			Reason:    event.Reason,
			Message:   event.Message,
			Count:     event.Count,
			FirstTime: event.FirstTimestamp.Format(time.RFC3339),
			LastTime:  event.LastTimestamp.Format(time.RFC3339),
			Source:    source,
		})
	}

	return events
}
