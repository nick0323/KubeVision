package service

import (
	"context"
	"strings"

	"github.com/nick0323/K8sVision/model"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func ListPVs(ctx context.Context, clientset *kubernetes.Clientset) ([]model.PVStatus, error) {
	pvList, err := clientset.CoreV1().PersistentVolumes().List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	var pvStatuses []model.PVStatus
	for _, pv := range pvList.Items {
		status := string(pv.Status.Phase)

		capacity := ""
		if pv.Spec.Capacity != nil {
			if storage, ok := pv.Spec.Capacity[v1.ResourceStorage]; ok {
				capacity = storage.String()
			}
		}

		accessModes := make([]string, 0)
		for _, mode := range pv.Spec.AccessModes {
			accessModes = append(accessModes, string(mode))
		}

		storageClass := ""
		if pv.Spec.StorageClassName != "" {
			storageClass = pv.Spec.StorageClassName
		}

		claimRef := ""
		if pv.Spec.ClaimRef != nil {
			claimRef = pv.Spec.ClaimRef.Namespace + "/" + pv.Spec.ClaimRef.Name
		}

		reclaimPolicy := string(pv.Spec.PersistentVolumeReclaimPolicy)

		pvStatus := model.PVStatus{
			Name:          pv.Name,
			Status:        status,
			Capacity:      capacity,
			AccessMode:    strings.Join(accessModes, ","),
			StorageClass:  storageClass,
			ClaimRef:      claimRef,
			ReclaimPolicy: reclaimPolicy,
			Age:           CalculateAge(pv.CreationTimestamp),
		}
		pvStatuses = append(pvStatuses, pvStatus)
	}

	return pvStatuses, nil
}
