//go:build ignore

package main

// ============================================================================
// 🔰 新手必看：这是一个完整的 API 请求处理示例
// ============================================================================
//
// 这个文件展示了从接收 HTTP 请求到返回数据的完整流程
// 每一行都有详细注释，帮助你理解代码在做什么
//
// 📌 阅读建议：
// 1. 从上往下看，不要跳行
// 2. 遇到不懂的术语先忽略，理解整体流程
// 3. 配合"新手代码导读.md"一起看效果更好
//
// ============================================================================

import (
	// --- 标准库 ---
	"context"  // 上下文，用于控制请求超时和取消
	"fmt"      // 格式化输出
	"net/http" // HTTP 相关功能

	// --- 第三方库 ---
	"github.com/gin-gonic/gin" // Web 框架，处理 HTTP 请求
	"go.uber.org/zap"          // 日志库，记录程序运行信息

	// --- 项目内部包 ---
	"github.com/nick0323/K8sVision/model" // 数据模型（定义数据结构）
)

// ============================================================================
// 第一部分：路由注册
// ============================================================================
// 作用：告诉程序"当用户访问某个 URL 时，调用哪个函数处理"
// 类似于餐厅的菜单，告诉服务员什么菜对应哪个厨师

// RegisterPod 注册 Pod 相关的路由
// 参数说明：
//   - r: 路由组，可以理解为"API 接口集合"
//   - logger: 日志记录器，用来记录程序运行信息
//   - getK8sClient: 获取 K8s 客户端的函数（用于连接 Kubernetes）
//   - listPodsWithRaw: 获取 Pod 数据的函数
func RegisterPod(
	r *gin.RouterGroup,
	logger *zap.Logger,
	getK8sClient K8sClientProvider,
	listPodsWithRaw func(context.Context, *kubernetes.Clientset, model.PodMetricsMap, string) ([]model.PodStatus, *v1.PodList, error),
) {
	// 注册两个路由：

	// 1. GET /api/pods - 获取 Pod 列表
	//    当用户访问这个 URL 时，调用 getPodList 函数处理
	r.GET("/pods", getPodList(logger, getK8sClient, listPodsWithRaw))

	// 2. GET /api/pods/:namespace/:name - 获取某个 Pod 的详情
	//    :namespace 和 :name 是参数，比如 /api/pods/default/my-pod
	r.GET("/pods/:namespace/:name", getPodDetail(logger, getK8sClient))
}

// ============================================================================
// 第二部分：处理列表请求
// ============================================================================
// 这是核心部分！处理用户"获取 Pod 列表"的请求

// getPodList 获取 Pod 列表的处理函数
// 返回值 gin.HandlerFunc 是一个函数，这个函数处理具体的 HTTP 请求
func getPodList(
	logger *zap.Logger,
	getK8sClient K8sClientProvider,
	listPodsWithRaw func(context.Context, *kubernetes.Clientset, model.PodMetricsMap, string) ([]model.PodStatus, *v1.PodList, error),
) gin.HandlerFunc {

	// 这里返回了一个函数，这个函数才是真正处理请求的
	return func(c *gin.Context) {

		// --- 第 1 步：获取请求参数 ---
		// 用户可能传递了 namespace 参数，比如 /api/pods?namespace=default
		namespace := c.Query("namespace")
		// 如果用户没传，就用空字符串（表示获取所有命名空间的 Pod）

		// --- 第 2 步：获取 K8s 客户端 ---
		// 客户端是用来和 Kubernetes API 服务器通信的
		clientset, metricsClient, err := getK8sClient()
		if err != nil {
			// 如果获取客户端失败，记录日志并返回错误
			logger.Error("获取 K8s 客户端失败", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    500,
				"message": "获取 K8s 客户端失败",
			})
			return
		}

		// --- 第 3 步：获取 Pod 指标数据（CPU、内存使用量）---
		// metrics-client 是用来获取资源使用情况的
		metricsList, _ := metricsClient.MetricsV1beta1().PodMetricses(namespace).List(
			context.Background(),
			metav1.ListOptions{},
		)

		// 把指标数据整理成一个 Map，方便后续查找
		podMetricsMap := make(model.PodMetricsMap)
		if metricsList != nil {
			for _, m := range metricsList.Items {
				// 计算每个 Pod 的 CPU 和内存总和
				var cpuSum, memSum int64
				for _, ctn := range m.Containers {
					cpuSum += ctn.Usage.Cpu().MilliValue() // CPU 单位是毫核（1000m = 1 核）
					memSum += ctn.Usage.Memory().Value()   // 内存单位是字节
				}
				// 以 "namespace/name" 为键存储
				podMetricsMap[m.Namespace+"/"+m.Name] = model.PodMetrics{
					CPU: cpuSum,
					Mem: memSum,
				}
			}
		}

		// --- 第 4 步：获取 Pod 列表 ---
		// 调用 Service 层的函数获取数据
		podStatuses, _, err := listPodsWithRaw(
			context.Background(), // 上下文
			clientset,            // K8s 客户端
			podMetricsMap,        // 指标数据
			namespace,            // 命名空间
		)

		if err != nil {
			// 如果获取失败，记录日志并返回错误
			logger.Error("获取 Pod 列表失败", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    500,
				"message": "获取 Pod 列表失败：" + err.Error(),
			})
			return
		}

		// --- 第 5 步：返回成功响应 ---
		// 把数据以 JSON 格式返回给前端
		c.JSON(http.StatusOK, gin.H{
			"code":    0,
			"message": "获取成功",
			"data":    podStatuses,      // Pod 列表数据
			"total":   len(podStatuses), // 总数
		})
	}
}

// ============================================================================
// 第三部分：处理详情请求
// ============================================================================
// 获取单个 Pod 的详细信息

// getPodDetail 获取 Pod 详情的处理函数
func getPodDetail(
	logger *zap.Logger,
	getK8sClient func() (*kubernetes.Clientset, *versioned.Clientset, error),
) gin.HandlerFunc {
	return func(c *gin.Context) {

		// --- 第 1 步：从 URL 中获取参数 ---
		// 比如 /api/pods/default/my-pod
		// c.Param("namespace") 会返回 "default"
		// c.Param("name") 会返回 "my-pod"
		namespace := c.Param("namespace")
		name := c.Param("name")

		// --- 第 2 步：获取 K8s 客户端 ---
		clientset, _, err := getK8sClient()
		if err != nil {
			logger.Error("获取 K8s 客户端失败", zap.Error(err))
			c.JSON(http.StatusInternalServerError, gin.H{
				"code":    500,
				"message": "获取 K8s 客户端失败",
			})
			return
		}

		// --- 第 3 步：从 K8s 获取 Pod 详情 ---
		// Get 方法需要三个参数：
		// 1. context: 上下文
		// 2. name: Pod 名字
		// 3. GetOptions: 获取选项（通常用空对象）
		pod, err := clientset.CoreV1().Pods(namespace).Get(
			context.Background(),
			name,
			metav1.GetOptions{},
		)

		if err != nil {
			// 如果获取失败（可能是 Pod 不存在），返回 404
			logger.Error("获取 Pod 详情失败", zap.Error(err))
			c.JSON(http.StatusNotFound, gin.H{
				"code":    404,
				"message": "Pod 不存在：" + err.Error(),
			})
			return
		}

		// --- 第 4 步：数据转换 ---
		// K8s 返回的数据结构很复杂，需要转换成前端友好的格式

		// 提取容器信息
		containers := make([]string, 0, len(pod.Spec.Containers))
		for _, ctn := range pod.Spec.Containers {
			// 格式："容器名 (镜像名)"
			containers = append(containers, ctn.Name+" ("+ctn.Image+")")
		}

		// 构建返回给前端的数据结构
		podDetail := model.PodDetail{
			CommonResourceFields: model.CommonResourceFields{
				Namespace: pod.Namespace,            // 命名空间
				Name:      pod.Name,                 // Pod 名字
				Status:    string(pod.Status.Phase), // 状态（Running/Pending 等）
				BaseMetadata: model.BaseMetadata{
					Labels:      pod.Labels,      // 标签
					Annotations: pod.Annotations, // 注解
				},
			},
			PodIP:      pod.Status.PodIP,                                   // Pod IP 地址
			NodeName:   pod.Spec.NodeName,                                  // 在哪个节点上
			StartTime:  pod.Status.StartTime.Format("2006-01-02 15:04:05"), // 启动时间
			Containers: containers,                                         // 容器列表
		}

		// --- 第 5 步：返回成功响应 ---
		c.JSON(http.StatusOK, gin.H{
			"code":    0,
			"message": "获取成功",
			"data":    podDetail,
		})
	}
}

// ============================================================================
// 🎓 总结：一个完整请求的处理流程
// ============================================================================
//
// 1. 用户发送请求：GET /api/pods?namespace=default
//                    ↓
// 2. 路由匹配：找到 RegisterPod 中注册的 r.GET("/pods", ...)
//                    ↓
// 3. API 层处理：getPodList 函数
//    - 获取参数（namespace）
//    - 获取 K8s 客户端
//    - 获取指标数据
//    - 调用 Service 层
//                    ↓
// 4. Service 层处理：listPodsWithRaw 函数
//    - 调用 K8s API 获取 Pod 列表
//    - 数据转换（K8s 格式 → 前端格式）
//                    ↓
// 5. 返回 JSON 响应：
//    {
//      "code": 0,
//      "message": "获取成功",
//      "data": [...]
//    }
//
// ============================================================================

// 💡 学习建议：
// 1. 先理解这个文件的代码
// 2. 然后打开 api/deployment.go 对比着看
// 3. 你会发现 90% 的代码都是一样的！
// 4. 这就是这个项目的"套路"
