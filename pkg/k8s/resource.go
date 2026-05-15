package k8s

import (
	"context"
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	autoscalingv2 "k8s.io/api/autoscaling/v2"
	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	policyv1 "k8s.io/api/policy/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	storagev1 "k8s.io/api/storage/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type ResourceType string

const (
	ResourcePod                     ResourceType = "pod"
	ResourceDeployment              ResourceType = "deployment"
	ResourceStatefulSet             ResourceType = "statefulset"
	ResourceDaemonSet               ResourceType = "daemonset"
	ResourceService                 ResourceType = "service"
	ResourceConfigMap               ResourceType = "configmap"
	ResourceSecret                  ResourceType = "secret"
	ResourceIngress                 ResourceType = "ingress"
	ResourceJob                     ResourceType = "job"
	ResourceCronJob                 ResourceType = "cronjob"
	ResourcePVC                     ResourceType = "persistentvolumeclaim"
	ResourcePV                      ResourceType = "persistentvolume"
	ResourceStorageClass            ResourceType = "storageclass"
	ResourceNamespace               ResourceType = "namespace"
	ResourceNode                    ResourceType = "node"
	ResourceEndpoint                ResourceType = "endpoint"
	ResourceEvent                   ResourceType = "event"
	ResourceHorizontalPodAutoscaler ResourceType = "horizontalpodautoscaler"
	ResourceNetworkPolicy           ResourceType = "networkpolicy"
	ResourceServiceAccount          ResourceType = "serviceaccount"
	ResourceRole                    ResourceType = "role"
	ResourceRoleBinding             ResourceType = "rolebinding"
	ResourceClusterRole             ResourceType = "clusterrole"
	ResourceClusterRoleBinding      ResourceType = "clusterrolebinding"
	ResourceResourceQuota           ResourceType = "resourcequota"
	ResourceLimitRange              ResourceType = "limitrange"
	ResourcePodDisruptionBudget     ResourceType = "poddisruptionbudget"
)

var clusterScopedResources = map[ResourceType]bool{
	ResourcePod:                     false,
	ResourceDeployment:              false,
	ResourceStatefulSet:             false,
	ResourceDaemonSet:               false,
	ResourceService:                 false,
	ResourceConfigMap:               false,
	ResourceSecret:                  false,
	ResourceIngress:                 false,
	ResourceJob:                     false,
	ResourceCronJob:                 false,
	ResourcePVC:                     false,
	ResourcePV:                      true,
	ResourceStorageClass:            true,
	ResourceNamespace:               true,
	ResourceNode:                    true,
	ResourceEndpoint:                false,
	ResourceEvent:                   false,
	ResourceHorizontalPodAutoscaler: false,
	ResourceNetworkPolicy:           false,
	ResourceServiceAccount:          false,
	ResourceRole:                    false,
	ResourceRoleBinding:             false,
	ResourceClusterRole:             true,
	ResourceClusterRoleBinding:      true,
	ResourceResourceQuota:           false,
	ResourceLimitRange:              false,
	ResourcePodDisruptionBudget:     false,
}

func (rt ResourceType) IsClusterScoped() bool {
	return clusterScopedResources[rt]
}

func (rt ResourceType) Normalize() ResourceType {
	switch string(rt) {
	case "pvc":
		return ResourcePVC
	case "pv":
		return ResourcePV
	case "sc":
		return ResourceStorageClass
	case "hpa":
		return ResourceHorizontalPodAutoscaler
	case "netpol":
		return ResourceNetworkPolicy
	case "sa":
		return ResourceServiceAccount
	case "quota":
		return ResourceResourceQuota
	case "pdb":
		return ResourcePodDisruptionBudget
	default:
		return rt
	}
}

type Getter interface {
	Get(ctx context.Context, namespace, name string) (any, error)
	List(ctx context.Context, namespace string, opts metav1.ListOptions) (any, error)
}

type Lister interface {
	List(ctx context.Context, namespace string, opts metav1.ListOptions) (any, error)
}

type Updater interface {
	Update(ctx context.Context, namespace, name string, obj any) error
}

type Deleter interface {
	Delete(ctx context.Context, namespace, name string) error
}

type Creator interface {
	Create(ctx context.Context, namespace string, obj any) error
}

type resourceGetter struct {
	getFn  func(ctx context.Context, namespace, name string) (any, error)
	listFn func(ctx context.Context, namespace string, opts metav1.ListOptions) (any, error)
}

func (g *resourceGetter) Get(ctx context.Context, namespace, name string) (any, error) {
	return g.getFn(ctx, namespace, name)
}

func (g *resourceGetter) List(ctx context.Context, namespace string, opts metav1.ListOptions) (any, error) {
	return g.listFn(ctx, namespace, opts)
}

type resourceUpdater struct {
	updateFn func(ctx context.Context, namespace, name string, obj any) error
	deleteFn func(ctx context.Context, namespace, name string) error
	createFn func(ctx context.Context, namespace string, obj any) error
}

func (u *resourceUpdater) Update(ctx context.Context, namespace, name string, obj any) error {
	return u.updateFn(ctx, namespace, name, obj)
}

func (u *resourceUpdater) Delete(ctx context.Context, namespace, name string) error {
	return u.deleteFn(ctx, namespace, name)
}

func (u *resourceUpdater) Create(ctx context.Context, namespace string, obj any) error {
	return u.createFn(ctx, namespace, obj)
}

type resourceDeleter struct {
	deleteFn func(ctx context.Context, namespace, name string) error
}

func (d *resourceDeleter) Delete(ctx context.Context, namespace, name string) error {
	return d.deleteFn(ctx, namespace, name)
}

type resourceCreator struct {
	createFn func(ctx context.Context, namespace string, obj any) error
}

func (c *resourceCreator) Create(ctx context.Context, namespace string, obj any) error {
	return c.createFn(ctx, namespace, obj)
}

func NewGetters(client kubernetes.Interface) map[ResourceType]Getter {
	c := client
	return map[ResourceType]Getter{
		ResourcePod: &resourceGetter{
			getFn: func(ctx context.Context, ns, name string) (any, error) {
				return c.CoreV1().Pods(ns).Get(ctx, name, metav1.GetOptions{})
			},
			listFn: func(ctx context.Context, ns string, opts metav1.ListOptions) (any, error) {
				return c.CoreV1().Pods(ns).List(ctx, opts)
			},
		},
		ResourceDeployment: &resourceGetter{
			getFn: func(ctx context.Context, ns, name string) (any, error) {
				return c.AppsV1().Deployments(ns).Get(ctx, name, metav1.GetOptions{})
			},
			listFn: func(ctx context.Context, ns string, opts metav1.ListOptions) (any, error) {
				return c.AppsV1().Deployments(ns).List(ctx, opts)
			},
		},
		ResourceStatefulSet: &resourceGetter{
			getFn: func(ctx context.Context, ns, name string) (any, error) {
				return c.AppsV1().StatefulSets(ns).Get(ctx, name, metav1.GetOptions{})
			},
			listFn: func(ctx context.Context, ns string, opts metav1.ListOptions) (any, error) {
				return c.AppsV1().StatefulSets(ns).List(ctx, opts)
			},
		},
		ResourceDaemonSet: &resourceGetter{
			getFn: func(ctx context.Context, ns, name string) (any, error) {
				return c.AppsV1().DaemonSets(ns).Get(ctx, name, metav1.GetOptions{})
			},
			listFn: func(ctx context.Context, ns string, opts metav1.ListOptions) (any, error) {
				return c.AppsV1().DaemonSets(ns).List(ctx, opts)
			},
		},
		ResourceService: &resourceGetter{
			getFn: func(ctx context.Context, ns, name string) (any, error) {
				return c.CoreV1().Services(ns).Get(ctx, name, metav1.GetOptions{})
			},
			listFn: func(ctx context.Context, ns string, opts metav1.ListOptions) (any, error) {
				return c.CoreV1().Services(ns).List(ctx, opts)
			},
		},
		ResourceConfigMap: &resourceGetter{
			getFn: func(ctx context.Context, ns, name string) (any, error) {
				return c.CoreV1().ConfigMaps(ns).Get(ctx, name, metav1.GetOptions{})
			},
			listFn: func(ctx context.Context, ns string, opts metav1.ListOptions) (any, error) {
				return c.CoreV1().ConfigMaps(ns).List(ctx, opts)
			},
		},
		ResourceSecret: &resourceGetter{
			getFn: func(ctx context.Context, ns, name string) (any, error) {
				return c.CoreV1().Secrets(ns).Get(ctx, name, metav1.GetOptions{})
			},
			listFn: func(ctx context.Context, ns string, opts metav1.ListOptions) (any, error) {
				return c.CoreV1().Secrets(ns).List(ctx, opts)
			},
		},
		ResourceIngress: &resourceGetter{
			getFn: func(ctx context.Context, ns, name string) (any, error) {
				return c.NetworkingV1().Ingresses(ns).Get(ctx, name, metav1.GetOptions{})
			},
			listFn: func(ctx context.Context, ns string, opts metav1.ListOptions) (any, error) {
				return c.NetworkingV1().Ingresses(ns).List(ctx, opts)
			},
		},
		ResourceJob: &resourceGetter{
			getFn: func(ctx context.Context, ns, name string) (any, error) {
				return c.BatchV1().Jobs(ns).Get(ctx, name, metav1.GetOptions{})
			},
			listFn: func(ctx context.Context, ns string, opts metav1.ListOptions) (any, error) {
				return c.BatchV1().Jobs(ns).List(ctx, opts)
			},
		},
		ResourceCronJob: &resourceGetter{
			getFn: func(ctx context.Context, ns, name string) (any, error) {
				return c.BatchV1().CronJobs(ns).Get(ctx, name, metav1.GetOptions{})
			},
			listFn: func(ctx context.Context, ns string, opts metav1.ListOptions) (any, error) {
				return c.BatchV1().CronJobs(ns).List(ctx, opts)
			},
		},
		ResourcePVC: &resourceGetter{
			getFn: func(ctx context.Context, ns, name string) (any, error) {
				return c.CoreV1().PersistentVolumeClaims(ns).Get(ctx, name, metav1.GetOptions{})
			},
			listFn: func(ctx context.Context, ns string, opts metav1.ListOptions) (any, error) {
				return c.CoreV1().PersistentVolumeClaims(ns).List(ctx, opts)
			},
		},
		ResourcePV: &resourceGetter{
			getFn: func(ctx context.Context, ns, name string) (any, error) {
				return c.CoreV1().PersistentVolumes().Get(ctx, name, metav1.GetOptions{})
			},
			listFn: func(ctx context.Context, ns string, opts metav1.ListOptions) (any, error) {
				return c.CoreV1().PersistentVolumes().List(ctx, opts)
			},
		},
		ResourceStorageClass: &resourceGetter{
			getFn: func(ctx context.Context, ns, name string) (any, error) {
				return c.StorageV1().StorageClasses().Get(ctx, name, metav1.GetOptions{})
			},
			listFn: func(ctx context.Context, ns string, opts metav1.ListOptions) (any, error) {
				return c.StorageV1().StorageClasses().List(ctx, opts)
			},
		},
		ResourceNamespace: &resourceGetter{
			getFn: func(ctx context.Context, ns, name string) (any, error) {
				return c.CoreV1().Namespaces().Get(ctx, name, metav1.GetOptions{})
			},
			listFn: func(ctx context.Context, ns string, opts metav1.ListOptions) (any, error) {
				return c.CoreV1().Namespaces().List(ctx, opts)
			},
		},
		ResourceNode: &resourceGetter{
			getFn: func(ctx context.Context, ns, name string) (any, error) {
				return c.CoreV1().Nodes().Get(ctx, name, metav1.GetOptions{})
			},
			listFn: func(ctx context.Context, ns string, opts metav1.ListOptions) (any, error) {
				return c.CoreV1().Nodes().List(ctx, opts)
			},
		},
		ResourceEndpoint: &resourceGetter{
			getFn: func(ctx context.Context, ns, name string) (any, error) {
				return c.CoreV1().Endpoints(ns).Get(ctx, name, metav1.GetOptions{})
			},
			listFn: func(ctx context.Context, ns string, opts metav1.ListOptions) (any, error) {
				return c.CoreV1().Endpoints(ns).List(ctx, opts)
			},
		},
		ResourceEvent: &resourceGetter{
			getFn: func(ctx context.Context, ns, name string) (any, error) {
				return c.CoreV1().Events(ns).Get(ctx, name, metav1.GetOptions{})
			},
			listFn: func(ctx context.Context, ns string, opts metav1.ListOptions) (any, error) {
				return c.CoreV1().Events(ns).List(ctx, opts)
			},
		},
		ResourceHorizontalPodAutoscaler: &resourceGetter{
			getFn: func(ctx context.Context, ns, name string) (any, error) {
				return c.AutoscalingV2().HorizontalPodAutoscalers(ns).Get(ctx, name, metav1.GetOptions{})
			},
			listFn: func(ctx context.Context, ns string, opts metav1.ListOptions) (any, error) {
				return c.AutoscalingV2().HorizontalPodAutoscalers(ns).List(ctx, opts)
			},
		},
		ResourceNetworkPolicy: &resourceGetter{
			getFn: func(ctx context.Context, ns, name string) (any, error) {
				return c.NetworkingV1().NetworkPolicies(ns).Get(ctx, name, metav1.GetOptions{})
			},
			listFn: func(ctx context.Context, ns string, opts metav1.ListOptions) (any, error) {
				return c.NetworkingV1().NetworkPolicies(ns).List(ctx, opts)
			},
		},
		ResourceServiceAccount: &resourceGetter{
			getFn: func(ctx context.Context, ns, name string) (any, error) {
				return c.CoreV1().ServiceAccounts(ns).Get(ctx, name, metav1.GetOptions{})
			},
			listFn: func(ctx context.Context, ns string, opts metav1.ListOptions) (any, error) {
				return c.CoreV1().ServiceAccounts(ns).List(ctx, opts)
			},
		},
		ResourceRole: &resourceGetter{
			getFn: func(ctx context.Context, ns, name string) (any, error) {
				return c.RbacV1().Roles(ns).Get(ctx, name, metav1.GetOptions{})
			},
			listFn: func(ctx context.Context, ns string, opts metav1.ListOptions) (any, error) {
				return c.RbacV1().Roles(ns).List(ctx, opts)
			},
		},
		ResourceRoleBinding: &resourceGetter{
			getFn: func(ctx context.Context, ns, name string) (any, error) {
				return c.RbacV1().RoleBindings(ns).Get(ctx, name, metav1.GetOptions{})
			},
			listFn: func(ctx context.Context, ns string, opts metav1.ListOptions) (any, error) {
				return c.RbacV1().RoleBindings(ns).List(ctx, opts)
			},
		},
		ResourceClusterRole: &resourceGetter{
			getFn: func(ctx context.Context, ns, name string) (any, error) {
				return c.RbacV1().ClusterRoles().Get(ctx, name, metav1.GetOptions{})
			},
			listFn: func(ctx context.Context, ns string, opts metav1.ListOptions) (any, error) {
				return c.RbacV1().ClusterRoles().List(ctx, opts)
			},
		},
		ResourceClusterRoleBinding: &resourceGetter{
			getFn: func(ctx context.Context, ns, name string) (any, error) {
				return c.RbacV1().ClusterRoleBindings().Get(ctx, name, metav1.GetOptions{})
			},
			listFn: func(ctx context.Context, ns string, opts metav1.ListOptions) (any, error) {
				return c.RbacV1().ClusterRoleBindings().List(ctx, opts)
			},
		},
		ResourceResourceQuota: &resourceGetter{
			getFn: func(ctx context.Context, ns, name string) (any, error) {
				return c.CoreV1().ResourceQuotas(ns).Get(ctx, name, metav1.GetOptions{})
			},
			listFn: func(ctx context.Context, ns string, opts metav1.ListOptions) (any, error) {
				return c.CoreV1().ResourceQuotas(ns).List(ctx, opts)
			},
		},
		ResourceLimitRange: &resourceGetter{
			getFn: func(ctx context.Context, ns, name string) (any, error) {
				return c.CoreV1().LimitRanges(ns).Get(ctx, name, metav1.GetOptions{})
			},
			listFn: func(ctx context.Context, ns string, opts metav1.ListOptions) (any, error) {
				return c.CoreV1().LimitRanges(ns).List(ctx, opts)
			},
		},
		ResourcePodDisruptionBudget: &resourceGetter{
			getFn: func(ctx context.Context, ns, name string) (any, error) {
				return c.PolicyV1().PodDisruptionBudgets(ns).Get(ctx, name, metav1.GetOptions{})
			},
			listFn: func(ctx context.Context, ns string, opts metav1.ListOptions) (any, error) {
				return c.PolicyV1().PodDisruptionBudgets(ns).List(ctx, opts)
			},
		},
	}
}

func NewUpdaters(client kubernetes.Interface) map[ResourceType]Updater {
	c := client
	u := func(update func(ctx context.Context, namespace, name string, obj any) error, del func(ctx context.Context, namespace, name string) error) *resourceUpdater {
		return &resourceUpdater{updateFn: update, deleteFn: del}
	}
	return map[ResourceType]Updater{
		ResourcePod: u(
			func(ctx context.Context, ns, name string, obj any) error {
				pod, _ := obj.(*v1.Pod)
				_, e := c.CoreV1().Pods(ns).Update(ctx, pod, metav1.UpdateOptions{})
				return e
			},
			func(ctx context.Context, ns, name string) error {
				return c.CoreV1().Pods(ns).Delete(ctx, name, metav1.DeleteOptions{})
			},
		),
		ResourceDeployment: u(
			func(ctx context.Context, ns, name string, obj any) error {
				dep, _ := obj.(*appsv1.Deployment)
				_, e := c.AppsV1().Deployments(ns).Update(ctx, dep, metav1.UpdateOptions{})
				return e
			},
			func(ctx context.Context, ns, name string) error {
				return c.AppsV1().Deployments(ns).Delete(ctx, name, metav1.DeleteOptions{})
			},
		),
		ResourceStatefulSet: u(
			func(ctx context.Context, ns, name string, obj any) error {
				sts, _ := obj.(*appsv1.StatefulSet)
				_, e := c.AppsV1().StatefulSets(ns).Update(ctx, sts, metav1.UpdateOptions{})
				return e
			},
			func(ctx context.Context, ns, name string) error {
				return c.AppsV1().StatefulSets(ns).Delete(ctx, name, metav1.DeleteOptions{})
			},
		),
		ResourceDaemonSet: u(
			func(ctx context.Context, ns, name string, obj any) error {
				ds, _ := obj.(*appsv1.DaemonSet)
				_, e := c.AppsV1().DaemonSets(ns).Update(ctx, ds, metav1.UpdateOptions{})
				return e
			},
			func(ctx context.Context, ns, name string) error {
				return c.AppsV1().DaemonSets(ns).Delete(ctx, name, metav1.DeleteOptions{})
			},
		),
		ResourceService: u(
			func(ctx context.Context, ns, name string, obj any) error {
				svc, _ := obj.(*v1.Service)
				_, e := c.CoreV1().Services(ns).Update(ctx, svc, metav1.UpdateOptions{})
				return e
			},
			func(ctx context.Context, ns, name string) error {
				return c.CoreV1().Services(ns).Delete(ctx, name, metav1.DeleteOptions{})
			},
		),
		ResourceConfigMap: u(
			func(ctx context.Context, ns, name string, obj any) error {
				cm, _ := obj.(*v1.ConfigMap)
				_, e := c.CoreV1().ConfigMaps(ns).Update(ctx, cm, metav1.UpdateOptions{})
				return e
			},
			func(ctx context.Context, ns, name string) error {
				return c.CoreV1().ConfigMaps(ns).Delete(ctx, name, metav1.DeleteOptions{})
			},
		),
		ResourceSecret: u(
			func(ctx context.Context, ns, name string, obj any) error {
				s, _ := obj.(*v1.Secret)
				_, e := c.CoreV1().Secrets(ns).Update(ctx, s, metav1.UpdateOptions{})
				return e
			},
			func(ctx context.Context, ns, name string) error {
				return c.CoreV1().Secrets(ns).Delete(ctx, name, metav1.DeleteOptions{})
			},
		),
		ResourceIngress: u(
			func(ctx context.Context, ns, name string, obj any) error {
				ing, _ := obj.(*networkingv1.Ingress)
				_, e := c.NetworkingV1().Ingresses(ns).Update(ctx, ing, metav1.UpdateOptions{})
				return e
			},
			func(ctx context.Context, ns, name string) error {
				return c.NetworkingV1().Ingresses(ns).Delete(ctx, name, metav1.DeleteOptions{})
			},
		),
		ResourceJob: u(
			func(ctx context.Context, ns, name string, obj any) error {
				job, _ := obj.(*batchv1.Job)
				_, e := c.BatchV1().Jobs(ns).Update(ctx, job, metav1.UpdateOptions{})
				return e
			},
			func(ctx context.Context, ns, name string) error {
				return c.BatchV1().Jobs(ns).Delete(ctx, name, metav1.DeleteOptions{})
			},
		),
		ResourceCronJob: u(
			func(ctx context.Context, ns, name string, obj any) error {
				cj, _ := obj.(*batchv1.CronJob)
				_, e := c.BatchV1().CronJobs(ns).Update(ctx, cj, metav1.UpdateOptions{})
				return e
			},
			func(ctx context.Context, ns, name string) error {
				return c.BatchV1().CronJobs(ns).Delete(ctx, name, metav1.DeleteOptions{})
			},
		),
		ResourcePVC: u(
			func(ctx context.Context, ns, name string, obj any) error {
				pvc, _ := obj.(*v1.PersistentVolumeClaim)
				_, e := c.CoreV1().PersistentVolumeClaims(ns).Update(ctx, pvc, metav1.UpdateOptions{})
				return e
			},
			func(ctx context.Context, ns, name string) error {
				return c.CoreV1().PersistentVolumeClaims(ns).Delete(ctx, name, metav1.DeleteOptions{})
			},
		),
		ResourcePV: u(
			func(ctx context.Context, ns, name string, obj any) error {
				pv, _ := obj.(*v1.PersistentVolume)
				_, e := c.CoreV1().PersistentVolumes().Update(ctx, pv, metav1.UpdateOptions{})
				return e
			},
			func(ctx context.Context, ns, name string) error {
				return c.CoreV1().PersistentVolumes().Delete(ctx, name, metav1.DeleteOptions{})
			},
		),
		ResourceStorageClass: u(
			func(ctx context.Context, ns, name string, obj any) error {
				sc, _ := obj.(*storagev1.StorageClass)
				_, e := c.StorageV1().StorageClasses().Update(ctx, sc, metav1.UpdateOptions{})
				return e
			},
			func(ctx context.Context, ns, name string) error {
				return c.StorageV1().StorageClasses().Delete(ctx, name, metav1.DeleteOptions{})
			},
		),
		ResourceNamespace: u(
			func(ctx context.Context, ns, name string, obj any) error {
				nsObj, _ := obj.(*v1.Namespace)
				_, e := c.CoreV1().Namespaces().Update(ctx, nsObj, metav1.UpdateOptions{})
				return e
			},
			func(ctx context.Context, ns, name string) error {
				return c.CoreV1().Namespaces().Delete(ctx, name, metav1.DeleteOptions{})
			},
		),
		ResourceNode: u(
			func(ctx context.Context, ns, name string, obj any) error {
				node, _ := obj.(*v1.Node)
				_, e := c.CoreV1().Nodes().Update(ctx, node, metav1.UpdateOptions{})
				return e
			},
			func(ctx context.Context, ns, name string) error {
				return c.CoreV1().Nodes().Delete(ctx, name, metav1.DeleteOptions{})
			},
		),
		ResourceHorizontalPodAutoscaler: &resourceUpdater{
			updateFn: func(ctx context.Context, ns, name string, obj any) error {
				hpa, _ := obj.(*autoscalingv2.HorizontalPodAutoscaler)
				_, e := c.AutoscalingV2().HorizontalPodAutoscalers(ns).Update(ctx, hpa, metav1.UpdateOptions{})
				return e
			},
			deleteFn: func(ctx context.Context, ns, name string) error {
				return c.AutoscalingV2().HorizontalPodAutoscalers(ns).Delete(ctx, name, metav1.DeleteOptions{})
			},
			createFn: func(ctx context.Context, ns string, obj any) error {
				hpa, _ := obj.(*autoscalingv2.HorizontalPodAutoscaler)
				_, e := c.AutoscalingV2().HorizontalPodAutoscalers(ns).Create(ctx, hpa, metav1.CreateOptions{})
				return e
			},
		},
		ResourceNetworkPolicy: &resourceUpdater{
			updateFn: func(ctx context.Context, ns, name string, obj any) error {
				np, _ := obj.(*networkingv1.NetworkPolicy)
				_, e := c.NetworkingV1().NetworkPolicies(ns).Update(ctx, np, metav1.UpdateOptions{})
				return e
			},
			deleteFn: func(ctx context.Context, ns, name string) error {
				return c.NetworkingV1().NetworkPolicies(ns).Delete(ctx, name, metav1.DeleteOptions{})
			},
			createFn: func(ctx context.Context, ns string, obj any) error {
				np, _ := obj.(*networkingv1.NetworkPolicy)
				_, e := c.NetworkingV1().NetworkPolicies(ns).Create(ctx, np, metav1.CreateOptions{})
				return e
			},
		},
		ResourceServiceAccount: &resourceUpdater{
			updateFn: func(ctx context.Context, ns, name string, obj any) error {
				sa, _ := obj.(*v1.ServiceAccount)
				_, e := c.CoreV1().ServiceAccounts(ns).Update(ctx, sa, metav1.UpdateOptions{})
				return e
			},
			deleteFn: func(ctx context.Context, ns, name string) error {
				return c.CoreV1().ServiceAccounts(ns).Delete(ctx, name, metav1.DeleteOptions{})
			},
			createFn: func(ctx context.Context, ns string, obj any) error {
				sa, _ := obj.(*v1.ServiceAccount)
				_, e := c.CoreV1().ServiceAccounts(ns).Create(ctx, sa, metav1.CreateOptions{})
				return e
			},
		},
		ResourceRole: &resourceUpdater{
			updateFn: func(ctx context.Context, ns, name string, obj any) error {
				role, _ := obj.(*rbacv1.Role)
				_, e := c.RbacV1().Roles(ns).Update(ctx, role, metav1.UpdateOptions{})
				return e
			},
			deleteFn: func(ctx context.Context, ns, name string) error {
				return c.RbacV1().Roles(ns).Delete(ctx, name, metav1.DeleteOptions{})
			},
			createFn: func(ctx context.Context, ns string, obj any) error {
				role, _ := obj.(*rbacv1.Role)
				_, e := c.RbacV1().Roles(ns).Create(ctx, role, metav1.CreateOptions{})
				return e
			},
		},
		ResourceRoleBinding: &resourceUpdater{
			updateFn: func(ctx context.Context, ns, name string, obj any) error {
				rb, _ := obj.(*rbacv1.RoleBinding)
				_, e := c.RbacV1().RoleBindings(ns).Update(ctx, rb, metav1.UpdateOptions{})
				return e
			},
			deleteFn: func(ctx context.Context, ns, name string) error {
				return c.RbacV1().RoleBindings(ns).Delete(ctx, name, metav1.DeleteOptions{})
			},
			createFn: func(ctx context.Context, ns string, obj any) error {
				rb, _ := obj.(*rbacv1.RoleBinding)
				_, e := c.RbacV1().RoleBindings(ns).Create(ctx, rb, metav1.CreateOptions{})
				return e
			},
		},
		ResourceClusterRole: &resourceUpdater{
			updateFn: func(ctx context.Context, ns, name string, obj any) error {
				cr, _ := obj.(*rbacv1.ClusterRole)
				_, e := c.RbacV1().ClusterRoles().Update(ctx, cr, metav1.UpdateOptions{})
				return e
			},
			deleteFn: func(ctx context.Context, ns, name string) error {
				return c.RbacV1().ClusterRoles().Delete(ctx, name, metav1.DeleteOptions{})
			},
			createFn: func(ctx context.Context, ns string, obj any) error {
				cr, _ := obj.(*rbacv1.ClusterRole)
				_, e := c.RbacV1().ClusterRoles().Create(ctx, cr, metav1.CreateOptions{})
				return e
			},
		},
		ResourceClusterRoleBinding: &resourceUpdater{
			updateFn: func(ctx context.Context, ns, name string, obj any) error {
				crb, _ := obj.(*rbacv1.ClusterRoleBinding)
				_, e := c.RbacV1().ClusterRoleBindings().Update(ctx, crb, metav1.UpdateOptions{})
				return e
			},
			deleteFn: func(ctx context.Context, ns, name string) error {
				return c.RbacV1().ClusterRoleBindings().Delete(ctx, name, metav1.DeleteOptions{})
			},
			createFn: func(ctx context.Context, ns string, obj any) error {
				crb, _ := obj.(*rbacv1.ClusterRoleBinding)
				_, e := c.RbacV1().ClusterRoleBindings().Create(ctx, crb, metav1.CreateOptions{})
				return e
			},
		},
		ResourceResourceQuota: &resourceUpdater{
			updateFn: func(ctx context.Context, ns, name string, obj any) error {
				rq, _ := obj.(*v1.ResourceQuota)
				_, e := c.CoreV1().ResourceQuotas(ns).Update(ctx, rq, metav1.UpdateOptions{})
				return e
			},
			deleteFn: func(ctx context.Context, ns, name string) error {
				return c.CoreV1().ResourceQuotas(ns).Delete(ctx, name, metav1.DeleteOptions{})
			},
			createFn: func(ctx context.Context, ns string, obj any) error {
				rq, _ := obj.(*v1.ResourceQuota)
				_, e := c.CoreV1().ResourceQuotas(ns).Create(ctx, rq, metav1.CreateOptions{})
				return e
			},
		},
		ResourceLimitRange: &resourceUpdater{
			updateFn: func(ctx context.Context, ns, name string, obj any) error {
				lr, _ := obj.(*v1.LimitRange)
				_, e := c.CoreV1().LimitRanges(ns).Update(ctx, lr, metav1.UpdateOptions{})
				return e
			},
			deleteFn: func(ctx context.Context, ns, name string) error {
				return c.CoreV1().LimitRanges(ns).Delete(ctx, name, metav1.DeleteOptions{})
			},
			createFn: func(ctx context.Context, ns string, obj any) error {
				lr, _ := obj.(*v1.LimitRange)
				_, e := c.CoreV1().LimitRanges(ns).Create(ctx, lr, metav1.CreateOptions{})
				return e
			},
		},
		ResourcePodDisruptionBudget: &resourceUpdater{
			updateFn: func(ctx context.Context, ns, name string, obj any) error {
				pdb, _ := obj.(*policyv1.PodDisruptionBudget)
				_, e := c.PolicyV1().PodDisruptionBudgets(ns).Update(ctx, pdb, metav1.UpdateOptions{})
				return e
			},
			deleteFn: func(ctx context.Context, ns, name string) error {
				return c.PolicyV1().PodDisruptionBudgets(ns).Delete(ctx, name, metav1.DeleteOptions{})
			},
			createFn: func(ctx context.Context, ns string, obj any) error {
				pdb, _ := obj.(*policyv1.PodDisruptionBudget)
				_, e := c.PolicyV1().PodDisruptionBudgets(ns).Create(ctx, pdb, metav1.CreateOptions{})
				return e
			},
		},
	}
}

// GetKindByResourceType 根据资源类型返回对应的 Kind
func GetKindByResourceType(resourceType string) string {
	switch resourceType {
	case "pod":
		return "Pod"
	case "deployment":
		return "Deployment"
	case "statefulset":
		return "StatefulSet"
	case "daemonset":
		return "DaemonSet"
	case "service":
		return "Service"
	case "configmap":
		return "ConfigMap"
	case "secret":
		return "Secret"
	case "ingress":
		return "Ingress"
	case "job":
		return "Job"
	case "cronjob":
		return "CronJob"
	case "persistentvolumeclaim", "pvc":
		return "PersistentVolumeClaim"
	case "persistentvolume", "pv":
		return "PersistentVolume"
	case "storageclass", "sc":
		return "StorageClass"
	case "namespace":
		return "Namespace"
	case "node":
		return "Node"
	case "endpoint":
		return "Endpoint"
	case "event":
		return "Event"
	case "horizontalpodautoscaler", "hpa":
		return "HorizontalPodAutoscaler"
	case "networkpolicy", "netpol":
		return "NetworkPolicy"
	case "serviceaccount", "sa":
		return "ServiceAccount"
	case "role":
		return "Role"
	case "rolebinding":
		return "RoleBinding"
	case "clusterrole":
		return "ClusterRole"
	case "clusterrolebinding":
		return "ClusterRoleBinding"
	case "resourcequota", "quota":
		return "ResourceQuota"
	case "limitrange":
		return "LimitRange"
	case "poddisruptionbudget", "pdb":
		return "PodDisruptionBudget"
	default:
		return ""
	}
}

// NewCreators 创建资源创建器映射
func NewCreators(client kubernetes.Interface) map[ResourceType]Creator {
	return map[ResourceType]Creator{
		ResourcePod:          &podsCreator{client},
		ResourceDeployment:   &deploymentsCreator{client},
		ResourceStatefulSet:  &statefulSetsCreator{client},
		ResourceDaemonSet:    &daemonSetsCreator{client},
		ResourceService:      &servicesCreator{client},
		ResourceConfigMap:    &configMapsCreator{client},
		ResourceSecret:       &secretsCreator{client},
		ResourceIngress:      &ingressesCreator{client},
		ResourceJob:          &jobsCreator{client},
		ResourceCronJob:      &cronJobsCreator{client},
		ResourcePVC:          &pvcsCreator{client},
		ResourcePV:           &pvsCreator{client},
		ResourceStorageClass: &storageClassesCreator{client},
		ResourceNamespace:    &namespacesCreator{client},
		ResourceNode:         &nodesCreator{client},
		ResourceHorizontalPodAutoscaler: &resourceUpdater{
			createFn: func(ctx context.Context, ns string, obj any) error {
				hpa, _ := obj.(*autoscalingv2.HorizontalPodAutoscaler)
				_, e := client.AutoscalingV2().HorizontalPodAutoscalers(ns).Create(ctx, hpa, metav1.CreateOptions{})
				return e
			},
		},
		ResourceNetworkPolicy: &resourceUpdater{
			createFn: func(ctx context.Context, ns string, obj any) error {
				np, _ := obj.(*networkingv1.NetworkPolicy)
				_, e := client.NetworkingV1().NetworkPolicies(ns).Create(ctx, np, metav1.CreateOptions{})
				return e
			},
		},
		ResourceServiceAccount: &resourceUpdater{
			createFn: func(ctx context.Context, ns string, obj any) error {
				sa, _ := obj.(*v1.ServiceAccount)
				_, e := client.CoreV1().ServiceAccounts(ns).Create(ctx, sa, metav1.CreateOptions{})
				return e
			},
		},
		ResourceRole: &resourceUpdater{
			createFn: func(ctx context.Context, ns string, obj any) error {
				role, _ := obj.(*rbacv1.Role)
				_, e := client.RbacV1().Roles(ns).Create(ctx, role, metav1.CreateOptions{})
				return e
			},
		},
		ResourceRoleBinding: &resourceUpdater{
			createFn: func(ctx context.Context, ns string, obj any) error {
				rb, _ := obj.(*rbacv1.RoleBinding)
				_, e := client.RbacV1().RoleBindings(ns).Create(ctx, rb, metav1.CreateOptions{})
				return e
			},
		},
		ResourceClusterRole: &resourceUpdater{
			createFn: func(ctx context.Context, ns string, obj any) error {
				cr, _ := obj.(*rbacv1.ClusterRole)
				_, e := client.RbacV1().ClusterRoles().Create(ctx, cr, metav1.CreateOptions{})
				return e
			},
		},
		ResourceClusterRoleBinding: &resourceUpdater{
			createFn: func(ctx context.Context, ns string, obj any) error {
				crb, _ := obj.(*rbacv1.ClusterRoleBinding)
				_, e := client.RbacV1().ClusterRoleBindings().Create(ctx, crb, metav1.CreateOptions{})
				return e
			},
		},
		ResourceResourceQuota: &resourceUpdater{
			createFn: func(ctx context.Context, ns string, obj any) error {
				rq, _ := obj.(*v1.ResourceQuota)
				_, e := client.CoreV1().ResourceQuotas(ns).Create(ctx, rq, metav1.CreateOptions{})
				return e
			},
		},
		ResourceLimitRange: &resourceUpdater{
			createFn: func(ctx context.Context, ns string, obj any) error {
				lr, _ := obj.(*v1.LimitRange)
				_, e := client.CoreV1().LimitRanges(ns).Create(ctx, lr, metav1.CreateOptions{})
				return e
			},
		},
		ResourcePodDisruptionBudget: &resourceUpdater{
			createFn: func(ctx context.Context, ns string, obj any) error {
				pdb, _ := obj.(*policyv1.PodDisruptionBudget)
				_, e := client.PolicyV1().PodDisruptionBudgets(ns).Create(ctx, pdb, metav1.CreateOptions{})
				return e
			},
		},
	}
}

// NewDeleters 创建资源删除器映射
func NewDeleters(client kubernetes.Interface) map[ResourceType]Deleter {
	return map[ResourceType]Deleter{
		ResourcePod: &resourceDeleter{deleteFn: func(ctx context.Context, ns, name string) error {
			return client.CoreV1().Pods(ns).Delete(ctx, name, metav1.DeleteOptions{})
		}},
		ResourceDeployment: &resourceDeleter{deleteFn: func(ctx context.Context, ns, name string) error {
			return client.AppsV1().Deployments(ns).Delete(ctx, name, metav1.DeleteOptions{})
		}},
		ResourceStatefulSet: &resourceDeleter{deleteFn: func(ctx context.Context, ns, name string) error {
			return client.AppsV1().StatefulSets(ns).Delete(ctx, name, metav1.DeleteOptions{})
		}},
		ResourceDaemonSet: &resourceDeleter{deleteFn: func(ctx context.Context, ns, name string) error {
			return client.AppsV1().DaemonSets(ns).Delete(ctx, name, metav1.DeleteOptions{})
		}},
		ResourceService: &resourceDeleter{deleteFn: func(ctx context.Context, ns, name string) error {
			return client.CoreV1().Services(ns).Delete(ctx, name, metav1.DeleteOptions{})
		}},
		ResourceConfigMap: &resourceDeleter{deleteFn: func(ctx context.Context, ns, name string) error {
			return client.CoreV1().ConfigMaps(ns).Delete(ctx, name, metav1.DeleteOptions{})
		}},
		ResourceSecret: &resourceDeleter{deleteFn: func(ctx context.Context, ns, name string) error {
			return client.CoreV1().Secrets(ns).Delete(ctx, name, metav1.DeleteOptions{})
		}},
		ResourceIngress: &resourceDeleter{deleteFn: func(ctx context.Context, ns, name string) error {
			return client.NetworkingV1().Ingresses(ns).Delete(ctx, name, metav1.DeleteOptions{})
		}},
		ResourceJob: &resourceDeleter{deleteFn: func(ctx context.Context, ns, name string) error {
			return client.BatchV1().Jobs(ns).Delete(ctx, name, metav1.DeleteOptions{})
		}},
		ResourceCronJob: &resourceDeleter{deleteFn: func(ctx context.Context, ns, name string) error {
			return client.BatchV1().CronJobs(ns).Delete(ctx, name, metav1.DeleteOptions{})
		}},
		ResourcePVC: &resourceDeleter{deleteFn: func(ctx context.Context, ns, name string) error {
			return client.CoreV1().PersistentVolumeClaims(ns).Delete(ctx, name, metav1.DeleteOptions{})
		}},
		ResourcePV: &resourceDeleter{deleteFn: func(ctx context.Context, ns, name string) error {
			return client.CoreV1().PersistentVolumes().Delete(ctx, name, metav1.DeleteOptions{})
		}},
		ResourceStorageClass: &resourceDeleter{deleteFn: func(ctx context.Context, ns, name string) error {
			return client.StorageV1().StorageClasses().Delete(ctx, name, metav1.DeleteOptions{})
		}},
		ResourceNamespace: &resourceDeleter{deleteFn: func(ctx context.Context, ns, name string) error {
			return client.CoreV1().Namespaces().Delete(ctx, name, metav1.DeleteOptions{})
		}},
		ResourceNode: &resourceDeleter{deleteFn: func(ctx context.Context, ns, name string) error {
			return client.CoreV1().Nodes().Delete(ctx, name, metav1.DeleteOptions{})
		}},
	}
}

// Creator 实现
type podsCreator struct{ client kubernetes.Interface }

func (c *podsCreator) Create(ctx context.Context, namespace string, obj any) error {
	pod, ok := obj.(*v1.Pod)
	if !ok {
		return fmt.Errorf("invalid Pod object")
	}
	_, err := c.client.CoreV1().Pods(namespace).Create(ctx, pod, metav1.CreateOptions{})
	return err
}

type deploymentsCreator struct{ client kubernetes.Interface }

func (c *deploymentsCreator) Create(ctx context.Context, namespace string, obj any) error {
	dep, ok := obj.(*appsv1.Deployment)
	if !ok {
		return fmt.Errorf("invalid Deployment object")
	}
	_, err := c.client.AppsV1().Deployments(namespace).Create(ctx, dep, metav1.CreateOptions{})
	return err
}

type statefulSetsCreator struct{ client kubernetes.Interface }

func (c *statefulSetsCreator) Create(ctx context.Context, namespace string, obj any) error {
	sts, ok := obj.(*appsv1.StatefulSet)
	if !ok {
		return fmt.Errorf("invalid StatefulSet object")
	}
	_, err := c.client.AppsV1().StatefulSets(namespace).Create(ctx, sts, metav1.CreateOptions{})
	return err
}

type daemonSetsCreator struct{ client kubernetes.Interface }

func (c *daemonSetsCreator) Create(ctx context.Context, namespace string, obj any) error {
	ds, ok := obj.(*appsv1.DaemonSet)
	if !ok {
		return fmt.Errorf("invalid DaemonSet object")
	}
	_, err := c.client.AppsV1().DaemonSets(namespace).Create(ctx, ds, metav1.CreateOptions{})
	return err
}

type servicesCreator struct{ client kubernetes.Interface }

func (c *servicesCreator) Create(ctx context.Context, namespace string, obj any) error {
	svc, ok := obj.(*v1.Service)
	if !ok {
		return fmt.Errorf("invalid Service object")
	}
	_, err := c.client.CoreV1().Services(namespace).Create(ctx, svc, metav1.CreateOptions{})
	return err
}

type configMapsCreator struct{ client kubernetes.Interface }

func (c *configMapsCreator) Create(ctx context.Context, namespace string, obj any) error {
	cm, ok := obj.(*v1.ConfigMap)
	if !ok {
		return fmt.Errorf("invalid ConfigMap object")
	}
	_, err := c.client.CoreV1().ConfigMaps(namespace).Create(ctx, cm, metav1.CreateOptions{})
	return err
}

type secretsCreator struct{ client kubernetes.Interface }

func (c *secretsCreator) Create(ctx context.Context, namespace string, obj any) error {
	s, ok := obj.(*v1.Secret)
	if !ok {
		return fmt.Errorf("invalid Secret object")
	}
	_, err := c.client.CoreV1().Secrets(namespace).Create(ctx, s, metav1.CreateOptions{})
	return err
}

type ingressesCreator struct{ client kubernetes.Interface }

func (c *ingressesCreator) Create(ctx context.Context, namespace string, obj any) error {
	ing, ok := obj.(*networkingv1.Ingress)
	if !ok {
		return fmt.Errorf("invalid Ingress object")
	}
	_, err := c.client.NetworkingV1().Ingresses(namespace).Create(ctx, ing, metav1.CreateOptions{})
	return err
}

type jobsCreator struct{ client kubernetes.Interface }

func (c *jobsCreator) Create(ctx context.Context, namespace string, obj any) error {
	job, ok := obj.(*batchv1.Job)
	if !ok {
		return fmt.Errorf("invalid Job object")
	}
	_, err := c.client.BatchV1().Jobs(namespace).Create(ctx, job, metav1.CreateOptions{})
	return err
}

type cronJobsCreator struct{ client kubernetes.Interface }

func (c *cronJobsCreator) Create(ctx context.Context, namespace string, obj any) error {
	cj, ok := obj.(*batchv1.CronJob)
	if !ok {
		return fmt.Errorf("invalid CronJob object")
	}
	_, err := c.client.BatchV1().CronJobs(namespace).Create(ctx, cj, metav1.CreateOptions{})
	return err
}

type pvcsCreator struct{ client kubernetes.Interface }

func (c *pvcsCreator) Create(ctx context.Context, namespace string, obj any) error {
	pvc, ok := obj.(*v1.PersistentVolumeClaim)
	if !ok {
		return fmt.Errorf("invalid PVC object")
	}
	_, err := c.client.CoreV1().PersistentVolumeClaims(namespace).Create(ctx, pvc, metav1.CreateOptions{})
	return err
}

type pvsCreator struct{ client kubernetes.Interface }

func (c *pvsCreator) Create(ctx context.Context, namespace string, obj any) error {
	pv, ok := obj.(*v1.PersistentVolume)
	if !ok {
		return fmt.Errorf("invalid PV object")
	}
	_, err := c.client.CoreV1().PersistentVolumes().Create(ctx, pv, metav1.CreateOptions{})
	return err
}

type storageClassesCreator struct{ client kubernetes.Interface }

func (c *storageClassesCreator) Create(ctx context.Context, namespace string, obj any) error {
	sc, ok := obj.(*storagev1.StorageClass)
	if !ok {
		return fmt.Errorf("invalid StorageClass object")
	}
	_, err := c.client.StorageV1().StorageClasses().Create(ctx, sc, metav1.CreateOptions{})
	return err
}

type namespacesCreator struct{ client kubernetes.Interface }

func (c *namespacesCreator) Create(ctx context.Context, namespace string, obj any) error {
	ns, ok := obj.(*v1.Namespace)
	if !ok {
		return fmt.Errorf("invalid Namespace object")
	}
	_, err := c.client.CoreV1().Namespaces().Create(ctx, ns, metav1.CreateOptions{})
	return err
}

type nodesCreator struct{ client kubernetes.Interface }

func (c *nodesCreator) Create(ctx context.Context, namespace string, obj any) error {
	node, ok := obj.(*v1.Node)
	if !ok {
		return fmt.Errorf("invalid Node object")
	}
	_, err := c.client.CoreV1().Nodes().Create(ctx, node, metav1.CreateOptions{})
	return err
}

// Compile-time interface assertions
var (
	_ Getter  = &resourceGetter{}
	_ Getter  = &resourceGetter{}
	_ Updater = &resourceUpdater{}
	_ Deleter = &resourceDeleter{}
	_ Creator = &resourceCreator{}
)
