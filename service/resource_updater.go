package service

import (
	"context"
	"encoding/json"
	"fmt"

	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
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

	obj, err := unmarshalResource(resourceType, jsonBytes)
	if err != nil {
		return err
	}

	return updater.Update(ctx, ns, name, obj)
}

func unmarshalResource(resourceType string, jsonBytes []byte) (interface{}, error) {
	switch resourceType {
	case "pod":
		obj := &v1.Pod{}
		if err := json.Unmarshal(jsonBytes, obj); err != nil {
			return nil, fmt.Errorf("invalid Pod object: %v", err)
		}
		if obj.ResourceVersion == "" {
			return nil, fmt.Errorf("missing required field: resourceVersion")
		}
		return obj, nil

	case "deployment":
		obj := &appsv1.Deployment{}
		if err := json.Unmarshal(jsonBytes, obj); err != nil {
			return nil, fmt.Errorf("invalid Deployment object: %v", err)
		}
		if obj.ResourceVersion == "" {
			return nil, fmt.Errorf("missing required field: resourceVersion")
		}
		return obj, nil

	case "statefulset":
		obj := &appsv1.StatefulSet{}
		if err := json.Unmarshal(jsonBytes, obj); err != nil {
			return nil, fmt.Errorf("invalid StatefulSet object: %v", err)
		}
		if obj.ResourceVersion == "" {
			return nil, fmt.Errorf("missing required field: resourceVersion")
		}
		return obj, nil

	case "daemonset":
		obj := &appsv1.DaemonSet{}
		if err := json.Unmarshal(jsonBytes, obj); err != nil {
			return nil, fmt.Errorf("invalid DaemonSet object: %v", err)
		}
		if obj.ResourceVersion == "" {
			return nil, fmt.Errorf("missing required field: resourceVersion")
		}
		return obj, nil

	case "service":
		obj := &v1.Service{}
		if err := json.Unmarshal(jsonBytes, obj); err != nil {
			return nil, fmt.Errorf("invalid Service object: %v", err)
		}
		if obj.ResourceVersion == "" {
			return nil, fmt.Errorf("missing required field: resourceVersion")
		}
		return obj, nil

	case "configmap":
		obj := &v1.ConfigMap{}
		if err := json.Unmarshal(jsonBytes, obj); err != nil {
			return nil, fmt.Errorf("invalid ConfigMap object: %v", err)
		}
		if obj.ResourceVersion == "" {
			return nil, fmt.Errorf("missing required field: resourceVersion")
		}
		return obj, nil

	case "secret":
		obj := &v1.Secret{}
		if err := json.Unmarshal(jsonBytes, obj); err != nil {
			return nil, fmt.Errorf("invalid Secret object: %v", err)
		}
		if obj.ResourceVersion == "" {
			return nil, fmt.Errorf("missing required field: resourceVersion")
		}
		return obj, nil

	case "ingress":
		obj := &networkingv1.Ingress{}
		if err := json.Unmarshal(jsonBytes, obj); err != nil {
			return nil, fmt.Errorf("invalid Ingress object: %v", err)
		}
		if obj.ResourceVersion == "" {
			return nil, fmt.Errorf("missing required field: resourceVersion")
		}
		return obj, nil

	case "job":
		obj := &batchv1.Job{}
		if err := json.Unmarshal(jsonBytes, obj); err != nil {
			return nil, fmt.Errorf("invalid Job object: %v", err)
		}
		if obj.ResourceVersion == "" {
			return nil, fmt.Errorf("missing required field: resourceVersion")
		}
		return obj, nil

	case "cronjob":
		obj := &batchv1.CronJob{}
		if err := json.Unmarshal(jsonBytes, obj); err != nil {
			return nil, fmt.Errorf("invalid CronJob object: %v", err)
		}
		if obj.ResourceVersion == "" {
			return nil, fmt.Errorf("missing required field: resourceVersion")
		}
		return obj, nil

	case "persistentvolumeclaim", "pvc":
		obj := &v1.PersistentVolumeClaim{}
		if err := json.Unmarshal(jsonBytes, obj); err != nil {
			return nil, fmt.Errorf("invalid PVC object: %v", err)
		}
		if obj.ResourceVersion == "" {
			return nil, fmt.Errorf("missing required field: resourceVersion")
		}
		return obj, nil

	case "persistentvolume", "pv":
		obj := &v1.PersistentVolume{}
		if err := json.Unmarshal(jsonBytes, obj); err != nil {
			return nil, fmt.Errorf("invalid PV object: %v", err)
		}
		if obj.ResourceVersion == "" {
			return nil, fmt.Errorf("missing required field: resourceVersion")
		}
		return obj, nil

	case "storageclass":
		obj := &storagev1.StorageClass{}
		if err := json.Unmarshal(jsonBytes, obj); err != nil {
			return nil, fmt.Errorf("invalid StorageClass object: %v", err)
		}
		if obj.ResourceVersion == "" {
			return nil, fmt.Errorf("missing required field: resourceVersion")
		}
		return obj, nil

	case "namespace":
		obj := &v1.Namespace{}
		if err := json.Unmarshal(jsonBytes, obj); err != nil {
			return nil, fmt.Errorf("invalid Namespace object: %v", err)
		}
		if obj.ResourceVersion == "" {
			return nil, fmt.Errorf("missing required field: resourceVersion")
		}
		return obj, nil

	case "node":
		obj := &v1.Node{}
		if err := json.Unmarshal(jsonBytes, obj); err != nil {
			return nil, fmt.Errorf("invalid Node object: %v", err)
		}
		if obj.ResourceVersion == "" {
			return nil, fmt.Errorf("missing required field: resourceVersion")
		}
		return obj, nil

	default:
		return nil, fmt.Errorf("unsupported resource type: %s", resourceType)
	}
}
