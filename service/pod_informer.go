package service

import (
	"context"
	"sync"

	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/tools/cache"
)

type PodInformer struct {
	clientset    *kubernetes.Clientset
	informer     cache.SharedIndexInformer
	podCache     map[string]*v1.Pod
	restartCache map[string]int32
	podRestarts  map[string]int32
	mu           sync.RWMutex
}

func NewPodInformer(clientset *kubernetes.Clientset, namespace string) *PodInformer {
	pi := &PodInformer{
		clientset:    clientset,
		podCache:     make(map[string]*v1.Pod),
		restartCache: make(map[string]int32),
		podRestarts:  make(map[string]int32),
	}

	informerFactory := informers.NewSharedInformerFactoryWithOptions(
		clientset,
		0,
		informers.WithNamespace(namespace),
	)

	pi.informer = informerFactory.Core().V1().Pods().Informer()
	pi.informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    pi.onPodAdd,
		UpdateFunc: pi.onPodUpdate,
		DeleteFunc: pi.onPodDelete,
	})

	return pi
}

func (pi *PodInformer) Start(ctx context.Context) {
	pi.informer.Run(ctx.Done())
}

func (pi *PodInformer) HasSynced() bool {
	return pi.informer.HasSynced()
}

func (pi *PodInformer) onPodAdd(obj interface{}) {
	pod, ok := obj.(*v1.Pod)
	if !ok {
		return
	}
	key := pod.Namespace + "/" + pod.Name
	pi.mu.Lock()
	defer pi.mu.Unlock()
	pi.podCache[key] = pod
	pi.updateOwnerRestarts(pod)
}

func (pi *PodInformer) onPodUpdate(oldObj, newObj interface{}) {
	newPod, ok := newObj.(*v1.Pod)
	if !ok {
		return
	}
	key := newPod.Namespace + "/" + newPod.Name
	pi.mu.Lock()
	defer pi.mu.Unlock()
	pi.podCache[key] = newPod
	pi.updateOwnerRestarts(newPod)
}

func (pi *PodInformer) onPodDelete(obj interface{}) {
	pod, ok := obj.(*v1.Pod)
	if !ok {
		return
	}
	key := pod.Namespace + "/" + pod.Name
	pi.mu.Lock()
	defer pi.mu.Unlock()

	podRestarts := pi.podRestarts[key]
	delete(pi.podCache, key)
	delete(pi.podRestarts, key)

	for _, ownerRef := range pod.OwnerReferences {
		ownerKey := string(ownerRef.UID)
		pi.restartCache[ownerKey] -= podRestarts
		if pi.restartCache[ownerKey] <= 0 {
			delete(pi.restartCache, ownerKey)
		}
	}
}

func (pi *PodInformer) updateOwnerRestarts(pod *v1.Pod) {
	key := pod.Namespace + "/" + pod.Name

	var currentRestarts int32
	for _, cs := range pod.Status.ContainerStatuses {
		currentRestarts += cs.RestartCount
	}

	lastRestarts := pi.podRestarts[key]
	pi.podRestarts[key] = currentRestarts
	delta := currentRestarts - lastRestarts

	for _, ownerRef := range pod.OwnerReferences {
		ownerKey := string(ownerRef.UID)
		pi.restartCache[ownerKey] += delta
	}
}

func (pi *PodInformer) GetRestartsByOwnerUID(ownerUID string) int32 {
	pi.mu.RLock()
	defer pi.mu.RUnlock()
	return pi.restartCache[ownerUID]
}

func (pi *PodInformer) GetRestartsByOwnerReference(ownerRef metav1.OwnerReference) int32 {
	return pi.GetRestartsByOwnerUID(string(ownerRef.UID))
}

func (pi *PodInformer) GetPod(namespace, name string) (*v1.Pod, bool) {
	pi.mu.RLock()
	defer pi.mu.RUnlock()
	pod, exists := pi.podCache[namespace+"/"+name]
	return pod, exists
}

var podInformerInstance *PodInformer

func SetPodInformer(pi *PodInformer) {
	podInformerInstance = pi
}

func GetPodInformer() *PodInformer {
	return podInformerInstance
}

func GetRestartsByOwnerReference(ownerRef metav1.OwnerReference) int32 {
	if podInformerInstance == nil {
		return 0
	}
	return podInformerInstance.GetRestartsByOwnerReference(ownerRef)
}
