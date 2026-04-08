package service

import (
	"context"
	"sort"
	"strings"
	"time"

	"github.com/nick0323/K8sVision/model"
	corev1 "k8s.io/api/core/v1"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// ListOptions K8s 列表选项
type ListOptions struct {
	Namespace     string
	LabelSelector string
	FieldSelector string
	Limit         int64
}

// Apply 应用列表选项到 metav1.ListOptions
func (o *ListOptions) Apply() metav1.ListOptions {
	opts := metav1.ListOptions{}
	if o.LabelSelector != "" {
		opts.LabelSelector = o.LabelSelector
	}
	if o.FieldSelector != "" {
		opts.FieldSelector = o.FieldSelector
	}
	if o.Limit > 0 {
		opts.Limit = o.Limit
	}
	return opts
}

// DefaultListOptions 默认列表选项
func DefaultListOptions() *ListOptions {
	return &ListOptions{
		Limit: 1000, // 默认分页大小
	}
}

// ==================== Pod 列表 ====================

// ListPods 获取 Pod 列表
func ListPods(ctx context.Context, clientset *kubernetes.Clientset, podMetricsMap map[string]model.PodMetrics, namespace string) ([]model.Pod, error) {
	pods, _, err := ListPodsWithRaw(ctx, clientset, podMetricsMap, namespace)
	return pods, err
}

// ListPodsWithRaw 获取 Pod 列表（包含原始 Pod 列表和指标数据）
// noLimit: 如果为 true，则不限制返回数量（用于集群概览等需要完整统计的场景）
func ListPodsWithRaw(ctx context.Context, clientset *kubernetes.Clientset, podMetricsMap map[string]model.PodMetrics, namespace string, noLimit ...bool) ([]model.Pod, *v1.PodList, error) {
	opts := DefaultListOptions()
	opts.Namespace = namespace

	// 如果指定了 noLimit 参数且为 true，则移除数量限制
	if len(noLimit) > 0 && noLimit[0] {
		opts.Limit = 0
	}

	pods, err := ListResourcesWithNamespace(ctx, namespace,
		func() (*v1.PodList, error) {
			return clientset.CoreV1().Pods("").List(ctx, opts.Apply())
		},
		func(ns string) (*v1.PodList, error) {
			return clientset.CoreV1().Pods(ns).List(ctx, opts.Apply())
		},
	)
	if err != nil {
		return nil, nil, err
	}

	// 使用通用映射函数
	podStatuses := MapPods(pods.Items, podMetricsMap)
	return podStatuses, pods, nil
}

// ==================== 工作负载列表 ====================

// ListDeployments 获取 Deployment 列表
func ListDeployments(ctx context.Context, clientset *kubernetes.Clientset, namespace string) ([]model.Deployment, error) {
	opts := DefaultListOptions()
	depList, err := clientset.AppsV1().Deployments(namespace).List(ctx, opts.Apply())
	if err != nil {
		return nil, err
	}

	// 使用通用映射函数
	result := MapDeployments(depList.Items)
	return result, nil
}

// ListStatefulSets 获取 StatefulSet 列表
func ListStatefulSets(ctx context.Context, clientset *kubernetes.Clientset, namespace string) ([]model.StatefulSet, error) {
	opts := DefaultListOptions()
	stsList, err := clientset.AppsV1().StatefulSets(namespace).List(ctx, opts.Apply())
	if err != nil {
		return nil, err
	}

	// 使用通用映射函数
	result := MapStatefulSets(stsList.Items)
	return result, nil
}

// ListDaemonSets 获取 DaemonSet 列表
func ListDaemonSets(ctx context.Context, clientset *kubernetes.Clientset, namespace string) ([]model.DaemonSet, error) {
	opts := DefaultListOptions()
	dsList, err := clientset.AppsV1().DaemonSets(namespace).List(ctx, opts.Apply())
	if err != nil {
		return nil, err
	}

	// 使用通用映射函数
	result := MapDaemonSets(dsList.Items)
	return result, nil
}

// ListJobs 获取 Job 列表
func ListJobs(ctx context.Context, clientset *kubernetes.Clientset, namespace string) ([]model.Job, error) {
	jobs, err := clientset.BatchV1().Jobs(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	// 使用通用映射函数
	result := MapJobs(jobs.Items)
	return result, nil
}

// ListCronJobs 获取 CronJob 列表
func ListCronJobs(ctx context.Context, clientset *kubernetes.Clientset, namespace string) ([]model.CronJob, error) {
	cronjobs, err := clientset.BatchV1().CronJobs(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	// 使用通用映射函数
	result := MapCronJobs(cronjobs.Items)
	return result, nil
}

// ==================== 服务和网络列表 ====================

// ListServices 获取 Service 列表
func ListServices(ctx context.Context, clientset *kubernetes.Clientset, namespace string) ([]model.Service, error) {
	opts := DefaultListOptions()
	svcs, err := clientset.CoreV1().Services(namespace).List(ctx, opts.Apply())
	if err != nil {
		return nil, err
	}

	// 使用通用映射函数
	result := MapServices(svcs.Items)
	return result, nil
}

// ListIngresses 获取 Ingress 列表
func ListIngresses(ctx context.Context, clientset *kubernetes.Clientset, namespace string) ([]model.Ingress, error) {
	opts := DefaultListOptions()
	ingresses, err := clientset.NetworkingV1().Ingresses(namespace).List(ctx, opts.Apply())
	if err != nil {
		return nil, err
	}

	// 使用通用映射函数
	result := MapIngresses(ingresses.Items)
	return result, nil
}

// ==================== 存储列表 ====================

// ListPVCs 获取 PVC 列表
func ListPVCs(ctx context.Context, clientset *kubernetes.Clientset, namespace string) ([]model.PVC, error) {
	opts := DefaultListOptions()
	pvcList, err := ListResourcesWithNamespace(ctx, namespace,
		func() (*v1.PersistentVolumeClaimList, error) {
			return clientset.CoreV1().PersistentVolumeClaims("").List(ctx, opts.Apply())
		},
		func(ns string) (*v1.PersistentVolumeClaimList, error) {
			return clientset.CoreV1().PersistentVolumeClaims(ns).List(ctx, opts.Apply())
		},
	)
	if err != nil {
		return nil, err
	}

	// 使用通用映射函数
	result := MapPVCs(pvcList.Items)
	return result, nil
}

// ListPVs 获取 PV 列表
func ListPVs(ctx context.Context, clientset *kubernetes.Clientset) ([]model.PV, error) {
	opts := DefaultListOptions()
	pvList, err := clientset.CoreV1().PersistentVolumes().List(ctx, opts.Apply())
	if err != nil {
		return nil, err
	}

	// 使用通用映射函数
	result := MapPVs(pvList.Items)
	return result, nil
}

// ListStorageClasses 获取 StorageClass 列表
func ListStorageClasses(ctx context.Context, clientset *kubernetes.Clientset) ([]model.StorageClass, error) {
	opts := DefaultListOptions()
	scList, err := clientset.StorageV1().StorageClasses().List(ctx, opts.Apply())
	if err != nil {
		return nil, err
	}

	// 使用通用映射函数
	result := MapStorageClasses(scList.Items)
	return result, nil
}

// ==================== 配置列表 ====================

// ListConfigMaps 获取 ConfigMap 列表
func ListConfigMaps(ctx context.Context, clientset *kubernetes.Clientset, namespace string) ([]model.ConfigMap, error) {
	opts := DefaultListOptions()
	cmList, err := ListResourcesWithNamespace(ctx, namespace,
		func() (*v1.ConfigMapList, error) {
			return clientset.CoreV1().ConfigMaps("").List(ctx, opts.Apply())
		},
		func(ns string) (*v1.ConfigMapList, error) {
			return clientset.CoreV1().ConfigMaps(ns).List(ctx, opts.Apply())
		},
	)
	if err != nil {
		return nil, err
	}

	// 使用通用映射函数
	result := MapConfigMaps(cmList.Items)
	return result, nil
}

// ListSecrets 获取 Secret 列表
func ListSecrets(ctx context.Context, clientset *kubernetes.Clientset, namespace string) ([]model.Secret, error) {
	opts := DefaultListOptions()
	// 优化：指定命名空间时只获取该命名空间的 Secrets
	var secretList *v1.SecretList
	var err error

	if namespace != "" {
		secretList, err = clientset.CoreV1().Secrets(namespace).List(ctx, opts.Apply())
	} else {
		// 全量获取（所有命名空间）- 性能开销较大
		secretList, err = clientset.CoreV1().Secrets("").List(ctx, opts.Apply())
	}

	if err != nil {
		return nil, err
	}

	// 使用通用映射函数
	result := MapSecrets(secretList.Items)
	return result, nil
}

// ==================== 集群资源列表 ====================

// ListNodes 获取 Node 列表
func ListNodes(ctx context.Context, clientset *kubernetes.Clientset, pods *v1.PodList, nodeMetricsMap map[string]model.NodeMetrics) ([]model.Node, error) {
	opts := DefaultListOptions()
	nodes, err := clientset.CoreV1().Nodes().List(ctx, opts.Apply())
	if err != nil {
		return nil, err
	}

	// 如果未提供 Pods 列表，尝试获取（用于计算 PodsUsed）
	if pods == nil {
		pods, err = clientset.CoreV1().Pods("").List(ctx, metav1.ListOptions{})
		if err != nil {
			// 获取失败不影响节点列表返回
			pods = nil
		}
	}

	// 使用通用映射函数
	result := MapNodes(nodes.Items, pods, nodeMetricsMap)
	return result, nil
}

// ListNamespaces 获取 Namespace 列表
func ListNamespaces(ctx context.Context, clientset *kubernetes.Clientset) ([]model.Namespace, error) {
	opts := DefaultListOptions()
	nsList, err := clientset.CoreV1().Namespaces().List(ctx, opts.Apply())
	if err != nil {
		return nil, err
	}

	// 使用通用映射函数
	result := MapNamespaces(nsList.Items)
	return result, nil
}

// ListEvents 获取 Event 列表
// 参数：
//   - namespace: 命名空间
//   - involvedObject: 关联对象（格式：Kind/Name，如：Deployment/my-app）
//   - since: 时间范围（如：1h, 30m, 1d），默认 1 小时
func ListEvents(ctx context.Context, clientset *kubernetes.Clientset, namespace string, involvedObject string, since string) ([]model.Event, error) {
	opts := DefaultListOptions()

	// 从 K8s API 获取事件
	events, err := clientset.CoreV1().Events(namespace).List(ctx, opts.Apply())
	if err != nil {
		return nil, err
	}

	// 解析时间范围
	duration, err := time.ParseDuration(since)
	if err != nil {
		// 默认 1 小时
		duration = 1 * time.Hour
	}

	// 在内存中过滤指定时间范围内的事件
	cutoffTime := time.Now().Add(-duration)
	filteredEvents := make([]corev1.Event, 0, len(events.Items))
	for _, event := range events.Items {
		eventTime := event.LastTimestamp.Time
		if eventTime.IsZero() {
			eventTime = event.EventTime.Time
		}
		if !eventTime.IsZero() && eventTime.After(cutoffTime) {
			// 如果指定了 involvedObject，进一步过滤
			if involvedObject != "" {
				// 解析 involvedObject（格式：Kind/Name）
				parts := strings.SplitN(involvedObject, "/", 2)
				if len(parts) == 2 {
					kind := parts[0]
					name := parts[1]
					// 检查事件的 involvedObject 是否匹配
					if event.InvolvedObject.Kind != kind || event.InvolvedObject.Name != name {
						continue
					}
				}
			}
			filteredEvents = append(filteredEvents, event)
		}
	}

	// 按 lastTimestamp 倒序排序（最新的事件在前）
	sort.Slice(filteredEvents, func(i, j int) bool {
		iTime := filteredEvents[i].LastTimestamp.Time
		if iTime.IsZero() {
			iTime = filteredEvents[i].EventTime.Time
		}
		jTime := filteredEvents[j].LastTimestamp.Time
		if jTime.IsZero() {
			jTime = filteredEvents[j].EventTime.Time
		}
		return iTime.After(jTime)
	})

	// 使用通用映射函数
	result := MapEvents(filteredEvents)
	return result, nil
}

// ListEndpoints 获取 Endpoints 列表
func ListEndpoints(ctx context.Context, clientset *kubernetes.Clientset, namespace string) ([]model.Endpoints, error) {
	opts := DefaultListOptions()
	endpoints, err := clientset.CoreV1().Endpoints(namespace).List(ctx, opts.Apply())
	if err != nil {
		return nil, err
	}

	// 使用通用映射函数
	result := MapEndpoints(endpoints.Items)
	return result, nil
}
