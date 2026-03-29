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

// PodInformer Pod Informer 管理器
type PodInformer struct {
	clientset    *kubernetes.Clientset
	informer     cache.SharedIndexInformer
	podCache     map[string]*v1.Pod // key: namespace/name
	restartCache map[string]int32   // key: ownerUID, value: restarts
	podRestarts  map[string]int32   // key: namespace/name, value: restarts (用于去重)
	mu           sync.RWMutex
}

// NewPodInformer 创建 Pod Informer
func NewPodInformer(clientset *kubernetes.Clientset, namespace string) *PodInformer {
	pi := &PodInformer{
		clientset:    clientset,
		podCache:     make(map[string]*v1.Pod),
		restartCache: make(map[string]int32),
		podRestarts:  make(map[string]int32),
	}

	// 创建 Informer
	informerFactory := informers.NewSharedInformerFactoryWithOptions(
		clientset,
		0, // 不设置 resync，完全依赖事件驱动
		informers.WithNamespace(namespace),
	)

	pi.informer = informerFactory.Core().V1().Pods().Informer()

	// 注册事件处理
	pi.informer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc:    pi.onPodAdd,
		UpdateFunc: pi.onPodUpdate,
		DeleteFunc: pi.onPodDelete,
	})

	return pi
}

// Start 启动 Informer
func (pi *PodInformer) Start(ctx context.Context) {
	pi.informer.Run(ctx.Done())
}

// HasSynced 检查是否已同步
func (pi *PodInformer) HasSynced() bool {
	return pi.informer.HasSynced()
}

// onPodAdd Pod 添加事件
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

// onPodUpdate Pod 更新事件
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

// onPodDelete Pod 删除事件
func (pi *PodInformer) onPodDelete(obj interface{}) {
	pod, ok := obj.(*v1.Pod)
	if !ok {
		return
	}

	key := pod.Namespace + "/" + pod.Name
	pi.mu.Lock()
	defer pi.mu.Unlock()

	// 获取该 Pod 的重启次数
	podRestarts := pi.podRestarts[key]

	// 从缓存中删除
	delete(pi.podCache, key)
	delete(pi.podRestarts, key)

	// 从 Owner 缓存中减去该 Pod 的重启次数
	for _, ownerRef := range pod.OwnerReferences {
		ownerKey := string(ownerRef.UID)
		pi.restartCache[ownerKey] -= podRestarts
		if pi.restartCache[ownerKey] <= 0 {
			delete(pi.restartCache, ownerKey)
		}
	}
}

// updateOwnerRestarts 更新 Owner 的重启次数
func (pi *PodInformer) updateOwnerRestarts(pod *v1.Pod) {
	key := pod.Namespace + "/" + pod.Name

	// 计算当前 Pod 的重启次数
	var currentRestarts int32
	for _, cs := range pod.Status.ContainerStatuses {
		currentRestarts += cs.RestartCount
	}

	// 获取上次记录的重启次数
	lastRestarts := pi.podRestarts[key]
	pi.podRestarts[key] = currentRestarts

	// 计算差值（可能为正、负或零）
	delta := currentRestarts - lastRestarts

	// 更新所有 Owner 的重启次数
	for _, ownerRef := range pod.OwnerReferences {
		ownerKey := string(ownerRef.UID)
		pi.restartCache[ownerKey] += delta
	}
}

// hasOwnerReference 检查 Pod 是否有指定的 Owner
func hasOwnerReference(pod *v1.Pod, ownerUID string) bool {
	for _, ownerRef := range pod.OwnerReferences {
		if string(ownerRef.UID) == ownerUID {
			return true
		}
	}
	return false
}

// GetRestartsByOwnerUID 根据 Owner UID 获取重启次数
func (pi *PodInformer) GetRestartsByOwnerUID(ownerUID string) int32 {
	pi.mu.RLock()
	defer pi.mu.RUnlock()
	return pi.restartCache[ownerUID]
}

// GetRestartsByOwnerReference 根据 OwnerReference 获取重启次数
func (pi *PodInformer) GetRestartsByOwnerReference(ownerRef metav1.OwnerReference) int32 {
	return pi.GetRestartsByOwnerUID(string(ownerRef.UID))
}

// GetPod 获取 Pod
func (pi *PodInformer) GetPod(namespace, name string) (*v1.Pod, bool) {
	pi.mu.RLock()
	defer pi.mu.RUnlock()
	pod, exists := pi.podCache[namespace+"/"+name]
	return pod, exists
}

// GetPodsByOwner 根据 Owner 获取所有 Pod
func (pi *PodInformer) GetPodsByOwner(ownerUID string) []*v1.Pod {
	pi.mu.RLock()
	defer pi.mu.RUnlock()

	var pods []*v1.Pod
	for _, pod := range pi.podCache {
		if hasOwnerReference(pod, ownerUID) {
			pods = append(pods, pod)
		}
	}
	return pods
}

// GetPodCount 获取 Pod 数量
func (pi *PodInformer) GetPodCount() int {
	pi.mu.RLock()
	defer pi.mu.RUnlock()
	return len(pi.podCache)
}

// GetRestartCache 获取重启缓存（用于调试）
func (pi *PodInformer) GetRestartCache() map[string]int32 {
	pi.mu.RLock()
	defer pi.mu.RUnlock()

	cacheCopy := make(map[string]int32, len(pi.restartCache))
	for k, v := range pi.restartCache {
		cacheCopy[k] = v
	}
	return cacheCopy
}
