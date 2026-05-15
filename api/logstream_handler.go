package api

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/nick0323/K8sVision/api/middleware"
	"go.uber.org/zap"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/kubernetes"
)

// RegisterLogStream 注册日志流接口（WebSocket）
func RegisterLogStream(
	r *gin.RouterGroup,
	logger *zap.Logger,
	getK8sClient K8sClientProvider,
	configProvider middleware.ConfigProvider,
	wsMgr *WebSocketManager,
) {
	r.GET("/ws/stream", streamPodLog(logger, getK8sClient, configProvider, wsMgr))
}

// streamPodLog 流式获取 Pod 日志（WebSocket 实时推送）
func streamPodLog(
	logger *zap.Logger,
	getK8sClient K8sClientProvider,
	configProvider middleware.ConfigProvider,
	wsMgr *WebSocketManager,
) gin.HandlerFunc {
	return func(c *gin.Context) {
		cluster := c.Query("cluster")
		clientset, _, err := getK8sClient(cluster)
		if err != nil {
			middleware.ResponseError(c, logger, err, http.StatusInternalServerError)
			return
		}

		ctx := GetRequestContext(c)
		ctx = wsMgr.ShutdownCtx(ctx)
		namespace := c.Query("namespace")
		podName := c.Query("pod")
		container := c.Query("container")
		tailLines := c.Query("tailLines")
		timestamps := c.Query("timestamps")
		previous := c.Query("previous")

		// 输入验证
		if err := validatePodLogParams(namespace, podName, container); err != nil {
			middleware.ResponseError(c, logger, err, http.StatusBadRequest)
			return
		}

		if err := validateWebSocketToken(c, logger, configProvider); err != nil {
			return
		}

		// 检查连接数限制
		if err := checkConnectionLimit(wsMgr, logger); err != nil {
			c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{
				"type":    "error",
				"message": err.Error(),
			})
			return
		}

		// 构建日志选项
		opts := buildPodLogOptions(container, timestamps, previous, tailLines)

		// WebSocket 升级
		ws, err := upgradeWebSocket(c, logger)
		if err != nil {
			decrementConnection(wsMgr)
			return
		}
		defer ws.Close()

		// 设置 WebSocket 关闭处理器（负责递减连接计数）
		wsCloseOnce := setupWebSocketCloseHandler(ws, logger, podName, wsMgr)

		logger.Info("Log WebSocket connected",
			zap.String("namespace", namespace),
			zap.String("pod", podName),
			zap.String("container", container),
			zap.Int32("activeConnections", wsMgr.ActiveCount()),
		)

		// 获取日志流
		podLogs, err := getPodLogStream(ctx, clientset, namespace, podName, opts, logger)
		if err != nil {
			logger.Warn("Failed to get log stream", zap.Error(err))
			if writeErr := ws.WriteJSON(gin.H{"type": "error", "message": fmt.Sprintf("Failed to get log stream: %v", err)}); writeErr != nil {
				logger.Error("Failed to send error message", zap.Error(writeErr))
			}
			return
		}
		defer podLogs.Close()

		// 发送连接成功消息
		if writeErr := ws.WriteJSON(gin.H{
			"type":    "connected",
			"message": fmt.Sprintf("Connected to %s/%s (%s)", namespace, podName, container),
		}); writeErr != nil {
			logger.Error("Failed to send connected message", zap.Error(writeErr))
			return
		}

		// 启动日志读取 goroutine
		// 优化：增加缓冲区大小防止阻塞，从 200 增加到 500
		logChan := make(chan string, 500)
		errorChan := make(chan error, 1)
		doneChan := make(chan struct{})

		go readPodLogs(ctx, podLogs, logChan, errorChan, doneChan, podName, logger)

		// 主循环：处理日志、心跳、超时
		runWebSocketLoop(ctx, ws, logChan, errorChan, doneChan, podName, logger, wsCloseOnce, wsMgr)
	}
}

// checkConnectionLimit 检查连接数限制
func checkConnectionLimit(wsMgr *WebSocketManager, logger *zap.Logger) error {
	err := wsMgr.Acquire()
	if err != nil {
		logger.Warn("Too many WebSocket connections", zap.Int32("current", wsMgr.ActiveCount()))
		return fmt.Errorf("connection limit exceeded")
	}
	return nil
}

func decrementConnection(wsMgr *WebSocketManager) {
	wsMgr.Release()
}

// buildPodLogOptions 构建 Pod 日志选项
func buildPodLogOptions(container, timestamps, previous, tailLines string) *corev1.PodLogOptions {
	opts := &corev1.PodLogOptions{
		Follow:    true,
		TailLines: nil,
	}
	if container != "" {
		opts.Container = container
	}
	if timestamps == "true" {
		opts.Timestamps = true
	}
	if previous == "true" {
		opts.Previous = true
	}
	if tailLines != "" && tailLines != "0" {
		var lines int64
		if _, err := fmt.Sscanf(tailLines, "%d", &lines); err == nil && lines > 0 {
			opts.TailLines = &lines
		}
	}
	return opts
}

// upgradeWebSocket 升级 WebSocket 连接
func upgradeWebSocket(c *gin.Context, logger *zap.Logger) (*websocket.Conn, error) {
	ws, err := upgrader.Upgrade(c.Writer, c.Request, buildWebSocketUpgradeHeaders(c))
	if err != nil {
		logger.Error("WebSocket upgrade failed", zap.Error(err))
		c.Abort()
		return nil, err
	}
	return ws, nil
}

// setupWebSocketCloseHandler 设置 WebSocket 关闭处理器
func setupWebSocketCloseHandler(ws *websocket.Conn, logger *zap.Logger, podName string, wsMgr *WebSocketManager) *sync.Once {
	var wsCloseOnce sync.Once
	ws.SetCloseHandler(func(code int, text string) error {
		logger.Info("Client disconnected WebSocket",
			zap.Int("code", code),
			zap.String("text", text),
			zap.String("pod", podName),
		)
		wsCloseOnce.Do(func() {
			decrementConnection(wsMgr)
		})
		return nil
	})
	return &wsCloseOnce
}

// getPodLogStream 获取 Pod 日志流
func getPodLogStream(ctx context.Context, clientset kubernetes.Interface, namespace, podName string, opts *corev1.PodLogOptions, logger *zap.Logger) (io.ReadCloser, error) {
	req := clientset.CoreV1().Pods(namespace).GetLogs(podName, opts)
	podLogs, err := req.Stream(ctx)
	if err != nil {
		return nil, err
	}

	logger.Info("Log stream opened",
		zap.String("namespace", namespace),
		zap.String("pod", podName),
	)

	return podLogs, nil
}

// readPodLogs 读取 Pod 日志
func readPodLogs(ctx context.Context, podLogs io.ReadCloser, logChan chan<- string, errorChan chan<- error, doneChan chan struct{}, podName string, logger *zap.Logger) {
	defer close(doneChan)

	reader := bufio.NewReader(podLogs)
	logCount := int64(0)
	lastReadTime := time.Now()
	backoff := 100 * time.Millisecond
	const maxBackoff = 5 * time.Second

	for {
		select {
		case <-ctx.Done():
			logger.Debug("Log reader goroutine received cancel signal", zap.String("pod", podName))
			return
		default:
			line, err := reader.ReadString('\n')
			readDuration := time.Since(lastReadTime)

			if len(line) > 0 {
				backoff = 100 * time.Millisecond
				count := atomic.AddInt64(&logCount, 1)
				logger.Debug("Read log line",
					zap.String("pod", podName),
					zap.Int64("lineNumber", count),
					zap.Int("length", len(line)),
					zap.Duration("sinceLastRead", readDuration),
				)
				select {
				case logChan <- line:
				default:
					logger.Warn("logChan full, discarding log",
						zap.String("pod", podName),
						zap.Int64("lineNumber", count),
					)
				}
				lastReadTime = time.Now()
			}

			if err != nil {
				if err == io.EOF {
					logger.Debug("Log stream EOF, waiting for new data", zap.String("pod", podName))
					time.Sleep(backoff)
					backoff *= 2
					if backoff > maxBackoff {
						backoff = maxBackoff
					}
					continue
				} else if isTimeoutError(err) {
					logger.Debug("Connection timeout, retrying", zap.String("pod", podName), zap.Error(err))
					time.Sleep(backoff)
					backoff *= 2
					if backoff > maxBackoff {
						backoff = maxBackoff
					}
					continue
				} else {
					logger.Warn("Error reading log stream", zap.String("pod", podName), zap.Error(err))
					select {
					case errorChan <- err:
					default:
					}
					return
				}
			}
		}
	}
}

// runWebSocketLoop 运行 WebSocket 主循环
func runWebSocketLoop(ctx context.Context, ws *websocket.Conn, logChan <-chan string, errorChan <-chan error, doneChan <-chan struct{}, podName string, logger *zap.Logger, wsCloseOnce *sync.Once, wsMgr *WebSocketManager) {
	heartbeatTicker := time.NewTicker(30 * time.Second)
	defer heartbeatTicker.Stop()

	maxDuration := 10 * time.Minute
	timeoutTimer := time.NewTimer(maxDuration)
	defer timeoutTimer.Stop()

	for {
		select {
		case <-ctx.Done():
			logger.Info("Context cancelled, closing log stream", zap.String("pod", podName))
			wsCloseOnce.Do(func() {
				decrementConnection(wsMgr)
			})
			return
		case <-timeoutTimer.C:
			logger.Info("Connection timeout, closing log stream", zap.String("pod", podName), zap.Duration("maxDuration", maxDuration))
			if writeErr := ws.WriteJSON(gin.H{"type": "info", "message": "Connection timeout, please reconnect"}); writeErr != nil {
				logger.Error("Failed to send timeout message", zap.Error(writeErr))
			}
			wsCloseOnce.Do(func() {
				decrementConnection(wsMgr)
			})
			return
		case <-heartbeatTicker.C:
			if err := ws.WriteJSON(gin.H{"type": "heartbeat"}); err != nil {
				// 客户端可能已断开连接
				if websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					logger.Info("WebSocket client disconnected (heartbeat check)",
						zap.String("pod", podName),
						zap.Error(err),
					)
				} else {
					logger.Debug("Failed to send heartbeat", zap.String("pod", podName), zap.Error(err))
				}
				wsCloseOnce.Do(func() {
					decrementConnection(wsMgr)
				})
				return
			}
			logger.Debug("Heartbeat sent", zap.String("pod", podName))
		case logLine := <-logChan:
			if err := ws.WriteJSON(gin.H{
				"type":    "log",
				"content": logLine,
			}); err != nil {
				// 客户端可能已断开连接
				if websocket.IsCloseError(err, websocket.CloseNormalClosure, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
					logger.Info("WebSocket client disconnected",
						zap.String("pod", podName),
						zap.Error(err),
					)
				} else {
					logger.Error("Failed to send log to WebSocket", zap.String("pod", podName), zap.Error(err))
				}
				wsCloseOnce.Do(func() {
					decrementConnection(wsMgr)
				})
				return
			}
		case err := <-errorChan:
			logger.Error("Log read error", zap.String("pod", podName), zap.Error(err))
			if writeErr := ws.WriteJSON(gin.H{"type": "error", "message": fmt.Sprintf("Failed to read log: %v", err)}); writeErr != nil {
				logger.Error("Failed to send error message", zap.Error(writeErr))
			}
			wsCloseOnce.Do(func() {
				decrementConnection(wsMgr)
			})
			return
		case <-doneChan:
			logger.Info("Log reader goroutine exited", zap.String("pod", podName))
			wsCloseOnce.Do(func() {
				decrementConnection(wsMgr)
			})
			return
		}
	}
}

// isTimeoutError 判断是否为超时错误
func isTimeoutError(err error) bool {
	if err == nil {
		return false
	}
	errStr := err.Error()
	return strings.Contains(errStr, "Client.Timeout") ||
		strings.Contains(errStr, "context cancellation") ||
		strings.Contains(errStr, "timeout") ||
		strings.Contains(errStr, "i/o timeout")
}
