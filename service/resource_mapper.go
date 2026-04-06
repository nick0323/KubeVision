package service

import (
	"fmt"
	"strings"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	storagev1 "k8s.io/api/storage/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/nick0323/K8sVision/model"
)

// MapPods 映射 Pods
func MapPods(pods []corev1.Pod, podMetricsMap map[string]model.PodMetrics) []model.Pod {
	result := make([]model.Pod, len(pods))
	for i, pod := range pods {
		result[i] = model.Pod{
			Namespace: pod.Namespace,
			Name:      pod.Name,
			Status:    string(pod.Status.Phase),
			Ready:     CalculatePodReady(pod),
			Restarts:  CalculatePodRestarts(pod),
			Age:       CalculateAge(pod.CreationTimestamp),
			PodIP:     pod.Status.PodIP,
			NodeName:  pod.Spec.NodeName,
		}
	}
	return result
}

// MapDeployments 映射 Deployments
func MapDeployments(deployments []appsv1.Deployment) []model.Deployment {
	result := make([]model.Deployment, len(deployments))
	for i, d := range deployments {
		status := GetWorkloadStatus(d.Status.ReadyReplicas, d.Status.Replicas)
		restarts := CalculateDeploymentRestarts(d)
		result[i] = model.Deployment{
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

// MapStatefulSets 映射 StatefulSets
func MapStatefulSets(sts []appsv1.StatefulSet) []model.StatefulSet {
	result := make([]model.StatefulSet, len(sts))
	for i, s := range sts {
		status := GetWorkloadStatus(s.Status.ReadyReplicas, s.Status.Replicas)
		restarts := CalculateStatefulSetRestarts(s)
		result[i] = model.StatefulSet{
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

// MapDaemonSets 映射 DaemonSets
func MapDaemonSets(ds []appsv1.DaemonSet) []model.DaemonSet {
	result := make([]model.DaemonSet, len(ds))
	for i, d := range ds {
		status := GetWorkloadStatus(d.Status.NumberReady, d.Status.DesiredNumberScheduled)
		restarts := CalculateDaemonSetRestarts(d)
		result[i] = model.DaemonSet{
			Namespace:       d.Namespace,
			Name:            d.Name,
			ReadyReplicas:   d.Status.NumberReady,
			UpdatedReplicas: d.Status.UpdatedNumberScheduled,
			Available:       d.Status.NumberAvailable,
			Desired:         d.Status.DesiredNumberScheduled,
			Restarts:        restarts,
			Status:          status,
			Age:             CalculateAge(d.CreationTimestamp),
		}
	}
	return result
}

// MapServices 映射 Services
func MapServices(services []corev1.Service) []model.Service {
	result := make([]model.Service, len(services))
	for i, s := range services {
		result[i] = model.Service{
			Namespace: s.Namespace,
			Name:      s.Name,
			Type:      string(s.Spec.Type),
			ClusterIP: s.Spec.ClusterIP,
			Ports:     getServicePorts(s),
			Age:       CalculateAge(s.CreationTimestamp),
		}
	}
	return result
}

// MapConfigMaps 映射 ConfigMaps
func MapConfigMaps(cms []corev1.ConfigMap) []model.ConfigMap {
	result := make([]model.ConfigMap, len(cms))
	for i, c := range cms {
		keys := make([]string, 0, len(c.Data))
		for k := range c.Data {
			keys = append(keys, k)
		}
		result[i] = model.ConfigMap{
			Namespace: c.Namespace,
			Name:      c.Name,
			DataCount: len(c.Data) + len(c.BinaryData),
			Keys:      keys,
			Age:       CalculateAge(c.CreationTimestamp),
		}
	}
	return result
}

// MapSecrets 映射 Secrets
func MapSecrets(secrets []corev1.Secret) []model.Secret {
	result := make([]model.Secret, len(secrets))
	for i, s := range secrets {
		keys := make([]string, 0, len(s.Data))
		for k := range s.Data {
			keys = append(keys, k)
		}
		result[i] = model.Secret{
			Namespace: s.Namespace,
			Name:      s.Name,
			Type:      string(s.Type),
			DataCount: len(s.Data),
			Keys:      keys,
			Age:       CalculateAge(s.CreationTimestamp),
		}
	}
	return result
}

// MapIngresses 映射 Ingresses
func MapIngresses(ingresses []networkingv1.Ingress) []model.Ingress {
	result := make([]model.Ingress, len(ingresses))
	for i, ing := range ingresses {
		hosts := getIngressHosts(ing)
		result[i] = model.Ingress{
			Namespace:     ing.Namespace,
			Name:          ing.Name,
			Class:         getIngressClass(ing),
			Hosts:         hosts,
			Ports:         getIngressPorts(ing),
			Status:        getIngressStatus(ing),
			Path:          getIngressPaths(ing),
			TargetService: getIngressServices(ing),
			Age:           CalculateAge(ing.CreationTimestamp),
		}
	}
	return result
}

// MapJobs 映射 Jobs
func MapJobs(jobs []batchv1.Job) []model.Job {
	result := make([]model.Job, len(jobs))
	for i, j := range jobs {
		result[i] = model.Job{
			Namespace:      j.Namespace,
			Name:           j.Name,
			Completions:    *j.Spec.Completions,
			Succeeded:      j.Status.Succeeded,
			Failed:         j.Status.Failed,
			Restarts:       calculateJobRestarts(j),
			StartTime:      model.FormatTime(j.Status.StartTime),
			CompletionTime: model.FormatTime(j.Status.CompletionTime),
			Duration:       calculateJobDuration(j),
			Status:         getJobStatus(j),
			Age:            CalculateAge(j.CreationTimestamp),
		}
	}
	return result
}

// MapCronJobs 映射 CronJobs
func MapCronJobs(cronJobs []batchv1.CronJob) []model.CronJob {
	result := make([]model.CronJob, len(cronJobs))
	for i, cj := range cronJobs {
		result[i] = model.CronJob{
			Namespace:        cj.Namespace,
			Name:             cj.Name,
			Schedule:         cj.Spec.Schedule,
			Suspend:          *cj.Spec.Suspend,
			Active:           len(cj.Status.Active),
			LastScheduleTime: model.FormatTime(cj.Status.LastScheduleTime),
			Restarts:         calculateCronJobRestarts(cj),
			Status:           getCronJobStatus(cj),
			Age:              CalculateAge(cj.CreationTimestamp),
		}
	}
	return result
}

// calculateCronJobRestarts 计算 CronJob 关联 Job 的总重启次数
func calculateCronJobRestarts(cj batchv1.CronJob) int32 {
	// CronJob 本身没有重启次数，需要统计其创建的 Job 的重启次数
	// 这里简化处理，返回 0 或根据 Active Job 计算
	var totalRestarts int32
	// 如果有活跃的 Job，可以查询其重启次数（需要额外的 API 调用）
	// 为简化，这里返回 0
	_ = totalRestarts
	return 0
}

// MapPVCs 映射 PVCs
func MapPVCs(pvcs []corev1.PersistentVolumeClaim) []model.PVC {
	result := make([]model.PVC, len(pvcs))
	for i, p := range pvcs {
		capacity := "-"
		if qty, ok := p.Status.Capacity[corev1.ResourceStorage]; ok {
			capacity = qty.String()
		}
		result[i] = model.PVC{
			Namespace:    p.Namespace,
			Name:         p.Name,
			Status:       string(p.Status.Phase),
			Capacity:     capacity,
			AccessMode:   getAccessMode(p),
			StorageClass: getStorageClassName(p),
			VolumeName:   p.Spec.VolumeName,
			Age:          CalculateAge(p.CreationTimestamp),
		}
	}
	return result
}

// MapPVs 映射 PVs
func MapPVs(pvs []corev1.PersistentVolume) []model.PV {
	result := make([]model.PV, len(pvs))
	for i, p := range pvs {
		capacity := "-"
		if qty, ok := p.Spec.Capacity[corev1.ResourceStorage]; ok {
			capacity = qty.String()
		}
		result[i] = model.PV{
			Name:          p.Name,
			Status:        string(p.Status.Phase),
			Capacity:      capacity,
			AccessMode:    getPVAccessMode(p),
			StorageClass:  p.Spec.StorageClassName,
			ClaimRef:      getClaimRef(p),
			ReclaimPolicy: string(p.Spec.PersistentVolumeReclaimPolicy),
			Age:           CalculateAge(p.CreationTimestamp),
		}
	}
	return result
}

// MapStorageClasses 映射 StorageClasses
func MapStorageClasses(scs []storagev1.StorageClass) []model.StorageClass {
	result := make([]model.StorageClass, len(scs))
	for i, s := range scs {
		result[i] = model.StorageClass{
			Name:              s.Name,
			Provisioner:       s.Provisioner,
			ReclaimPolicy:     getReclaimPolicy(s),
			VolumeBindingMode: string(*s.VolumeBindingMode),
			IsDefault:         isDefaultStorageClass(s),
			Age:               CalculateAge(s.CreationTimestamp),
		}
	}
	return result
}

// MapNamespaces 映射 Namespaces
func MapNamespaces(namespaces []corev1.Namespace) []model.Namespace {
	result := make([]model.Namespace, len(namespaces))
	for i, n := range namespaces {
		result[i] = model.Namespace{
			Name:   n.Name,
			Status: string(n.Status.Phase),
			Age:    CalculateAge(n.CreationTimestamp),
		}
	}
	return result
}

// MapNodes 映射 Nodes
func MapNodes(nodes []corev1.Node, pods *corev1.PodList, metrics map[string]model.NodeMetrics) []model.Node {
	result := make([]model.Node, len(nodes))
	for i, n := range nodes {
		cpuUsage := 0.0
		memUsage := 0.0

		// 如果有 metrics 数据，计算使用率
		if metrics != nil {
			if m, ok := metrics[n.Name]; ok {
				// 计算 CPU 使用率
				cpuCapacity := n.Status.Capacity.Cpu()
				if !cpuCapacity.IsZero() && m.CPU != "" {
					cpuUsage = calculateCPUUsage(cpuCapacity, m.CPU)
				}
				// 计算内存使用率
				memCapacity := n.Status.Capacity.Memory()
				if !memCapacity.IsZero() && m.Memory != "" {
					memUsage = calculateMemoryUsage(memCapacity, m.Memory)
				}
			}
		}

		result[i] = model.Node{
			Name:         n.Name,
			IP:           getInternalIP(n),
			Status:       getNodeStatus(n),
			CPUUsage:     cpuUsage,
			MemoryUsage:  memUsage,
			Role:         getNodeRoles(n),
			PodsUsed:     countPodsOnNode(pods, n.Name),
			PodsCapacity: int(n.Status.Capacity.Pods().Value()),
			Age:          CalculateAge(n.CreationTimestamp),
		}
	}
	return result
}

// calculateCPUUsage 计算 CPU 使用率
func calculateCPUUsage(capacity *resource.Quantity, usageStr string) float64 {
	usage, err := resource.ParseQuantity(usageStr)
	if err != nil {
		return 0
	}

	capMilli := float64(capacity.MilliValue())
	usageMilli := float64(usage.MilliValue())

	if capMilli <= 0 {
		return 0
	}

	return (usageMilli / capMilli) * 100
}

// calculateMemoryUsage 计算内存使用率
func calculateMemoryUsage(capacity *resource.Quantity, usageStr string) float64 {
	usage, err := resource.ParseQuantity(usageStr)
	if err != nil {
		return 0
	}

	capBytes := capacity.Value()
	usageBytes := usage.Value()

	if capBytes <= 0 {
		return 0
	}

	return (float64(usageBytes) / float64(capBytes)) * 100
}

// MapEndpoints 映射 Endpoints
func MapEndpoints(endpoints []corev1.Endpoints) []model.Endpoints {
	result := make([]model.Endpoints, len(endpoints))
	for i, e := range endpoints {
		result[i] = model.Endpoints{
			Namespace: e.Namespace,
			Name:      e.Name,
			Addresses: countEndpointAddresses(e),
			Ports:     getEndpointPorts(e),
			Age:       CalculateAge(e.CreationTimestamp),
		}
	}
	return result
}

// MapEvents 映射 Events
func MapEvents(events []corev1.Event) []model.Event {
	result := make([]model.Event, len(events))
	for i, e := range events {
		result[i] = model.Event{
			Namespace: e.Namespace,
			Name:      e.Name,
			Reason:    e.Reason,
			Message:   e.Message,
			Type:      e.Type,
			Count:     e.Count,
			Object:    e.InvolvedObject.Kind + "/" + e.InvolvedObject.Name,
			Source:    e.Source.Component,
			LastSeen:  model.FormatTime(&e.LastTimestamp),
			Duration:  calculateEventDuration(e),
			Age:       CalculateAge(e.CreationTimestamp),
		}
	}
	return result
}

// 辅助函数

// CalculateAge 计算资源存在时间
func CalculateAge(t metav1.Time) string {
	return model.FormatAge(t.Time)
}

// CalculatePodReady 计算 Pod Ready 状态
func CalculatePodReady(pod corev1.Pod) string {
	readyContainers := 0
	totalContainers := len(pod.Status.ContainerStatuses)

	for _, cs := range pod.Status.ContainerStatuses {
		if cs.Ready {
			readyContainers++
		}
	}

	return fmt.Sprintf("%d/%d", readyContainers, totalContainers)
}

// CalculatePodRestarts 计算 Pod 重启次数
func CalculatePodRestarts(pod corev1.Pod) int32 {
	var restarts int32
	for _, cs := range pod.Status.ContainerStatuses {
		restarts += cs.RestartCount
	}
	return restarts
}

// CalculateDeploymentRestarts 计算 Deployment 重启次数
func CalculateDeploymentRestarts(d appsv1.Deployment) int32 {
	// 使用 PodInformer 获取精确的重启次数
	if len(d.OwnerReferences) > 0 {
		return GetRestartsByOwnerReference(d.OwnerReferences[0])
	}
	return 0
}

// CalculateStatefulSetRestarts 计算 StatefulSet 重启次数
func CalculateStatefulSetRestarts(s appsv1.StatefulSet) int32 {
	// 使用 PodInformer 获取精确的重启次数
	if len(s.OwnerReferences) > 0 {
		return GetRestartsByOwnerReference(s.OwnerReferences[0])
	}
	return 0
}

// CalculateDaemonSetRestarts 计算 DaemonSet 重启次数
func CalculateDaemonSetRestarts(d appsv1.DaemonSet) int32 {
	// 使用 PodInformer 获取精确的重启次数
	if len(d.OwnerReferences) > 0 {
		return GetRestartsByOwnerReference(d.OwnerReferences[0])
	}
	return 0
}

func getExternalIP(svc corev1.Service) string {
	if len(svc.Status.LoadBalancer.Ingress) > 0 {
		if ip := svc.Status.LoadBalancer.Ingress[0].IP; ip != "" {
			return ip
		}
		if hostname := svc.Status.LoadBalancer.Ingress[0].Hostname; hostname != "" {
			return hostname
		}
	}
	if len(svc.Spec.ExternalIPs) > 0 {
		return svc.Spec.ExternalIPs[0]
	}
	return "-"
}

func getServicePorts(svc corev1.Service) []string {
	ports := make([]string, 0, len(svc.Spec.Ports))
	for _, p := range svc.Spec.Ports {
		ports = append(ports, fmt.Sprintf("%d", p.Port))
	}
	return ports
}

func getIngressHosts(ing networkingv1.Ingress) []string {
	hosts := make([]string, 0)
	for _, r := range ing.Spec.Rules {
		hosts = append(hosts, r.Host)
	}
	if len(hosts) == 0 {
		hosts = append(hosts, "*")
	}
	return hosts
}

func getIngressClass(ing networkingv1.Ingress) string {
	if ing.Spec.IngressClassName != nil {
		return *ing.Spec.IngressClassName
	}
	return "-"
}

func getIngressPorts(ing networkingv1.Ingress) []string {
	ports := make([]string, 0)
	for _, rule := range ing.Spec.Rules {
		if rule.HTTP != nil {
			for _, path := range rule.HTTP.Paths {
				if path.Backend.Service != nil && path.Backend.Service.Port.Number > 0 {
					ports = append(ports, fmt.Sprintf("%d", path.Backend.Service.Port.Number))
				}
			}
		}
	}
	if len(ports) == 0 {
		// 尝试从 TLS 配置获取
		if len(ing.Spec.TLS) > 0 {
			ports = append(ports, "443")
		}
	}
	if len(ports) == 0 {
		ports = append(ports, "80")
	}
	return ports
}

func getIngressAddress(ing networkingv1.Ingress) string {
	if len(ing.Status.LoadBalancer.Ingress) > 0 {
		if ip := ing.Status.LoadBalancer.Ingress[0].IP; ip != "" {
			return ip
		}
	}
	return "-"
}

func getIngressStatus(ing networkingv1.Ingress) string {
	if len(ing.Status.LoadBalancer.Ingress) > 0 {
		return "Ready"
	}
	return "Pending"
}

func getIngressPaths(ing networkingv1.Ingress) []string {
	paths := make([]string, 0)
	for _, r := range ing.Spec.Rules {
		if r.HTTP != nil {
			for _, p := range r.HTTP.Paths {
				paths = append(paths, p.Path)
			}
		}
	}
	return paths
}

func getIngressServices(ing networkingv1.Ingress) []string {
	services := make([]string, 0)
	for _, r := range ing.Spec.Rules {
		if r.HTTP != nil {
			for _, p := range r.HTTP.Paths {
				if p.Backend.Service != nil {
					services = append(services, p.Backend.Service.Name)
				}
			}
		}
	}
	return services
}

func calculateJobRestarts(j batchv1.Job) int32 {
	return j.Status.Failed
}

func calculateJobDuration(j batchv1.Job) string {
	if j.Status.StartTime == nil {
		return "-"
	}
	endTime := j.Status.CompletionTime
	if endTime == nil {
		endTime = &metav1.Time{Time: time.Now()}
	}
	return endTime.Sub(j.Status.StartTime.Time).Round(time.Second).String()
}

func getJobStatus(j batchv1.Job) string {
	if j.Status.Failed > 0 {
		return "Failed"
	}
	if j.Status.Succeeded >= *j.Spec.Completions {
		return "Completed"
	}
	if j.Status.StartTime != nil {
		return "Running"
	}
	return "Pending"
}

func getCronJobStatus(cj batchv1.CronJob) string {
	if *cj.Spec.Suspend {
		return "Suspended"
	}
	return "Active"
}

func getAccessMode(pvc corev1.PersistentVolumeClaim) string {
	if len(pvc.Spec.AccessModes) > 0 {
		return string(pvc.Spec.AccessModes[0])
	}
	return "-"
}

func getStorageClassName(pvc corev1.PersistentVolumeClaim) string {
	if pvc.Spec.StorageClassName != nil {
		return *pvc.Spec.StorageClassName
	}
	return "-"
}

func getPVAccessMode(pv corev1.PersistentVolume) string {
	if len(pv.Spec.AccessModes) > 0 {
		return string(pv.Spec.AccessModes[0])
	}
	return "-"
}

func getClaimRef(pv corev1.PersistentVolume) string {
	if pv.Spec.ClaimRef != nil {
		return pv.Spec.ClaimRef.Namespace + "/" + pv.Spec.ClaimRef.Name
	}
	return "-"
}

func getReclaimPolicy(sc storagev1.StorageClass) string {
	if sc.ReclaimPolicy != nil {
		return string(*sc.ReclaimPolicy)
	}
	return "Delete"
}

func isDefaultStorageClass(sc storagev1.StorageClass) bool {
	return sc.Annotations["storageclass.kubernetes.io/is-default-class"] == "true"
}

func getInternalIP(node corev1.Node) string {
	for _, addr := range node.Status.Addresses {
		if addr.Type == corev1.NodeInternalIP {
			return addr.Address
		}
	}
	return "-"
}

func getNodeStatus(node corev1.Node) string {
	for _, condition := range node.Status.Conditions {
		if condition.Type == corev1.NodeReady {
			if condition.Status == corev1.ConditionTrue {
				return "Ready"
			}
			return "NotReady"
		}
	}
	return "Unknown"
}

func getNodeRoles(node corev1.Node) []string {
	roles := make([]string, 0)
	hasControlPlane := false
	hasEtcd := false
	hasWorker := false

	// 遍历所有 node-role.kubernetes.io/ 开头的 labels
	for k := range node.Labels {
		if strings.HasPrefix(k, "node-role.kubernetes.io/") {
			role := strings.TrimPrefix(k, "node-role.kubernetes.io/")
			switch role {
			case "controlplane":
				hasControlPlane = true
			case "etcd":
				hasEtcd = true
			case "worker":
				hasWorker = true
			default:
				roles = append(roles, role)
			}
		}
	}

	// 按优先级添加角色：control-plane > etcd > worker
	if hasControlPlane {
		roles = append(roles, "control-plane")
	}
	if hasEtcd {
		roles = append(roles, "etcd")
	}
	if hasWorker {
		roles = append(roles, "worker")
	}

	// 如果没有设置任何角色标签，默认为 worker
	if len(roles) == 0 {
		roles = append(roles, "worker")
	}

	return roles
}

func countPodsOnNode(pods *corev1.PodList, nodeName string) int {
	if pods == nil {
		return 0
	}
	count := 0
	for _, p := range pods.Items {
		if p.Spec.NodeName == nodeName {
			count++
		}
	}
	return count
}

func countEndpointAddresses(ep corev1.Endpoints) int {
	count := 0
	for _, subset := range ep.Subsets {
		count += len(subset.Addresses)
	}
	return count
}

func getEndpointPorts(ep corev1.Endpoints) []string {
	ports := make([]string, 0)
	for _, subset := range ep.Subsets {
		for _, p := range subset.Ports {
			ports = append(ports, string(p.Port))
		}
	}
	return ports
}

func calculateEventDuration(e corev1.Event) string {
	if e.FirstTimestamp.IsZero() {
		return "-"
	}
	endTime := e.LastTimestamp.Time
	if e.LastTimestamp.IsZero() {
		endTime = e.EventTime.Time
	}
	return endTime.Sub(e.FirstTimestamp.Time).Round(time.Second).String()
}
