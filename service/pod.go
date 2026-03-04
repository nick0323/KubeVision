package service

import (
	"context"

	"github.com/nick0323/K8sVision/model"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func ListPodsWithRaw(ctx context.Context, clientset *kubernetes.Clientset, podMetricsMap model.PodMetricsMap, namespace string) ([]model.PodStatus, *v1.PodList, error) {
	pods, err := ListResourcesWithNamespace(ctx, namespace,
		func() (*v1.PodList, error) {
			return clientset.CoreV1().Pods("").List(ctx, metav1.ListOptions{})
		},
		func(ns string) (*v1.PodList, error) {
			return clientset.CoreV1().Pods(ns).List(ctx, metav1.ListOptions{})
		},
	)
	if err != nil {
		return nil, nil, err
	}

	// 使用通用映射函数
	podStatuses := MapPods(pods.Items, podMetricsMap)
	return podStatuses, pods, nil
}
