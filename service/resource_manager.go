package service

import (
	"context"
	"fmt"
	"sort"
	"time"

	"github.com/nick0323/K8sVision/model"
	"github.com/nick0323/K8sVision/pkg/k8s"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	storagev1 "k8s.io/api/storage/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func GetResourceByName(ctx context.Context, clientset kubernetes.Interface, resourceType, namespace, name string) (interface{}, error) {
	rt := k8s.ResourceType(resourceType).Normalize()
	getters := k8s.NewGetters(clientset)

	getter, ok := getters[rt]
	if !ok {
		return nil, fmt.Errorf("unsupported resource type: %s", resourceType)
	}

	ns := namespace
	if rt.IsClusterScoped() {
		ns = ""
	}

	return getter.Get(ctx, ns, name)
}

func DeleteResourceByType(ctx context.Context, clientset kubernetes.Interface, resourceType, namespace, name string) error {
	rt := k8s.ResourceType(resourceType).Normalize()
	deleters := k8s.NewDeleters(clientset)

	deleter, ok := deleters[rt]
	if !ok {
		return fmt.Errorf("unsupported resource type: %s", resourceType)
	}

	ns := namespace
	if rt.IsClusterScoped() {
		ns = ""
	}

	return deleter.Delete(ctx, ns, name)
}

func ListResourcesByType(ctx context.Context, clientset kubernetes.Interface, resourceType, namespace, labelSelector, fieldSelector, involvedObject, since string) ([]model.SearchableItem, error) {
	rt := k8s.ResourceType(resourceType).Normalize()
	getters := k8s.NewGetters(clientset)

	getter, ok := getters[rt]
	if !ok {
		return nil, fmt.Errorf("unsupported resource type: %s", resourceType)
	}

	opts := metav1.ListOptions{}
	if labelSelector != "" {
		opts.LabelSelector = labelSelector
	}
	if fieldSelector != "" {
		opts.FieldSelector = fieldSelector
	}

	ns := namespace
	if rt.IsClusterScoped() {
		ns = ""
	}

	result, err := getter.List(ctx, ns, opts)
	if err != nil {
		return nil, err
	}

	return convertToSearchableItems(result, resourceType, rt, since)
}

func convertToSearchableItems(result interface{}, resourceType string, rt k8s.ResourceType, since string) ([]model.SearchableItem, error) {
	switch rt {
	case k8s.ResourcePod:
		items, ok := result.(*v1.PodList)
		if !ok {
			return nil, fmt.Errorf("invalid pod list")
		}
		pods := MapPods(items.Items)
		res := make([]model.SearchableItem, len(pods))
		for i := range pods {
			res[i] = &pods[i]
		}
		return res, nil

	case k8s.ResourceDeployment:
		items, ok := result.(*appsv1.DeploymentList)
		if !ok {
			return nil, fmt.Errorf("invalid deployment list")
		}
		deps := MapDeployments(items.Items)
		res := make([]model.SearchableItem, len(deps))
		for i := range deps {
			res[i] = &deps[i]
		}
		return res, nil

	case k8s.ResourceStatefulSet:
		items, ok := result.(*appsv1.StatefulSetList)
		if !ok {
			return nil, fmt.Errorf("invalid statefulset list")
		}
		sts := MapStatefulSets(items.Items)
		res := make([]model.SearchableItem, len(sts))
		for i := range sts {
			res[i] = &sts[i]
		}
		return res, nil

	case k8s.ResourceDaemonSet:
		items, ok := result.(*appsv1.DaemonSetList)
		if !ok {
			return nil, fmt.Errorf("invalid daemonset list")
		}
		dss := MapDaemonSets(items.Items)
		res := make([]model.SearchableItem, len(dss))
		for i := range dss {
			res[i] = &dss[i]
		}
		return res, nil

	case k8s.ResourceService:
		items, ok := result.(*v1.ServiceList)
		if !ok {
			return nil, fmt.Errorf("invalid service list")
		}
		svcs := MapServices(items.Items)
		res := make([]model.SearchableItem, len(svcs))
		for i := range svcs {
			res[i] = &svcs[i]
		}
		return res, nil

	case k8s.ResourceConfigMap:
		items, ok := result.(*v1.ConfigMapList)
		if !ok {
			return nil, fmt.Errorf("invalid configmap list")
		}
		cms := MapConfigMaps(items.Items)
		res := make([]model.SearchableItem, len(cms))
		for i := range cms {
			res[i] = &cms[i]
		}
		return res, nil

	case k8s.ResourceSecret:
		items, ok := result.(*v1.SecretList)
		if !ok {
			return nil, fmt.Errorf("invalid secret list")
		}
		secrets := MapSecrets(items.Items)
		res := make([]model.SearchableItem, len(secrets))
		for i := range secrets {
			res[i] = &secrets[i]
		}
		return res, nil

	case k8s.ResourceIngress:
		items, ok := result.(*networkingv1.IngressList)
		if !ok {
			return nil, fmt.Errorf("invalid ingress list")
		}
		ings := MapIngresses(items.Items)
		res := make([]model.SearchableItem, len(ings))
		for i := range ings {
			res[i] = &ings[i]
		}
		return res, nil

	case k8s.ResourceJob:
		items, ok := result.(*batchv1.JobList)
		if !ok {
			return nil, fmt.Errorf("invalid job list")
		}
		jobs := MapJobs(items.Items)
		res := make([]model.SearchableItem, len(jobs))
		for i := range jobs {
			res[i] = &jobs[i]
		}
		return res, nil

	case k8s.ResourceCronJob:
		items, ok := result.(*batchv1.CronJobList)
		if !ok {
			return nil, fmt.Errorf("invalid cronjob list")
		}
		cjs := MapCronJobs(items.Items)
		res := make([]model.SearchableItem, len(cjs))
		for i := range cjs {
			res[i] = &cjs[i]
		}
		return res, nil

	case k8s.ResourcePVC:
		items, ok := result.(*v1.PersistentVolumeClaimList)
		if !ok {
			return nil, fmt.Errorf("invalid pvc list")
		}
		pvcs := MapPVCs(items.Items)
		res := make([]model.SearchableItem, len(pvcs))
		for i := range pvcs {
			res[i] = &pvcs[i]
		}
		return res, nil

	case k8s.ResourcePV:
		items, ok := result.(*v1.PersistentVolumeList)
		if !ok {
			return nil, fmt.Errorf("invalid pv list")
		}
		pvs := MapPVs(items.Items)
		res := make([]model.SearchableItem, len(pvs))
		for i := range pvs {
			res[i] = &pvs[i]
		}
		return res, nil

	case k8s.ResourceStorageClass:
		items, ok := result.(*storagev1.StorageClassList)
		if !ok {
			return nil, fmt.Errorf("invalid storageclass list")
		}
		scs := MapStorageClasses(items.Items)
		res := make([]model.SearchableItem, len(scs))
		for i := range scs {
			res[i] = &scs[i]
		}
		return res, nil

	case k8s.ResourceNamespace:
		items, ok := result.(*v1.NamespaceList)
		if !ok {
			return nil, fmt.Errorf("invalid namespace list")
		}
		nss := MapNamespaces(items.Items)
		res := make([]model.SearchableItem, len(nss))
		for i := range nss {
			res[i] = &nss[i]
		}
		return res, nil

	case k8s.ResourceNode:
		items, ok := result.(*v1.NodeList)
		if !ok {
			return nil, fmt.Errorf("invalid node list")
		}
		nodes := MapNodes(items.Items, nil)
		res := make([]model.SearchableItem, len(nodes))
		for i := range nodes {
			res[i] = &nodes[i]
		}
		return res, nil

	case k8s.ResourceEndpoint:
		items, ok := result.(*v1.EndpointsList)
		if !ok {
			return nil, fmt.Errorf("invalid endpoints list")
		}
		eps := MapEndpoints(items.Items)
		res := make([]model.SearchableItem, len(eps))
		for i := range eps {
			res[i] = &eps[i]
		}
		return res, nil

	case k8s.ResourceEvent:
		items, ok := result.(*v1.EventList)
		if !ok {
			return nil, fmt.Errorf("invalid event list")
		}
		events := MapEvents(filterEventsByTime(items.Items, since))
		res := make([]model.SearchableItem, len(events))
		for i := range events {
			res[i] = &events[i]
		}
		return res, nil

	default:
		return nil, fmt.Errorf("unsupported resource type: %s", resourceType)
	}
}

func filterEventsByTime(events []v1.Event, since string) []v1.Event {
	if since == "" {
		return events
	}

	duration, err := time.ParseDuration(since)
	if err != nil {
		duration = 1 * time.Hour
	}

	cutoffTime := time.Now().Add(-duration)
	filtered := make([]v1.Event, 0, len(events))
	for _, e := range events {
		eventTime := e.LastTimestamp.Time
		if eventTime.IsZero() {
			eventTime = e.EventTime.Time
		}
		if !eventTime.IsZero() && eventTime.After(cutoffTime) {
			filtered = append(filtered, e)
		}
	}

	sort.Slice(filtered, func(i, j int) bool {
		iTime := filtered[i].LastTimestamp.Time
		if iTime.IsZero() {
			iTime = filtered[i].EventTime.Time
		}
		jTime := filtered[j].LastTimestamp.Time
		if jTime.IsZero() {
			jTime = filtered[j].EventTime.Time
		}
		return iTime.After(jTime)
	})

	return filtered
}