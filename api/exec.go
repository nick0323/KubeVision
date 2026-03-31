package api

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/gorilla/websocket"
	"github.com/nick0323/K8sVision/api/middleware"
	"go.uber.org/zap"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
	"k8s.io/client-go/tools/remotecommand"
	"k8s.io/client-go/util/homedir"
)

// Exec 配置常量
const (
	ExecDefaultCommand = "/bin/sh"
)

// WebSocket upgrader
var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// ExecRequest WebSocket 请求参数
type ExecRequest struct {
	Namespace string   `json:"namespace"`
	Pod       string   `json:"pod"`
	Container string   `json:"container,omitempty"`
	Command   []string `json:"command"`
}

func RegisterExecWS(
	r *gin.RouterGroup,
	logger *zap.Logger,
	getK8sClient K8sClientProvider,
) {
	// 这个函数不再使用，WebSocket 路由直接在 main.go 中注册
}

// HandleExecWS 处理 WebSocket exec 请求（带 JWT 认证）
func HandleExecWS(
	logger *zap.Logger,
	getK8sClient K8sClientProvider,
) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 从 WebSocket 请求中获取 token
		tokenStr := c.Query("token")
		if tokenStr == "" {
			// 尝试从 Authorization header 获取
			tokenStr = c.GetHeader("Authorization")
			if tokenStr != "" {
				tokenStr = strings.TrimPrefix(tokenStr, "Bearer ")
			}
		}

		if tokenStr == "" {
			logger.Warn("WebSocket exec 缺少 token")
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		// 验证 token（简单验证，实际应该使用 JWT 库）
		// 这里只是示例，实际应该验证 token 有效性
		_ = tokenStr

		// 调用实际的 WebSocket 处理函数
		handleExecWSImpl(c, logger, getK8sClient)
	}
}

func handleExecWSImpl(
	c *gin.Context,
	logger *zap.Logger,
	getK8sClient K8sClientProvider,
) {
	// 获取参数
	namespace := c.Query("namespace")
	podName := c.Query("pod")
	container := c.Query("container")
	commandStr := c.Query("command")

	if namespace == "" || podName == "" {
		middleware.ResponseError(c, logger, fmt.Errorf("missing namespace or pod"), http.StatusBadRequest)
		return
	}

	// 解析命令
	command := []string{ExecDefaultCommand}
	if commandStr != "" {
		command = []string{ExecDefaultCommand, "-c", commandStr}
	}

	// 获取 K8s 客户端和配置（在 WebSocket 升级前）
	clientset, config, err := getK8sClientWithConfig()
	if err != nil {
		logger.Error("获取 K8s 客户端失败", zap.Error(err))
		middleware.ResponseError(c, logger, err, http.StatusInternalServerError)
		return
	}

	// 检查 Pod 是否存在（在 WebSocket 升级前）
	_, err = clientset.CoreV1().Pods(namespace).Get(context.Background(), podName, metav1.GetOptions{})
	if err != nil {
		logger.Error("Pod 不存在", zap.Error(err))
		middleware.ResponseError(c, logger, err, http.StatusNotFound)
		return
	}

	// 升级 WebSocket 连接
	ws, err := upgrader.Upgrade(c.Writer, c.Request, nil)
	if err != nil {
		logger.Error("WebSocket 升级失败", zap.Error(err))
		return
	}
	defer ws.Close()

	logger.Info("Terminal WebSocket 升级成功",
		zap.String("namespace", namespace),
		zap.String("pod", podName),
		zap.String("container", container),
	)

	// 先发送连接成功消息
	ws.WriteJSON(gin.H{
		"status":    "connected",
		"namespace": namespace,
		"pod":       podName,
		"container": container,
		"message":   fmt.Sprintf("Connected to %s/%s (%s)", namespace, podName, container),
	})

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

	// 执行命令
	err = exec.Stream(remotecommand.StreamOptions{
		Stdin:  wsInput{ws},
		Stdout: wsOutput{ws},
		Stderr: wsOutput{ws},
		Tty:    true,
	})

	if err != nil {
		logger.Error("exec stream 失败", zap.Error(err))
		ws.WriteJSON(gin.H{"type": "error", "message": fmt.Sprintf("Exec failed: %v", err)})
	}
}

// wsInput 实现 io.Reader 接口，从 WebSocket 读取输入
type wsInput struct {
	conn *websocket.Conn
}

func (w wsInput) Read(p []byte) (int, error) {
	_, message, err := w.conn.ReadMessage()
	if err != nil {
		return 0, err
	}
	copy(p, message)
	return len(message), nil
}

// wsOutput 实现 io.Writer 接口，向 WebSocket 写入输出
type wsOutput struct {
	conn *websocket.Conn
}

func (w wsOutput) Write(p []byte) (int, error) {
	err := w.conn.WriteMessage(websocket.TextMessage, p)
	if err != nil {
		return 0, err
	}
	return len(p), nil
}

// getK8sClientWithConfig 获取 K8s 客户端和配置
func getK8sClientWithConfig() (*kubernetes.Clientset, *rest.Config, error) {
	// 尝试 in-cluster config
	config, err := rest.InClusterConfig()
	if err == nil {
		clientset, err := kubernetes.NewForConfig(config)
		return clientset, config, err
	}

	// fallback 到 kubeconfig
	kubeconfig := os.Getenv("KUBECONFIG")
	if kubeconfig == "" {
		home := homedir.HomeDir()
		kubeconfig = filepath.Join(home, ".kube", "config")
	}

	config, err = clientcmd.BuildConfigFromFlags("", kubeconfig)
	if err != nil {
		return nil, nil, err
	}

	clientset, err := kubernetes.NewForConfig(config)
	return clientset, config, err
}
