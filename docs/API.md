# KubeVision API 文档

完整的 API 接口文档，包含请求示例和响应格式。

**版本**: v2.0.0  
**Base URL**: `http://localhost:8080/api`

## 目录

- [认证](#认证)
- [资源操作](#资源操作)
- [YAML 操作](#yaml 操作)
- [日志流](#日志流)
- [终端访问](#终端访问)
- [管理接口](#管理接口)
- [监控指标](#监控指标)
- [错误码](#错误码)

---

## 认证

所有 API 请求（除登录外）都需要在 Header 中携带 JWT Token：

```http
Authorization: Bearer <token>
```

### 登录

**接口**: `POST /api/login`

**请求**:
```json
{
  "username": "admin",
  "password": "your-password"
}
```

**响应**:
```json
{
  "code": 200,
  "message": "Login successful",
  "data": {
    "token": "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9..."
  },
  "timestamp": 1234567890
}
```

**错误响应**:
```json
{
  "code": 401,
  "message": "Invalid username or password",
  "details": {
    "remainingAttempts": 3,
    "maxFailCount": 5
  },
  "timestamp": 1234567890
}
```

---

## 资源操作

### 获取资源列表

**接口**: `GET /api/:resourceType`

**路径参数**:
- `resourceType`: 资源类型 (pods, deployments, services, etc.)

**查询参数**:
| 参数 | 类型 | 说明 |
|------|------|------|
| `namespace` | string | 命名空间（集群资源不需要） |
| `labelSelector` | string | Label 选择器 |
| `fieldSelector` | string | Field 选择器 |
| `offset` | int | 分页偏移量 |
| `limit` | int | 每页数量 (默认 20, 最大 100) |
| `search` | string | 搜索关键词 |
| `sortBy` | string | 排序字段 |
| `sortOrder` | string | 排序方式 (asc/desc) |

**示例**:
```http
GET /api/pods?namespace=default&labelSelector=app=nginx&offset=0&limit=20
Authorization: Bearer <token>
```

**响应**:
```json
{
  "code": 200,
  "message": "List retrieved successfully",
  "data": [
    {
      "namespace": "default",
      "name": "nginx-deployment-abc123",
      "status": "Running",
      "ready": "1/1",
      "restarts": 0,
      "age": "2d",
      "podIP": "10.244.0.5",
      "nodeName": "node-1"
    }
  ],
  "page": {
    "total": 50,
    "limit": 20,
    "offset": 0
  },
  "timestamp": 1234567890
}
```

### 获取资源详情

**接口**: `GET /api/:resourceType/:namespace/:name`

**示例**:
```http
GET /api/deployments/default/nginx-deployment
Authorization: Bearer <token>
```

**响应**:
```json
{
  "code": 200,
  "message": "Resource details retrieved successfully",
  "data": {
    "namespace": "default",
    "name": "nginx-deployment",
    "readyReplicas": 3,
    "updatedReplicas": 3,
    "available": 3,
    "desiredReplicas": 3,
    "status": "Available",
    "age": "30d"
  },
  "timestamp": 1234567890
}
```

### 删除资源

**接口**: `DELETE /api/:resourceType/:namespace/:name`

**示例**:
```http
DELETE /api/pods/default/nginx-pod
Authorization: Bearer <token>
```

**响应**:
```json
{
  "code": 200,
  "message": "Resource deleted successfully",
  "data": null,
  "timestamp": 1234567890
}
```

---

## YAML 操作

### 获取资源 YAML

**接口**: `GET /api/:resourceType/:namespace/:name/yaml`

**示例**:
```http
GET /api/deployments/default/nginx-deployment/yaml
Authorization: Bearer <token>
```

**响应**:
```json
{
  "code": 200,
  "message": "YAML retrieved successfully",
  "data": "apiVersion: apps/v1\nkind: Deployment\nmetadata:\n  name: nginx-deployment\n  namespace: default\nspec:\n  replicas: 3\n  selector:\n    matchLabels:\n      app: nginx\n  template:\n    metadata:\n      labels:\n        app: nginx\n    spec:\n      containers:\n      - name: nginx\n        image: nginx:1.21\n        ports:\n        - containerPort: 80",
  "timestamp": 1234567890
}
```

### 更新资源 YAML

**接口**: `PUT /api/:resourceType/:namespace/:name/yaml`

**请求体格式 1** (直接 YAML):
```json
{
  "apiVersion": "apps/v1",
  "kind": "Deployment",
  "metadata": {
    "name": "nginx-deployment",
    "namespace": "default",
    "resourceVersion": "12345"
  },
  "spec": {
    "replicas": 5
  }
}
```

**请求体格式 2** (嵌套 yaml 字段):
```json
{
  "yaml": {
    "apiVersion": "apps/v1",
    "kind": "Deployment",
    "metadata": {
      "name": "nginx-deployment",
      "namespace": "default",
      "resourceVersion": "12345"
    },
    "spec": {
      "replicas": 5
    }
  }
}
```

**响应**:
```json
{
  "code": 200,
  "message": "Resource updated successfully",
  "data": null,
  "timestamp": 1234567890
}
```

**注意**:
- 必须包含 `resourceVersion` 字段
- Event 资源不支持 YAML 更新

---

## 日志流

### WebSocket Pod 日志

**接口**: `WS /ws/logs/pod`

**连接 URL**:
```
ws://localhost:8080/ws/logs/pod?namespace=default&pod=my-pod&container=app&tailLines=100&timestamps=true&token=<jwt-token>
```

**查询参数**:
| 参数 | 类型 | 说明 |
|------|------|------|
| `namespace` | string | 命名空间（必填） |
| `pod` | string | Pod 名称（必填） |
| `container` | string | 容器名称 |
| `tailLines` | int | 日志行数 |
| `timestamps` | bool | 显示时间戳 |
| `previous` | bool | 获取之前容器的日志 |
| `token` | string | JWT Token |

**消息格式**:

服务端 → 客户端:
```json
{
  "type": "connected",
  "message": "Connected to default/my-pod (app)"
}

{
  "type": "log",
  "content": "2024-01-01T00:00:00Z INFO Application started"
}

{
  "type": "heartbeat"
}

{
  "type": "error",
  "message": "Pod not found"
}
```

**JavaScript 示例**:
```javascript
const token = 'eyJhbGc...';
const ws = new WebSocket(
  `ws://localhost:8080/ws/logs/pod?namespace=default&pod=my-pod&token=${token}`
);

ws.onopen = () => {
  console.log('Connected to log stream');
};

ws.onmessage = (event) => {
  const data = JSON.parse(event.data);
  switch (data.type) {
    case 'connected':
      console.log(data.message);
      break;
    case 'log':
      console.log(data.content);
      break;
    case 'heartbeat':
      // 心跳响应
      break;
    case 'error':
      console.error(data.message);
      break;
  }
};

ws.onerror = (error) => {
  console.error('WebSocket error:', error);
};

ws.onclose = () => {
  console.log('Connection closed');
};
```

---

## 终端访问

### WebSocket Pod Exec

**接口**: `WS /ws/exec`

**连接 URL**:
```
ws://localhost:8080/ws/exec?namespace=default&pod=my-pod&container=app&command=/bin/bash&token=<jwt-token>
```

**查询参数**:
| 参数 | 类型 | 说明 |
|------|------|------|
| `namespace` | string | 命名空间（必填） |
| `pod` | string | Pod 名称（必填） |
| `container` | string | 容器名称 |
| `command` | string | 执行的命令（默认 /bin/sh） |
| `token` | string | JWT Token |

**消息格式**:

客户端 → 服务端:
```json
{
  "type": "stdin",
  "data": "ls -la\n"
}

{
  "type": "resize",
  "cols": 120,
  "rows": 40
}
```

服务端 → 客户端:
```json
{
  "type": "stdout",
  "data": "total 0\ndrwxr-xr-x ..."
}

{
  "type": "stderr",
  "data": "error message"
}

{
  "type": "connected",
  "namespace": "default",
  "pod": "my-pod",
  "container": "app",
  "message": "Connected to default/my-pod (app)"
}
```

**xterm.js 集成示例**:
```javascript
import { Terminal } from 'xterm';
import { FitAddon } from 'xterm-addon-fit';

const term = new Terminal();
const fitAddon = new FitAddon();
term.loadAddon(fitAddon);
term.open(document.getElementById('terminal'));

const token = 'eyJhbGc...';
const ws = new WebSocket(
  `ws://localhost:8080/ws/exec?namespace=default&pod=my-pod&container=app&token=${token}`
);

ws.onopen = () => {
  fitAddon.fit();
  // 发送初始尺寸
  ws.send(JSON.stringify({
    type: 'resize',
    cols: term.cols,
    rows: term.rows
  }));
};

term.onData((data) => {
  ws.send(JSON.stringify({
    type: 'stdin',
    data: data
  }));
});

ws.onmessage = (event) => {
  const data = JSON.parse(event.data);
  if (data.type === 'stdout' || data.type === 'stderr') {
    term.write(data.data);
  }
};

// 处理窗口大小变化
window.addEventListener('resize', () => {
  fitAddon.fit();
  ws.send(JSON.stringify({
    type: 'resize',
    cols: term.cols,
    rows: term.rows
  }));
});
```

---

## 关联资源

### 获取关联资源

**接口**: `GET /api/:resourceType/:namespace/:name/related`

**示例**:
```http
GET /api/deployments/default/nginx-deployment/related
Authorization: Bearer <token>
```

**响应**:
```json
{
  "code": 200,
  "message": "Related resources retrieved successfully",
  "data": [
    {
      "kind": "ReplicaSet",
      "name": "nginx-deployment-abc123",
      "relation": "child"
    },
    {
      "kind": "Service",
      "name": "nginx-service",
      "relation": "exposedBy"
    },
    {
      "kind": "HorizontalPodAutoscaler",
      "name": "nginx-hpa",
      "relation": "autoscaled"
    }
  ],
  "timestamp": 1234567890
}
```

---

## 管理接口

### 修改密码

**接口**: `POST /api/admin/password/change`

**请求**:
```json
{
  "oldPassword": "old-password",
  "newPassword": "new-secure-password"
}
```

**响应**:
```json
{
  "code": 200,
  "message": "Password changed successfully",
  "data": {
    "message": "密码修改成功"
  },
  "timestamp": 1234567890
}
```

### 生成随机密码

**接口**: `POST /api/admin/password/generate`

**请求**:
```json
{
  "length": 16
}
```

**响应**:
```json
{
  "code": 200,
  "message": "Password generated successfully",
  "data": {
    "password": "aB3$xY9!mN2@pQ7#",
    "hashedPassword": "$2a$12$...",
    "length": 16,
    "warning": "请安全保存明文密码，系统将不会再次显示"
  },
  "timestamp": 1234567890
}
```

### 密码哈希

**接口**: `POST /api/admin/password/hash`

**请求**:
```json
{
  "password": "plain-password"
}
```

**响应**:
```json
{
  "code": 200,
  "message": "Password hashed successfully",
  "data": {
    "hashedPassword": "$2a$12$...",
    "cost": 12
  },
  "timestamp": 1234567890
}
```

### 验证密码

**接口**: `POST /api/admin/password/validate`

**请求**:
```json
{
  "password": "plain-password",
  "hashedPassword": "$2a$12$..."
}
```

**响应**:
```json
{
  "code": 200,
  "message": "Verification completed",
  "data": {
    "valid": true,
    "message": "Password verification passed"
  },
  "timestamp": 1234567890
}
```

---

## 监控指标

### 获取所有指标

**接口**: `GET /api/metrics`

**响应**:
```json
{
  "code": 200,
  "message": "Metrics retrieved successfully",
  "data": {
    "timestamp": "2024-01-01T00:00:00Z",
    "system": {
      "cpu": {
        "usage_percent": 25.5,
        "cores": 4
      },
      "memory": {
        "used_mb": 512,
        "total_mb": 2048,
        "usage_percent": 25.0
      },
      "network": {
        "bytes_in": 1000000,
        "bytes_out": 2000000
      },
      "connections": {
        "active": 50,
        "idle": 10
      },
      "collected_at": "2024-01-01T00:00:00Z"
    },
    "business": {
      "totalRequests": 10000,
      "cacheHitRate": 85.5,
      "k8sApiCalls": 5000
    },
    "summary": {
      "total_count": 1,
      "system_count": 1,
      "business_count": 3
    }
  },
  "timestamp": 1234567890
}
```

### 获取业务指标

**接口**: `GET /api/metrics/business`

### 获取系统指标

**接口**: `GET /api/metrics/system`

### 获取健康指标

**接口**: `GET /api/metrics/health`

**响应**:
```json
{
  "code": 200,
  "message": "Health metrics retrieved successfully",
  "data": {
    "status": "healthy",
    "score": 95.0,
    "timestamp": "2024-01-01T00:00:00Z"
  },
  "timestamp": 1234567890
}
```

---

## 健康检查

**接口**: `GET /health`

**响应**:
```json
{
  "status": "healthy",
  "timestamp": 1234567890,
  "version": "2.0.0-optimized",
  "k8sConnected": true
}
```

---

## 缓存统计

**接口**: `GET /cache/stats`

**响应**:
```json
{
  "size": 150,
  "maxSize": 1000,
  "hits": 5000,
  "misses": 500,
  "hitRate": 90.9,
  "evictions": 50
}
```

---

## 错误码

| HTTP 状态码 | 错误码 | 说明 |
|------------|--------|------|
| 200 | 200 | 成功 |
| 400 | 400 | 请求参数错误 |
| 401 | 401 | 未授权/认证失败 |
| 403 | 403 | 禁止访问 |
| 404 | 404 | 资源不存在 |
| 409 | 409 | 资源冲突 |
| 422 | 422 | 验证失败 |
| 429 | 429 | 请求过多 |
| 500 | 500 | 服务器内部错误 |
| 503 | 503 | 服务不可用 |

### 错误响应格式

```json
{
  "code": 400,
  "message": "Invalid request parameter format",
  "details": "具体错误信息",
  "traceId": "追踪 ID",
  "timestamp": 1234567890
}
```

---

## 速率限制

所有 API 端点都有速率限制：

- 默认：100 请求/秒
- 登录接口：5 次失败后锁定 10 分钟
- WebSocket 连接：最大 100 个并发连接

超过限制会返回 `429 Too Many Requests`。

---

## 版本历史

| API 版本 | KubeVision 版本 | 说明 |
|---------|----------------|------|
| v1 | 1.0.0 | 初始版本 |
| v2 | 2.0.0 | 重构资源处理器，统一响应格式 |
