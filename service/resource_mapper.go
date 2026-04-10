package service

import (
	"fmt"
	"strconv"
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

// ============================================================================
// 主映射函数（严格匹配 model 包字段定义）
// ============================================================================

// MapPods 映射 Pods
func MapPods(pods []corev1.Pod, podMetricsMap map[string]model.PodMetrics) []model.Pod {
	result := make([]model.Pod, len(pods))
	for i, pod := range pods {
		result[i] = model.Pod{
			Namespace: pod.Namespace,
			Name:      pod.Name,
			Status:    getPodPhaseDisplay(pod.Status.Phase),
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
		// GetWorkloadStatus 已在 common.go 定义，直接调用
		status := GetWorkloadStatus(d.Status.ReadyReplicas, d.Status.Replicas)
		result[i] = model.Deployment{
			Namespace:       d.Namespace,
			Name:            d.Name,
			ReadyReplicas:   d.Status.ReadyReplicas,
			UpdatedReplicas: d.Status.UpdatedReplicas,
			Available:       d.Status.AvailableReplicas,
			Desired:         d.Status.Replicas,
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
		result[i] = model.StatefulSet{
			Namespace:       s.Namespace,
			Name:            s.Name,
			ReadyReplicas:   s.Status.ReadyReplicas,
			UpdatedReplicas: s.Status.UpdatedReplicas,
			Available:       s.Status.AvailableReplicas,
			Desired:         s.Status.Replicas,
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
		result[i] = model.DaemonSet{
			Namespace:       d.Namespace,
			Name:            d.Name,
			ReadyReplicas:   d.Status.NumberReady,
			UpdatedReplicas: d.Status.UpdatedNumberScheduled,
			Available:       d.Status.NumberAvailable,
			Desired:         d.Status.DesiredNumberScheduled,
			Status:          status,
			Age:             CalculateAge(d.CreationTimestamp),
		}
	}
	return result
}

// MapServices 映射 Services（仅使用 model.Service 实际存在的字段）
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
		keys := make([]string, 0, len(c.Data)+len(c.BinaryData))
		for k := range c.Data {
			keys = append(keys, k)
		}
		for k := range c.BinaryData {
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

func MapIngresses(ingresses []networkingv1.Ingress) []model.Ingress {
	result := make([]model.Ingress, len(ingresses))
	for i, ing := range ingresses {
		result[i] = model.Ingress{
			Namespace:     ing.Namespace,
			Name:          ing.Name,
			Class:         getIngressClass(ing),
			Hosts:         getIngressHosts(ing),
			Ports:         getIngressPorts(ing),
			Path:          getIngressPaths(ing),    // ✅ 恢复返回 []string
			TargetService: getIngressServices(ing), // ✅ 恢复返回 []string
			Age:           CalculateAge(ing.CreationTimestamp),
		}
	}
	return result
}

// MapJobs 映射 Jobs
func MapJobs(jobs []batchv1.Job) []model.Job {
	result := make([]model.Job, len(jobs))
	for i, j := range jobs {
		completions := int32(1)
		if j.Spec.Completions != nil {
			completions = *j.Spec.Completions
		}
		result[i] = model.Job{
			Namespace:      j.Namespace,
			Name:           j.Name,
			Completions:    completions,
			Succeeded:      j.Status.Succeeded,
			Failed:         j.Status.Failed,
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
		suspend := false
		if cj.Spec.Suspend != nil {
			suspend = *cj.Spec.Suspend
		}
		result[i] = model.CronJob{
			Namespace:        cj.Namespace,
			Name:             cj.Name,
			Schedule:         cj.Spec.Schedule,
			Suspend:          suspend,
			Active:           len(cj.Status.Active),
			LastScheduleTime: model.FormatTime(cj.Status.LastScheduleTime),
			Status:           getCronJobStatus(cj),
			Age:              CalculateAge(cj.CreationTimestamp),
		}
	}
	return result
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
			ReclaimPolicy: getReclaimPolicyStr(p),
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
			VolumeBindingMode: getVolumeBindingMode(s),
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
	// 预构建 nodeName → podCount 映射，优化性能
	podCountMap := buildNodePodCountMap(pods)

	result := make([]model.Node, len(nodes))
	for i, n := range nodes {
		cpuUsage := 0.0
		memUsage := 0.0

		if metrics != nil {
			if m, ok := metrics[n.Name]; ok {
				cpuUsage = calculateCPUUsage(n.Status.Capacity.Cpu(), m.CPU)
				memUsage = calculateMemoryUsage(n.Status.Capacity.Memory(), m.Memory)
			}
		}

		result[i] = model.Node{
			Name:         n.Name,
			IP:           getInternalIP(n),
			Status:       getNodeStatus(n),
			CPUUsage:     cpuUsage,
			MemoryUsage:  memUsage,
			Role:         getNodeRoles(n),
			PodsUsed:     podCountMap[n.Name],
			PodsCapacity: int(n.Status.Capacity.Pods().Value()),
			Age:          CalculateAge(n.CreationTimestamp),
		}
	}
	return result
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
			Object:    formatEventObject(e.InvolvedObject),
			Source:    e.Source.Component,
			LastSeen:  formatEventLastSeen(e),
			Duration:  calculateEventDuration(e),
			Age:       CalculateAge(e.CreationTimestamp),
		}
	}
	return result
}

// ============================================================================
// 辅助函数 - 计算类
// ============================================================================

// CalculateAge 计算资源存在时间
func CalculateAge(t metav1.Time) string {
	return model.FormatAge(t.Time)
}

// CalculatePodReady 计算 Pod Ready 状态（优化显示）
func CalculatePodReady(pod corev1.Pod) string {
	total := len(pod.Status.ContainerStatuses)
	if total == 0 {
		// 根据 Phase 显示更友好的状态
		switch pod.Status.Phase {
		case corev1.PodPending:
			return "Pending"
		case corev1.PodFailed:
			return "Failed"
		case corev1.PodSucceeded:
			return "Succeeded"
		default:
			return "N/A"
		}
	}

	ready := 0
	for _, cs := range pod.Status.ContainerStatuses {
		if cs.Ready {
			ready++
		}
	}
	return fmt.Sprintf("%d/%d", ready, total)
}

// CalculatePodRestarts 计算 Pod 重启次数
func CalculatePodRestarts(pod corev1.Pod) int32 {
	var restarts int32
	for _, cs := range pod.Status.ContainerStatuses {
		restarts += cs.RestartCount
	}
	return restarts
}

// calculateCPUUsage 计算 CPU 使用率
func calculateCPUUsage(capacity *resource.Quantity, usageStr string) float64 {
	if capacity == nil || capacity.IsZero() || usageStr == "" {
		return 0
	}

	usage, err := resource.ParseQuantity(usageStr)
	if err != nil {
		return -1 // 返回哨兵值，前端显示 "N/A"
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
	if capacity == nil || capacity.IsZero() || usageStr == "" {
		return 0
	}

	usage, err := resource.ParseQuantity(usageStr)
	if err != nil {
		return -1
	}

	capBytes := capacity.Value()
	usageBytes := usage.Value()

	if capBytes <= 0 {
		return 0
	}
	return (float64(usageBytes) / float64(capBytes)) * 100
}

// calculateJobDuration 计算 Job 执行时长
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

// calculateEventDuration 计算事件持续时间
func calculateEventDuration(e corev1.Event) string {
	if e.FirstTimestamp.IsZero() {
		return "-"
	}
	endTime := e.LastTimestamp.Time
	if e.LastTimestamp.IsZero() {
		endTime = e.EventTime.Time
	}
	if endTime.IsZero() {
		return "-"
	}
	return endTime.Sub(e.FirstTimestamp.Time).Round(time.Second).String()
}

// ============================================================================
// 辅助函数 - 状态判断类
// ============================================================================

// getPodPhaseDisplay 返回更友好的 Pod 状态显示
func getPodPhaseDisplay(phase corev1.PodPhase) string {
	switch phase {
	case corev1.PodRunning:
		return "Running"
	case corev1.PodPending:
		return "Pending"
	case corev1.PodSucceeded:
		return "Succeeded"
	case corev1.PodFailed:
		return "Failed"
	case corev1.PodUnknown:
		return "Unknown"
	default:
		return string(phase)
	}
}

// getJobStatus 获取 Job 状态（✅ 修复指针解引用）
func getJobStatus(j batchv1.Job) string {
	// Completions 默认值为 1
	completions := int32(1)
	if j.Spec.Completions != nil {
		completions = *j.Spec.Completions
	}

	if j.Status.Failed > 0 {
		return "Failed"
	}
	if j.Status.Succeeded >= completions {
		return "Completed"
	}
	if j.Status.StartTime != nil {
		return "Running"
	}
	return "Pending"
}

// getCronJobStatus 获取 CronJob 状态（✅ 修复指针解引用）
func getCronJobStatus(cj batchv1.CronJob) string {
	// Suspend 默认值为 false
	if cj.Spec.Suspend != nil && *cj.Spec.Suspend {
		return "Suspended"
	}
	return "Active"
}

// getNodeStatus 获取 Node 状态
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

// getIngressStatus 获取 Ingress 状态
func getIngressStatus(ing networkingv1.Ingress) string {
	if len(ing.Status.LoadBalancer.Ingress) > 0 {
		return "Ready"
	}
	return "Pending"
}

// ============================================================================
// 辅助函数 - 字段提取类
// ============================================================================

// getServicePorts 提取 Service 端口
func getServicePorts(svc corev1.Service) []string {
	if len(svc.Spec.Ports) == 0 {
		return []string{"-"}
	}
	ports := make([]string, 0, len(svc.Spec.Ports))
	for _, p := range svc.Spec.Ports {
		ports = append(ports, fmt.Sprintf("%d", p.Port))
	}
	return ports
}

// getIngressHosts 提取 Ingress 主机名
func getIngressHosts(ing networkingv1.Ingress) []string {
	if len(ing.Spec.Rules) == 0 {
		return []string{"*"}
	}
	hosts := make([]string, 0, len(ing.Spec.Rules))
	for _, r := range ing.Spec.Rules {
		if r.Host != "" {
			hosts = append(hosts, r.Host)
		}
	}
	if len(hosts) == 0 {
		return []string{"*"}
	}
	return hosts
}

// getIngressClass 提取 Ingress 类
func getIngressClass(ing networkingv1.Ingress) string {
	if ing.Spec.IngressClassName != nil {
		return *ing.Spec.IngressClassName
	}
	// 兼容旧版 annotation
	if class, ok := ing.Annotations["kubernetes.io/ingress.class"]; ok {
		return class
	}
	return "-"
}

// getIngressPorts 提取 Ingress 端口（保持原函数名，返回后端服务端口）
func getIngressPorts(ing networkingv1.Ingress) []string {
	ports := make([]string, 0)
	for _, rule := range ing.Spec.Rules {
		if rule.HTTP != nil {
			for _, path := range rule.HTTP.Paths {
				if path.Backend.Service != nil {
					if path.Backend.Service.Port.Number > 0 {
						ports = append(ports, fmt.Sprintf("%d", path.Backend.Service.Port.Number))
					} else if path.Backend.Service.Port.Name != "" {
						ports = append(ports, path.Backend.Service.Port.Name)
					}
				}
			}
		}
	}
	if len(ports) == 0 {
		// 尝试从 TLS 配置推断
		if len(ing.Spec.TLS) > 0 {
			return []string{"443"}
		}
		return []string{"80"}
	}
	return ports
}

func getIngressPaths(ing networkingv1.Ingress) []string {
	paths := make([]string, 0)
	for _, r := range ing.Spec.Rules {
		if r.HTTP != nil {
			for _, p := range r.HTTP.Paths {
				if p.Path != "" {
					paths = append(paths, p.Path)
				}
			}
		}
	}
	if len(paths) == 0 {
		return []string{"/"}
	}
	return paths
}

func getIngressServices(ing networkingv1.Ingress) []string {
	services := make([]string, 0)
	for _, r := range ing.Spec.Rules {
		if r.HTTP != nil {
			for _, p := range r.HTTP.Paths {
				if p.Backend.Service != nil && p.Backend.Service.Name != "" {
					services = append(services, p.Backend.Service.Name)
				}
			}
		}
	}
	if len(services) == 0 {
		return []string{"-"}
	}
	return services
}

// getAccessMode 提取 PVC 访问模式
func getAccessMode(pvc corev1.PersistentVolumeClaim) string {
	if len(pvc.Spec.AccessModes) > 0 {
		return string(pvc.Spec.AccessModes[0])
	}
	return "-"
}

// getStorageClassName 提取 PVC 存储类
func getStorageClassName(pvc corev1.PersistentVolumeClaim) string {
	if pvc.Spec.StorageClassName != nil {
		return *pvc.Spec.StorageClassName
	}
	// 兼容旧版 annotation
	if class, ok := pvc.Annotations["volume.beta.kubernetes.io/storage-class"]; ok {
		return class
	}
	return "-"
}

// getPVAccessMode 提取 PV 访问模式
func getPVAccessMode(pv corev1.PersistentVolume) string {
	if len(pv.Spec.AccessModes) > 0 {
		return string(pv.Spec.AccessModes[0])
	}
	return "-"
}

// getClaimRef 提取 PV 绑定引用
func getClaimRef(pv corev1.PersistentVolume) string {
	if pv.Spec.ClaimRef != nil && pv.Spec.ClaimRef.Name != "" {
		ns := pv.Spec.ClaimRef.Namespace
		if ns == "" {
			ns = "default"
		}
		return fmt.Sprintf("%s/%s", ns, pv.Spec.ClaimRef.Name)
	}
	return "-"
}

// getReclaimPolicy 提取 StorageClass 回收策略
func getReclaimPolicy(sc storagev1.StorageClass) string {
	if sc.ReclaimPolicy != nil {
		return string(*sc.ReclaimPolicy)
	}
	return "Delete"
}

// getReclaimPolicyStr 提取 PV 回收策略
func getReclaimPolicyStr(pv corev1.PersistentVolume) string {
	if pv.Spec.PersistentVolumeReclaimPolicy != "" {
		return string(pv.Spec.PersistentVolumeReclaimPolicy)
	}
	return "Delete"
}

// getVolumeBindingMode 提取 StorageClass 绑定模式
func getVolumeBindingMode(sc storagev1.StorageClass) string {
	if sc.VolumeBindingMode != nil {
		return string(*sc.VolumeBindingMode)
	}
	return "Immediate"
}

// isDefaultStorageClass 判断是否为默认存储类
func isDefaultStorageClass(sc storagev1.StorageClass) bool {
	return sc.Annotations["storageclass.kubernetes.io/is-default-class"] == "true"
}

// getInternalIP 获取 Node 内网 IP
func getInternalIP(node corev1.Node) string {
	for _, addr := range node.Status.Addresses {
		if addr.Type == corev1.NodeInternalIP && addr.Address != "" {
			return addr.Address
		}
	}
	return "-"
}

// getNodeRoles 获取 Node 角色（✅ 兼容新旧标签）
func getNodeRoles(node corev1.Node) []string {
	roles := make([]string, 0)
	hasControlPlane := false
	hasWorker := false

	for k := range node.Labels {
		if strings.HasPrefix(k, "node-role.kubernetes.io/") {
			role := strings.TrimPrefix(k, "node-role.kubernetes.io/")
			// 兼容新旧控制面标签
			if role == "control-plane" || role == "master" || role == "" {
				hasControlPlane = true
			} else if role == "worker" {
				hasWorker = true
			} else if role != "" {
				roles = append(roles, role)
			}
		}
	}

	if hasControlPlane {
		roles = append([]string{"control-plane"}, roles...)
	}
	if hasWorker && !hasControlPlane {
		roles = append(roles, "worker")
	}
	if len(roles) == 0 {
		roles = append(roles, "worker")
	}
	return roles
}

// getEndpointPorts 提取 Endpoint 端口（✅ 修复类型转换错误）
func getEndpointPorts(ep corev1.Endpoints) []string {
	if len(ep.Subsets) == 0 {
		return []string{"-"}
	}
	ports := make([]string, 0)
	for _, subset := range ep.Subsets {
		for _, p := range subset.Ports {
			// ✅ 修复：使用格式化而非直接类型转换
			ports = append(ports, strconv.Itoa(int(p.Port)))
		}
	}
	if len(ports) == 0 {
		return []string{"-"}
	}
	return ports
}

// formatEventLastSeen 格式化事件最后可见时间（✅ 修复类型转换错误）
func formatEventLastSeen(e corev1.Event) string {
	// 优先使用 EventTime（新 API）
	if !e.EventTime.IsZero() {
		// ✅ 修复：正确转换 MicroTime → Time
		return model.FormatTime(&metav1.Time{Time: e.EventTime.Time})
	}
	// 回退到 LastTimestamp
	if !e.LastTimestamp.IsZero() {
		return model.FormatTime(&e.LastTimestamp)
	}
	// 最后回退到 CreationTimestamp
	return model.FormatTime(&e.CreationTimestamp)
}

// formatEventObject 格式化事件关联对象
func formatEventObject(obj corev1.ObjectReference) string {
	if obj.Kind == "" && obj.Name == "" {
		return "-"
	}
	if obj.Kind == "" {
		return obj.Name
	}
	if obj.Name == "" {
		return obj.Kind
	}
	return fmt.Sprintf("%s/%s", obj.Kind, obj.Name)
}

// ============================================================================
// 辅助函数 - 性能优化类
// ============================================================================

// buildNodePodCountMap 预构建 nodeName → podCount 映射
func buildNodePodCountMap(pods *corev1.PodList) map[string]int {
	result := make(map[string]int)
	if pods == nil {
		return result
	}
	for _, p := range pods.Items {
		if p.Spec.NodeName != "" {
			result[p.Spec.NodeName]++
		}
	}
	return result
}

// countEndpointAddresses 统计 Endpoint 地址数
func countEndpointAddresses(ep corev1.Endpoints) int {
	count := 0
	for _, subset := range ep.Subsets {
		count += len(subset.Addresses)
	}
	return count
}
