package api

import (
	"context"
	"fmt"
	"strings"

	"github.com/nick0323/K8sVision/model"
	"github.com/nick0323/K8sVision/service"
	"go.uber.org/zap"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// ResourceHandler 资源处理器接口
type ResourceHandler interface {
	// List 获取资源列表
	List(ctx context.Context, clientset *kubernetes.Clientset, namespace, labelSelector, fieldSelector string) ([]model.SearchableItem, error)
	// Get 获取资源详情
	Get(ctx context.Context, clientset *kubernetes.Clientset, namespace, name string) (interface{}, error)
	// Delete 删除资源
	Delete(ctx context.Context, clientset *kubernetes.Clientset, namespace, name string) error
	// IsClusterScoped 是否为集群级资源
	IsClusterScoped() bool
}

// ResourceRegistry 资源注册表
type ResourceRegistry struct {
	handlers map[string]ResourceHandler
	logger   *zap.Logger
}

// NewResourceRegistry 创建资源注册表
func NewResourceRegistry(logger *zap.Logger) *ResourceRegistry {
	registry := &ResourceRegistry{
		handlers: make(map[string]ResourceHandler),
		logger:   logger,
	}
	registry.registerAllResources()
	return registry
}

// GetHandler 获取资源处理器
func (r *ResourceRegistry) GetHandler(resourceType string) (ResourceHandler, bool) {
	handler, exists := r.handlers[strings.ToLower(resourceType)]
	return handler, exists
}

// GetSupportedResourceTypes 获取所有支持的资源类型
func (r *ResourceRegistry) GetSupportedResourceTypes() []string {
	types := make([]string, 0, len(r.handlers))
	for t := range r.handlers {
		types = append(types, t)
	}
	return types
}

// registerAllResources 注册所有资源处理器
func (r *ResourceRegistry) registerAllResources() {
	// Namespaced resources
	r.register("pod", newPodHandler())
	r.register("deployment", newDeploymentHandler())
	r.register("statefulset", newStatefulSetHandler())
	r.register("daemonset", newDaemonSetHandler())
	r.register("service", newServiceHandler())
	r.register("configmap", newConfigMapHandler())
	r.register("secret", newSecretHandler())
	r.register("ingress", newIngressHandler())
	r.register("job", newJobHandler())
	r.register("cronjob", newCronJobHandler())
	r.register("persistentvolumeclaim", newPVCHandler())
	r.register("pvc", newPVCHandler())
	r.register("endpoint", newEndpointsHandler())
	r.register("event", newEventHandler())

	// Cluster-scoped resources
	r.register("persistentvolume", newPVHandler())
	r.register("pv", newPVHandler())
	r.register("storageclass", newStorageClassHandler())
	r.register("namespace", newNamespaceHandler())
	r.register("node", newNodeHandler())
}

// register 注册单个资源处理器
func (r *ResourceRegistry) register(name string, handler ResourceHandler) {
	r.handlers[strings.ToLower(name)] = handler
	r.logger.Debug("Resource handler registered", zap.String("resource", name))
}

// ============== 具体资源处理器实现 ==============

// PodHandler Pod 资源处理器
type PodHandler struct {
	clusterScoped bool
}

func newPodHandler() *PodHandler {
	return &PodHandler{clusterScoped: false}
}

func (h *PodHandler) List(ctx context.Context, clientset *kubernetes.Clientset, namespace, labelSelector, fieldSelector string) ([]model.SearchableItem, error) {
	pods, err := service.ListPods(ctx, clientset, nil, namespace, labelSelector, fieldSelector)
	if err != nil {
		return nil, err
	}
	result := make([]model.SearchableItem, len(pods))
	for i := range pods {
		result[i] = &pods[i]
	}
	return result, nil
}

func (h *PodHandler) Get(ctx context.Context, clientset *kubernetes.Clientset, namespace, name string) (interface{}, error) {
	return clientset.CoreV1().Pods(namespace).Get(ctx, name, metav1.GetOptions{})
}

func (h *PodHandler) Delete(ctx context.Context, clientset *kubernetes.Clientset, namespace, name string) error {
	return clientset.CoreV1().Pods(namespace).Delete(ctx, name, metav1.DeleteOptions{})
}

func (h *PodHandler) IsClusterScoped() bool {
	return h.clusterScoped
}

// DeploymentHandler Deployment 资源处理器
type DeploymentHandler struct {
	clusterScoped bool
}

func newDeploymentHandler() *DeploymentHandler {
	return &DeploymentHandler{clusterScoped: false}
}

func (h *DeploymentHandler) List(ctx context.Context, clientset *kubernetes.Clientset, namespace, labelSelector, fieldSelector string) ([]model.SearchableItem, error) {
	deployments, err := service.ListDeployments(ctx, clientset, namespace, labelSelector, fieldSelector)
	if err != nil {
		return nil, err
	}
	result := make([]model.SearchableItem, len(deployments))
	for i := range deployments {
		result[i] = &deployments[i]
	}
	return result, nil
}

func (h *DeploymentHandler) Get(ctx context.Context, clientset *kubernetes.Clientset, namespace, name string) (interface{}, error) {
	return clientset.AppsV1().Deployments(namespace).Get(ctx, name, metav1.GetOptions{})
}

func (h *DeploymentHandler) Delete(ctx context.Context, clientset *kubernetes.Clientset, namespace, name string) error {
	return clientset.AppsV1().Deployments(namespace).Delete(ctx, name, metav1.DeleteOptions{})
}

func (h *DeploymentHandler) IsClusterScoped() bool {
	return h.clusterScoped
}

// StatefulSetHandler StatefulSet 资源处理器
type StatefulSetHandler struct {
	clusterScoped bool
}

func newStatefulSetHandler() *StatefulSetHandler {
	return &StatefulSetHandler{clusterScoped: false}
}

func (h *StatefulSetHandler) List(ctx context.Context, clientset *kubernetes.Clientset, namespace, labelSelector, fieldSelector string) ([]model.SearchableItem, error) {
	statefulSets, err := service.ListStatefulSets(ctx, clientset, namespace, labelSelector, fieldSelector)
	if err != nil {
		return nil, err
	}
	result := make([]model.SearchableItem, len(statefulSets))
	for i := range statefulSets {
		result[i] = &statefulSets[i]
	}
	return result, nil
}

func (h *StatefulSetHandler) Get(ctx context.Context, clientset *kubernetes.Clientset, namespace, name string) (interface{}, error) {
	return clientset.AppsV1().StatefulSets(namespace).Get(ctx, name, metav1.GetOptions{})
}

func (h *StatefulSetHandler) Delete(ctx context.Context, clientset *kubernetes.Clientset, namespace, name string) error {
	return clientset.AppsV1().StatefulSets(namespace).Delete(ctx, name, metav1.DeleteOptions{})
}

func (h *StatefulSetHandler) IsClusterScoped() bool {
	return h.clusterScoped
}

// DaemonSetHandler DaemonSet 资源处理器
type DaemonSetHandler struct {
	clusterScoped bool
}

func newDaemonSetHandler() *DaemonSetHandler {
	return &DaemonSetHandler{clusterScoped: false}
}

func (h *DaemonSetHandler) List(ctx context.Context, clientset *kubernetes.Clientset, namespace, labelSelector, fieldSelector string) ([]model.SearchableItem, error) {
	daemonSets, err := service.ListDaemonSets(ctx, clientset, namespace, labelSelector, fieldSelector)
	if err != nil {
		return nil, err
	}
	result := make([]model.SearchableItem, len(daemonSets))
	for i := range daemonSets {
		result[i] = &daemonSets[i]
	}
	return result, nil
}

func (h *DaemonSetHandler) Get(ctx context.Context, clientset *kubernetes.Clientset, namespace, name string) (interface{}, error) {
	return clientset.AppsV1().DaemonSets(namespace).Get(ctx, name, metav1.GetOptions{})
}

func (h *DaemonSetHandler) Delete(ctx context.Context, clientset *kubernetes.Clientset, namespace, name string) error {
	return clientset.AppsV1().DaemonSets(namespace).Delete(ctx, name, metav1.DeleteOptions{})
}

func (h *DaemonSetHandler) IsClusterScoped() bool {
	return h.clusterScoped
}

// ServiceHandler Service 资源处理器
type ServiceHandler struct {
	clusterScoped bool
}

func newServiceHandler() *ServiceHandler {
	return &ServiceHandler{clusterScoped: false}
}

func (h *ServiceHandler) List(ctx context.Context, clientset *kubernetes.Clientset, namespace, labelSelector, fieldSelector string) ([]model.SearchableItem, error) {
	services, err := service.ListServices(ctx, clientset, namespace, labelSelector, fieldSelector)
	if err != nil {
		return nil, err
	}
	result := make([]model.SearchableItem, len(services))
	for i := range services {
		result[i] = &services[i]
	}
	return result, nil
}

func (h *ServiceHandler) Get(ctx context.Context, clientset *kubernetes.Clientset, namespace, name string) (interface{}, error) {
	return clientset.CoreV1().Services(namespace).Get(ctx, name, metav1.GetOptions{})
}

func (h *ServiceHandler) Delete(ctx context.Context, clientset *kubernetes.Clientset, namespace, name string) error {
	return clientset.CoreV1().Services(namespace).Delete(ctx, name, metav1.DeleteOptions{})
}

func (h *ServiceHandler) IsClusterScoped() bool {
	return h.clusterScoped
}

// ConfigMapHandler ConfigMap 资源处理器
type ConfigMapHandler struct {
	clusterScoped bool
}

func newConfigMapHandler() *ConfigMapHandler {
	return &ConfigMapHandler{clusterScoped: false}
}

func (h *ConfigMapHandler) List(ctx context.Context, clientset *kubernetes.Clientset, namespace, labelSelector, fieldSelector string) ([]model.SearchableItem, error) {
	configMaps, err := service.ListConfigMaps(ctx, clientset, namespace, labelSelector, fieldSelector)
	if err != nil {
		return nil, err
	}
	result := make([]model.SearchableItem, len(configMaps))
	for i := range configMaps {
		result[i] = &configMaps[i]
	}
	return result, nil
}

func (h *ConfigMapHandler) Get(ctx context.Context, clientset *kubernetes.Clientset, namespace, name string) (interface{}, error) {
	return clientset.CoreV1().ConfigMaps(namespace).Get(ctx, name, metav1.GetOptions{})
}

func (h *ConfigMapHandler) Delete(ctx context.Context, clientset *kubernetes.Clientset, namespace, name string) error {
	return clientset.CoreV1().ConfigMaps(namespace).Delete(ctx, name, metav1.DeleteOptions{})
}

func (h *ConfigMapHandler) IsClusterScoped() bool {
	return h.clusterScoped
}

// SecretHandler Secret 资源处理器
type SecretHandler struct {
	clusterScoped bool
}

func newSecretHandler() *SecretHandler {
	return &SecretHandler{clusterScoped: false}
}

func (h *SecretHandler) List(ctx context.Context, clientset *kubernetes.Clientset, namespace, labelSelector, fieldSelector string) ([]model.SearchableItem, error) {
	secrets, err := service.ListSecrets(ctx, clientset, namespace, labelSelector, fieldSelector)
	if err != nil {
		return nil, err
	}
	result := make([]model.SearchableItem, len(secrets))
	for i := range secrets {
		result[i] = &secrets[i]
	}
	return result, nil
}

func (h *SecretHandler) Get(ctx context.Context, clientset *kubernetes.Clientset, namespace, name string) (interface{}, error) {
	return clientset.CoreV1().Secrets(namespace).Get(ctx, name, metav1.GetOptions{})
}

func (h *SecretHandler) Delete(ctx context.Context, clientset *kubernetes.Clientset, namespace, name string) error {
	return clientset.CoreV1().Secrets(namespace).Delete(ctx, name, metav1.DeleteOptions{})
}

func (h *SecretHandler) IsClusterScoped() bool {
	return h.clusterScoped
}

// IngressHandler Ingress 资源处理器
type IngressHandler struct {
	clusterScoped bool
}

func newIngressHandler() *IngressHandler {
	return &IngressHandler{clusterScoped: false}
}

func (h *IngressHandler) List(ctx context.Context, clientset *kubernetes.Clientset, namespace, labelSelector, fieldSelector string) ([]model.SearchableItem, error) {
	ingresses, err := service.ListIngresses(ctx, clientset, namespace, labelSelector, fieldSelector)
	if err != nil {
		return nil, err
	}
	result := make([]model.SearchableItem, len(ingresses))
	for i := range ingresses {
		result[i] = &ingresses[i]
	}
	return result, nil
}

func (h *IngressHandler) Get(ctx context.Context, clientset *kubernetes.Clientset, namespace, name string) (interface{}, error) {
	return clientset.NetworkingV1().Ingresses(namespace).Get(ctx, name, metav1.GetOptions{})
}

func (h *IngressHandler) Delete(ctx context.Context, clientset *kubernetes.Clientset, namespace, name string) error {
	return clientset.NetworkingV1().Ingresses(namespace).Delete(ctx, name, metav1.DeleteOptions{})
}

func (h *IngressHandler) IsClusterScoped() bool {
	return h.clusterScoped
}

// JobHandler Job 资源处理器
type JobHandler struct {
	clusterScoped bool
}

func newJobHandler() *JobHandler {
	return &JobHandler{clusterScoped: false}
}

func (h *JobHandler) List(ctx context.Context, clientset *kubernetes.Clientset, namespace, labelSelector, fieldSelector string) ([]model.SearchableItem, error) {
	jobs, err := service.ListJobs(ctx, clientset, namespace, labelSelector, fieldSelector)
	if err != nil {
		return nil, err
	}
	result := make([]model.SearchableItem, len(jobs))
	for i := range jobs {
		result[i] = &jobs[i]
	}
	return result, nil
}

func (h *JobHandler) Get(ctx context.Context, clientset *kubernetes.Clientset, namespace, name string) (interface{}, error) {
	return clientset.BatchV1().Jobs(namespace).Get(ctx, name, metav1.GetOptions{})
}

func (h *JobHandler) Delete(ctx context.Context, clientset *kubernetes.Clientset, namespace, name string) error {
	return clientset.BatchV1().Jobs(namespace).Delete(ctx, name, metav1.DeleteOptions{})
}

func (h *JobHandler) IsClusterScoped() bool {
	return h.clusterScoped
}

// CronJobHandler CronJob 资源处理器
type CronJobHandler struct {
	clusterScoped bool
}

func newCronJobHandler() *CronJobHandler {
	return &CronJobHandler{clusterScoped: false}
}

func (h *CronJobHandler) List(ctx context.Context, clientset *kubernetes.Clientset, namespace, labelSelector, fieldSelector string) ([]model.SearchableItem, error) {
	cronJobs, err := service.ListCronJobs(ctx, clientset, namespace, labelSelector, fieldSelector)
	if err != nil {
		return nil, err
	}
	result := make([]model.SearchableItem, len(cronJobs))
	for i := range cronJobs {
		result[i] = &cronJobs[i]
	}
	return result, nil
}

func (h *CronJobHandler) Get(ctx context.Context, clientset *kubernetes.Clientset, namespace, name string) (interface{}, error) {
	return clientset.BatchV1().CronJobs(namespace).Get(ctx, name, metav1.GetOptions{})
}

func (h *CronJobHandler) Delete(ctx context.Context, clientset *kubernetes.Clientset, namespace, name string) error {
	return clientset.BatchV1().CronJobs(namespace).Delete(ctx, name, metav1.DeleteOptions{})
}

func (h *CronJobHandler) IsClusterScoped() bool {
	return h.clusterScoped
}

// PVCHandler PersistentVolumeClaim 资源处理器
type PVCHandler struct {
	clusterScoped bool
}

func newPVCHandler() *PVCHandler {
	return &PVCHandler{clusterScoped: false}
}

func (h *PVCHandler) List(ctx context.Context, clientset *kubernetes.Clientset, namespace, labelSelector, fieldSelector string) ([]model.SearchableItem, error) {
	pvcs, err := service.ListPVCs(ctx, clientset, namespace, labelSelector, fieldSelector)
	if err != nil {
		return nil, err
	}
	result := make([]model.SearchableItem, len(pvcs))
	for i := range pvcs {
		result[i] = &pvcs[i]
	}
	return result, nil
}

func (h *PVCHandler) Get(ctx context.Context, clientset *kubernetes.Clientset, namespace, name string) (interface{}, error) {
	return clientset.CoreV1().PersistentVolumeClaims(namespace).Get(ctx, name, metav1.GetOptions{})
}

func (h *PVCHandler) Delete(ctx context.Context, clientset *kubernetes.Clientset, namespace, name string) error {
	return clientset.CoreV1().PersistentVolumeClaims(namespace).Delete(ctx, name, metav1.DeleteOptions{})
}

func (h *PVCHandler) IsClusterScoped() bool {
	return h.clusterScoped
}

// PVHandler PersistentVolume 资源处理器
type PVHandler struct {
	clusterScoped bool
}

func newPVHandler() *PVHandler {
	return &PVHandler{clusterScoped: true}
}

func (h *PVHandler) List(ctx context.Context, clientset *kubernetes.Clientset, namespace, labelSelector, fieldSelector string) ([]model.SearchableItem, error) {
	pvs, err := service.ListPVs(ctx, clientset, labelSelector, fieldSelector)
	if err != nil {
		return nil, err
	}
	result := make([]model.SearchableItem, len(pvs))
	for i := range pvs {
		result[i] = &pvs[i]
	}
	return result, nil
}

func (h *PVHandler) Get(ctx context.Context, clientset *kubernetes.Clientset, namespace, name string) (interface{}, error) {
	return clientset.CoreV1().PersistentVolumes().Get(ctx, name, metav1.GetOptions{})
}

func (h *PVHandler) Delete(ctx context.Context, clientset *kubernetes.Clientset, namespace, name string) error {
	return clientset.CoreV1().PersistentVolumes().Delete(ctx, name, metav1.DeleteOptions{})
}

func (h *PVHandler) IsClusterScoped() bool {
	return h.clusterScoped
}

// StorageClassHandler StorageClass 资源处理器
type StorageClassHandler struct {
	clusterScoped bool
}

func newStorageClassHandler() *StorageClassHandler {
	return &StorageClassHandler{clusterScoped: true}
}

func (h *StorageClassHandler) List(ctx context.Context, clientset *kubernetes.Clientset, namespace, labelSelector, fieldSelector string) ([]model.SearchableItem, error) {
	storageClasses, err := service.ListStorageClasses(ctx, clientset, labelSelector, fieldSelector)
	if err != nil {
		return nil, err
	}
	result := make([]model.SearchableItem, len(storageClasses))
	for i := range storageClasses {
		result[i] = &storageClasses[i]
	}
	return result, nil
}

func (h *StorageClassHandler) Get(ctx context.Context, clientset *kubernetes.Clientset, namespace, name string) (interface{}, error) {
	return clientset.StorageV1().StorageClasses().Get(ctx, name, metav1.GetOptions{})
}

func (h *StorageClassHandler) Delete(ctx context.Context, clientset *kubernetes.Clientset, namespace, name string) error {
	return clientset.StorageV1().StorageClasses().Delete(ctx, name, metav1.DeleteOptions{})
}

func (h *StorageClassHandler) IsClusterScoped() bool {
	return h.clusterScoped
}

// NamespaceHandler Namespace 资源处理器
type NamespaceHandler struct {
	clusterScoped bool
}

func newNamespaceHandler() *NamespaceHandler {
	return &NamespaceHandler{clusterScoped: true}
}

func (h *NamespaceHandler) List(ctx context.Context, clientset *kubernetes.Clientset, namespace, labelSelector, fieldSelector string) ([]model.SearchableItem, error) {
	namespaces, err := service.ListNamespaces(ctx, clientset, labelSelector, fieldSelector)
	if err != nil {
		return nil, err
	}
	result := make([]model.SearchableItem, len(namespaces))
	for i := range namespaces {
		result[i] = &namespaces[i]
	}
	return result, nil
}

func (h *NamespaceHandler) Get(ctx context.Context, clientset *kubernetes.Clientset, namespace, name string) (interface{}, error) {
	return clientset.CoreV1().Namespaces().Get(ctx, name, metav1.GetOptions{})
}

func (h *NamespaceHandler) Delete(ctx context.Context, clientset *kubernetes.Clientset, namespace, name string) error {
	return clientset.CoreV1().Namespaces().Delete(ctx, name, metav1.DeleteOptions{})
}

func (h *NamespaceHandler) IsClusterScoped() bool {
	return h.clusterScoped
}

// NodeHandler Node 资源处理器
type NodeHandler struct {
	clusterScoped bool
}

func newNodeHandler() *NodeHandler {
	return &NodeHandler{clusterScoped: true}
}

func (h *NodeHandler) List(ctx context.Context, clientset *kubernetes.Clientset, namespace, labelSelector, fieldSelector string) ([]model.SearchableItem, error) {
	// Note: metrics 需要在调用方处理，这里只返回基础节点信息
	nodes, err := service.ListNodes(ctx, clientset, nil, nil, labelSelector, fieldSelector)
	if err != nil {
		return nil, err
	}
	result := make([]model.SearchableItem, len(nodes))
	for i := range nodes {
		result[i] = &nodes[i]
	}
	return result, nil
}

func (h *NodeHandler) Get(ctx context.Context, clientset *kubernetes.Clientset, namespace, name string) (interface{}, error) {
	return clientset.CoreV1().Nodes().Get(ctx, name, metav1.GetOptions{})
}

func (h *NodeHandler) Delete(ctx context.Context, clientset *kubernetes.Clientset, namespace, name string) error {
	return clientset.CoreV1().Nodes().Delete(ctx, name, metav1.DeleteOptions{})
}

func (h *NodeHandler) IsClusterScoped() bool {
	return h.clusterScoped
}

// EndpointsHandler Endpoints 资源处理器
type EndpointsHandler struct {
	clusterScoped bool
}

func newEndpointsHandler() *EndpointsHandler {
	return &EndpointsHandler{clusterScoped: false}
}

func (h *EndpointsHandler) List(ctx context.Context, clientset *kubernetes.Clientset, namespace, labelSelector, fieldSelector string) ([]model.SearchableItem, error) {
	endpoints, err := service.ListEndpoints(ctx, clientset, namespace, labelSelector, fieldSelector)
	if err != nil {
		return nil, err
	}
	result := make([]model.SearchableItem, len(endpoints))
	for i := range endpoints {
		result[i] = &endpoints[i]
	}
	return result, nil
}

func (h *EndpointsHandler) Get(ctx context.Context, clientset *kubernetes.Clientset, namespace, name string) (interface{}, error) {
	return clientset.CoreV1().Endpoints(namespace).Get(ctx, name, metav1.GetOptions{})
}

func (h *EndpointsHandler) Delete(ctx context.Context, clientset *kubernetes.Clientset, namespace, name string) error {
	return clientset.CoreV1().Endpoints(namespace).Delete(ctx, name, metav1.DeleteOptions{})
}

func (h *EndpointsHandler) IsClusterScoped() bool {
	return h.clusterScoped
}

// EventHandler Event 资源处理器
type EventHandler struct {
	clusterScoped bool
}

func newEventHandler() *EventHandler {
	return &EventHandler{clusterScoped: false}
}

func (h *EventHandler) List(ctx context.Context, clientset *kubernetes.Clientset, namespace, labelSelector, fieldSelector string) ([]model.SearchableItem, error) {
	// Event 使用 involvedObject 和 since 参数
	events, err := service.ListEvents(ctx, clientset, namespace, "", "", labelSelector, fieldSelector)
	if err != nil {
		return nil, err
	}
	result := make([]model.SearchableItem, len(events))
	for i := range events {
		result[i] = &events[i]
	}
	return result, nil
}

func (h *EventHandler) Get(ctx context.Context, clientset *kubernetes.Clientset, namespace, name string) (interface{}, error) {
	return clientset.CoreV1().Events(namespace).Get(ctx, name, metav1.GetOptions{})
}

func (h *EventHandler) Delete(ctx context.Context, clientset *kubernetes.Clientset, namespace, name string) error {
	return clientset.CoreV1().Events(namespace).Delete(ctx, name, metav1.DeleteOptions{})
}

func (h *EventHandler) IsClusterScoped() bool {
	return h.clusterScoped
}

// ============== 工具函数 ==============

// getResourceByName 根据资源类型和名称获取对象（供 operations.go 使用）
func getResourceByName(ctx context.Context, clientset *kubernetes.Clientset, resourceType, namespace, name string) (interface{}, error) {
	registry := getResourceRegistry()
	if registry == nil {
		return nil, fmt.Errorf("resource registry not initialized")
	}

	handler, exists := registry.GetHandler(resourceType)
	if !exists {
		return nil, fmt.Errorf("unsupported resource type: %s", resourceType)
	}

	// 集群资源忽略 namespace
	if handler.IsClusterScoped() {
		namespace = ""
	}

	return handler.Get(ctx, clientset, namespace, name)
}
