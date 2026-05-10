package model

import (
	"fmt"
	"time"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ==================== API 响应 ====================

type APIResponse struct {
	Code      int       `json:"code"`
	Message   string    `json:"message"`
	Data      any       `json:"data"`
	TraceID   string    `json:"traceId,omitempty"`
	Timestamp int64     `json:"timestamp"`
	Page      *PageMeta `json:"page,omitempty"`
}

type PageMeta struct {
	Total  int `json:"total"`
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
}

type APIError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
	Details any    `json:"details,omitempty"`
}

func (e *APIError) Error() string {
	return e.Message
}

// ==================== 请求 ====================

type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// ==================== 集群概览 ====================

type OverviewStatus struct {
	NodeCount      int     `json:"nodeCount"`
	NodeReadyCount int     `json:"nodeReadyCount"`
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

// ==================== K8s 资源 ====================

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

type Pod struct {
	Namespace       string                  `json:"namespace"`
	Name            string                  `json:"name"`
	Status          string                  `json:"status"`
	Ready           string                  `json:"ready"`
	Restarts        int32                   `json:"restarts"`
	Age             string                  `json:"age"`
	PodIP           string                  `json:"podIP"`
	NodeName        string                  `json:"nodeName"`
	OwnerReferences []metav1.OwnerReference `json:"ownerReferences,omitempty"`
}

type Deployment struct {
	Namespace       string `json:"namespace"`
	Name            string `json:"name"`
	ReadyReplicas   int32  `json:"readyReplicas"`
	UpdatedReplicas int32  `json:"updatedReplicas"`
	Available       int32  `json:"availableReplicas"`
	Desired         int32  `json:"desiredReplicas"`
	Status          string `json:"status"`
	Age             string `json:"age"`
}

type StatefulSet struct {
	Namespace       string `json:"namespace"`
	Name            string `json:"name"`
	ReadyReplicas   int32  `json:"readyReplicas"`
	UpdatedReplicas int32  `json:"updatedReplicas"`
	Available       int32  `json:"availableReplicas"`
	Desired         int32  `json:"desiredReplicas"`
	Status          string `json:"status"`
	Age             string `json:"age"`
}

type DaemonSet struct {
	Namespace       string `json:"namespace"`
	Name            string `json:"name"`
	ReadyReplicas   int32  `json:"readyReplicas"`
	UpdatedReplicas int32  `json:"updatedReplicas"`
	Available       int32  `json:"availableReplicas"`
	Desired         int32  `json:"desiredReplicas"`
	Status          string `json:"status"`
	Age             string `json:"age"`
}

type Service struct {
	Namespace string   `json:"namespace"`
	Name      string   `json:"name"`
	Type      string   `json:"type"`
	ClusterIP string   `json:"clusterIP"`
	Ports     []string `json:"ports"`
	Age       string   `json:"age"`
}

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

type CronJob struct {
	Namespace        string `json:"namespace"`
	Name             string `json:"name"`
	Schedule         string `json:"schedule"`
	Suspend          bool   `json:"suspend"`
	Active           int    `json:"active"`
	LastScheduleTime string `json:"lastScheduleTime"`
	Status           string `json:"status"`
	Age              string `json:"age"`
}

type Job struct {
	Namespace      string `json:"namespace"`
	Name           string `json:"name"`
	Completions    int32  `json:"completions"`
	Succeeded      int32  `json:"succeeded"`
	Failed         int32  `json:"failed"`
	StartTime      string `json:"startTime"`
	CompletionTime string `json:"completionTime"`
	Duration       string `json:"duration"`
	Status         string `json:"status"`
	Age            string `json:"age"`
}

type Ingress struct {
	Namespace     string   `json:"namespace"`
	Name          string   `json:"name"`
	Class         string   `json:"class"`
	Hosts         []string `json:"hosts"`
	Path          []string `json:"path"`
	TargetService []string `json:"targetService"`
	Age           string   `json:"age"`
}

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

type StorageClass struct {
	Name              string `json:"name"`
	Provisioner       string `json:"provisioner"`
	ReclaimPolicy     string `json:"reclaimPolicy"`
	VolumeBindingMode string `json:"volumeBindingMode"`
	IsDefault         bool   `json:"isDefault"`
	Age               string `json:"age"`
}

type ConfigMap struct {
	Namespace string   `json:"namespace"`
	Name      string   `json:"name"`
	DataCount int      `json:"dataCount"`
	Keys      []string `json:"keys"`
	Age       string   `json:"age"`
}

type Secret struct {
	Namespace string   `json:"namespace"`
	Name      string   `json:"name"`
	Type      string   `json:"type"`
	DataCount int      `json:"dataCount"`
	Keys      []string `json:"keys"`
	Age       string   `json:"age"`
}

type Namespace struct {
	Name   string `json:"name"`
	Status string `json:"status"`
	Age    string `json:"age"`
}

type Endpoints struct {
	Namespace string   `json:"namespace"`
	Name      string   `json:"name"`
	Addresses int      `json:"addresses"`
	Ports     []string `json:"ports"`
	Age       string   `json:"age"`
}

type NodeMetrics struct {
	CPU    string `json:"cpu"`
	Memory string `json:"memory"`
}

type NetworkPolicy struct {
	Namespace   string   `json:"namespace"`
	Name        string   `json:"name"`
	PodSelector string   `json:"podSelector"`
	PolicyTypes []string `json:"policyTypes"`
	Age         string   `json:"age"`
}

type ServiceAccount struct {
	Namespace string `json:"namespace"`
	Name      string `json:"name"`
	Secrets   int    `json:"secrets"`
	Age       string `json:"age"`
}

type Role struct {
	Namespace string `json:"namespace"`
	Name      string `json:"name"`
	Rules     int    `json:"rules"`
	Age       string `json:"age"`
}

type RoleBinding struct {
	Namespace string `json:"namespace"`
	Name      string `json:"name"`
	RoleRef   string `json:"roleRef"`
	Subjects  int    `json:"subjects"`
	Age       string `json:"age"`
}

type ClusterRole struct {
	Name  string `json:"name"`
	Rules int    `json:"rules"`
	Age   string `json:"age"`
}

type ClusterRoleBinding struct {
	Name     string `json:"name"`
	RoleRef  string `json:"roleRef"`
	Subjects int    `json:"subjects"`
	Age      string `json:"age"`
}

type ResourceQuota struct {
	Namespace string `json:"namespace"`
	Name      string `json:"name"`
	Requests  string `json:"requests"`
	Limits    string `json:"limits"`
	Age       string `json:"age"`
}

type LimitRange struct {
	Namespace string `json:"namespace"`
	Name      string `json:"name"`
	Limits    string `json:"limits"`
	Age       string `json:"age"`
}

type PodDisruptionBudget struct {
	Namespace      string `json:"namespace"`
	Name           string `json:"name"`
	MinAvailable   string `json:"minAvailable"`
	MaxUnavailable string `json:"maxUnavailable"`
	CurrentHealthy int32  `json:"currentHealthy"`
	DesiredHealthy int32  `json:"desiredHealthy"`
	Age            string `json:"age"`
}

type HorizontalPodAutoscaler struct {
	Namespace       string `json:"namespace"`
	Name            string `json:"name"`
	MinReplicas     int32  `json:"minReplicas"`
	MaxReplicas     int32  `json:"maxReplicas"`
	CurrentReplicas int32  `json:"currentReplicas"`
	DesiredReplicas int32  `json:"desiredReplicas"`
	Metrics         string `json:"metrics"`
	Age             string `json:"age"`
}

// ==================== 接口 ====================

type SearchableItem interface {
	GetSearchableFields() map[string]string
}

func (p Pod) GetSearchableFields() map[string]string {
	return map[string]string{"Name": p.Name, "Namespace": p.Namespace, "Status": p.Status, "PodIP": p.PodIP, "NodeName": p.NodeName}
}

func (d Deployment) GetSearchableFields() map[string]string {
	return map[string]string{"Name": d.Name, "Namespace": d.Namespace, "Status": d.Status}
}

func (s StatefulSet) GetSearchableFields() map[string]string {
	return map[string]string{"Name": s.Name, "Namespace": s.Namespace, "Status": s.Status}
}

func (d DaemonSet) GetSearchableFields() map[string]string {
	return map[string]string{"Name": d.Name, "Namespace": d.Namespace, "Status": d.Status}
}

func (s Service) GetSearchableFields() map[string]string {
	return map[string]string{"Name": s.Name, "Namespace": s.Namespace, "Type": s.Type, "ClusterIP": s.ClusterIP}
}

func (n Node) GetSearchableFields() map[string]string {
	return map[string]string{"Name": n.Name, "IP": n.IP, "Status": n.Status}
}

func (e Event) GetSearchableFields() map[string]string {
	return map[string]string{"Namespace": e.Namespace, "Reason": e.Reason, "Message": e.Message, "Type": e.Type}
}

func (c CronJob) GetSearchableFields() map[string]string {
	return map[string]string{"Name": c.Name, "Namespace": c.Namespace, "Status": c.Status}
}

func (j Job) GetSearchableFields() map[string]string {
	return map[string]string{"Name": j.Name, "Namespace": j.Namespace, "Status": j.Status}
}

func (i Ingress) GetSearchableFields() map[string]string {
	return map[string]string{"Name": i.Name, "Namespace": i.Namespace}
}

func (p PVC) GetSearchableFields() map[string]string {
	return map[string]string{"Name": p.Name, "Status": p.Status}
}

func (p PV) GetSearchableFields() map[string]string {
	return map[string]string{"Name": p.Name, "Status": p.Status}
}

func (s StorageClass) GetSearchableFields() map[string]string {
	return map[string]string{"Name": s.Name}
}

func (c ConfigMap) GetSearchableFields() map[string]string {
	return map[string]string{"Name": c.Name}
}

func (s Secret) GetSearchableFields() map[string]string {
	return map[string]string{"Name": s.Name}
}

func (n Namespace) GetSearchableFields() map[string]string {
	return map[string]string{"Name": n.Name, "Status": n.Status}
}

func (e Endpoints) GetSearchableFields() map[string]string {
	return map[string]string{"Name": e.Name, "Namespace": e.Namespace}
}

func (n NetworkPolicy) GetSearchableFields() map[string]string {
	return map[string]string{"Name": n.Name, "Namespace": n.Namespace}
}

func (s ServiceAccount) GetSearchableFields() map[string]string {
	return map[string]string{"Name": s.Name, "Namespace": s.Namespace}
}

func (r Role) GetSearchableFields() map[string]string {
	return map[string]string{"Name": r.Name, "Namespace": r.Namespace}
}

func (r RoleBinding) GetSearchableFields() map[string]string {
	return map[string]string{"Name": r.Name, "Namespace": r.Namespace, "RoleRef": r.RoleRef}
}

func (c ClusterRole) GetSearchableFields() map[string]string {
	return map[string]string{"Name": c.Name}
}

func (c ClusterRoleBinding) GetSearchableFields() map[string]string {
	return map[string]string{"Name": c.Name, "RoleRef": c.RoleRef}
}

func (r ResourceQuota) GetSearchableFields() map[string]string {
	return map[string]string{"Name": r.Name, "Namespace": r.Namespace}
}

func (l LimitRange) GetSearchableFields() map[string]string {
	return map[string]string{"Name": l.Name, "Namespace": l.Namespace}
}

func (p PodDisruptionBudget) GetSearchableFields() map[string]string {
	return map[string]string{"Name": p.Name, "Namespace": p.Namespace}
}

func (h HorizontalPodAutoscaler) GetSearchableFields() map[string]string {
	return map[string]string{"Name": h.Name, "Namespace": h.Namespace}
}

// ==================== 工具函数 ====================

func FormatTime(t *metav1.Time) string {
	if t == nil || t.IsZero() {
		return ""
	}
	return t.Time.Format(TimeFormatRFC3339)
}

func FormatAge(t time.Time) string {
	duration := time.Since(t)
	if duration < time.Minute {
		return "0s"
	}
	if duration < time.Hour {
		return fmt.Sprintf("%dm", int(duration/time.Minute))
	}
	if duration < 24*time.Hour {
		return fmt.Sprintf("%dh", int(duration/time.Hour))
	}
	return fmt.Sprintf("%dd", int(duration/(24*time.Hour)))
}
