package service

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/nick0323/K8sVision/model"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

type ListOptions struct {
	Namespace     string
	LabelSelector string
	FieldSelector string
	Limit         int64
}

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

func DefaultListOptions() *ListOptions {
	return &ListOptions{Limit: 1000}
}

func ListPods(ctx context.Context, clientset *kubernetes.Clientset, namespace, labelSelector, fieldSelector string) ([]model.Pod, error) {
	pods, _, err := ListPodsWithRaw(ctx, clientset, namespace, labelSelector, fieldSelector)
	return pods, err
}

func ListPodsWithRaw(ctx context.Context, clientset *kubernetes.Clientset, namespace, labelSelector, fieldSelector string, noLimit ...bool) ([]model.Pod, *corev1.PodList, error) {
	opts := DefaultListOptions()
	opts.Namespace = namespace
	opts.LabelSelector = labelSelector
	opts.FieldSelector = fieldSelector

	if len(noLimit) > 0 && noLimit[0] {
		opts.Limit = 0
	}

	pods, err := ListResourcesWithNamespace(ctx, namespace,
		func() (*corev1.PodList, error) {
			return clientset.CoreV1().Pods("").List(ctx, opts.Apply())
		},
		func(ns string) (*corev1.PodList, error) {
			return clientset.CoreV1().Pods(ns).List(ctx, opts.Apply())
		},
	)
	if err != nil {
		return nil, nil, err
	}

	return MapPods(pods.Items), pods, nil
}

func ListDeployments(ctx context.Context, clientset *kubernetes.Clientset, namespace, labelSelector, fieldSelector string) ([]model.Deployment, error) {
	opts := DefaultListOptions()
	opts.LabelSelector = labelSelector
	opts.FieldSelector = fieldSelector
	depList, err := clientset.AppsV1().Deployments(namespace).List(ctx, opts.Apply())
	if err != nil {
		return nil, err
	}
	return MapDeployments(depList.Items), nil
}

func ListStatefulSets(ctx context.Context, clientset *kubernetes.Clientset, namespace, labelSelector, fieldSelector string) ([]model.StatefulSet, error) {
	opts := DefaultListOptions()
	opts.LabelSelector = labelSelector
	opts.FieldSelector = fieldSelector
	stsList, err := clientset.AppsV1().StatefulSets(namespace).List(ctx, opts.Apply())
	if err != nil {
		return nil, err
	}
	return MapStatefulSets(stsList.Items), nil
}

func ListDaemonSets(ctx context.Context, clientset *kubernetes.Clientset, namespace, labelSelector, fieldSelector string) ([]model.DaemonSet, error) {
	opts := DefaultListOptions()
	opts.LabelSelector = labelSelector
	opts.FieldSelector = fieldSelector
	dsList, err := clientset.AppsV1().DaemonSets(namespace).List(ctx, opts.Apply())
	if err != nil {
		return nil, err
	}
	return MapDaemonSets(dsList.Items), nil
}

func ListJobs(ctx context.Context, clientset *kubernetes.Clientset, namespace, labelSelector, fieldSelector string) ([]model.Job, error) {
	opts := DefaultListOptions()
	opts.LabelSelector = labelSelector
	opts.FieldSelector = fieldSelector
	jobs, err := clientset.BatchV1().Jobs(namespace).List(ctx, opts.Apply())
	if err != nil {
		return nil, err
	}
	return MapJobs(jobs.Items), nil
}

func ListCronJobs(ctx context.Context, clientset *kubernetes.Clientset, namespace, labelSelector, fieldSelector string) ([]model.CronJob, error) {
	opts := DefaultListOptions()
	opts.LabelSelector = labelSelector
	opts.FieldSelector = fieldSelector
	cronjobs, err := clientset.BatchV1().CronJobs(namespace).List(ctx, opts.Apply())
	if err != nil {
		return nil, err
	}
	return MapCronJobs(cronjobs.Items), nil
}

func ListServices(ctx context.Context, clientset *kubernetes.Clientset, namespace, labelSelector, fieldSelector string) ([]model.Service, error) {
	opts := DefaultListOptions()
	opts.LabelSelector = labelSelector
	opts.FieldSelector = fieldSelector
	svcs, err := clientset.CoreV1().Services(namespace).List(ctx, opts.Apply())
	if err != nil {
		return nil, err
	}
	return MapServices(svcs.Items), nil
}

func ListIngresses(ctx context.Context, clientset *kubernetes.Clientset, namespace, labelSelector, fieldSelector string) ([]model.Ingress, error) {
	opts := DefaultListOptions()
	opts.LabelSelector = labelSelector
	opts.FieldSelector = fieldSelector
	ingresses, err := clientset.NetworkingV1().Ingresses(namespace).List(ctx, opts.Apply())
	if err != nil {
		return nil, err
	}
	return MapIngresses(ingresses.Items), nil
}

func ListPVCs(ctx context.Context, clientset *kubernetes.Clientset, namespace, labelSelector, fieldSelector string) ([]model.PVC, error) {
	opts := DefaultListOptions()
	opts.LabelSelector = labelSelector
	opts.FieldSelector = fieldSelector
	pvcList, err := ListResourcesWithNamespace(ctx, namespace,
		func() (*corev1.PersistentVolumeClaimList, error) {
			return clientset.CoreV1().PersistentVolumeClaims("").List(ctx, opts.Apply())
		},
		func(ns string) (*corev1.PersistentVolumeClaimList, error) {
			return clientset.CoreV1().PersistentVolumeClaims(ns).List(ctx, opts.Apply())
		},
	)
	if err != nil {
		return nil, err
	}
	return MapPVCs(pvcList.Items), nil
}

func ListPVs(ctx context.Context, clientset *kubernetes.Clientset, labelSelector, fieldSelector string) ([]model.PV, error) {
	opts := DefaultListOptions()
	opts.LabelSelector = labelSelector
	opts.FieldSelector = fieldSelector
	pvList, err := clientset.CoreV1().PersistentVolumes().List(ctx, opts.Apply())
	if err != nil {
		return nil, err
	}
	return MapPVs(pvList.Items), nil
}

func ListStorageClasses(ctx context.Context, clientset *kubernetes.Clientset, labelSelector, fieldSelector string) ([]model.StorageClass, error) {
	opts := DefaultListOptions()
	opts.LabelSelector = labelSelector
	opts.FieldSelector = fieldSelector
	scList, err := clientset.StorageV1().StorageClasses().List(ctx, opts.Apply())
	if err != nil {
		return nil, err
	}
	return MapStorageClasses(scList.Items), nil
}

func ListConfigMaps(ctx context.Context, clientset *kubernetes.Clientset, namespace, labelSelector, fieldSelector string) ([]model.ConfigMap, error) {
	opts := DefaultListOptions()
	opts.LabelSelector = labelSelector
	opts.FieldSelector = fieldSelector
	cmList, err := ListResourcesWithNamespace(ctx, namespace,
		func() (*corev1.ConfigMapList, error) {
			return clientset.CoreV1().ConfigMaps("").List(ctx, opts.Apply())
		},
		func(ns string) (*corev1.ConfigMapList, error) {
			return clientset.CoreV1().ConfigMaps(ns).List(ctx, opts.Apply())
		},
	)
	if err != nil {
		return nil, err
	}
	return MapConfigMaps(cmList.Items), nil
}

func ListSecrets(ctx context.Context, clientset *kubernetes.Clientset, namespace, labelSelector, fieldSelector string) ([]model.Secret, error) {
	opts := DefaultListOptions()
	opts.LabelSelector = labelSelector
	opts.FieldSelector = fieldSelector

	var secretList *corev1.SecretList
	var err error

	if namespace != "" {
		secretList, err = clientset.CoreV1().Secrets(namespace).List(ctx, opts.Apply())
	} else {
		secretList, err = clientset.CoreV1().Secrets("").List(ctx, opts.Apply())
	}

	if err != nil {
		return nil, err
	}
	return MapSecrets(secretList.Items), nil
}

func ListNodes(ctx context.Context, clientset *kubernetes.Clientset, pods *corev1.PodList, labelSelector, fieldSelector string) ([]model.Node, error) {
	opts := DefaultListOptions()
	opts.LabelSelector = labelSelector
	opts.FieldSelector = fieldSelector
	nodes, err := clientset.CoreV1().Nodes().List(ctx, opts.Apply())
	if err != nil {
		return nil, err
	}

	if pods == nil {
		pods, err = clientset.CoreV1().Pods("").List(ctx, metav1.ListOptions{})
		if err != nil {
			pods = nil
		}
	}

	return MapNodes(nodes.Items, pods), nil
}

func ListNamespaces(ctx context.Context, clientset *kubernetes.Clientset, labelSelector, fieldSelector string) ([]model.Namespace, error) {
	opts := DefaultListOptions()
	opts.LabelSelector = labelSelector
	opts.FieldSelector = fieldSelector
	nsList, err := clientset.CoreV1().Namespaces().List(ctx, opts.Apply())
	if err != nil {
		return nil, err
	}
	return MapNamespaces(nsList.Items), nil
}

func ListEvents(ctx context.Context, clientset *kubernetes.Clientset, namespace, involvedObject, since, labelSelector, fieldSelector string) ([]model.Event, error) {
	opts := DefaultListOptions()
	opts.LabelSelector = labelSelector

	fieldSelectors := []string{}
	if fieldSelector != "" {
		fieldSelectors = append(fieldSelectors, fieldSelector)
	}

	if involvedObject != "" {
		parts := strings.SplitN(involvedObject, "/", 2)
		if len(parts) == 2 {
			fieldSelectors = append(fieldSelectors,
				fmt.Sprintf("involvedObject.kind=%s", parts[0]),
				fmt.Sprintf("involvedObject.name=%s", parts[1]),
			)
		}
	}

	if len(fieldSelectors) > 0 {
		opts.FieldSelector = strings.Join(fieldSelectors, ",")
	}

	events, err := clientset.CoreV1().Events(namespace).List(ctx, opts.Apply())
	if err != nil {
		return nil, err
	}

	duration, err := time.ParseDuration(since)
	if err != nil {
		duration = 1 * time.Hour
	}

	cutoffTime := time.Now().Add(-duration)
	filteredEvents := make([]corev1.Event, 0, len(events.Items))
	for _, event := range events.Items {
		eventTime := event.LastTimestamp.Time
		if eventTime.IsZero() {
			eventTime = event.EventTime.Time
		}
		if !eventTime.IsZero() && eventTime.After(cutoffTime) {
			filteredEvents = append(filteredEvents, event)
		}
	}

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

	return MapEvents(filteredEvents), nil
}

func ListEndpoints(ctx context.Context, clientset *kubernetes.Clientset, namespace, labelSelector, fieldSelector string) ([]model.Endpoints, error) {
	opts := DefaultListOptions()
	opts.LabelSelector = labelSelector
	opts.FieldSelector = fieldSelector
	endpoints, err := clientset.CoreV1().Endpoints(namespace).List(ctx, opts.Apply())
	if err != nil {
		return nil, err
	}
	return MapEndpoints(endpoints.Items), nil
}
