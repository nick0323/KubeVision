package service

import (
	"context"

	"github.com/nick0323/K8sVision/model"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func ListConfigMaps(ctx context.Context, clientset *kubernetes.Clientset, namespace string) ([]model.ConfigMapStatus, error) {
	cmList, err := ListResourcesWithNamespace(ctx, namespace,
		func() (*v1.ConfigMapList, error) {
			return clientset.CoreV1().ConfigMaps("").List(ctx, metav1.ListOptions{})
		},
		func(ns string) (*v1.ConfigMapList, error) {
			return clientset.CoreV1().ConfigMaps(ns).List(ctx, metav1.ListOptions{})
		},
	)
	if err != nil {
		return nil, err
	}

	cmStatuses := make([]model.ConfigMapStatus, 0, len(cmList.Items))
	for _, cm := range cmList.Items {
		cmStatus := model.ConfigMapStatus{
			Namespace: cm.Namespace,
			Name:      cm.Name,
			DataCount: len(cm.Data),
			Keys:      ExtractKeys(cm.Data),
			Age:       CalculateAge(cm.CreationTimestamp),
		}
		cmStatuses = append(cmStatuses, cmStatus)
	}

	return cmStatuses, nil
}
