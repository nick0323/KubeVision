package service

import (
	"context"

	"go.uber.org/zap"
	appsv1 "k8s.io/api/apps/v1"
	batchv1 "k8s.io/api/batch/v1"
	v1 "k8s.io/api/core/v1"
	networkingv1 "k8s.io/api/networking/v1"
	storagev1 "k8s.io/api/storage/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// 关联资源查询的最大数量限制
const maxRelatedResources = 100

// FindRelatedResources 查找关联资源（支持多种资源类型）
// 返回格式：[]map[string]string{kind, name, relation}
func FindRelatedResources(
	obj interface{},
	resourceType string,
	namespace string,
	clientset kubernetes.Interface,
	ctx context.Context,
	logger *zap.Logger,
) []interface{} {
	result := make([]interface{}, 0, 50) // 预分配容量

	switch o := obj.(type) {
	// ==================== Pod ====================
	case *v1.Pod:
		// 1. 父资源 (OwnerReferences) - ReplicaSet, Deployment, Job 等
		for _, ownerRef := range o.OwnerReferences {
			if len(result) >= maxRelatedResources {
				break
			}
			result = append(result, map[string]string{
				"kind":     ownerRef.Kind,
				"name":     ownerRef.Name,
				"relation": "owner",
			})
		}
		// 2. 关联的 Service (通过 label 匹配)
		if o.Labels != nil && len(result) < maxRelatedResources {
			svcList, err := clientset.CoreV1().Services(namespace).List(ctx, metav1.ListOptions{})
			if err != nil {
				logger.Warn("Failed to query Pod Service", zap.Error(err))
			} else {
				for _, svc := range svcList.Items {
					if len(result) >= maxRelatedResources {
						break
					}
					if svc.Spec.Selector != nil && matchesSelector(o.Labels, svc.Spec.Selector) {
						result = append(result, map[string]string{
							"kind":     "Service",
							"name":     svc.Name,
							"relation": "selectedBy",
						})
					}
				}
			}
		}
		// 3. Volume 相关 - ConfigMap, Secret, PVC
		for _, vol := range o.Spec.Volumes {
			if vol.ConfigMap != nil {
				result = append(result, map[string]string{
					"kind":     "ConfigMap",
					"name":     vol.ConfigMap.Name,
					"relation": "volume",
				})
			}
			if vol.Secret != nil {
				result = append(result, map[string]string{
					"kind":     "Secret",
					"name":     vol.Secret.SecretName,
					"relation": "volume",
				})
			}
			if vol.PersistentVolumeClaim != nil {
				result = append(result, map[string]string{
					"kind":     "PersistentVolumeClaim",
					"name":     vol.PersistentVolumeClaim.ClaimName,
					"relation": "volume",
				})
			}
		}
		// 4. Node
		if o.Spec.NodeName != "" {
			result = append(result, map[string]string{
				"kind":     "Node",
				"name":     o.Spec.NodeName,
				"relation": "scheduledOn",
			})
		}

	// ==================== Deployment ====================
	case *appsv1.Deployment:
		// 1. 子资源 - ReplicaSet
		rsList, err := clientset.AppsV1().ReplicaSets(namespace).List(ctx, metav1.ListOptions{})
		if err != nil {
			logger.Warn("Failed to query Deployment ReplicaSet", zap.Error(err))
		} else {
			for _, rs := range rsList.Items {
				for _, ownerRef := range rs.OwnerReferences {
					if ownerRef.Kind == "Deployment" && ownerRef.Name == o.Name && ownerRef.UID == o.UID {
						result = append(result, map[string]string{
							"kind":     "ReplicaSet",
							"name":     rs.Name,
							"relation": "child",
						})
					}
				}
			}
		}
		// 2. 关联的 Service (通过 selector 匹配)
		if o.Spec.Selector != nil {
			svcList, err := clientset.CoreV1().Services(namespace).List(ctx, metav1.ListOptions{})
			if err != nil {
				logger.Warn("Failed to query Deployment Service", zap.Error(err))
			} else {
				for _, svc := range svcList.Items {
					if svc.Spec.Selector != nil && matchesSelector(o.Spec.Selector.MatchLabels, svc.Spec.Selector) {
						result = append(result, map[string]string{
							"kind":     "Service",
							"name":     svc.Name,
							"relation": "exposedBy",
						})
					}
				}
			}
		}
		// 3. HPA (HorizontalPodAutoscaler)
		hpaList, err := clientset.AutoscalingV1().HorizontalPodAutoscalers(namespace).List(ctx, metav1.ListOptions{})
		if err != nil {
			logger.Warn("Failed to query Deployment HPA", zap.Error(err))
		} else {
			for _, hpa := range hpaList.Items {
				if hpa.Spec.ScaleTargetRef.Kind == "Deployment" && hpa.Spec.ScaleTargetRef.Name == o.Name {
					result = append(result, map[string]string{
						"kind":     "HorizontalPodAutoscaler",
						"name":     hpa.Name,
						"relation": "autoscaled",
					})
				}
			}
		}
		// 4. PDB (PodDisruptionBudget)
		pdbList, err := clientset.PolicyV1().PodDisruptionBudgets(namespace).List(ctx, metav1.ListOptions{})
		if err != nil {
			logger.Warn("Failed to query Deployment PDB", zap.Error(err))
		} else {
			for _, pdb := range pdbList.Items {
				if pdb.Spec.Selector != nil && matchesSelector(o.Spec.Selector.MatchLabels, pdb.Spec.Selector.MatchLabels) {
					result = append(result, map[string]string{
						"kind":     "PodDisruptionBudget",
						"name":     pdb.Name,
						"relation": "protected",
					})
				}
			}
		}
		// 5. Ingress (如果 Service 关联了 Ingress)
		ingressList, err := clientset.NetworkingV1().Ingresses(namespace).List(ctx, metav1.ListOptions{})
		if err != nil {
			logger.Warn("Failed to query Deployment Ingress", zap.Error(err))
		} else {
			// 先找到关联的 Service
			svcNames := make(map[string]bool)
			svcList, _ := clientset.CoreV1().Services(namespace).List(ctx, metav1.ListOptions{})
			for _, svc := range svcList.Items {
				if svc.Spec.Selector != nil && matchesSelector(o.Spec.Selector.MatchLabels, svc.Spec.Selector) {
					svcNames[svc.Name] = true
				}
			}
			// 再找关联这些 Service 的 Ingress
			for _, ing := range ingressList.Items {
				for _, rule := range ing.Spec.Rules {
					if rule.HTTP != nil {
						for _, path := range rule.HTTP.Paths {
							if svcNames[path.Backend.Service.Name] {
								result = append(result, map[string]string{
									"kind":     "Ingress",
									"name":     ing.Name,
									"relation": "routedBy",
								})
							}
						}
					}
				}
			}
		}

	// ==================== StatefulSet ====================
	case *appsv1.StatefulSet:
		// 1. 子资源 - Pod
		podList, err := clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})
		if err != nil {
			logger.Warn("Failed to query StatefulSet Pod", zap.Error(err))
		} else {
			for _, pod := range podList.Items {
				for _, ownerRef := range pod.OwnerReferences {
					if ownerRef.Kind == "StatefulSet" && ownerRef.Name == o.Name && ownerRef.UID == o.UID {
						result = append(result, map[string]string{
							"kind":     "Pod",
							"name":     pod.Name,
							"relation": "child",
						})
					}
				}
			}
		}
		// 2. 关联的 Service
		if o.Spec.Selector != nil {
			svcList, err := clientset.CoreV1().Services(namespace).List(ctx, metav1.ListOptions{})
			if err != nil {
				logger.Warn("Failed to query StatefulSet Service", zap.Error(err))
			} else {
				for _, svc := range svcList.Items {
					if svc.Spec.Selector != nil && matchesSelector(o.Spec.Selector.MatchLabels, svc.Spec.Selector) {
						result = append(result, map[string]string{
							"kind":     "Service",
							"name":     svc.Name,
							"relation": "exposedBy",
						})
					}
				}
			}
		}
		// 3. Headless Service (spec.serviceName)
		if o.Spec.ServiceName != "" {
			result = append(result, map[string]string{
				"kind":     "Service",
				"name":     o.Spec.ServiceName,
				"relation": "headlessService",
			})
		}
		// 4. HPA
		hpaList, err := clientset.AutoscalingV1().HorizontalPodAutoscalers(namespace).List(ctx, metav1.ListOptions{})
		if err != nil {
			logger.Warn("Failed to query StatefulSet HPA", zap.Error(err))
		} else {
			for _, hpa := range hpaList.Items {
				if hpa.Spec.ScaleTargetRef.Kind == "StatefulSet" && hpa.Spec.ScaleTargetRef.Name == o.Name {
					result = append(result, map[string]string{
						"kind":     "HorizontalPodAutoscaler",
						"name":     hpa.Name,
						"relation": "autoscaled",
					})
				}
			}
		}
		// 5. PVC (volumeClaimTemplates)
		for _, pvc := range o.Spec.VolumeClaimTemplates {
			result = append(result, map[string]string{
				"kind":     "PersistentVolumeClaim",
				"name":     pvc.Name,
				"relation": "volumeClaim",
			})
		}

	// ==================== DaemonSet ====================
	case *appsv1.DaemonSet:
		// 1. 子资源 - Pod
		podList, err := clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})
		if err != nil {
			logger.Warn("Failed to query DaemonSet Pod", zap.Error(err))
		} else {
			for _, pod := range podList.Items {
				for _, ownerRef := range pod.OwnerReferences {
					if ownerRef.Kind == "DaemonSet" && ownerRef.Name == o.Name && ownerRef.UID == o.UID {
						result = append(result, map[string]string{
							"kind":     "Pod",
							"name":     pod.Name,
							"relation": "child",
						})
					}
				}
			}
		}
		// 2. 关联的 Service
		if o.Spec.Selector != nil {
			svcList, err := clientset.CoreV1().Services(namespace).List(ctx, metav1.ListOptions{})
			if err != nil {
				logger.Warn("Failed to query DaemonSet Service", zap.Error(err))
			} else {
				for _, svc := range svcList.Items {
					if svc.Spec.Selector != nil && matchesSelector(o.Spec.Selector.MatchLabels, svc.Spec.Selector) {
						result = append(result, map[string]string{
							"kind":     "Service",
							"name":     svc.Name,
							"relation": "exposedBy",
						})
					}
				}
			}
		}

	// ==================== Job ====================
	case *batchv1.Job:
		// 1. 子资源 - Pod
		podList, err := clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})
		if err != nil {
			logger.Warn("Failed to query Job Pod", zap.Error(err))
		} else {
			for _, pod := range podList.Items {
				for _, ownerRef := range pod.OwnerReferences {
					if ownerRef.Kind == "Job" && ownerRef.Name == o.Name && ownerRef.UID == o.UID {
						result = append(result, map[string]string{
							"kind":     "Pod",
							"name":     pod.Name,
							"relation": "child",
						})
					}
				}
			}
		}
		// 2. 父资源 - CronJob
		for _, ownerRef := range o.OwnerReferences {
			if ownerRef.Kind == "CronJob" {
				result = append(result, map[string]string{
					"kind":     "CronJob",
					"name":     ownerRef.Name,
					"relation": "owner",
				})
			}
		}

	// ==================== CronJob ====================
	case *batchv1.CronJob:
		// 1. 子资源 - Job
		jobList, err := clientset.BatchV1().Jobs(namespace).List(ctx, metav1.ListOptions{})
		if err != nil {
			logger.Warn("Failed to query CronJob Job", zap.Error(err))
		} else {
			for _, job := range jobList.Items {
				for _, ownerRef := range job.OwnerReferences {
					if ownerRef.Kind == "CronJob" && ownerRef.Name == o.Name && ownerRef.UID == o.UID {
						result = append(result, map[string]string{
							"kind":     "Job",
							"name":     job.Name,
							"relation": "child",
						})
					}
				}
			}
		}

	// ==================== Service ====================
	case *v1.Service:
		// 1. 关联的 Pod (通过 selector 匹配)
		if o.Spec.Selector != nil {
			podList, err := clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})
			if err != nil {
				logger.Warn("Failed to query Service Pod", zap.Error(err))
			} else {
				for _, pod := range podList.Items {
					if pod.Labels != nil && matchesSelector(pod.Labels, o.Spec.Selector) {
						result = append(result, map[string]string{
							"kind":     "Pod",
							"name":     pod.Name,
							"relation": "selects",
						})
					}
				}
			}
		}
		// 2. Endpoint
		epList, err := clientset.CoreV1().Endpoints(namespace).List(ctx, metav1.ListOptions{})
		if err != nil {
			logger.Warn("Failed to query Service Endpoint", zap.Error(err))
		} else {
			for _, ep := range epList.Items {
				if ep.Name == o.Name {
					result = append(result, map[string]string{
						"kind":     "Endpoints",
						"name":     ep.Name,
						"relation": "endpoints",
					})
				}
			}
		}
		// 3. Ingress
		ingressList, err := clientset.NetworkingV1().Ingresses(namespace).List(ctx, metav1.ListOptions{})
		if err != nil {
			logger.Warn("Failed to query Service Ingress", zap.Error(err))
		} else {
			for _, ing := range ingressList.Items {
				for _, rule := range ing.Spec.Rules {
					if rule.HTTP != nil {
						for _, path := range rule.HTTP.Paths {
							if path.Backend.Service.Name == o.Name {
								result = append(result, map[string]string{
									"kind":     "Ingress",
									"name":     ing.Name,
									"relation": "routedBy",
								})
							}
						}
					}
				}
			}
		}

	// ==================== ConfigMap ====================
	case *v1.ConfigMap:
		// 使用 map 去重
		addedPods := make(map[string]bool)

		// 1. 查询通过 Volume 引用此 ConfigMap 的 Pod
		podList, err := clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})
		if err != nil {
			logger.Warn("Failed to query ConfigMap Pod", zap.Error(err))
		} else {
			for _, pod := range podList.Items {
				for _, vol := range pod.Spec.Volumes {
					if vol.ConfigMap != nil && vol.ConfigMap.Name == o.Name {
						if !addedPods[pod.Name] {
							result = append(result, map[string]string{
								"kind":     "Pod",
								"name":     pod.Name,
								"relation": "usedBy",
							})
							addedPods[pod.Name] = true
						}
						break
					}
				}
			}
		}
		// 2. 查询通过 envFrom.configMapRef 引用此 ConfigMap 的 Pod（containers 和 initContainers）
		for _, pod := range podList.Items {
			if addedPods[pod.Name] {
				continue
			}
			// 检查 containers
			for _, container := range pod.Spec.Containers {
				for _, envFrom := range container.EnvFrom {
					if envFrom.ConfigMapRef != nil && envFrom.ConfigMapRef.Name == o.Name {
						result = append(result, map[string]string{
							"kind":     "Pod",
							"name":     pod.Name,
							"relation": "usedBy",
						})
						addedPods[pod.Name] = true
						break
					}
				}
				if addedPods[pod.Name] {
					break
				}
			}
			// 检查 initContainers
			if !addedPods[pod.Name] {
				for _, container := range pod.Spec.InitContainers {
					for _, envFrom := range container.EnvFrom {
						if envFrom.ConfigMapRef != nil && envFrom.ConfigMapRef.Name == o.Name {
							result = append(result, map[string]string{
								"kind":     "Pod",
								"name":     pod.Name,
								"relation": "usedBy",
							})
							addedPods[pod.Name] = true
							break
						}
					}
					if addedPods[pod.Name] {
						break
					}
				}
			}
		}
		// 3. 查询通过 env.valueFrom.configMapKeyRef 引用此 ConfigMap 的 Pod
		for _, pod := range podList.Items {
			if addedPods[pod.Name] {
				continue
			}
			for _, container := range pod.Spec.Containers {
				for _, env := range container.Env {
					if env.ValueFrom != nil && env.ValueFrom.ConfigMapKeyRef != nil {
						if env.ValueFrom.ConfigMapKeyRef.Name == o.Name {
							result = append(result, map[string]string{
								"kind":     "Pod",
								"name":     pod.Name,
								"relation": "usedBy",
							})
							addedPods[pod.Name] = true
							break
						}
					}
				}
				if addedPods[pod.Name] {
					break
				}
			}
		}
		// 4. 查询通过 projection 引用的 Pod（高级用法）
		for _, pod := range podList.Items {
			if addedPods[pod.Name] {
				continue
			}
			for _, vol := range pod.Spec.Volumes {
				if vol.Projected != nil {
					for _, source := range vol.Projected.Sources {
						if source.ConfigMap != nil && source.ConfigMap.Name == o.Name {
							result = append(result, map[string]string{
								"kind":     "Pod",
								"name":     pod.Name,
								"relation": "usedBy",
							})
							addedPods[pod.Name] = true
							break
						}
					}
				}
				if addedPods[pod.Name] {
					break
				}
			}
		}

	// ==================== Secret ====================
	case *v1.Secret:
		// 使用 map 去重
		addedPods := make(map[string]bool)
		addedSA := make(map[string]bool)

		// 1. 查询通过 Volume 引用此 Secret 的 Pod
		podList, err := clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})
		if err != nil {
			logger.Warn("Failed to query Secret Pod", zap.Error(err))
		} else {
			for _, pod := range podList.Items {
				for _, vol := range pod.Spec.Volumes {
					if vol.Secret != nil && vol.Secret.SecretName == o.Name {
						if !addedPods[pod.Name] {
							result = append(result, map[string]string{
								"kind":     "Pod",
								"name":     pod.Name,
								"relation": "usedBy",
							})
							addedPods[pod.Name] = true
						}
						break
					}
				}
			}
		}
		// 2. 查询通过 imagePullSecrets 引用此 Secret 的 Pod
		for _, pod := range podList.Items {
			for _, ips := range pod.Spec.ImagePullSecrets {
				if ips.Name == o.Name {
					if !addedPods[pod.Name] {
						result = append(result, map[string]string{
							"kind":     "Pod",
							"name":     pod.Name,
							"relation": "usedBy",
						})
						addedPods[pod.Name] = true
					}
					break
				}
			}
		}
		// 3. 查询通过 envFrom.secretRef 引用此 Secret 的 Pod（containers 和 initContainers）
		for _, pod := range podList.Items {
			if addedPods[pod.Name] {
				continue
			}
			// 检查 containers
			for _, container := range pod.Spec.Containers {
				for _, envFrom := range container.EnvFrom {
					if envFrom.SecretRef != nil && envFrom.SecretRef.Name == o.Name {
						result = append(result, map[string]string{
							"kind":     "Pod",
							"name":     pod.Name,
							"relation": "usedBy",
						})
						addedPods[pod.Name] = true
						break
					}
				}
				if addedPods[pod.Name] {
					break
				}
			}
			// 检查 initContainers
			if !addedPods[pod.Name] {
				for _, container := range pod.Spec.InitContainers {
					for _, envFrom := range container.EnvFrom {
						if envFrom.SecretRef != nil && envFrom.SecretRef.Name == o.Name {
							result = append(result, map[string]string{
								"kind":     "Pod",
								"name":     pod.Name,
								"relation": "usedBy",
							})
							addedPods[pod.Name] = true
							break
						}
					}
					if addedPods[pod.Name] {
						break
					}
				}
			}
		}
		// 4. 查询通过 env.valueFrom.secretKeyRef 引用此 Secret 的 Pod
		for _, pod := range podList.Items {
			if addedPods[pod.Name] {
				continue
			}
			for _, container := range pod.Spec.Containers {
				for _, env := range container.Env {
					if env.ValueFrom != nil && env.ValueFrom.SecretKeyRef != nil {
						if env.ValueFrom.SecretKeyRef.Name == o.Name {
							result = append(result, map[string]string{
								"kind":     "Pod",
								"name":     pod.Name,
								"relation": "usedBy",
							})
							addedPods[pod.Name] = true
							break
						}
					}
				}
				if addedPods[pod.Name] {
					break
				}
			}
		}
		// 5. 查询引用此 Secret 的 ServiceAccount
		saList, err := clientset.CoreV1().ServiceAccounts(namespace).List(ctx, metav1.ListOptions{})
		if err != nil {
			logger.Warn("Failed to query Secret ServiceAccount", zap.Error(err))
		} else {
			for _, sa := range saList.Items {
				// 检查 imagePullSecrets
				for _, ips := range sa.ImagePullSecrets {
					if ips.Name == o.Name && !addedSA[sa.Name] {
						result = append(result, map[string]string{
							"kind":     "ServiceAccount",
							"name":     sa.Name,
							"relation": "usedBy",
						})
						addedSA[sa.Name] = true
						break
					}
				}
				// 检查 secrets
				for _, secret := range sa.Secrets {
					if secret.Name == o.Name && !addedSA[sa.Name] {
						result = append(result, map[string]string{
							"kind":     "ServiceAccount",
							"name":     sa.Name,
							"relation": "usedBy",
						})
						addedSA[sa.Name] = true
						break
					}
				}
			}
		}

	// ==================== Ingress ====================
	case *networkingv1.Ingress:
		// 1. 关联的 Service
		for _, rule := range o.Spec.Rules {
			if rule.HTTP != nil {
				for _, path := range rule.HTTP.Paths {
					result = append(result, map[string]string{
						"kind":     "Service",
						"name":     path.Backend.Service.Name,
						"relation": "routesTo",
					})
				}
			}
		}
		// 2. TLS Secret
		for _, tls := range o.Spec.TLS {
			if tls.SecretName != "" {
				result = append(result, map[string]string{
					"kind":     "Secret",
					"name":     tls.SecretName,
					"relation": "tlsSecret",
				})
			}
		}

	// ==================== PVC ====================
	case *v1.PersistentVolumeClaim:
		// 1. 关联的 PV
		if o.Spec.VolumeName != "" {
			result = append(result, map[string]string{
				"kind":     "PersistentVolume",
				"name":     o.Spec.VolumeName,
				"relation": "boundPV",
			})
		}
		// 2. 关联的 Pod
		podList, err := clientset.CoreV1().Pods(namespace).List(ctx, metav1.ListOptions{})
		if err != nil {
			logger.Warn("Failed to query PVC Pod", zap.Error(err))
		} else {
			for _, pod := range podList.Items {
				for _, vol := range pod.Spec.Volumes {
					if vol.PersistentVolumeClaim != nil && vol.PersistentVolumeClaim.ClaimName == o.Name {
						result = append(result, map[string]string{
							"kind":     "Pod",
							"name":     pod.Name,
							"relation": "usedBy",
						})
					}
				}
			}
		}

	// ==================== PV ====================
	case *v1.PersistentVolume:
		// 1. 关联的 PVC
		if o.Spec.ClaimRef != nil {
			result = append(result, map[string]string{
				"kind":     "PersistentVolumeClaim",
				"name":     o.Spec.ClaimRef.Name,
				"relation": "boundPVC",
			})
			// 2. 查询使用此 PVC 的 Pod
			pvcNamespace := o.Spec.ClaimRef.Namespace
			if pvcNamespace != "" {
				podList, err := clientset.CoreV1().Pods(pvcNamespace).List(ctx, metav1.ListOptions{})
				if err != nil {
					logger.Warn("Failed to query PV Pod", zap.Error(err))
				} else {
					for _, pod := range podList.Items {
						for _, vol := range pod.Spec.Volumes {
							if vol.PersistentVolumeClaim != nil && vol.PersistentVolumeClaim.ClaimName == o.Spec.ClaimRef.Name {
								result = append(result, map[string]string{
									"kind":     "Pod",
									"name":     pod.Name,
									"relation": "usedBy",
								})
							}
						}
					}
				}
			}
		}
		// 3. StorageClass
		if o.Spec.StorageClassName != "" {
			result = append(result, map[string]string{
				"kind":     "StorageClass",
				"name":     o.Spec.StorageClassName,
				"relation": "storageClass",
			})
		}

	// ==================== Node ====================
	case *v1.Node:
		// 查询运行在此 Node 上的 Pod
		podList, err := clientset.CoreV1().Pods("").List(ctx, metav1.ListOptions{
			FieldSelector: "spec.nodeName=" + o.Name,
		})
		if err != nil {
			logger.Warn("Failed to query Node Pod", zap.Error(err))
		} else {
			for _, pod := range podList.Items {
				result = append(result, map[string]string{
					"kind":     "Pod",
					"name":     pod.Name,
					"relation": "scheduled",
				})
			}
		}

	// ==================== Namespace ====================
	case *v1.Namespace:
		// 查询 Namespace 中的主要资源数量
		result = append(result, map[string]string{
			"kind":     "ResourceQuota",
			"name":     "ResourceQuota",
			"relation": "quota",
		})
		// 查询此 Namespace 中的 Pod
		podList, err := clientset.CoreV1().Pods(o.Name).List(ctx, metav1.ListOptions{})
		if err != nil {
			logger.Warn("Failed to query Namespace Pod", zap.Error(err))
		} else {
			for _, pod := range podList.Items {
				result = append(result, map[string]string{
					"kind":     "Pod",
					"name":     pod.Name,
					"relation": "contains",
				})
			}
		}

	// ==================== StorageClass ====================
	case *storagev1.StorageClass:
		// 查询使用此 StorageClass 的 PVC
		pvcList, err := clientset.CoreV1().PersistentVolumeClaims("").List(ctx, metav1.ListOptions{})
		if err != nil {
			logger.Warn("Failed to query StorageClass PVC", zap.Error(err))
		} else {
			for _, pvc := range pvcList.Items {
				if pvc.Spec.StorageClassName != nil && *pvc.Spec.StorageClassName == o.Name {
					result = append(result, map[string]string{
						"kind":     "PersistentVolumeClaim",
						"name":     pvc.Name,
						"relation": "provisionedPVC",
					})
				}
			}
		}
		// 查询使用此 StorageClass 的 PV
		pvList, err := clientset.CoreV1().PersistentVolumes().List(ctx, metav1.ListOptions{})
		if err != nil {
			logger.Warn("Failed to query StorageClass PV", zap.Error(err))
		} else {
			for _, pv := range pvList.Items {
				if pv.Spec.StorageClassName == o.Name {
					result = append(result, map[string]string{
						"kind":     "PersistentVolume",
						"name":     pv.Name,
						"relation": "provisionedPV",
					})
				}
			}
		}
	}

	return result
}

// matchesSelector 检查 labels 是否匹配 selector
func matchesSelector(labels, selector map[string]string) bool {
	if len(selector) == 0 {
		return false
	}
	for k, v := range selector {
		if labels[k] != v {
			return false
		}
	}
	return true
}
