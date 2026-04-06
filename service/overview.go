package service

import (
	"context"
	"math"
	"sort"

	"github.com/nick0323/K8sVision/model"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// OverviewService 概览服务
type OverviewService struct {
	clientset *kubernetes.Clientset
}

// NewOverviewService 创建概览服务实例
func NewOverviewService(clientset *kubernetes.Clientset) *OverviewService {
	return &OverviewService{
		clientset: clientset,
	}
}

// GetOverview 获取集群概览信息
func (s *OverviewService) GetOverview(ctx context.Context) (*model.OverviewStatus, error) {
	overview := &model.OverviewStatus{}

	// 并行获取所有资源数据（包括 Events）
	data := s.fetchAllResources(ctx)

	// 处理 Pod 数据 - 直接使用列表接口的状态数据
	if data.podsErr == nil {
		overview.PodCount = len(data.pods)
		podNotReady := 0
		for _, p := range data.pods {
			// Pod 的 Status 字段是 K8s Phase（Running, Pending, Failed, Succeeded, Unknown）
			// Running 和 Succeeded 表示 Pod 就绪/正常
			if p.Status != "Running" && p.Status != "Succeeded" {
				podNotReady++
			}
		}
		overview.PodNotReady = podNotReady
	} else {
		// Pod 获取失败，设置错误标记
		overview.PodCount = -1
	}

	// 处理 Node 数据 - 直接使用列表接口的状态数据
	if data.nodesErr == nil {
		overview.NodeCount = len(data.nodes)
		nodeReady := 0
		for _, n := range data.nodes {
			// Node 的 Status 字段是 K8s 状态（Ready, NotReady, Unknown）
			// Ready 表示节点就绪
			if n.Status == model.NodeReady {
				nodeReady++
			}
		}
		overview.NodeReady = nodeReady
	} else {
		// Node 获取失败，设置错误标记
		overview.NodeCount = -1
	}

	// 处理 Namespace 数据
	if data.nsErr == nil {
		overview.NamespaceCount = len(data.nsList)
	}

	// 处理 Service 数据
	if data.servicesErr == nil {
		overview.ServiceCount = len(data.services)
	}

	// 计算资源使用率 - 仅在需要时使用原始数据
	if data.nodesRaw != nil && data.podsRaw != nil {
		s.calcResourceUsage(data.nodesRaw, data.podsRaw, overview)
	}

	// 处理 Events 数据
	if data.eventsErr == nil {
		overview.Events = data.events
	}

	return overview, nil
}

// overviewData 并行获取的资源数据
type overviewData struct {
	pods    []model.Pod
	podsRaw *corev1.PodList
	podsErr error

	nodes    []model.Node
	nodesRaw *corev1.NodeList
	nodesErr error

	nsList []model.Namespace
	nsErr  error

	services    []model.Service
	servicesErr error

	events    []model.Event
	eventsErr error
}

// fetchAllResources 并行获取所有资源
func (s *OverviewService) fetchAllResources(ctx context.Context) *overviewData {
	data := &overviewData{}

	// 定义 channel 类型
	type podResult struct {
		pods    []model.Pod     // 简化列表数据（用于统计）
		podsRaw *corev1.PodList // 原始数据（用于资源计算）
		err     error
	}
	type nodeResult struct {
		nodes    []model.Node     // 简化列表数据（用于统计）
		nodesRaw *corev1.NodeList // 原始数据（用于资源计算）
		err      error
	}
	type nsResult struct {
		nsList []model.Namespace
		err    error
	}
	type svcResult struct {
		services []model.Service
		err      error
	}
	type eventResult struct {
		events []model.Event
		err    error
	}

	// 创建带缓冲的 channel（避免 goroutine 阻塞）
	podChan := make(chan podResult, 1)
	nodeChan := make(chan nodeResult, 1)
	nsChan := make(chan nsResult, 1)
	svcChan := make(chan svcResult, 1)
	eventChan := make(chan eventResult, 1)

	// 并行获取 Pods（同时返回简化数据和原始数据）
	// 传入 true 表示不限制数量，获取完整的 Pod 列表用于统计
	go func() {
		pods, podsRaw, err := ListPodsWithRaw(ctx, s.clientset, nil, "", true)
		podChan <- podResult{pods: pods, podsRaw: podsRaw, err: err}
	}()

	// 并行获取 Nodes（同时返回简化数据和原始数据）
	go func() {
		nodesRaw, err := s.clientset.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
		var nodes []model.Node
		if err == nil && nodesRaw != nil {
			// 使用 MapNodes 转换简化数据
			nodes = MapNodes(nodesRaw.Items, &corev1.PodList{}, nil)
		}
		nodeChan <- nodeResult{nodes: nodes, nodesRaw: nodesRaw, err: err}
	}()

	// 并行获取 Namespaces
	go func() {
		nsList, err := ListNamespaces(ctx, s.clientset)
		nsChan <- nsResult{nsList: nsList, err: err}
	}()

	// 并行获取 Services
	go func() {
		services, err := ListServices(ctx, s.clientset, "")
		svcChan <- svcResult{services: services, err: err}
	}()

	// 并行获取 Events（只获取最近事件，减少数据传输）
	go func() {
		events, err := s.getRecentEvents(ctx, model.DefaultOverviewEventsLimit)
		eventChan <- eventResult{events: events, err: err}
	}()

	// 收集所有结果
	podRes := <-podChan
	data.pods = podRes.pods
	data.podsRaw = podRes.podsRaw
	data.podsErr = podRes.err

	nodeRes := <-nodeChan
	data.nodes = nodeRes.nodes
	data.nodesRaw = nodeRes.nodesRaw
	data.nodesErr = nodeRes.err

	nsRes := <-nsChan
	data.nsList = nsRes.nsList
	data.nsErr = nsRes.err

	svcRes := <-svcChan
	data.services = svcRes.services
	data.servicesErr = svcRes.err

	eventRes := <-eventChan
	data.events = eventRes.events
	data.eventsErr = eventRes.err

	return data
}

// calcResourceUsage 计算资源使用率
func (s *OverviewService) calcResourceUsage(nodesRaw *corev1.NodeList, podsRaw *corev1.PodList, overview *model.OverviewStatus) {
	var cpuCap, memCap, cpuReq, cpuLim, memReq, memLim float64

	// 计算节点容量
	for _, node := range nodesRaw.Items {
		if c, ok := node.Status.Capacity[corev1.ResourceCPU]; ok {
			cpuCap += float64(c.MilliValue()) / 1000.0
		}
		if m, ok := node.Status.Capacity[corev1.ResourceMemory]; ok {
			memCap += float64(m.Value()) / BytesPerGiB
		}
	}

	// 计算 Pod 资源请求和限制
	for _, pod := range podsRaw.Items {
		for _, container := range pod.Spec.Containers {
			resources := container.Resources

			// Requests
			if v, ok := resources.Requests[corev1.ResourceCPU]; ok {
				cpuReq += float64(v.MilliValue()) / 1000.0
			}
			if v, ok := resources.Requests[corev1.ResourceMemory]; ok {
				memReq += float64(v.Value()) / BytesPerGiB
			}

			// Limits
			if v, ok := resources.Limits[corev1.ResourceCPU]; ok {
				cpuLim += float64(v.MilliValue()) / 1000.0
			}
			if v, ok := resources.Limits[corev1.ResourceMemory]; ok {
				memLim += float64(v.Value()) / BytesPerGiB
			}
		}
	}

	// 保留一位小数
	overview.CPUCapacity = round(cpuCap)
	overview.MemoryCapacity = round(memCap)
	overview.CPURequests = round(cpuReq)
	overview.CPULimits = round(cpuLim)
	overview.MemoryRequests = round(memReq)
	overview.MemoryLimits = round(memLim)
}

// round 四舍五入保留一位小数
func round(val float64) float64 {
	return math.Round(val*10) / 10
}

// getRecentEvents 获取最近事件
func (s *OverviewService) getRecentEvents(ctx context.Context, limit int) ([]model.Event, error) {
	eventList, err := s.clientset.CoreV1().Events("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}

	if eventList == nil || len(eventList.Items) == 0 {
		return []model.Event{}, nil
	}

	// 按 LastSeen 时间倒序排序
	events := eventList.Items
	sort.Slice(events, func(i, j int) bool {
		return events[i].LastTimestamp.Time.After(events[j].LastTimestamp.Time)
	})

	// 取最近的 N 条事件
	count := min(limit, len(events))
	recentEvents := make([]model.Event, count)
	for i := 0; i < count; i++ {
		e := events[i]
		recentEvents[i] = model.Event{
			Name:      e.Name,
			Namespace: e.Namespace,
			Reason:    e.Reason,
			Message:   e.Message,
			Type:      e.Type,
			Count:     e.Count,
			LastSeen:  model.FormatTime(&e.LastTimestamp),
		}
	}

	return recentEvents, nil
}

// min 返回两个整数中的较小值
func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
