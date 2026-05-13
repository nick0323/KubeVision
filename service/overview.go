package service

import (
	"context"
	"math"
	"sort"
	"sync"
	"time"

	"github.com/nick0323/K8sVision/model"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	versioned "k8s.io/metrics/pkg/client/clientset/versioned"
)

const BytesPerGiB = 1024 * 1024 * 1024

type OverviewService struct {
	clientset     kubernetes.Interface
	metricsClient interface{}
}

func NewOverviewService(clientset kubernetes.Interface, metricsClient interface{}) *OverviewService {
	return &OverviewService{clientset: clientset, metricsClient: metricsClient}
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
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		pods, podsRaw, err := ListPodsWithRaw(ctx, s.clientset, "", "", "")
		data.pods = pods
		data.podsRaw = podsRaw
		data.podsErr = err
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		nodesRaw, err := s.clientset.CoreV1().Nodes().List(ctx, metav1.ListOptions{})
		if err != nil {
			data.nodesErr = err
			return
		}
		metricsMap := make(map[string]*model.NodeMetrics)
		if s.metricsClient != nil {
			if mc, ok := s.metricsClient.(versioned.Interface); ok {
				metricsList, err := mc.MetricsV1beta1().NodeMetricses().List(ctx, metav1.ListOptions{})
				if err == nil {
					for i := range metricsList.Items {
						metricsMap[metricsList.Items[i].Name] = &model.NodeMetrics{
							CPU:    metricsList.Items[i].Usage.Cpu().String(),
							Memory: metricsList.Items[i].Usage.Memory().String(),
						}
					}
				}
			}
		}
		data.nodes = MapNodes(nodesRaw.Items, &corev1.PodList{}, metricsMap)
		data.nodesRaw = nodesRaw
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		nsList, err := ListNamespaces(ctx, s.clientset, "", "")
		data.nsList = nsList
		data.nsErr = err
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		services, err := ListServices(ctx, s.clientset, "", "", "")
		data.services = services
		data.servicesErr = err
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()
		events, err := s.getRecentEvents(ctx, model.DefaultOverviewEventsLimit)
		data.events = events
		data.eventsErr = err
	}()

	wg.Wait()
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
	eventList, err := s.clientset.CoreV1().Events("").List(ctx, metav1.ListOptions{
		Limit: 500,
	})
	if err != nil {
		return nil, err
	}
	if len(eventList.Items) == 0 {
		return []model.Event{}, nil
	}

	since := time.Now().Add(-1 * time.Hour)
	filtered := make([]corev1.Event, 0, len(eventList.Items))
	for _, e := range eventList.Items {
		eventTime := e.LastTimestamp.Time
		if eventTime.IsZero() {
			eventTime = e.EventTime.Time
		}
		if !eventTime.IsZero() && eventTime.After(since) {
			filtered = append(filtered, e)
		}
	}

	sort.Slice(filtered, func(i, j int) bool {
		iTime := filtered[i].LastTimestamp.Time
		if iTime.IsZero() {
			iTime = filtered[i].EventTime.Time
		}
		jTime := filtered[j].LastTimestamp.Time
		if jTime.IsZero() {
			jTime = filtered[j].EventTime.Time
		}
		return iTime.After(jTime)
	})

	count := min(limit, len(filtered))
	recentEvents := make([]model.Event, count)
	for i := 0; i < count; i++ {
		e := filtered[i]
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
