package api

import (
	"context"
	"sort"
	"time"

	"github.com/nick0323/K8sVision/model"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
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
}

// getEventList 获取 Event 列表的处理函数
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

// sortEventsByLastSeen 按 LastSeen 时间倒序排列 Events（最新事件在前）
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
