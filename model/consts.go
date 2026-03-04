package model

import (
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ==================== 资源状态常量 ====================
const (
	// 通用状态
	StatusActive       = "Active"
	StatusSuspended    = "Suspended"
	StatusHealthy      = "Healthy"
	StatusAbnormal     = "Abnormal"
	StatusPartial      = "PartialAvailable"
	StatusReady        = "Ready"
	StatusNotReady     = "Not Ready"
	StatusScaledToZero = "Scaled to zero"

	// PVC 状态
	StatusBound = "Bound"
	StatusLost  = "Lost"
)

// ==================== 时间格式常量 ====================
const (
	TimeFormat = "2006-01-02 15:04:05"
)

// ==================== Kubernetes 注解和标签常量 ====================
const (
	// StorageClass 注解
	AnnotationStorageClassDefault = "storageclass.kubernetes.io/is-default-class"

	// Node 角色标签前缀
	LabelNodeRolePrefix = "node-role.kubernetes.io/"
)

// ==================== Kubernetes 资源名称常量 ====================
const (
	ResourceCPU     = "cpu"
	ResourceMemory  = "memory"
	ResourceStorage = "storage"
	ResourcePods    = "pods"
)

// ==================== HTTP 状态码常量 ====================
const (
	HTTPStatusOK                  = 200
	HTTPStatusBadRequest          = 400
	HTTPStatusUnauthorized        = 401
	HTTPStatusForbidden           = 403
	HTTPStatusNotFound            = 404
	HTTPStatusInternalServerError = 500
	HTTPStatusServiceUnavailable  = 503
)

// ==================== 默认配置常量 ====================
const (
	DefaultPageSize     = 20
	DefaultPageOffset   = 0
	DefaultCacheTTL     = 300
	DefaultJWTSecretLen = 32
	DefaultPasswordLen  = 12
	MaxPasswordLen      = 128
	MinPasswordLen      = 8
	DefaultOverviewEventsLimit = 5
)

// ==================== 缓存键前缀常量 ====================
const (
	CacheKeyPrefixK8sClient = "k8s_client_"
	CacheKeyPrefixResource  = "resource_"
	CacheKeyPrefixMetrics   = "metrics_"
)

// ==================== 日志配置常量 ====================
const (
	LogLevelDebug = "debug"
	LogLevelInfo  = "info"
	LogLevelWarn  = "warn"
	LogLevelError = "error"
)

// ==================== 日志格式常量 ====================
const (
	LogFormatConsole = "console"
	LogFormatJSON    = "json"
)

// ==================== 业务错误码常量 ====================
const (
	CodeSuccess = 0

	// 客户端错误 (1000-1999)
	CodeBadRequest       = 1000
	CodeUnauthorized     = 1001
	CodeForbidden        = 1002
	CodeNotFound         = 1003
	CodeMethodNotAllowed = 1004
	CodeRequestTimeout   = 1005
	CodeConflict         = 1006
	CodeValidationFailed = 1007
	CodeInvalidParameter = 1008
	CodeMissingParameter = 1009

	// 服务器错误 (2000-2999)
	CodeInternalServerError = 2000
	CodeServiceUnavailable  = 2001
	CodeDatabaseError       = 2002
	CodeK8sClientError      = 2003
	CodeK8sAPIError         = 2004
	CodeConfigError         = 2005
	CodeAuthError           = 2006

	// 资源错误 (3000-3999)
	CodeResourceNotFound    = 3000
	CodeResourceExists      = 3001
	CodeResourceInUse       = 3002
	CodeResourceQuotaExceed = 3003
	CodePermissionDenied    = 3004
)

// ==================== Kubernetes 端口常量 ====================
const (
	PortHTTP  = 80
	PortHTTPS = 443
)

// ==================== Ingress 默认端口 ====================
var DefaultIngressPorts = []string{"80", "443"}

// ==================== 密码生成常量 ====================
const (
	PasswordCharset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%^&*"
)

// ==================== 缓存采样大小 ====================
const (
	CacheSampleSize = 10
)

var ErrorMessages = map[int]string{
	CodeSuccess:             "操作成功",
	CodeBadRequest:          "请求参数错误",
	CodeUnauthorized:        "未授权访问",
	CodeForbidden:           "访问被禁止",
	CodeNotFound:            "资源不存在",
	CodeMethodNotAllowed:    "请求方法不允许",
	CodeRequestTimeout:      "请求超时",
	CodeConflict:            "资源冲突",
	CodeValidationFailed:    "数据验证失败",
	CodeInvalidParameter:    "无效参数",
	CodeMissingParameter:    "缺少必要参数",
	CodeInternalServerError: "服务器内部错误",
	CodeServiceUnavailable:  "服务不可用",
	CodeDatabaseError:       "数据库错误",
	CodeK8sClientError:      "Kubernetes客户端错误",
	CodeK8sAPIError:         "Kubernetes API错误",
	CodeConfigError:         "配置错误",
	CodeAuthError:           "认证错误",
	CodeResourceNotFound:    "资源不存在",
	CodeResourceExists:      "资源已存在",
	CodeResourceInUse:       "资源正在使用中",
	CodeResourceQuotaExceed: "资源配额超限",
	CodePermissionDenied:    "权限不足",
}

func GetErrorMessage(code int) string {
	if msg, exists := ErrorMessages[code]; exists {
		return msg
	}
	return "未知错误"
}

func FormatTime(t *metav1.Time) string {
	if t == nil || t.IsZero() {
		return ""
	}
	return t.Time.Format(TimeFormat)
}

func FormatTimeValue(t time.Time) string {
	if t.IsZero() {
		return ""
	}
	return t.Format(TimeFormat)
}
