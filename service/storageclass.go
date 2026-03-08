package service

import (
	"context"

	"github.com/nick0323/K8sVision/model"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func ListStorageClasses(ctx context.Context, clientset *kubernetes.Clientset) ([]model.StorageClassStatus, error) {
	scList, err := clientset.StorageV1().StorageClasses().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	var scStatuses []model.StorageClassStatus
	for _, sc := range scList.Items {
		reclaimPolicy := string(*sc.ReclaimPolicy)

		volumeBindingMode := string(*sc.VolumeBindingMode)

		isDefault := false
		if sc.Annotations != nil {
			if _, ok := sc.Annotations[model.AnnotationStorageClassDefault]; ok {
				isDefault = true
			}
		}

		scStatus := model.StorageClassStatus{
			Name:              sc.Name,
			Provisioner:       sc.Provisioner,
			ReclaimPolicy:     reclaimPolicy,
			VolumeBindingMode: volumeBindingMode,
			IsDefault:         isDefault,
			Age:               CalculateAge(sc.CreationTimestamp),
		}
		scStatuses = append(scStatuses, scStatus)
	}

	return scStatuses, nil
}
