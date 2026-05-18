# KubeVision 优化计划

> 基于 7 个维度的代码审查：DRY、Consistency、Single Responsibility、Performance、Safety、Maintainability、Separation of Concerns

---

## P0 — 高优先级（有隐患/影响体验）

### 0-1. Cluster Update 非原子操作

**文件**: `api/cluster_handler.go`
**问题**: `updateCluster` 先 `RemoveFromConfig` 再 `AddCluster`，如果 `AddCluster` 失败会导致配置丢失
**方案**: 将 `SaveToConfig` 放到最后，失败时不写 config

```diff
+ k8sCfg := ...
+ if err := svc.TestConnection(ctx, k8sCfg); err != nil { return }
+ if err := svc.AddCluster(ctx, name, k8sCfg); err != nil { return }
- svc.RemoveFromConfig(name)
- svc.AddCluster(...)
- svc.SaveToConfig(...)
+ svc.SaveToConfig(name, clusterCfg)  // 只有全部成功才持久化
```

### 0-2. Cache Get() 锁降级窗口

**文件**: `cache/memory.go:172-192`
**问题**: `Get()` 做 `RUnlock()` → `Lock()` 期间，其他 goroutine 可删除/修改同一 key
**方案**: 改为单次 `Lock()` 或提取过期清理到独立流程

```go
// 方案 A: 直接用写锁（16 shards 下可接受）
func (c *MemoryCache[T]) Get(key string) (T, bool) {
    s := c.getShard(key)
    s.mutex.Lock()
    defer s.mutex.Unlock()
    // ...
}
```

---

## P1 — 中优先级（一致性/质量）

### 1-1. 用泛型消除 List* 模板函数

**文件**: `service/list.go` (328行)
**问题**: 20 个函数 `ListPods`/`ListDeployments`/`ListStatefulSets`... 全是同一个 8 行模板，仅调用的 client 方法和 Map 函数不同
**方案**: 引入泛型包装

```go
func listResource[T any](
    ctx context.Context,
    clientset kubernetes.Interface,
    namespace, labelSelector, fieldSelector string,
    listFn func(kubernetes.Interface, string, metav1.ListOptions) ([]T, error),
) ([]T, error) {
    opts := DefaultListOptions()
    opts.LabelSelector = labelSelector
    opts.FieldSelector = fieldSelector
    return listFn(clientset, namespace, opts.Apply())
}
```

### 1-2. 合并 pkg/k8s/resource.go 的四张 map

**文件**: `pkg/k8s/resource.go` (1084行)
**问题**: `NewGetters` / `NewUpdaters` / `NewCreators` / `NewDeleters` 四张 map + `GetKindByResourceType` switch，每种资源重复 5 次
**方案**: 合并为单一 `ResourceRegistry`

```go
type ResourceRegistration struct {
    Kind          string
    ClusterScoped bool
    Get           func(kubernetes.Interface, string, string) (any, error)
    List          func(kubernetes.Interface, string, metav1.ListOptions) (any, error)
    Create        func(kubernetes.Interface, string, any) error
    Update        func(kubernetes.Interface, string, string, any) error
    Delete        func(kubernetes.Interface, string, string) error
}

var registry = map[ResourceType]*ResourceRegistration{...}
```

### 1-3. 统一 API 调用模式 — 全部走 apiClient

**文件**: `ui/src/pages/ResourceDetailPage.tsx`
**问题**: scale/restart/delete 用 raw `authFetch` 绕过 `apiClient` 的重试/超时/集群参数注入
**方案**: 替换为 `apiClient.request()`，或在 `apiClient` 上加 type-safe 的快捷方法

```diff
- await authFetch(`/api/v1/...?cluster=${cluster}`, { method: 'POST' })
+ await apiClient.request('POST', `/api/v1/...`)
```

### 1-4. 内联 `<style>` 提取为 CSS 文件

**文件**: `ui/src/pages/ClusterManagementPage.tsx:368-530`
**问题**: 162 行 CSS 写在 JSX `<style>` 里，其他页面用 `.css` 文件
**方案**: 新建 `ClusterManagementPage.css`，移入所有样式

---

## P2 — 低优先级（可维护性/整洁）

### 2-1. convertToSearchableItems 与 resource_mapper 合并

**文件**: `service/resource_manager.go:114-312` + `service/resource_mapper.go`
**问题**: 30 路 type switch 路由到 Map* 函数，本质是第二份映射
**方案**: 在 `pkg/k8s/resource.go` 的 registry 中注册 Mapper 函数，消除 `convertToSearchableItems`

### 2-2. related_finder.go 的 16 个 finder 去重

**文件**: `service/related_finder.go` (717行)
**问题**: 每个 finder 重复 List → 过滤 → add 三步，secret/configmap finder 有深层嵌套循环
**方案**: 提取公共 finder 骨架，finder 只提供过滤逻辑

```go
type RelatedFinder interface {
    Find(ctx context.Context, clientset kubernetes.Interface, resource model.ResourceRef) ([]model.RelatedResource, error)
}

// 通用骨架
func findRelated[T any](
    ctx context.Context,
    clientset kubernetes.Interface,
    namespace string,
    listFn func() ([]T, error),
    matchFn func(T, model.ResourceRef) bool,
) ([]model.RelatedResource, error)
```

### 2-3. ArgoCD 状态颜色三重复合并

**文件**: `ui/src/pages/ArgoCDPage.tsx`
**问题**: `getStatusColor` / `getStatusBgColor` / `getStatusIcon` 三个函数对同一组状态值做 3 次 switch
**方案**: 合并为配置对象

```ts
const STATUS_CONFIG = {
  Synced:      { color: '#52c41a', bg: 'rgba(82,196,26,0.1)', icon: FaCheckCircle },
  OutOfSync:   { color: '#faad14', bg: 'rgba(250,173,20,0.1)', icon: FaExclamationTriangle },
  Progressing: { color: '#1890ff', bg: 'rgba(24,144,255,0.1)', icon: FaSync },
  Missing:     { color: '#ff4d4f', bg: 'rgba(245,34,45,0.1)',  icon: FaQuestionCircle },
  Unknown:     { color: '#666',    bg: 'rgba(0,0,0,0.05)',    icon: FaMinusCircle },
} as const;
```

### 2-4. Sidebar 拆分为三个独立组件

**文件**: `ui/src/common/Sidebar.tsx` (436行)
**问题**: 导航菜单 + 集群选择器 + Settings 弹窗 + 密码修改弹窗混在一起
**方案**: 提取 `SettingsModal.tsx` 和 `ChangePasswordModal.tsx`

```
Sidebar.tsx (～200行)
├── 导航菜单渲染
├── 集群选择器
└── 引用 SettingsModal / ChangePasswordModal
```

### 2-5. 静态内联 style 改为 CSS class

**文件**: 散见 `ClusterManagementPage.tsx`、`CRDPage.tsx`、`PodsTab.tsx` 等
**问题**: 52 处 `style={}`，其中 20+ 是纯静态值（如 `<th style={{width:'15%'}}>`），每次 render 创建新对象
**方案**: 定义 CSS class，必要时用 `data-*` 属性驱动动态值

### 2-6. 添加 React.memo 减少重渲染

**文件**: `ui/src/pages/ArgoCDPage.tsx`、`OverviewTab.tsx`、`ResourceListPage.tsx`
**问题**: 列表页搜索/筛选时所有行/卡片重渲染；`Table.tsx` 的子组件有 memo 但上层没用
**方案**: `React.memo` + `useCallback` 策略性添加:
- `ArgoCDPage` → app cards
- `OverviewTab` → container cards  
- `ResourceSummary` / `InfoCard` / `EventCard`

### 2-7. Config.GetConfig() 返回副本

**文件**: `config/manager.go`
**问题**: `GetConfig()` 返回 `*Config` 指针，调用方可修改内部状态
**方案**: 返回副本而非指针

---

## 优先级速查

| ID | 项目 | 影响 | 文件数改动 | 预估工时 |
|----|------|------|-----------|---------|
| P0-1 | Cluster update 原子性 | 数据丢失风险 | 1 | ~0.5h |
| P0-2 | Cache 锁竞争 | 偶发竞态 | 1 | ~1h |
| P1-1 | List* 泛型化 | 减少 300 行模板 | 1 | ~2h |
| P1-2 | resource.go 四合一 | 减少 600 行 | 1 | ~3h |
| P1-3 | 统一 apiClient | 一致性 | 1 | ~1h |
| P1-4 | 内联 style 提取 | 一致性 | 2 | ~0.5h |
| P2-1 | searchableItems 合并 | 减少重复路由 | 1 | ~2h |
| P2-2 | related_finder 去重 | 减少 300 行 | 1 | ~3h |
| P2-3 | ArgoCD 状态合并 | 减少 40 行 | 1 | ~0.5h |
| P2-4 | Sidebar 拆分 | 可维护性 | 4 | ~2h |
| P2-5 | 静态 style 清理 | 一致性 | 6+ | ~1h |
| P2-6 | React.memo | 渲染性能 | 5+ | ~1h |
| P2-7 | Config 返回副本 | 安全 | 1 | ~0.5h |

---

*生成日期: 2026-05-18*
*分析框架: DRY / Consistency / Single Responsibility / Performance / Safety / Maintainability / Separation of Concerns*
