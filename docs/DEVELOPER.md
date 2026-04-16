# KubeVision 开发者指南

本文档面向 KubeVision 的开发者，介绍项目架构、编码规范和开发流程。

## 目录

- [项目架构](#项目架构)
- [技术栈](#技术栈)
- [开发环境](#开发环境)
- [代码规范](#代码规范)
- [测试指南](#测试指南)
- [调试技巧](#调试技巧)
- [贡献流程](#贡献流程)

---

## 项目架构

### 整体架构

```
┌─────────────────────────────────────────────────────────┐
│                      Frontend (React)                    │
│  ┌─────────┐  ┌─────────┐  ┌─────────┐  ┌─────────┐    │
│  │  Pods   │  │Deploy...│  │ Services│  │  Logs   │    │
│  └─────────┘  └─────────┘  └─────────┘  └─────────┘    │
└─────────────────────────────────────────────────────────┘
                            │
                            │ HTTP / WebSocket
                            ▼
┌─────────────────────────────────────────────────────────┐
│                      Backend (Go)                        │
│  ┌─────────────────────────────────────────────────┐    │
│  │                 API Layer                        │    │
│  │  ┌─────────┐  ┌─────────┐  ┌─────────┐         │    │
│  │  │Resource │  │  YAML   │  │   Log   │         │    │
│  │  │ Handler │  │ Handler │  │ Stream  │         │    │
│  │  └─────────┘  └─────────┘  └─────────┘         │    │
│  └─────────────────────────────────────────────────┘    │
│                            │                             │
│  ┌─────────────────────────────────────────────────┐    │
│  │               Middleware Layer                   │    │
│  │  JWT Auth │ CORS │ Logging │ Rate Limit        │    │
│  └─────────────────────────────────────────────────┘    │
│                            │                             │
│  ┌─────────────────────────────────────────────────┐    │
│  │               Service Layer                      │    │
│  │  ┌─────────┐  ┌─────────┐  ┌─────────┐         │    │
│  │  │  List   │  │  Watch  │  │  Exec   │         │    │
│  │  │ Service │  │ Service │  │ Service │         │    │
│  │  └─────────┘  └─────────┘  └─────────┘         │    │
│  └─────────────────────────────────────────────────┘    │
│                            │                             │
│  ┌─────────────────────────────────────────────────┐    │
│  │                Cache Layer                       │    │
│  │         LRU Cache with TTL support              │    │
│  └─────────────────────────────────────────────────┘    │
└─────────────────────────────────────────────────────────┘
                            │
                            │ K8s client-go
                            ▼
┌─────────────────────────────────────────────────────────┐
│                  Kubernetes Cluster                      │
│  ┌─────────┐  ┌─────────┐  ┌─────────┐  ┌─────────┐    │
│  │  Pods   │  │Deploy...│  │Services │  │  ...    │    │
│  └─────────┘  └─────────┘  └─────────┘  └─────────┘    │
└─────────────────────────────────────────────────────────┘
```

### 目录结构

```
KubeVision/
├── api/                        # API 层 (HTTP handlers)
│   ├── middleware/            # 中间件
│   │   ├── auth.go           # JWT 认证
│   │   ├── cors.go           # CORS
│   │   ├── error.go          # 错误处理
│   │   ├── logging.go        # 日志中间件
│   │   └── metrics.go        # 监控中间件
│   ├── exec.go               # Pod Exec WebSocket
│   ├── login.go              # 登录接口
│   ├── metrics.go            # 监控指标接口
│   ├── metrics_handler.go    # K8s Metrics handler
│   ├── operations.go         # 资源操作 (YAML, 关联资源，日志流)
│   ├── overview.go           # 集群概览
│   ├── password_management.go # 密码管理
│   └── resource_handler.go   # 通用资源接口
│
├── cache/                      # 缓存层
│   ├── memory.go              # 内存缓存实现
│   └── memory_test.go         # 缓存测试
│
├── config/                     # 配置管理
│   ├── manager.go             # 配置管理器
│   └── manager_test.go        # 配置测试
│
├── model/                      # 数据模型
│   ├── config.go              # 配置结构
│   ├── consts.go              # 常量定义
│   ├── types.go               # API 类型定义
│   └── config_test.go         # 配置验证测试
│
├── monitor/                    # 监控指标
│   ├── metrics.go             # 指标采集
│   └── metrics_test.go        # 指标测试
│
├── service/                    # 业务逻辑层
│   ├── client_manager.go      # K8s 客户端管理
│   ├── informer.go            # K8s Informer
│   ├── list.go                # 资源列表查询
│   └── k8s_util.go            # K8s 工具函数
│
├── ui/                         # 前端 React 应用
│   ├── src/
│   │   ├── components/        # React 组件
│   │   ├── pages/             # 页面
│   │   ├── services/          # API 服务
│   │   ├── types/             # TypeScript 类型
│   │   └── utils/             # 工具函数
│   └── package.json
│
├── docs/                       # 文档
│   ├── API.md                 # API 文档
│   ├── DEPLOYMENT.md          # 部署指南
│   └── DEVELOPER.md           # 开发者指南 (本文档)
│
├── cmd/                        # 命令行工具
│   └── tools/                 # 工具命令
│
├── main.go                     # 应用入口
├── config.yaml                 # 配置文件示例
├── Dockerfile                  # Docker 构建
├── go.mod                      # Go 模块定义
└── README.md                   # 项目说明
```

---

## 技术栈

### 后端

| 组件 | 技术 | 版本 | 用途 |
|------|------|------|------|
| 语言 | Go | 1.24 | 主要编程语言 |
| Web 框架 | Gin | v1.10+ | HTTP 路由和中间件 |
| K8s 客户端 | client-go | v0.33+ | Kubernetes API 访问 |
| WebSocket | gorilla/websocket | v1.5+ | WebSocket 连接 |
| JWT | golang-jwt/jwt | v5+ | Token 认证 |
| 日志 | zap | v1.27+ | 结构化日志 |
| 配置 | Viper | v1.19+ | 配置管理 |
| 密码加密 | bcrypt | - | 密码哈希 |

### 前端

| 组件 | 技术 | 版本 | 用途 |
|------|------|------|------|
| 框架 | React | 19 | UI 框架 |
| 语言 | TypeScript | 5.9 | 类型系统 |
| 构建 | Vite | 7 | 构建工具 |
| 终端 | xterm.js | v5+ | 终端模拟 |
| HTTP | Axios | v1.6+ | HTTP 客户端 |

---

## 开发环境

### 1. 安装依赖

```bash
# Go 依赖
go mod download

# 前端依赖
cd ui
npm install
```

### 2. 配置环境

```bash
# 复制环境配置示例
cp .env.example .env

# 编辑 .env 文件，设置必要的环境变量
export K8SVISION_JWT_SECRET="your-secret-key"
export K8SVISION_AUTH_USERNAME="admin"
export K8SVISION_AUTH_PASSWORD="your-password"
```

### 3. 启动开发服务器

```bash
# 方式 1: 同时启动前后端
make dev

# 方式 2: 分别启动
# 后端
go run main.go

# 前端
cd ui
npm run dev
```

### 4. 运行测试

```bash
# 所有测试
go test ./...

# 带覆盖率
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out

# 单个包测试
go test ./api/...
go test ./service/...

# 竞态检测
go test ./... -race
```

### 5. 代码检查

```bash
# 格式化
go fmt ./...

# 静态分析
go vet ./...

# Lint (需要安装 golangci-lint)
golangci-lint run
```

---

## 代码规范

### Go 代码规范

#### 命名规范

```go
// ✅ 好的命名
type PodHandler struct{}           // 类型：大驼峰
var maxRetries = 3                 // 变量：小驼峰
const DefaultTimeout = 30 * time.Second  // 常量：大驼峰
func GetUserByID(id string) (*User, error) {}  // 函数：大驼峰

// ❌ 避免的命名
var MaxRetries = 3                 // 变量不应大驼峰
type pod_handler struct{}          // 类型不应下划线
func get_user(id string) {}        // 函数不应下划线
```

#### 错误处理

```go
// ✅ 正确的错误处理
result, err := doSomething()
if err != nil {
    logger.Error("Failed to do something", zap.Error(err))
    return nil, fmt.Errorf("failed to do something: %w", err)
}

// ❌ 避免忽略错误
result, _ := doSomething()  // 除非确定可以忽略

// ✅ 错误包装
if err != nil {
    return fmt.Errorf("failed to process request: %w", err)
}
```

#### 注释规范

```go
// GetPod 获取 Pod 详情
// 参数:
//   - ctx: 上下文
//   - namespace: 命名空间
//   - name: Pod 名称
// 返回:
//   - *v1.Pod: Pod 对象
//   - error: 错误信息
func GetPod(ctx context.Context, namespace, name string) (*v1.Pod, error) {
    // 实现...
}
```

#### 并发规范

```go
// ✅ 使用 context 控制超时
ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
defer cancel()

// ✅ 使用 waitgroup 等待 goroutine
var wg sync.WaitGroup
wg.Add(1)
go func() {
    defer wg.Done()
    // 工作...
}()
wg.Wait()

// ✅ 使用 channel 传递数据
resultChan := make(chan Result, 10)
go worker(resultChan)
```

### 前端代码规范

#### TypeScript 规范

```typescript
// ✅ 使用接口定义类型
interface Pod {
  namespace: string;
  name: string;
  status: string;
}

// ✅ 使用泛型处理响应
interface APIResponse<T> {
  code: number;
  message: string;
  data: T;
}

// ✅ 使用 async/await
async function fetchPods(): Promise<Pod[]> {
  const response = await api.get('/api/pods');
  return response.data;
}
```

#### React 规范

```tsx
// ✅ 使用函数组件 + Hooks
const PodList: React.FC<Props> = ({ namespace }) => {
  const [pods, setPods] = useState<Pod[]>([]);
  
  useEffect(() => {
    fetchPods(namespace);
  }, [namespace]);
  
  return <div>...</div>;
};

// ✅ 使用 TypeScript 定义 Props
interface Props {
  namespace: string;
  onSelectPod: (pod: Pod) => void;
}
```

---

## 测试指南

### 单元测试

```go
// api/login_test.go
package api

import (
    "testing"
    "github.com/stretchr/testify/assert"
)

func TestLoginHandler_Success(t *testing.T) {
    // 准备测试数据
    req := &model.LoginRequest{
        Username: "admin",
        Password: "correct-password",
    }
    
    // 执行测试
    response := callLoginHandler(req)
    
    // 断言结果
    assert.Equal(t, 200, response.Code)
    assert.NotNil(t, response.Data.Token)
}

func TestLoginHandler_WrongPassword(t *testing.T) {
    req := &model.LoginRequest{
        Username: "admin",
        Password: "wrong-password",
    }
    
    response := callLoginHandler(req)
    
    assert.Equal(t, 401, response.Code)
    assert.Equal(t, "Invalid username or password", response.Message)
}
```

### 集成测试

```go
// api/integration_test.go
func TestResourceCRUD(t *testing.T) {
    // 1. 创建资源
    created := createDeployment(testDeployment)
    assert.NotNil(t, created)
    
    // 2. 获取资源
    retrieved := getDeployment(created.Name)
    assert.Equal(t, created.Name, retrieved.Name)
    
    // 3. 更新资源
    updated := updateDeployment(created.Name, 5)
    assert.Equal(t, int32(5), updated.Replicas)
    
    // 4. 删除资源
    deleteDeployment(created.Name)
    deleted := getDeployment(created.Name)
    assert.Nil(t, deleted)
}
```

### WebSocket 测试

```go
func TestPodLogWebSocket(t *testing.T) {
    // 建立 WebSocket 连接
    ws, _, err := websocket.DefaultDialer.Dial(
        "ws://localhost:8080/ws/logs/pod?namespace=default&pod=test-pod&token="+token,
        nil,
    )
    assert.NoError(t, err)
    defer ws.Close()
    
    // 等待连接成功消息
    _, message, err := ws.ReadMessage()
    assert.NoError(t, err)
    
    var response map[string]interface{}
    json.Unmarshal(message, &response)
    assert.Equal(t, "connected", response["type"])
}
```

---

## 调试技巧

### 日志调试

```bash
# 设置 debug 日志级别
export K8SVISION_LOG_LEVEL=debug

# 查看日志
kubectl logs -f deployment/k8svision | grep "ERROR"
```

### 本地调试 K8s

```bash
# 使用本地 kubeconfig
export KUBECONFIG=~/.kube/config

# 或使用 in-cluster 模式（在集群内运行时）
# 无需配置，自动使用 ServiceAccount
```

### 性能分析

```bash
# 启用 pprof
import _ "net/http/pprof"

# 访问 profiling 端点
go tool pprof http://localhost:8080/debug/pprof/heap
go tool pprof http://localhost:8080/debug/pprof/profile?seconds=30
```

### 网络调试

```bash
# 端口转发
kubectl port-forward svc/k8svision -n k8svision 8080:80

# 测试 API
curl -H "Authorization: Bearer $TOKEN" http://localhost:8080/api/pods
```

---

## 贡献流程

### 1. Fork 项目

```bash
git clone https://github.com/your-username/KubeVision.git
cd KubeVision
git remote add upstream https://github.com/nick0323/K8sVision.git
```

### 2. 创建分支

```bash
git checkout -b feature/your-feature-name
```

### 3. 开发并提交

```bash
# 开发...

# 确保测试通过
go test ./...

# 提交
git add .
git commit -m "feat: add your feature description"
```

### 4. 推送并创建 PR

```bash
git push origin feature/your-feature-name
# 然后到 GitHub 创建 Pull Request
```

### 提交信息规范

```
feat: 新功能
fix: 修复 bug
docs: 文档更新
style: 代码格式（不影响功能）
refactor: 重构（非新功能/修复）
test: 添加测试
chore: 构建/工具/配置更新
```

示例：
```
feat: add Pod exec WebSocket support
fix: fix memory leak in cache cleanup
docs: update API documentation
refactor: extract resource handlers to registry
```

---

## 发布流程

### 1. 更新版本号

```bash
# 更新 main.go 中的 Version 常量
# 更新 CHANGELOG.md
```

### 2. 打标签

```bash
git tag -a v2.0.0 -m "Release v2.0.0"
git push origin v2.0.0
```

### 3. 构建发布

```bash
# 构建多平台二进制
GOOS=linux GOARCH=amd64 go build -o KubeVision-linux-amd64
GOOS=darwin GOARCH=amd64 go build -o KubeVision-darwin-amd64
GOOS=windows GOARCH=amd64 go build -o KubeVision-windows-amd64.exe

# 构建 Docker 镜像
docker build -t k8svision:v2.0.0 .
docker push k8svision:v2.0.0
```

### 4. 发布到 GitHub

在 GitHub Releases 页面创建新版本，上传二进制文件。

---

## 常见问题

### Q: 如何添加新的资源类型支持？

A: 
1. 在 `service/list.go` 添加 `List<Resource>` 函数
2. 在 `api/resource_registry.go` 创建并注册 Handler
3. 在前端 `ui/src/types/k8s.ts` 添加类型定义
4. 在前端添加对应的列表/详情页面

### Q: WebSocket 连接数如何限制？

A: 在 `api/operations.go` 中使用 `atomic.Int32` 计数，超过阈值返回错误。

### Q: 如何优化 K8s API 调用性能？

A: 
1. 使用 Informer 缓存资源列表
2. 使用 label/field selector 过滤
3. 启用 LRU 缓存
4. 调整 QPS/Burst 配置

---

## 参考资源

- [Gin 文档](https://gin-gonic.com/docs/)
- [Kubernetes client-go](https://github.com/kubernetes/client-go)
- [Zap 日志](https://pkg.go.dev/go.uber.org/zap)
- [React 文档](https://react.dev/)
- [TypeScript 文档](https://www.typescriptlang.org/docs/)

---

## 联系

遇到问题请提交 Issue 或联系维护者。
