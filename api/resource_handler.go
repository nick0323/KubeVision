package api

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/nick0323/K8sVision/api/middleware"
	"github.com/nick0323/K8sVision/model"
	"github.com/nick0323/K8sVision/service"
	"go.uber.org/zap"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	versioned "k8s.io/metrics/pkg/client/clientset/versioned"
)

// isClusterResource 判断是否为集群级资源（不需要 namespace）
func isClusterResource(resourceType string) bool {
	clusterResources := map[string]bool{
		"node":             true,
		"pv":               true,
		"persistentvolume": true,
		"storageclass":     true,
		"namespace":        true,
	}
	return clusterResources[strings.ToLower(resourceType)]
}

// RegisterRoutes 注册通用资源接口
func RegisterRoutes(
	r *gin.RouterGroup,
	logger *zap.Logger,
	getK8sClient K8sClientProvider,
) {
	r.GET("/:resourceType", getResourceList(logger, getK8sClient))
	// 资源详情：/api/:resourceType/:namespace/:name
	// 对于集群资源（node, pv, storageclass, namespace），:namespace 参数会被忽略
	r.GET("/:resourceType/:namespace/:name", getResourceDetail(logger, getK8sClient))
	r.DELETE("/:resourceType/:namespace/:name", deleteResource(logger, getK8sClient))
}

// getResourceList 获取资源列表
func getResourceList(
	logger *zap.Logger,
	getK8sClient K8sClientProvider,
) gin.HandlerFunc {
	return func(c *gin.Context) {
		resourceType := c.Param("resourceType")
		if resourceType == "" {
			middleware.ResponseError(c, logger, fmt.Errorf("resource type is required"), http.StatusBadRequest)
			return
		}

		clientset, metricsClient, err := getK8sClient()
		if err != nil {
			middleware.ResponseError(c, logger, err, http.StatusInternalServerError)
			return
		}

		ctx := GetRequestContext(c)
		namespace := c.Query("namespace")
		labelSelector := c.Query("labelSelector") // 支持 label selector 查询
		if labelSelector == "" {
			labelSelector = c.Query("selector") // 兼容 selector 别名
		}
		fieldSelector := c.Query("fieldSelector")   // 支持 field selector 查询
		involvedObject := c.Query("involvedObject") // Events 专用参数
		since := c.Query("since")                   // Events 专用参数

		// 调用 service 层获取数据
		result, err := getResourceListByType(ctx, clientset, metricsClient, resourceType, namespace, labelSelector, fieldSelector, involvedObject, since)
		if err != nil {
			middleware.ResponseError(c, logger, err, http.StatusInternalServerError)
			return
		}

		// 分页和搜索
		params := ParsePaginationParams(c)
		filteredItems := GenericSearchFilter(result, params.Search)
		if params.SortBy != "" && params.SortOrder != "" {
			filteredItems = SortItems(filteredItems, params.SortBy, params.SortOrder)
		}
		paged := Paginate(filteredItems, params.Offset, params.Limit)

		middleware.ResponseSuccess(c, paged, "获取列表成功", &model.PageMeta{
			Total:  len(filteredItems),
			Limit:  params.Limit,
			Offset: params.Offset,
		})
	}
}

// getResourceDetail 获取资源详情
func getResourceDetail(
	logger *zap.Logger,
	getK8sClient K8sClientProvider,
) gin.HandlerFunc {
	return func(c *gin.Context) {
		resourceType := c.Param("resourceType")
		namespace := c.Param("namespace")
		name := c.Param("name")

		logger.Info("获取资源详情",
			zap.String("resourceType", resourceType),
			zap.String("namespace", namespace),
			zap.String("name", name),
		)

		if resourceType == "" {
			middleware.ResponseError(c, logger, fmt.Errorf("resource type is required"), http.StatusBadRequest)
			return
		}
		if name == "" {
			middleware.ResponseError(c, logger, fmt.Errorf("name is required"), http.StatusBadRequest)
			return
		}

		clientset, _, err := getK8sClient()
		if err != nil {
			middleware.ResponseError(c, logger, err, http.StatusInternalServerError)
			return
		}

		ctx := GetRequestContext(c)

		// 判断是否为集群资源（不需要 namespace）
		ns := namespace
		if isClusterResource(resourceType) {
			ns = ""
		}

		// 直接调用 K8s API 获取原始资源对象
		obj, err := getResourceByName(ctx, clientset, resourceType, ns, name)
		if err != nil {
			logger.Error("获取资源失败", zap.Error(err))
			middleware.ResponseError(c, logger, err, http.StatusNotFound)
			return
		}

		middleware.ResponseSuccess(c, obj, "获取资源详情成功", nil)
	}
}

// deleteResource 删除资源
func deleteResource(
	logger *zap.Logger,
	getK8sClient K8sClientProvider,
) gin.HandlerFunc {
	return func(c *gin.Context) {
		resourceType := c.Param("resourceType")
		namespace := c.Param("namespace")
		name := c.Param("name")

		if resourceType == "" {
			middleware.ResponseError(c, logger, fmt.Errorf("resource type is required"), http.StatusBadRequest)
			return
		}
		if name == "" {
			middleware.ResponseError(c, logger, fmt.Errorf("name is required"), http.StatusBadRequest)
			return
		}

		clientset, _, err := getK8sClient()
		if err != nil {
			middleware.ResponseError(c, logger, err, http.StatusInternalServerError)
			return
		}

		ctx := GetRequestContext(c)

		// 判断是否为集群资源（不需要 namespace）
		ns := namespace
		if isClusterResource(resourceType) {
			ns = ""
		}

		// 根据资源类型调用不同的删除方法
		err = deleteResourceByType(ctx, clientset, resourceType, ns, name)
		if err != nil {
			logger.Error("删除资源失败", zap.Error(err))
			middleware.ResponseError(c, logger, err, http.StatusInternalServerError)
			return
		}

		middleware.ResponseSuccess(c, nil, "资源删除成功", nil)
	}
}

// deleteResourceByType 根据资源类型删除资源
func deleteResourceByType(ctx context.Context, clientset *kubernetes.Clientset, resourceType, namespace, name string) error {
	resourceType = strings.ToLower(resourceType)

	switch resourceType {
	case "pod":
		return clientset.CoreV1().Pods(namespace).Delete(ctx, name, metav1.DeleteOptions{})
	case "deployment":
		return clientset.AppsV1().Deployments(namespace).Delete(ctx, name, metav1.DeleteOptions{})
	case "statefulset":
		return clientset.AppsV1().StatefulSets(namespace).Delete(ctx, name, metav1.DeleteOptions{})
	case "daemonset":
		return clientset.AppsV1().DaemonSets(namespace).Delete(ctx, name, metav1.DeleteOptions{})
	case "service":
		return clientset.CoreV1().Services(namespace).Delete(ctx, name, metav1.DeleteOptions{})
	case "configmap":
		return clientset.CoreV1().ConfigMaps(namespace).Delete(ctx, name, metav1.DeleteOptions{})
	case "secret":
		return clientset.CoreV1().Secrets(namespace).Delete(ctx, name, metav1.DeleteOptions{})
	case "ingress":
		return clientset.NetworkingV1().Ingresses(namespace).Delete(ctx, name, metav1.DeleteOptions{})
	case "job":
		return clientset.BatchV1().Jobs(namespace).Delete(ctx, name, metav1.DeleteOptions{})
	case "cronjob":
		return clientset.BatchV1().CronJobs(namespace).Delete(ctx, name, metav1.DeleteOptions{})
	case "persistentvolumeclaim", "pvc":
		return clientset.CoreV1().PersistentVolumeClaims(namespace).Delete(ctx, name, metav1.DeleteOptions{})
	case "persistentvolume", "pv":
		return clientset.CoreV1().PersistentVolumes().Delete(ctx, name, metav1.DeleteOptions{})
	case "storageclass":
		return clientset.StorageV1().StorageClasses().Delete(ctx, name, metav1.DeleteOptions{})
	case "namespace":
		return clientset.CoreV1().Namespaces().Delete(ctx, name, metav1.DeleteOptions{})
	case "node":
		return clientset.CoreV1().Nodes().Delete(ctx, name, metav1.DeleteOptions{})
	default:
		return fmt.Errorf("不支持的资源类型：%s", resourceType)
	}
}

// getResourceListByType 根据资源类型获取列表
func getResourceListByType(ctx context.Context, clientset *kubernetes.Clientset, metricsClient *versioned.Clientset, resourceType, namespace, labelSelector, fieldSelector, involvedObject, since string) ([]model.SearchableItem, error) {

	switch resourceType {
	case "pod":
		pods, err := service.ListPods(ctx, clientset, nil, namespace, labelSelector, fieldSelector)
		if err != nil {
			return nil, err
		}
		result := make([]model.SearchableItem, len(pods))
		for i := range pods {
			result[i] = &pods[i]
		}
		return result, nil

	case "deployment":
		deployments, err := service.ListDeployments(ctx, clientset, namespace, labelSelector, fieldSelector)
		if err != nil {
			return nil, err
		}
		result := make([]model.SearchableItem, len(deployments))
		for i := range deployments {
			result[i] = &deployments[i]
		}
		return result, nil

	case "statefulset":
		statefulSets, err := service.ListStatefulSets(ctx, clientset, namespace, labelSelector, fieldSelector)
		if err != nil {
			return nil, err
		}
		result := make([]model.SearchableItem, len(statefulSets))
		for i := range statefulSets {
			result[i] = &statefulSets[i]
		}
		return result, nil

	case "daemonset":
		daemonSets, err := service.ListDaemonSets(ctx, clientset, namespace, labelSelector, fieldSelector)
		if err != nil {
			return nil, err
		}
		result := make([]model.SearchableItem, len(daemonSets))
		for i := range daemonSets {
			result[i] = &daemonSets[i]
		}
		return result, nil

	case "service":
		services, err := service.ListServices(ctx, clientset, namespace, labelSelector, fieldSelector)
		if err != nil {
			return nil, err
		}
		result := make([]model.SearchableItem, len(services))
		for i := range services {
			result[i] = &services[i]
		}
		return result, nil

	case "configmap":
		configMaps, err := service.ListConfigMaps(ctx, clientset, namespace, labelSelector, fieldSelector)
		if err != nil {
			return nil, err
		}
		result := make([]model.SearchableItem, len(configMaps))
		for i := range configMaps {
			result[i] = &configMaps[i]
		}
		return result, nil

	case "secret":
		secrets, err := service.ListSecrets(ctx, clientset, namespace, labelSelector, fieldSelector)
		if err != nil {
			return nil, err
		}
		result := make([]model.SearchableItem, len(secrets))
		for i := range secrets {
			result[i] = &secrets[i]
		}
		return result, nil

	case "ingress":
		ingresses, err := service.ListIngresses(ctx, clientset, namespace, labelSelector, fieldSelector)
		if err != nil {
			return nil, err
		}
		result := make([]model.SearchableItem, len(ingresses))
		for i := range ingresses {
			result[i] = &ingresses[i]
		}
		return result, nil

	case "job":
		jobs, err := service.ListJobs(ctx, clientset, namespace, labelSelector, fieldSelector)
		if err != nil {
			return nil, err
		}
		result := make([]model.SearchableItem, len(jobs))
		for i := range jobs {
			result[i] = &jobs[i]
		}
		return result, nil

	case "cronjob":
		cronJobs, err := service.ListCronJobs(ctx, clientset, namespace, labelSelector, fieldSelector)
		if err != nil {
			return nil, err
		}
		result := make([]model.SearchableItem, len(cronJobs))
		for i := range cronJobs {
			result[i] = &cronJobs[i]
		}
		return result, nil

	case "persistentvolumeclaim", "pvc":
		pvcs, err := service.ListPVCs(ctx, clientset, namespace, labelSelector, fieldSelector)
		if err != nil {
			return nil, err
		}
		result := make([]model.SearchableItem, len(pvcs))
		for i := range pvcs {
			result[i] = &pvcs[i]
		}
		return result, nil

	case "persistentvolume", "pv":
		pvs, err := service.ListPVs(ctx, clientset, labelSelector, fieldSelector)
		if err != nil {
			return nil, err
		}
		result := make([]model.SearchableItem, len(pvs))
		for i := range pvs {
			result[i] = &pvs[i]
		}
		return result, nil

	case "storageclass":
		storageClasses, err := service.ListStorageClasses(ctx, clientset, labelSelector, fieldSelector)
		if err != nil {
			return nil, err
		}
		result := make([]model.SearchableItem, len(storageClasses))
		for i := range storageClasses {
			result[i] = &storageClasses[i]
		}
		return result, nil

	case "namespace":
		namespaces, err := service.ListNamespaces(ctx, clientset, labelSelector, fieldSelector)
		if err != nil {
			return nil, err
		}
		result := make([]model.SearchableItem, len(namespaces))
		for i := range namespaces {
			result[i] = &namespaces[i]
		}
		return result, nil

	case "node":
		// 获取节点 metrics
		var nodeMetricsMap map[string]model.NodeMetrics
		if metricsClient != nil {
			nodeMetricsMap, _ = GetNodeMetrics(ctx, metricsClient)
		}

		nodes, err := service.ListNodes(ctx, clientset, nil, nodeMetricsMap, labelSelector, fieldSelector)
		if err != nil {
			return nil, err
		}
		result := make([]model.SearchableItem, len(nodes))
		for i := range nodes {
			result[i] = &nodes[i]
		}
		return result, nil

	case "endpoint":
		endpoints, err := service.ListEndpoints(ctx, clientset, namespace, labelSelector, fieldSelector)
		if err != nil {
			return nil, err
		}
		result := make([]model.SearchableItem, len(endpoints))
		for i := range endpoints {
			result[i] = &endpoints[i]
		}
		return result, nil

	case "event":
		events, err := service.ListEvents(ctx, clientset, namespace, involvedObject, since, labelSelector, fieldSelector)
		if err != nil {
			return nil, err
		}
		result := make([]model.SearchableItem, len(events))
		for i := range events {
			result[i] = &events[i]
		}
		return result, nil

	default:
		return nil, fmt.Errorf("unsupported resource type: %s", resourceType)
	}
}

// getResourceByName 根据资源类型和名称获取对象（返回 K8s 原始对象）
// 只支持单数形式，如：/api/deployment/ns/name、/api/ingress/ns/name
func getResourceByName(ctx context.Context, clientset *kubernetes.Clientset, resourceType, namespace, name string) (interface{}, error) {
	// 规范化资源类型（转为小写）
	resourceType = strings.ToLower(resourceType)

	switch resourceType {
	case "pod":
		return clientset.CoreV1().Pods(namespace).Get(ctx, name, metav1.GetOptions{})
	case "deployment":
		return clientset.AppsV1().Deployments(namespace).Get(ctx, name, metav1.GetOptions{})
	case "statefulset":
		return clientset.AppsV1().StatefulSets(namespace).Get(ctx, name, metav1.GetOptions{})
	case "daemonset":
		return clientset.AppsV1().DaemonSets(namespace).Get(ctx, name, metav1.GetOptions{})
	case "service":
		return clientset.CoreV1().Services(namespace).Get(ctx, name, metav1.GetOptions{})
	case "configmap":
		return clientset.CoreV1().ConfigMaps(namespace).Get(ctx, name, metav1.GetOptions{})
	case "secret":
		return clientset.CoreV1().Secrets(namespace).Get(ctx, name, metav1.GetOptions{})
	case "ingress":
		return clientset.NetworkingV1().Ingresses(namespace).Get(ctx, name, metav1.GetOptions{})
	case "job":
		return clientset.BatchV1().Jobs(namespace).Get(ctx, name, metav1.GetOptions{})
	case "cronjob":
		return clientset.BatchV1().CronJobs(namespace).Get(ctx, name, metav1.GetOptions{})
	case "persistentvolumeclaim", "pvc":
		return clientset.CoreV1().PersistentVolumeClaims(namespace).Get(ctx, name, metav1.GetOptions{})
	case "persistentvolume", "pv":
		return clientset.CoreV1().PersistentVolumes().Get(ctx, name, metav1.GetOptions{})
	case "storageclass":
		return clientset.StorageV1().StorageClasses().Get(ctx, name, metav1.GetOptions{})
	case "namespace":
		return clientset.CoreV1().Namespaces().Get(ctx, name, metav1.GetOptions{})
	case "node":
		return clientset.CoreV1().Nodes().Get(ctx, name, metav1.GetOptions{})
	case "endpoint":
		return clientset.CoreV1().Endpoints(namespace).Get(ctx, name, metav1.GetOptions{})
	default:
		return nil, fmt.Errorf("不支持的资源类型：%s", resourceType)
	}
}
