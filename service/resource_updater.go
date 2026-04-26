package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	storagev1 "k8s.io/api/storage/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// UpdateResourceByType 根据资源类型更新资源
// 注意：K8s Update 操作需要 resourceVersion 字段
func UpdateResourceByType(ctx context.Context, clientset kubernetes.Interface, resourceType, namespace, name string, jsonBytes []byte) error {
	resourceType = strings.ToLower(resourceType)

	switch resourceType {
	case "pod":
		pod := &v1.Pod{}
		if err := json.Unmarshal(jsonBytes, pod); err != nil {
			return fmt.Errorf("invalid Pod object: %v", err)
		}
		// 检查必需字段
		if pod.ResourceVersion == "" {
			return fmt.Errorf("missing required field: resourceVersion")
		}
		_, err := clientset.CoreV1().Pods(namespace).Update(ctx, pod, metav1.UpdateOptions{})
		return err
	case "deployment":
		dep := &appsv1.Deployment{}
		if err := json.Unmarshal(jsonBytes, dep); err != nil {
			return fmt.Errorf("invalid Deployment object: %v", err)
		}
		if dep.ResourceVersion == "" {
			return fmt.Errorf("missing required field: resourceVersion")
		}
		_, err := clientset.AppsV1().Deployments(namespace).Update(ctx, dep, metav1.UpdateOptions{})
		return err
	case "statefulset":
		sts := &appsv1.StatefulSet{}
		if err := json.Unmarshal(jsonBytes, sts); err != nil {
			return fmt.Errorf("invalid StatefulSet object: %v", err)
		}
		if sts.ResourceVersion == "" {
			return fmt.Errorf("missing required field: resourceVersion")
		}
		_, err := clientset.AppsV1().StatefulSets(namespace).Update(ctx, sts, metav1.UpdateOptions{})
		return err
	case "daemonset":
		ds := &appsv1.DaemonSet{}
		if err := json.Unmarshal(jsonBytes, ds); err != nil {
			return fmt.Errorf("invalid DaemonSet object: %v", err)
		}
		if ds.ResourceVersion == "" {
			return fmt.Errorf("missing required field: resourceVersion")
		}
		_, err := clientset.AppsV1().DaemonSets(namespace).Update(ctx, ds, metav1.UpdateOptions{})
		return err
	case "service":
		svc := &v1.Service{}
		if err := json.Unmarshal(jsonBytes, svc); err != nil {
			return fmt.Errorf("invalid Service object: %v", err)
		}
		if svc.ResourceVersion == "" {
			return fmt.Errorf("missing required field: resourceVersion")
		}
		_, err := clientset.CoreV1().Services(namespace).Update(ctx, svc, metav1.UpdateOptions{})
		return err
	case "configmap":
		cm := &v1.ConfigMap{}
		if err := json.Unmarshal(jsonBytes, cm); err != nil {
			return fmt.Errorf("invalid ConfigMap object: %v", err)
		}
		if cm.ResourceVersion == "" {
			return fmt.Errorf("missing required field: resourceVersion")
		}
		_, err := clientset.CoreV1().ConfigMaps(namespace).Update(ctx, cm, metav1.UpdateOptions{})
		return err
	case "secret":
		secret := &v1.Secret{}
		if err := json.Unmarshal(jsonBytes, secret); err != nil {
			return fmt.Errorf("invalid Secret object: %v", err)
		}
		if secret.ResourceVersion == "" {
			return fmt.Errorf("missing required field: resourceVersion")
		}
		_, err := clientset.CoreV1().Secrets(namespace).Update(ctx, secret, metav1.UpdateOptions{})
		return err
	case "ingress":
		ing := &networkingv1.Ingress{}
		if err := json.Unmarshal(jsonBytes, ing); err != nil {
			return fmt.Errorf("invalid Ingress object: %v", err)
		}
		if ing.ResourceVersion == "" {
			return fmt.Errorf("missing required field: resourceVersion")
		}
		_, err := clientset.NetworkingV1().Ingresses(namespace).Update(ctx, ing, metav1.UpdateOptions{})
		return err
	case "job":
		job := &batchv1.Job{}
		if err := json.Unmarshal(jsonBytes, job); err != nil {
			return fmt.Errorf("invalid Job object: %v", err)
		}
		if job.ResourceVersion == "" {
			return fmt.Errorf("missing required field: resourceVersion")
		}
		_, err := clientset.BatchV1().Jobs(namespace).Update(ctx, job, metav1.UpdateOptions{})
		return err
	case "cronjob":
		cj := &batchv1.CronJob{}
		if err := json.Unmarshal(jsonBytes, cj); err != nil {
			return fmt.Errorf("invalid CronJob object: %v", err)
		}
		if cj.ResourceVersion == "" {
			return fmt.Errorf("missing required field: resourceVersion")
		}
		_, err := clientset.BatchV1().CronJobs(namespace).Update(ctx, cj, metav1.UpdateOptions{})
		return err
	case "persistentvolumeclaim", "pvc":
		pvc := &v1.PersistentVolumeClaim{}
		if err := json.Unmarshal(jsonBytes, pvc); err != nil {
			return fmt.Errorf("invalid PVC object: %v", err)
		}
		if pvc.ResourceVersion == "" {
			return fmt.Errorf("missing required field: resourceVersion")
		}
		_, err := clientset.CoreV1().PersistentVolumeClaims(namespace).Update(ctx, pvc, metav1.UpdateOptions{})
		return err
	case "persistentvolume", "pv":
		pv := &v1.PersistentVolume{}
		if err := json.Unmarshal(jsonBytes, pv); err != nil {
			return fmt.Errorf("invalid PV object: %v", err)
		}
		if pv.ResourceVersion == "" {
			return fmt.Errorf("missing required field: resourceVersion")
		}
		_, err := clientset.CoreV1().PersistentVolumes().Update(ctx, pv, metav1.UpdateOptions{})
		return err
	case "storageclass":
		sc := &storagev1.StorageClass{}
		if err := json.Unmarshal(jsonBytes, sc); err != nil {
			return fmt.Errorf("invalid StorageClass object: %v", err)
		}
		if sc.ResourceVersion == "" {
			return fmt.Errorf("missing required field: resourceVersion")
		}
		_, err := clientset.StorageV1().StorageClasses().Update(ctx, sc, metav1.UpdateOptions{})
		return err
	case "namespace":
		ns := &v1.Namespace{}
		if err := json.Unmarshal(jsonBytes, ns); err != nil {
			return fmt.Errorf("invalid Namespace object: %v", err)
		}
		if ns.ResourceVersion == "" {
			return fmt.Errorf("missing required field: resourceVersion")
		}
		_, err := clientset.CoreV1().Namespaces().Update(ctx, ns, metav1.UpdateOptions{})
		return err
	case "node":
		node := &v1.Node{}
		if err := json.Unmarshal(jsonBytes, node); err != nil {
			return fmt.Errorf("invalid Node object: %v", err)
		}
		if node.ResourceVersion == "" {
			return fmt.Errorf("missing required field: resourceVersion")
		}
		_, err := clientset.CoreV1().Nodes().Update(ctx, node, metav1.UpdateOptions{})
		return err
	default:
		return fmt.Errorf("unsupported resource type: %s", resourceType)
	}
}
