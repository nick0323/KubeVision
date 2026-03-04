package service

import (
	"context"

	"github.com/nick0323/K8sVision/model"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func ListServices(ctx context.Context, clientset *kubernetes.Clientset, namespace string) ([]model.ServiceStatus, error) {
	svcs, err := clientset.CoreV1().Services(namespace).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	
	// 使用通用映射函数
	result := MapServices(svcs.Items)
	return result, nil
}
