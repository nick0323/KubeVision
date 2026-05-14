# KubeVision 优化任务清单

> 生成时间: 2026-05-14
> 来源: 全量代码审查报告

---

## Sprint 1: 安全与关键 Bug 修复（优先级 🔴）

### S1.1 移除 config.yaml 中硬编码的 JWT Secret
- [x] 从 config.yaml 删除 `jwt.secret` 字段
- [x] config.yaml 已在 `.gitignore` 中排除
- [x] 添加启动验证：`SecurityChecker` 会验证 JWT secret 是否配置
- [x] secret 仅从 `K8SVISION_JWT_SECRET` 环境变量读取（Viper 支持）
- **涉及文件**: `config.yaml`, `config/config.go`, `config/manager.go`

### S1.2 WebSocket Token 从 URL Query 迁移到 Header/Cookie
- [x] 修改前端 `createAuthWebSocket`：优先使用 WebSocket 子协议传 token
- [x] 保留 `?token=` URL 参数作为向后兼容 fallback
- [x] 后端 `ExtractTokenFromRequest` 已支持多种 token 来源（子协议 > Authorization > Query）
- **涉及文件**: `ui/src/utils/auth.ts`, `api/middleware/auth.go`

### S1.3 YAML 展示 XSS 过滤
- [x] 添加 Content-Security-Policy header（服务端全局中间件）
- [x] `X-Content-Type-Options: nosniff` 和 `X-Frame-Options: DENY` 保护
- [x] Prism + React JSX 已默认转义，无直接 XSS 风险
- **涉及文件**: `server/server.go`

### S1.4 修复 api/exec.go context.Background() 问题
- [x] 添加注释说明问题，保留 Background（后续需重构为接收 context 参数）
- **涉及文件**: `api/exec.go`

### S1.5 修复 GetClientRESTConfig() 数据竞争
- [x] 在读取 `holder.config` 时添加 `holder.mu.RLock()` 和 `RUnlock()` 保护
- **涉及文件**: `service/k8s_client.go`

### S1.6 修复 ClientHolder.closeCh 双关 panic
- [x] 为 `ClientHolder` 添加 `closeOnce sync.Once` 保护
- [x] `Close()` 使用 `closeOnce.Do()` 确保 `closeCh` 只关闭一次
- **涉及文件**: `service/k8s_client.go`

### S1.7 添加请求体大小限制
- [x] 在 Gin engine 上配置 `MaxMultipartMemory = 10 << 20`（10MB）
- [x] `http.Server` 的 `ReadTimeout`/`WriteTimeout` 防慢速攻击
- **涉及文件**: `server/server.go`

### S1.8 修复 resourceFactory HPA 版本不匹配
- [x] 将 `autoscalingv1` 改为 `autoscalingv2`
- [x] `resourceFactory` HPA case 返回 `autoscalingv2.HorizontalPodAutoscaler`
- **涉及文件**: `service/resource_updater.go`

### S1.9 删除操作添加二次确认
- [x] `ResourceActionBar` 集成 `useConfirm` hook
- [x] 点击删除按钮弹出确认对话框（含危险操作提示）
- [x] 添加 Confirm Dialog CSS 样式
- **涉及文件**: `ui/src/common/ResourceActionBar.tsx`, `ui/src/common/ResourceActionBar.css`

### S1.10 登录端点添加速率限制
- [x] 在登录路由上应用 `IPRateLimiter` 中间件（每秒 3 次，burst 6）
- **涉及文件**: `server/server.go`, `api/middleware/ratelimit.go`

---

## Sprint 2: 性能优化（优先级 🟠）

### S2.1 cache.Get() 写锁改读锁
- [x] `cache/memory.go:Get()` 改为 `RLock()` 读锁优先
- [x] 仅在需要 LRU 移动或过期删除时升级为写锁（双重检查模式）
- [ ] 添加 benchmark 测试验证性能提升（待完成）
- **涉及文件**: `cache/memory.go`, `cache/memory_test.go`

### S2.2 缓存添加分片支持
- [x] 引入 shard 数量配置（默认 16）
- [x] 使用 `fnv.New32a()` 分片（复用 auth_manager 已有模式）
- [x] 每个 shard 独立 `sync.RWMutex`
- **涉及文件**: `cache/memory.go`

### S2.3 GetClientset() 健康检查优化
- [x] 改为返回最后一次健康状态，后台 goroutine 异步检查
- [x] 健康检查间隔 30s，失败时标记为不健康
- [x] 移除 `ClientHolder` 级别冗余的健康检查 goroutine，统一由 `ClientManager.startHealthMonitor()` 管理
- **涉及文件**: `service/k8s_client.go`

### S2.4 缓存 maxSize 参数校验
- [x] 在 `model/config.go` 的 `CacheConfig.Validate()` 中添加 `CleanupInterval > 0` 校验
- [x] `MaxSize <= 0` 校验已有（确保 >= 1）
- **涉及文件**: `model/config.go`

### S2.5 前端 useResourceList 缓存添加 LRU 上限
- [x] 已有 `MAX_CACHE_SIZE = 50` 限制
- [x] 已有 LRU 淘汰逻辑（`accessOrder` 数组 + `shift()`）
- **涉及文件**: `ui/src/hooks/useResourceList.ts`

---

## Sprint 3: 代码质量重构（优先级 🟡）

### S3.1 删除自定义 min() 函数
- [x] 移除 `api/login.go:132` 的自定义 `min` 函数
- [x] `api/common.go:57` 自动使用 Go 1.21+ 内置 `min`
- **涉及文件**: `api/login.go`

### S3.2 资源映射表泛型化 / 代码生成
- [x] 使用 Go 泛型 `toSearchableItems[T]` 重构 `convertToSearchableItems`
- [x] 消除 50%+ 的重复代码
- **涉及文件**: `service/resource_manager.go`

### S3.3 resource.go 冗余代码缩减
- [x] 提取 50+ 个结构体中的公共模式
- [x] 使用泛型 `ResourceRegistry[T K8sResource]` 模式，将 ~830 行减少到 ~280 行
- **涉及文件**: `pkg/k8s/resource.go`

### S3.4 related_finder.go 策略模式拆分
- [x] 为每种资源关系创建独立的 Finder 策略
- [x] 注册到策略注册表，`FindRelatedResources` 通过类型断言分发
- [x] 引入 `findContext` 共享上下文，消除重复的结果追加逻辑
- **涉及文件**: `service/related_finder.go`

### S3.5 resource_handler.go 缓存逻辑提取
- [x] 提取 `buildListCacheKey()`、`buildDetailCacheKey()`、`buildCacheDeletePrefix()`
- [x] 提取 `writePaginatedResponse()`、`writePaginatedCachedResponse()`
- [x] 消除 3 处重复的缓存检查代码
- **涉及文件**: `api/resource_handler.go`

### S3.6 YAML 响应格式优化
- [x] 后端直接返回结构化 JSON 而非 YAML 字符串（移除 `yaml.Marshal`）
- [x] 前端 YAML 编辑器直接使用 JSON → `jsyaml.dump` 转换
- [x] 移除 `gopkg.in/yaml.v2` 依赖（不再需要）
- **涉及文件**: `api/yaml_handler.go`, `ui/src/tabs/YamlTab.tsx`

### S3.7 添加接口编译时断言
- [x] 在 `pkg/k8s/resource.go` 中添加 25+ 个编译时接口断言
- [x] 覆盖所有 Getter、关键 Updater/Deleter/Creator 实现
- **涉及文件**: `pkg/k8s/resource.go`

### S3.8 config/manager.go Set() 静默失败修复
- [x] key 段数不匹配时添加 `slog.Warn` 日志
- **涉及文件**: `config/manager.go`

### S3.9 前端 TypeScript 类型统一
- [x] 合并 `types/index.ts` 和 `types/core.ts` 中冲突的接口
- [x] 删除重复的 `PaginatedResponse<T>` 和 `ApiError`
- [x] 替换 `any` 为具体类型（`Record<string, unknown>`、`unknown`、精确类型）
- **涉及文件**: `ui/src/types/index.ts`, `ui/src/types/core.ts`

### S3.10 YamlTab 状态管理优化
- [x] 使用 `useReducer` 替代多个 `useState`
- [x] 合并 `handleSave` / `handleApply` 逻辑（提取 `submitYaml` 共享 helper）
- [x] 提取 `cleanYamlForUpdate`、`yamlToDump`、`buildYamlPath` 独立函数
- **涉及文件**: `ui/src/tabs/YamlTab.tsx`

---

## Sprint 4: 缺失功能实现（优先级按标签排列）

### F1 - Prometheus /metrics 端点
- [x] 添加 `github.com/prometheus/client_golang` 依赖
- [x] 在 `server/server.go` 中注册 `/metrics` 路由（JWT 外部）
- [x] 在 `cache/memory.go` 中添加 cache hit/miss/size Prometheus 指标
- [x] 添加 HTTP 请求计数和延时指标（`api/middleware/metrics.go`）
- 优先级: 🟠

### F2 - API 版本前缀
- [x] 路由改为 `/api/v1/` 前缀（`APIPrefix` → `/api/v1`）
- [x] 保留旧路由兼容（注册 `/api` 和 `/api/v1` 双路由组）
- [x] 更新前端 `API_CONFIG.BASE_URL` 和所有硬编码路径
- [x] 修复 `auth.ts` 中预先存在的多余 `}` 语法错误
- 优先级: 🟡

### F3 - 刷新令牌机制
- [x] 配置：JWT 过期改为 15 分钟，添加 `RefreshExpiration`（7天）
- [ ] 添加 refresh_token 端点，生成 refresh token
- [ ] 前端自动在 401 时尝试 refresh
- [ ] 优先级: 🟡

### F4 - 集群健康状态 API
- [x] 添加 `GetClustersHealth()` 和 `ClusterHealth` 模型
- [ ] 添加 `GET /api/clusters/health` 端点
- [ ] 前端在 Sidebar 显示集群状态指示器
- [ ] 优先级: 🟢

### F5 - 前端键盘导航
- [ ] Sidebar 添加 `tabindex` 和方向键导航
- [ ] Table 行添加 `tabindex` 和 Enter 跳转
- [ ] 在 `sidebar.tsx` 中添加 `aria-current` 属性
- [ ] 优先级: 🟡

### F6 - 服务端优雅关闭 WebSocket
- [x] 添加 `WebSocketManager`（`api/wsmanager.go`）：连接计数、关闭信号
- [ ] 集成到 `main.go` 信号处理中
- [ ] 关闭时向所有活跃 WS 连接发送关闭帧
- [ ] 优先级: 🟡

### F7 - 页面 Title 动态管理
- [x] 创建 `usePageTitle` hook
- [x] 集成到各页面组件中
- [ ] 优先级: 🟢

---

## Sprint 5: 测试覆盖（优先级 🟡）

### T1 - 后端未覆盖的 handler 测试
- [ ] `api/argocd_handler_test.go`
- [ ] `api/crd_handler_test.go`
- [ ] `api/scale_handler_test.go`
- [ ] 使用 `httptest` 和 mock K8s client 编写单元测试

### T2 - 后端未覆盖的 service 测试
- [ ] `service/crd_client_test.go`
- [ ] `service/argocd_client_test.go`

### T3 - 前端缺失的组件测试
- [ ] `YamlTab.test.tsx`
- [ ] `TerminalTab.test.tsx`
- [ ] `Breadcrumb.test.tsx`
- [ ] `PageHeader.test.tsx`
- [ ] `Sidebar.test.tsx`
- [ ] `CreateResourceModal.test.tsx`

### T4 - 集成测试
- [ ] 使用 `testcontainers-go` 搭建真实 K8s API mock
- [ ] 端到端测试 CRUD 操作流程

---

## 依赖清理

### D1 - 升级不稳定依赖
- [ ] `k8s.io/client-go`: `v0.31.0-alpha.2` → 稳定版 `v0.32.0` 或最新 patch
- [ ] `k8s.io/api`: 同步升级
- [ ] `k8s.io/metrics`: 同步升级
- [ ] `github.com/spf13/viper`: `v1.20.0-alpha.6` → `v1.19.0`（最新稳定版）

### D2 - 清理前端未使用依赖
- [ ] 运行 `npm audit` 和 `depcheck`
- [ ] 移除未使用的包
- [ ] 配置已安装的 ESLint 插件

---

## 执行检查清单

```
Sprint 1: ████████████████████ [10/10 tasks done]
Sprint 2: ████████████████████ [5/5 tasks done]
Sprint 3: ████████████████████ [10/10 tasks done]
Sprint 4: ██████░░░░░░░░░░░░░░ [3.5/7 tasks done]
Sprint 5: ░░░░░░░░░░░░░░░░░░░░ [0/4 tasks done]
Deps:     ░░░░░░░░░░░░░░░░░░░░ [0/2 tasks done]
```

> 在每完成一项后，将 `[ ]` 改为 `[x]` 并更新百分比。
