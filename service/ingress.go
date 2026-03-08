package service

import (
	"context"

	"github.com/nick0323/K8sVision/model"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func ListIngresses(ctx context.Context, clientset *kubernetes.Clientset, namespace string) ([]model.IngressStatus, error) {
	ingresses, err := clientset.NetworkingV1().Ingresses(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	// 使用通用映射函数
	result := MapIngresses(ingresses.Items)
	return result, nil
}
