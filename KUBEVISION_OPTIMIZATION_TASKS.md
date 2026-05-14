# KubeVision 优化任务清单

> 生成时间: 2026-05-14
> 来源: 全量代码审查报告

---

## Sprint 1: 安全与关键 Bug 修复（优先级 🔴）

### S1.1 移除 config.yaml 中硬编码的 JWT Secret
- [ ] 从 config.yaml 删除 `jwt.secret` 字段
- [ ] 更新 `.gitignore` 排除 config.yaml（或确保 secret 仅从环境变量读取）
- [ ] 在 config/manager.go 中添加说明文档：secret 应通过 `K8SVISION_JWT_SECRET` 环境变量注入
- [ ] 验证启动时若缺少 JWT secret 给出明确错误信息
- **涉及文件**: `config.yaml`, `.gitignore`, `config/manager.go`, `config/config.go`

### S1.2 WebSocket Token 从 URL Query 迁移到 Header/Cookie
- [ ] 修改 `api/common.go:InitWebSocketUpgrader` 支持从 Header 或 Cookie 读取 token
- [ ] 修改前端 WebSocket 连接逻辑：token 不再拼入 URL 参数
- [ ] 保留现有 `?token=` 作为 fallback（需在文档中标明废弃）
- **涉及文件**: `api/common.go`, `ui/src/utils/auth.ts`, `ui/src/hooks/useResourceList.ts`

### S1.3 YAML 展示 XSS 过滤
- [ ] 在 `ui/src/tabs/YamlTab.tsx` 中引入 `DOMPurify` 或对 YAML 内容进行 HTML 实体转义
- [ ] 添加 content-security-policy header（服务端）
- **涉及文件**: `ui/src/tabs/YamlTab.tsx`, `server/server.go`

### S1.4 修复 api/exec.go context.Background() 问题
- [ ] 将 `api/exec.go:233` 的 `context.Background()` 替换为传入的请求 context
- [ ] 验证 context 取消后 pod exec 连接正确关闭
- **涉及文件**: `api/exec.go`

### S1.5 修复 GetClientRESTConfig() 数据竞争
- [ ] 在 `service/k8s_client.go` 中为 `GetClientRESTConfig()` 添加 `holder.mu.RLock()` / `RLock()`
- [ ] 确保所有对 `holder.config` 的读取都受 mutex 保护
- **涉及文件**: `service/k8s_client.go`

### S1.6 修复 ClientHolder.closeCh 双关 panic
- [ ] 为 `ClientHolder.closeCh` 添加 `sync.Once` 保护
- [ ] 添加单元测试验证重复调用 Close() 不 panic
- **涉及文件**: `service/k8s_client.go`, `service/k8s_client_test.go`

### S1.7 添加请求体大小限制
- [ ] 在 `server/server.go` 中配置 Gin engine: `engine.MaxMultipartMemory = 10 << 20` 或全局 body 限制
- [ ] 考虑使用中间件对特定端点（YAML apply、create）应用更严格限制
- **涉及文件**: `server/server.go`

### S1.8 修复 resourceFactory HPA 版本不匹配
- [ ] 统一 `autoscaling` 版本为 v2
- [ ] 修改 `service/resource_updater.go:resourceFactory` 中 HPA 创建逻辑
- **涉及文件**: `service/resource_updater.go`, `pkg/k8s/resource.go`

### S1.9 删除操作添加二次确认
- [ ] 前端 `ResourceActionBar.tsx` 中 DELETE 操作前弹出确认对话框
- [ ] 使用已有的 `useConfirm` hook
- **涉及文件**: `ui/src/components/ResourceActionBar.tsx`

### S1.10 登录端点添加速率限制
- [ ] 在 `api/login.go` 的 login 路由上应用 rate-limit 中间件
- [ ] 配置每秒最多 3 次尝试
- **涉及文件**: `api/login.go`, `api/middleware/ratelimit.go`, `api/operations.go`

---

## Sprint 2: 性能优化（优先级 🟠）

### S2.1 cache.Get() 写锁改读锁
- [ ] `cache/memory.go:Get()` 中 `Lock()` → `RLock()`
- [ ] 仅当需要做 LRU 移动时才升级为写锁（或使用 `sync.Map` + 原子操作）
- [ ] 添加 benchmark 测试验证性能提升
- **涉及文件**: `cache/memory.go`, `cache/memory_test.go`

### S2.2 缓存添加分片支持
- [ ] 引入 shard 数量配置（默认 16）
- [ ] 使用 `fnv.New32a()` 分片（复用 auth_manager 已有模式）
- [ ] 每个 shard 独立 `sync.RWMutex`
- **涉及文件**: `cache/memory.go`

### S2.3 GetClientset() 健康检查优化
- [ ] 改为返回最后一次健康状态，后台 goroutine 异步检查
- [ ] 健康检查间隔 30s，失败时标记为不健康
- **涉及文件**: `service/k8s_client.go`

### S2.4 缓存 maxSize 参数校验
- [ ] 在 `model/config.go` 的 `CacheConfig` 中添加校验：`maxSize >= 1`
- [ ] 启动时打印警告日志
- **涉及文件**: `model/config.go`, `bootstrap/initializer.go`

### S2.5 前端 useResourceList 缓存添加 LRU 上限
- [ ] 设置最大缓存条目数（默认 50）
- [ ] 超出时淘汰最久未使用的条目
- **涉及文件**: `ui/src/hooks/useResourceList.ts`

---

## Sprint 3: 代码质量重构（优先级 🟡）

### S3.1 删除自定义 min() 函数
- [ ] 移除 `api/login.go:132` 的自定义 `min`
- [ ] 全局替换为 Go 内置 `min`
- **涉及文件**: `api/login.go`

### S3.2 资源映射表泛型化 / 代码生成
- [ ] 评估选项：
  - a) 使用 Go 1.24 泛型重构 `convertToSearchableItems`
  - b) 使用 `go generate` + 模板生成 switch 分支
- [ ] 目标：减少 `resource_mapper.go` 中 50% 的重复代码
- **涉及文件**: `service/resource_mapper.go`, `service/resource_manager.go`

### S3.3 resource.go 冗余代码缩减
- [ ] 提取 50+ 个结构体中的公共模式
- [ ] 考虑用泛型 `ResourceRegistry[T K8sResource]` 模式
- **涉及文件**: `pkg/k8s/resource.go`

### S3.4 related_finder.go 策略模式拆分
- [ ] 为每种资源关系创建独立的 Finder 策略
- [ ] 注册到策略注册表，主函数做分发
- **涉及文件**: `service/related_finder.go`

### S3.5 resource_handler.go 缓存逻辑提取
- [ ] 提取 `buildCacheKey()`、`getFromCache()`、`setToCache()` 为公共方法
- [ ] 消除 3 处重复的缓存检查代码
- **涉及文件**: `api/resource_handler.go`

### S3.6 YAML 响应格式优化
- [ ] 后端直接返回结构化 JSON 而非 YAML 字符串
- [ ] 前端 YAML 编辑器负责 JSON ↔ YAML 转换（使用 `yaml.js`）
- **涉及文件**: `api/yaml_handler.go`, `ui/src/tabs/YamlTab.tsx`

### S3.7 添加接口编译时断言
- [ ] 在 `pkg/k8s/resource.go` 中添加 `var _ Getter = &podsGetter{}` 等
- [ ] 覆盖所有 Getter/Deleter/Updater/Creator 实现
- **涉及文件**: `pkg/k8s/resource.go`

### S3.8 config/manager.go Set() 静默失败修复
- [ ] key 段数不匹配时添加 `slog.Warn` 日志
- **涉及文件**: `config/manager.go`

### S3.9 前端 TypeScript 类型统一
- [ ] 合并 `types/index.ts` 和 `types/core.ts` 中冲突的接口
- [ ] 删除重复的 `PaginatedResponse<T>` 和 `ApiError`
- [ ] 逐步替换 `any` 为具体类型
- **涉及文件**: `ui/src/types/index.ts`, `ui/src/types/core.ts`

### S3.10 YamlTab 状态管理优化
- [ ] 使用 `useReducer` 替代多个 `useState`
- [ ] 合并 `handleSave` / `handleApply` 逻辑
- **涉及文件**: `ui/src/tabs/YamlTab.tsx`

---

## Sprint 4: 缺失功能实现（优先级按标签排列）

### F1 - Prometheus /metrics 端点
- [ ] 添加 `github.com/prometheus/client_golang` 依赖
- [ ] 在 `server/server.go` 中注册 `/metrics` 路由
- [ ] 在 `cache/memory.go` 中添加 cache hit/miss 指标
- [ ] 添加 HTTP 请求计数和延时指标
- [ ] 优先级: 🟠

### F2 - API 版本前缀
- [ ] 路由改为 `/api/v1/` 前缀
- [ ] 保留旧路由兼容（通过重定向或双注册）
- [ ] 更新前端 apiClient baseURL
- [ ] 优先级: 🟡

### F3 - 刷新令牌机制
- [ ] 添加 refresh_token 端点，生成 7 天有效期的 refresh token
- [ ] 前端自动在 401 时尝试 refresh
- [ ] JWT 过期时间改为 15 分钟 + refresh token 机制
- [ ] 优先级: 🟡

### F4 - 集群健康状态 API
- [ ] 添加 `GET /api/clusters/health` 端点
- [ ] 返回每个集群的连接状态、API 版本、node 数量
- [ ] 前端在 Sidebar 显示集群状态指示器
- [ ] 优先级: 🟢

### F5 - 前端键盘导航
- [ ] Sidebar 添加 `tabindex` 和方向键导航
- [ ] Table 行添加 `tabindex` 和 Enter 跳转
- [ ] 在 `sidebar.tsx` 中添加 `aria-current` 属性
- [ ] 优先级: 🟡

### F6 - 服务端优雅关闭 WebSocket
- [ ] 在 `main.go` 信号处理中添加活跃 WebSocket 连接跟踪
- [ ] 关闭时向所有活跃 WS 连接发送关闭帧
- [ ] 设置 5 秒优雅关闭超时
- [ ] 优先级: 🟡

### F7 - 页面 Title 动态管理
- [ ] 创建 `usePageTitle` hook
- [ ] 在路由级根据当前页面设置 HTML title
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
Sprint 1: ████████░░░░░░░░░░░░ [8/10 tasks done]
Sprint 2: ░░░░░░░░░░░░░░░░░░░░ [0/5 tasks done]
Sprint 3: ░░░░░░░░░░░░░░░░░░░░ [0/10 tasks done]
Sprint 4: ░░░░░░░░░░░░░░░░░░░░ [0/7 tasks done]
Sprint 5: ░░░░░░░░░░░░░░░░░░░░ [0/4 tasks done]
Deps:     ░░░░░░░░░░░░░░░░░░░░ [0/2 tasks done]
```

> 在每完成一项后，将 `[ ]` 改为 `[x]` 并更新百分比。
