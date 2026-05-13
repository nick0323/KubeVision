package service

import (
	"fmt"
	"math"
	"strings"
	"time"

	appsv1 "k8s.io/api/apps/v1"
	autoscalingv2 "k8s.io/api/autoscaling/v2"
	batchv1 "k8s.io/api/batch/v1"
	corev1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	policyv1 "k8s.io/api/policy/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	storagev1 "k8s.io/api/storage/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/nick0323/K8sVision/model"
)

func MapPods(pods []corev1.Pod) []model.Pod {
	result := make([]model.Pod, len(pods))
	for i, pod := range pods {
		result[i] = model.Pod{
			Namespace:       pod.Namespace,
			Name:            pod.Name,
			Status:          getPodPhaseDisplay(pod.Status.Phase),
			Ready:           CalculatePodReady(pod),
			Restarts:        CalculatePodRestarts(pod),
			Age:             CalculateAge(pod.CreationTimestamp),
			PodIP:           pod.Status.PodIP,
			NodeName:        pod.Spec.NodeName,
			OwnerReferences: pod.OwnerReferences,
		}
	}
	return result
}

func MapDeployments(deployments []appsv1.Deployment) []model.Deployment {
	result := make([]model.Deployment, len(deployments))
	for i, d := range deployments {
		result[i] = model.Deployment{
			Namespace:       d.Namespace,
			Name:            d.Name,
			ReadyReplicas:   d.Status.ReadyReplicas,
			UpdatedReplicas: d.Status.UpdatedReplicas,
			Available:       d.Status.AvailableReplicas,
			Desired:         d.Status.Replicas,
			Status:          GetWorkloadStatus(d.Status.ReadyReplicas, d.Status.Replicas),
			Age:             CalculateAge(d.CreationTimestamp),
		}
	}
	return result
}

func MapStatefulSets(sts []appsv1.StatefulSet) []model.StatefulSet {
	result := make([]model.StatefulSet, len(sts))
	for i, s := range sts {
		result[i] = model.StatefulSet{
			Namespace:       s.Namespace,
			Name:            s.Name,
			ReadyReplicas:   s.Status.ReadyReplicas,
			UpdatedReplicas: s.Status.UpdatedReplicas,
			Available:       s.Status.AvailableReplicas,
			Desired:         s.Status.Replicas,
			Status:          GetWorkloadStatus(s.Status.ReadyReplicas, s.Status.Replicas),
			Age:             CalculateAge(s.CreationTimestamp),
		}
	}
	return result
}

func MapDaemonSets(ds []appsv1.DaemonSet) []model.DaemonSet {
	result := make([]model.DaemonSet, len(ds))
	for i, d := range ds {
		result[i] = model.DaemonSet{
			Namespace:       d.Namespace,
			Name:            d.Name,
			ReadyReplicas:   d.Status.NumberReady,
			UpdatedReplicas: d.Status.UpdatedNumberScheduled,
			Available:       d.Status.NumberAvailable,
			Desired:         d.Status.DesiredNumberScheduled,
			Status:          GetWorkloadStatus(d.Status.NumberReady, d.Status.DesiredNumberScheduled),
			Age:             CalculateAge(d.CreationTimestamp),
		}
	}
	return result
}

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

func MapSecrets(secrets []corev1.Secret) []model.Secret {
	result := make([]model.Secret, len(secrets))
	for i, s := range secrets {
		// 安全考虑：不返回 Keys 字段，防止泄露敏感信息（如 tls.key、password 等）
		result[i] = model.Secret{
			Namespace: s.Namespace,
			Name:      s.Name,
			Type:      string(s.Type),
			DataCount: len(s.Data),
			Keys:      nil, // 不返回密钥键名，防止信息泄露
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
			Path:          getIngressPaths(ing),
			TargetService: getIngressServices(ing),
			Age:           CalculateAge(ing.CreationTimestamp),
		}
	}
	return result
}

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

func MapNodes(nodes []corev1.Node, pods *corev1.PodList, nodeMetricsMap map[string]*model.NodeMetrics) []model.Node {
	podCountMap := buildNodePodCountMap(pods)

	result := make([]model.Node, len(nodes))
	for i, n := range nodes {
		node := model.Node{
			Name:         n.Name,
			IP:           getInternalIP(n),
			Status:       getNodeStatus(n),
			Role:         getNodeRoles(n),
			PodsUsed:     podCountMap[n.Name],
			PodsCapacity: int(n.Status.Capacity.Pods().Value()),
			Age:          CalculateAge(n.CreationTimestamp),
		}

		// 计算CPU和Memory使用率
		if metrics, ok := nodeMetricsMap[n.Name]; ok && metrics != nil {
			if cpuCapacity, ok := n.Status.Capacity[corev1.ResourceCPU]; ok {
				node.CPUUsage = calculateCPUUsage(&cpuCapacity, metrics.CPU)
			}
			if memCapacity, ok := n.Status.Capacity[corev1.ResourceMemory]; ok {
				node.MemoryUsage = calculateMemoryUsage(&memCapacity, metrics.Memory)
			}
		}

		result[i] = node
	}
	return result
}

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

func MapHPAs(hpas []autoscalingv2.HorizontalPodAutoscaler) []model.HorizontalPodAutoscaler {
	result := make([]model.HorizontalPodAutoscaler, len(hpas))
	for i, h := range hpas {
		minReplicas := int32(1)
		if h.Spec.MinReplicas != nil {
			minReplicas = *h.Spec.MinReplicas
		}
		metrics := formatHPAMetrics(h.Spec.Metrics, h.Status.CurrentMetrics)
		result[i] = model.HorizontalPodAutoscaler{
			Namespace:       h.Namespace,
			Name:            h.Name,
			MinReplicas:     minReplicas,
			MaxReplicas:     h.Spec.MaxReplicas,
			CurrentReplicas: h.Status.CurrentReplicas,
			DesiredReplicas: h.Status.DesiredReplicas,
			Metrics:         metrics,
			Age:             CalculateAge(h.CreationTimestamp),
		}
	}
	return result
}

func MapNetworkPolicies(policies []networkingv1.NetworkPolicy) []model.NetworkPolicy {
	result := make([]model.NetworkPolicy, len(policies))
	for i, np := range policies {
		result[i] = model.NetworkPolicy{
			Namespace:   np.Namespace,
			Name:        np.Name,
			PodSelector: formatLabelSelector(&np.Spec.PodSelector),
			PolicyTypes: policyTypesToStrings(np.Spec.PolicyTypes),
			Age:         CalculateAge(np.CreationTimestamp),
		}
	}
	return result
}

func MapServiceAccounts(sas []corev1.ServiceAccount) []model.ServiceAccount {
	result := make([]model.ServiceAccount, len(sas))
	for i, sa := range sas {
		result[i] = model.ServiceAccount{
			Namespace: sa.Namespace,
			Name:      sa.Name,
			Secrets:   len(sa.Secrets),
			Age:       CalculateAge(sa.CreationTimestamp),
		}
	}
	return result
}

func MapRoles(roles []rbacv1.Role) []model.Role {
	result := make([]model.Role, len(roles))
	for i, r := range roles {
		result[i] = model.Role{
			Namespace: r.Namespace,
			Name:      r.Name,
			Rules:     len(r.Rules),
			Age:       CalculateAge(r.CreationTimestamp),
		}
	}
	return result
}

func MapRoleBindings(rbs []rbacv1.RoleBinding) []model.RoleBinding {
	result := make([]model.RoleBinding, len(rbs))
	for i, rb := range rbs {
		result[i] = model.RoleBinding{
			Namespace: rb.Namespace,
			Name:      rb.Name,
			RoleRef:   rb.RoleRef.Name,
			Subjects:  len(rb.Subjects),
			Age:       CalculateAge(rb.CreationTimestamp),
		}
	}
	return result
}

func MapClusterRoles(crs []rbacv1.ClusterRole) []model.ClusterRole {
	result := make([]model.ClusterRole, len(crs))
	for i, cr := range crs {
		result[i] = model.ClusterRole{
			Name:  cr.Name,
			Rules: len(cr.Rules),
			Age:   CalculateAge(cr.CreationTimestamp),
		}
	}
	return result
}

func MapClusterRoleBindings(crbs []rbacv1.ClusterRoleBinding) []model.ClusterRoleBinding {
	result := make([]model.ClusterRoleBinding, len(crbs))
	for i, crb := range crbs {
		result[i] = model.ClusterRoleBinding{
			Name:     crb.Name,
			RoleRef:  crb.RoleRef.Name,
			Subjects: len(crb.Subjects),
			Age:      CalculateAge(crb.CreationTimestamp),
		}
	}
	return result
}

func MapResourceQuotas(rqs []corev1.ResourceQuota) []model.ResourceQuota {
	result := make([]model.ResourceQuota, len(rqs))
	for i, rq := range rqs {
		result[i] = model.ResourceQuota{
			Namespace: rq.Namespace,
			Name:      rq.Name,
			Requests:  formatResourceQuotaHard(rq.Status.Used),
			Limits:    formatResourceQuotaHard(rq.Status.Hard),
			Age:       CalculateAge(rq.CreationTimestamp),
		}
	}
	return result
}

func MapLimitRanges(lrs []corev1.LimitRange) []model.LimitRange {
	result := make([]model.LimitRange, len(lrs))
	for i, lr := range lrs {
		result[i] = model.LimitRange{
			Namespace: lr.Namespace,
			Name:      lr.Name,
			Limits:    formatLimitRangeLimits(lr.Spec.Limits),
			Age:       CalculateAge(lr.CreationTimestamp),
		}
	}
	return result
}

func MapPodDisruptionBudgets(pdbs []policyv1.PodDisruptionBudget) []model.PodDisruptionBudget {
	result := make([]model.PodDisruptionBudget, len(pdbs))
	for i, pdb := range pdbs {
		minAvailable := "-"
		if pdb.Spec.MinAvailable != nil {
			minAvailable = pdb.Spec.MinAvailable.String()
		}
		maxUnavailable := "-"
		if pdb.Spec.MaxUnavailable != nil {
			maxUnavailable = pdb.Spec.MaxUnavailable.String()
		}
		result[i] = model.PodDisruptionBudget{
			Namespace:      pdb.Namespace,
			Name:           pdb.Name,
			MinAvailable:   minAvailable,
			MaxUnavailable: maxUnavailable,
			CurrentHealthy: pdb.Status.CurrentHealthy,
			DesiredHealthy: pdb.Status.DesiredHealthy,
			Age:            CalculateAge(pdb.CreationTimestamp),
		}
	}
	return result
}

func formatHPAMetrics(specMetrics []autoscalingv2.MetricSpec, statusMetrics []autoscalingv2.MetricStatus) string {
	if len(specMetrics) == 0 {
		return "-"
	}
	parts := make([]string, 0, len(specMetrics))
	for _, sm := range specMetrics {
		metricName := string(sm.Type)
		target := "-"
		current := "-"

		switch sm.Type {
		case autoscalingv2.ResourceMetricSourceType:
			if sm.Resource != nil {
				metricName = string(sm.Resource.Name)
				if sm.Resource.Target.AverageUtilization != nil {
					target = fmt.Sprintf("%d%%", *sm.Resource.Target.AverageUtilization)
				} else if sm.Resource.Target.AverageValue != nil {
					target = sm.Resource.Target.AverageValue.String()
				}
				// find matching status
				for _, cm := range statusMetrics {
					if cm.Type == autoscalingv2.ResourceMetricSourceType && cm.Resource != nil && cm.Resource.Name == sm.Resource.Name {
						if cm.Resource.Current.AverageUtilization != nil {
							current = fmt.Sprintf("%d%%", *cm.Resource.Current.AverageUtilization)
						} else if cm.Resource.Current.AverageValue != nil {
							current = cm.Resource.Current.AverageValue.String()
						}
						break
					}
				}
			}
		case autoscalingv2.PodsMetricSourceType:
			if sm.Pods != nil {
				metricName = sm.Pods.Metric.Name
				target = sm.Pods.Target.AverageValue.String()
				for _, cm := range statusMetrics {
					if cm.Type == autoscalingv2.PodsMetricSourceType && cm.Pods != nil && cm.Pods.Metric.Name == sm.Pods.Metric.Name {
						current = cm.Pods.Current.AverageValue.String()
						break
					}
				}
			}
		case autoscalingv2.ObjectMetricSourceType:
			if sm.Object != nil {
				metricName = sm.Object.Metric.Name
				if sm.Object.Target.Value != nil {
					target = sm.Object.Target.Value.String()
				} else if sm.Object.Target.AverageValue != nil {
					target = sm.Object.Target.AverageValue.String()
				}
				for _, cm := range statusMetrics {
					if cm.Type == autoscalingv2.ObjectMetricSourceType && cm.Object != nil && cm.Object.Metric.Name == sm.Object.Metric.Name {
						if cm.Object.Current.Value != nil {
							current = cm.Object.Current.Value.String()
						} else if cm.Object.Current.AverageValue != nil {
							current = cm.Object.Current.AverageValue.String()
						}
						break
					}
				}
			}
		case autoscalingv2.ExternalMetricSourceType:
			if sm.External != nil {
				metricName = sm.External.Metric.Name
				if sm.External.Target.Value != nil {
					target = sm.External.Target.Value.String()
				} else if sm.External.Target.AverageValue != nil {
					target = sm.External.Target.AverageValue.String()
				}
				for _, cm := range statusMetrics {
					if cm.Type == autoscalingv2.ExternalMetricSourceType && cm.External != nil && cm.External.Metric.Name == sm.External.Metric.Name {
						if cm.External.Current.Value != nil {
							current = cm.External.Current.Value.String()
						} else if cm.External.Current.AverageValue != nil {
							current = cm.External.Current.AverageValue.String()
						}
						break
					}
				}
			}
		}

		parts = append(parts, fmt.Sprintf("%s: %s/%s", metricName, current, target))
	}
	return strings.Join(parts, ", ")
}

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

func CalculateAge(t metav1.Time) string {
	return model.FormatAge(t.Time)
}

func CalculatePodReady(pod corev1.Pod) string {
	total := len(pod.Status.ContainerStatuses)
	if total == 0 {
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

func CalculatePodRestarts(pod corev1.Pod) int32 {
	var restarts int32
	for _, cs := range pod.Status.ContainerStatuses {
		restarts += cs.RestartCount
	}
	return restarts
}

func calculateCPUUsage(capacity *resource.Quantity, usageStr string) float64 {
	if capacity == nil || capacity.IsZero() || usageStr == "" {
		return 0
	}

	usage, err := resource.ParseQuantity(usageStr)
	if err != nil {
		return -1
	}

	capMilli := float64(capacity.MilliValue())
	usageMilli := float64(usage.MilliValue())

	if capMilli <= 0 {
		return 0
	}

	return math.Round((usageMilli / capMilli) * 100)
}

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

	return math.Round((float64(usageBytes) / float64(capBytes)) * 100)
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

func getJobStatus(j batchv1.Job) string {
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

func getCronJobStatus(cj batchv1.CronJob) string {
	if cj.Spec.Suspend != nil && *cj.Spec.Suspend {
		return "Suspended"
	}
	return "Active"
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

func getIngressClass(ing networkingv1.Ingress) string {
	if ing.Spec.IngressClassName != nil {
		return *ing.Spec.IngressClassName
	}
	if class, ok := ing.Annotations["kubernetes.io/ingress.class"]; ok {
		return class
	}
	return "-"
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
	if class, ok := pvc.Annotations["volume.beta.kubernetes.io/storage-class"]; ok {
		return class
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
	if pv.Spec.ClaimRef != nil && pv.Spec.ClaimRef.Name != "" {
		ns := pv.Spec.ClaimRef.Namespace
		if ns == "" {
			ns = "default"
		}
		return fmt.Sprintf("%s/%s", ns, pv.Spec.ClaimRef.Name)
	}
	return "-"
}

func getReclaimPolicy(sc storagev1.StorageClass) string {
	if sc.ReclaimPolicy != nil {
		return string(*sc.ReclaimPolicy)
	}
	return "Delete"
}

func getReclaimPolicyStr(pv corev1.PersistentVolume) string {
	if pv.Spec.PersistentVolumeReclaimPolicy != "" {
		return string(pv.Spec.PersistentVolumeReclaimPolicy)
	}
	return "Delete"
}

func getVolumeBindingMode(sc storagev1.StorageClass) string {
	if sc.VolumeBindingMode != nil {
		return string(*sc.VolumeBindingMode)
	}
	return "Immediate"
}

func isDefaultStorageClass(sc storagev1.StorageClass) bool {
	return sc.Annotations["storageclass.kubernetes.io/is-default-class"] == "true"
}

func getInternalIP(node corev1.Node) string {
	for _, addr := range node.Status.Addresses {
		if addr.Type == corev1.NodeInternalIP && addr.Address != "" {
			return addr.Address
		}
	}
	return "-"
}

func getNodeRoles(node corev1.Node) []string {
	roles := make([]string, 0)
	hasControlPlane := false
	hasWorker := false

	for k := range node.Labels {
		if strings.HasPrefix(k, "node-role.kubernetes.io/") {
			role := strings.TrimPrefix(k, "node-role.kubernetes.io/")
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

func getEndpointPorts(ep corev1.Endpoints) []string {
	if len(ep.Subsets) == 0 {
		return []string{"-"}
	}
	ports := make([]string, 0)
	for _, subset := range ep.Subsets {
		for _, p := range subset.Ports {
			ports = append(ports, fmt.Sprintf("%d", p.Port))
		}
	}
	if len(ports) == 0 {
		return []string{"-"}
	}
	return ports
}

func formatEventLastSeen(e corev1.Event) string {
	if !e.EventTime.IsZero() {
		return model.FormatTime(&metav1.Time{Time: e.EventTime.Time})
	}
	if !e.LastTimestamp.IsZero() {
		return model.FormatTime(&e.LastTimestamp)
	}
	return model.FormatTime(&e.CreationTimestamp)
}

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

func countEndpointAddresses(ep corev1.Endpoints) int {
	count := 0
	for _, subset := range ep.Subsets {
		count += len(subset.Addresses)
	}
	return count
}

func formatLabelSelector(selector *metav1.LabelSelector) string {
	if selector == nil {
		return "<none>"
	}
	parts := make([]string, 0, len(selector.MatchLabels))
	for k, v := range selector.MatchLabels {
		parts = append(parts, fmt.Sprintf("%s=%s", k, v))
	}
	for _, expr := range selector.MatchExpressions {
		parts = append(parts, fmt.Sprintf("%s %s %s", expr.Key, expr.Operator, strings.Join(expr.Values, ",")))
	}
	if len(parts) == 0 {
		return "<none>"
	}
	return strings.Join(parts, ", ")
}

func policyTypesToStrings(pts []networkingv1.PolicyType) []string {
	result := make([]string, len(pts))
	for i, pt := range pts {
		result[i] = string(pt)
	}
	return result
}

func formatResourceQuotaHard(hard corev1.ResourceList) string {
	if len(hard) == 0 {
		return "-"
	}
	parts := make([]string, 0, len(hard))
	for k, v := range hard {
		parts = append(parts, fmt.Sprintf("%s=%s", k, v.String()))
	}
	return strings.Join(parts, ", ")
}

func formatLimitRangeLimits(limits []corev1.LimitRangeItem) string {
	if len(limits) == 0 {
		return "-"
	}
	parts := make([]string, 0, len(limits))
	for _, item := range limits {
		itemStr := string(item.Type)
		for k, v := range item.Max {
			itemStr += fmt.Sprintf(" max-%s=%s", k, v.String())
		}
		for k, v := range item.Min {
			itemStr += fmt.Sprintf(" min-%s=%s", k, v.String())
		}
		for k, v := range item.Default {
			itemStr += fmt.Sprintf(" default-%s=%s", k, v.String())
		}
		for k, v := range item.DefaultRequest {
			itemStr += fmt.Sprintf(" defaultRequest-%s=%s", k, v.String())
		}
		parts = append(parts, itemStr)
	}
	return strings.Join(parts, "; ")
}
