package k8s

import (
	"context"

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

type ResourceEntry struct {
	Get    func(ctx context.Context, namespace, name string) (any, error)
	List   func(ctx context.Context, namespace string, opts metav1.ListOptions) (any, error)
	Update func(ctx context.Context, namespace, name string, obj any) error
	Delete func(ctx context.Context, namespace, name string) error
	Create func(ctx context.Context, namespace string, obj any) error
	Kind   string
}

func NewRegistry(client kubernetes.Interface) map[ResourceType]*ResourceEntry {
	c := client
	return map[ResourceType]*ResourceEntry{
		ResourcePod: {
			Get: func(ctx context.Context, ns, name string) (any, error) {
				return c.CoreV1().Pods(ns).Get(ctx, name, metav1.GetOptions{})
			},
			List: func(ctx context.Context, ns string, opts metav1.ListOptions) (any, error) {
				return c.CoreV1().Pods(ns).List(ctx, opts)
			},
			Update: func(ctx context.Context, ns, name string, obj any) error {
				_, e := c.CoreV1().Pods(ns).Update(ctx, obj.(*v1.Pod), metav1.UpdateOptions{})
				return e
			},
			Delete: func(ctx context.Context, ns, name string) error {
				return c.CoreV1().Pods(ns).Delete(ctx, name, metav1.DeleteOptions{})
			},
			Create: func(ctx context.Context, ns string, obj any) error {
				_, err := c.CoreV1().Pods(ns).Create(ctx, obj.(*v1.Pod), metav1.CreateOptions{})
				return err
			},
			Kind: "Pod",
		},
		ResourceDeployment: {
			Get: func(ctx context.Context, ns, name string) (any, error) {
				return c.AppsV1().Deployments(ns).Get(ctx, name, metav1.GetOptions{})
			},
			List: func(ctx context.Context, ns string, opts metav1.ListOptions) (any, error) {
				return c.AppsV1().Deployments(ns).List(ctx, opts)
			},
			Update: func(ctx context.Context, ns, name string, obj any) error {
				_, e := c.AppsV1().Deployments(ns).Update(ctx, obj.(*appsv1.Deployment), metav1.UpdateOptions{})
				return e
			},
			Delete: func(ctx context.Context, ns, name string) error {
				return c.AppsV1().Deployments(ns).Delete(ctx, name, metav1.DeleteOptions{})
			},
			Create: func(ctx context.Context, ns string, obj any) error {
				_, err := c.AppsV1().Deployments(ns).Create(ctx, obj.(*appsv1.Deployment), metav1.CreateOptions{})
				return err
			},
			Kind: "Deployment",
		},
		ResourceStatefulSet: {
			Get: func(ctx context.Context, ns, name string) (any, error) {
				return c.AppsV1().StatefulSets(ns).Get(ctx, name, metav1.GetOptions{})
			},
			List: func(ctx context.Context, ns string, opts metav1.ListOptions) (any, error) {
				return c.AppsV1().StatefulSets(ns).List(ctx, opts)
			},
			Update: func(ctx context.Context, ns, name string, obj any) error {
				_, e := c.AppsV1().StatefulSets(ns).Update(ctx, obj.(*appsv1.StatefulSet), metav1.UpdateOptions{})
				return e
			},
			Delete: func(ctx context.Context, ns, name string) error {
				return c.AppsV1().StatefulSets(ns).Delete(ctx, name, metav1.DeleteOptions{})
			},
			Create: func(ctx context.Context, ns string, obj any) error {
				_, err := c.AppsV1().StatefulSets(ns).Create(ctx, obj.(*appsv1.StatefulSet), metav1.CreateOptions{})
				return err
			},
			Kind: "StatefulSet",
		},
		ResourceDaemonSet: {
			Get: func(ctx context.Context, ns, name string) (any, error) {
				return c.AppsV1().DaemonSets(ns).Get(ctx, name, metav1.GetOptions{})
			},
			List: func(ctx context.Context, ns string, opts metav1.ListOptions) (any, error) {
				return c.AppsV1().DaemonSets(ns).List(ctx, opts)
			},
			Update: func(ctx context.Context, ns, name string, obj any) error {
				_, e := c.AppsV1().DaemonSets(ns).Update(ctx, obj.(*appsv1.DaemonSet), metav1.UpdateOptions{})
				return e
			},
			Delete: func(ctx context.Context, ns, name string) error {
				return c.AppsV1().DaemonSets(ns).Delete(ctx, name, metav1.DeleteOptions{})
			},
			Create: func(ctx context.Context, ns string, obj any) error {
				_, err := c.AppsV1().DaemonSets(ns).Create(ctx, obj.(*appsv1.DaemonSet), metav1.CreateOptions{})
				return err
			},
			Kind: "DaemonSet",
		},
		ResourceService: {
			Get: func(ctx context.Context, ns, name string) (any, error) {
				return c.CoreV1().Services(ns).Get(ctx, name, metav1.GetOptions{})
			},
			List: func(ctx context.Context, ns string, opts metav1.ListOptions) (any, error) {
				return c.CoreV1().Services(ns).List(ctx, opts)
			},
			Update: func(ctx context.Context, ns, name string, obj any) error {
				_, e := c.CoreV1().Services(ns).Update(ctx, obj.(*v1.Service), metav1.UpdateOptions{})
				return e
			},
			Delete: func(ctx context.Context, ns, name string) error {
				return c.CoreV1().Services(ns).Delete(ctx, name, metav1.DeleteOptions{})
			},
			Create: func(ctx context.Context, ns string, obj any) error {
				_, err := c.CoreV1().Services(ns).Create(ctx, obj.(*v1.Service), metav1.CreateOptions{})
				return err
			},
			Kind: "Service",
		},
		ResourceConfigMap: {
			Get: func(ctx context.Context, ns, name string) (any, error) {
				return c.CoreV1().ConfigMaps(ns).Get(ctx, name, metav1.GetOptions{})
			},
			List: func(ctx context.Context, ns string, opts metav1.ListOptions) (any, error) {
				return c.CoreV1().ConfigMaps(ns).List(ctx, opts)
			},
			Update: func(ctx context.Context, ns, name string, obj any) error {
				_, e := c.CoreV1().ConfigMaps(ns).Update(ctx, obj.(*v1.ConfigMap), metav1.UpdateOptions{})
				return e
			},
			Delete: func(ctx context.Context, ns, name string) error {
				return c.CoreV1().ConfigMaps(ns).Delete(ctx, name, metav1.DeleteOptions{})
			},
			Create: func(ctx context.Context, ns string, obj any) error {
				_, err := c.CoreV1().ConfigMaps(ns).Create(ctx, obj.(*v1.ConfigMap), metav1.CreateOptions{})
				return err
			},
			Kind: "ConfigMap",
		},
		ResourceSecret: {
			Get: func(ctx context.Context, ns, name string) (any, error) {
				return c.CoreV1().Secrets(ns).Get(ctx, name, metav1.GetOptions{})
			},
			List: func(ctx context.Context, ns string, opts metav1.ListOptions) (any, error) {
				return c.CoreV1().Secrets(ns).List(ctx, opts)
			},
			Update: func(ctx context.Context, ns, name string, obj any) error {
				_, e := c.CoreV1().Secrets(ns).Update(ctx, obj.(*v1.Secret), metav1.UpdateOptions{})
				return e
			},
			Delete: func(ctx context.Context, ns, name string) error {
				return c.CoreV1().Secrets(ns).Delete(ctx, name, metav1.DeleteOptions{})
			},
			Create: func(ctx context.Context, ns string, obj any) error {
				_, err := c.CoreV1().Secrets(ns).Create(ctx, obj.(*v1.Secret), metav1.CreateOptions{})
				return err
			},
			Kind: "Secret",
		},
		ResourceIngress: {
			Get: func(ctx context.Context, ns, name string) (any, error) {
				return c.NetworkingV1().Ingresses(ns).Get(ctx, name, metav1.GetOptions{})
			},
			List: func(ctx context.Context, ns string, opts metav1.ListOptions) (any, error) {
				return c.NetworkingV1().Ingresses(ns).List(ctx, opts)
			},
			Update: func(ctx context.Context, ns, name string, obj any) error {
				_, e := c.NetworkingV1().Ingresses(ns).Update(ctx, obj.(*networkingv1.Ingress), metav1.UpdateOptions{})
				return e
			},
			Delete: func(ctx context.Context, ns, name string) error {
				return c.NetworkingV1().Ingresses(ns).Delete(ctx, name, metav1.DeleteOptions{})
			},
			Create: func(ctx context.Context, ns string, obj any) error {
				_, err := c.NetworkingV1().Ingresses(ns).Create(ctx, obj.(*networkingv1.Ingress), metav1.CreateOptions{})
				return err
			},
			Kind: "Ingress",
		},
		ResourceJob: {
			Get: func(ctx context.Context, ns, name string) (any, error) {
				return c.BatchV1().Jobs(ns).Get(ctx, name, metav1.GetOptions{})
			},
			List: func(ctx context.Context, ns string, opts metav1.ListOptions) (any, error) {
				return c.BatchV1().Jobs(ns).List(ctx, opts)
			},
			Update: func(ctx context.Context, ns, name string, obj any) error {
				_, e := c.BatchV1().Jobs(ns).Update(ctx, obj.(*batchv1.Job), metav1.UpdateOptions{})
				return e
			},
			Delete: func(ctx context.Context, ns, name string) error {
				return c.BatchV1().Jobs(ns).Delete(ctx, name, metav1.DeleteOptions{})
			},
			Create: func(ctx context.Context, ns string, obj any) error {
				_, err := c.BatchV1().Jobs(ns).Create(ctx, obj.(*batchv1.Job), metav1.CreateOptions{})
				return err
			},
			Kind: "Job",
		},
		ResourceCronJob: {
			Get: func(ctx context.Context, ns, name string) (any, error) {
				return c.BatchV1().CronJobs(ns).Get(ctx, name, metav1.GetOptions{})
			},
			List: func(ctx context.Context, ns string, opts metav1.ListOptions) (any, error) {
				return c.BatchV1().CronJobs(ns).List(ctx, opts)
			},
			Update: func(ctx context.Context, ns, name string, obj any) error {
				_, e := c.BatchV1().CronJobs(ns).Update(ctx, obj.(*batchv1.CronJob), metav1.UpdateOptions{})
				return e
			},
			Delete: func(ctx context.Context, ns, name string) error {
				return c.BatchV1().CronJobs(ns).Delete(ctx, name, metav1.DeleteOptions{})
			},
			Create: func(ctx context.Context, ns string, obj any) error {
				_, err := c.BatchV1().CronJobs(ns).Create(ctx, obj.(*batchv1.CronJob), metav1.CreateOptions{})
				return err
			},
			Kind: "CronJob",
		},
		ResourcePVC: {
			Get: func(ctx context.Context, ns, name string) (any, error) {
				return c.CoreV1().PersistentVolumeClaims(ns).Get(ctx, name, metav1.GetOptions{})
			},
			List: func(ctx context.Context, ns string, opts metav1.ListOptions) (any, error) {
				return c.CoreV1().PersistentVolumeClaims(ns).List(ctx, opts)
			},
			Update: func(ctx context.Context, ns, name string, obj any) error {
				_, e := c.CoreV1().PersistentVolumeClaims(ns).Update(ctx, obj.(*v1.PersistentVolumeClaim), metav1.UpdateOptions{})
				return e
			},
			Delete: func(ctx context.Context, ns, name string) error {
				return c.CoreV1().PersistentVolumeClaims(ns).Delete(ctx, name, metav1.DeleteOptions{})
			},
			Create: func(ctx context.Context, ns string, obj any) error {
				_, err := c.CoreV1().PersistentVolumeClaims(ns).Create(ctx, obj.(*v1.PersistentVolumeClaim), metav1.CreateOptions{})
				return err
			},
			Kind: "PersistentVolumeClaim",
		},
		ResourcePV: {
			Get: func(ctx context.Context, ns, name string) (any, error) {
				return c.CoreV1().PersistentVolumes().Get(ctx, name, metav1.GetOptions{})
			},
			List: func(ctx context.Context, ns string, opts metav1.ListOptions) (any, error) {
				return c.CoreV1().PersistentVolumes().List(ctx, opts)
			},
			Update: func(ctx context.Context, ns, name string, obj any) error {
				_, e := c.CoreV1().PersistentVolumes().Update(ctx, obj.(*v1.PersistentVolume), metav1.UpdateOptions{})
				return e
			},
			Delete: func(ctx context.Context, ns, name string) error {
				return c.CoreV1().PersistentVolumes().Delete(ctx, name, metav1.DeleteOptions{})
			},
			Create: func(ctx context.Context, ns string, obj any) error {
				_, err := c.CoreV1().PersistentVolumes().Create(ctx, obj.(*v1.PersistentVolume), metav1.CreateOptions{})
				return err
			},
			Kind: "PersistentVolume",
		},
		ResourceStorageClass: {
			Get: func(ctx context.Context, ns, name string) (any, error) {
				return c.StorageV1().StorageClasses().Get(ctx, name, metav1.GetOptions{})
			},
			List: func(ctx context.Context, ns string, opts metav1.ListOptions) (any, error) {
				return c.StorageV1().StorageClasses().List(ctx, opts)
			},
			Update: func(ctx context.Context, ns, name string, obj any) error {
				_, e := c.StorageV1().StorageClasses().Update(ctx, obj.(*storagev1.StorageClass), metav1.UpdateOptions{})
				return e
			},
			Delete: func(ctx context.Context, ns, name string) error {
				return c.StorageV1().StorageClasses().Delete(ctx, name, metav1.DeleteOptions{})
			},
			Create: func(ctx context.Context, ns string, obj any) error {
				_, err := c.StorageV1().StorageClasses().Create(ctx, obj.(*storagev1.StorageClass), metav1.CreateOptions{})
				return err
			},
			Kind: "StorageClass",
		},
		ResourceNamespace: {
			Get: func(ctx context.Context, ns, name string) (any, error) {
				return c.CoreV1().Namespaces().Get(ctx, name, metav1.GetOptions{})
			},
			List: func(ctx context.Context, ns string, opts metav1.ListOptions) (any, error) {
				return c.CoreV1().Namespaces().List(ctx, opts)
			},
			Update: func(ctx context.Context, ns, name string, obj any) error {
				_, e := c.CoreV1().Namespaces().Update(ctx, obj.(*v1.Namespace), metav1.UpdateOptions{})
				return e
			},
			Delete: func(ctx context.Context, ns, name string) error {
				return c.CoreV1().Namespaces().Delete(ctx, name, metav1.DeleteOptions{})
			},
			Create: func(ctx context.Context, ns string, obj any) error {
				_, err := c.CoreV1().Namespaces().Create(ctx, obj.(*v1.Namespace), metav1.CreateOptions{})
				return err
			},
			Kind: "Namespace",
		},
		ResourceNode: {
			Get: func(ctx context.Context, ns, name string) (any, error) {
				return c.CoreV1().Nodes().Get(ctx, name, metav1.GetOptions{})
			},
			List: func(ctx context.Context, ns string, opts metav1.ListOptions) (any, error) {
				return c.CoreV1().Nodes().List(ctx, opts)
			},
			Update: func(ctx context.Context, ns, name string, obj any) error {
				_, e := c.CoreV1().Nodes().Update(ctx, obj.(*v1.Node), metav1.UpdateOptions{})
				return e
			},
			Delete: func(ctx context.Context, ns, name string) error {
				return c.CoreV1().Nodes().Delete(ctx, name, metav1.DeleteOptions{})
			},
			Create: func(ctx context.Context, ns string, obj any) error {
				_, err := c.CoreV1().Nodes().Create(ctx, obj.(*v1.Node), metav1.CreateOptions{})
				return err
			},
			Kind: "Node",
		},
		ResourceHorizontalPodAutoscaler: {
			Get: func(ctx context.Context, ns, name string) (any, error) {
				return c.AutoscalingV2().HorizontalPodAutoscalers(ns).Get(ctx, name, metav1.GetOptions{})
			},
			List: func(ctx context.Context, ns string, opts metav1.ListOptions) (any, error) {
				return c.AutoscalingV2().HorizontalPodAutoscalers(ns).List(ctx, opts)
			},
			Update: func(ctx context.Context, ns, name string, obj any) error {
				_, e := c.AutoscalingV2().HorizontalPodAutoscalers(ns).Update(ctx, obj.(*autoscalingv2.HorizontalPodAutoscaler), metav1.UpdateOptions{})
				return e
			},
			Delete: func(ctx context.Context, ns, name string) error {
				return c.AutoscalingV2().HorizontalPodAutoscalers(ns).Delete(ctx, name, metav1.DeleteOptions{})
			},
			Create: func(ctx context.Context, ns string, obj any) error {
				_, err := c.AutoscalingV2().HorizontalPodAutoscalers(ns).Create(ctx, obj.(*autoscalingv2.HorizontalPodAutoscaler), metav1.CreateOptions{})
				return err
			},
			Kind: "HorizontalPodAutoscaler",
		},
		ResourceNetworkPolicy: {
			Get: func(ctx context.Context, ns, name string) (any, error) {
				return c.NetworkingV1().NetworkPolicies(ns).Get(ctx, name, metav1.GetOptions{})
			},
			List: func(ctx context.Context, ns string, opts metav1.ListOptions) (any, error) {
				return c.NetworkingV1().NetworkPolicies(ns).List(ctx, opts)
			},
			Update: func(ctx context.Context, ns, name string, obj any) error {
				_, e := c.NetworkingV1().NetworkPolicies(ns).Update(ctx, obj.(*networkingv1.NetworkPolicy), metav1.UpdateOptions{})
				return e
			},
			Delete: func(ctx context.Context, ns, name string) error {
				return c.NetworkingV1().NetworkPolicies(ns).Delete(ctx, name, metav1.DeleteOptions{})
			},
			Create: func(ctx context.Context, ns string, obj any) error {
				_, err := c.NetworkingV1().NetworkPolicies(ns).Create(ctx, obj.(*networkingv1.NetworkPolicy), metav1.CreateOptions{})
				return err
			},
			Kind: "NetworkPolicy",
		},
		ResourceServiceAccount: {
			Get: func(ctx context.Context, ns, name string) (any, error) {
				return c.CoreV1().ServiceAccounts(ns).Get(ctx, name, metav1.GetOptions{})
			},
			List: func(ctx context.Context, ns string, opts metav1.ListOptions) (any, error) {
				return c.CoreV1().ServiceAccounts(ns).List(ctx, opts)
			},
			Update: func(ctx context.Context, ns, name string, obj any) error {
				_, e := c.CoreV1().ServiceAccounts(ns).Update(ctx, obj.(*v1.ServiceAccount), metav1.UpdateOptions{})
				return e
			},
			Delete: func(ctx context.Context, ns, name string) error {
				return c.CoreV1().ServiceAccounts(ns).Delete(ctx, name, metav1.DeleteOptions{})
			},
			Create: func(ctx context.Context, ns string, obj any) error {
				_, err := c.CoreV1().ServiceAccounts(ns).Create(ctx, obj.(*v1.ServiceAccount), metav1.CreateOptions{})
				return err
			},
			Kind: "ServiceAccount",
		},
		ResourceRole: {
			Get: func(ctx context.Context, ns, name string) (any, error) {
				return c.RbacV1().Roles(ns).Get(ctx, name, metav1.GetOptions{})
			},
			List: func(ctx context.Context, ns string, opts metav1.ListOptions) (any, error) {
				return c.RbacV1().Roles(ns).List(ctx, opts)
			},
			Update: func(ctx context.Context, ns, name string, obj any) error {
				_, e := c.RbacV1().Roles(ns).Update(ctx, obj.(*rbacv1.Role), metav1.UpdateOptions{})
				return e
			},
			Delete: func(ctx context.Context, ns, name string) error {
				return c.RbacV1().Roles(ns).Delete(ctx, name, metav1.DeleteOptions{})
			},
			Create: func(ctx context.Context, ns string, obj any) error {
				_, err := c.RbacV1().Roles(ns).Create(ctx, obj.(*rbacv1.Role), metav1.CreateOptions{})
				return err
			},
			Kind: "Role",
		},
		ResourceRoleBinding: {
			Get: func(ctx context.Context, ns, name string) (any, error) {
				return c.RbacV1().RoleBindings(ns).Get(ctx, name, metav1.GetOptions{})
			},
			List: func(ctx context.Context, ns string, opts metav1.ListOptions) (any, error) {
				return c.RbacV1().RoleBindings(ns).List(ctx, opts)
			},
			Update: func(ctx context.Context, ns, name string, obj any) error {
				_, e := c.RbacV1().RoleBindings(ns).Update(ctx, obj.(*rbacv1.RoleBinding), metav1.UpdateOptions{})
				return e
			},
			Delete: func(ctx context.Context, ns, name string) error {
				return c.RbacV1().RoleBindings(ns).Delete(ctx, name, metav1.DeleteOptions{})
			},
			Create: func(ctx context.Context, ns string, obj any) error {
				_, err := c.RbacV1().RoleBindings(ns).Create(ctx, obj.(*rbacv1.RoleBinding), metav1.CreateOptions{})
				return err
			},
			Kind: "RoleBinding",
		},
		ResourceClusterRole: {
			Get: func(ctx context.Context, ns, name string) (any, error) {
				return c.RbacV1().ClusterRoles().Get(ctx, name, metav1.GetOptions{})
			},
			List: func(ctx context.Context, ns string, opts metav1.ListOptions) (any, error) {
				return c.RbacV1().ClusterRoles().List(ctx, opts)
			},
			Update: func(ctx context.Context, ns, name string, obj any) error {
				_, e := c.RbacV1().ClusterRoles().Update(ctx, obj.(*rbacv1.ClusterRole), metav1.UpdateOptions{})
				return e
			},
			Delete: func(ctx context.Context, ns, name string) error {
				return c.RbacV1().ClusterRoles().Delete(ctx, name, metav1.DeleteOptions{})
			},
			Create: func(ctx context.Context, ns string, obj any) error {
				_, err := c.RbacV1().ClusterRoles().Create(ctx, obj.(*rbacv1.ClusterRole), metav1.CreateOptions{})
				return err
			},
			Kind: "ClusterRole",
		},
		ResourceClusterRoleBinding: {
			Get: func(ctx context.Context, ns, name string) (any, error) {
				return c.RbacV1().ClusterRoleBindings().Get(ctx, name, metav1.GetOptions{})
			},
			List: func(ctx context.Context, ns string, opts metav1.ListOptions) (any, error) {
				return c.RbacV1().ClusterRoleBindings().List(ctx, opts)
			},
			Update: func(ctx context.Context, ns, name string, obj any) error {
				_, e := c.RbacV1().ClusterRoleBindings().Update(ctx, obj.(*rbacv1.ClusterRoleBinding), metav1.UpdateOptions{})
				return e
			},
			Delete: func(ctx context.Context, ns, name string) error {
				return c.RbacV1().ClusterRoleBindings().Delete(ctx, name, metav1.DeleteOptions{})
			},
			Create: func(ctx context.Context, ns string, obj any) error {
				_, err := c.RbacV1().ClusterRoleBindings().Create(ctx, obj.(*rbacv1.ClusterRoleBinding), metav1.CreateOptions{})
				return err
			},
			Kind: "ClusterRoleBinding",
		},
		ResourceResourceQuota: {
			Get: func(ctx context.Context, ns, name string) (any, error) {
				return c.CoreV1().ResourceQuotas(ns).Get(ctx, name, metav1.GetOptions{})
			},
			List: func(ctx context.Context, ns string, opts metav1.ListOptions) (any, error) {
				return c.CoreV1().ResourceQuotas(ns).List(ctx, opts)
			},
			Update: func(ctx context.Context, ns, name string, obj any) error {
				_, e := c.CoreV1().ResourceQuotas(ns).Update(ctx, obj.(*v1.ResourceQuota), metav1.UpdateOptions{})
				return e
			},
			Delete: func(ctx context.Context, ns, name string) error {
				return c.CoreV1().ResourceQuotas(ns).Delete(ctx, name, metav1.DeleteOptions{})
			},
			Create: func(ctx context.Context, ns string, obj any) error {
				_, err := c.CoreV1().ResourceQuotas(ns).Create(ctx, obj.(*v1.ResourceQuota), metav1.CreateOptions{})
				return err
			},
			Kind: "ResourceQuota",
		},
		ResourceLimitRange: {
			Get: func(ctx context.Context, ns, name string) (any, error) {
				return c.CoreV1().LimitRanges(ns).Get(ctx, name, metav1.GetOptions{})
			},
			List: func(ctx context.Context, ns string, opts metav1.ListOptions) (any, error) {
				return c.CoreV1().LimitRanges(ns).List(ctx, opts)
			},
			Update: func(ctx context.Context, ns, name string, obj any) error {
				_, e := c.CoreV1().LimitRanges(ns).Update(ctx, obj.(*v1.LimitRange), metav1.UpdateOptions{})
				return e
			},
			Delete: func(ctx context.Context, ns, name string) error {
				return c.CoreV1().LimitRanges(ns).Delete(ctx, name, metav1.DeleteOptions{})
			},
			Create: func(ctx context.Context, ns string, obj any) error {
				_, err := c.CoreV1().LimitRanges(ns).Create(ctx, obj.(*v1.LimitRange), metav1.CreateOptions{})
				return err
			},
			Kind: "LimitRange",
		},
		ResourcePodDisruptionBudget: {
			Get: func(ctx context.Context, ns, name string) (any, error) {
				return c.PolicyV1().PodDisruptionBudgets(ns).Get(ctx, name, metav1.GetOptions{})
			},
			List: func(ctx context.Context, ns string, opts metav1.ListOptions) (any, error) {
				return c.PolicyV1().PodDisruptionBudgets(ns).List(ctx, opts)
			},
			Update: func(ctx context.Context, ns, name string, obj any) error {
				_, e := c.PolicyV1().PodDisruptionBudgets(ns).Update(ctx, obj.(*policyv1.PodDisruptionBudget), metav1.UpdateOptions{})
				return e
			},
			Delete: func(ctx context.Context, ns, name string) error {
				return c.PolicyV1().PodDisruptionBudgets(ns).Delete(ctx, name, metav1.DeleteOptions{})
			},
			Create: func(ctx context.Context, ns string, obj any) error {
				_, err := c.PolicyV1().PodDisruptionBudgets(ns).Create(ctx, obj.(*policyv1.PodDisruptionBudget), metav1.CreateOptions{})
				return err
			},
			Kind: "PodDisruptionBudget",
		},
	}
}
func GetKindByResourceType(resourceType string) string {
	rt := ResourceType(resourceType).Normalize()
	reg := map[ResourceType]string{
		ResourcePod:                     "Pod",
		ResourceDeployment:              "Deployment",
		ResourceStatefulSet:             "StatefulSet",
		ResourceDaemonSet:               "DaemonSet",
		ResourceService:                 "Service",
		ResourceConfigMap:               "ConfigMap",
		ResourceSecret:                  "Secret",
		ResourceIngress:                 "Ingress",
		ResourceJob:                     "Job",
		ResourceCronJob:                 "CronJob",
		ResourcePVC:                     "PersistentVolumeClaim",
		ResourcePV:                      "PersistentVolume",
		ResourceStorageClass:            "StorageClass",
		ResourceNamespace:               "Namespace",
		ResourceNode:                    "Node",
		ResourceEndpoint:                "Endpoint",
		ResourceEvent:                   "Event",
		ResourceHorizontalPodAutoscaler: "HorizontalPodAutoscaler",
		ResourceNetworkPolicy:           "NetworkPolicy",
		ResourceServiceAccount:          "ServiceAccount",
		ResourceRole:                    "Role",
		ResourceRoleBinding:             "RoleBinding",
		ResourceClusterRole:             "ClusterRole",
		ResourceClusterRoleBinding:      "ClusterRoleBinding",
		ResourceResourceQuota:           "ResourceQuota",
		ResourceLimitRange:              "LimitRange",
		ResourcePodDisruptionBudget:     "PodDisruptionBudget",
	}
	return reg[rt]
}
