package service

import (
	"context"
	"math"
	"sort"

	"github.com/nick0323/K8sVision/model"
	"golang.org/x/sync/errgroup"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

const BytesPerGiB = 1024 * 1024 * 1024

type OverviewService struct {
	clientset kubernetes.Interface
}

func NewOverviewService(clientset kubernetes.Interface) *OverviewService {
	return &OverviewService{clientset: clientset}
}

func (s *OverviewService) GetOverview(ctx context.Context) (*model.OverviewStatus, error) {
	overview := &model.OverviewStatus{}
	data, err := s.fetchAllResources(ctx)
	if err != nil {
		return nil, err
	}

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

func (s *OverviewService) fetchAllResources(ctx context.Context) (*overviewData, error) {
	data := &overviewData{}
	g, ctx := errgroup.WithContext(ctx)

	// 1. 并发查询 Pod
	g.Go(func() error {
		pods, podsRaw, err := ListPodsWithRaw(ctx, s.clientset, "", "", "", true)
		data.pods = pods
		data.podsRaw = podsRaw
		data.podsErr = err
		return err
	})

	// 2. 并发查询 Node
	g.Go(func() error {
		nodesRaw, err := s.clientset.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
		if err != nil {
			data.nodesErr = err
			return err
		}
		data.nodes = MapNodes(nodesRaw.Items, &corev1.PodList{})
		data.nodesRaw = nodesRaw
		return nil
	})

	// 3. 并发查询 Namespace
	g.Go(func() error {
		nsList, err := ListNamespaces(ctx, s.clientset, "", "")
		data.nsList = nsList
		data.nsErr = err
		return err
	})

	// 4. 并发查询 Service
	g.Go(func() error {
		services, err := ListServices(ctx, s.clientset, "", "", "")
		data.services = services
		data.servicesErr = err
		return err
	})

	// 5. 并发查询 Events
	g.Go(func() error {
		events, err := s.getRecentEvents(ctx, model.DefaultOverviewEventsLimit)
		data.events = events
		data.eventsErr = err
		return err
	})

	// 等待所有任务完成
	if err := g.Wait(); err != nil {
		return nil, err
	}

	return data, nil
}

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
	if len(eventList.Items) == 0 {
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
