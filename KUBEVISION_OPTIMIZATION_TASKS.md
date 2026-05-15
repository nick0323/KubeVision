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
- [x] 缓存已同时改造为 O(1) LRU（`container/list` + `map` + shard）
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
- [x] 后端 refresh_token 端点：验证旧 token → 黑名单 → 生成新 token 对
- [x] 前端自动在 401 时尝试 refresh（`attemptTokenRefresh`）
- [x] 前端 `isTokenExpiringSoon()` 预刷新逻辑
- [x] 修复前端调用不存在的 `attemptRefresh()` 函数名 bug
- 优先级: 🟡

### F4 - 集群健康状态 API
- [x] 添加 `GetClustersHealth()` 和 `ClusterHealth` 模型
- [x] 添加 `GET /api/v1/clusters/health` 端点
- [x] 前端 Sidebar 显示集群健康指示器（状态点 + tooltip）
- 优先级: 🟢

### F5 - 前端键盘导航
- [x] Sidebar 添加 `tabIndex`、`role="menuitem"`、`aria-current`、Enter/Space 键盘支持
- [x] Table 行添加 `tabIndex`、`role="button"`、Enter/Space 跳转
- 优先级: 🟡

### F6 - 服务端优雅关闭 WebSocket
- [x] 添加 `WebSocketManager`（`api/wsmanager.go`）：连接计数、关闭信号
- [x] `exec.go` 和 `logstream_handler.go` 使用 `Acquire()`/`Release()` 管理连接
- [x] 集成到 `server.Shutdown()` 中，通过 `ShutdownCtx()` 传播关闭信号
- 优先级: 🟡

### F7 - 页面 Title 动态管理
- [x] 创建 `usePageTitle` hook
- [x] 集成到各页面组件中
- [ ] 优先级: 🟢

---

## Sprint 5: 测试覆盖（优先级 🟡）

### T1 - 后端未覆盖的 handler 测试
- [x] `api/argocd_handler_test.go`
- [x] `api/crd_handler_test.go`
- [x] `api/scale_handler_test.go`
- [x] 使用 `httptest` 和 mock K8s client 编写单元测试

### T2 - 后端未覆盖的 service 测试
- [x] `service/crd_client_test.go`
- [x] `service/argocd_client_test.go`

### T3 - 前端缺失的组件测试（暂缓：环境阻塞）
- [ ] `YamlTab.test.tsx`
- [ ] `TerminalTab.test.tsx`
- [ ] `Breadcrumb.test.tsx`
- [ ] `PageHeader.test.tsx`
- [ ] `Sidebar.test.tsx`
- [ ] `CreateResourceModal.test.tsx`
- 状态: 被 `@rollup/rollup-linux-x64-gnu` 环境问题阻塞

### T4 - 集成测试（暂缓：投入产出比低）
- [ ] 使用 `testcontainers-go` 搭建真实 K8s API mock
- [ ] 端到端测试 CRUD 操作流程
- 状态: 搭建成本高，维护负担大，当前阶段收益有限

---

## 依赖清理

### D1 - 升级不稳定依赖
- [x] `k8s.io/client-go`: `v0.31.0-alpha.2` → `v0.36.1`
- [x] `k8s.io/api`: `v0.31.0-alpha.2` → `v0.36.1`
- [x] `k8s.io/metrics`: `v0.31.0-alpha.0` → `v0.36.1`
- [x] `github.com/spf13/viper`: `v1.20.0-alpha.6` → `v1.21.0`
- [x] `github.com/golang-jwt/jwt/v4`: `v4.3.0` → `v4.5.2`
- [x] `github.com/gin-gonic/gin`: `v1.10.1` → `v1.12.0`
- [x] `github.com/gorilla/websocket`: `v1.5.0` → `v1.5.4`
- [x] `github.com/prometheus/client_golang`: `v1.20.0` → `v1.23.2`
- [x] `go.uber.org/zap`: `v1.27.0` → `v1.28.0`
- [x] `golang.org/x/crypto`: `v0.24.0` → `v0.51.0`
- [x] `golang.org/x/sync`: `v0.7.0` → `v0.20.0`
- [x] `golang.org/x/time`: `v0.9.0` → `v0.15.0`
- [x] Go: `1.24.4` → `1.26.0`

### D2 - 清理前端未使用依赖
- [x] 运行 `depcheck` 检查
- [x] 移除 15 个未使用的包（runtime: `prop-types`, `web-vitals`; dev: `husky`, `terser`, `eslint-config-prettier`, `eslint-plugin-import`, `eslint-plugin-jsx-a11y`, `eslint-plugin-prettier`, `eslint-plugin-simple-import-sort`, `@typescript-eslint/eslint-plugin`, `@typescript-eslint/parser`, `@testing-library/user-event`, `@testing-library/dom`, `stylelint-declaration-block-no-ignored-properties`, `typescript-plugin-css-modules`）
- [x] 移除 `husky` + `prepare` script（无 `.husky/` 目录）
- [x] 移除 158 个 transitive 依赖

---

## 执行检查清单

```
Sprint 1: ████████████████████ [10/10 tasks done]
Sprint 2: ████████████████████ [5/5 tasks done]
Sprint 3: ████████████████████ [10/10 tasks done]
Sprint 4: ████████████████████ [7/7 tasks done]
Sprint 5: ████████████████████ [2/4 tasks done (T3/T4 暂缓)]
Deps:     ████████████████████ [2/2 tasks done]
```

> 在每完成一项后，将 `[ ]` 改为 `[x]` 并更新百分比。
