package service

import (
	"context"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/nick0323/K8sVision/model"
	"github.com/nick0323/K8sVision/pkg/k8s"
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

func toSearchableItems[T model.SearchableItem](items []T) []model.SearchableItem {
	res := make([]model.SearchableItem, len(items))
	for i := range items {
		res[i] = items[i]
	}
	return res
}

func GetResourceByName(ctx context.Context, clientset kubernetes.Interface, resourceType, namespace, name string) (any, error) {
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

	fieldSelectors := []string{}
	if fieldSelector != "" {
		fieldSelectors = append(fieldSelectors, fieldSelector)
	}

	if rt == k8s.ResourceEvent && involvedObject != "" {
		parts := strings.SplitN(involvedObject, "/", 2)
		if len(parts) == 2 {
			fieldSelectors = append(fieldSelectors,
				fmt.Sprintf("involvedObject.kind=%s", parts[0]),
				fmt.Sprintf("involvedObject.name=%s", parts[1]),
			)
		} else {
			fieldSelectors = append(fieldSelectors, "involvedObject.name="+involvedObject)
		}
	}

	if len(fieldSelectors) > 0 {
		opts.FieldSelector = strings.Join(fieldSelectors, ",")
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

func convertToSearchableItems(result any, resourceType string, rt k8s.ResourceType, since string) ([]model.SearchableItem, error) {
	mappers := map[k8s.ResourceType]func() ([]model.SearchableItem, error){
		k8s.ResourcePod: func() ([]model.SearchableItem, error) {
			items, ok := result.(*v1.PodList)
			if !ok {
				return nil, fmt.Errorf("invalid pod list")
			}
			return toSearchableItems(MapPods(items.Items)), nil
		},
		k8s.ResourceDeployment: func() ([]model.SearchableItem, error) {
			items, ok := result.(*appsv1.DeploymentList)
			if !ok {
				return nil, fmt.Errorf("invalid deployment list")
			}
			return toSearchableItems(MapDeployments(items.Items)), nil
		},
		k8s.ResourceStatefulSet: func() ([]model.SearchableItem, error) {
			items, ok := result.(*appsv1.StatefulSetList)
			if !ok {
				return nil, fmt.Errorf("invalid statefulset list")
			}
			return toSearchableItems(MapStatefulSets(items.Items)), nil
		},
		k8s.ResourceDaemonSet: func() ([]model.SearchableItem, error) {
			items, ok := result.(*appsv1.DaemonSetList)
			if !ok {
				return nil, fmt.Errorf("invalid daemonset list")
			}
			return toSearchableItems(MapDaemonSets(items.Items)), nil
		},
		k8s.ResourceService: func() ([]model.SearchableItem, error) {
			items, ok := result.(*v1.ServiceList)
			if !ok {
				return nil, fmt.Errorf("invalid service list")
			}
			return toSearchableItems(MapServices(items.Items)), nil
		},
		k8s.ResourceConfigMap: func() ([]model.SearchableItem, error) {
			items, ok := result.(*v1.ConfigMapList)
			if !ok {
				return nil, fmt.Errorf("invalid configmap list")
			}
			return toSearchableItems(MapConfigMaps(items.Items)), nil
		},
		k8s.ResourceSecret: func() ([]model.SearchableItem, error) {
			items, ok := result.(*v1.SecretList)
			if !ok {
				return nil, fmt.Errorf("invalid secret list")
			}
			return toSearchableItems(MapSecrets(items.Items)), nil
		},
		k8s.ResourceIngress: func() ([]model.SearchableItem, error) {
			items, ok := result.(*networkingv1.IngressList)
			if !ok {
				return nil, fmt.Errorf("invalid ingress list")
			}
			return toSearchableItems(MapIngresses(items.Items)), nil
		},
		k8s.ResourceJob: func() ([]model.SearchableItem, error) {
			items, ok := result.(*batchv1.JobList)
			if !ok {
				return nil, fmt.Errorf("invalid job list")
			}
			return toSearchableItems(MapJobs(items.Items)), nil
		},
		k8s.ResourceCronJob: func() ([]model.SearchableItem, error) {
			items, ok := result.(*batchv1.CronJobList)
			if !ok {
				return nil, fmt.Errorf("invalid cronjob list")
			}
			return toSearchableItems(MapCronJobs(items.Items)), nil
		},
		k8s.ResourcePVC: func() ([]model.SearchableItem, error) {
			items, ok := result.(*v1.PersistentVolumeClaimList)
			if !ok {
				return nil, fmt.Errorf("invalid pvc list")
			}
			return toSearchableItems(MapPVCs(items.Items)), nil
		},
		k8s.ResourcePV: func() ([]model.SearchableItem, error) {
			items, ok := result.(*v1.PersistentVolumeList)
			if !ok {
				return nil, fmt.Errorf("invalid pv list")
			}
			return toSearchableItems(MapPVs(items.Items)), nil
		},
		k8s.ResourceStorageClass: func() ([]model.SearchableItem, error) {
			items, ok := result.(*storagev1.StorageClassList)
			if !ok {
				return nil, fmt.Errorf("invalid storageclass list")
			}
			return toSearchableItems(MapStorageClasses(items.Items)), nil
		},
		k8s.ResourceNamespace: func() ([]model.SearchableItem, error) {
			items, ok := result.(*v1.NamespaceList)
			if !ok {
				return nil, fmt.Errorf("invalid namespace list")
			}
			return toSearchableItems(MapNamespaces(items.Items)), nil
		},
		k8s.ResourceNode: func() ([]model.SearchableItem, error) {
			items, ok := result.(*v1.NodeList)
			if !ok {
				return nil, fmt.Errorf("invalid node list")
			}
			return toSearchableItems(MapNodes(items.Items, nil, make(map[string]*model.NodeMetrics))), nil
		},
		k8s.ResourceEndpoint: func() ([]model.SearchableItem, error) {
			items, ok := result.(*v1.EndpointsList)
			if !ok {
				return nil, fmt.Errorf("invalid endpoints list")
			}
			return toSearchableItems(MapEndpoints(items.Items)), nil
		},
		k8s.ResourceEvent: func() ([]model.SearchableItem, error) {
			items, ok := result.(*v1.EventList)
			if !ok {
				return nil, fmt.Errorf("invalid event list")
			}
			return toSearchableItems(MapEvents(filterEventsByTime(items.Items, since))), nil
		},
		k8s.ResourceNetworkPolicy: func() ([]model.SearchableItem, error) {
			items, ok := result.(*networkingv1.NetworkPolicyList)
			if !ok {
				return nil, fmt.Errorf("invalid networkpolicy list")
			}
			return toSearchableItems(MapNetworkPolicies(items.Items)), nil
		},
		k8s.ResourceServiceAccount: func() ([]model.SearchableItem, error) {
			items, ok := result.(*v1.ServiceAccountList)
			if !ok {
				return nil, fmt.Errorf("invalid serviceaccount list")
			}
			return toSearchableItems(MapServiceAccounts(items.Items)), nil
		},
		k8s.ResourceRole: func() ([]model.SearchableItem, error) {
			items, ok := result.(*rbacv1.RoleList)
			if !ok {
				return nil, fmt.Errorf("invalid role list")
			}
			return toSearchableItems(MapRoles(items.Items)), nil
		},
		k8s.ResourceRoleBinding: func() ([]model.SearchableItem, error) {
			items, ok := result.(*rbacv1.RoleBindingList)
			if !ok {
				return nil, fmt.Errorf("invalid rolebinding list")
			}
			return toSearchableItems(MapRoleBindings(items.Items)), nil
		},
		k8s.ResourceClusterRole: func() ([]model.SearchableItem, error) {
			items, ok := result.(*rbacv1.ClusterRoleList)
			if !ok {
				return nil, fmt.Errorf("invalid clusterrole list")
			}
			return toSearchableItems(MapClusterRoles(items.Items)), nil
		},
		k8s.ResourceClusterRoleBinding: func() ([]model.SearchableItem, error) {
			items, ok := result.(*rbacv1.ClusterRoleBindingList)
			if !ok {
				return nil, fmt.Errorf("invalid clusterrolebinding list")
			}
			return toSearchableItems(MapClusterRoleBindings(items.Items)), nil
		},
		k8s.ResourceResourceQuota: func() ([]model.SearchableItem, error) {
			items, ok := result.(*v1.ResourceQuotaList)
			if !ok {
				return nil, fmt.Errorf("invalid resourcequota list")
			}
			return toSearchableItems(MapResourceQuotas(items.Items)), nil
		},
		k8s.ResourceLimitRange: func() ([]model.SearchableItem, error) {
			items, ok := result.(*v1.LimitRangeList)
			if !ok {
				return nil, fmt.Errorf("invalid limitrange list")
			}
			return toSearchableItems(MapLimitRanges(items.Items)), nil
		},
		k8s.ResourcePodDisruptionBudget: func() ([]model.SearchableItem, error) {
			items, ok := result.(*policyv1.PodDisruptionBudgetList)
			if !ok {
				return nil, fmt.Errorf("invalid poddisruptionbudget list")
			}
			return toSearchableItems(MapPodDisruptionBudgets(items.Items)), nil
		},
		k8s.ResourceHorizontalPodAutoscaler: func() ([]model.SearchableItem, error) {
			items, ok := result.(*autoscalingv2.HorizontalPodAutoscalerList)
			if !ok {
				return nil, fmt.Errorf("invalid hpa list")
			}
			return toSearchableItems(MapHPAs(items.Items)), nil
		},
	}

	mapper, ok := mappers[rt]
	if !ok {
		return nil, fmt.Errorf("unsupported resource type: %s", resourceType)
	}
	return mapper()
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

	slices.SortFunc(filtered, func(a, b v1.Event) int {
		aTime := a.LastTimestamp.Time
		if aTime.IsZero() {
			aTime = a.EventTime.Time
		}
		bTime := b.LastTimestamp.Time
		if bTime.IsZero() {
			bTime = b.EventTime.Time
		}
		return bTime.Compare(aTime)
	})

	return filtered
}
