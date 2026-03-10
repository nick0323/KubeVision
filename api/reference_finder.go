package api

import (
	"fmt"
	"net/http"

	"github.com/nick0323/K8sVision/api/middleware"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// ReferenceInfo 引用信息
type ReferenceInfo struct {
	Kind      string `json:"kind"`
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
	RefType   string `json:"refType"` // configMapKeyRef, secretKeyRef, volume, etc.
	Field     string `json:"field"`
}

// ReferenceList 引用列表
type ReferenceList struct {
	Name      string          `json:"name"`
	Namespace string          `json:"namespace"`
	Type      string          `json:"type"` // configmap or secret
	References []ReferenceInfo `json:"references"`
}

func RegisterReferenceFinder(
	r *gin.RouterGroup,
	logger *zap.Logger,
	getK8sClient K8sClientProvider,
) {
	r.GET("/configmaps/:namespace/:name/references", findConfigMapReferences(logger, getK8sClient))
	r.GET("/secrets/:namespace/:name/references", findSecretReferences(logger, getK8sClient))
}

// findConfigMapReferences 查找 ConfigMap 被哪些资源引用
func findConfigMapReferences(
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

		references := []ReferenceInfo{}

		// 1. 查找引用此 ConfigMap 的 Pod
		podList, err := clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})
		if err != nil {
			logger.Warn("列出 Pod 失败", zap.Error(err))
		} else {
			for _, pod := range podList.Items {
				// 检查卷引用
				for _, vol := range pod.Spec.Volumes {
					if vol.ConfigMap != nil && vol.ConfigMap.Name == name {
						references = append(references, ReferenceInfo{
							Kind:      "Pod",
							Name:      pod.Name,
							Namespace: pod.Namespace,
							RefType:   "volume",
							Field:     fmt.Sprintf("spec.volumes[%s]", vol.Name),
						})
					}
				}

				// 检查容器环境变量引用
				for _, container := range pod.Spec.Containers {
					for _, env := range container.Env {
						if env.ValueFrom != nil && env.ValueFrom.ConfigMapKeyRef != nil {
							if env.ValueFrom.ConfigMapKeyRef.LocalObjectReference.Name == name {
								references = append(references, ReferenceInfo{
									Kind:      "Pod",
									Name:      pod.Name,
									Namespace: pod.Namespace,
									RefType:   "configMapKeyRef",
									Field:     fmt.Sprintf("spec.containers[%s].env[%s]", container.Name, env.Name),
								})
							}
						}
					}

					// 检查环境变量从 ConfigMap 所有键值
					for _, envFrom := range container.EnvFrom {
						if envFrom.ConfigMapRef != nil && envFrom.ConfigMapRef.Name == name {
							references = append(references, ReferenceInfo{
								Kind:      "Pod",
								Name:      pod.Name,
								Namespace: pod.Namespace,
								RefType:   "configMapRef",
								Field:     fmt.Sprintf("spec.containers[%s].envFrom", container.Name),
							})
						}
					}
				}
			}
		}

		// 2. 查找 Deployment
		deployList, err := clientset.AppsV1().Deployments(namespace).List(ctx, metav1.ListOptions{})
		if err != nil {
			logger.Warn("列出 Deployment 失败", zap.Error(err))
		} else {
			for _, dep := range deployList.Items {
				for _, vol := range dep.Spec.Template.Spec.Volumes {
					if vol.ConfigMap != nil && vol.ConfigMap.Name == name {
						references = append(references, ReferenceInfo{
							Kind:      "Deployment",
							Name:      dep.Name,
							Namespace: dep.Namespace,
							RefType:   "volume",
							Field:     fmt.Sprintf("spec.template.spec.volumes[%s]", vol.Name),
						})
					}
				}
			}
		}

		// 3. 查找 StatefulSet
		stsList, err := clientset.AppsV1().StatefulSets(namespace).List(ctx, metav1.ListOptions{})
		if err != nil {
			logger.Warn("列出 StatefulSet 失败", zap.Error(err))
		} else {
			for _, sts := range stsList.Items {
				for _, vol := range sts.Spec.Template.Spec.Volumes {
					if vol.ConfigMap != nil && vol.ConfigMap.Name == name {
						references = append(references, ReferenceInfo{
							Kind:      "StatefulSet",
							Name:      sts.Name,
							Namespace: sts.Namespace,
							RefType:   "volume",
							Field:     fmt.Sprintf("spec.template.spec.volumes[%s]", vol.Name),
						})
					}
				}
			}
		}

		// 4. 查找 DaemonSet
		dsList, err := clientset.AppsV1().DaemonSets(namespace).List(ctx, metav1.ListOptions{})
		if err != nil {
			logger.Warn("列出 DaemonSet 失败", zap.Error(err))
		} else {
			for _, ds := range dsList.Items {
				for _, vol := range ds.Spec.Template.Spec.Volumes {
					if vol.ConfigMap != nil && vol.ConfigMap.Name == name {
						references = append(references, ReferenceInfo{
							Kind:      "DaemonSet",
							Name:      ds.Name,
							Namespace: ds.Namespace,
							RefType:   "volume",
							Field:     fmt.Sprintf("spec.template.spec.volumes[%s]", vol.Name),
						})
					}
				}
			}
		}

		result := ReferenceList{
			Name:       name,
			Namespace:  namespace,
			Type:       "configmap",
			References: references,
		}

		middleware.ResponseSuccess(c, result, "引用查找成功", nil)
	}
}

// findSecretReferences 查找 Secret 被哪些资源引用
func findSecretReferences(
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

		references := []ReferenceInfo{}

		// 1. 查找引用此 Secret 的 Pod
		podList, err := clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})
		if err != nil {
			logger.Warn("列出 Pod 失败", zap.Error(err))
		} else {
			for _, pod := range podList.Items {
				// 检查卷引用
				for _, vol := range pod.Spec.Volumes {
					if vol.Secret != nil && vol.Secret.SecretName == name {
						references = append(references, ReferenceInfo{
							Kind:      "Pod",
							Name:      pod.Name,
							Namespace: pod.Namespace,
							RefType:   "volume",
							Field:     fmt.Sprintf("spec.volumes[%s]", vol.Name),
						})
					}
				}

				// 检查容器环境变量引用
				for _, container := range pod.Spec.Containers {
					for _, env := range container.Env {
						if env.ValueFrom != nil && env.ValueFrom.SecretKeyRef != nil {
							if env.ValueFrom.SecretKeyRef.LocalObjectReference.Name == name {
								references = append(references, ReferenceInfo{
									Kind:      "Pod",
									Name:      pod.Name,
									Namespace: pod.Namespace,
									RefType:   "secretKeyRef",
									Field:     fmt.Sprintf("spec.containers[%s].env[%s]", container.Name, env.Name),
								})
							}
						}
					}

					// 检查 imagePullSecrets
					for _, ips := range pod.Spec.ImagePullSecrets {
						if ips.Name == name {
							references = append(references, ReferenceInfo{
								Kind:      "Pod",
								Name:      pod.Name,
								Namespace: pod.Namespace,
								RefType:   "imagePullSecret",
								Field:     "spec.imagePullSecrets",
							})
						}
					}
				}
			}
		}

		// 2. 查找 Deployment
		deployList, err := clientset.AppsV1().Deployments(namespace).List(ctx, metav1.ListOptions{})
		if err != nil {
			logger.Warn("列出 Deployment 失败", zap.Error(err))
		} else {
			for _, dep := range deployList.Items {
				for _, vol := range dep.Spec.Template.Spec.Volumes {
					if vol.Secret != nil && vol.Secret.SecretName == name {
						references = append(references, ReferenceInfo{
							Kind:      "Deployment",
							Name:      dep.Name,
							Namespace: dep.Namespace,
							RefType:   "volume",
							Field:     fmt.Sprintf("spec.template.spec.volumes[%s]", vol.Name),
						})
					}
				}
			}
		}

		// 3. 查找 ServiceAccount
		saList, err := clientset.CoreV1().ServiceAccounts(namespace).List(ctx, metav1.ListOptions{})
		if err != nil {
			logger.Warn("列出 ServiceAccount 失败", zap.Error(err))
		} else {
			for _, sa := range saList.Items {
				for _, secret := range sa.Secrets {
					if secret.Name == name {
						references = append(references, ReferenceInfo{
							Kind:      "ServiceAccount",
							Name:      sa.Name,
							Namespace: sa.Namespace,
							RefType:   "secret",
							Field:     "secrets",
						})
					}
				}
				for _, ips := range sa.ImagePullSecrets {
					if ips.Name == name {
						references = append(references, ReferenceInfo{
							Kind:      "ServiceAccount",
							Name:      sa.Name,
							Namespace: sa.Namespace,
							RefType:   "imagePullSecret",
							Field:     "imagePullSecrets",
						})
					}
				}
			}
		}

		// 4. 查找 Ingress TLS
		ingressList, err := clientset.NetworkingV1().Ingresses(namespace).List(ctx, metav1.ListOptions{})
		if err != nil {
			logger.Warn("列出 Ingress 失败", zap.Error(err))
		} else {
			for _, ingress := range ingressList.Items {
				for _, tls := range ingress.Spec.TLS {
					if tls.SecretName == name {
						references = append(references, ReferenceInfo{
							Kind:      "Ingress",
							Name:      ingress.Name,
							Namespace: ingress.Namespace,
							RefType:   "tls",
							Field:     "spec.tls",
						})
					}
				}
			}
		}

		result := ReferenceList{
			Name:       name,
			Namespace:  namespace,
			Type:       "secret",
			References: references,
		}

		middleware.ResponseSuccess(c, result, "引用查找成功", nil)
	}
}
