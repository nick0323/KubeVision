# KubeVision

Kubernetes Web 管理面板 - 可视化和管理 K8s 集群资源

[![Go Version](https://img.shields.io/badge/Go-1.24-blue)](https://golang.org)
[![Kubernetes](https://img.shields.io/badge/Kubernetes-1.25+-blue)](https://kubernetes.io)
[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](LICENSE)

## 📖 简介

KubeVision 是一个功能完整的 Kubernetes Web 管理面板，提供直观的 UI 界面来管理和监控 K8s 集群资源。支持多集群管理、实时日志、终端访问、YAML 编辑等高级功能。

### 核心特性

- **资源管理**: 支持 15+ 种 Kubernetes 资源类型
  - 工作负载：Pods, Deployments, StatefulSets, DaemonSets, Jobs, CronJobs
  - 服务网络：Services, Ingresses, Endpoints
  - 配置管理：ConfigMaps, Secrets
  - 存储：PVCs, PVs, StorageClasses
  - 集群：Namespaces, Nodes, Events

- **实时操作**:
  - 📝 YAML 查看和编辑
  - 📊 实时 Pod 日志流 (WebSocket)
  - 💻 Pod 终端访问 (WebSocket + xterm.js)
  - 🔗 关联资源查询

- **企业级功能**:
  - 🔐 JWT 认证 + 登录失败锁定
  - 🚀 LRU 缓存优化
  - 📈 监控指标采集
  - 🔄 多集群支持
  - ⚡ 优雅关闭

---

## 🏗️ 技术架构

### 后端技术栈

| 组件 | 技术 | 版本 |
|------|------|------|
| 语言 | Go | 1.24 |
| Web 框架 | Gin | v1.10+ |
| 日志 | zap | v1.27+ |
| WebSocket | gorilla/websocket | v1.5+ |
| JWT | golang-jwt/jwt | v5+ |
| 配置 | Viper | v1.19+ |
| K8s 客户端 | client-go | v0.33+ |
| 密码加密 | bcrypt | - |
| 缓存 | 内存 LRU | 自研 |

### 前端技术栈

| 组件 | 技术 | 版本 |
|------|------|------|
| 框架 | React | 19 |
| 语言 | TypeScript | 5.9 |
| 构建 | Vite | 7 |
| 终端 | xterm.js | v5+ |
| HTTP | Axios | v1.6+ |

### 系统架构

```
┌─────────────┐     ┌──────────────┐     ┌─────────────┐
│   Browser   │────▶│  KubeVision  │────▶│  Kubernetes │
│  (React UI) │     │   (Go API)   │     │   Cluster   │
└─────────────┘     └──────────────┘     └─────────────┘
       │                    │                    │
       │                    │                    │
       ▼                    ▼                    ▼
  WebSocket           JWT Auth            K8s API Server
  xterm.js            LRU Cache           Informers
                      Metrics             Pod Exec
```

---

## 📦 快速开始

### 前置要求

- Go 1.24+
- Node.js 18+
- Kubernetes 集群访问权限 (1.25+)

### 1. 克隆项目

```bash
git clone https://github.com/nick0323/K8sVision.git
cd KubeVision
```

### 2. 后端启动

```bash
# 安装依赖
go mod download

# 设置环境变量
export K8SVISION_JWT_SECRET="your-secret-key-at-least-32-chars"
export K8SVISION_AUTH_USERNAME="admin"
export K8SVISION_AUTH_PASSWORD="your-password"

# 方式 1: 直接运行
go run main.go

# 方式 2: 编译后运行
go build -o KubeVision
./KubeVision -config config.yaml

# 方式 3: Docker 运行
docker build -t k8svision:latest .
docker run -p 8080:8080 \
  -e K8SVISION_JWT_SECRET="your-secret" \
  -e K8SVISION_AUTH_USERNAME="admin" \
  -e K8SVISION_AUTH_PASSWORD="your-password" \
  k8svision:latest
```

### 3. 前端启动

```bash
cd ui

# 安装依赖
npm install

# 开发模式
npm run dev

# 生产构建
npm run build
```

### 4. 访问应用

打开浏览器访问：`http://localhost:8080`

默认账号：
- 用户名：`admin`
- 密码：启动时自动生成或配置

---

## ⚙️ 配置说明

### 环境变量

| 变量名 | 说明 | 默认值 | 必填 |
|--------|------|--------|------|
| `K8SVISION_SERVER_PORT` | 服务端口 | `8080` | 否 |
| `K8SVISION_SERVER_HOST` | 服务主机 | `0.0.0.0` | 否 |
| `K8SVISION_JWT_SECRET` | JWT 密钥 | 自动生成 | 是* |
| `K8SVISION_AUTH_USERNAME` | 管理员用户名 | `admin` | 是* |
| `K8SVISION_AUTH_PASSWORD` | 管理员密码 | 自动生成 | 是* |
| `K8SVISION_LOG_LEVEL` | 日志级别 | `info` | 否 |
| `KUBECONFIG` | K8s 配置文件路径 | `~/.kube/config` | 否 |

*首次启动可自动生成，生产环境建议显式配置

### 配置文件 (config.yaml)

```yaml
server:
  port: "8080"
  host: "0.0.0.0"
  allowedOrigin:
    - "http://localhost:3000"
    - "http://localhost:8080"

kubernetes:
  kubeconfig: "~/.kube/config"
  timeout: 30s
  qps: 100
  burst: 200

jwt:
  secret: "your-secret-key-at-least-32-characters-long"
  expiration: 24h

log:
  level: "info"
  format: "json"

auth:
  username: "admin"
  password: "$2a$12$..."  # bcrypt 哈希后的密码
  maxLoginFail: 5
  lockDuration: 10m
  sessionTimeout: 24h
  enableRateLimit: true
  rateLimit: 100

cache:
  enabled: true
  type: "memory"
  ttl: 5m
  maxSize: 1000
  cleanupInterval: 10m
```

---

## 📚 API 文档

### 认证接口

#### 登录
```http
POST /api/login
Content-Type: application/json

{
  "username": "admin",
  "password": "your-password"
}

Response:
{
  "code": 200,
  "message": "登录成功",
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
  }
}
```

### 资源接口

#### 获取资源列表
```http
GET /api/:resourceType?namespace=default&labelSelector=&fieldSelector=
Authorization: Bearer <token>
```

#### 获取资源详情
```http
GET /api/:resourceType/:namespace/:name
Authorization: Bearer <token>
```

#### 获取资源 YAML
```http
GET /api/:resourceType/:namespace/:name/yaml
Authorization: Bearer <token>
```

#### 更新资源 YAML
```http
PUT /api/:resourceType/:namespace/:name/yaml
Authorization: Bearer <token>
Content-Type: application/json

{
  "yaml": {
    "apiVersion": "v1",
    "kind": "Pod",
    ...
  }
}
```

#### 删除资源
```http
DELETE /api/:resourceType/:namespace/:name
Authorization: Bearer <token>
```

### WebSocket 接口

#### Pod 日志流
```javascript
const ws = new WebSocket('ws://localhost:8080/ws/logs/pod?namespace=default&pod=my-pod&token=<jwt-token>');

ws.onmessage = (event) => {
  const data = JSON.parse(event.data);
  if (data.type === 'log') {
    console.log(data.content);
  }
};
```

#### Pod 终端
```javascript
const ws = new WebSocket('ws://localhost:8080/ws/exec?namespace=default&pod=my-pod&container=app&token=<jwt-token>');

// 发送命令
ws.send(JSON.stringify({
  type: 'stdin',
  data: 'ls -la\n'
}));

// 接收输出
ws.onmessage = (event) => {
  const data = JSON.parse(event.data);
  if (data.type === 'stdout') {
    console.log(data.data);
  }
};
```

---

## 🔒 安全说明

### JWT 配置

JWT Secret 必须满足：
- 长度至少 32 字符
- 建议包含大小写字母、数字、特殊字符
- 生产环境必须通过环境变量配置

```bash
# 生成安全的 JWT Secret
openssl rand -base64 32
```

### 密码策略

- 最小长度：8 字符
- 最大长度：128 字符
- 复杂度要求：至少 3 种字符类型（大写、小写、数字、特殊字符）
- 历史记录：不使用最近 5 次用过的密码

### 登录保护

- 最大失败次数：5 次
- 锁定时长：10 分钟
- 速率限制：100 请求/秒

---

## 📊 监控指标

### 内置指标

访问 `/metrics` 获取：

```json
{
  "system": {
    "cpu": { "usage_percent": 25.5, "cores": 4 },
    "memory": { "used_mb": 512, "total_mb": 2048, "usage_percent": 25.0 }
  },
  "business": {
    "totalRequests": 10000,
    "cacheHitRate": 85.5,
    "k8sApiCalls": 5000
  }
}
```

### 健康检查

```http
GET /health

Response:
{
  "status": "healthy",
  "timestamp": 1234567890,
  "k8sConnected": true
}
```

---

## 🛠️ 开发指南

### 项目结构

```
KubeVision/
├── api/                    # API 层 (HTTP handlers)
│   ├── middleware/        # 中间件 (JWT, CORS, Logging)
│   ├── exec.go            # Pod Exec WebSocket
│   ├── login.go           # 登录接口
│   ├── operations.go      # 资源操作 (YAML, 关联资源)
│   └── resource_handler.go # 通用资源接口
├── cache/                  # 缓存层 (LRU, TTL)
├── config/                 # 配置管理 (Viper)
├── model/                  # 数据模型
├── monitor/                # 监控指标
├── service/                # 业务逻辑层
│   ├── client_manager.go  # K8s 客户端管理
│   ├── list.go            # 资源列表查询
│   └── informer.go        # K8s Informer
├── ui/                     # 前端 React 应用
├── main.go                 # 入口文件
├── config.yaml             # 配置文件
└── Dockerfile              # Docker 构建
```

### 添加新资源类型

1. 在 `service/list.go` 添加 List 函数
2. 在 `api/resource_registry.go` 注册 Handler
3. 在前端 `ui/src/types/k8s.ts` 添加类型定义

### 测试

```bash
# 单元测试
go test ./...

# 覆盖率
go test ./... -coverprofile=coverage.out
go tool cover -html=coverage.out

# 竞态检测
go test ./... -race
```

---

## 🚀 部署指南

### Kubernetes 部署

```yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: k8svision
  namespace: kube-system
spec:
  replicas: 2
  selector:
    matchLabels:
      app: k8svision
  template:
    metadata:
      labels:
        app: k8svision
    spec:
      serviceAccountName: k8svision-sa
      containers:
      - name: k8svision
        image: k8svision:latest
        ports:
        - containerPort: 8080
        env:
        - name: K8SVISION_JWT_SECRET
          valueFrom:
            secretKeyRef:
              name: k8svision-secret
              key: jwt-secret
        - name: K8SVISION_AUTH_USERNAME
          value: "admin"
        - name: K8SVISION_AUTH_PASSWORD
          valueFrom:
            secretKeyRef:
              name: k8svision-secret
              key: admin-password
---
apiVersion: v1
kind: Service
metadata:
  name: k8svision
  namespace: kube-system
spec:
  selector:
    app: k8svision
  ports:
  - port: 80
    targetPort: 8080
  type: ClusterIP
```

### ServiceAccount 配置

```yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  name: k8svision-sa
  namespace: kube-system
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: k8svision-role
rules:
- apiGroups: [""]
  resources: ["*"]
  verbs: ["get", "list", "watch", "create", "update", "delete"]
- apiGroups: ["apps"]
  resources: ["*"]
  verbs: ["get", "list", "watch", "create", "update", "delete"]
# ... 添加其他 API 组权限
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRoleBinding
metadata:
  name: k8svision-role-binding
roleRef:
  apiGroup: rbac.authorization.k8s.io
  kind: ClusterRole
  name: k8svision-role
subjects:
- kind: ServiceAccount
  name: k8svision-sa
  namespace: kube-system
```

---

## ❓ 常见问题

### Q: 首次启动密码是多少？
A: 如果未配置密码，系统会自动生成并记录在日志中。查看启动日志获取临时密码。

### Q: 如何重置密码？
A: 调用 `/api/admin/password/change` 接口修改密码（需要登录）。

### Q: 支持哪些 K8s 版本？
A: 支持 Kubernetes 1.25+ 版本。

### Q: 如何配置多集群？
A: 当前版本支持单集群，多集群功能开发中。

### Q: WebSocket 连接失败？
A: 检查：
1. JWT token 是否有效
2. 防火墙是否允许 WebSocket
3. Pod 是否运行中
4. 容器名称是否正确

---

## 📝 更新日志

### v2.0.0 (2026-04)
- ✨ 重构资源处理器为 Registry 模式
- ✨ 中文消息全部翻译为英文
- 🐛 修复多处 Bug
- 📈 性能优化

### v1.0.0 (2025-12)
- 🎉 首次发布
- ✅ 基础资源管理功能
- ✅ WebSocket 日志和终端

---

## 🤝 贡献指南

欢迎提交 Issue 和 Pull Request！

1. Fork 项目
2. 创建特性分支 (`git checkout -b feature/AmazingFeature`)
3. 提交更改 (`git commit -m 'Add some AmazingFeature'`)
4. 推送到分支 (`git push origin feature/AmazingFeature`)
5. 开启 Pull Request

---

## 📄 许可证

Apache License 2.0 - 详见 [LICENSE](LICENSE) 文件

---

## 📬 联系方式

- 项目地址：https://github.com/nick0323/K8sVision
- 问题反馈：https://github.com/nick0323/K8sVision/issues
