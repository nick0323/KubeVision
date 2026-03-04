package api

import (
	"context"
	"net/http"

	"github.com/nick0323/K8sVision/api/middleware"
	"github.com/nick0323/K8sVision/model"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

func RegisterIngress(
	r *gin.RouterGroup,
	logger *zap.Logger,
	getK8sClient K8sClientProvider,
	listIngresses func(context.Context, *kubernetes.Clientset, string) ([]model.IngressStatus, error),
) {
	r.GET("/ingress", getIngressList(logger, getK8sClient, listIngresses))
	r.GET("/ingress/:namespace/:name", getIngressDetail(logger, getK8sClient))
}

func getIngressList(
	logger *zap.Logger,
	getK8sClient K8sClientProvider,
	listIngresses func(context.Context, *kubernetes.Clientset, string) ([]model.IngressStatus, error),
) gin.HandlerFunc {
	return func(c *gin.Context) {
		HandleListWithPagination(c, logger, func(ctx context.Context, params PaginationParams) ([]model.IngressStatus, error) {
			clientset, _, err := getK8sClient()
			if err != nil {
				return nil, err
			}
			return listIngresses(ctx, clientset, params.Namespace)
		}, ListSuccessMessage)
	}
}

func getIngressDetail(
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
		ns := c.Param("namespace")
		name := c.Param("name")
		ingress, err := clientset.NetworkingV1().Ingresses(ns).Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			middleware.ResponseError(c, logger, err, http.StatusNotFound)
			return
		}

		hosts := make([]string, 0, len(ingress.Spec.Rules))
		paths := make([]string, 0)
		targetServices := make([]string, 0)

		for _, rule := range ingress.Spec.Rules {
			hosts = append(hosts, rule.Host)

			if rule.HTTP != nil {
				for _, path := range rule.HTTP.Paths {
					if path.Path != "" {
						paths = append(paths, path.Path)
					} else {
						paths = append(paths, "/")
					}

					if path.Backend.Service != nil {
						targetServices = append(targetServices, path.Backend.Service.Name)
					}
				}
			}
		}

		address := ""
		if len(ingress.Status.LoadBalancer.Ingress) > 0 {
			if ingress.Status.LoadBalancer.Ingress[0].IP != "" {
				address = ingress.Status.LoadBalancer.Ingress[0].IP
			} else if ingress.Status.LoadBalancer.Ingress[0].Hostname != "" {
				address = ingress.Status.LoadBalancer.Ingress[0].Hostname
			}
		}

		class := ""
		if ingress.Spec.IngressClassName != nil {
			class = *ingress.Spec.IngressClassName
		}

		ingressDetail := model.IngressDetail{
			CommonResourceFields: model.CommonResourceFields{
				Namespace: ingress.Namespace,
				Name:      ingress.Name,
				Status:    "Ready",
				BaseMetadata: model.BaseMetadata{
					Labels:      ingress.Labels,
					Annotations: ingress.Annotations,
				},
			},
			Hosts:         hosts,
			Address:       address,
			Ports:         model.DefaultIngressPorts,
			Class:         class,
			Path:          paths,
			TargetService: targetServices,
		}
		middleware.ResponseSuccess(c, ingressDetail, DetailSuccessMessage, nil)
	}
}
