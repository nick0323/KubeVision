package k8s

import (
	"context"
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	storagev1 "k8s.io/api/storage/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type ResourceType string

const (
	ResourcePod           ResourceType = "pod"
	ResourceDeployment   ResourceType = "deployment"
	ResourceStatefulSet ResourceType = "statefulset"
	ResourceDaemonSet   ResourceType = "daemonset"
	ResourceService     ResourceType = "service"
	ResourceConfigMap   ResourceType = "configmap"
	ResourceSecret     ResourceType = "secret"
	ResourceIngress   ResourceType = "ingress"
	ResourceJob       ResourceType = "job"
	ResourceCronJob   ResourceType = "cronjob"
	ResourcePVC      ResourceType = "persistentvolumeclaim"
	ResourcePV      ResourceType = "persistentvolume"
	ResourceStorageClass ResourceType = "storageclass"
	ResourceNamespace ResourceType = "namespace"
	ResourceNode     ResourceType = "node"
	ResourceEndpoint ResourceType = "endpoint"
	ResourceEvent   ResourceType = "event"
)

var clusterScopedResources = map[ResourceType]bool{
	ResourcePod:           false,
	ResourceDeployment:   false,
	ResourceStatefulSet:    false,
	ResourceDaemonSet:     false,
	ResourceService:      false,
	ResourceConfigMap:   false,
	ResourceSecret:      false,
	ResourceIngress:    false,
	ResourceJob:        false,
	ResourceCronJob:    false,
	ResourcePVC:         false,
	ResourcePV:         true,
	ResourceStorageClass: true,
	ResourceNamespace:   true,
	ResourceNode:        true,
	ResourceEndpoint:    false,
	ResourceEvent:      false,
}

func (rt ResourceType) IsClusterScoped() bool {
	return clusterScopedResources[rt]
}

func (rt ResourceType) Normalize() ResourceType {
	normalized := string(rt)
	switch normalized {
	case "pvc":
		normalized = "persistentvolumeclaim"
	case "pv":
		normalized = "persistentvolume"
	case "sc":
		normalized = "storageclass"
	default:
		normalized = rt.normalizeSingular()
	}
	return ResourceType(normalized)
}

func (rt ResourceType) normalizeSingular() string {
	s := string(rt)
	if len(s) > 3 && s[len(s)-3:] == "ies" {
		return s[:len(s)-3] + "y"
	}
	if len(s) > 1 && s[len(s)-1:] == "s" && s[len(s)-2:] != "ss" {
		return s[:len(s)-1]
	}
	return s
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
		ResourcePod:             &podsGetter{client},
		ResourceDeployment:     &deploymentsGetter{client},
		ResourceStatefulSet:     &statefulSetsGetter{client},
		ResourceDaemonSet:       &daemonSetsGetter{client},
		ResourceService:       &servicesGetter{client},
		ResourceConfigMap:     &configMapsGetter{client},
		ResourceSecret:        &secretsGetter{client},
		ResourceIngress:       &ingressesGetter{client},
		ResourceJob:           &jobsGetter{client},
		ResourceCronJob:       &cronJobsGetter{client},
		ResourcePVC:            &pvcsGetter{client},
		ResourcePV:             &pvsGetter{client},
		ResourceStorageClass:   &storageClassesGetter{client},
		ResourceNamespace:     &namespacesGetter{client},
		ResourceNode:          &nodesGetter{client},
		ResourceEndpoint:     &endpointsGetter{client},
		ResourceEvent:        &eventsGetter{client},
	}
}

func NewUpdaters(client kubernetes.Interface) map[ResourceType]Updater {
	return map[ResourceType]Updater{
		ResourcePod:           &podsUpdater{client},
		ResourceDeployment:   &deploymentsUpdater{client},
		ResourceStatefulSet:   &statefulSetsUpdater{client},
		ResourceDaemonSet:     &daemonSetsUpdater{client},
		ResourceService:       &servicesUpdater{client},
		ResourceConfigMap:     &configMapsUpdater{client},
		ResourceSecret:        &secretsUpdater{client},
		ResourceIngress:      &ingressesUpdater{client},
		ResourceJob:          &jobsUpdater{client},
		ResourceCronJob:      &cronJobsUpdater{client},
		ResourcePVC:          &pvcsUpdater{client},
		ResourcePV:           &pvsUpdater{client},
		ResourceStorageClass: &storageClassesUpdater{client},
		ResourceNamespace:   &namespacesUpdater{client},
		ResourceNode:        &nodesUpdater{client},
	}
}

func NewDeleters(client kubernetes.Interface) map[ResourceType]Deleter {
	return map[ResourceType]Deleter{
		ResourcePod:           &podsUpdater{client},
		ResourceDeployment:   &deploymentsUpdater{client},
		ResourceStatefulSet:   &statefulSetsUpdater{client},
		ResourceDaemonSet:     &daemonSetsUpdater{client},
		ResourceService:       &servicesUpdater{client},
		ResourceConfigMap:     &configMapsUpdater{client},
		ResourceSecret:        &secretsUpdater{client},
		ResourceIngress:      &ingressesUpdater{client},
		ResourceJob:          &jobsUpdater{client},
		ResourceCronJob:      &cronJobsUpdater{client},
		ResourcePVC:          &pvcsUpdater{client},
		ResourcePV:           &pvsUpdater{client},
		ResourceStorageClass: &storageClassesUpdater{client},
		ResourceNamespace:   &namespacesUpdater{client},
		ResourceNode:        &nodesUpdater{client},
	}
}