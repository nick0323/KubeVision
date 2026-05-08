package service

import (
	"context"
	"fmt"

	"go.uber.org/zap"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/apis/meta/v1/unstructured"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/client-go/dynamic"
	"k8s.io/client-go/rest"
)

// ArgoCD 资源 GVR
var (
	ArgoCDApplicationGVR = schema.GroupVersionResource{
		Group:    "argoproj.io",
		Version:  "v1alpha1",
		Resource: "applications",
	}
)

// ArgoCDManager 管理 ArgoCD 客户端
type ArgoCDManager struct {
	dynamicClient dynamic.Interface
	logger        *zap.Logger
}

// NewArgoCDManager 创建 ArgoCD 客户端管理器
func NewArgoCDManager(restConfig *rest.Config, logger *zap.Logger) (*ArgoCDManager, error) {
	if restConfig == nil {
		return nil, fmt.Errorf("rest config is required")
	}

	dynamicClient, err := dynamic.NewForConfig(restConfig)
	if err != nil {
		return nil, fmt.Errorf("failed to create dynamic client: %w", err)
	}

	logger.Info("ArgoCD client initialized (CRD mode)")
	return &ArgoCDManager{
		dynamicClient: dynamicClient,
		logger:        logger,
	}, nil
}

// ListApplications 列出 ArgoCD 应用
func (m *ArgoCDManager) ListApplications(ctx context.Context, namespace string) (*unstructured.UnstructuredList, error) {
	if m == nil {
		return nil, fmt.Errorf("ArgoCD manager not initialized")
	}

	ns := namespace
	if ns == "" {
		ns = metav1.NamespaceAll
	}

	return m.dynamicClient.Resource(ArgoCDApplicationGVR).Namespace(ns).List(ctx, metav1.ListOptions{})
}

// GetApplicationByName 通过名称获取应用（自动查找所有 namespace）
func (m *ArgoCDManager) GetApplicationByName(ctx context.Context, name string) (*unstructured.Unstructured, error) {
	if m == nil {
		return nil, fmt.Errorf("ArgoCD manager not initialized")
	}

	// 跨所有 namespace 查找应用，避免硬编码 argocd namespace 导致找不到资源
	apps, err := m.dynamicClient.Resource(ArgoCDApplicationGVR).Namespace(metav1.NamespaceAll).List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	for i := range apps.Items {
		if apps.Items[i].GetName() == name {
			return &apps.Items[i], nil
		}
	}

	return nil, fmt.Errorf("application %s not found", name)
}

// SyncApplicationByName 同步应用
func (m *ArgoCDManager) SyncApplicationByName(ctx context.Context, name string) error {
	app, err := m.GetApplicationByName(ctx, name)
	if err != nil {
		return err
	}

	annotations := app.GetAnnotations()
	if annotations == nil {
		annotations = make(map[string]string)
	}
	annotations["argocd.argoproj.io/refresh"] = "normal"
	app.SetAnnotations(annotations)

	// 使用应用实际的 namespace 进行更新
	_, err = m.dynamicClient.Resource(ArgoCDApplicationGVR).Namespace(app.GetNamespace()).Update(ctx, app, metav1.UpdateOptions{})
	return err
}

// RefreshApplicationByName 刷新应用
func (m *ArgoCDManager) RefreshApplicationByName(ctx context.Context, name string) error {
	return m.SyncApplicationByName(ctx, name)
}

// DeleteApplicationByName 删除应用
func (m *ArgoCDManager) DeleteApplicationByName(ctx context.Context, name string) error {
	app, err := m.GetApplicationByName(ctx, name)
	if err != nil {
		return err
	}

	propagationPolicy := metav1.DeletePropagationForeground
	return m.dynamicClient.Resource(ArgoCDApplicationGVR).Namespace(app.GetNamespace()).Delete(ctx, name, metav1.DeleteOptions{
		PropagationPolicy: &propagationPolicy,
	})
}
