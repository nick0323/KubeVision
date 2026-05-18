# KubeVision 架构文档

## 系统架构

```
┌──────────────────────────────────────────────────────────┐
│                    Browser (React SPA)                    │
│  Pages (列表/详情)  →  Common (Table/Sidebar/Tabs)      │
│  Hooks (useResourceList)  →  apiClient (统一 HTTP)      │
└──────────────────────────┬──────────────────────────────┘
                           │ HTTP / WebSocket
                           ↓
┌──────────────────────────────────────────────────────────┐
│                     Go Backend (Gin)                     │
│  ┌─────────────────────────────────────────────────┐    │
│  │              API Layer                          │    │
│  │  resource_handler → ResourceEntry registry     │    │
│  │  operations (YAML/logs/exec/related)           │    │
│  │  middleware (auth/cors/logging/metrics)        │    │
│  │  cluster_handler / password_management         │    │
│  └─────────────────────────────────────────────────┘    │
│                          ↓                               │
│  ┌─────────────────────────────────────────────────┐    │
│  │            Service Layer                        │    │
│  │  ClientManager → 多集群 K8s 客户端管理          │    │
│  │  ResourceManager → 列表/搜索/映射               │    │
│  │  RelatedFinder → 16 种资源的关联查找             │    │
│  │  Informer → Pod 变更监听                        │    │
│  └─────────────────────────────────────────────────┘    │
│                          ↓                               │
│  ┌─────────────────────────────────────────────────┐    │
│  │         Infrastructure Layer                    │    │
│  │  MemoryCache (LRU + 16 shards + TTL)           │    │
│  │  ConfigManager (Viper + 值拷贝返回)             │    │
│  │  Monitor (Prometheus 指标)                      │    │
│  └─────────────────────────────────────────────────┘    │
└──────────────────────────┬──────────────────────────────┘
                           │ client-go
                           ↓
┌──────────────────────────────────────────────────────────┐
│                  Kubernetes API Server                    │
└──────────────────────────────────────────────────────────┘
```

---

## 核心模块

### 1. ResourceEntry 注册表 (`pkg/k8s/resource.go`)

所有 K8s 资源的 CRUD 操作集中管理。核心结构：

```go
type ResourceEntry struct {
    Kind          string
    ClusterScoped bool
    Get           func(kubernetes.Interface, string, string) (any, error)
    List          func(kubernetes.Interface, string, metav1.ListOptions) (any, error)
    Create        func(kubernetes.Interface, string, any) error
    Update        func(kubernetes.Interface, string, string, any) error
    Delete        func(kubernetes.Interface, string, string) error
}

func NewRegistry(c kubernetes.Interface) map[ResourceType]*ResourceEntry
```

新增资源只需在 `NewRegistry` 中添加一条记录，无需新增文件或 switch-case。

### 2. 多集群客户端管理 (`service/k8s_client.go`)

```go
type ClientManager struct {
    defaultClient *ClientHolder   // config.yaml 顶层 kubernetes: 段
    clientPool    sync.Map        // clusters: 列表中的命名集群
}
```

- `GetClient("")` / `GetClient("default")` → 路由到 `defaultClient`
- `GetClusterNames()` → 始终将 `"default"` 放在首位
- `RemoveCluster("default")` → 返回错误，不可删除
- `GetClustersHealth()` → 同时检查所有集群健康状态

### 3. 关联资源查找 (`service/related_finder.go`)

```go
var relatedFinders map[string]func(*findContext)

func init() {
    relatedFinders = map[string]func(*findContext){
        "Pod":         func(fc *findContext) { /* ownerRef + service + volumes + node */ },
        "Deployment":  func(fc *findContext) { /* RS + service + HPA + PDB + ingress */ },
        // ... 共 16 种资源
    }
}

func FindRelatedResources(obj any, resourceType string, ...) []any
```

使用 `k8s.GetKindByResourceType(resourceType)` 映射类型名，直接在闭包 map 中查找，无需 switch。

### 4. 缓存层 (`cache/memory.go`)

泛型 LRU 缓存，16 个 shard 减少锁竞争：

```go
type MemoryCache[T any] struct {
    shards [16]cacheShard[T]
}

type cacheShard[T any] struct {
    mutex sync.Mutex
    data  map[string]cacheItem[T]
    order *list.List
}
```

- `Get()` / `Set()` 使用单次 `Lock()`（无 RUnlock→Lock 窗口）
- 淘汰策略：LRU，全局 maxSize 分摊到每个 shard
- TTL：列表 5min，详情 2min，ConfigMap/Secret 30s
- 删除/更新时自动失效对应缓存

### 5. 密码管理 (`api/password_management.go`)

```go
type PasswordManager struct {
    configMgr *config.Manager
}

func (pm *PasswordManager) ValidatePasswordStrength(password string) error
```

验证规则：
- 长度 8-128
- 至少 3 种字符类型（大写/小写/数字/特殊符号）
- 禁止 3+ 连续递增数字
- 禁止单一字符占比 >50%
- 禁止常见弱密码
- 禁止与最近 5 次历史密码相同

### 6. Config 安全 (`config/manager.go`)

```go
func (m *Manager) GetConfig() Config       // 返回值拷贝，非指针
func (m *Manager) UpdateConfig(fn func(*Config)) // 原子更新
```

`GetConfig()` 返回值拷贝，调用方无法意外修改内部状态。测试用 `UpdateConfig` 避免状态污染。

---

## 数据流

### 资源列表查询

```
前端请求 /api/pods?namespace=default
  ↓
resource_handler.go → 解析 resourceType
  ↓
NewRegistry(clientset)[ResourcePod].List()
  ↓
cache.Get("list:pod:default:")  ← 命中则直接返回
  ↓ 未命中
clientset.CoreV1().Pods("default").List()
  ↓
resource_mapper.go → 映射为前端模型
  ↓
cache.Set("list:pod:default:", result, 5min)
  ↓
返回 JSON
```

### WebSocket 日志流

```
前端 → /ws/logs?namespace=default&pod=nginx&token=xxx
  ↓
middleware/auth.go → 验证 JWT
  ↓
operations.go → 建立 Pod 日志流 (client-go)
  ↓
gorilla/websocket ↔ kubernetes.APIContainer.Logs() 双向转发
```

### 多集群请求路由

```
前端请求 (cluster=dev)
  ↓
apiClient 自动注入 cluster 参数 (default 时不传)
  ↓
resource_handler → GetClient(clusterName)
  ├─ "default" / "" → defaultClient
  └─ 其他 → clientPool.Load(clusterName)
```

---

## 扩展指南

### 添加新资源类型

1. **注册表**: 在 `pkg/k8s/resource.go` `NewRegistry()` 中添加 `ResourceEntry`
2. **搜索映射**: 在 `service/resource_manager.go` `searchItemMappers()` 中添加
3. **关联查找** (可选): 在 `service/related_finder.go` `init()` 中添加
4. **前端常量**: 在 `ui/src/constants/index.ts` 添加资源类型
5. **页面配置**: 在 `ui/src/constants/pageConfigs.ts` 添加列表/详情配置
6. **详情 Tab** (可选): 在 `ui/src/tabs/` 添加组件

### 添加新集群

在 `config.yaml` 的 `clusters:` 列表中添加：

```yaml
clusters:
  - name: prod
    kubeconfig: /path/to/prod/kubeconfig
    apiserver: ""
    token: ""
    insecure: false
```

通过 UI "Cluster Management" 页面也可动态添加/删除。

---

## 已完成的优化

| 项目 | 描述 |
|------|------|
| Cluster update 原子性 | TestConnection + AddCluster 通过后再写配置 |
| Cache 锁安全 | Get() 单次 Lock() 消除竞态窗口 |
| ResourceEntry 注册表 | 5 接口 + 4 工厂 + switch → 1 个 map |
| related_finder 去重 | 16 结构体 + interface → map 闭包 |
| Sidebar 拆分 | 436 行 → 4 个独立组件 |
| Config 安全 | 返回值拷贝 + UpdateConfig 测试辅助 |
| React.memo | 15+ 展示组件添加 memo |
| CSS 变量化 | 硬编码值 → CSS 变量 |
| GetKindByResourceType | switch → map 查找 |
