# KubeVision 架构文档

## 系统架构概述

KubeVision 是一个基于 Go + React 的 Kubernetes Web 管理面板，采用前后端分离架构。

### 整体架构

```
┌──────────────────────────────────────────────────────────┐
│                        Browser                           │
│  ┌─────────────────────────────────────────────────┐    │
│  │           React SPA (TypeScript)                │    │
│  │  - Pages (列表/详情/日志/终端)                  │    │
│  │  - Components (Table/Sidebar/Tabs)              │    │
│  │  - Hooks (useResourceList/useResourceDetail)    │    │
│  └─────────────────────────────────────────────────┘    │
└──────────────────────────┬──────────────────────────────┘
                           │ HTTP / WebSocket
                           ↓
┌──────────────────────────────────────────────────────────┐
│                   Go Backend (Gin)                       │
│  ┌─────────────────────────────────────────────────┐    │
│  │              API Layer                          │    │
│  │  - HTTP Handlers (resource_handler.go)         │    │
│  │  - WebSocket (exec.go, operations.go)          │    │
│  │  - Middleware (auth/cors/logging/metrics)      │    │
│  └─────────────────────────────────────────────────┘    │
│                          ↓                               │
│  ┌─────────────────────────────────────────────────┐    │
│  │            Service Layer                        │    │
│  │  - K8s Client Manager (k8s_client.go)          │    │
│  │  - Resource Listers (list.go)                  │    │
│  │  - Pod Informer (pod_informer.go)              │    │
│  └─────────────────────────────────────────────────┘    │
│                          ↓                               │
│  ┌─────────────────────────────────────────────────┐    │
│  │         Infrastructure Layer                    │    │
│  │  - Cache (memory.go - LRU)                     │    │
│  │  - Monitor (metrics/tracing)                   │    │
│  │  - Config (manager.go - Viper)                 │    │
│  └─────────────────────────────────────────────────┘    │
└──────────────────────────┬──────────────────────────────┘
                           │ client-go
                           ↓
┌──────────────────────────────────────────────────────────┐
│              Kubernetes API Server                       │
└──────────────────────────────────────────────────────────┘
```

## 核心模块设计

### 1. API 层 (api/)

**职责**: HTTP 请求处理、WebSocket 连接、中间件

#### 关键文件
- `resource_handler.go` - 通用资源 CRUD 路由
- `operations.go` - YAML 操作、关联资源、日志流
- `exec.go` - Pod Exec WebSocket
- `middleware/` - 认证、CORS、日志、指标

#### 缓存集成
```go
// 列表查询 - 带缓存
result, err := getResourceListByTypeWithCache(...)

// 详情查询 - 带缓存
obj, err := getResourceByName(...)

// 删除/更新 - 清除缓存
invalidateCache(resourceType, namespace)
```

### 2. Service 层 (service/)

**职责**: 业务逻辑、K8s 操作、资源映射

#### 关键组件

**ClientManager** - K8s 客户端管理器
```go
type ClientManager struct {
    defaultClient  *ClientHolder
    clientPool     sync.Map  // 多集群连接池
}

// 获取客户端（支持多集群）
func (m *ClientManager) GetClient(clusterName string)
```

**PodInformer** - Pod 缓存管理器
```go
type PodInformer struct {
    podCache     map[string]*v1.Pod
    restartCache map[string]int32
}

// 从缓存列出 Pod
func (pi *PodInformer) ListPods(namespace string) []*v1.Pod
```

### 3. 缓存层 (cache/)

**设计**: 泛型 LRU 缓存，支持 TTL 和自动清理

```go
type MemoryCache[T any] struct {
    data            map[string]CacheItem[T]
    maxSize         int
    ttl             time.Duration
    hits, misses    atomic.Int64
}

// 使用示例
cache.SetWithTTL("pod:default:nginx", pods, 5*time.Minute)
if cached, found := cache.Get("pod:default:nginx"); found {
    return cached
}
```

#### 缓存策略
- **列表查询**: 5 分钟 TTL
- **资源详情**: 2 分钟 TTL
- **淘汰策略**: LRU（最近最少使用）
- **自动清理**: 定期清理过期项

### 4. 多集群架构

#### 配置加载
```yaml
clusters:
  - name: "cluster-dev"
    kubeconfig: "/path/to/dev-kubeconfig"
  - name: "cluster-prod"
    apiServer: "https://prod-api:6443"
    token: "xxx"
```

#### 动态加载流程
```
1. 请求指定 clusterName
2. 检查 clientPool 是否已缓存
3. 未命中 → 从配置加载 ClusterConfig
4. 创建 ClientHolder 并 Store 到 clientPool
5. 返回客户端
```

#### 连接池管理
```go
// LoadOrStore 保证原子性，避免并发加载
actual, loaded := m.clientPool.LoadOrStore(clusterName, holder)
if loaded {
    holder.Close()  // 丢弃重复创建
    return actual.(*ClientHolder)
}
```

### 5. 安全机制

#### JWT 认证流程
```
1. 用户登录 → 验证用户名/密码
2. 生成 JWT token (24h 过期)
3. 后续请求携带 token → 中间件验证
4. 失败 5 次 → 锁定 10 分钟
```

#### 密码安全
- 使用 bcrypt 哈希存储
- 首次启动自动生成随机密码
- 支持登录后修改

#### WebSocket 安全
- 携带 JWT token 验证
- CheckOrigin 可配置
- 连接超时和心跳检测

## 数据流

### 资源列表查询流程
```
前端请求
  ↓
API Handler (resource_handler.go)
  ↓
检查缓存
  ├─ 命中 → 返回缓存
  └─ 未命中 ↓
Service Layer (list.go)
  ↓
K8s API (client-go)
  ↓
写入缓存 (5min TTL)
  ↓
返回前端
```

### Pod 终端流程
```
前端 WebSocket 连接
  ↓
/api/ws/exec
  ↓
验证 JWT token
  ↓
创建 Pod Exec (client-go)
  ↓
WebSocket ↔ exec 双向转发
  ↓
连接关闭 → 清理资源
```

## 性能优化

### 1. 缓存优化
- 列表查询减少 K8s API 调用
- LRU 淘汰避免内存溢出
- 删除/更新时自动失效

### 2. 连接优化
- K8s 客户端连接池复用
- WebSocket 心跳保活
- HTTP 连接复用

### 3. 前端优化
- 虚拟滚动（日志列表）
- 按需加载（Tab 切换）
- 防抖/节流（搜索/resize）

## 监控指标

### 缓存指标
```go
type CacheStats struct {
    Hits          int64
    Misses        int64
    HitRate       float64
    Size          int
    Utilization   float64
}
```

### 请求指标
- 请求延迟（P50/P99）
- 错误率
- QPS

## 部署架构

### Docker 部署
```
┌─────────────────────┐
│  KubeVision Pod     │
│  ┌───────────────┐  │
│  │ Go Backend    │  │
│  │   (8080)      │  │
│  └───────────────┘  │
│  ┌───────────────┐  │
│  │ React Static  │  │
│  │   (Nginx)     │  │
│  └───────────────┘  │
└─────────────────────┘
```

### K8s 部署
```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: kubevision
spec:
  template:
    spec:
      serviceAccountName: kubevision-sa
      containers:
      - name: kubevision
        image: kubevision:latest
        env:
        - name: K8SVISION_JWT_SECRET
          valueFrom:
            secretKeyRef:
              name: kubevision-secret
              key: jwt-secret
```

## 扩展指南

### 添加新资源类型

1. **Service 层**: 添加列表/详情函数
   ```go
   func ListNewResource(ctx context.Context, clientset *kubernetes.Clientset, ...) ([]model.SearchableItem, error)
   ```

2. **API 层**: 添加 switch-case
   ```go
   case "newresource":
       return service.ListNewResource(...)
   ```

3. **前端**: 添加路由和页面

### 自定义缓存策略

修改 `config.yaml`:
```yaml
cache:
  ttl: "10m"        # 调整 TTL
  maxSize: 2000     # 调整容量
```

## 故障排查

### 缓存问题
```bash
# 查看缓存统计
curl http://localhost:8080/cache/stats
```

### 连接问题
```bash
# 检查 K8s 连接
curl http://localhost:8080/health
```

### 日志级别
```yaml
log:
  level: "debug"  # 调整为 debug 查看详细日志
```

## 未来规划

1. **RBAC 权限系统**: 多用户、角色权限、命名空间级别访问控制
2. **Dynamic Client**: 使用 dynamic client 重构 switch-case，支持 CRD
3. **Informer 扩展**: 为更多资源类型添加 Informer 缓存
4. **多集群 UI**: 前端支持切换集群查看
5. **告警系统**: 基于监控指标的自动告警
