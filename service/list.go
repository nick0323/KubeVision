package service

import (
	"context"

	"github.com/nick0323/K8sVision/model"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// ==================== Pod 列表 ====================

// ListPodsWithRaw 获取 Pod 列表（包含原始 Pod 列表和指标数据）
func ListPodsWithRaw(ctx context.Context, clientset *kubernetes.Clientset, podMetricsMap model.PodMetricsMap, namespace string) ([]model.PodStatus, *v1.PodList, error) {
	pods, err := ListResourcesWithNamespace(ctx, namespace,
		func() (*v1.PodList, error) {
			return clientset.CoreV1().Pods("").List(ctx, metav1.ListOptions{})
		},
		func(ns string) (*v1.PodList, error) {
			return clientset.CoreV1().Pods(ns).List(ctx, metav1.ListOptions{})
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
func ListDeployments(ctx context.Context, clientset *kubernetes.Clientset, namespace string) ([]model.DeploymentStatus, error) {
	depList, err := clientset.AppsV1().Deployments(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	// 使用通用映射函数
	result := MapDeployments(depList.Items)
	return result, nil
}

// ListStatefulSets 获取 StatefulSet 列表
func ListStatefulSets(ctx context.Context, clientset *kubernetes.Clientset, namespace string) ([]model.StatefulSetStatus, error) {
	stsList, err := clientset.AppsV1().StatefulSets(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	// 使用通用映射函数
	result := MapStatefulSets(stsList.Items)
	return result, nil
}

// ListDaemonSets 获取 DaemonSet 列表
func ListDaemonSets(ctx context.Context, clientset *kubernetes.Clientset, namespace string) ([]model.DaemonSetStatus, error) {
	dsList, err := clientset.AppsV1().DaemonSets(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	// 使用通用映射函数
	result := MapDaemonSets(dsList.Items)
	return result, nil
}

// ListJobs 获取 Job 列表
func ListJobs(ctx context.Context, clientset *kubernetes.Clientset, namespace string) ([]model.JobStatus, error) {
	jobs, err := clientset.BatchV1().Jobs(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	// 使用通用映射函数
	result := MapJobs(jobs.Items)
	return result, nil
}

// ListCronJobs 获取 CronJob 列表
func ListCronJobs(ctx context.Context, clientset *kubernetes.Clientset, namespace string) ([]model.CronJobStatus, error) {
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
func ListServices(ctx context.Context, clientset *kubernetes.Clientset, namespace string) ([]model.ServiceStatus, error) {
	svcs, err := clientset.CoreV1().Services(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	// 使用通用映射函数
	result := MapServices(svcs.Items)
	return result, nil
}

// ListIngresses 获取 Ingress 列表
func ListIngresses(ctx context.Context, clientset *kubernetes.Clientset, namespace string) ([]model.IngressStatus, error) {
	ingresses, err := clientset.NetworkingV1().Ingresses(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	// 使用通用映射函数
	result := MapIngresses(ingresses.Items)
	return result, nil
}

// ==================== 存储列表 ====================

// ListPVCs 获取 PVC 列表
func ListPVCs(ctx context.Context, clientset *kubernetes.Clientset, namespace string) ([]model.PVCStatus, error) {
	pvcList, err := ListResourcesWithNamespace(ctx, namespace,
		func() (*v1.PersistentVolumeClaimList, error) {
			return clientset.CoreV1().PersistentVolumeClaims("").List(ctx, metav1.ListOptions{})
		},
		func(ns string) (*v1.PersistentVolumeClaimList, error) {
			return clientset.CoreV1().PersistentVolumeClaims(ns).List(ctx, metav1.ListOptions{})
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
func ListPVs(ctx context.Context, clientset *kubernetes.Clientset) ([]model.PVStatus, error) {
	pvList, err := clientset.CoreV1().PersistentVolumes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	// 使用通用映射函数
	result := MapPVs(pvList.Items)
	return result, nil
}

// ListStorageClasses 获取 StorageClass 列表
func ListStorageClasses(ctx context.Context, clientset *kubernetes.Clientset) ([]model.StorageClassStatus, error) {
	scList, err := clientset.StorageV1().StorageClasses().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	// 使用通用映射函数
	result := MapStorageClasses(scList.Items)
	return result, nil
}

// ==================== 配置列表 ====================

// ListConfigMaps 获取 ConfigMap 列表
func ListConfigMaps(ctx context.Context, clientset *kubernetes.Clientset, namespace string) ([]model.ConfigMapStatus, error) {
	cmList, err := ListResourcesWithNamespace(ctx, namespace,
		func() (*v1.ConfigMapList, error) {
			return clientset.CoreV1().ConfigMaps("").List(ctx, metav1.ListOptions{})
		},
		func(ns string) (*v1.ConfigMapList, error) {
			return clientset.CoreV1().ConfigMaps(ns).List(ctx, metav1.ListOptions{})
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
func ListSecrets(ctx context.Context, clientset *kubernetes.Clientset, namespace string) ([]model.SecretStatus, error) {
	secretList, err := ListResourcesWithNamespace(ctx, namespace,
		func() (*v1.SecretList, error) {
			return clientset.CoreV1().Secrets("").List(ctx, metav1.ListOptions{})
		},
		func(ns string) (*v1.SecretList, error) {
			return clientset.CoreV1().Secrets(ns).List(ctx, metav1.ListOptions{})
		},
	)
	if err != nil {
		return nil, err
	}

	// 使用通用映射函数
	result := MapSecrets(secretList.Items)
	return result, nil
}

// ==================== 集群资源列表 ====================

// ListNodes 获取 Node 列表
func ListNodes(ctx context.Context, clientset *kubernetes.Clientset, pods *v1.PodList, nodeMetricsMap model.NodeMetricsMap) ([]model.NodeStatus, error) {
	nodes, err := clientset.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	// 使用通用映射函数
	result := MapNodes(nodes.Items, pods, nodeMetricsMap)
	return result, nil
}

// ListNamespaces 获取 Namespace 列表
func ListNamespaces(ctx context.Context, clientset *kubernetes.Clientset) ([]model.NamespaceDetail, error) {
	nsList, err := clientset.CoreV1().Namespaces().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	// 使用通用映射函数
	result := MapNamespaces(nsList.Items)
	return result, nil
}

// ListEvents 获取 Event 列表
func ListEvents(ctx context.Context, clientset *kubernetes.Clientset, namespace string) ([]model.EventStatus, error) {
	events, err := clientset.CoreV1().Events(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	// 使用通用映射函数
	result := MapEvents(events.Items)
	return result, nil
}
