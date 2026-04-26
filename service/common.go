package service

import (
	"github.com/nick0323/K8sVision/model"
)

func GetWorkloadStatus(ready, desired int32) string {
	if ready == desired && desired > 0 {
		return model.WorkloadAvailable
	} else if ready > 0 {
		return model.WorkloadPartial
	}
	return model.WorkloadUnavailable
}
