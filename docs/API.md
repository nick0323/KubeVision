# KubeVision API 文档

**Base URL**: `http://localhost:8080/api`

## 目录

- [认证](#认证)
- [资源操作](#资源操作)
- [集群管理](#集群管理)
- [密码管理](#密码管理)
- [YAML 操作](#yaml-操作)
- [日志流](#日志流)
- [终端访问](#终端访问)
- [关联资源](#关联资源)
- [监控指标](#监控指标)
- [缓存统计](#缓存统计)
- [健康检查](#健康检查)

---

## 认证

所有 API（除登录外）需在 Header 携带 JWT Token：

```
Authorization: Bearer <token>
```

### 登录

**`POST /api/login`**

```json
// Request
{ "username": "admin", "password": "your-password" }

// Response 200
{ "code": 200, "message": "Login successful", "data": { "token": "eyJ..." }, "timestamp": ... }

// Response 401
{ "code": 401, "message": "Invalid username or password", "details": { "remainingAttempts": 3, "maxFailCount": 5 } }
```

---

## 资源操作

所有资源操作通过 `ResourceEntry` 注册表中的统一接口处理，无需为每种资源单独写 handler。

### 获取资源列表

**`GET /api/:resourceType`**

查询参数：

| 参数 | 类型 | 说明 |
|------|------|------|
| `namespace` | string | 命名空间 |
| `cluster` | string | 集群名称（默认集群可不传） |
| `labelSelector` | string | Label 选择器 |
| `fieldSelector` | string | Field 选择器 |
| `search` | string | 搜索关键词 |
| `offset` | int | 分页偏移 |
| `limit` | int | 每页数量 |

**示例**: `GET /api/pods?namespace=default&cluster=dev`

### 获取资源详情

**`GET /api/:resourceType/:namespace/:name`**

**示例**: `GET /api/deployments/default/nginx`

### 删除资源

**`DELETE /api/:resourceType/:namespace/:name`**

### 扩缩容

**`POST /api/:resourceType/:namespace/:name/scale`**

```json
// Request
{ "replicas": 5 }
```

### 重启 (Deployment/StatefulSet/DaemonSet)

**`POST /api/:resourceType/:namespace/:name/restart`**

无需请求体。

---

## 集群管理

### 获取集群列表

**`GET /api/v1/clusters`**

```json
// Response 200
{ "code": 200, "data": [
  { "name": "default", "apiServer": "https://...", "version": "1.28", "healthy": true, "nodeCount": 3, "lastCheck": 1700000000 },
  { "name": "dev", "apiServer": "https://...", "version": "1.27", "healthy": true, "nodeCount": 5, "lastCheck": 1700000000 }
]}
```

### 获取集群健康状态

**`GET /api/v1/clusters/health`**

```json
// Response 200
{ "code": 200, "data": [
  { "name": "default", "healthy": true, "host": "https://...", "version": "1.28", "nodeCount": 3 },
  { "name": "dev", "healthy": true, "host": "https://...", "version": "1.27", "nodeCount": 5 }
]}
```

### 测试集群连接

**`POST /api/v1/clusters/test`**

```json
// Request
{ "apiServer": "https://...", "token": "...", "kubeconfig": "/path/to/config", "caFile": "...", "insecure": false }
```

### 添加集群

**`POST /api/v1/clusters`**

```json
// Request
{ "name": "prod", "apiServer": "https://...", "token": "...", "kubeconfig": "...", "caFile": "...", "insecure": false }
```

注意：名称不能为 `"default"` 或空字符串。

### 删除集群

**`DELETE /api/v1/clusters/:name`**

注意：无法删除 `"default"` 集群。

---

## 密码管理

### 修改密码

**`POST /api/v1/admin/password/change`**

```json
// Request
{ "oldPassword": "current-password", "newPassword": "new-password" }
```

密码强度要求：

| 规则 | 要求 |
|------|------|
| 长度 | 8-128 字符 |
| 字符多样性 | ≥3 种：大写 / 小写 / 数字 / 特殊符号 |
| 连续数字 | 禁止 3+ 位连续递增（如 `1234`） |
| 重复字符 | 禁止单字符占比 >50% |
| 弱密码 | 禁止常见密码（password、admin、123456 等） |
| 历史 | 不能与最近 5 次密码相同 |

### 生成随机密码

**`POST /api/v1/admin/password/generate`**

```json
// Request
{ "length": 16 }

// Response 200
{ "code": 200, "data": { "password": "aB3$xY9!mN2@pQ7#", "hashedPassword": "$2a$12$...", "length": 16 } }
```

### 验证密码哈希

**`POST /api/v1/admin/password/validate`**

```json
// Request
{ "password": "plain-text", "hashedPassword": "$2a$12$..." }

// Response 200
{ "code": 200, "data": { "valid": true } }
```

### 密码哈希

**`POST /api/v1/admin/password/hash`**

```json
// Request
{ "password": "plain-text" }

// Response 200
{ "code": 200, "data": { "hashedPassword": "$2a$12$...", "cost": 12 } }
```

---

## YAML 操作

### 获取资源 YAML

**`GET /api/:resourceType/:namespace/:name/yaml`**

### 更新资源 YAML

**`PUT /api/:resourceType/:namespace/:name/yaml`**

```json
// Request (两种格式均可)
{ "apiVersion": "apps/v1", "kind": "Deployment", "metadata": { "name": "nginx", "resourceVersion": "12345" }, "spec": { "replicas": 5 } }

// 或嵌套格式
{ "yaml": { "apiVersion": "apps/v1", ... } }
```

注意：必须包含 `resourceVersion`。

---

## 日志流

### WebSocket Pod 日志

**`WS /ws/logs`**

```
ws://localhost:8080/ws/logs?namespace=default&pod=nginx&container=app&tailLines=100&token=<jwt>
```

服务端消息格式：

```json
{ "type": "connected", "message": "Connected to default/nginx (app)" }
{ "type": "log", "content": "2024-01-01 INFO Application started" }
{ "type": "heartbeat" }
{ "type": "error", "message": "Pod not found" }
```

---

## 终端访问

### WebSocket Pod Exec

**`WS /ws/exec`**

```
ws://localhost:8080/ws/exec?namespace=default&pod=nginx&container=app&command=/bin/bash&token=<jwt>
```

客户端 → 服务端：

```json
{ "type": "stdin", "data": "ls -la\n" }
{ "type": "resize", "cols": 120, "rows": 40 }
```

服务端 → 客户端：

```json
{ "type": "stdout", "data": "total 0\ndrwxr-xr-x ..." }
{ "type": "connected", "namespace": "default", "pod": "nginx", "container": "app" }
```

---

## 关联资源

### 获取关联资源

**`GET /api/:resourceType/:namespace/:name/related`**

```json
// Response 200
{ "code": 200, "data": [
  { "kind": "ReplicaSet", "name": "nginx-abc123", "relation": "child" },
  { "kind": "Service", "name": "nginx-svc", "relation": "exposedBy" }
]}
```

支持 16 种资源的关联查找（Pod → owner/Service/volumes/node, Deployment → RS/Service/HPA/PDB/Ingress 等）。

---

## 监控指标

### 获取所有指标

**`GET /api/v1/metrics`**

### 缓存统计

**`GET /api/v1/cache/stats`**

```json
// Response
{ "code": 200, "data": { "size": 150, "maxSize": 1000, "hits": 5000, "misses": 500, "hitRate": 90.9, "evictions": 50 } }
```

---

## 健康检查

**`GET /health`**

```json
// Response 200
{ "status": "healthy", "k8sConnected": true }
```

---

## 速率限制

| 端点 | 限制 |
|------|------|
| 通用 API | 100 req/s (可配置) |
| 登录 | 5 次失败后锁定 10 分钟 |
| WebSocket | 最大 100 并发 |

超出返回 `429 Too Many Requests`。

## 错误码

| 状态码 | 说明 |
|--------|------|
| 200 | 成功 |
| 400 | 参数错误 |
| 401 | 未认证 |
| 403 | 无权限 |
| 404 | 资源不存在 |
| 409 | 冲突 |
| 429 | 限流 |
| 500 | 服务端错误 |

```json
// 错误响应格式
{ "code": 400, "message": "...", "details": "...", "traceId": "...", "timestamp": ... }
```
