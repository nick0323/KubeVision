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

type OverviewService struct {
	clientset *kubernetes.Clientset
}

func NewOverviewService(clientset *kubernetes.Clientset) *OverviewService {
	return &OverviewService{clientset: clientset}
}

func (s *OverviewService) GetOverview(ctx context.Context) (*model.OverviewStatus, error) {
	overview := &model.OverviewStatus{}
	data := s.fetchAllResources(ctx)

	if data.podsErr == nil {
		overview.PodCount = len(data.pods)
		for _, p := range data.pods {
			if p.Status != "Running" && p.Status != "Succeeded" {
				overview.PodNotReady++
			}
		}
	} else {
		overview.PodCount = -1
	}

	if data.nodesErr == nil {
		overview.NodeCount = len(data.nodes)
		for _, n := range data.nodes {
			if n.Status == model.NodeReady {
				overview.NodeReadyCount++
			}
		}
	} else {
		overview.NodeCount = -1
	}

	if data.nsErr == nil {
		overview.NamespaceCount = len(data.nsList)
	}
	if data.servicesErr == nil {
		overview.ServiceCount = len(data.services)
	}

	if data.nodesRaw != nil && data.podsRaw != nil {
		s.calcResourceUsage(data.nodesRaw, data.podsRaw, overview)
	}
	if data.eventsErr == nil {
		overview.Events = data.events
	}

	return overview, nil
}

type overviewData struct {
	pods        []model.Pod
	podsRaw     *corev1.PodList
	podsErr     error
	nodes       []model.Node
	nodesRaw    *corev1.NodeList
	nodesErr    error
	nsList      []model.Namespace
	nsErr       error
	services    []model.Service
	servicesErr error
	events      []model.Event
	eventsErr   error
}

func (s *OverviewService) fetchAllResources(ctx context.Context) *overviewData {
	data := &overviewData{}

	type podResult struct {
		pods    []model.Pod
		podsRaw *corev1.PodList
		err     error
	}
	type nodeResult struct {
		nodes    []model.Node
		nodesRaw *corev1.NodeList
		err      error
	}

	podChan := make(chan podResult, 1)
	nodeChan := make(chan nodeResult, 1)
	nsChan := make(chan []model.Namespace, 1)
	svcChan := make(chan []model.Service, 1)
	eventChan := make(chan []model.Event, 1)

	go func() {
		pods, podsRaw, err := ListPodsWithRaw(ctx, s.clientset, "", "", "", true)
		podChan <- podResult{pods: pods, podsRaw: podsRaw, err: err}
	}()

	go func() {
		nodesRaw, err := s.clientset.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
		var nodes []model.Node
		if err == nil && nodesRaw != nil {
			nodes = MapNodes(nodesRaw.Items, &corev1.PodList{})
		}
		nodeChan <- nodeResult{nodes: nodes, nodesRaw: nodesRaw, err: err}
	}()

	go func() { nsChan <- must(ListNamespaces(ctx, s.clientset, "", "")) }()
	go func() { svcChan <- must(ListServices(ctx, s.clientset, "", "", "")) }()
	go func() {
		events, err := s.getRecentEvents(ctx, model.DefaultOverviewEventsLimit)
		eventChan <- events
		_ = err
	}()

	podRes := <-podChan
	data.pods, data.podsRaw, data.podsErr = podRes.pods, podRes.podsRaw, podRes.err

	nodeRes := <-nodeChan
	data.nodes, data.nodesRaw, data.nodesErr = nodeRes.nodes, nodeRes.nodesRaw, nodeRes.err

	data.nsList, data.nsErr = <-nsChan, nil
	data.services, data.servicesErr = <-svcChan, nil
	data.events, data.eventsErr = <-eventChan, nil

	return data
}

func must[T any](res T, _ error) T { return res }

func (s *OverviewService) calcResourceUsage(nodesRaw *corev1.NodeList, podsRaw *corev1.PodList, overview *model.OverviewStatus) {
	var cpuCap, memCap, cpuReq, cpuLim, memReq, memLim float64

	for _, node := range nodesRaw.Items {
		if c, ok := node.Status.Capacity[corev1.ResourceCPU]; ok {
			cpuCap += float64(c.MilliValue()) / 1000.0
		}
		if m, ok := node.Status.Capacity[corev1.ResourceMemory]; ok {
			memCap += float64(m.Value()) / BytesPerGiB
		}
	}

	for _, pod := range podsRaw.Items {
		for _, container := range pod.Spec.Containers {
			resources := container.Resources
			if v, ok := resources.Requests[corev1.ResourceCPU]; ok {
				cpuReq += float64(v.MilliValue()) / 1000.0
			}
			if v, ok := resources.Requests[corev1.ResourceMemory]; ok {
				memReq += float64(v.Value()) / BytesPerGiB
			}
			if v, ok := resources.Limits[corev1.ResourceCPU]; ok {
				cpuLim += float64(v.MilliValue()) / 1000.0
			}
			if v, ok := resources.Limits[corev1.ResourceMemory]; ok {
				memLim += float64(v.Value()) / BytesPerGiB
			}
		}
	}

	overview.CPUCapacity = round(cpuCap)
	overview.MemoryCapacity = round(memCap)
	overview.CPURequests = round(cpuReq)
	overview.CPULimits = round(cpuLim)
	overview.MemoryRequests = round(memReq)
	overview.MemoryLimits = round(memLim)
}

func round(val float64) float64 { return math.Round(val*10) / 10 }

func (s *OverviewService) getRecentEvents(ctx context.Context, limit int) ([]model.Event, error) {
	eventList, err := s.clientset.CoreV1().Events("").List(ctx, metav1.ListOptions{})
	if err != nil {
		return nil, err
	}
	if eventList == nil || len(eventList.Items) == 0 {
		return []model.Event{}, nil
	}

	events := eventList.Items
	sort.Slice(events, func(i, j int) bool {
		return events[i].LastTimestamp.Time.After(events[j].LastTimestamp.Time)
	})

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

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
