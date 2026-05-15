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

const maxRelatedResources = 100

type findContext struct {
	obj       any
	ctx       context.Context
	clientset kubernetes.Interface
	namespace string
	logger    *zap.Logger
	result    []any
}

type RelatedResourceFinder interface {
	Find(fc *findContext)
}

var relatedFinders = map[string]RelatedResourceFinder{}

func registerFinder(kind string, f RelatedResourceFinder) {
	relatedFinders[kind] = f
}

func init() {
	registerFinder("Pod", &podFinder{})
	registerFinder("Deployment", &deploymentFinder{})
	registerFinder("StatefulSet", &statefulSetFinder{})
	registerFinder("DaemonSet", &daemonSetFinder{})
	registerFinder("Job", &jobFinder{})
	registerFinder("CronJob", &cronJobFinder{})
	registerFinder("Service", &serviceFinder{})
	registerFinder("ConfigMap", &configMapFinder{})
	registerFinder("Secret", &secretFinder{})
	registerFinder("Ingress", &ingressFinder{})
	registerFinder("PersistentVolumeClaim", &pvcFinder{})
	registerFinder("PersistentVolume", &pvFinder{})
	registerFinder("Node", &nodeFinder{})
	registerFinder("Namespace", &namespaceFinder{})
	registerFinder("StorageClass", &storageClassFinder{})
}

func FindRelatedResources(
	obj any,
	resourceType string,
	namespace string,
	clientset kubernetes.Interface,
	ctx context.Context,
	logger *zap.Logger,
) []any {
	fc := &findContext{
		obj:       obj,
		ctx:       ctx,
		clientset: clientset,
		namespace: namespace,
		logger:    logger,
		result:    make([]any, 0, 50),
	}

	var finder RelatedResourceFinder
	switch obj.(type) {
	case *v1.Pod:
		finder = relatedFinders["Pod"]
	case *appsv1.Deployment:
		finder = relatedFinders["Deployment"]
	case *appsv1.StatefulSet:
		finder = relatedFinders["StatefulSet"]
	case *appsv1.DaemonSet:
		finder = relatedFinders["DaemonSet"]
	case *batchv1.Job:
		finder = relatedFinders["Job"]
	case *batchv1.CronJob:
		finder = relatedFinders["CronJob"]
	case *v1.Service:
		finder = relatedFinders["Service"]
	case *v1.ConfigMap:
		finder = relatedFinders["ConfigMap"]
	case *v1.Secret:
		finder = relatedFinders["Secret"]
	case *networkingv1.Ingress:
		finder = relatedFinders["Ingress"]
	case *v1.PersistentVolumeClaim:
		finder = relatedFinders["PersistentVolumeClaim"]
	case *v1.PersistentVolume:
		finder = relatedFinders["PersistentVolume"]
	case *v1.Node:
		finder = relatedFinders["Node"]
	case *v1.Namespace:
		finder = relatedFinders["Namespace"]
	case *storagev1.StorageClass:
		finder = relatedFinders["StorageClass"]
	}
	if finder != nil {
		finder.Find(fc)
	}
	return fc.result
}

func (fc *findContext) add(kind, name, relation string) {
	if len(fc.result) >= maxRelatedResources {
		return
	}
	fc.result = append(fc.result, map[string]string{
		"kind":     kind,
		"name":     name,
		"relation": relation,
	})
}

type podFinder struct{}

func (f *podFinder) Find(fc *findContext) {
	o := fc.obj.(*v1.Pod)
	for _, ownerRef := range o.OwnerReferences {
		fc.add(ownerRef.Kind, ownerRef.Name, "owner")
	}
	if o.Labels != nil {
		svcList, err := fc.clientset.CoreV1().Services(fc.namespace).List(fc.ctx, metav1.ListOptions{})
		if err != nil {
			fc.logger.Warn("Failed to query Pod Service", zap.Error(err))
		} else {
			for _, svc := range svcList.Items {
				if svc.Spec.Selector != nil && matchesSelector(o.Labels, svc.Spec.Selector) {
					fc.add("Service", svc.Name, "selectedBy")
				}
			}
		}
	}
	for _, vol := range o.Spec.Volumes {
		if vol.ConfigMap != nil {
			fc.add("ConfigMap", vol.ConfigMap.Name, "volume")
		}
		if vol.Secret != nil {
			fc.add("Secret", vol.Secret.SecretName, "volume")
		}
		if vol.PersistentVolumeClaim != nil {
			fc.add("PersistentVolumeClaim", vol.PersistentVolumeClaim.ClaimName, "volume")
		}
	}
	if o.Spec.NodeName != "" {
		fc.add("Node", o.Spec.NodeName, "scheduledOn")
	}
}

type deploymentFinder struct{}

func (f *deploymentFinder) Find(fc *findContext) {
	o := fc.obj.(*appsv1.Deployment)
	rsList, err := fc.clientset.AppsV1().ReplicaSets(fc.namespace).List(fc.ctx, metav1.ListOptions{})
	if err != nil {
		fc.logger.Warn("Failed to query Deployment ReplicaSet", zap.Error(err))
	} else {
		for _, rs := range rsList.Items {
			for _, ownerRef := range rs.OwnerReferences {
				if ownerRef.Kind == "Deployment" && ownerRef.Name == o.Name && ownerRef.UID == o.UID {
					fc.add("ReplicaSet", rs.Name, "child")
				}
			}
		}
	}
	if o.Spec.Selector != nil {
		svcList, err := fc.clientset.CoreV1().Services(fc.namespace).List(fc.ctx, metav1.ListOptions{})
		if err != nil {
			fc.logger.Warn("Failed to query Deployment Service", zap.Error(err))
		} else {
			for _, svc := range svcList.Items {
				if svc.Spec.Selector != nil && matchesSelector(o.Spec.Selector.MatchLabels, svc.Spec.Selector) {
					fc.add("Service", svc.Name, "exposedBy")
				}
			}
		}
	}
	hpaList, err := fc.clientset.AutoscalingV1().HorizontalPodAutoscalers(fc.namespace).List(fc.ctx, metav1.ListOptions{})
	if err != nil {
		fc.logger.Warn("Failed to query Deployment HPA", zap.Error(err))
	} else {
		for _, hpa := range hpaList.Items {
			if hpa.Spec.ScaleTargetRef.Kind == "Deployment" && hpa.Spec.ScaleTargetRef.Name == o.Name {
				fc.add("HorizontalPodAutoscaler", hpa.Name, "autoscaled")
			}
		}
	}
	pdbList, err := fc.clientset.PolicyV1().PodDisruptionBudgets(fc.namespace).List(fc.ctx, metav1.ListOptions{})
	if err != nil {
		fc.logger.Warn("Failed to query Deployment PDB", zap.Error(err))
	} else {
		for _, pdb := range pdbList.Items {
			if pdb.Spec.Selector != nil && matchesSelector(o.Spec.Selector.MatchLabels, pdb.Spec.Selector.MatchLabels) {
				fc.add("PodDisruptionBudget", pdb.Name, "protected")
			}
		}
	}
	ingressList, err := fc.clientset.NetworkingV1().Ingresses(fc.namespace).List(fc.ctx, metav1.ListOptions{})
	if err != nil {
		fc.logger.Warn("Failed to query Deployment Ingress", zap.Error(err))
	} else {
		svcNames := make(map[string]bool)
		svcList, _ := fc.clientset.CoreV1().Services(fc.namespace).List(fc.ctx, metav1.ListOptions{})
		for _, svc := range svcList.Items {
			if svc.Spec.Selector != nil && matchesSelector(o.Spec.Selector.MatchLabels, svc.Spec.Selector) {
				svcNames[svc.Name] = true
			}
		}
		for _, ing := range ingressList.Items {
			for _, rule := range ing.Spec.Rules {
				if rule.HTTP != nil {
					for _, path := range rule.HTTP.Paths {
						if svcNames[path.Backend.Service.Name] {
							fc.add("Ingress", ing.Name, "routedBy")
						}
					}
				}
			}
		}
	}
}

type statefulSetFinder struct{}

func (f *statefulSetFinder) Find(fc *findContext) {
	o := fc.obj.(*appsv1.StatefulSet)
	podList, err := fc.clientset.CoreV1().Pods(fc.namespace).List(fc.ctx, metav1.ListOptions{})
	if err != nil {
		fc.logger.Warn("Failed to query StatefulSet Pod", zap.Error(err))
	} else {
		for _, pod := range podList.Items {
			for _, ownerRef := range pod.OwnerReferences {
				if ownerRef.Kind == "StatefulSet" && ownerRef.Name == o.Name && ownerRef.UID == o.UID {
					fc.add("Pod", pod.Name, "child")
				}
			}
		}
	}
	if o.Spec.Selector != nil {
		svcList, err := fc.clientset.CoreV1().Services(fc.namespace).List(fc.ctx, metav1.ListOptions{})
		if err != nil {
			fc.logger.Warn("Failed to query StatefulSet Service", zap.Error(err))
		} else {
			for _, svc := range svcList.Items {
				if svc.Spec.Selector != nil && matchesSelector(o.Spec.Selector.MatchLabels, svc.Spec.Selector) {
					fc.add("Service", svc.Name, "exposedBy")
				}
			}
		}
	}
	if o.Spec.ServiceName != "" {
		fc.add("Service", o.Spec.ServiceName, "headlessService")
	}
	hpaList, err := fc.clientset.AutoscalingV1().HorizontalPodAutoscalers(fc.namespace).List(fc.ctx, metav1.ListOptions{})
	if err != nil {
		fc.logger.Warn("Failed to query StatefulSet HPA", zap.Error(err))
	} else {
		for _, hpa := range hpaList.Items {
			if hpa.Spec.ScaleTargetRef.Kind == "StatefulSet" && hpa.Spec.ScaleTargetRef.Name == o.Name {
				fc.add("HorizontalPodAutoscaler", hpa.Name, "autoscaled")
			}
		}
	}
	for _, pvc := range o.Spec.VolumeClaimTemplates {
		fc.add("PersistentVolumeClaim", pvc.Name, "volumeClaim")
	}
}

type daemonSetFinder struct{}

func (f *daemonSetFinder) Find(fc *findContext) {
	o := fc.obj.(*appsv1.DaemonSet)
	podList, err := fc.clientset.CoreV1().Pods(fc.namespace).List(fc.ctx, metav1.ListOptions{})
	if err != nil {
		fc.logger.Warn("Failed to query DaemonSet Pod", zap.Error(err))
	} else {
		for _, pod := range podList.Items {
			for _, ownerRef := range pod.OwnerReferences {
				if ownerRef.Kind == "DaemonSet" && ownerRef.Name == o.Name && ownerRef.UID == o.UID {
					fc.add("Pod", pod.Name, "child")
				}
			}
		}
	}
	if o.Spec.Selector != nil {
		svcList, err := fc.clientset.CoreV1().Services(fc.namespace).List(fc.ctx, metav1.ListOptions{})
		if err != nil {
			fc.logger.Warn("Failed to query DaemonSet Service", zap.Error(err))
		} else {
			for _, svc := range svcList.Items {
				if svc.Spec.Selector != nil && matchesSelector(o.Spec.Selector.MatchLabels, svc.Spec.Selector) {
					fc.add("Service", svc.Name, "exposedBy")
				}
			}
		}
	}
}

type jobFinder struct{}

func (f *jobFinder) Find(fc *findContext) {
	o := fc.obj.(*batchv1.Job)
	podList, err := fc.clientset.CoreV1().Pods(fc.namespace).List(fc.ctx, metav1.ListOptions{})
	if err != nil {
		fc.logger.Warn("Failed to query Job Pod", zap.Error(err))
	} else {
		for _, pod := range podList.Items {
			for _, ownerRef := range pod.OwnerReferences {
				if ownerRef.Kind == "Job" && ownerRef.Name == o.Name && ownerRef.UID == o.UID {
					fc.add("Pod", pod.Name, "child")
				}
			}
		}
	}
	for _, ownerRef := range o.OwnerReferences {
		if ownerRef.Kind == "CronJob" {
			fc.add("CronJob", ownerRef.Name, "owner")
		}
	}
}

type cronJobFinder struct{}

func (f *cronJobFinder) Find(fc *findContext) {
	o := fc.obj.(*batchv1.CronJob)
	jobList, err := fc.clientset.BatchV1().Jobs(fc.namespace).List(fc.ctx, metav1.ListOptions{})
	if err != nil {
		fc.logger.Warn("Failed to query CronJob Job", zap.Error(err))
	} else {
		for _, job := range jobList.Items {
			for _, ownerRef := range job.OwnerReferences {
				if ownerRef.Kind == "CronJob" && ownerRef.Name == o.Name && ownerRef.UID == o.UID {
					fc.add("Job", job.Name, "child")
				}
			}
		}
	}
}

type serviceFinder struct{}

func (f *serviceFinder) Find(fc *findContext) {
	o := fc.obj.(*v1.Service)
	if o.Spec.Selector != nil {
		podList, err := fc.clientset.CoreV1().Pods(fc.namespace).List(fc.ctx, metav1.ListOptions{})
		if err != nil {
			fc.logger.Warn("Failed to query Service Pod", zap.Error(err))
		} else {
			for _, pod := range podList.Items {
				if pod.Labels != nil && matchesSelector(pod.Labels, o.Spec.Selector) {
					fc.add("Pod", pod.Name, "selects")
				}
			}
		}
	}
	epList, err := fc.clientset.CoreV1().Endpoints(fc.namespace).List(fc.ctx, metav1.ListOptions{})
	if err != nil {
		fc.logger.Warn("Failed to query Service Endpoint", zap.Error(err))
	} else {
		for _, ep := range epList.Items {
			if ep.Name == o.Name {
				fc.add("Endpoints", ep.Name, "endpoints")
			}
		}
	}
	ingressList, err := fc.clientset.NetworkingV1().Ingresses(fc.namespace).List(fc.ctx, metav1.ListOptions{})
	if err != nil {
		fc.logger.Warn("Failed to query Service Ingress", zap.Error(err))
	} else {
		for _, ing := range ingressList.Items {
			for _, rule := range ing.Spec.Rules {
				if rule.HTTP != nil {
					for _, path := range rule.HTTP.Paths {
						if path.Backend.Service.Name == o.Name {
							fc.add("Ingress", ing.Name, "routedBy")
						}
					}
				}
			}
		}
	}
}

type configMapFinder struct{}

func (f *configMapFinder) Find(fc *findContext) {
	o := fc.obj.(*v1.ConfigMap)
	addedPods := make(map[string]bool)

	podList, err := fc.clientset.CoreV1().Pods(fc.namespace).List(fc.ctx, metav1.ListOptions{})
	if err != nil {
		fc.logger.Warn("Failed to query ConfigMap Pod", zap.Error(err))
		return
	}
	for _, pod := range podList.Items {
		if addedPods[pod.Name] {
			continue
		}
		for _, vol := range pod.Spec.Volumes {
			if vol.ConfigMap != nil && vol.ConfigMap.Name == o.Name {
				fc.add("Pod", pod.Name, "usedBy")
				addedPods[pod.Name] = true
				break
			}
		}
		if addedPods[pod.Name] {
			continue
		}
		found := false
		for _, container := range pod.Spec.Containers {
			for _, envFrom := range container.EnvFrom {
				if envFrom.ConfigMapRef != nil && envFrom.ConfigMapRef.Name == o.Name {
					fc.add("Pod", pod.Name, "usedBy")
					addedPods[pod.Name] = true
					found = true
					break
				}
			}
			if found {
				break
			}
		}
		if !found {
			for _, container := range pod.Spec.InitContainers {
				for _, envFrom := range container.EnvFrom {
					if envFrom.ConfigMapRef != nil && envFrom.ConfigMapRef.Name == o.Name {
						fc.add("Pod", pod.Name, "usedBy")
						addedPods[pod.Name] = true
						found = true
						break
					}
				}
				if found {
					break
				}
			}
		}
		if addedPods[pod.Name] {
			continue
		}
		for _, container := range pod.Spec.Containers {
			for _, env := range container.Env {
				if env.ValueFrom != nil && env.ValueFrom.ConfigMapKeyRef != nil && env.ValueFrom.ConfigMapKeyRef.Name == o.Name {
					fc.add("Pod", pod.Name, "usedBy")
					addedPods[pod.Name] = true
					break
				}
			}
			if addedPods[pod.Name] {
				break
			}
		}
		if addedPods[pod.Name] {
			continue
		}
		for _, vol := range pod.Spec.Volumes {
			if vol.Projected != nil {
				for _, source := range vol.Projected.Sources {
					if source.ConfigMap != nil && source.ConfigMap.Name == o.Name {
						fc.add("Pod", pod.Name, "usedBy")
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
}

type secretFinder struct{}

func (f *secretFinder) Find(fc *findContext) {
	o := fc.obj.(*v1.Secret)
	addedPods := make(map[string]bool)
	addedSA := make(map[string]bool)

	podList, err := fc.clientset.CoreV1().Pods(fc.namespace).List(fc.ctx, metav1.ListOptions{})
	if err != nil {
		fc.logger.Warn("Failed to query Secret Pod", zap.Error(err))
	} else {
		for _, pod := range podList.Items {
			if addedPods[pod.Name] {
				continue
			}
			for _, vol := range pod.Spec.Volumes {
				if vol.Secret != nil && vol.Secret.SecretName == o.Name {
					fc.add("Pod", pod.Name, "usedBy")
					addedPods[pod.Name] = true
					break
				}
			}
			if addedPods[pod.Name] {
				continue
			}
			for _, ips := range pod.Spec.ImagePullSecrets {
				if ips.Name == o.Name {
					fc.add("Pod", pod.Name, "usedBy")
					addedPods[pod.Name] = true
					break
				}
			}
			if addedPods[pod.Name] {
				continue
			}
			found := false
			for _, container := range pod.Spec.Containers {
				for _, envFrom := range container.EnvFrom {
					if envFrom.SecretRef != nil && envFrom.SecretRef.Name == o.Name {
						fc.add("Pod", pod.Name, "usedBy")
						addedPods[pod.Name] = true
						found = true
						break
					}
				}
				if found {
					break
				}
			}
			if !found {
				for _, container := range pod.Spec.InitContainers {
					for _, envFrom := range container.EnvFrom {
						if envFrom.SecretRef != nil && envFrom.SecretRef.Name == o.Name {
							fc.add("Pod", pod.Name, "usedBy")
							addedPods[pod.Name] = true
							found = true
							break
						}
					}
					if found {
						break
					}
				}
			}
			if addedPods[pod.Name] {
				continue
			}
			for _, container := range pod.Spec.Containers {
				for _, env := range container.Env {
					if env.ValueFrom != nil && env.ValueFrom.SecretKeyRef != nil && env.ValueFrom.SecretKeyRef.Name == o.Name {
						fc.add("Pod", pod.Name, "usedBy")
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
	saList, err := fc.clientset.CoreV1().ServiceAccounts(fc.namespace).List(fc.ctx, metav1.ListOptions{})
	if err != nil {
		fc.logger.Warn("Failed to query Secret ServiceAccount", zap.Error(err))
		return
	}
	for _, sa := range saList.Items {
		for _, ips := range sa.ImagePullSecrets {
			if ips.Name == o.Name && !addedSA[sa.Name] {
				fc.add("ServiceAccount", sa.Name, "usedBy")
				addedSA[sa.Name] = true
				break
			}
		}
		for _, secret := range sa.Secrets {
			if secret.Name == o.Name && !addedSA[sa.Name] {
				fc.add("ServiceAccount", sa.Name, "usedBy")
				addedSA[sa.Name] = true
				break
			}
		}
	}
}

type ingressFinder struct{}

func (f *ingressFinder) Find(fc *findContext) {
	o := fc.obj.(*networkingv1.Ingress)
	for _, rule := range o.Spec.Rules {
		if rule.HTTP != nil {
			for _, path := range rule.HTTP.Paths {
				fc.add("Service", path.Backend.Service.Name, "routesTo")
			}
		}
	}
	for _, tls := range o.Spec.TLS {
		if tls.SecretName != "" {
			fc.add("Secret", tls.SecretName, "tlsSecret")
		}
	}
}

type pvcFinder struct{}

func (f *pvcFinder) Find(fc *findContext) {
	o := fc.obj.(*v1.PersistentVolumeClaim)
	if o.Spec.VolumeName != "" {
		fc.add("PersistentVolume", o.Spec.VolumeName, "boundPV")
	}
	podList, err := fc.clientset.CoreV1().Pods(fc.namespace).List(fc.ctx, metav1.ListOptions{})
	if err != nil {
		fc.logger.Warn("Failed to query PVC Pod", zap.Error(err))
		return
	}
	for _, pod := range podList.Items {
		for _, vol := range pod.Spec.Volumes {
			if vol.PersistentVolumeClaim != nil && vol.PersistentVolumeClaim.ClaimName == o.Name {
				fc.add("Pod", pod.Name, "usedBy")
			}
		}
	}
}

type pvFinder struct{}

func (f *pvFinder) Find(fc *findContext) {
	o := fc.obj.(*v1.PersistentVolume)
	if o.Spec.ClaimRef != nil {
		fc.add("PersistentVolumeClaim", o.Spec.ClaimRef.Name, "boundPVC")
		pvcNamespace := o.Spec.ClaimRef.Namespace
		if pvcNamespace != "" {
			podList, err := fc.clientset.CoreV1().Pods(pvcNamespace).List(fc.ctx, metav1.ListOptions{})
			if err != nil {
				fc.logger.Warn("Failed to query PV Pod", zap.Error(err))
			} else {
				for _, pod := range podList.Items {
					for _, vol := range pod.Spec.Volumes {
						if vol.PersistentVolumeClaim != nil && vol.PersistentVolumeClaim.ClaimName == o.Spec.ClaimRef.Name {
							fc.add("Pod", pod.Name, "usedBy")
						}
					}
				}
			}
		}
	}
	if o.Spec.StorageClassName != "" {
		fc.add("StorageClass", o.Spec.StorageClassName, "storageClass")
	}
}

type nodeFinder struct{}

func (f *nodeFinder) Find(fc *findContext) {
	o := fc.obj.(*v1.Node)
	podList, err := fc.clientset.CoreV1().Pods("").List(fc.ctx, metav1.ListOptions{
		FieldSelector: "spec.nodeName=" + o.Name,
	})
	if err != nil {
		fc.logger.Warn("Failed to query Node Pod", zap.Error(err))
		return
	}
	for _, pod := range podList.Items {
		fc.add("Pod", pod.Name, "scheduled")
	}
}

type namespaceFinder struct{}

func (f *namespaceFinder) Find(fc *findContext) {
	o := fc.obj.(*v1.Namespace)
	fc.add("ResourceQuota", "ResourceQuota", "quota")
	podList, err := fc.clientset.CoreV1().Pods(o.Name).List(fc.ctx, metav1.ListOptions{})
	if err != nil {
		fc.logger.Warn("Failed to query Namespace Pod", zap.Error(err))
		return
	}
	for _, pod := range podList.Items {
		fc.add("Pod", pod.Name, "contains")
	}
}

type storageClassFinder struct{}

func (f *storageClassFinder) Find(fc *findContext) {
	o := fc.obj.(*storagev1.StorageClass)
	pvcList, err := fc.clientset.CoreV1().PersistentVolumeClaims("").List(fc.ctx, metav1.ListOptions{})
	if err != nil {
		fc.logger.Warn("Failed to query StorageClass PVC", zap.Error(err))
	} else {
		for _, pvc := range pvcList.Items {
			if pvc.Spec.StorageClassName != nil && *pvc.Spec.StorageClassName == o.Name {
				fc.add("PersistentVolumeClaim", pvc.Name, "provisionedPVC")
			}
		}
	}
	pvList, err := fc.clientset.CoreV1().PersistentVolumes().List(fc.ctx, metav1.ListOptions{})
	if err != nil {
		fc.logger.Warn("Failed to query StorageClass PV", zap.Error(err))
		return
	}
	for _, pv := range pvList.Items {
		if pv.Spec.StorageClassName == o.Name {
			fc.add("PersistentVolume", pv.Name, "provisionedPV")
		}
	}
}

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
