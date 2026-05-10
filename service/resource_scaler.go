package service

import (
	"context"
	"fmt"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"

	"github.com/nick0323/K8sVision/pkg/k8s"
)

var scalableTypes = map[k8s.ResourceType]bool{
	k8s.ResourceDeployment:  true,
	k8s.ResourceStatefulSet: true,
}

var restartableTypes = map[k8s.ResourceType]bool{
	k8s.ResourceDeployment:  true,
	k8s.ResourceStatefulSet: true,
	k8s.ResourceDaemonSet:   true,
}

func isScalable(rt k8s.ResourceType) bool {
	return scalableTypes[rt]
}

func ScaleResource(ctx context.Context, clientset kubernetes.Interface, resourceType, namespace, name string, replicas int32) error {
	rt := k8s.ResourceType(resourceType).Normalize()
	if !isScalable(rt) {
		return fmt.Errorf("unsupported resource type for scaling: %s", resourceType)
	}

	ns := namespace
	if rt.IsClusterScoped() {
		ns = ""
	}

	switch rt {
	case k8s.ResourceDeployment:
		dep, err := clientset.AppsV1().Deployments(ns).Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			return fmt.Errorf("failed to get deployment: %w", err)
		}
		dep.Spec.Replicas = &replicas
		_, err = clientset.AppsV1().Deployments(ns).Update(ctx, dep, metav1.UpdateOptions{})
		return err

	case k8s.ResourceStatefulSet:
		sts, err := clientset.AppsV1().StatefulSets(ns).Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			return fmt.Errorf("failed to get statefulset: %w", err)
		}
		sts.Spec.Replicas = &replicas
		_, err = clientset.AppsV1().StatefulSets(ns).Update(ctx, sts, metav1.UpdateOptions{})
		return err

	}

	return fmt.Errorf("unsupported resource type: %s", resourceType)
}

func RestartResource(ctx context.Context, clientset kubernetes.Interface, resourceType, namespace, name string) error {
	rt := k8s.ResourceType(resourceType).Normalize()
	if !isScalable(rt) {
		return fmt.Errorf("unsupported resource type for restart: %s", resourceType)
	}

	ns := namespace
	if rt.IsClusterScoped() {
		ns = ""
	}

	restartTime := time.Now().Format(time.RFC3339)

	switch rt {
	case k8s.ResourceDeployment:
		dep, err := clientset.AppsV1().Deployments(ns).Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			return fmt.Errorf("failed to get deployment: %w", err)
		}
		if dep.Spec.Template.Annotations == nil {
			dep.Spec.Template.Annotations = make(map[string]string)
		}
		dep.Spec.Template.Annotations["kubectl.kubernetes.io/restartedAt"] = restartTime
		_, err = clientset.AppsV1().Deployments(ns).Update(ctx, dep, metav1.UpdateOptions{})
		return err

	case k8s.ResourceStatefulSet:
		sts, err := clientset.AppsV1().StatefulSets(ns).Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			return fmt.Errorf("failed to get statefulset: %w", err)
		}
		if sts.Spec.Template.Annotations == nil {
			sts.Spec.Template.Annotations = make(map[string]string)
		}
		sts.Spec.Template.Annotations["kubectl.kubernetes.io/restartedAt"] = restartTime
		_, err = clientset.AppsV1().StatefulSets(ns).Update(ctx, sts, metav1.UpdateOptions{})
		return err

	case k8s.ResourceDaemonSet:
		ds, err := clientset.AppsV1().DaemonSets(ns).Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			return fmt.Errorf("failed to get daemonset: %w", err)
		}
		if ds.Spec.Template.Annotations == nil {
			ds.Spec.Template.Annotations = make(map[string]string)
		}
		ds.Spec.Template.Annotations["kubectl.kubernetes.io/restartedAt"] = restartTime
		_, err = clientset.AppsV1().DaemonSets(ns).Update(ctx, ds, metav1.UpdateOptions{})
		return err
	}

	return fmt.Errorf("unsupported resource type: %s", resourceType)
}

type ScaleRequest struct {
	Replicas int32 `json:"replicas" binding:"required,min=0"`
}
