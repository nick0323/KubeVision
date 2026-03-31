package model

// ==================== API 响应结构 ====================

// APIResponse 统一 API 响应结构
type APIResponse struct {
	Code      int         `json:"code"`
	Message   string      `json:"message"`
	Data      interface{} `json:"data"`
	TraceID   string      `json:"traceId,omitempty"`
	Timestamp int64       `json:"timestamp"`
	Page      *PageMeta   `json:"page,omitempty"`
}

// PageMeta 分页元数据
type PageMeta struct {
	Total  int `json:"total"`
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
}

// APIError API 错误结构
type APIError struct {
	Code    int         `json:"code"`
	Message string      `json:"message"`
	Details interface{} `json:"details,omitempty"`
}

// Error 实现 error 接口
func (e *APIError) Error() string {
	return e.Message
}

// ==================== 登录相关 ====================

// LoginRequest 登录请求
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// ==================== 指标类型 ====================

// PodMetrics Pod 指标
type PodMetrics struct {
	CPU string `json:"cpu"` // 如 "123m"
	Mem string `json:"mem"` // 如 "512Mi"
}

// PodMetricsMap Pod 指标映射
type PodMetricsMap map[string]PodMetrics

// NodeMetrics Node 指标
type NodeMetrics struct {
	CPU string `json:"cpu"`
	Mem string `json:"mem"`
}

// NodeMetricsMap Node 指标映射
type NodeMetricsMap map[string]NodeMetrics

// ==================== 集群概览 ====================

// OverviewStatus 集群概览
type OverviewStatus struct {
	NodeCount      int           `json:"nodeCount"`
	NodeReady      int           `json:"nodeReady"`
	PodCount       int           `json:"podCount"`
	PodNotReady    int           `json:"podNotReady"`
	NamespaceCount int           `json:"namespaceCount"`
	ServiceCount   int           `json:"serviceCount"`
	CPUCapacity    float64       `json:"cpuCapacity"`
	CPURequests    float64       `json:"cpuRequests"`
	CPULimits      float64       `json:"cpuLimits"`
	MemoryCapacity float64       `json:"memoryCapacity"`
	MemoryRequests float64       `json:"memoryRequests"`
	MemoryLimits   float64       `json:"memoryLimits"`
	Events         []EventStatus `json:"events"`
}

// ==================== 资源状态结构 ====================

// NodeStatus 节点状态
type NodeStatus struct {
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

// PodStatus Pod 状态
type PodStatus struct {
	Namespace   string `json:"namespace"`
	Name        string `json:"name"`
	Status      string `json:"status"`
	Ready       string `json:"ready"`
	Restarts    int32  `json:"restarts"`
	Age         string `json:"age"`
	CPUUsage    string `json:"cpuUsage"`
	MemoryUsage string `json:"memoryUsage"`
	PodIP       string `json:"podIP"`
	NodeName    string `json:"nodeName"`
}

// DeploymentStatus Deployment 状态
type DeploymentStatus struct {
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

// StatefulSetStatus StatefulSet 状态
type StatefulSetStatus struct {
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

// DaemonSetStatus DaemonSet 状态
type DaemonSetStatus struct {
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

// ServiceStatus Service 状态
type ServiceStatus struct {
	Namespace  string   `json:"namespace"`
	Name       string   `json:"name"`
	Type       string   `json:"type"`
	ClusterIP  string   `json:"clusterIP"`
	ExternalIP string   `json:"externalIP"`
	Ports      []string `json:"ports"`
	Age        string   `json:"age"`
}

// EventStatus Event 状态
type EventStatus struct {
	Namespace string `json:"namespace"`
	Name      string `json:"name"`
	Reason    string `json:"reason"`
	Message   string `json:"message"`
	Type      string `json:"type"`
	Count     int32  `json:"count"`
	Object    string `json:"object"`
	Source    string `json:"source"`
	FirstSeen string `json:"firstSeen"`
	LastSeen  string `json:"lastSeen"`
	Duration  string `json:"duration"`
	Age       string `json:"age"`
}

// CronJobStatus CronJob 状态
type CronJobStatus struct {
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

// JobStatus Job 状态
type JobStatus struct {
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

// IngressStatus Ingress 状态
type IngressStatus struct {
	Namespace     string   `json:"namespace"`
	Name          string   `json:"name"`
	Class         string   `json:"class"`
	Hosts         []string `json:"hosts"`
	Address       string   `json:"address"`
	Ports         []string `json:"ports"`
	Status        string   `json:"status"`
	Path          []string `json:"path"`
	TargetService []string `json:"targetService"`
	Age           string   `json:"age"`
}

// PVCStatus PVC 状态
type PVCStatus struct {
	Namespace    string `json:"namespace"`
	Name         string `json:"name"`
	Status       string `json:"status"`
	Capacity     string `json:"capacity"`
	AccessMode   string `json:"accessMode"`
	StorageClass string `json:"storageClass"`
	VolumeName   string `json:"volumeName"`
	Age          string `json:"age"`
}

// PVStatus PV 状态
type PVStatus struct {
	Name          string `json:"name"`
	Status        string `json:"status"`
	Capacity      string `json:"capacity"`
	AccessMode    string `json:"accessMode"`
	StorageClass  string `json:"storageClass"`
	ClaimRef      string `json:"claimRef"`
	ReclaimPolicy string `json:"reclaimPolicy"`
	Age           string `json:"age"`
}

// StorageClassStatus StorageClass 状态
type StorageClassStatus struct {
	Name              string `json:"name"`
	Provisioner       string `json:"provisioner"`
	ReclaimPolicy     string `json:"reclaimPolicy"`
	VolumeBindingMode string `json:"volumeBindingMode"`
	IsDefault         bool   `json:"isDefault"`
	Age               string `json:"age"`
}

// ConfigMapStatus ConfigMap 状态
type ConfigMapStatus struct {
	Namespace string   `json:"namespace"`
	Name      string   `json:"name"`
	DataCount int      `json:"dataCount"`
	Keys      []string `json:"keys"`
	Age       string   `json:"age"`
}

// SecretStatus Secret 状态
type SecretStatus struct {
	Namespace string   `json:"namespace"`
	Name      string   `json:"name"`
	Type      string   `json:"type"`
	DataCount int      `json:"dataCount"`
	Keys      []string `json:"keys"`
	Age       string   `json:"age"`
}

// ==================== 详情结构（用于 Detail API） ====================

// NamespaceDetail Namespace 详情
type NamespaceDetail struct {
	Name   string            `json:"name"`
	Status string            `json:"status"`
	Age    string            `json:"age"`
	Labels map[string]string `json:"labels"`
}

// NodeDetail Node 详情
type NodeDetail struct {
	Name         string            `json:"name"`
	Status       string            `json:"status"`
	IP           string            `json:"ip"`
	CPUUsage     float64           `json:"cpuUsage"`
	MemoryUsage  float64           `json:"memoryUsage"`
	Role         []string          `json:"role"`
	PodsUsed     int               `json:"podsUsed"`
	PodsCapacity int               `json:"podsCapacity"`
	Labels       map[string]string `json:"labels"`
}

// EventDetail Event 详情
type EventDetail struct {
	Namespace string            `json:"namespace"`
	Name      string            `json:"name"`
	Status    string            `json:"status"`
	Reason    string            `json:"reason"`
	Message   string            `json:"message"`
	Type      string            `json:"type"`
	Count     int32             `json:"count"`
	FirstSeen string            `json:"firstSeen"`
	LastSeen  string            `json:"lastSeen"`
	Duration  string            `json:"duration"`
	Labels    map[string]string `json:"labels"`
}

// PVCDetail PVC 详情
type PVCDetail struct {
	Namespace    string            `json:"namespace"`
	Name         string            `json:"name"`
	Status       string            `json:"status"`
	Capacity     string            `json:"capacity"`
	AccessMode   []string          `json:"accessMode"`
	StorageClass string            `json:"storageClass"`
	VolumeName   string            `json:"volumeName"`
	Labels       map[string]string `json:"labels"`
}

// PVDetail PV 详情
type PVDetail struct {
	Name          string            `json:"name"`
	Status        string            `json:"status"`
	Capacity      string            `json:"capacity"`
	AccessMode    []string          `json:"accessMode"`
	StorageClass  string            `json:"storageClass"`
	ClaimRef      string            `json:"claimRef"`
	ReclaimPolicy string            `json:"reclaimPolicy"`
	Labels        map[string]string `json:"labels"`
}

// StorageClassDetail StorageClass 详情
type StorageClassDetail struct {
	Name              string            `json:"name"`
	Provisioner       string            `json:"provisioner"`
	ReclaimPolicy     string            `json:"reclaimPolicy"`
	VolumeBindingMode string            `json:"volumeBindingMode"`
	IsDefault         bool              `json:"isDefault"`
	Parameters        map[string]string `json:"parameters"`
	Labels            map[string]string `json:"labels"`
}

// ConfigMapDetail ConfigMap 详情
type ConfigMapDetail struct {
	Namespace string            `json:"namespace"`
	Name      string            `json:"name"`
	Status    string            `json:"status"`
	DataCount int               `json:"dataCount"`
	Keys      []string          `json:"keys"`
	Labels    map[string]string `json:"labels"`
}

// SecretDetail Secret 详情
type SecretDetail struct {
	Namespace string            `json:"namespace"`
	Name      string            `json:"name"`
	Status    string            `json:"status"`
	Type      string            `json:"type"`
	DataCount int               `json:"dataCount"`
	Keys      []string          `json:"keys"`
	Labels    map[string]string `json:"labels"`
}

// ==================== 搜索接口 ====================

// SearchableItem 可搜索接口
type SearchableItem interface {
	GetSearchableFields() map[string]string
}

// GetSearchableFields PodStatus 搜索字段
func (p PodStatus) GetSearchableFields() map[string]string {
	return map[string]string{
		"Name":      p.Name,
		"Namespace": p.Namespace,
		"Status":    p.Status,
		"PodIP":     p.PodIP,
		"NodeName":  p.NodeName,
	}
}

// GetSearchableFields DeploymentStatus 搜索字段
func (d DeploymentStatus) GetSearchableFields() map[string]string {
	return map[string]string{
		"Name":      d.Name,
		"Namespace": d.Namespace,
		"Status":    d.Status,
	}
}

// GetSearchableFields StatefulSetStatus 搜索字段
func (s StatefulSetStatus) GetSearchableFields() map[string]string {
	return map[string]string{
		"Name":      s.Name,
		"Namespace": s.Namespace,
		"Status":    s.Status,
	}
}

// GetSearchableFields DaemonSetStatus 搜索字段
func (d DaemonSetStatus) GetSearchableFields() map[string]string {
	return map[string]string{
		"Name":      d.Name,
		"Namespace": d.Namespace,
		"Status":    d.Status,
	}
}

// GetSearchableFields ServiceStatus 搜索字段
func (s ServiceStatus) GetSearchableFields() map[string]string {
	return map[string]string{
		"Name":      s.Name,
		"Namespace": s.Namespace,
		"Type":      s.Type,
		"ClusterIP": s.ClusterIP,
	}
}

// GetSearchableFields NodeStatus 搜索字段
func (n NodeStatus) GetSearchableFields() map[string]string {
	return map[string]string{
		"Name":   n.Name,
		"IP":     n.IP,
		"Status": n.Status,
	}
}

// GetSearchableFields EventStatus 搜索字段
func (e EventStatus) GetSearchableFields() map[string]string {
	return map[string]string{
		"Name":      e.Name,
		"Namespace": e.Namespace,
		"Reason":    e.Reason,
		"Message":   e.Message,
		"Type":      e.Type,
	}
}

// GetSearchableFields CronJobStatus 搜索字段
func (c CronJobStatus) GetSearchableFields() map[string]string {
	return map[string]string{
		"Name":      c.Name,
		"Namespace": c.Namespace,
		"Status":    c.Status,
	}
}

// GetSearchableFields JobStatus 搜索字段
func (j JobStatus) GetSearchableFields() map[string]string {
	return map[string]string{
		"Name":      j.Name,
		"Namespace": j.Namespace,
		"Status":    j.Status,
	}
}

// GetSearchableFields IngressStatus 搜索字段
func (i IngressStatus) GetSearchableFields() map[string]string {
	return map[string]string{
		"Name":      i.Name,
		"Namespace": i.Namespace,
		"Status":    i.Status,
	}
}

// GetSearchableFields PVCStatus 搜索字段
func (p PVCStatus) GetSearchableFields() map[string]string {
	return map[string]string{
		"Name":   p.Name,
		"Status": p.Status,
	}
}

// GetSearchableFields PVStatus 搜索字段
func (p PVStatus) GetSearchableFields() map[string]string {
	return map[string]string{
		"Name":   p.Name,
		"Status": p.Status,
	}
}

// GetSearchableFields StorageClassStatus 搜索字段
func (s StorageClassStatus) GetSearchableFields() map[string]string {
	return map[string]string{
		"Name": s.Name,
	}
}

// GetSearchableFields ConfigMapStatus 搜索字段
func (c ConfigMapStatus) GetSearchableFields() map[string]string {
	return map[string]string{
		"Name": c.Name,
	}
}

// GetSearchableFields SecretStatus 搜索字段
func (s SecretStatus) GetSearchableFields() map[string]string {
	return map[string]string{
		"Name": s.Name,
	}
}

// GetSearchableFields NamespaceDetail 搜索字段
func (n NamespaceDetail) GetSearchableFields() map[string]string {
	return map[string]string{
		"Name":   n.Name,
		"Status": n.Status,
	}
}
