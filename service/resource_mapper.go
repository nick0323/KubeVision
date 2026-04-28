package service

import (
	"fmt"
	"math"
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

func MapPods(pods []corev1.Pod) []model.Pod {
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
