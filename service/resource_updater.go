package service

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"

	appsv1 "k8s.io/api/apps/v1"
	autoscalingv2 "k8s.io/api/autoscaling/v2"
	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	policyv1 "k8s.io/api/policy/v1"
	rbacv1 "k8s.io/api/rbac/v1"
	storagev1 "k8s.io/api/storage/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/nick0323/K8sVision/pkg/k8s"
)

func UpdateResourceByType(ctx context.Context, clientset kubernetes.Interface, resourceType, namespace, name string, jsonBytes []byte) error {
	rt := k8s.ResourceType(resourceType).Normalize()
	updaters := k8s.NewUpdaters(clientset)

	updater, ok := updaters[rt]
	if !ok {
		return fmt.Errorf("unsupported resource type: %s", resourceType)
	}

	ns := namespace
	if rt.IsClusterScoped() {
		ns = ""
	}

	// 先获取最新资源，确保使用最新的 resourceVersion
	latestObj, err := GetResourceByName(ctx, clientset, resourceType, ns, name)
	if err != nil {
		return fmt.Errorf("failed to get latest resource: %w", err)
	}

	// 解析用户提交的更新
	obj, err := resourceFactory(resourceType)
	if err != nil {
		return err
	}

	if err := json.Unmarshal(jsonBytes, obj); err != nil {
		return fmt.Errorf("invalid %s object: %w", resourceType, err)
	}

	// 将最新资源的 resourceVersion 应用到更新对象上
	copyResourceVersion(latestObj, obj)

	return updater.Update(ctx, ns, name, obj)
}

// CreateResourceByType 创建资源
func CreateResourceByType(ctx context.Context, clientset kubernetes.Interface, resourceType string, jsonBytes []byte) error {
	rt := k8s.ResourceType(resourceType).Normalize()
	creators := k8s.NewCreators(clientset)

	creator, ok := creators[rt]
	if !ok {
		return fmt.Errorf("unsupported resource type: %s", resourceType)
	}

	// 创建资源时不需要 resourceVersion，使用专门的解析函数
	obj, err := unmarshalResourceForCreate(resourceType, jsonBytes)
	if err != nil {
		return err
	}

	// 获取 namespace（如果是集群资源则为空）
	ns := ""
	if !rt.IsClusterScoped() {
		// 从 obj 中获取 namespace
		val := reflect.ValueOf(obj)
		if val.Kind() == reflect.Ptr {
			val = val.Elem()
		}
		if val.Kind() == reflect.Struct {
			field := val.FieldByName("ObjectMeta")
			if field.IsValid() {
				nsField := field.FieldByName("Namespace")
				if nsField.IsValid() {
					ns = nsField.String()
				}
			}
		}
	}

	return creator.Create(ctx, ns, obj)
}

// resourceFactory creates a new instance of the specified resource type
func resourceFactory(resourceType string) (any, error) {
	switch resourceType {
	case "pod":
		return &v1.Pod{}, nil
	case "deployment":
		return &appsv1.Deployment{}, nil
	case "statefulset":
		return &appsv1.StatefulSet{}, nil
	case "daemonset":
		return &appsv1.DaemonSet{}, nil
	case "service":
		return &v1.Service{}, nil
	case "configmap":
		return &v1.ConfigMap{}, nil
	case "secret":
		return &v1.Secret{}, nil
	case "ingress":
		return &networkingv1.Ingress{}, nil
	case "job":
		return &batchv1.Job{}, nil
	case "cronjob":
		return &batchv1.CronJob{}, nil
	case "persistentvolumeclaim", "pvc":
		return &v1.PersistentVolumeClaim{}, nil
	case "persistentvolume", "pv":
		return &v1.PersistentVolume{}, nil
	case "storageclass":
		return &storagev1.StorageClass{}, nil
	case "namespace":
		return &v1.Namespace{}, nil
	case "node":
		return &v1.Node{}, nil
	case "endpoints":
		return &v1.Endpoints{}, nil
	case "events":
		return &v1.Event{}, nil
	case "networkpolicies", "networkpolicy":
		return &networkingv1.NetworkPolicy{}, nil
	case "serviceaccounts", "serviceaccount":
		return &v1.ServiceAccount{}, nil
	case "roles", "role":
		return &rbacv1.Role{}, nil
	case "rolebindings", "rolebinding":
		return &rbacv1.RoleBinding{}, nil
	case "clusterroles", "clusterrole":
		return &rbacv1.ClusterRole{}, nil
	case "clusterrolebindings", "clusterrolebinding":
		return &rbacv1.ClusterRoleBinding{}, nil
	case "resourcequotas", "resourcequota":
		return &v1.ResourceQuota{}, nil
	case "limitranges", "limitrange":
		return &v1.LimitRange{}, nil
	case "poddisruptionbudgets", "poddisruptionbudget":
		return &policyv1.PodDisruptionBudget{}, nil
	case "horizontalpodautoscalers", "horizontalpodautoscaler", "hpa":
		return &autoscalingv2.HorizontalPodAutoscaler{}, nil
	default:
		return nil, fmt.Errorf("unsupported resource type: %s", resourceType)
	}
}

// getResourceVersion extracts ResourceVersion from any K8s object
func getResourceVersion(obj any) string {
	val := reflect.ValueOf(obj)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	if val.Kind() != reflect.Struct {
		return ""
	}
	field := val.FieldByName("ObjectMeta")
	if !field.IsValid() {
		return ""
	}
	rv := field.FieldByName("ResourceVersion")
	if !rv.IsValid() {
		return ""
	}
	return rv.String()
}

func unmarshalResource(resourceType string, jsonBytes []byte) (any, error) {
	obj, err := resourceFactory(resourceType)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(jsonBytes, obj); err != nil {
		return nil, fmt.Errorf("invalid %s object: %v", resourceType, err)
	}

	resourceVersion := getResourceVersion(obj)
	if resourceVersion == "" {
		return nil, fmt.Errorf("missing required field: resourceVersion")
	}

	return obj, nil
}

// unmarshalResourceForCreate 解析资源用于创建（不需要 resourceVersion）
func unmarshalResourceForCreate(resourceType string, jsonBytes []byte) (any, error) {
	obj, err := resourceFactory(resourceType)
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal(jsonBytes, obj); err != nil {
		return nil, fmt.Errorf("invalid %s object: %v", resourceType, err)
	}

	// 创建资源时清除 resourceVersion（Kubernetes 会自动分配）
	clearResourceVersion(obj)

	return obj, nil
}

// clearResourceVersion 清除资源的 resourceVersion（用于创建新资源）
func clearResourceVersion(obj any) {
	val := reflect.ValueOf(obj)
	if val.Kind() == reflect.Ptr {
		val = val.Elem()
	}
	if val.Kind() != reflect.Struct {
		return
	}
	field := val.FieldByName("ObjectMeta")
	if !field.IsValid() {
		return
	}
	rv := field.FieldByName("ResourceVersion")
	if rv.IsValid() && rv.CanSet() {
		rv.SetString("")
	}
	// 也清除 UID 和 creationTimestamp（创建时不应包含）
	uid := field.FieldByName("UID")
	if uid.IsValid() && uid.CanSet() {
		uid.SetString("")
	}
	creationTimestamp := field.FieldByName("CreationTimestamp")
	if creationTimestamp.IsValid() && creationTimestamp.CanSet() {
		creationTimestamp.Set(reflect.Zero(creationTimestamp.Type()))
	}
}

// copyResourceVersion 将源资源的 resourceVersion 复制到目标资源（用于更新操作）
func copyResourceVersion(src, dst any) {
	srcVal := reflect.ValueOf(src)
	if srcVal.Kind() == reflect.Ptr {
		srcVal = srcVal.Elem()
	}
	if srcVal.Kind() != reflect.Struct {
		return
	}

	dstVal := reflect.ValueOf(dst)
	if dstVal.Kind() == reflect.Ptr {
		dstVal = dstVal.Elem()
	}
	if dstVal.Kind() != reflect.Struct {
		return
	}

	srcField := srcVal.FieldByName("ObjectMeta")
	if !srcField.IsValid() {
		return
	}

	dstField := dstVal.FieldByName("ObjectMeta")
	if !dstField.IsValid() {
		return
	}

	// 复制 resourceVersion
	srcRV := srcField.FieldByName("ResourceVersion")
	dstRV := dstField.FieldByName("ResourceVersion")
	if srcRV.IsValid() && dstRV.IsValid() && dstRV.CanSet() {
		dstRV.Set(srcRV)
	}

	// 复制 UID
	srcUID := srcField.FieldByName("UID")
	dstUID := dstField.FieldByName("UID")
	if srcUID.IsValid() && dstUID.IsValid() && dstUID.CanSet() {
		dstUID.Set(srcUID)
	}
}
