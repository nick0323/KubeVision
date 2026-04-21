package model

import (
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

type NodeMetrics struct {
	CPU    string `json:"cpu"`
	Memory string `json:"memory"`
}

const (
	Version         = "2.0.0-optimized"
	MonitorInterval = 5 * time.Minute

	HealthCheckPath = "/health"
	CacheStatsPath  = "/cache/stats"
	APIPrefix       = "/api"
	LoginPath       = "/api/login"

	WorkloadAvailable   = "Available"
	WorkloadPartial     = "Partial"
	WorkloadUnavailable = "Unavailable"

	NodeReady    = "Ready"
	NodeNotReady = "NotReady"
	NodeUnknown  = "Unknown"

	ResourceCPU    = "cpu"
	ResourceMemory = "memory"
	ResourcePods   = "pods"

	TimeFormatRFC3339 = time.RFC3339

	DefaultPageSize   = 15
	DefaultPageOffset = 0
	MaxPageSize       = 1000

	PasswordCharset    = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*"
	MinPasswordLen     = 8
	MaxPasswordLen     = 128
	DefaultPasswordLen = 12

	DefaultOverviewEventsLimit = 5

	CodeSuccess             = 0
	CodeBadRequest          = 400
	CodeUnauthorized        = 401
	CodeValidationFailed    = 422
	CodeInternalServerError = 500
)

func FormatTime(t *metav1.Time) string {
	if t == nil || t.IsZero() {
		return ""
	}
	return t.Time.Format(TimeFormatRFC3339)
}
