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
	"go.uber.org/zap"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/kubernetes/scheme"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/remotecommand"
)

const (
	ExecDefaultCommand     = "/bin/sh"
	ExecSessionTimeout  = 30 * time.Minute
	MaxExecConnections = 100
)

var activeExecConnections atomic.Int32

type ExecClientProvider interface {
	GetClientset() (*kubernetes.Clientset, error)
	GetRESTConfig() (*rest.Config, error)
}

type wsInput struct {
	conn     *websocket.Conn
	sizeChan chan *remotecommand.TerminalSize
	mu       sync.Mutex
	buf      []byte
}

type termSizeQueue struct {
	sizeChan <-chan *remotecommand.TerminalSize
	size     *remotecommand.TerminalSize
}

func (t *termSizeQueue) Next() *remotecommand.TerminalSize {
	if t.size != nil {
		size := t.size
		t.size = nil
		return size
	}
	size, ok := <-t.sizeChan
	if !ok {
		return nil
	}
	return size
}

func (w *wsInput) Read(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if len(w.buf) > 0 {
		n := copy(p, w.buf)
		w.buf = w.buf[n:]
		return n, nil
	}

	for {
		_, message, err := w.conn.ReadMessage()
		if err != nil {
			return 0, err
		}

		var msg struct {
			Type string `json:"type"`
			Cols uint16 `json:"cols"`
			Rows uint16 `json:"rows"`
		}
		if err := json.Unmarshal(message, &msg); err == nil && msg.Type == "resize" {
			w.sizeChan <- &remotecommand.TerminalSize{Width: msg.Cols, Height: msg.Rows}
			continue
		}

		if len(message) > len(p) {
			w.buf = append([]byte(nil), message[len(p):]...)
		}
		n := copy(p, message)
		return n, nil
	}
}

type wsOutput struct {
	conn *websocket.Conn
	mu   sync.Mutex
}

func (w *wsOutput) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()
	return len(p), w.conn.WriteMessage(websocket.BinaryMessage, p)
}

func RegisterExecWS(r *gin.RouterGroup, logger *zap.Logger, execClientProvider ExecClientProvider, configProvider middleware.ConfigProvider) {
	r.GET("/ws/exec", HandleExecWS(logger, execClientProvider, configProvider))
}

func HandleExecWS(logger *zap.Logger, execClientProvider ExecClientProvider, configProvider middleware.ConfigProvider) gin.HandlerFunc {
	return func(c *gin.Context) {
		tokenStr := ExtractTokenFromRequest(c)
		if tokenStr == "" {
			logger.Warn("WebSocket exec missing token")
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		claims, err := middleware.VerifyToken(tokenStr, configProvider.GetJWTSecret())
		if err != nil {
			logger.Warn("WebSocket exec token verification failed", zap.Error(err))
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		username, _ := claims["username"].(string)
		handleExecWSImpl(c, logger, execClientProvider, username)
		c.Abort()
	}
}

func handleExecWSImpl(c *gin.Context, logger *zap.Logger, execClientProvider ExecClientProvider, username string) {
	namespace := c.Query("namespace")
	podName := c.Query("pod")
	container := c.Query("container")
	commandStr := c.Query("command")

	if err := validateExecParams(namespace, podName); err != nil {
		middleware.ResponseError(c, logger, err, http.StatusBadRequest)
		return
	}

	if err := checkExecConnectionLimit(logger, username); err != nil {
		middleware.ResponseError(c, logger, err, http.StatusServiceUnavailable)
		return
	}
	defer activeExecConnections.Add(-1)

	clientset, config, err := getExecClient(execClientProvider)
	if err != nil {
		middleware.ResponseError(c, logger, err, http.StatusInternalServerError)
		return
	}

	pod, err := validatePodAndContainer(clientset, namespace, podName, container)
	if err != nil {
		middleware.ResponseError(c, logger, err, http.StatusNotFound)
		return
	}

	ws, err := upgradeExecWebSocket(c, logger, namespace, podName, container, username)
	if err != nil {
		return
	}
	defer ws.Close()

	executeRemoteCommand(ws, clientset, config, pod, container, parseExecCommand(commandStr), logger)
}

func validateExecParams(namespace, podName string) error {
	if !isValidNamespace(namespace) {
		return fmt.Errorf("invalid namespace format")
	}
	if podName == "" {
		return fmt.Errorf("missing pod name")
	}
	return nil
}

func checkExecConnectionLimit(logger *zap.Logger, username string) error {
	if activeExecConnections.Load() >= MaxExecConnections {
		logger.Warn("exec connection limit reached", zap.Int32("active", activeExecConnections.Load()), zap.String("user", username))
		return fmt.Errorf("service busy, please try again later")
	}
	activeExecConnections.Add(1)
	return nil
}

func parseExecCommand(commandStr string) []string {
	if commandStr == "" {
		return []string{ExecDefaultCommand}
	}
	return []string{ExecDefaultCommand, "-c", commandStr}
}

func getExecClient(provider ExecClientProvider) (*kubernetes.Clientset, *rest.Config, error) {
	clientset, err := provider.GetClientset()
	if err != nil {
		return nil, nil, err
	}
	config, err := provider.GetRESTConfig()
	if err != nil {
		return nil, nil, err
	}
	return clientset, config, nil
}

func validatePodAndContainer(clientset *kubernetes.Clientset, namespace, podName, container string) (*v1.Pod, error) {
	pod, err := clientset.CoreV1().Pods(namespace).Get(context.Background(), podName, metav1.GetOptions{})
	if err != nil {
		return nil, err
	}

	if container != "" && !hasContainer(pod, container) {
		return nil, fmt.Errorf("container '%s' not found in pod '%s'", container, podName)
	}
	return pod, nil
}

func upgradeExecWebSocket(c *gin.Context, logger *zap.Logger, namespace, podName, container, username string) (*websocket.Conn, error) {
	ws, err := upgrader.Upgrade(c.Writer, c.Request, buildWebSocketUpgradeHeaders(c))
	if err != nil {
		return nil, err
	}

	logger.Info("Terminal WebSocket upgraded",
		zap.String("namespace", namespace),
		zap.String("pod", podName),
		zap.String("container", container),
		zap.String("user", username),
	)

	if err := ws.WriteJSON(gin.H{
		"status":    "connected",
		"namespace": namespace,
		"pod":      podName,
		"container": container,
		"message":  fmt.Sprintf("Connected to %s/%s (%s)", namespace, podName, container),
	}); err != nil {
		ws.Close()
		return nil, err
	}
	return ws, nil
}

func executeRemoteCommand(ws *websocket.Conn, clientset *kubernetes.Clientset, config *rest.Config, pod *v1.Pod, container string, command []string, logger *zap.Logger) {
	req := clientset.CoreV1().RESTClient().
		Post().
		Resource("pods").
		Name(pod.Name).
		Namespace(pod.Namespace).
		SubResource("exec").
		VersionedParams(&v1.PodExecOptions{
			Container: container,
			Command:   command,
			Stdin:     true,
			Stdout:    true,
			Stderr:    true,
			TTY:       true,
		}, scheme.ParameterCodec)

	exec, err := remotecommand.NewSPDYExecutor(config, "POST", req.URL())
	if err != nil {
		sendWebSocketError(ws, fmt.Sprintf("Failed to create executor: %v", err), logger)
		return
	}

	ctx, cancel := context.WithTimeout(context.Background(), ExecSessionTimeout)
	defer cancel()

	sizeChan := make(chan *remotecommand.TerminalSize, 1)
	input := &wsInput{conn: ws, sizeChan: sizeChan}
	output := &wsOutput{conn: ws}

	err = exec.StreamWithContext(ctx, remotecommand.StreamOptions{
		Stdin:              input,
		Stdout:             output,
		Stderr:            output,
		Tty:               true,
		TerminalSizeQueue:  &termSizeQueue{sizeChan: sizeChan},
	})

	if err != nil && err != io.EOF {
		sendWebSocketError(ws, fmt.Sprintf("Exec failed: %v", err), logger)
	}
}

func sendWebSocketError(ws *websocket.Conn, msg string, logger *zap.Logger) {
	if err := ws.WriteJSON(gin.H{"type": "error", "message": msg}); err != nil {
		logger.Error("Failed to send error message", zap.Error(err))
	}
}

func isValidNamespace(namespace string) bool {
	if namespace == "" || len(namespace) > 63 {
		return false
	}
	return regexp.MustCompile(`^[a-z0-9]([-a-z0-9]*[a-z0-9])?$`).MatchString(namespace)
}

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