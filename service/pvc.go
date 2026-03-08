package service

import (
	"context"
	"strings"

	"github.com/nick0323/K8sVision/model"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func ListPVCs(ctx context.Context, clientset *kubernetes.Clientset, namespace string) ([]model.PVCStatus, error) {
	pvcList, err := ListResourcesWithNamespace(ctx, namespace,
		func() (*v1.PersistentVolumeClaimList, error) {
			return clientset.CoreV1().PersistentVolumeClaims("").List(ctx, metav1.ListOptions{})
		},
		func(ns string) (*v1.PersistentVolumeClaimList, error) {
			return clientset.CoreV1().PersistentVolumeClaims(ns).List(ctx, metav1.ListOptions{})
		},
	)
	if err != nil {
		return nil, err
	}

	pvcStatuses := make([]model.PVCStatus, 0, len(pvcList.Items))
	for _, pvc := range pvcList.Items {
		status := "Pending"
		if pvc.Status.Phase == v1.ClaimBound {
			status = "Bound"
		} else if pvc.Status.Phase == v1.ClaimLost {
			status = "Lost"
		}

		capacity := ""
		if pvc.Status.Capacity != nil {
			if storage, ok := pvc.Status.Capacity[v1.ResourceStorage]; ok {
				capacity = storage.String()
			}
		}

		accessModes := make([]string, 0)
		for _, mode := range pvc.Spec.AccessModes {
			accessModes = append(accessModes, string(mode))
		}

		storageClass := ""
		if pvc.Spec.StorageClassName != nil {
			storageClass = *pvc.Spec.StorageClassName
		}

		volumeName := ""
		if pvc.Spec.VolumeName != "" {
			volumeName = pvc.Spec.VolumeName
		}

		pvcStatus := model.PVCStatus{
			Namespace:    pvc.Namespace,
			Name:         pvc.Name,
			Status:       status,
			Capacity:     capacity,
			AccessMode:   strings.Join(accessModes, ","),
			StorageClass: storageClass,
			VolumeName:   volumeName,
			Age:          CalculateAge(pvc.CreationTimestamp),
		}
		pvcStatuses = append(pvcStatuses, pvcStatus)
	}

	return pvcStatuses, nil
}
