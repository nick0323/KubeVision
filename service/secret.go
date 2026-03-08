package service

import (
	"context"

	"github.com/nick0323/K8sVision/model"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func ListSecrets(ctx context.Context, clientset *kubernetes.Clientset, namespace string) ([]model.SecretStatus, error) {
	secretList, err := ListResourcesWithNamespace(ctx, namespace,
		func() (*v1.SecretList, error) {
			return clientset.CoreV1().Secrets("").List(ctx, metav1.ListOptions{})
		},
		func(ns string) (*v1.SecretList, error) {
			return clientset.CoreV1().Secrets(ns).List(ctx, metav1.ListOptions{})
		},
	)
	if err != nil {
		return nil, err
	}

	secretStatuses := make([]model.SecretStatus, 0, len(secretList.Items))
	for _, secret := range secretList.Items {
		secretStatus := model.SecretStatus{
			Namespace: secret.Namespace,
			Name:      secret.Name,
			Type:      string(secret.Type),
			DataCount: len(secret.Data),
			Keys:      ExtractKeys(secret.Data),
			Age:       CalculateAge(secret.CreationTimestamp),
		}
		secretStatuses = append(secretStatuses, secretStatus)
	}

	return secretStatuses, nil
}
