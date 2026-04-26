package service

import (
	"context"
	"fmt"
	"strings"

	"github.com/nick0323/K8sVision/model"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// GetResourceByName 根据资源类型和名称获取对象（返回 K8s 原始对象）
func GetResourceByName(ctx context.Context, clientset kubernetes.Interface, resourceType, namespace, name string) (interface{}, error) {
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
		return nil, fmt.Errorf("unsupported resource type: %s", resourceType)
	}
}

// DeleteResourceByType 根据资源类型删除资源
func DeleteResourceByType(ctx context.Context, clientset kubernetes.Interface, resourceType, namespace, name string) error {
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
		return fmt.Errorf("unsupported resource type: %s", resourceType)
	}
}

// ListResourcesByType 根据资源类型获取列表
func ListResourcesByType(ctx context.Context, clientset kubernetes.Interface, resourceType, namespace, labelSelector, fieldSelector, involvedObject, since string) ([]model.SearchableItem, error) {
	switch resourceType {
	case "pod":
		pods, err := ListPods(ctx, clientset, namespace, labelSelector, fieldSelector)
		if err != nil {
			return nil, err
		}
		result := make([]model.SearchableItem, len(pods))
		for i := range pods {
			result[i] = &pods[i]
		}
		return result, nil

	case "deployment":
		deployments, err := ListDeployments(ctx, clientset, namespace, labelSelector, fieldSelector)
		if err != nil {
			return nil, err
		}
		result := make([]model.SearchableItem, len(deployments))
		for i := range deployments {
			result[i] = &deployments[i]
		}
		return result, nil

	case "statefulset":
		statefulSets, err := ListStatefulSets(ctx, clientset, namespace, labelSelector, fieldSelector)
		if err != nil {
			return nil, err
		}
		result := make([]model.SearchableItem, len(statefulSets))
		for i := range statefulSets {
			result[i] = &statefulSets[i]
		}
		return result, nil

	case "daemonset":
		daemonSets, err := ListDaemonSets(ctx, clientset, namespace, labelSelector, fieldSelector)
		if err != nil {
			return nil, err
		}
		result := make([]model.SearchableItem, len(daemonSets))
		for i := range daemonSets {
			result[i] = &daemonSets[i]
		}
		return result, nil

	case "service":
		services, err := ListServices(ctx, clientset, namespace, labelSelector, fieldSelector)
		if err != nil {
			return nil, err
		}
		result := make([]model.SearchableItem, len(services))
		for i := range services {
			result[i] = &services[i]
		}
		return result, nil

	case "configmap":
		configMaps, err := ListConfigMaps(ctx, clientset, namespace, labelSelector, fieldSelector)
		if err != nil {
			return nil, err
		}
		result := make([]model.SearchableItem, len(configMaps))
		for i := range configMaps {
			result[i] = &configMaps[i]
		}
		return result, nil

	case "secret":
		secrets, err := ListSecrets(ctx, clientset, namespace, labelSelector, fieldSelector)
		if err != nil {
			return nil, err
		}
		result := make([]model.SearchableItem, len(secrets))
		for i := range secrets {
			result[i] = &secrets[i]
		}
		return result, nil

	case "ingress":
		ingresses, err := ListIngresses(ctx, clientset, namespace, labelSelector, fieldSelector)
		if err != nil {
			return nil, err
		}
		result := make([]model.SearchableItem, len(ingresses))
		for i := range ingresses {
			result[i] = &ingresses[i]
		}
		return result, nil

	case "job":
		jobs, err := ListJobs(ctx, clientset, namespace, labelSelector, fieldSelector)
		if err != nil {
			return nil, err
		}
		result := make([]model.SearchableItem, len(jobs))
		for i := range jobs {
			result[i] = &jobs[i]
		}
		return result, nil

	case "cronjob":
		cronJobs, err := ListCronJobs(ctx, clientset, namespace, labelSelector, fieldSelector)
		if err != nil {
			return nil, err
		}
		result := make([]model.SearchableItem, len(cronJobs))
		for i := range cronJobs {
			result[i] = &cronJobs[i]
		}
		return result, nil

	case "persistentvolumeclaim", "pvc":
		pvcs, err := ListPVCs(ctx, clientset, namespace, labelSelector, fieldSelector)
		if err != nil {
			return nil, err
		}
		result := make([]model.SearchableItem, len(pvcs))
		for i := range pvcs {
			result[i] = &pvcs[i]
		}
		return result, nil

	case "persistentvolume", "pv":
		pvs, err := ListPVs(ctx, clientset, labelSelector, fieldSelector)
		if err != nil {
			return nil, err
		}
		result := make([]model.SearchableItem, len(pvs))
		for i := range pvs {
			result[i] = &pvs[i]
		}
		return result, nil

	case "storageclass":
		storageClasses, err := ListStorageClasses(ctx, clientset, labelSelector, fieldSelector)
		if err != nil {
			return nil, err
		}
		result := make([]model.SearchableItem, len(storageClasses))
		for i := range storageClasses {
			result[i] = &storageClasses[i]
		}
		return result, nil

	case "namespace":
		namespaces, err := ListNamespaces(ctx, clientset, labelSelector, fieldSelector)
		if err != nil {
			return nil, err
		}
		result := make([]model.SearchableItem, len(namespaces))
		for i := range namespaces {
			result[i] = &namespaces[i]
		}
		return result, nil

	case "node":
		nodes, err := ListNodes(ctx, clientset, nil, labelSelector, fieldSelector)
		if err != nil {
			return nil, err
		}
		result := make([]model.SearchableItem, len(nodes))
		for i := range nodes {
			result[i] = &nodes[i]
		}
		return result, nil

	case "endpoint":
		endpoints, err := ListEndpoints(ctx, clientset, namespace, labelSelector, fieldSelector)
		if err != nil {
			return nil, err
		}
		result := make([]model.SearchableItem, len(endpoints))
		for i := range endpoints {
			result[i] = &endpoints[i]
		}
		return result, nil

	case "event":
		events, err := ListEvents(ctx, clientset, namespace, involvedObject, since, labelSelector, fieldSelector)
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
