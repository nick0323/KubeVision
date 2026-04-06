package model

import (
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// Kubernetes 原生资源状态
const (
	PodPending   = "Pending"
	PodRunning   = "Running"
	PodSucceeded = "Succeeded"
	PodFailed    = "Failed"

	PVCBound = "Bound"
	PVCLost  = "Lost"
)

// 业务状态（前端展示用）
const (
	WorkloadAvailable    = "Available"
	WorkloadPartial      = "Partial"
	WorkloadUnavailable  = "Unavailable"
	WorkloadScaledToZero = "ScaledToZero"

	// Node 状态常量
	NodeReady    = "Ready"
	NodeNotReady = "NotReady"
	NodeUnknown  = "Unknown"
)

// Kubernetes 注解和标签常量
const (
	AnnotationStorageClassDefault = "storageclass.kubernetes.io/is-default-class"
	LabelNodeRolePrefix           = "node-role.kubernetes.io/"
)

// Kubernetes 资源名称常量
const (
	ResourceCPU    = "cpu"
	ResourceMemory = "memory"
	ResourcePods   = "pods"
)

// 时间格式常量
const (
	TimeFormatRFC3339 = time.RFC3339
	TimeFormatConsole = "2006-01-02 15:04:05"
)

// 分页配置常量
const (
	DefaultPageSize   = 15
	DefaultPageOffset = 0
	MaxPageSize       = 1000
)

// 密码策略常量
const (
	PasswordCharset    = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*"
	MinPasswordLen     = 8
	MaxPasswordLen     = 128
	DefaultPasswordLen = 12
)

// 缓存配置常量
const (
	CacheSampleSize = 10
)

// 响应配置常量
const (
	DefaultOverviewEventsLimit = 5
)

// Ingress 配置
var DefaultIngressPorts = []string{"80", "443"}

// 简化的错误码（只保留常用的）
const (
	CodeSuccess             = 0
	CodeBadRequest          = 400
	CodeUnauthorized        = 401
	CodeForbidden           = 403
	CodeNotFound            = 404
	CodeConflict            = 409
	CodeValidationFailed    = 422
	CodeInternalServerError = 500
)

// FormatTime 格式化 K8s 时间为 RFC3339 格式
func FormatTime(t *metav1.Time) string {
	if t == nil || t.IsZero() {
		return ""
	}
	return t.Time.Format(TimeFormatRFC3339)
}

// FormatConsoleTime 格式化时间为控制台显示格式
func FormatConsoleTime(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format(TimeFormatConsole)
}
