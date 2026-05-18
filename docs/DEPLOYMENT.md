# KubeVision 部署指南

本文档介绍 KubeVision 在生产环境的部署方案。

## 目录

- [前置要求](#前置要求)
- [Docker Compose 部署](#docker-compose-部署)
- [手动 Docker 部署](#手动-docker-部署)
- [Kubernetes 部署](#kubernetes-部署)
- [配置说明](#配置说明)
- [CI/CD](#cicd)
- [故障排查](#故障排查)

---

## 前置要求

### 硬件要求

| 环境 | CPU | 内存 | 磁盘 |
|------|-----|------|------|
| 开发 | 1 核 | 1GB | 500MB |
| 生产 | 2 核 | 2GB | 1GB |

### 软件要求

- Kubernetes 1.25+
- Go 1.26+ (如需自行编译)
- Docker 20.10+ / Docker Compose v2+

### 权限要求

需要以下 Kubernetes RBAC 权限：

```yaml
- apiGroups: [""]
  resources: ["pods", "services", "configmaps", "secrets", "events", "namespaces", "nodes"]
  verbs: ["get", "list", "watch", "create", "update", "delete"]
- apiGroups: ["apps"]
  resources: ["deployments", "statefulsets", "daemonsets", "replicasets"]
  verbs: ["get", "list", "watch", "create", "update", "delete"]
- apiGroups: ["batch"]
  resources: ["jobs", "cronjobs"]
  verbs: ["get", "list", "watch", "create", "update", "delete"]
- apiGroups: ["networking.k8s.io"]
  resources: ["ingresses"]
  verbs: ["get", "list", "watch", "create", "update", "delete"]
- apiGroups: ["storage.k8s.io"]
  resources: ["storageclasses", "persistentvolumes", "persistentvolumeclaims"]
  verbs: ["get", "list", "watch", "create", "update", "delete"]
- apiGroups: [""]
  resources: ["pods/exec", "pods/log"]
  verbs: ["create", "get"]
```

---

## Docker Compose 部署

### 1. 准备配置

```bash
cp .env.example .env
```

编辑 `.env` 文件：

```bash
K8SVISION_JWT_SECRET=<64字符以上的随机字符串>
K8SVISION_AUTH_USERNAME=admin
K8SVISION_AUTH_PASSWORD=<你的密码>
K8SVISION_LOG_LEVEL=info
```

### 2. 启动

```bash
docker compose up -d
```

### 3. 验证

```bash
docker compose logs -f
curl http://localhost:8080/health
```

打开浏览器访问 `http://localhost:8080`。

---

## 手动 Docker 部署

### 拉取镜像

```bash
docker pull ghcr.io/nick0323/KubeVision:latest
```

### 运行

```bash
docker run -d \
  --name k8svision \
  -p 8080:8080 \
  -v ~/.kube/config:/root/.kube/config:ro \
  -e K8SVISION_JWT_SECRET="<your-secret>" \
  -e K8SVISION_AUTH_USERNAME="admin" \
  -e K8SVISION_AUTH_PASSWORD="<your-password>" \
  -e K8SVISION_LOG_LEVEL="info" \
  --restart unless-stopped \
  ghcr.io/nick0323/KubeVision:latest
```

访问 `http://localhost:8080`。

---

## Kubernetes 部署

### 1. 准备命名空间和 ServiceAccount

```yaml
# k8svision-namespace.yaml
apiVersion: v1
kind: Namespace
metadata:
  name: k8svision
```

```yaml
# k8svision-rbac.yaml
apiVersion: v1
kind: ServiceAccount
metadata:
  name: k8svision-sa
  namespace: k8svision
---
apiVersion: rbac.authorization.k8s.io/v1
kind: ClusterRole
metadata:
  name: k8svision-role
rules:
- apiGroups: ["*"]
  resources: ["*"]
  verbs: ["get", "list", "watch", "create", "update", "delete"]
- apiGroups: [""]
  resources: ["pods/exec", "pods/log"]
  verbs: ["create", "get"]
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
  namespace: k8svision
```

应用：

```bash
kubectl apply -f k8svision-namespace.yaml
kubectl apply -f k8svision-rbac.yaml
```

### 2. 创建密钥

```bash
JWT_SECRET=$(openssl rand -base64 32)
kubectl create secret generic k8svision-secret \
  -n k8svision \
  --from-literal=jwt-secret="$JWT_SECRET" \
  --from-literal=admin-username="admin" \
  --from-literal=admin-password="<bcrypt-hashed-password>"
```

### 3. 部署应用

```yaml
# k8svision-deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: k8svision
  namespace: k8svision
spec:
  replicas: 1
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
        image: ghcr.io/nick0323/KubeVision:latest
        imagePullPolicy: Always
        ports:
        - containerPort: 8080
        env:
        - name: K8SVISION_JWT_SECRET
          valueFrom:
            secretKeyRef:
              name: k8svision-secret
              key: jwt-secret
        - name: K8SVISION_AUTH_USERNAME
          valueFrom:
            secretKeyRef:
              name: k8svision-secret
              key: admin-username
        - name: K8SVISION_AUTH_PASSWORD
          valueFrom:
            secretKeyRef:
              name: k8svision-secret
              key: admin-password
        - name: K8SVISION_LOG_LEVEL
          value: "info"
        livenessProbe:
          httpGet: { path: /health, port: 8080 }
          initialDelaySeconds: 10
          periodSeconds: 30
        readinessProbe:
          httpGet: { path: /health, port: 8080 }
          initialDelaySeconds: 5
          periodSeconds: 10
        resources:
          requests: { cpu: "100m", memory: "128Mi" }
          limits:   { cpu: "500m", memory: "512Mi" }
```

Service targets port 8080:

```yaml
# k8svision-svc.yaml
apiVersion: v1
kind: Service
metadata:
  name: k8svision
  namespace: k8svision
spec:
  type: ClusterIP
  ports:
  - port: 80
    targetPort: 8080
  selector:
    app: k8svision
```

应用：

```bash
kubectl apply -f k8svision-deployment.yaml
kubectl apply -f k8svision-svc.yaml
```

验证：

```bash
kubectl get pods -n k8svision
kubectl logs -f deployment/k8svision -n k8svision
kubectl port-forward svc/k8svision -n k8svision 8080:80
```

---

## 配置说明

### 环境变量

| 变量 | 必填 | 说明 |
|------|------|------|
| `K8SVISION_AUTH_PASSWORD` | 是 | bcrypt 哈希密码 |
| `K8SVISION_JWT_SECRET` | 是 | JWT 签名密钥（64+ 字符） |
| `KUBECONFIG` | 否 | K8s 配置路径（为空时使用 in-cluster 配置） |
| `K8SVISION_LOG_LEVEL` | 否 | debug/info/warn/error，默认 info |
| `K8SVISION_SERVER_HOST` | 否 | 监听地址，默认 0.0.0.0 |
| `K8SVISION_SERVER_PORT` | 否 | 监听端口，默认 8080 |

### 生成密码哈希

```bash
go run cmd/tools/generate_password.go <your_password>
```

---

## CI/CD

镜像仓库：`ghcr.io/nick0323/KubeVision`（前后端合并为单镜像）

| 触发 | 标签 |
|------|------|
| push to main | `main`, `sha-<hash>` |
| push tag `v1.2.3` | `1.2.3`, `1.2`, `sha-<hash>` |

PR 时只构建不推送。

---

## 故障排查

### 无法连接 K8s

```bash
# 检查 kubeconfig 挂载
docker exec k8svision cat /root/.kube/config

# in-cluster 模式权限
kubectl get sa k8svision-sa -n k8svision
kubectl auth can-i get pods --as=system:serviceaccount:k8svision:k8svision-sa
```

### 服务无法访问

```bash
# 检查健康状态
curl http://localhost:8080/health

# 查看日志
docker logs k8svision 2>&1 | tail -50
```

### 登录失败

- 确认 `K8SVISION_JWT_SECRET` 为 64+ 字符
- 确认 `K8SVISION_AUTH_PASSWORD` 为 bcrypt 哈希格式
```bash
docker logs k8svision 2>&1 | grep -i auth
```

### 性能问题

```bash
docker stats k8svision
# 或 K8s: kubectl top pod -n k8svision
docker logs k8svision 2>&1 | grep "duration"
```

---

## 升级

```bash
kubectl set image deployment/k8svision -n k8svision \
  k8svision=ghcr.io/nick0323/KubeVision:latest
kubectl rollout status deployment/k8svision -n k8svision
```

回滚：

```bash
kubectl rollout undo deployment/k8svision -n k8svision
```

---

## 安全加固

### 启用 HTTPS

通过 Ingress 配置 TLS：

```yaml
spec:
  tls:
  - hosts: ["k8svision.example.com"]
    secretName: k8svision-tls
```

### 网络策略

```yaml
apiVersion: networking.k8s.io/v1
kind: NetworkPolicy
metadata:
  name: k8svision-network-policy
  namespace: k8svision
spec:
  podSelector:
    matchLabels:
      app: k8svision
  policyTypes: [Ingress, Egress]
  ingress:
  - from:
    - namespaceSelector:
        matchLabels:
          name: ingress-nginx
    ports: [{ protocol: TCP, port: 8080 }]
  egress:
  - to: [{ namespaceSelector: {} }]
    ports: [{ protocol: TCP, port: 443 }]
```

---

## 联系支持

遇到问题请提交 Issue：https://github.com/nick0323/KubeVision/issues
