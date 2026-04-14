package api

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"sync"
	"sync/atomic"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/nick0323/K8sVision/api/middleware"
	"github.com/nick0323/K8sVision/service"
	"go.uber.org/zap"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/tools/remotecommand"
)

// Exec 配置常量
const (
	ExecDefaultCommand = "/bin/sh"
	WebSocketTimeout   = 5 * time.Second
	ExecSessionTimeout = 30 * time.Minute // 会话超时
	MaxExecConnections = 100              // 最大并发连接数
)

// 全局连接计数
var activeExecConnections atomic.Int32

// 全局 ClientManager（由 main.go 初始化时设置）
var globalClientManager *service.ClientManager

// SetGlobalClientManager 设置全局 ClientManager（在 main.go 初始化时调用）
func SetGlobalClientManager(cm *service.ClientManager) {
	globalClientManager = cm
}

// dnsLabelRegex 验证 namespace 名称 (DNS label 规范)
var dnsLabelRegex = regexp.MustCompile(`^[a-z0-9]([-a-z0-9]*[a-z0-9])?$`)

// wsInput 实现 io.Reader 接口，从 WebSocket 读取输入
type wsInput struct {
	conn     *websocket.Conn
	sizeChan chan *remotecommand.TerminalSize // resize 消息通道
	mu       sync.Mutex
	buf      []byte
}

// termSizeQueue 实现 TerminalSizeQueue 接口，支持 resize 消息
type termSizeQueue struct {
	sizeChan <-chan *remotecommand.TerminalSize
	size     *remotecommand.TerminalSize
}

// Next 阻塞直到收到新的尺寸或错误
func (t *termSizeQueue) Next() *remotecommand.TerminalSize {
	// 初始尺寸
	if t.size != nil {
		size := t.size
		t.size = nil
		return size
	}

	// 阻塞等待 resize 消息
	size, ok := <-t.sizeChan
	if !ok {
		return nil
	}
	return size
}

func (w *wsInput) Read(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	// 如果缓冲区有数据，先返回
	if len(w.buf) > 0 {
		n := copy(p, w.buf)
		w.buf = w.buf[n:]
		return n, nil
	}

	// 读取新消息
	for {
		_, message, err := w.conn.ReadMessage()
		if err != nil {
			return 0, err
		}

		// 检查是否是 resize 消息
		var msg struct {
			Type string `json:"type"`
			Cols uint16 `json:"cols"`
			Rows uint16 `json:"rows"`
		}
		if err := json.Unmarshal(message, &msg); err == nil && msg.Type == "resize" {
			// 发送 resize 到通道
			w.sizeChan <- &remotecommand.TerminalSize{
				Width:  msg.Cols,
				Height: msg.Rows,
			}
			continue // 不返回给 stdin，继续读取下一条
		}

		// 普通输入数据
		if len(message) > len(p) {
			w.buf = append([]byte(nil), message[len(p):]...)
		}

		n := copy(p, message)
		return n, nil
	}
}

// wsOutput 实现 io.Writer 接口，向 WebSocket 写入输出
type wsOutput struct {
	conn *websocket.Conn
	mu   sync.Mutex
}

func (w *wsOutput) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	// 使用 BinaryMessage 支持所有字符
	if err := w.conn.WriteMessage(websocket.BinaryMessage, p); err != nil {
		return 0, err
	}
	return len(p), nil
}

func RegisterExecWS(
	r *gin.RouterGroup,
	logger *zap.Logger,
	getK8sClient K8sClientProvider,
	configProvider middleware.ConfigProvider,
) {
	// 注册 Exec WebSocket 路由
	r.GET("/ws/exec", HandleExecWS(logger, getK8sClient, configProvider))
}

// HandleExecWS 处理 WebSocket exec 请求（带 JWT 认证）
func HandleExecWS(
	logger *zap.Logger,
	getK8sClient K8sClientProvider,
	configProvider middleware.ConfigProvider,
) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从请求中提取 token
		tokenStr := ExtractTokenFromRequest(c)

		if tokenStr == "" {
			logger.Warn("WebSocket exec 缺少 token")
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		// 验证 token 并获取用户名
		claims, err := middleware.VerifyToken(tokenStr, configProvider.GetJWTSecret())
		if err != nil {
			logger.Warn("WebSocket exec token 验证失败", zap.Error(err))
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		username, _ := claims["username"].(string)

		// 调用实际的 WebSocket 处理函数
		// 注意：handleExecWSImpl 会处理 WebSocket 升级，不需要 c.Next()
		handleExecWSImpl(c, logger, getK8sClient, username)

		// 停止 Gin 继续执行中间件
		c.Abort()
	}
}

func handleExecWSImpl(
	c *gin.Context,
	logger *zap.Logger,
	getK8sClient K8sClientProvider,
	username string,
) {
	// 获取参数
	namespace := c.Query("namespace")
	podName := c.Query("pod")
	container := c.Query("container")
	commandStr := c.Query("command")

	// 验证 namespace 格式
	if !isValidNamespace(namespace) {
		middleware.ResponseError(c, logger, fmt.Errorf("invalid namespace format"), http.StatusBadRequest)
		return
	}

	if podName == "" {
		middleware.ResponseError(c, logger, fmt.Errorf("missing pod name"), http.StatusBadRequest)
		return
	}

	// 检查并发连接数
	if activeExecConnections.Load() >= MaxExecConnections {
		logger.Warn("exec 连接数已达上限",
			zap.Int32("active", activeExecConnections.Load()),
			zap.String("user", username),
		)
		middleware.ResponseError(c, logger, fmt.Errorf("服务繁忙，请稍后重试"), http.StatusServiceUnavailable)
		return
	}

	// 解析命令
	command := []string{ExecDefaultCommand}
	if commandStr != "" {
		command = []string{ExecDefaultCommand, "-c", commandStr}
	}

	// 获取 K8s 客户端和配置（在 WebSocket 升级前）
	// 使用全局 ClientManager 而不是重新创建客户端
	if globalClientManager == nil {
		logger.Error("ClientManager 未初始化")
		middleware.ResponseError(c, logger, fmt.Errorf("系统未初始化"), http.StatusInternalServerError)
		return
	}

	clientset, _, err := globalClientManager.GetDefaultClient()
	if err != nil {
		logger.Error("获取 K8s 客户端失败", zap.Error(err))
		middleware.ResponseError(c, logger, err, http.StatusInternalServerError)
		return
	}

	config := globalClientManager.GetDefaultRESTConfig()

	// 检查 Pod 是否存在并获取容器列表（在 WebSocket 升级前）
	pod, err := clientset.CoreV1().Pods(namespace).Get(context.Background(), podName, metav1.GetOptions{})
	if err != nil {
		logger.Error("Pod 不存在", zap.Error(err))
		middleware.ResponseError(c, logger, err, http.StatusNotFound)
		return
	}

	// 验证 container 是否属于该 pod
	if container != "" && !hasContainer(pod, container) {
		logger.Error("container 不存在于 pod 中",
			zap.String("pod", podName),
			zap.String("container", container),
		)
		middleware.ResponseError(c, logger, fmt.Errorf("container '%s' not found in pod '%s'", container, podName), http.StatusBadRequest)
		return
	}

	// 升级 WebSocket 连接
	ws, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		logger.Error("WebSocket 升级失败", zap.Error(err))
		return
	}
	defer ws.Close()

	// 增加连接计数
	activeExecConnections.Add(1)
	defer activeExecConnections.Add(-1)

	logger.Info("Terminal WebSocket 升级成功",
		zap.String("namespace", namespace),
		zap.String("pod", podName),
		zap.String("container", container),
		zap.String("user", username),
	)

	// 先发送连接成功消息
	if err := ws.WriteJSON(gin.H{
		"status":    "connected",
		"namespace": namespace,
		"pod":       podName,
		"container": container,
		"message":   fmt.Sprintf("Connected to %s/%s (%s)", namespace, podName, container),
	}); err != nil {
		logger.Error("发送连接消息失败", zap.Error(err))
		return
	}

	logger.Info("已发送连接消息")

	// 创建 exec 请求
	req := clientset.CoreV1().RESTClient().
		Post().
		Resource("pods").
		Name(podName).
		Namespace(namespace).
		SubResource("exec").
		VersionedParams(&v1.PodExecOptions{
			Container: container,
			Command:   command,
			Stdin:     true,
			Stdout:    true,
			Stderr:    true,
			TTY:       true,
		}, scheme.ParameterCodec)

	// 创建 executor
	exec, err := remotecommand.NewSPDYExecutor(config, "POST", req.URL())
	if err != nil {
		logger.Error("创建 executor 失败", zap.Error(err))
		ws.WriteJSON(gin.H{"type": "error", "message": fmt.Sprintf("Failed to create executor: %v", err)})
		return
	}

	logger.Info("Executor 创建成功，开始 stream")

	// 创建带超时的上下文
	ctx, cancel := context.WithTimeout(context.Background(), ExecSessionTimeout)
	defer cancel()

	// 创建 resize 通道
	sizeChan := make(chan *remotecommand.TerminalSize, 1)

	// 创建带锁的输入输出
	input := &wsInput{
		conn:     ws,
		sizeChan: sizeChan,
	}
	output := &wsOutput{conn: ws}

	// 创建 TerminalSizeQueue 支持 resize
	sizeQueue := &termSizeQueue{sizeChan: sizeChan}

	// 执行命令
	err = exec.StreamWithContext(ctx, remotecommand.StreamOptions{
		Stdin:             input,
		Stdout:            output,
		Stderr:            output,
		Tty:               true,
		TerminalSizeQueue: sizeQueue,
	})

	if err != nil {
		if err == io.EOF {
			logger.Info("exec stream 正常结束",
				zap.String("user", username),
				zap.String("namespace", namespace),
				zap.String("pod", podName),
			)
		} else {
			logger.Error("exec stream 失败", zap.Error(err))
			ws.WriteJSON(gin.H{"type": "error", "message": fmt.Sprintf("Exec failed: %v", err)})
		}
	}
}

// isValidNamespace 验证 namespace 名称是否符合 DNS label 规范
func isValidNamespace(namespace string) bool {
	if namespace == "" {
		return false
	}
	if len(namespace) > 63 {
		return false
	}
	return dnsLabelRegex.MatchString(namespace)
}

// hasContainer 检查 pod 是否包含指定容器
func hasContainer(pod *v1.Pod, containerName string) bool {
	if containerName == "" {
		return true
	}
	for _, c := range pod.Spec.Containers {
		if c.Name == containerName {
			return true
		}
	}
	for _, c := range pod.Spec.InitContainers {
		if c.Name == containerName {
			return true
		}
	}
	return false
}
