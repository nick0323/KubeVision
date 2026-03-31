package service

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"github.com/nick0323/K8sVision/model"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	storagev1 "k8s.io/api/storage/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// podInformerInstance 全局 PodInformer 实例
var podInformerInstance *PodInformer

// SetPodInformer 设置 PodInformer 实例
func SetPodInformer(pi *PodInformer) {
	podInformerInstance = pi
}

// GetPodInformer 获取 PodInformer 实例
func GetPodInformer() *PodInformer {
	return podInformerInstance
}

// 时间单位常量（用于 CalculateAge）
const (
	SecondsPerMinute = 60
	SecondsPerHour   = 3600
	SecondsPerDay    = 86400
)

// CalculateAge 计算资源运行时长
func CalculateAge(creationTime metav1.Time) string {
	// 如果是零值时间，返回空字符串
	if creationTime.IsZero() || creationTime.Time.IsZero() {
		return ""
	}

	duration := time.Since(creationTime.Time)
	seconds := int(duration.Seconds())

	if seconds < SecondsPerMinute {
		return fmt.Sprintf("%ds", seconds)
	} else if seconds < SecondsPerHour {
		return fmt.Sprintf("%dm", seconds/SecondsPerMinute)
	} else if seconds < SecondsPerDay {
		return fmt.Sprintf("%dh", seconds/SecondsPerHour)
	} else {
		return fmt.Sprintf("%dd", seconds/SecondsPerDay)
	}
}

// CalculatePodRestarts 计算 Pod 重启次数
func CalculatePodRestarts(pod v1.Pod) int32 {
	var totalRestarts int32 = 0
	for _, cs := range pod.Status.ContainerStatuses {
		totalRestarts += cs.RestartCount
	}
	return totalRestarts
}

// CalculatePodReady 计算 Pod 就绪状态
func CalculatePodReady(pod v1.Pod) string {
	var readyContainers, totalContainers int32 = 0, 0
	for _, cs := range pod.Status.ContainerStatuses {
		totalContainers++
		if cs.Ready {
			readyContainers++
		}
	}
	return fmt.Sprintf("%d/%d", readyContainers, totalContainers)
}

// CalculateDeploymentRestarts 计算 Deployment 总重启次数
func CalculateDeploymentRestarts(d appsv1.Deployment) int32 {
	if podInformerInstance == nil {
		return 0
	}
	if len(d.OwnerReferences) == 0 {
		return 0
	}
	return podInformerInstance.GetRestartsByOwnerReference(d.OwnerReferences[0])
}

// CalculateStatefulSetRestarts 计算 StatefulSet 总重启次数
func CalculateStatefulSetRestarts(s appsv1.StatefulSet) int32 {
	if podInformerInstance == nil {
		return 0
	}
	if len(s.OwnerReferences) == 0 {
		return 0
	}
	return podInformerInstance.GetRestartsByOwnerReference(s.OwnerReferences[0])
}

// CalculateDaemonSetRestarts 计算 DaemonSet 总重启次数
func CalculateDaemonSetRestarts(d appsv1.DaemonSet) int32 {
	if podInformerInstance == nil {
		return 0
	}
	if len(d.OwnerReferences) == 0 {
		return 0
	}
	return podInformerInstance.GetRestartsByOwnerReference(d.OwnerReferences[0])
}

// CalculateJobRestarts 计算 Job 总重启次数
func CalculateJobRestarts(j batchv1.Job) int32 {
	if podInformerInstance == nil {
		return 0
	}
	if len(j.OwnerReferences) == 0 {
		return 0
	}
	return podInformerInstance.GetRestartsByOwnerReference(j.OwnerReferences[0])
}

// MapDeployments 专门用于映射 Deployment 的函数
func MapDeployments(deployments []appsv1.Deployment) []model.DeploymentStatus {
	result := make([]model.DeploymentStatus, len(deployments))
	for i, d := range deployments {
		status := GetWorkloadStatus(d.Status.ReadyReplicas, d.Status.Replicas)
		restarts := CalculateDeploymentRestarts(d)
		result[i] = model.DeploymentStatus{
			Namespace:       d.Namespace,
			Name:            d.Name,
			ReadyReplicas:   d.Status.ReadyReplicas,
			UpdatedReplicas: d.Status.UpdatedReplicas,
			Available:       d.Status.AvailableReplicas,
			Desired:         d.Status.Replicas,
			Restarts:        restarts,
			Status:          status,
			Age:             CalculateAge(d.CreationTimestamp),
		}
	}
	return result
}

// MapPods 专门用于映射 Pod 的函数
func MapPods(pods []v1.Pod, podMetricsMap model.PodMetricsMap) []model.PodStatus {
	result := make([]model.PodStatus, len(pods))
	for i, pod := range pods {
		cpuVal, memVal := FormatPodResourceUsage(podMetricsMap, pod.Namespace, pod.Name)
		result[i] = model.PodStatus{
			Namespace:   pod.Namespace,
			Name:        pod.Name,
			Status:      string(pod.Status.Phase),
			Ready:       CalculatePodReady(pod),
			Restarts:    CalculatePodRestarts(pod),
			Age:         CalculateAge(pod.CreationTimestamp),
			CPUUsage:    cpuVal,
			MemoryUsage: memVal,
			PodIP:       pod.Status.PodIP,
			NodeName:    pod.Spec.NodeName,
		}
	}
	return result
}

// MapServices 专门用于映射 Service 的函数
func MapServices(services []v1.Service) []model.ServiceStatus {
	result := make([]model.ServiceStatus, len(services))
	for i, svc := range services {
		ports := make([]string, len(svc.Spec.Ports))
		for j, port := range svc.Spec.Ports {
			ports[j] = fmt.Sprintf("%d/%s", port.Port, port.Protocol)
		}

		// 获取 External IP
		externalIP := ""
		if len(svc.Status.LoadBalancer.Ingress) > 0 {
			if svc.Status.LoadBalancer.Ingress[0].IP != "" {
				externalIP = svc.Status.LoadBalancer.Ingress[0].IP
			} else if svc.Status.LoadBalancer.Ingress[0].Hostname != "" {
				externalIP = svc.Status.LoadBalancer.Ingress[0].Hostname
			}
		}

		result[i] = model.ServiceStatus{
			Namespace:  svc.Namespace,
			Name:       svc.Name,
			Type:       string(svc.Spec.Type),
			ClusterIP:  svc.Spec.ClusterIP,
			ExternalIP: externalIP,
			Ports:      ports,
			Age:        CalculateAge(svc.CreationTimestamp),
		}
	}
	return result
}

// MapStatefulSets 专门用于映射 StatefulSet 的函数
func MapStatefulSets(statefulSets []appsv1.StatefulSet) []model.StatefulSetStatus {
	result := make([]model.StatefulSetStatus, len(statefulSets))
	for i, s := range statefulSets {
		status := GetWorkloadStatus(s.Status.ReadyReplicas, s.Status.Replicas)
		restarts := CalculateStatefulSetRestarts(s)
		result[i] = model.StatefulSetStatus{
			Namespace:       s.Namespace,
			Name:            s.Name,
			ReadyReplicas:   s.Status.ReadyReplicas,
			UpdatedReplicas: s.Status.UpdatedReplicas,
			Available:       s.Status.AvailableReplicas,
			Desired:         s.Status.Replicas,
			Restarts:        restarts,
			Status:          status,
			Age:             CalculateAge(s.CreationTimestamp),
		}
	}
	return result
}

// MapDaemonSets 专门用于映射 DaemonSet 的函数
func MapDaemonSets(daemonSets []appsv1.DaemonSet) []model.DaemonSetStatus {
	result := make([]model.DaemonSetStatus, len(daemonSets))
	for i, d := range daemonSets {
		status := GetWorkloadStatus(d.Status.NumberReady, d.Status.DesiredNumberScheduled)
		restarts := CalculateDaemonSetRestarts(d)
		result[i] = model.DaemonSetStatus{
			Namespace:       d.Namespace,
			Name:            d.Name,
			ReadyReplicas:   d.Status.NumberReady,
			UpdatedReplicas: d.Status.UpdatedNumberScheduled,
			Available:       d.Status.NumberReady,
			Desired:         d.Status.DesiredNumberScheduled,
			Restarts:        restarts,
			Status:          status,
			Age:             CalculateAge(d.CreationTimestamp),
		}
	}
	return result
}

// MapJobs 专门用于映射 Job 的函数
func MapJobs(jobs []batchv1.Job) []model.JobStatus {
	result := make([]model.JobStatus, len(jobs))
	for i, j := range jobs {
		duration := CalculateDuration(j.Status.StartTime, j.Status.CompletionTime)
		restarts := CalculateJobRestarts(j)
		result[i] = model.JobStatus{
			Namespace:      j.Namespace,
			Name:           j.Name,
			Completions:    *j.Spec.Completions,
			Succeeded:      j.Status.Succeeded,
			Failed:         j.Status.Failed,
			Restarts:       restarts,
			StartTime:      formatTime(j.Status.StartTime),
			CompletionTime: formatTime(j.Status.CompletionTime),
			Duration:       duration,
			Status:         GetJobStatus(j.Status.Succeeded, j.Status.Failed, j.Status.Active),
			Age:            CalculateAge(j.CreationTimestamp),
		}
	}
	return result
}

// MapCronJobs 专门用于映射 CronJob 的函数
func MapCronJobs(cronJobs []batchv1.CronJob) []model.CronJobStatus {
	result := make([]model.CronJobStatus, len(cronJobs))
	for i, c := range cronJobs {
		result[i] = model.CronJobStatus{
			Namespace:        c.Namespace,
			Name:             c.Name,
			Schedule:         c.Spec.Schedule,
			Suspend:          *c.Spec.Suspend,
			Active:           len(c.Status.Active),
			LastScheduleTime: formatTime(c.Status.LastScheduleTime),
			Restarts:         0, // CronJob 本身不直接存储重启次数
			Status:           GetCronJobStatus(len(c.Status.Active), c.Status.LastSuccessfulTime),
			Age:              CalculateAge(c.CreationTimestamp),
		}
	}
	return result
}

// MapIngresses 专门用于映射 Ingress 的函数
func MapIngresses(ingresses []networkingv1.Ingress) []model.IngressStatus {
	result := make([]model.IngressStatus, len(ingresses))
	for i, ing := range ingresses {
		hosts := make([]string, 0)
		paths := make([]string, 0)
		targetServices := make([]string, 0)

		for _, rule := range ing.Spec.Rules {
			hosts = append(hosts, rule.Host)
			if rule.HTTP != nil {
				for _, path := range rule.HTTP.Paths {
					paths = append(paths, path.Path)
					if path.Backend.Service != nil {
						targetServices = append(targetServices, path.Backend.Service.Name)
					}
				}
			}
		}

		result[i] = model.IngressStatus{
			Namespace:     ing.Namespace,
			Name:          ing.Name,
			Class:         getClass(ing.Spec.IngressClassName),
			Hosts:         hosts,
			Address:       getAddress(ing.Status.LoadBalancer.Ingress),
			Ports:         []string{"80", "443"},
			Status:        model.StatusActive,
			Path:          paths,
			TargetService: targetServices,
			Age:           CalculateAge(ing.CreationTimestamp),
		}
	}
	return result
}

// formatTime 格式化时间
func formatTime(t *metav1.Time) string {
	if t == nil {
		return ""
	}
	return t.Format("2006-01-02 15:04:05")
}

// getClass 获取 Ingress 类别名
func getClass(className *string) string {
	if className == nil {
		return ""
	}
	return *className
}

// getAddress 获取 Ingress 地址
func getAddress(ingressList []networkingv1.IngressLoadBalancerIngress) string {
	if len(ingressList) == 0 {
		return ""
	}
	if ingressList[0].IP != "" {
		return ingressList[0].IP
	}
	return ingressList[0].Hostname
}

// MapNodes 专门用于映射 Node 的函数
func MapNodes(nodes []v1.Node, pods *v1.PodList, nodeMetricsMap model.NodeMetricsMap) []model.NodeStatus {
	result := make([]model.NodeStatus, len(nodes))
	for i, node := range nodes {
		status := "Unknown"
		ip := ""
		for _, addr := range node.Status.Addresses {
			if addr.Type == "InternalIP" {
				ip = addr.Address
				break
			}
		}
		for _, cond := range node.Status.Conditions {
			if cond.Type == "Ready" && cond.Status == "True" {
				status = "Active"
				break
			}
		}
		roles := []string{}
		for k := range node.Labels {
			if strings.HasPrefix(k, model.LabelNodeRolePrefix) {
				role := strings.TrimPrefix(k, model.LabelNodeRolePrefix)
				if role == "" {
					role = "worker"
				}
				roles = append(roles, role)
			}
		}
		if len(roles) == 0 {
			roles = append(roles, "worker")
		} else {
			// 排序 roles 数组，保证顺序一致
			sort.Strings(roles)
		}
		podsUsed := 0
		for _, pod := range pods.Items {
			if pod.Spec.NodeName == node.Name {
				podsUsed++
			}
		}
		podsCapacity := 0
		if v, ok := node.Status.Allocatable[v1.ResourceName(model.ResourcePods)]; ok {
			podsCapacity = int(v.Value())
		}
		metric := nodeMetricsMap[node.Name]
		cpuUsed := ParseCPU(metric.CPU)
		memUsed := ParseMemory(metric.Mem)
		cpuTotal := GetNodeAllocatableCPU(node)
		memTotal := GetNodeAllocatableMemory(node)
		cpuPercent := 0.0
		memPercent := 0.0
		if cpuTotal > 0 {
			cpuPercent = cpuUsed / cpuTotal * 100
		}
		if memTotal > 0 {
			memPercent = memUsed / memTotal * 100
		}
		result[i] = model.NodeStatus{
			Name:         node.Name,
			IP:           ip,
			Status:       status,
			CPUUsage:     math.Round(cpuPercent*10) / 10,
			MemoryUsage:  math.Round(memPercent*10) / 10,
			Role:         roles,
			PodsUsed:     podsUsed,
			PodsCapacity: podsCapacity,
			Age:          CalculateAge(node.CreationTimestamp),
		}
	}
	return result
}

// MapNamespaces 专门用于映射 Namespace 的函数
func MapNamespaces(namespaces []v1.Namespace) []model.NamespaceDetail {
	result := make([]model.NamespaceDetail, len(namespaces))
	for i, ns := range namespaces {
		result[i] = model.NamespaceDetail{
			Name:   ns.Name,
			Status: string(ns.Status.Phase),
			Labels: ns.Labels,
			Age:    CalculateAge(ns.CreationTimestamp),
		}
	}
	return result
}

// MapConfigMaps 专门用于映射 ConfigMap 的函数
func MapConfigMaps(configMaps []v1.ConfigMap) []model.ConfigMapStatus {
	result := make([]model.ConfigMapStatus, 0, len(configMaps))
	for _, cm := range configMaps {
		result = append(result, model.ConfigMapStatus{
			Namespace: cm.Namespace,
			Name:      cm.Name,
			DataCount: len(cm.Data),
			Keys:      ExtractKeys(cm.Data),
			Age:       CalculateAge(cm.CreationTimestamp),
		})
	}
	return result
}

// MapSecrets 专门用于映射 Secret 的函数
func MapSecrets(secrets []v1.Secret) []model.SecretStatus {
	result := make([]model.SecretStatus, 0, len(secrets))
	for _, secret := range secrets {
		result = append(result, model.SecretStatus{
			Namespace: secret.Namespace,
			Name:      secret.Name,
			Type:      string(secret.Type),
			DataCount: len(secret.Data),
			Keys:      ExtractKeys(secret.Data),
			Age:       CalculateAge(secret.CreationTimestamp),
		})
	}
	return result
}

// MapPVCs 专门用于映射 PVC 的函数
func MapPVCs(pvcs []v1.PersistentVolumeClaim) []model.PVCStatus {
	result := make([]model.PVCStatus, 0, len(pvcs))
	for _, pvc := range pvcs {
		status := "Pending"
		if pvc.Status.Phase == v1.ClaimBound {
			status = "Bound"
		} else if pvc.Status.Phase == v1.ClaimLost {
			status = "Lost"
		}

		capacity := ""
		if pvc.Status.Capacity != nil {
			if storage, ok := pvc.Status.Capacity[v1.ResourceStorage]; ok {
				capacity = storage.String()
			}
		}

		accessModes := make([]string, 0)
		for _, mode := range pvc.Spec.AccessModes {
			accessModes = append(accessModes, string(mode))
		}

		storageClass := ""
		if pvc.Spec.StorageClassName != nil {
			storageClass = *pvc.Spec.StorageClassName
		}

		volumeName := ""
		if pvc.Spec.VolumeName != "" {
			volumeName = pvc.Spec.VolumeName
		}

		result = append(result, model.PVCStatus{
			Namespace:    pvc.Namespace,
			Name:         pvc.Name,
			Status:       status,
			Capacity:     capacity,
			AccessMode:   strings.Join(accessModes, ","),
			StorageClass: storageClass,
			VolumeName:   volumeName,
			Age:          CalculateAge(pvc.CreationTimestamp),
		})
	}
	return result
}

// MapPVs 专门用于映射 PV 的函数
func MapPVs(pvs []v1.PersistentVolume) []model.PVStatus {
	result := make([]model.PVStatus, 0, len(pvs))
	for _, pv := range pvs {
		status := string(pv.Status.Phase)

		capacity := ""
		if pv.Spec.Capacity != nil {
			if storage, ok := pv.Spec.Capacity[v1.ResourceStorage]; ok {
				capacity = storage.String()
			}
		}

		accessModes := make([]string, 0)
		for _, mode := range pv.Spec.AccessModes {
			accessModes = append(accessModes, string(mode))
		}

		storageClass := ""
		if pv.Spec.StorageClassName != "" {
			storageClass = pv.Spec.StorageClassName
		}

		claimRef := ""
		if pv.Spec.ClaimRef != nil {
			claimRef = pv.Spec.ClaimRef.Namespace + "/" + pv.Spec.ClaimRef.Name
		}

		reclaimPolicy := string(pv.Spec.PersistentVolumeReclaimPolicy)

		result = append(result, model.PVStatus{
			Name:          pv.Name,
			Status:        status,
			Capacity:      capacity,
			AccessMode:    strings.Join(accessModes, ","),
			StorageClass:  storageClass,
			ClaimRef:      claimRef,
			ReclaimPolicy: reclaimPolicy,
			Age:           CalculateAge(pv.CreationTimestamp),
		})
	}
	return result
}

// MapStorageClasses 专门用于映射 StorageClass 的函数
func MapStorageClasses(storageClasses []storagev1.StorageClass) []model.StorageClassStatus {
	result := make([]model.StorageClassStatus, 0, len(storageClasses))
	for _, sc := range storageClasses {
		reclaimPolicy := ""
		if sc.ReclaimPolicy != nil {
			reclaimPolicy = string(*sc.ReclaimPolicy)
		}

		volumeBindingMode := ""
		if sc.VolumeBindingMode != nil {
			volumeBindingMode = string(*sc.VolumeBindingMode)
		}

		isDefault := false
		if sc.Annotations != nil {
			if _, ok := sc.Annotations[model.AnnotationStorageClassDefault]; ok {
				isDefault = true
			}
		}

		result = append(result, model.StorageClassStatus{
			Name:              sc.Name,
			Provisioner:       sc.Provisioner,
			ReclaimPolicy:     reclaimPolicy,
			VolumeBindingMode: volumeBindingMode,
			IsDefault:         isDefault,
			Age:               CalculateAge(sc.CreationTimestamp),
		})
	}
	return result
}

// MapEvents 专门用于映射 Event 的函数
func MapEvents(events []v1.Event) []model.EventStatus {
	result := make([]model.EventStatus, 0, len(events))
	for _, e := range events {
		// 构建对象引用
		objectRef := fmt.Sprintf("%s/%s", e.InvolvedObject.Kind, e.InvolvedObject.Name)
		if e.InvolvedObject.Namespace != "" {
			objectRef = fmt.Sprintf("%s/%s/%s", e.InvolvedObject.Namespace, e.InvolvedObject.Kind, e.InvolvedObject.Name)
		}

		result = append(result, model.EventStatus{
			Namespace: e.Namespace,
			Name:      e.Name,
			Reason:    e.Reason,
			Message:   e.Message,
			Type:      e.Type,
			Count:     e.Count,
			Object:    objectRef,
			Source:    e.Source.Component,
			FirstSeen: model.FormatTime(&e.FirstTimestamp),
			LastSeen:  model.FormatTime(&e.LastTimestamp),
			Duration:  e.LastTimestamp.Sub(e.FirstTimestamp.Time).String(),
		})
	}
	return result
}
