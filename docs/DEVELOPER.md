# KubeVision 开发者指南

本文档面向 KubeVision 的开发者，介绍项目架构、编码规范和开发流程。

## 目录

- [技术栈](#技术栈)
- [项目结构](#项目结构)
- [开发环境](#开发环境)
- [代码规范](#代码规范)
- [测试指南](#测试指南)
- [添加新资源类型](#添加新资源类型)
- [常见问题](#常见问题)

---

## 技术栈

### 后端

| 组件 | 技术 | 用途 |
|------|------|------|
| 语言 | Go 1.26 | 主要编程语言 |
| Web 框架 | Gin v1.10+ | HTTP 路由和中间件 |
| K8s 客户端 | client-go v0.33+ | Kubernetes API 访问 |
| WebSocket | gorilla/websocket v1.5+ | Pod exec / 日志流 |
| JWT | golang-jwt/jwt v5+ | Token 认证 |
| 日志 | zap v1.27+ | 结构化日志 |
| 配置 | Viper v1.19+ | 配置管理 |

### 前端

| 组件 | 技术 | 用途 |
|------|------|------|
| 框架 | React 19 | UI 框架 |
| 语言 | TypeScript 5.9 | 类型系统 |
| 构建 | Vite 7 | 构建工具 |
| 终端 | xterm.js v5+ | 终端模拟 |
| HTTP | fetch (封装) | HTTP 客户端 |

---

## 项目结构

```
KubeVision/
├── api/                        # API 层 (HTTP handlers)
│   ├── middleware/            # 中间件
│   │   ├── auth.go           # JWT 认证
│   │   ├── cors.go           # CORS
│   │   ├── error.go          # 错误处理
│   │   ├── logging.go        # 日志中间件
│   │   └── metrics.go        # 监控中间件
│   ├── cluster_handler.go    # 集群管理接口
│   ├── exec.go               # Pod Exec WebSocket
│   ├── login.go              # 登录接口
│   ├── metrics_handler.go    # K8s Metrics handler
│   ├── operations.go         # YAML 操作、关联资源、日志流
│   ├── overview.go           # 集群概览
│   ├── password_management.go # 密码管理
│   ├── related_handler.go    # 关联资源接口
│   ├── resource_handler.go   # 通用资源 CRUD 路由
│   └── yaml_handler.go       # YAML 编解码
│
├── bootstrap/                  # 启动初始化
│   └── initializer.go         # 默认数据初始化
│
├── cache/                      # 缓存层
│   ├── memory.go              # 泛型 LRU 缓存 (16 shards)
│   └── memory_test.go
│
├── config/                     # 配置管理
│   ├── manager.go             # 配置管理器 (Viper)
│   └── manager_test.go
│
├── model/                      # 数据模型
│   ├── config.go              # 配置结构体
│   ├── consts.go              # 常量定义
│   ├── types.go               # API 类型定义
│   └── config_test.go
│
├── monitor/                    # 监控指标
│   ├── metrics.go
│   └── metrics_test.go
│
├── pkg/k8s/                    # K8s 工具包
│   └── resource.go            # ResourceEntry 注册表 (核心)
│
├── service/                    # 业务逻辑层
│   ├── cluster_service.go     # 集群增删改查
│   ├── client_manager.go      # K8s 客户端管理 (多集群)
│   ├── informer.go            # Pod Informer 缓存
│   ├── k8s_client.go          # 客户端创建与健康检查
│   ├── related_finder.go      # 关联资源查找 (16 种资源)
│   ├── resource_manager.go    # 资源列表/搜索
│   ├── resource_mapper.go     # K8s 对象 → 前端模型映射
│   ├── resource_updater.go    # 资源更新
│   └── resource_scaler.go     # 资源扩缩容
│
├── ui/                         # 前端 React 应用
│   ├── src/
│   │   ├── common/            # 通用组件
│   │   │   ├── Sidebar.tsx    # 侧栏容器
│   │   │   ├── SidebarMenu.tsx# 导航菜单
│   │   │   ├── ClusterSelector.tsx # 集群选择器
│   │   │   ├── SettingsModal.tsx   # 设置弹窗
│   │   │   ├── Table.tsx      # 通用表格 (React.memo)
│   │   │   ├── StatusBadge.tsx# 状态标签
│   │   │   └── ...            # 其他组件
│   │   ├── pages/             # 页面
│   │   ├── tabs/              # 详情页 Tab
│   │   ├── hooks/             # 自定义 Hooks
│   │   ├── utils/             # 工具函数
│   │   ├── styles/            # 全局样式 + CSS 变量
│   │   └── constants/         # 常量配置
│   └── package.json
│
├── docs/                       # 文档
│   ├── API.md
│   ├── ARCHITECTURE.md
│   └── DEVELOPER.md
│
├── cmd/tools/                  # 命令行工具
│   └── generate_password.go   # 密码生成工具
│
├── main.go                     # 应用入口
├── config.yaml                 # 配置文件
├── Dockerfile
├── go.mod
└── Makefile
```

---

## 开发环境

### 1. 安装依赖

```bash
# Go 依赖
go mod download

# 前端依赖
cd ui && npm install
```

### 2. 配置环境

```bash
cp .env.example .env
# 编辑 .env，设置 JWT_SECRET (64+ chars) 和 AUTH_PASSWORD (bcrypt hash)
```

### 3. 启动

```bash
# 后端 (默认 :8080)
go run main.go

# 前端 (默认 :5173，代理到 :8080)
cd ui && npm run dev
```

### 4. 运行测试

```bash
# 后端
go test ./... -race

# 前端
cd ui && npm run test
```

### 5. 代码检查

```bash
go fmt ./... && go vet ./...
cd ui && npm run lint && npm run type-check
```

---

## 代码规范

### Go

| 规范 | 要求 |
|------|------|
| 命名 | 类型大驼峰，变量小驼峰，常量大驼峰 |
| 错误 | `fmt.Errorf("context: %w", err)` 包装，zap 结构化日志 |
| 并发 | context 控制超时，sync.Map 管理连接池 |
| 导入 | 标准库 → 第三方 → 内部，每组空行分隔 |

### 前端

| 规范 | 要求 |
|------|------|
| 类型 | Props 用 interface，API 响应用泛型 |
| 组件 | 函数组件 + Hooks，纯展示组件加 `React.memo` |
| CSS | 使用 CSS 变量（`var(--spacing-md)`），kebab-class 命名 |
| 状态 | 业务状态用 `apiClient` + hooks，UI 状态用 `useState` |
| 集群参数 | `apiClient` 自动注入，无需手动传 `cluster` |

---

## 测试指南

### 后端测试模式

```bash
# 单元测试
go test ./api/... -v

# 竞态检测
go test ./service/... -race

# 覆盖率
go test ./... -coverprofile=coverage.out && go tool cover -html=coverage.out
```

### 前端测试

```bash
cd ui && npm run test
```

---

## 添加新资源类型

### 后端

在 `pkg/k8s/resource.go` 的 `NewRegistry()` 中添加一条记录：

```go
ResourceYourType: {
    Kind:          "YourType",
    ClusterScoped: false,
    Get:           func(c kubernetes.Interface, ns, name string) (any, error) {
        return c.AppsV1().YourTypes(ns).Get(ctx, name, metav1.GetOptions{})
    },
    List:          func(c kubernetes.Interface, ns string, opts metav1.ListOptions) (any, error) {
        return c.AppsV1().YourTypes(ns).List(ctx, opts)
    },
    Create:        func(c kubernetes.Interface, ns string, obj any) error { ... },
    Update:        func(c kubernetes.Interface, ns, name string, obj any) error { ... },
    Delete:        func(c kubernetes.Interface, ns, name string) error { ... },
},
```

然后在 `service/resource_manager.go` 的 `searchItemMappers()` 中添加映射函数。

### 前端

1. 在 `ui/src/constants/index.ts` 添加资源类型常量
2. 在 `ui/src/constants/pageConfigs.ts` 添加列表/详情页配置
3. 在 `ui/src/tabs/` 添加详情 Tab 组件（如需）

---

## 常见问题

**Q: WebSocket 连接数有限制吗？**
A: `api/operations.go` 中用 `atomic.Int32` 计数，最大 100 并发。

**Q: 如何调优 K8s API 性能？**
A: LRU 缓存（5min TTL 列表，2min 详情）+ Informer 缓存 + QPS/Burst 配置。

**Q: 密码强度规则是什么？**
A: ≥8 字符，≤128 字符，至少含 3 种字符类型（大写/小写/数字/特殊字符），不能有 3+ 连续数字，不能是常见弱密码。
