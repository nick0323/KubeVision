# KubeVision 代码优化任务

> 创建日期: 2026-05-12
>
> 优先级: P0(紧急) → P1(高) → P2(中) → P3(低)
> 投入产出: ★★★(最高) → ★★☆(中) → ★☆☆(低)
> ✅ = 已完成  🔄 = 进行中  ⬜ = 待开始

---

## P0 — 紧急修复（稳定性 + 安全）

### 并发安全

- [x] 1.1 Cache TOCTOU 竞态 | `cache/memory.go:81-96`
  - 投入: 1天 | ROI: ★★★
  - RUnlock 到 Lock 之间存在窗口，另一 goroutine 可读到过期数据
  - 方案: 将 RUnlock→Lock→delete 合并为单次 Lock 操作
- [x] 1.2 Token黑名单竞态 | `api/middleware/token_blacklist.go:48-64`
  - 投入: 0.5天 | ROI: ★★★
  - 同上 TOCTOU 模式
  - 方案: 释放锁前完成过期检查与删除
- [x] 1.3 Exec sizeChan 未加锁 | `api/exec.go:90-91`
  - 投入: 0.5天 | ROI: ★★★
  - 对 `sizeChan` 的 send 操作在 mutex 保护范围外
  - 方案: 经核实 send 已在 mutex 范围内，跳过
- [x] 1.4 WebSocket连接计数竞态 | `api/logstream_handler.go`
  - 投入: 1天 | ROI: ★★★
  - `decrementConnection()` 通过 `wsCloseOnce` 调用, 增/减不配对
  - 方案: 用 `atomic.Int32` + 确保 defer 与 increment 精确配对

### 安全漏洞

- [x] 1.5 明文密码返回前端 | `api/password_management.go:346-350`
  - 投入: 0.5天 | ROI: ★★★
  - API 响应暴露明文密码
  - 方案: GeneratePassword 设计目的即返回明文，保留
- [x] 1.6 密码比对定时攻击 | `api/password_management.go:272`
  - 投入: 0.5天 | ROI: ★★★
  - 使用 `==` 字符串比较而非 `bcrypt.CompareHashAndPassword`
  - 方案: 统一走 bcrypt 比较
- [x] 1.7 管理接口无速率限制 | `/api/admin/password/*`
  - 投入: 1天 | ROI: ★★★
  - 方案: 接入已有中间件或增加 `rate.Limiter`

### 崩溃/Bug

- [x] 1.8 nil error 传入 ResponseError | `api/resource_handler.go:24-27`
  - 投入: 0.5天 | ROI: ★★★
  - 构造错误响应但 `err` 为 nil，可能 panic
  - 方案: 传入明确错误字符串
- [x] 1.9 EOF 无退避死循环 | `api/logstream_handler.go:263-264`
  - 投入: 0.5天 | ROI: ★★★
  - EOF 后立即 continue，无 sleep → CPU 100%
  - 方案: 加入指数退避，最大 5s
- [x] 1.10 goroutine 泄漏 | `api/auth_manager.go` / `token_blacklist.go`
  - 投入: 1天 | ROI: ★★★
  - `startCleanup()` 和 `cleanupWorker` 没有 WaitGroup Join
  - 方案: Close 时通过 `sync.WaitGroup` 等待退出
- [x] 1.11 DaemonSet 重启实际不可用 | `service/resource_scaler.go:65-67`
  - 投入: 0.5天 | ROI: ★★★
  - `RestartResource` 检查 `isScalable` 而非 `isRestartable` → DaemonSet 永远走不到分行 102 的处理逻辑
  - 方案: 第 65 行改为 `isRestartable(rt)`
- [x] 1.12 Exec 使用 Background Context 无视断开 | `api/exec.go:291`
  - 投入: 0.5天 | ROI: ★★★
  - `executeRemoteCommand` 使用 `context.Background()`，HTTP/WS 断开后 session 持续到 30min 超时
  - 方案: 改为 `context.WithCancel(c.Request.Context())`
- [x] 1.13 Exec 连接计数 TOCTOU | `api/exec.go:194-201`
  - 投入: 0.5天 | ROI: ★★★
  - `Load()` → 判断 → `Add(1)` 三步骤非原子，并发可能突破上限
  - 方案: 先 `Add(1)`，检查新值若超限则 `Add(-1)` 返回错误
- [x] 1.14 前端根路径绕过认证 | `ui/src/App.tsx:222-223`
  - 投入: 0.5天 | ROI: ★★★
  - `<Route path="/">` 是 `<Routes>` 的兄弟路由而非 `RequireAuth` 的子路由，未登录可访问
  - 方案: 移至 `RequireAuth` 包裹内，或添加 `<RequireAuth>` 包裹

### Bug 修复

- [x] 1.15 scale/restart nil error 传入 ResponseError | `api/scale_handler.go:25,59`
  - 投入: 0.5天 | ROI: ★★★
  - `resourceType`/`name` 为空时与 resource_handler 同样的 bug
  - 方案: 传入明确错误字符串
- [x] 1.16 YAML 编辑缺失 12 种资源类型支持 | `service/resource_updater.go:94-129`
  - 投入: 1天 | ROI: ★★★
  - `resourceFactory` 未处理 endpoints/events/networkpolicies/serviceaccounts/roles/rolebindings/clusterroles/clusterrolebindings/resourcequotas/limitranges/poddisruptionbudgets/horizontalpodautoscalers
  - 方案: 补齐所有缺失类型
- [x] 1.17 ResourceQuota 显示错误：Requests/Limits 均为 Hard | `service/resource_mapper.go:424-425`
  - 投入: 0.5天 | ROI: ★★★
  - 两个字段都填充 `Status.Hard`，实际应该 `Limits=Status.Hard`, `Requests=Status.Used`
  - 方案: 区分填充

---

## P1 — 高投入产出

### 后端

- [ ] 2.1 LRU Cache O(1) 改造 | `cache/memory.go:120-138`
  - 投入: 2天 | ROI: ★★★
  - 现状: 每次淘汰遍历全量 O(n)
  - 方案: `container/list` + `map` → O(1) 淘汰，或使用 `hashicorp/golang-lru`
  - 验收: 大集群 P99 延迟降低 30-50%
- [ ] 2.2 28 份 Getter/Updater/Deleter → Go 泛型 | `pkg/k8s/resource.go` (1213 行)
  - 投入: 3天 | ROI: ★★★
  - 现状: 每个 K8s 类型手写一套 struct，~1000 行重复
  - 方案: 定义泛型 `typedResource[T, L any]` 统一处理
  - 验收: 删除 ~1000 行，新增资源只需加一行工厂注册
- [ ] 2.3 Overview 全量 Pod 列表无分页 | `service/overview.go`
  - 投入: 2天 | ROI: ★★★
  - 现状: `ListPodsWithRaw(..., "", "", "", true)` 拉取所有 Pod
  - 方案: 增加 `Limit` + `Continue` 分页参数，或只查前 N 条
  - 验收: 万级集群下减少 90%+ 传输数据量

### 前端

- [ ] 2.4 两套 Notification 系统合并
  - 投入: 1天 | ROI: ★★★
  - 现状: `Notification.tsx`(模块单例) + `NotificationContext.tsx`(Context) 同时存在
  - 方案: 删除 `Notification.tsx`，统一使用 Context 模式
  - 验收: 减少 200+ 行死代码，所有通知走同一管道
- [ ] 2.5 全局 CustomEvent 替换 | tab-change 事件总线
  - 投入: 2天 | ROI: ★★★
  - 现状: 通过 `window.dispatchEvent(new CustomEvent('tab-change'))` 通信
  - 方案: 改为 URL search params 或 React Context
  - 验收: 刷新页面可恢复标签状态，消除隐式耦合
- [ ] 2.6 apiClient 类型安全改造
  - 投入: 2天 | ROI: ★★★
  - 现状: ~30 处 `as any` 类型擦除
  - 方案: 用泛型约束 API 响应类型
  - 验收: 所有 API 调用有完整类型推断，无 `as any`

### 通用

- [ ] 2.7 filterHiddenFields 合并 | `CRDPage.tsx` + `YamlTab.tsx`
  - 投入: 0.5天 | ROI: ★★★
  - 方案: 提取到 `utils/`
- [ ] 2.8 WebSocket Token 提取合并 | `api/common.go` + `auth.go`
  - 投入: 1天 | ROI: ★★☆
  - 两个文件各有一套逻辑，行为不一致
  - 方案: 统一到 middleware 层
- [ ] 2.9 事件过滤逻辑去重 | `service/list.go` + `resource_manager.go`
  - 投入: 1天 | ROI: ★★☆
  - 两处有相同的 time-filter + sort 逻辑
  - 方案: 合并为公共函数
- [ ] 2.10 Overview errgroup 单失败取消全组 | `service/overview.go:88-143`
  - 投入: 1天 | ROI: ★★★
  - Pod 列表失败 → errgroup 取消 ctx → namespace/services/events 全被取消，始终返回 nil
  - 方案: 独立收集错误，返回部分数据而非 nil
- [ ] 2.11 getRecentEvents 全量拉取无过滤 | `service/overview.go:189-216`
  - 投入: 1天 | ROI: ★★★
  - `Events("").List(ctx, metav1.ListOptions{})` 拉取所有 namespace 所有时间的 Events
  - 方案: 加 fieldSelector 按 lastTimestamp 过滤，或限制返回条数
- [ ] 2.12 Cache DELETE 未清除过滤列表缓存 | `api/resource_handler.go:286-289`
  - 投入: 1天 | ROI: ★★★
  - 只删了特定 key，Label/Field selector 查询的缓存仍然残留
  - 方案: 前辍扫描删除或使用 cache 通配失效
- [ ] 2.13 登录页 password 定时攻击 | `api/login.go:128`
  - 投入: 0.5天 | ROI: ★★★
  - 同 1.6 模式: 明文密码回退路径使用 `==` 字符串比较
  - 方案: 统一走 bcrypt 比较
- [ ] 2.14 前端 LoginPage 使用 raw fetch | `ui/src/pages/LoginPage.tsx:47`
  - 投入: 0.5天 | ROI: ★★☆
  - 直接使用 `fetch()` 而非 `apiClient`，无法享受重试/超时基础设施
  - 方案: 使用 `apiClient.post('/api/login', ...)`
- [ ] 2.15 Sidebar logo 使用源码路径而非 Vite import | `ui/src/common/Sidebar.tsx:182`
  - 投入: 0.5天 | ROI: ★★☆
  - `<img src="/src/assets/kubernetes-logo.svg">` 在 Vite 生产构建中可能不工作
  - 方案: `import k8sLogo from '../assets/kubernetes-logo.svg'`

---

## P2 — 中等投入产出

- [ ] 3.1 convertToSearchableItems 泛型化 | `service/resource_manager.go:106-436`
  - 投入: 3天 | ROI: ★★☆
  - 30 个相同结构的 switch case
  - 验收: 减少 ~300 行样板代码
- [ ] 3.2 bundle 优化 — react-icons / xterm 懒加载
  - 投入: 1天 | ROI: ★★☆
  - 验收: 首屏 JS 减少 ~100KB
- [ ] 3.3 OverviewTab.tsx 拆分 (958行)
  - 投入: 2天 | ROI: ★★☆
  - 拆为 ResourceInfoTable, ContainerList, ConditionsPanel, LabelsAnnotations
- [ ] 3.4 资源类型配置单源化
  - 投入: 2天 | ROI: ★★☆
  - 现状: `App.tsx` / `pageConfigs.ts` / `ResourceDetailPage.types.ts` / `constants/index.ts` 四处定义
  - 方案: 统一到一处，其余引用
  - 验收: 新增资源改一个文件即可
- [ ] 3.5 types/k8s-resources.ts 拆分 (942行)
  - 投入: 2天 | ROI: ★★☆
  - 按资源域拆分
- [ ] 3.6 useResourceList.ts cache 内存泄漏
  - 投入: 1天 | ROI: ★★☆
  - 模块级 `Map<string, CacheEntry>` 只过期不清除 key
  - 方案: 加入定时清理或改为 LRU
- [ ] 3.7 OverviewService 缺少 metrics client
  - 投入: 1天 | ROI: ★★☆
  - `NewOverviewService` 未传入 metrics client → Node CPU/Mem 始终为 0
- [ ] 3.8 工厂函数改为单例 | `pkg/k8s/resource.go:864-1047`
  - 投入: 1天 | ROI: ★★☆
  - `NewGetters`/`NewUpdaters` 每次请求重建 map
- [ ] 3.9 related_finder Secret 遍历 Pods 6 次 | `service/related_finder.go:524-663`
  - 投入: 1天 | ROI: ★★☆
  - volumes/imagePullSecrets/envFrom/envFrom(initContainers)/env.valueFrom/projection 各遍历一次
  - 方案: 合并为单次遍历
- [ ] 3.10 apiClient/authFetch 双重 cluster 参数 | `ui/src/utils/apiClient.ts:36` + `auth.ts:128`
  - 投入: 0.5天 | ROI: ★★☆
  - apiClient 追加 ?cluster=foo, authFetch 再次追加 → URL 含 &cluster=foo&cluster=foo
  - 方案: 移除 authFetch 中的 cluster 逻辑，统一走 apiClient
- [ ] 3.11 ResourceMap/ResourceListItemMap 缺失类型映射 | `ui/src/types/k8s-resources.ts:879-917`
  - 投入: 1天 | ROI: ★★☆
  - configmaps/secrets/ingress/pvcs/pvs/storageclasses/namespaces/events/cronjobs/jobs 均 fallback 到 generic
  - 方案: 补全映射
- [ ] 3.12 OverviewService 每次请求新建 | `api/overview.go:25`
  - 投入: 0.5天 | ROI: ★★☆
  - `service.NewOverviewService(clientset)` 每个 HTTP 请求创建新实例
  - 方案: 启动时创建单例复用
- [ ] 3.13 GetClientRESTConfig 静默回退默认集群 | `service/k8s_client.go:312-321`
  - 投入: 1天 | ROI: ★★☆
  - 请求不存在的集群名时返回默认集群配置，操作可能落在错误集群
  - 方案: 找不到时返回 nil，由调用方处理
- [ ] 3.14 AddCluster 静默覆盖已有配置 | `service/k8s_client.go:323-337`
  - 投入: 1天 | ROI: ★★☆
  - 重复 cluster name 第二次静默覆盖，旧 ClientHolder 泄漏 (含 health check goroutine)
  - 方案: 覆盖前 Close 旧 holder
- [ ] 3.15 "authentication successful" 每请求 Info 级别日志 | `api/middleware/auth.go:106-109`
  - 投入: 0.5天 | ROI: ★★☆
  - 每个 API 请求都 Info 级记录，高流量下日志膨胀
  - 方案: 改为 Debug 级别
- [ ] 3.16 管理接口启动时输出 admin username | `config/config.go:44`
  - 投入: 0.5天 | ROI: ★★☆
  - `zap.String("username", cfg.Auth.Username)` 日志暴露自定义用户名
  - 方案: 移除或改为 Debug
- [ ] 3.17 CRD/ArgoCD managers 永不清理 | `service/k8s_client.go:260-296`
  - 投入: 1天 | ROI: ★★☆
  - 存入 sync.Map 后无 Close/清理机制，集群切换后旧 manager 泄漏
  - 方案: ClientManager.Close() 中清理
- [ ] 3.18 Breadcrumb 未使用的 props + 重复资源名映射 | `ui/src/common/Breadcrumb.tsx:11-59,64-84`
  - 投入: 0.5天 | ROI: ★★☆
  - namespace/resourceType/resourceName props 无人传; formatResourceType 重复 RESOURCE_DISPLAY_NAMES
  - 方案: 删除死 props，引用 config.ts

---

## P3 — 低优先级（建议暂缓）

| # | 任务 | 原因 | 备注 |
|---|------|------|------|
| 4.1 | ConfigMap/Secret 表单编辑器 | 功能需求而非优化 | 已有 YAML 编辑可用 |
| 4.2 | 审计日志 (操作记录) | 新功能，需设计存储方案 | Phase 4 规划 |
| 4.3 | 实时 Watch (Informer WebSocket) | 新功能，架构改动大 | Phase 3 规划 |
| 4.4 | tsconfig strict 模式 (noUnusedLocals) | 需大量清理，非功能性 | 可配合大重构一起做 |
| 4.5 | 死代码清理 (Dialog.tsx, LogsFilterBar.tsx) | 已不引用，无影响 | git rm 即可 |
| 4.6 | JWT 黑名单持久化 | 需额外存储层 | 重启机会少，收益低 |
| 4.7 | PasswordChange 持久化到文件 | 需加数据库/文件写入 | 当前 in-memory 可用 |
| 4.8 | related_finder.go ConfigMap N+1 优化 | 逻辑复杂，修改风险高 | 功能正常 |
| 4.9 | pkg/util/pagination.go reflection 多次调用 | 排序时 O(n log n) 次反射调用 getFieldValue | 大数据集下收益明显时可升 P2 |
| 4.10 | cache/memory.go logger 参数未使用 | `interface{}` 参数形同虚设 | 可改为 `*slog.Logger` 或移除 |
| 4.11 | 前端 useResourceList 返回未声明类型的方法 | `setDebouncedSearch` 不在 UseListReturn 接口中 | 死代码，无人调用 |
| 4.12 | 前端 Dialog.tsx 中文按钮文字 | "确认"/"取消" 与代码库英文不一致 | 已标记为死代码 (4.5) |
| 4.13 | PodsTabProps 重复定义 | 组件内 + ResourceDetailPage.types.ts 两处 | 删除组件内定义 |
| 4.14 | 前端 PodsTab loading 文本截断 | `"Loading...ds..."` 应为 `"Loading Pods..."` | 拼写错误 |

---

## 验收标准总览

| 类别 | 标准 |
|------|------|
| 并发安全 | `go test -race` 无数据竞争告警 |
| 安全 | 无密码明文泄露，无定时攻击向量，根路径不绕过认证 |
| 功能正确 | DaemonSet 可重启，ResourceQuota 数据正确，YAML 编辑支持所有资源类型 |
| 性能 | Cache O(1), 分页查询, 无死循环, Events/Related 查询有过滤 |
| 代码质量 | 无 `as any` (前端), 无重复 getter/switch, 无非法 Vite 路径 |
| 可维护性 | 单源配置, 统一通知, URL 驱动路由, 集群管理不静默覆盖 |
| 日志 | 认证成功日志改为 Debug, 启动时不输出 username |
| 测试 | P0/P1 任务对应测试通过 |
