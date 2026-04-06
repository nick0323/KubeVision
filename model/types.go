package model

import (
	"errors"
	"fmt"
	"time"
)

// ==================== API 响应结构 ====================

// APIResponse 统一 API 响应结构
type APIResponse struct {
	Code      int       `json:"code"`
	Message   string    `json:"message"`
	Data      any       `json:"data"`
	TraceID   string    `json:"traceId,omitempty"`
	Timestamp int64     `json:"timestamp"`
	Page      *PageMeta `json:"page,omitempty"`
}

// PageMeta 分页元数据
type PageMeta struct {
	Total  int `json:"total"`
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
}

// APIError API 错误结构
type APIError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Details any    `json:"details,omitempty"`
}

// Error 实现 error 接口
func (e *APIError) Error() string {
	return e.Message
}

// NewAPIError 创建 API 错误
func NewAPIError(code int, message string) *APIError {
	return &APIError{
		Code:    code,
		Message: message,
	}
}

// NewAPIErrorWithDetails 创建带详细信息的 API 错误
func NewAPIErrorWithDetails(code int, message string, details any) *APIError {
	return &APIError{
		Code:    code,
		Message: message,
		Details: details,
	}
}

// ==================== 通用请求结构 ====================

// LoginRequest 登录请求
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// Validate 验证登录请求
func (r *LoginRequest) Validate() error {
	if r.Username == "" {
		return errors.New("用户名不能为空")
	}
	if r.Password == "" {
		return errors.New("密码不能为空")
	}
	if len(r.Username) < 3 {
		return errors.New("用户名至少 3 个字符")
	}
	if len(r.Password) < 6 {
		return errors.New("密码至少 6 个字符")
	}
	return nil
}

// ChangePasswordRequest 修改密码请求
type ChangePasswordRequest struct {
	OldPassword string `json:"oldPassword" binding:"required"`
	NewPassword string `json:"newPassword" binding:"required,min=8,max=128"`
}

// TokenResponse Token 响应
type TokenResponse struct {
	AccessToken  string `json:"accessToken"`
	RefreshToken string `json:"refreshToken,omitempty"`
	ExpiresIn    int64  `json:"expiresIn"`
	TokenType    string `json:"tokenType"`
}

// ==================== 指标类型 ====================

// PodMetrics Pod 指标
type PodMetrics struct {
	CPU        string  `json:"cpu"`                  // 如 "123m"
	Memory     string  `json:"memory"`               // 如 "512Mi"
	CPUNumeric float64 `json:"cpuNumeric,omitempty"` // CPU 毫核数
	MemNumeric int64   `json:"memNumeric,omitempty"` // 内存字节数
}

// NodeMetrics Node 指标
type NodeMetrics struct {
	CPU         string  `json:"cpu"`
	Memory      string  `json:"memory"`
	CPUUsage    float64 `json:"cpuUsage,omitempty"`    // CPU 使用率 0-100
	MemoryUsage float64 `json:"memoryUsage,omitempty"` // 内存使用率 0-100
}

// ==================== 集群概览 ====================

// OverviewStatus 集群概览
type OverviewStatus struct {
	NodeCount      int     `json:"nodeCount"`
	NodeReady      int     `json:"nodeReady"`
	PodCount       int     `json:"podCount"`
	PodNotReady    int     `json:"podNotReady"`
	NamespaceCount int     `json:"namespaceCount"`
	ServiceCount   int     `json:"serviceCount"`
	CPUCapacity    float64 `json:"cpuCapacity"`
	CPURequests    float64 `json:"cpuRequests"`
	CPULimits      float64 `json:"cpuLimits"`
	MemoryCapacity float64 `json:"memoryCapacity"`
	MemoryRequests float64 `json:"memoryRequests"`
	MemoryLimits   float64 `json:"memoryLimits"`
	Events         []Event `json:"events"`
}

// ==================== 资源状态结构 ====================

// Node 节点状态
type Node struct {
	Name         string   `json:"name"`
	IP           string   `json:"ip"`
	Status       string   `json:"status"`
	CPUUsage     float64  `json:"cpuUsage"`
	MemoryUsage  float64  `json:"memoryUsage"`
	Role         []string `json:"role"`
	PodsUsed     int      `json:"podsUsed"`
	PodsCapacity int      `json:"podsCapacity"`
	Age          string   `json:"age"`
}

// Pod Pod 状态
type Pod struct {
	Namespace string `json:"namespace"`
	Name      string `json:"name"`
	Status    string `json:"status"`
	Ready     string `json:"ready"`
	Restarts  int32  `json:"restarts"`
	Age       string `json:"age"`
	PodIP     string `json:"podIP"`
	NodeName  string `json:"nodeName"`
}

// Deployment Deployment 状态
type Deployment struct {
	Namespace       string `json:"namespace"`
	Name            string `json:"name"`
	ReadyReplicas   int32  `json:"readyReplicas"`
	UpdatedReplicas int32  `json:"updatedReplicas"`
	Available       int32  `json:"availableReplicas"`
	Desired         int32  `json:"desiredReplicas"`
	Restarts        int32  `json:"restarts"`
	Status          string `json:"status"`
	Age             string `json:"age"`
}

// GetStatus 获取 Deployment 状态
func (d *Deployment) GetStatus() string {
	if d.Desired == 0 {
		return WorkloadScaledToZero
	}
	if d.Available == d.Desired {
		return WorkloadAvailable
	}
	if d.Available > 0 {
		return WorkloadPartial
	}
	return WorkloadUnavailable
}

// StatefulSet StatefulSet 状态
type StatefulSet struct {
	Namespace       string `json:"namespace"`
	Name            string `json:"name"`
	ReadyReplicas   int32  `json:"readyReplicas"`
	UpdatedReplicas int32  `json:"updatedReplicas"`
	Available       int32  `json:"availableReplicas"`
	Desired         int32  `json:"desiredReplicas"`
	Restarts        int32  `json:"restarts"`
	Status          string `json:"status"`
	Age             string `json:"age"`
}

// GetStatus 获取 StatefulSet 状态
func (s *StatefulSet) GetStatus() string {
	if s.Desired == 0 {
		return WorkloadScaledToZero
	}
	if s.Available == s.Desired {
		return WorkloadAvailable
	}
	if s.Available > 0 {
		return WorkloadPartial
	}
	return WorkloadUnavailable
}

// DaemonSet DaemonSet 状态
type DaemonSet struct {
	Namespace       string `json:"namespace"`
	Name            string `json:"name"`
	ReadyReplicas   int32  `json:"readyReplicas"`
	UpdatedReplicas int32  `json:"updatedReplicas"`
	Available       int32  `json:"availableReplicas"`
	Desired         int32  `json:"desiredReplicas"`
	Restarts        int32  `json:"restarts"`
	Status          string `json:"status"`
	Age             string `json:"age"`
}

// GetStatus 获取 DaemonSet 状态
func (d *DaemonSet) GetStatus() string {
	if d.Desired == 0 {
		return WorkloadScaledToZero
	}
	if d.Available == d.Desired {
		return WorkloadAvailable
	}
	if d.Available > 0 {
		return WorkloadPartial
	}
	return WorkloadUnavailable
}

// Service Service 状态
type Service struct {
	Namespace string   `json:"namespace"`
	Name      string   `json:"name"`
	Type      string   `json:"type"`
	ClusterIP string   `json:"clusterIP"`
	Ports     []string `json:"ports"`
	Age       string   `json:"age"`
}

// Event Event 状态
type Event struct {
	Namespace string `json:"namespace"`
	Name      string `json:"name"`
	Reason    string `json:"reason"`
	Message   string `json:"message"`
	Type      string `json:"type"`
	Count     int32  `json:"count"`
	Object    string `json:"object"`
	Source    string `json:"source"`
	LastSeen  string `json:"lastSeen"`
	Duration  string `json:"duration"`
	Age       string `json:"age"`
}

// CronJob CronJob 状态
type CronJob struct {
	Namespace        string `json:"namespace"`
	Name             string `json:"name"`
	Schedule         string `json:"schedule"`
	Suspend          bool   `json:"suspend"`
	Active           int    `json:"active"`
	LastScheduleTime string `json:"lastScheduleTime"`
	Restarts         int32  `json:"restarts"`
	Status           string `json:"status"`
	Age              string `json:"age"`
}

// IsActive 检查 CronJob 是否活跃
func (c *CronJob) IsActive() bool {
	return !c.Suspend && c.Active > 0
}

// Job Job 状态
type Job struct {
	Namespace      string `json:"namespace"`
	Name           string `json:"name"`
	Completions    int32  `json:"completions"`
	Succeeded      int32  `json:"succeeded"`
	Failed         int32  `json:"failed"`
	Restarts       int32  `json:"restarts"`
	StartTime      string `json:"startTime"`
	CompletionTime string `json:"completionTime"`
	Duration       string `json:"duration"`
	Status         string `json:"status"`
	Age            string `json:"age"`
}

// GetStatus 获取 Job 状态
func (j *Job) GetStatus() string {
	if j.Failed > 0 {
		return "Failed"
	}
	if j.Succeeded >= j.Completions {
		return "Completed"
	}
	if j.StartTime != "" {
		return "Running"
	}
	return "Pending"
}

// Ingress Ingress 状态
type Ingress struct {
	Namespace     string   `json:"namespace"`
	Name          string   `json:"name"`
	Class         string   `json:"class"`
	Hosts         []string `json:"hosts"`
	Ports         []string `json:"ports"`
	Status        string   `json:"status"`
	Path          []string `json:"path"`
	TargetService []string `json:"targetService"`
	Age           string   `json:"age"`
}

// PVC PVC 状态
type PVC struct {
	Namespace    string `json:"namespace"`
	Name         string `json:"name"`
	Status       string `json:"status"`
	Capacity     string `json:"capacity"`
	AccessMode   string `json:"accessMode"`
	StorageClass string `json:"storageClass"`
	VolumeName   string `json:"volumeName"`
	Age          string `json:"age"`
}

// PV PV 状态
type PV struct {
	Name          string `json:"name"`
	Status        string `json:"status"`
	Capacity      string `json:"capacity"`
	AccessMode    string `json:"accessMode"`
	StorageClass  string `json:"storageClass"`
	ClaimRef      string `json:"claimRef"`
	ReclaimPolicy string `json:"reclaimPolicy"`
	Age           string `json:"age"`
}

// StorageClass StorageClass 状态
type StorageClass struct {
	Name              string `json:"name"`
	Provisioner       string `json:"provisioner"`
	ReclaimPolicy     string `json:"reclaimPolicy"`
	VolumeBindingMode string `json:"volumeBindingMode"`
	IsDefault         bool   `json:"isDefault"`
	Age               string `json:"age"`
}

// ConfigMap ConfigMap 状态
type ConfigMap struct {
	Namespace string   `json:"namespace"`
	Name      string   `json:"name"`
	DataCount int      `json:"dataCount"`
	Keys      []string `json:"keys"`
	Age       string   `json:"age"`
}

// Secret Secret 状态
type Secret struct {
	Namespace string   `json:"namespace"`
	Name      string   `json:"name"`
	Type      string   `json:"type"`
	DataCount int      `json:"dataCount"`
	Keys      []string `json:"keys"`
	Age       string   `json:"age"`
}

// Namespace Namespace 状态
type Namespace struct {
	Name   string `json:"name"`
	Status string `json:"status"`
	Age    string `json:"age"`
}

// Endpoints Endpoints 状态
type Endpoints struct {
	Namespace string   `json:"namespace"`
	Name      string   `json:"name"`
	Addresses int      `json:"addresses"`
	Ports     []string `json:"ports"`
	Age       string   `json:"age"`
}

// ==================== 搜索接口 ====================

// SearchableItem 可搜索资源接口
type SearchableItem interface {
	GetSearchableFields() map[string]string
}

// GetSearchableFields Pod 搜索字段
func (p Pod) GetSearchableFields() map[string]string {
	return map[string]string{
		"Name":      p.Name,
		"Namespace": p.Namespace,
		"Status":    p.Status,
		"PodIP":     p.PodIP,
		"NodeName":  p.NodeName,
	}
}

// GetSearchableFields Deployment 搜索字段
func (d Deployment) GetSearchableFields() map[string]string {
	return map[string]string{
		"Name":      d.Name,
		"Namespace": d.Namespace,
		"Status":    d.Status,
	}
}

// GetSearchableFields StatefulSet 搜索字段
func (s StatefulSet) GetSearchableFields() map[string]string {
	return map[string]string{
		"Name":      s.Name,
		"Namespace": s.Namespace,
		"Status":    s.Status,
	}
}

// GetSearchableFields DaemonSet 搜索字段
func (d DaemonSet) GetSearchableFields() map[string]string {
	return map[string]string{
		"Name":      d.Name,
		"Namespace": d.Namespace,
		"Status":    d.Status,
	}
}

// GetSearchableFields Service 搜索字段
func (s Service) GetSearchableFields() map[string]string {
	return map[string]string{
		"Name":      s.Name,
		"Namespace": s.Namespace,
		"Type":      s.Type,
		"ClusterIP": s.ClusterIP,
	}
}

// GetSearchableFields Node 搜索字段
func (n Node) GetSearchableFields() map[string]string {
	return map[string]string{
		"Name":   n.Name,
		"IP":     n.IP,
		"Status": n.Status,
	}
}

// GetSearchableFields Event 搜索字段
func (e Event) GetSearchableFields() map[string]string {
	return map[string]string{
		"Namespace": e.Namespace,
		"Reason":    e.Reason,
		"Message":   e.Message,
		"Type":      e.Type,
	}
}

// GetSearchableFields CronJob 搜索字段
func (c CronJob) GetSearchableFields() map[string]string {
	return map[string]string{
		"Name":      c.Name,
		"Namespace": c.Namespace,
		"Status":    c.Status,
	}
}

// GetSearchableFields Job 搜索字段
func (j Job) GetSearchableFields() map[string]string {
	return map[string]string{
		"Name":      j.Name,
		"Namespace": j.Namespace,
		"Status":    j.Status,
	}
}

// GetSearchableFields Ingress 搜索字段
func (i Ingress) GetSearchableFields() map[string]string {
	return map[string]string{
		"Name":      i.Name,
		"Namespace": i.Namespace,
		"Status":    i.Status,
	}
}

// GetSearchableFields PVC 搜索字段
func (p PVC) GetSearchableFields() map[string]string {
	return map[string]string{
		"Name":   p.Name,
		"Status": p.Status,
	}
}

// GetSearchableFields PV 搜索字段
func (p PV) GetSearchableFields() map[string]string {
	return map[string]string{
		"Name":   p.Name,
		"Status": p.Status,
	}
}

// GetSearchableFields StorageClass 搜索字段
func (s StorageClass) GetSearchableFields() map[string]string {
	return map[string]string{
		"Name": s.Name,
	}
}

// GetSearchableFields ConfigMap 搜索字段
func (c ConfigMap) GetSearchableFields() map[string]string {
	return map[string]string{
		"Name": c.Name,
	}
}

// GetSearchableFields Secret 搜索字段
func (s Secret) GetSearchableFields() map[string]string {
	return map[string]string{
		"Name": s.Name,
	}
}

// GetSearchableFields Namespace 搜索字段
func (n Namespace) GetSearchableFields() map[string]string {
	return map[string]string{
		"Name":   n.Name,
		"Status": n.Status,
	}
}

// GetSearchableFields Endpoints 搜索字段
func (e Endpoints) GetSearchableFields() map[string]string {
	return map[string]string{
		"Name":      e.Name,
		"Namespace": e.Namespace,
	}
}

// ==================== 时间工具 ====================

// FormatAge 格式化时间为人类可读的相对时间（英文格式）
// 统一使用 d（天）作为最大单位，不显示 M（月）和 y（年）
func FormatAge(t time.Time) string {
	now := time.Now()
	duration := now.Sub(t)

	if duration < time.Minute {
		return "0s"
	}
	if duration < time.Hour {
		return formatDuration(duration, time.Minute, "m")
	}
	if duration < 24*time.Hour {
		return formatDuration(duration, time.Hour, "h")
	}
	// 全部使用天（d）作为单位，不显示月和年
	return formatDuration(duration, 24*time.Hour, "d")
}

// formatDuration 格式化时间间隔为人类可读格式
func formatDuration(d time.Duration, unit time.Duration, unitName string) string {
	count := int(d / unit)
	return fmt.Sprintf("%d%s", count, unitName)
}
