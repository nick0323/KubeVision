# KubeVision

Kubernetes Web 管理面板 - 可视化和管理 K8s 集群资源

## 功能特性

### 核心功能
- **资源管理**: 支持 15+ 种 Kubernetes 资源类型的列表和详情查看
  - Pods, Deployments, StatefulSets, DaemonSets
  - Services, Ingresses, Endpoints
  - ConfigMaps, Secrets
  - Jobs, CronJobs
  - PVCs, PVs, StorageClasses
  - Namespaces, Nodes, Events

- **YAML 编辑**: 查看和更新资源 YAML 配置
- **实时日志**: WebSocket 实时查看 Pod 日志（支持 Follow 模式）
- **终端访问**: WebSocket Pod Exec（基于 xterm.js）
- **关联资源**: 查询 Pod/Deployment 等资源的关联关系
- **多集群支持**: 支持配置和管理多个 K8s 集群

### 高级功能
- **JWT 认证**: 安全的登录系统，失败锁定机制
- **缓存优化**: LRU 缓存减少 K8s API 调用
- **监控指标**: 请求统计、缓存命中率、K8s API 调用统计
- **优雅关闭**: 支持 signal 处理和 context 超时

## 技术栈

### 后端
- **语言**: Go 1.24
- **Web框架**: Gin
- **日志**: zap
- **WebSocket**: gorilla/websocket
- **认证**: golang-jwt/jwt
- **配置**: Viper
- **K8s**: client-go 0.33

### 前端
- **框架**: React 19 + TypeScript 5.9
- **构建**: Vite 7
- **终端**: xterm.js
- **样式**: 原生 CSS

## 快速开始

### 前置要求
- Go 1.24+
- Node.js 18+
- Kubernetes 集群访问权限

### 后端启动

1. 配置 `config.yaml`（可选，默认使用 ~/.kube/config）
2. 设置环境变量（推荐）:
   ```bash
   export K8SVISION_JWT_SECRET="your-random-secret-at-least-32-chars"
   export K8SVISION_AUTH_USERNAME="admin"
   export K8SVISION_AUTH_PASSWORD="your-secure-password"
   ```
3. 启动服务:
   ```bash
   go run main.go
   ```

### 前端启动

```bash
cd ui
npm install
npm run dev
```

### Docker 部署

```bash
docker build -t kubevision:latest .
docker run -d -p 8080:8080 \
  -e K8SVISION_JWT_SECRET="your-secret" \
  -e K8SVISION_AUTH_PASSWORD="your-password" \
  -v ~/.kube/config:/root/.kube/config \
  kubevision:latest
```

## 配置说明

### 配置文件结构

```yaml
server:
  port: "8080"
  host: "0.0.0.0"

kubernetes:
  kubeconfig: ""  # 空则使用默认 ~/.kube/config
  timeout: "30s"
  qps: 100
  burst: 200

jwt:
  secret: ""  # 必须设置！至少32位随机字符
  expiration: "24h"

auth:
  username: "admin"
  password: ""  # 首次启动会自动生成并显示在日志中
  maxLoginFail: 5
  lockDuration: "10m"

cache:
  enabled: true
  ttl: "5m"
  maxSize: 1000

# 多集群配置（可选）
clusters:
  - name: "cluster-dev"
    kubeconfig: "/path/to/dev-kubeconfig"
  - name: "cluster-prod"
    apiServer: "https://prod-k8s-api:6443"
    token: "your-token"
```

### 环境变量优先级

环境变量 > 配置文件 > 默认值

主要环境变量：
- `K8SVISION_JWT_SECRET`: JWT 密钥（必须）
- `K8SVISION_AUTH_USERNAME`: 管理员用户名
- `K8SVISION_AUTH_PASSWORD`: 管理员密码
- `KUBECONFIG`: K8s 配置文件路径

## 架构设计

### 分层架构
```
┌─────────────────┐
│   API Layer     │  ← HTTP handlers, WebSocket
├─────────────────┤
│  Service Layer  │  ← Business logic, K8s operations
├─────────────────┤
│   Model Layer   │  ← Data types, Config
└─────────────────┘
         ↓
┌─────────────────┐
│   Cache Layer   │  ← LRU memory cache
│   Monitor Layer │  ← Metrics, Tracing
└─────────────────┘
```

### 缓存策略
- **列表查询**: 5 分钟 TTL
- **资源详情**: 2 分钟 TTL
- **LRU 淘汰**: 超过 maxSize 时淘汰最少访问的项
- **自动失效**: 删除/更新资源时清除相关缓存

### 多集群架构
```
                  ┌─────────────┐
                  │ClientManager│
                  └──────┬──────┘
                         │
        ┌────────────────┼────────────────┐
        ↓                ↓                ↓
  ┌──────────┐   ┌──────────┐   ┌──────────┐
  │ Default  │   │Cluster A │   │Cluster B │
  │ Client   │   │ Client   │   │ Client   │
  └──────────┘   └──────────┘   └──────────┘
```

## API 文档

### 认证
- `POST /api/login` - 用户登录

### 资源操作
- `GET /api/:resourceType` - 获取资源列表
- `GET /api/:resourceType/:namespace/:name` - 获取资源详情
- `DELETE /api/:resourceType/:namespace/:name` - 删除资源
- `PUT /api/:resourceType/:namespace/:name/yaml` - 更新资源 YAML

### WebSocket
- `/api/ws/exec` - Pod 终端
- `/api/ws/stream` - Pod 日志流

### 其他
- `GET /api/overview` - 集群概览
- `GET /health` - 健康检查
- `GET /cache/stats` - 缓存统计

## 开发指南

### 添加新资源类型

1. 在 `service/list.go` 中添加列表函数
2. 在 `api/resource_handler.go` 的 switch-case 中添加 case
3. 前端添加对应路由和页面

### 测试

```bash
# 运行所有测试
go test ./...

# 运行特定包测试
go test ./cache -v
go test ./service -v
```

### 性能优化建议

1. **缓存利用**: 确保缓存已启用并合理配置 TTL
2. **连接池**: K8s 客户端使用连接池，避免重复创建
3. **Informer**: 高频查询资源可使用 Informer 缓存
4. **分页**: 列表接口默认分页，避免大数据量传输

## 安全建议

1. **JWT Secret**: 使用强随机密钥（至少 32 位）
2. **密码管理**: 定期修改管理员密码
3. **RBAC**: 配置 K8s ServiceAccount 限制权限
4. **HTTPS**: 生产环境启用 TLS
5. **网络策略**: 限制 API 访问来源

## 更新日志

### v2.0.0-optimized (2026-04-14)

**新增功能**
- ✨ API 缓存机制（列表 5min TTL，详情 2min TTL）
- ✨ 多集群动态加载支持
- ✨ PodInformer 缓存优化
- ✨ 安全配置自动检查和生成

**优化改进**
- 🔧 修复 WebSocket 硬编码 localhost 问题
- 🔧 优化 LRU 缓存淘汰逻辑
- 🔧 增强 JWT 和密码安全管理
- 🔧 添加核心模块单元测试

**已知问题**
- 详情见 GitHub Issues

## 许可证

MIT License

## 贡献

欢迎提交 Issue 和 Pull Request！
