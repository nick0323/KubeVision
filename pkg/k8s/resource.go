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
	Get(ctx context.Context, namespace, name string) (interface{}, error)
	List(ctx context.Context, namespace string, opts metav1.ListOptions) (interface{}, error)
}

type Lister interface {
	List(ctx context.Context, namespace string, opts metav1.ListOptions) (interface{}, error)
}

type Updater interface {
	Update(ctx context.Context, namespace, name string, obj interface{}) error
}

type Deleter interface {
	Delete(ctx context.Context, namespace, name string) error
}

type Creator interface {
	Create(ctx context.Context, namespace string, obj interface{}) error
}

type podsGetter struct{ client kubernetes.Interface }

func (g *podsGetter) Get(ctx context.Context, namespace, name string) (interface{}, error) {
	return g.client.CoreV1().Pods(namespace).Get(ctx, name, metav1.GetOptions{})
}

func (g *podsGetter) List(ctx context.Context, namespace string, opts metav1.ListOptions) (interface{}, error) {
	return g.client.CoreV1().Pods(namespace).List(ctx, opts)
}

type deploymentsGetter struct{ client kubernetes.Interface }

func (g *deploymentsGetter) Get(ctx context.Context, namespace, name string) (interface{}, error) {
	return g.client.AppsV1().Deployments(namespace).Get(ctx, name, metav1.GetOptions{})
}

func (g *deploymentsGetter) List(ctx context.Context, namespace string, opts metav1.ListOptions) (interface{}, error) {
	return g.client.AppsV1().Deployments(namespace).List(ctx, opts)
}

type statefulSetsGetter struct{ client kubernetes.Interface }

func (g *statefulSetsGetter) Get(ctx context.Context, namespace, name string) (interface{}, error) {
	return g.client.AppsV1().StatefulSets(namespace).Get(ctx, name, metav1.GetOptions{})
}

func (g *statefulSetsGetter) List(ctx context.Context, namespace string, opts metav1.ListOptions) (interface{}, error) {
	return g.client.AppsV1().StatefulSets(namespace).List(ctx, opts)
}

type daemonSetsGetter struct{ client kubernetes.Interface }

func (g *daemonSetsGetter) Get(ctx context.Context, namespace, name string) (interface{}, error) {
	return g.client.AppsV1().DaemonSets(namespace).Get(ctx, name, metav1.GetOptions{})
}

func (g *daemonSetsGetter) List(ctx context.Context, namespace string, opts metav1.ListOptions) (interface{}, error) {
	return g.client.AppsV1().DaemonSets(namespace).List(ctx, opts)
}

type servicesGetter struct{ client kubernetes.Interface }

func (g *servicesGetter) Get(ctx context.Context, namespace, name string) (interface{}, error) {
	return g.client.CoreV1().Services(namespace).Get(ctx, name, metav1.GetOptions{})
}

func (g *servicesGetter) List(ctx context.Context, namespace string, opts metav1.ListOptions) (interface{}, error) {
	return g.client.CoreV1().Services(namespace).List(ctx, opts)
}

type configMapsGetter struct{ client kubernetes.Interface }

func (g *configMapsGetter) Get(ctx context.Context, namespace, name string) (interface{}, error) {
	return g.client.CoreV1().ConfigMaps(namespace).Get(ctx, name, metav1.GetOptions{})
}

func (g *configMapsGetter) List(ctx context.Context, namespace string, opts metav1.ListOptions) (interface{}, error) {
	return g.client.CoreV1().ConfigMaps(namespace).List(ctx, opts)
}

type secretsGetter struct{ client kubernetes.Interface }

func (g *secretsGetter) Get(ctx context.Context, namespace, name string) (interface{}, error) {
	return g.client.CoreV1().Secrets(namespace).Get(ctx, name, metav1.GetOptions{})
}

func (g *secretsGetter) List(ctx context.Context, namespace string, opts metav1.ListOptions) (interface{}, error) {
	return g.client.CoreV1().Secrets(namespace).List(ctx, opts)
}

type ingressesGetter struct{ client kubernetes.Interface }

func (g *ingressesGetter) Get(ctx context.Context, namespace, name string) (interface{}, error) {
	return g.client.NetworkingV1().Ingresses(namespace).Get(ctx, name, metav1.GetOptions{})
}

func (g *ingressesGetter) List(ctx context.Context, namespace string, opts metav1.ListOptions) (interface{}, error) {
	return g.client.NetworkingV1().Ingresses(namespace).List(ctx, opts)
}

type jobsGetter struct{ client kubernetes.Interface }

func (g *jobsGetter) Get(ctx context.Context, namespace, name string) (interface{}, error) {
	return g.client.BatchV1().Jobs(namespace).Get(ctx, name, metav1.GetOptions{})
}

func (g *jobsGetter) List(ctx context.Context, namespace string, opts metav1.ListOptions) (interface{}, error) {
	return g.client.BatchV1().Jobs(namespace).List(ctx, opts)
}

type cronJobsGetter struct{ client kubernetes.Interface }

func (g *cronJobsGetter) Get(ctx context.Context, namespace, name string) (interface{}, error) {
	return g.client.BatchV1().CronJobs(namespace).Get(ctx, name, metav1.GetOptions{})
}

func (g *cronJobsGetter) List(ctx context.Context, namespace string, opts metav1.ListOptions) (interface{}, error) {
	return g.client.BatchV1().CronJobs(namespace).List(ctx, opts)
}

type pvcsGetter struct{ client kubernetes.Interface }

func (g *pvcsGetter) Get(ctx context.Context, namespace, name string) (interface{}, error) {
	return g.client.CoreV1().PersistentVolumeClaims(namespace).Get(ctx, name, metav1.GetOptions{})
}

func (g *pvcsGetter) List(ctx context.Context, namespace string, opts metav1.ListOptions) (interface{}, error) {
	return g.client.CoreV1().PersistentVolumeClaims(namespace).List(ctx, opts)
}

type pvsGetter struct{ client kubernetes.Interface }

func (g *pvsGetter) Get(ctx context.Context, namespace, name string) (interface{}, error) {
	return g.client.CoreV1().PersistentVolumes().Get(ctx, name, metav1.GetOptions{})
}

func (g *pvsGetter) List(ctx context.Context, namespace string, opts metav1.ListOptions) (interface{}, error) {
	return g.client.CoreV1().PersistentVolumes().List(ctx, opts)
}

type storageClassesGetter struct{ client kubernetes.Interface }

func (g *storageClassesGetter) Get(ctx context.Context, namespace, name string) (interface{}, error) {
	return g.client.StorageV1().StorageClasses().Get(ctx, name, metav1.GetOptions{})
}

func (g *storageClassesGetter) List(ctx context.Context, namespace string, opts metav1.ListOptions) (interface{}, error) {
	return g.client.StorageV1().StorageClasses().List(ctx, opts)
}

type namespacesGetter struct{ client kubernetes.Interface }

func (g *namespacesGetter) Get(ctx context.Context, namespace, name string) (interface{}, error) {
	return g.client.CoreV1().Namespaces().Get(ctx, name, metav1.GetOptions{})
}

func (g *namespacesGetter) List(ctx context.Context, namespace string, opts metav1.ListOptions) (interface{}, error) {
	return g.client.CoreV1().Namespaces().List(ctx, opts)
}

type nodesGetter struct{ client kubernetes.Interface }

func (g *nodesGetter) Get(ctx context.Context, namespace, name string) (interface{}, error) {
	return g.client.CoreV1().Nodes().Get(ctx, name, metav1.GetOptions{})
}

func (g *nodesGetter) List(ctx context.Context, namespace string, opts metav1.ListOptions) (interface{}, error) {
	return g.client.CoreV1().Nodes().List(ctx, opts)
}

type endpointsGetter struct{ client kubernetes.Interface }

func (g *endpointsGetter) Get(ctx context.Context, namespace, name string) (interface{}, error) {
	return g.client.CoreV1().Endpoints(namespace).Get(ctx, name, metav1.GetOptions{})
}

func (g *endpointsGetter) List(ctx context.Context, namespace string, opts metav1.ListOptions) (interface{}, error) {
	return g.client.CoreV1().Endpoints(namespace).List(ctx, opts)
}

type hpasGetter struct{ client kubernetes.Interface }

func (g *hpasGetter) Get(ctx context.Context, namespace, name string) (interface{}, error) {
	return g.client.AutoscalingV2().HorizontalPodAutoscalers(namespace).Get(ctx, name, metav1.GetOptions{})
}

func (g *hpasGetter) List(ctx context.Context, namespace string, opts metav1.ListOptions) (interface{}, error) {
	return g.client.AutoscalingV2().HorizontalPodAutoscalers(namespace).List(ctx, opts)
}

type networkPoliciesGetter struct{ client kubernetes.Interface }

func (g *networkPoliciesGetter) Get(ctx context.Context, namespace, name string) (interface{}, error) {
	return g.client.NetworkingV1().NetworkPolicies(namespace).Get(ctx, name, metav1.GetOptions{})
}

func (g *networkPoliciesGetter) List(ctx context.Context, namespace string, opts metav1.ListOptions) (interface{}, error) {
	return g.client.NetworkingV1().NetworkPolicies(namespace).List(ctx, opts)
}

type serviceAccountsGetter struct{ client kubernetes.Interface }

func (g *serviceAccountsGetter) Get(ctx context.Context, namespace, name string) (interface{}, error) {
	return g.client.CoreV1().ServiceAccounts(namespace).Get(ctx, name, metav1.GetOptions{})
}

func (g *serviceAccountsGetter) List(ctx context.Context, namespace string, opts metav1.ListOptions) (interface{}, error) {
	return g.client.CoreV1().ServiceAccounts(namespace).List(ctx, opts)
}

type rolesGetter struct{ client kubernetes.Interface }

func (g *rolesGetter) Get(ctx context.Context, namespace, name string) (interface{}, error) {
	return g.client.RbacV1().Roles(namespace).Get(ctx, name, metav1.GetOptions{})
}

func (g *rolesGetter) List(ctx context.Context, namespace string, opts metav1.ListOptions) (interface{}, error) {
	return g.client.RbacV1().Roles(namespace).List(ctx, opts)
}

type roleBindingsGetter struct{ client kubernetes.Interface }

func (g *roleBindingsGetter) Get(ctx context.Context, namespace, name string) (interface{}, error) {
	return g.client.RbacV1().RoleBindings(namespace).Get(ctx, name, metav1.GetOptions{})
}

func (g *roleBindingsGetter) List(ctx context.Context, namespace string, opts metav1.ListOptions) (interface{}, error) {
	return g.client.RbacV1().RoleBindings(namespace).List(ctx, opts)
}

type clusterRolesGetter struct{ client kubernetes.Interface }

func (g *clusterRolesGetter) Get(ctx context.Context, namespace, name string) (interface{}, error) {
	return g.client.RbacV1().ClusterRoles().Get(ctx, name, metav1.GetOptions{})
}

func (g *clusterRolesGetter) List(ctx context.Context, namespace string, opts metav1.ListOptions) (interface{}, error) {
	return g.client.RbacV1().ClusterRoles().List(ctx, opts)
}

type clusterRoleBindingsGetter struct{ client kubernetes.Interface }

func (g *clusterRoleBindingsGetter) Get(ctx context.Context, namespace, name string) (interface{}, error) {
	return g.client.RbacV1().ClusterRoleBindings().Get(ctx, name, metav1.GetOptions{})
}

func (g *clusterRoleBindingsGetter) List(ctx context.Context, namespace string, opts metav1.ListOptions) (interface{}, error) {
	return g.client.RbacV1().ClusterRoleBindings().List(ctx, opts)
}

type resourceQuotasGetter struct{ client kubernetes.Interface }

func (g *resourceQuotasGetter) Get(ctx context.Context, namespace, name string) (interface{}, error) {
	return g.client.CoreV1().ResourceQuotas(namespace).Get(ctx, name, metav1.GetOptions{})
}

func (g *resourceQuotasGetter) List(ctx context.Context, namespace string, opts metav1.ListOptions) (interface{}, error) {
	return g.client.CoreV1().ResourceQuotas(namespace).List(ctx, opts)
}

type limitRangesGetter struct{ client kubernetes.Interface }

func (g *limitRangesGetter) Get(ctx context.Context, namespace, name string) (interface{}, error) {
	return g.client.CoreV1().LimitRanges(namespace).Get(ctx, name, metav1.GetOptions{})
}

func (g *limitRangesGetter) List(ctx context.Context, namespace string, opts metav1.ListOptions) (interface{}, error) {
	return g.client.CoreV1().LimitRanges(namespace).List(ctx, opts)
}

type podDisruptionBudgetsGetter struct{ client kubernetes.Interface }

func (g *podDisruptionBudgetsGetter) Get(ctx context.Context, namespace, name string) (interface{}, error) {
	return g.client.PolicyV1().PodDisruptionBudgets(namespace).Get(ctx, name, metav1.GetOptions{})
}

func (g *podDisruptionBudgetsGetter) List(ctx context.Context, namespace string, opts metav1.ListOptions) (interface{}, error) {
	return g.client.PolicyV1().PodDisruptionBudgets(namespace).List(ctx, opts)
}

type hpasUpdater struct{ client kubernetes.Interface }

func (u *hpasUpdater) Update(ctx context.Context, namespace, name string, obj interface{}) error {
	hpa, ok := obj.(*autoscalingv2.HorizontalPodAutoscaler)
	if !ok {
		return fmt.Errorf("invalid HorizontalPodAutoscaler object")
	}
	_, err := u.client.AutoscalingV2().HorizontalPodAutoscalers(namespace).Update(ctx, hpa, metav1.UpdateOptions{})
	return err
}

func (u *hpasUpdater) Delete(ctx context.Context, namespace, name string) error {
	return u.client.AutoscalingV2().HorizontalPodAutoscalers(namespace).Delete(ctx, name, metav1.DeleteOptions{})
}

func (u *hpasUpdater) Create(ctx context.Context, namespace string, obj interface{}) error {
	hpa, ok := obj.(*autoscalingv2.HorizontalPodAutoscaler)
	if !ok {
		return fmt.Errorf("invalid HorizontalPodAutoscaler object")
	}
	_, err := u.client.AutoscalingV2().HorizontalPodAutoscalers(namespace).Create(ctx, hpa, metav1.CreateOptions{})
	return err
}

type networkPoliciesUpdater struct{ client kubernetes.Interface }

func (u *networkPoliciesUpdater) Update(ctx context.Context, namespace, name string, obj interface{}) error {
	np, ok := obj.(*networkingv1.NetworkPolicy)
	if !ok {
		return fmt.Errorf("invalid NetworkPolicy object")
	}
	_, err := u.client.NetworkingV1().NetworkPolicies(namespace).Update(ctx, np, metav1.UpdateOptions{})
	return err
}

func (u *networkPoliciesUpdater) Delete(ctx context.Context, namespace, name string) error {
	return u.client.NetworkingV1().NetworkPolicies(namespace).Delete(ctx, name, metav1.DeleteOptions{})
}

func (u *networkPoliciesUpdater) Create(ctx context.Context, namespace string, obj interface{}) error {
	np, ok := obj.(*networkingv1.NetworkPolicy)
	if !ok {
		return fmt.Errorf("invalid NetworkPolicy object")
	}
	_, err := u.client.NetworkingV1().NetworkPolicies(namespace).Create(ctx, np, metav1.CreateOptions{})
	return err
}

type serviceAccountsUpdater struct{ client kubernetes.Interface }

func (u *serviceAccountsUpdater) Update(ctx context.Context, namespace, name string, obj interface{}) error {
	sa, ok := obj.(*v1.ServiceAccount)
	if !ok {
		return fmt.Errorf("invalid ServiceAccount object")
	}
	_, err := u.client.CoreV1().ServiceAccounts(namespace).Update(ctx, sa, metav1.UpdateOptions{})
	return err
}

func (u *serviceAccountsUpdater) Delete(ctx context.Context, namespace, name string) error {
	return u.client.CoreV1().ServiceAccounts(namespace).Delete(ctx, name, metav1.DeleteOptions{})
}

func (u *serviceAccountsUpdater) Create(ctx context.Context, namespace string, obj interface{}) error {
	sa, ok := obj.(*v1.ServiceAccount)
	if !ok {
		return fmt.Errorf("invalid ServiceAccount object")
	}
	_, err := u.client.CoreV1().ServiceAccounts(namespace).Create(ctx, sa, metav1.CreateOptions{})
	return err
}

type rolesUpdater struct{ client kubernetes.Interface }

func (u *rolesUpdater) Update(ctx context.Context, namespace, name string, obj interface{}) error {
	role, ok := obj.(*rbacv1.Role)
	if !ok {
		return fmt.Errorf("invalid Role object")
	}
	_, err := u.client.RbacV1().Roles(namespace).Update(ctx, role, metav1.UpdateOptions{})
	return err
}

func (u *rolesUpdater) Delete(ctx context.Context, namespace, name string) error {
	return u.client.RbacV1().Roles(namespace).Delete(ctx, name, metav1.DeleteOptions{})
}

func (u *rolesUpdater) Create(ctx context.Context, namespace string, obj interface{}) error {
	role, ok := obj.(*rbacv1.Role)
	if !ok {
		return fmt.Errorf("invalid Role object")
	}
	_, err := u.client.RbacV1().Roles(namespace).Create(ctx, role, metav1.CreateOptions{})
	return err
}

type roleBindingsUpdater struct{ client kubernetes.Interface }

func (u *roleBindingsUpdater) Update(ctx context.Context, namespace, name string, obj interface{}) error {
	rb, ok := obj.(*rbacv1.RoleBinding)
	if !ok {
		return fmt.Errorf("invalid RoleBinding object")
	}
	_, err := u.client.RbacV1().RoleBindings(namespace).Update(ctx, rb, metav1.UpdateOptions{})
	return err
}

func (u *roleBindingsUpdater) Delete(ctx context.Context, namespace, name string) error {
	return u.client.RbacV1().RoleBindings(namespace).Delete(ctx, name, metav1.DeleteOptions{})
}

func (u *roleBindingsUpdater) Create(ctx context.Context, namespace string, obj interface{}) error {
	rb, ok := obj.(*rbacv1.RoleBinding)
	if !ok {
		return fmt.Errorf("invalid RoleBinding object")
	}
	_, err := u.client.RbacV1().RoleBindings(namespace).Create(ctx, rb, metav1.CreateOptions{})
	return err
}

type clusterRolesUpdater struct{ client kubernetes.Interface }

func (u *clusterRolesUpdater) Update(ctx context.Context, namespace, name string, obj interface{}) error {
	cr, ok := obj.(*rbacv1.ClusterRole)
	if !ok {
		return fmt.Errorf("invalid ClusterRole object")
	}
	_, err := u.client.RbacV1().ClusterRoles().Update(ctx, cr, metav1.UpdateOptions{})
	return err
}

func (u *clusterRolesUpdater) Delete(ctx context.Context, namespace, name string) error {
	return u.client.RbacV1().ClusterRoles().Delete(ctx, name, metav1.DeleteOptions{})
}

func (u *clusterRolesUpdater) Create(ctx context.Context, namespace string, obj interface{}) error {
	cr, ok := obj.(*rbacv1.ClusterRole)
	if !ok {
		return fmt.Errorf("invalid ClusterRole object")
	}
	_, err := u.client.RbacV1().ClusterRoles().Create(ctx, cr, metav1.CreateOptions{})
	return err
}

type clusterRoleBindingsUpdater struct{ client kubernetes.Interface }

func (u *clusterRoleBindingsUpdater) Update(ctx context.Context, namespace, name string, obj interface{}) error {
	crb, ok := obj.(*rbacv1.ClusterRoleBinding)
	if !ok {
		return fmt.Errorf("invalid ClusterRoleBinding object")
	}
	_, err := u.client.RbacV1().ClusterRoleBindings().Update(ctx, crb, metav1.UpdateOptions{})
	return err
}

func (u *clusterRoleBindingsUpdater) Delete(ctx context.Context, namespace, name string) error {
	return u.client.RbacV1().ClusterRoleBindings().Delete(ctx, name, metav1.DeleteOptions{})
}

func (u *clusterRoleBindingsUpdater) Create(ctx context.Context, namespace string, obj interface{}) error {
	crb, ok := obj.(*rbacv1.ClusterRoleBinding)
	if !ok {
		return fmt.Errorf("invalid ClusterRoleBinding object")
	}
	_, err := u.client.RbacV1().ClusterRoleBindings().Create(ctx, crb, metav1.CreateOptions{})
	return err
}

type resourceQuotasUpdater struct{ client kubernetes.Interface }

func (u *resourceQuotasUpdater) Update(ctx context.Context, namespace, name string, obj interface{}) error {
	rq, ok := obj.(*v1.ResourceQuota)
	if !ok {
		return fmt.Errorf("invalid ResourceQuota object")
	}
	_, err := u.client.CoreV1().ResourceQuotas(namespace).Update(ctx, rq, metav1.UpdateOptions{})
	return err
}

func (u *resourceQuotasUpdater) Delete(ctx context.Context, namespace, name string) error {
	return u.client.CoreV1().ResourceQuotas(namespace).Delete(ctx, name, metav1.DeleteOptions{})
}

func (u *resourceQuotasUpdater) Create(ctx context.Context, namespace string, obj interface{}) error {
	rq, ok := obj.(*v1.ResourceQuota)
	if !ok {
		return fmt.Errorf("invalid ResourceQuota object")
	}
	_, err := u.client.CoreV1().ResourceQuotas(namespace).Create(ctx, rq, metav1.CreateOptions{})
	return err
}

type limitRangesUpdater struct{ client kubernetes.Interface }

func (u *limitRangesUpdater) Update(ctx context.Context, namespace, name string, obj interface{}) error {
	lr, ok := obj.(*v1.LimitRange)
	if !ok {
		return fmt.Errorf("invalid LimitRange object")
	}
	_, err := u.client.CoreV1().LimitRanges(namespace).Update(ctx, lr, metav1.UpdateOptions{})
	return err
}

func (u *limitRangesUpdater) Delete(ctx context.Context, namespace, name string) error {
	return u.client.CoreV1().LimitRanges(namespace).Delete(ctx, name, metav1.DeleteOptions{})
}

func (u *limitRangesUpdater) Create(ctx context.Context, namespace string, obj interface{}) error {
	lr, ok := obj.(*v1.LimitRange)
	if !ok {
		return fmt.Errorf("invalid LimitRange object")
	}
	_, err := u.client.CoreV1().LimitRanges(namespace).Create(ctx, lr, metav1.CreateOptions{})
	return err
}

type podDisruptionBudgetsUpdater struct{ client kubernetes.Interface }

func (u *podDisruptionBudgetsUpdater) Update(ctx context.Context, namespace, name string, obj interface{}) error {
	pdb, ok := obj.(*policyv1.PodDisruptionBudget)
	if !ok {
		return fmt.Errorf("invalid PodDisruptionBudget object")
	}
	_, err := u.client.PolicyV1().PodDisruptionBudgets(namespace).Update(ctx, pdb, metav1.UpdateOptions{})
	return err
}

func (u *podDisruptionBudgetsUpdater) Delete(ctx context.Context, namespace, name string) error {
	return u.client.PolicyV1().PodDisruptionBudgets(namespace).Delete(ctx, name, metav1.DeleteOptions{})
}

func (u *podDisruptionBudgetsUpdater) Create(ctx context.Context, namespace string, obj interface{}) error {
	pdb, ok := obj.(*policyv1.PodDisruptionBudget)
	if !ok {
		return fmt.Errorf("invalid PodDisruptionBudget object")
	}
	_, err := u.client.PolicyV1().PodDisruptionBudgets(namespace).Create(ctx, pdb, metav1.CreateOptions{})
	return err
}

type eventsGetter struct{ client kubernetes.Interface }

func (g *eventsGetter) Get(ctx context.Context, namespace, name string) (interface{}, error) {
	return g.client.CoreV1().Events(namespace).Get(ctx, name, metav1.GetOptions{})
}

func (g *eventsGetter) List(ctx context.Context, namespace string, opts metav1.ListOptions) (interface{}, error) {
	return g.client.CoreV1().Events(namespace).List(ctx, opts)
}

type podsUpdater struct{ client kubernetes.Interface }

func (u *podsUpdater) Update(ctx context.Context, namespace, name string, obj interface{}) error {
	pod, ok := obj.(*v1.Pod)
	if !ok {
		return fmt.Errorf("invalid Pod object")
	}
	_, err := u.client.CoreV1().Pods(namespace).Update(ctx, pod, metav1.UpdateOptions{})
	return err
}

func (u *podsUpdater) Delete(ctx context.Context, namespace, name string) error {
	return u.client.CoreV1().Pods(namespace).Delete(ctx, name, metav1.DeleteOptions{})
}

type deploymentsUpdater struct{ client kubernetes.Interface }

func (u *deploymentsUpdater) Update(ctx context.Context, namespace, name string, obj interface{}) error {
	dep, ok := obj.(*appsv1.Deployment)
	if !ok {
		return fmt.Errorf("invalid Deployment object")
	}
	_, err := u.client.AppsV1().Deployments(namespace).Update(ctx, dep, metav1.UpdateOptions{})
	return err
}

func (u *deploymentsUpdater) Delete(ctx context.Context, namespace, name string) error {
	return u.client.AppsV1().Deployments(namespace).Delete(ctx, name, metav1.DeleteOptions{})
}

type statefulSetsUpdater struct{ client kubernetes.Interface }

func (u *statefulSetsUpdater) Update(ctx context.Context, namespace, name string, obj interface{}) error {
	sts, ok := obj.(*appsv1.StatefulSet)
	if !ok {
		return fmt.Errorf("invalid StatefulSet object")
	}
	_, err := u.client.AppsV1().StatefulSets(namespace).Update(ctx, sts, metav1.UpdateOptions{})
	return err
}

func (u *statefulSetsUpdater) Delete(ctx context.Context, namespace, name string) error {
	return u.client.AppsV1().StatefulSets(namespace).Delete(ctx, name, metav1.DeleteOptions{})
}

type daemonSetsUpdater struct{ client kubernetes.Interface }

func (u *daemonSetsUpdater) Update(ctx context.Context, namespace, name string, obj interface{}) error {
	ds, ok := obj.(*appsv1.DaemonSet)
	if !ok {
		return fmt.Errorf("invalid DaemonSet object")
	}
	_, err := u.client.AppsV1().DaemonSets(namespace).Update(ctx, ds, metav1.UpdateOptions{})
	return err
}

func (u *daemonSetsUpdater) Delete(ctx context.Context, namespace, name string) error {
	return u.client.AppsV1().DaemonSets(namespace).Delete(ctx, name, metav1.DeleteOptions{})
}

type servicesUpdater struct{ client kubernetes.Interface }

func (u *servicesUpdater) Update(ctx context.Context, namespace, name string, obj interface{}) error {
	svc, ok := obj.(*v1.Service)
	if !ok {
		return fmt.Errorf("invalid Service object")
	}
	_, err := u.client.CoreV1().Services(namespace).Update(ctx, svc, metav1.UpdateOptions{})
	return err
}

func (u *servicesUpdater) Delete(ctx context.Context, namespace, name string) error {
	return u.client.CoreV1().Services(namespace).Delete(ctx, name, metav1.DeleteOptions{})
}

type configMapsUpdater struct{ client kubernetes.Interface }

func (u *configMapsUpdater) Update(ctx context.Context, namespace, name string, obj interface{}) error {
	cm, ok := obj.(*v1.ConfigMap)
	if !ok {
		return fmt.Errorf("invalid ConfigMap object")
	}
	_, err := u.client.CoreV1().ConfigMaps(namespace).Update(ctx, cm, metav1.UpdateOptions{})
	return err
}

func (u *configMapsUpdater) Delete(ctx context.Context, namespace, name string) error {
	return u.client.CoreV1().ConfigMaps(namespace).Delete(ctx, name, metav1.DeleteOptions{})
}

type secretsUpdater struct{ client kubernetes.Interface }

func (u *secretsUpdater) Update(ctx context.Context, namespace, name string, obj interface{}) error {
	s, ok := obj.(*v1.Secret)
	if !ok {
		return fmt.Errorf("invalid Secret object")
	}
	_, err := u.client.CoreV1().Secrets(namespace).Update(ctx, s, metav1.UpdateOptions{})
	return err
}

func (u *secretsUpdater) Delete(ctx context.Context, namespace, name string) error {
	return u.client.CoreV1().Secrets(namespace).Delete(ctx, name, metav1.DeleteOptions{})
}

type ingressesUpdater struct{ client kubernetes.Interface }

func (u *ingressesUpdater) Update(ctx context.Context, namespace, name string, obj interface{}) error {
	ing, ok := obj.(*networkingv1.Ingress)
	if !ok {
		return fmt.Errorf("invalid Ingress object")
	}
	_, err := u.client.NetworkingV1().Ingresses(namespace).Update(ctx, ing, metav1.UpdateOptions{})
	return err
}

func (u *ingressesUpdater) Delete(ctx context.Context, namespace, name string) error {
	return u.client.NetworkingV1().Ingresses(namespace).Delete(ctx, name, metav1.DeleteOptions{})
}

type jobsUpdater struct{ client kubernetes.Interface }

func (u *jobsUpdater) Update(ctx context.Context, namespace, name string, obj interface{}) error {
	job, ok := obj.(*batchv1.Job)
	if !ok {
		return fmt.Errorf("invalid Job object")
	}
	_, err := u.client.BatchV1().Jobs(namespace).Update(ctx, job, metav1.UpdateOptions{})
	return err
}

func (u *jobsUpdater) Delete(ctx context.Context, namespace, name string) error {
	return u.client.BatchV1().Jobs(namespace).Delete(ctx, name, metav1.DeleteOptions{})
}

type cronJobsUpdater struct{ client kubernetes.Interface }

func (u *cronJobsUpdater) Update(ctx context.Context, namespace, name string, obj interface{}) error {
	cj, ok := obj.(*batchv1.CronJob)
	if !ok {
		return fmt.Errorf("invalid CronJob object")
	}
	_, err := u.client.BatchV1().CronJobs(namespace).Update(ctx, cj, metav1.UpdateOptions{})
	return err
}

func (u *cronJobsUpdater) Delete(ctx context.Context, namespace, name string) error {
	return u.client.BatchV1().CronJobs(namespace).Delete(ctx, name, metav1.DeleteOptions{})
}

type pvcsUpdater struct{ client kubernetes.Interface }

func (u *pvcsUpdater) Update(ctx context.Context, namespace, name string, obj interface{}) error {
	pvc, ok := obj.(*v1.PersistentVolumeClaim)
	if !ok {
		return fmt.Errorf("invalid PVC object")
	}
	_, err := u.client.CoreV1().PersistentVolumeClaims(namespace).Update(ctx, pvc, metav1.UpdateOptions{})
	return err
}

func (u *pvcsUpdater) Delete(ctx context.Context, namespace, name string) error {
	return u.client.CoreV1().PersistentVolumeClaims(namespace).Delete(ctx, name, metav1.DeleteOptions{})
}

type pvsUpdater struct{ client kubernetes.Interface }

func (u *pvsUpdater) Update(ctx context.Context, namespace, name string, obj interface{}) error {
	pv, ok := obj.(*v1.PersistentVolume)
	if !ok {
		return fmt.Errorf("invalid PV object")
	}
	_, err := u.client.CoreV1().PersistentVolumes().Update(ctx, pv, metav1.UpdateOptions{})
	return err
}

func (u *pvsUpdater) Delete(ctx context.Context, namespace, name string) error {
	return u.client.CoreV1().PersistentVolumes().Delete(ctx, name, metav1.DeleteOptions{})
}

type storageClassesUpdater struct{ client kubernetes.Interface }

func (u *storageClassesUpdater) Update(ctx context.Context, namespace, name string, obj interface{}) error {
	sc, ok := obj.(*storagev1.StorageClass)
	if !ok {
		return fmt.Errorf("invalid StorageClass object")
	}
	_, err := u.client.StorageV1().StorageClasses().Update(ctx, sc, metav1.UpdateOptions{})
	return err
}

func (u *storageClassesUpdater) Delete(ctx context.Context, namespace, name string) error {
	return u.client.StorageV1().StorageClasses().Delete(ctx, name, metav1.DeleteOptions{})
}

type namespacesUpdater struct{ client kubernetes.Interface }

func (u *namespacesUpdater) Update(ctx context.Context, namespace, name string, obj interface{}) error {
	ns, ok := obj.(*v1.Namespace)
	if !ok {
		return fmt.Errorf("invalid Namespace object")
	}
	_, err := u.client.CoreV1().Namespaces().Update(ctx, ns, metav1.UpdateOptions{})
	return err
}

func (u *namespacesUpdater) Delete(ctx context.Context, namespace, name string) error {
	return u.client.CoreV1().Namespaces().Delete(ctx, name, metav1.DeleteOptions{})
}

type nodesUpdater struct{ client kubernetes.Interface }

func (u *nodesUpdater) Update(ctx context.Context, namespace, name string, obj interface{}) error {
	node, ok := obj.(*v1.Node)
	if !ok {
		return fmt.Errorf("invalid Node object")
	}
	_, err := u.client.CoreV1().Nodes().Update(ctx, node, metav1.UpdateOptions{})
	return err
}

func (u *nodesUpdater) Delete(ctx context.Context, namespace, name string) error {
	return u.client.CoreV1().Nodes().Delete(ctx, name, metav1.DeleteOptions{})
}

func NewGetters(client kubernetes.Interface) map[ResourceType]Getter {
	return map[ResourceType]Getter{
		ResourcePod:                     &podsGetter{client},
		ResourceDeployment:              &deploymentsGetter{client},
		ResourceStatefulSet:             &statefulSetsGetter{client},
		ResourceDaemonSet:               &daemonSetsGetter{client},
		ResourceService:                 &servicesGetter{client},
		ResourceConfigMap:               &configMapsGetter{client},
		ResourceSecret:                  &secretsGetter{client},
		ResourceIngress:                 &ingressesGetter{client},
		ResourceJob:                     &jobsGetter{client},
		ResourceCronJob:                 &cronJobsGetter{client},
		ResourcePVC:                     &pvcsGetter{client},
		ResourcePV:                      &pvsGetter{client},
		ResourceStorageClass:            &storageClassesGetter{client},
		ResourceNamespace:               &namespacesGetter{client},
		ResourceNode:                    &nodesGetter{client},
		ResourceEndpoint:                &endpointsGetter{client},
		ResourceEvent:                   &eventsGetter{client},
		ResourceHorizontalPodAutoscaler: &hpasGetter{client},
		ResourceNetworkPolicy:           &networkPoliciesGetter{client},
		ResourceServiceAccount:          &serviceAccountsGetter{client},
		ResourceRole:                    &rolesGetter{client},
		ResourceRoleBinding:             &roleBindingsGetter{client},
		ResourceClusterRole:             &clusterRolesGetter{client},
		ResourceClusterRoleBinding:      &clusterRoleBindingsGetter{client},
		ResourceResourceQuota:           &resourceQuotasGetter{client},
		ResourceLimitRange:              &limitRangesGetter{client},
		ResourcePodDisruptionBudget:     &podDisruptionBudgetsGetter{client},
	}
}

func NewUpdaters(client kubernetes.Interface) map[ResourceType]Updater {
	return map[ResourceType]Updater{
		ResourcePod:                     &podsUpdater{client},
		ResourceDeployment:              &deploymentsUpdater{client},
		ResourceStatefulSet:             &statefulSetsUpdater{client},
		ResourceDaemonSet:               &daemonSetsUpdater{client},
		ResourceService:                 &servicesUpdater{client},
		ResourceConfigMap:               &configMapsUpdater{client},
		ResourceSecret:                  &secretsUpdater{client},
		ResourceIngress:                 &ingressesUpdater{client},
		ResourceJob:                     &jobsUpdater{client},
		ResourceCronJob:                 &cronJobsUpdater{client},
		ResourcePVC:                     &pvcsUpdater{client},
		ResourcePV:                      &pvsUpdater{client},
		ResourceStorageClass:            &storageClassesUpdater{client},
		ResourceNamespace:               &namespacesUpdater{client},
		ResourceNode:                    &nodesUpdater{client},
		ResourceHorizontalPodAutoscaler: &hpasUpdater{client},
		ResourceNetworkPolicy:           &networkPoliciesUpdater{client},
		ResourceServiceAccount:          &serviceAccountsUpdater{client},
		ResourceRole:                    &rolesUpdater{client},
		ResourceRoleBinding:             &roleBindingsUpdater{client},
		ResourceClusterRole:             &clusterRolesUpdater{client},
		ResourceClusterRoleBinding:      &clusterRoleBindingsUpdater{client},
		ResourceResourceQuota:           &resourceQuotasUpdater{client},
		ResourceLimitRange:              &limitRangesUpdater{client},
		ResourcePodDisruptionBudget:     &podDisruptionBudgetsUpdater{client},
	}
}

func NewDeleters(client kubernetes.Interface) map[ResourceType]Deleter {
	return map[ResourceType]Deleter{
		ResourcePod:                     &podsUpdater{client},
		ResourceDeployment:              &deploymentsUpdater{client},
		ResourceStatefulSet:             &statefulSetsUpdater{client},
		ResourceDaemonSet:               &daemonSetsUpdater{client},
		ResourceService:                 &servicesUpdater{client},
		ResourceConfigMap:               &configMapsUpdater{client},
		ResourceSecret:                  &secretsUpdater{client},
		ResourceIngress:                 &ingressesUpdater{client},
		ResourceJob:                     &jobsUpdater{client},
		ResourceCronJob:                 &cronJobsUpdater{client},
		ResourcePVC:                     &pvcsUpdater{client},
		ResourcePV:                      &pvsUpdater{client},
		ResourceStorageClass:            &storageClassesUpdater{client},
		ResourceNamespace:               &namespacesUpdater{client},
		ResourceNode:                    &nodesUpdater{client},
		ResourceHorizontalPodAutoscaler: &hpasUpdater{client},
		ResourceNetworkPolicy:           &networkPoliciesUpdater{client},
		ResourceServiceAccount:          &serviceAccountsUpdater{client},
		ResourceRole:                    &rolesUpdater{client},
		ResourceRoleBinding:             &roleBindingsUpdater{client},
		ResourceClusterRole:             &clusterRolesUpdater{client},
		ResourceClusterRoleBinding:      &clusterRoleBindingsUpdater{client},
		ResourceResourceQuota:           &resourceQuotasUpdater{client},
		ResourceLimitRange:              &limitRangesUpdater{client},
		ResourcePodDisruptionBudget:     &podDisruptionBudgetsUpdater{client},
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
		ResourcePod:                     &podsCreator{client},
		ResourceDeployment:              &deploymentsCreator{client},
		ResourceStatefulSet:             &statefulSetsCreator{client},
		ResourceDaemonSet:               &daemonSetsCreator{client},
		ResourceService:                 &servicesCreator{client},
		ResourceConfigMap:               &configMapsCreator{client},
		ResourceSecret:                  &secretsCreator{client},
		ResourceIngress:                 &ingressesCreator{client},
		ResourceJob:                     &jobsCreator{client},
		ResourceCronJob:                 &cronJobsCreator{client},
		ResourcePVC:                     &pvcsCreator{client},
		ResourcePV:                      &pvsCreator{client},
		ResourceStorageClass:            &storageClassesCreator{client},
		ResourceNamespace:               &namespacesCreator{client},
		ResourceNode:                    &nodesCreator{client},
		ResourceHorizontalPodAutoscaler: &hpasUpdater{client},
		ResourceNetworkPolicy:           &networkPoliciesUpdater{client},
		ResourceServiceAccount:          &serviceAccountsUpdater{client},
		ResourceRole:                    &rolesUpdater{client},
		ResourceRoleBinding:             &roleBindingsUpdater{client},
		ResourceClusterRole:             &clusterRolesUpdater{client},
		ResourceClusterRoleBinding:      &clusterRoleBindingsUpdater{client},
		ResourceResourceQuota:           &resourceQuotasUpdater{client},
		ResourceLimitRange:              &limitRangesUpdater{client},
		ResourcePodDisruptionBudget:     &podDisruptionBudgetsUpdater{client},
	}
}

// Creator 实现
type podsCreator struct{ client kubernetes.Interface }

func (c *podsCreator) Create(ctx context.Context, namespace string, obj interface{}) error {
	pod, ok := obj.(*v1.Pod)
	if !ok {
		return fmt.Errorf("invalid Pod object")
	}
	_, err := c.client.CoreV1().Pods(namespace).Create(ctx, pod, metav1.CreateOptions{})
	return err
}

type deploymentsCreator struct{ client kubernetes.Interface }

func (c *deploymentsCreator) Create(ctx context.Context, namespace string, obj interface{}) error {
	dep, ok := obj.(*appsv1.Deployment)
	if !ok {
		return fmt.Errorf("invalid Deployment object")
	}
	_, err := c.client.AppsV1().Deployments(namespace).Create(ctx, dep, metav1.CreateOptions{})
	return err
}

type statefulSetsCreator struct{ client kubernetes.Interface }

func (c *statefulSetsCreator) Create(ctx context.Context, namespace string, obj interface{}) error {
	sts, ok := obj.(*appsv1.StatefulSet)
	if !ok {
		return fmt.Errorf("invalid StatefulSet object")
	}
	_, err := c.client.AppsV1().StatefulSets(namespace).Create(ctx, sts, metav1.CreateOptions{})
	return err
}

type daemonSetsCreator struct{ client kubernetes.Interface }

func (c *daemonSetsCreator) Create(ctx context.Context, namespace string, obj interface{}) error {
	ds, ok := obj.(*appsv1.DaemonSet)
	if !ok {
		return fmt.Errorf("invalid DaemonSet object")
	}
	_, err := c.client.AppsV1().DaemonSets(namespace).Create(ctx, ds, metav1.CreateOptions{})
	return err
}

type servicesCreator struct{ client kubernetes.Interface }

func (c *servicesCreator) Create(ctx context.Context, namespace string, obj interface{}) error {
	svc, ok := obj.(*v1.Service)
	if !ok {
		return fmt.Errorf("invalid Service object")
	}
	_, err := c.client.CoreV1().Services(namespace).Create(ctx, svc, metav1.CreateOptions{})
	return err
}

type configMapsCreator struct{ client kubernetes.Interface }

func (c *configMapsCreator) Create(ctx context.Context, namespace string, obj interface{}) error {
	cm, ok := obj.(*v1.ConfigMap)
	if !ok {
		return fmt.Errorf("invalid ConfigMap object")
	}
	_, err := c.client.CoreV1().ConfigMaps(namespace).Create(ctx, cm, metav1.CreateOptions{})
	return err
}

type secretsCreator struct{ client kubernetes.Interface }

func (c *secretsCreator) Create(ctx context.Context, namespace string, obj interface{}) error {
	s, ok := obj.(*v1.Secret)
	if !ok {
		return fmt.Errorf("invalid Secret object")
	}
	_, err := c.client.CoreV1().Secrets(namespace).Create(ctx, s, metav1.CreateOptions{})
	return err
}

type ingressesCreator struct{ client kubernetes.Interface }

func (c *ingressesCreator) Create(ctx context.Context, namespace string, obj interface{}) error {
	ing, ok := obj.(*networkingv1.Ingress)
	if !ok {
		return fmt.Errorf("invalid Ingress object")
	}
	_, err := c.client.NetworkingV1().Ingresses(namespace).Create(ctx, ing, metav1.CreateOptions{})
	return err
}

type jobsCreator struct{ client kubernetes.Interface }

func (c *jobsCreator) Create(ctx context.Context, namespace string, obj interface{}) error {
	job, ok := obj.(*batchv1.Job)
	if !ok {
		return fmt.Errorf("invalid Job object")
	}
	_, err := c.client.BatchV1().Jobs(namespace).Create(ctx, job, metav1.CreateOptions{})
	return err
}

type cronJobsCreator struct{ client kubernetes.Interface }

func (c *cronJobsCreator) Create(ctx context.Context, namespace string, obj interface{}) error {
	cj, ok := obj.(*batchv1.CronJob)
	if !ok {
		return fmt.Errorf("invalid CronJob object")
	}
	_, err := c.client.BatchV1().CronJobs(namespace).Create(ctx, cj, metav1.CreateOptions{})
	return err
}

type pvcsCreator struct{ client kubernetes.Interface }

func (c *pvcsCreator) Create(ctx context.Context, namespace string, obj interface{}) error {
	pvc, ok := obj.(*v1.PersistentVolumeClaim)
	if !ok {
		return fmt.Errorf("invalid PVC object")
	}
	_, err := c.client.CoreV1().PersistentVolumeClaims(namespace).Create(ctx, pvc, metav1.CreateOptions{})
	return err
}

type pvsCreator struct{ client kubernetes.Interface }

func (c *pvsCreator) Create(ctx context.Context, namespace string, obj interface{}) error {
	pv, ok := obj.(*v1.PersistentVolume)
	if !ok {
		return fmt.Errorf("invalid PV object")
	}
	_, err := c.client.CoreV1().PersistentVolumes().Create(ctx, pv, metav1.CreateOptions{})
	return err
}

type storageClassesCreator struct{ client kubernetes.Interface }

func (c *storageClassesCreator) Create(ctx context.Context, namespace string, obj interface{}) error {
	sc, ok := obj.(*storagev1.StorageClass)
	if !ok {
		return fmt.Errorf("invalid StorageClass object")
	}
	_, err := c.client.StorageV1().StorageClasses().Create(ctx, sc, metav1.CreateOptions{})
	return err
}

type namespacesCreator struct{ client kubernetes.Interface }

func (c *namespacesCreator) Create(ctx context.Context, namespace string, obj interface{}) error {
	ns, ok := obj.(*v1.Namespace)
	if !ok {
		return fmt.Errorf("invalid Namespace object")
	}
	_, err := c.client.CoreV1().Namespaces().Create(ctx, ns, metav1.CreateOptions{})
	return err
}

type nodesCreator struct{ client kubernetes.Interface }

func (c *nodesCreator) Create(ctx context.Context, namespace string, obj interface{}) error {
	node, ok := obj.(*v1.Node)
	if !ok {
		return fmt.Errorf("invalid Node object")
	}
	_, err := c.client.CoreV1().Nodes().Create(ctx, node, metav1.CreateOptions{})
	return err
}
