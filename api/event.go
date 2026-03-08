package api

import (
	"context"
	"net/http"
	"sort"
	"time"

	"github.com/nick0323/K8sVision/api/middleware"
	"github.com/nick0323/K8sVision/model"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
)

// RegisterEvent 注册 Event 相关路由
func RegisterEvent(
	r *gin.RouterGroup,
	logger *zap.Logger,
	getK8sClient K8sClientProvider, // 使用类型别名简化签名
	listEvents func(context.Context, *kubernetes.Clientset, string) ([]model.EventStatus, error),
) {
	r.GET("/events", getEventList(logger, getK8sClient, listEvents))
	r.GET("/events/:namespace/:name", getEventDetail(logger, getK8sClient))
}

// getEventList 获取Event列表的处理函数
func getEventList(
	logger *zap.Logger,
	getK8sClient K8sClientProvider,
	listEvents func(context.Context, *kubernetes.Clientset, string) ([]model.EventStatus, error),
) gin.HandlerFunc {
	return func(c *gin.Context) {
		HandleListWithPagination(c, logger, func(ctx context.Context, params PaginationParams) ([]model.EventStatus, error) {
			clientset, _, err := getK8sClient()
			if err != nil {
				return nil, err
			}
			events, err := listEvents(ctx, clientset, params.Namespace)
			if err != nil {
				return nil, err
			}
			// 对原始数据进行排序（在搜索和分页之前）
			sortEventsByLastSeen(events)
			return events, nil
		}, ListSuccessMessage)
	}
}

// getEventDetail 获取Event详情的处理函数
func getEventDetail(
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
		event, err := clientset.CoreV1().Events(ns).Get(ctx, name, metav1.GetOptions{})
		if err != nil {
			middleware.ResponseError(c, logger, err, http.StatusNotFound)
			return
		}

		eventDetail := model.EventDetail{
			Namespace: event.Namespace,
			Name:      event.Name,
			Status:    event.Type,
			Labels:    event.Labels,
			Reason:    event.Reason,
			Message:   event.Message,
			Type:      event.Type,
			Count:     event.Count,
			FirstSeen: event.FirstTimestamp.Format("2006-01-02 15:04:05"),
			LastSeen:  event.LastTimestamp.Format("2006-01-02 15:04:05"),
			Duration:  event.LastTimestamp.Sub(event.FirstTimestamp.Time).String(),
		}
		middleware.ResponseSuccess(c, eventDetail, DetailSuccessMessage, nil)
	}
}

// sortEventsByLastSeen 按LastSeen时间倒序排列Events（最新事件在前）
func sortEventsByLastSeen(events []model.EventStatus) {
	sort.Slice(events, func(i, j int) bool {
		// 解析时间字符串进行比较
		timeI, errI := time.Parse("2006-01-02 15:04:05", events[i].LastSeen)
		timeJ, errJ := time.Parse("2006-01-02 15:04:05", events[j].LastSeen)

		// 如果解析失败，保持原始顺序
		if errI != nil || errJ != nil {
			return false
		}

		// 倒序排列：最新的时间在前
		return timeI.After(timeJ)
	})
}
