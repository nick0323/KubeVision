package api

import (
	"context"
	"net/http"
	"time"

	"github.com/nick0323/K8sVision/api/middleware"
	"github.com/nick0323/K8sVision/model"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func RegisterDeployment(
	r *gin.RouterGroup,
	logger *zap.Logger,
	getK8sClient K8sClientProvider,
	listDeployments func(context.Context, *kubernetes.Clientset, string) ([]model.DeploymentStatus, error),
) {
	r.GET("/deployments", getDeploymentList(logger, getK8sClient, listDeployments))
	r.GET("/deployments/:namespace/:name", getDeploymentDetail(logger, getK8sClient))
}

func getDeploymentList(
	logger *zap.Logger,
	getK8sClient K8sClientProvider,
	listDeployments func(context.Context, *kubernetes.Clientset, string) ([]model.DeploymentStatus, error),
) gin.HandlerFunc {
	return func(c *gin.Context) {
		HandleListWithPagination(c, logger, func(ctx context.Context, params PaginationParams) ([]model.DeploymentStatus, error) {
			clientset, _, err := getK8sClient()
			if err != nil {
				return nil, err
			}
			return listDeployments(ctx, clientset, params.Namespace)
		}, ListSuccessMessage)
	}
}

func getDeploymentDetail(
	logger *zap.Logger,
	getK8sClient K8sClientProvider,
) gin.HandlerFunc {
	return func(c *gin.Context) {
		clientset, _, err := getK8sClient()
		if err != nil {
			middleware.ResponseError(c, logger, err, http.StatusInternalServerError)
			return
		}
		ctx := GetRequestContext(c)
		namespace := c.Param("namespace")
		name := c.Param("name")

		// 获取原始 Deployment 对象
		dep, err := clientset.AppsV1().Deployments(namespace).Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			middleware.ResponseError(c, logger, err, http.StatusNotFound)
			return
		}

		// 构建完整的对象（包含 kind 和 apiVersion）
		fullObj := map[string]interface{}{
			"apiVersion": "apps/v1",
			"kind":       "Deployment",
			"metadata": map[string]interface{}{
				"name":              dep.Name,
				"namespace":         dep.Namespace,
				"labels":            dep.Labels,
				"annotations":       dep.Annotations,
				"creationTimestamp": dep.CreationTimestamp.Format(time.RFC3339),
				"uid":               dep.UID,
				"resourceVersion":   dep.ResourceVersion,
			},
			"spec": map[string]interface{}{
				"replicas": dep.Spec.Replicas,
				"selector": dep.Spec.Selector,
				"template": dep.Spec.Template,
				"strategy": dep.Spec.Strategy,
				"minReadySeconds": dep.Spec.MinReadySeconds,
			},
			"status": map[string]interface{}{
				"replicas":          dep.Status.Replicas,
				"readyReplicas":     dep.Status.ReadyReplicas,
				"updatedReplicas":   dep.Status.UpdatedReplicas,
				"availableReplicas": dep.Status.AvailableReplicas,
				"unavailableReplicas": dep.Status.UnavailableReplicas,
				"conditions":        dep.Status.Conditions,
				"observedGeneration": dep.Status.ObservedGeneration,
			},
		}

		middleware.ResponseSuccess(c, fullObj, DetailSuccessMessage, nil)
	}
}
