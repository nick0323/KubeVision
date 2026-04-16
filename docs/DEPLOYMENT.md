# KubeVision 部署指南

本文档介绍 KubeVision 在生产环境的部署方案。

## 目录

- [前置要求](#前置要求)
- [Docker 部署](#docker-部署)
- [Kubernetes 部署](#kubernetes-部署)
- [配置说明](#配置说明)
- [监控告警](#监控告警)
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
- Go 1.24+ (如需自行编译)
- Docker 20.10+ (如需容器化部署)

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

## Docker 部署

### 1. 构建镜像

```bash
# 方式 1: 使用 Dockerfile
docker build -t k8svision:latest .

# 方式 2: 使用 docker-compose
docker-compose build
```

### 2. 运行容器

```bash
docker run -d \
  --name k8svision \
  -p 8080:8080 \
  -v ~/.kube/config:/root/.kube/config:ro \
  -e K8SVISION_JWT_SECRET="your-secret-key-at-least-32-characters" \
  -e K8SVISION_AUTH_USERNAME="admin" \
  -e K8SVISION_AUTH_PASSWORD="your-secure-password" \
  -e K8SVISION_LOG_LEVEL="info" \
  --restart unless-stopped \
  k8svision:latest
```

### 3. 验证部署

```bash
# 查看日志
docker logs -f k8svision

# 健康检查
curl http://localhost:8080/health

# 访问 UI
# 打开浏览器 http://localhost:8080
```

### 4. Docker Compose

创建 `docker-compose.yml`：

```yaml
version: '3.8'

services:
  k8svision:
    image: k8svision:latest
    container_name: k8svision
    ports:
      - "8080:8080"
    volumes:
      - ~/.kube/config:/root/.kube/config:ro
    environment:
      - K8SVISION_JWT_SECRET=${K8SVISION_JWT_SECRET}
      - K8SVISION_AUTH_USERNAME=${K8SVISION_AUTH_USERNAME}
      - K8SVISION_AUTH_PASSWORD=${K8SVISION_AUTH_PASSWORD}
      - K8SVISION_LOG_LEVEL=info
    restart: unless-stopped
    healthcheck:
      test: ["CMD", "curl", "-f", "http://localhost:8080/health"]
      interval: 30s
      timeout: 10s
      retries: 3
```

启动：

```bash
docker-compose up -d
```

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
- apiGroups: [""]
  resources: ["*"]
  verbs: ["get", "list", "watch", "create", "update", "delete"]
- apiGroups: ["apps"]
  resources: ["*"]
  verbs: ["get", "list", "watch", "create", "update", "delete"]
- apiGroups: ["batch"]
  resources: ["*"]
  verbs: ["get", "list", "watch", "create", "update", "delete"]
- apiGroups: ["networking.k8s.io"]
  resources: ["*"]
  verbs: ["get", "list", "watch", "create", "update", "delete"]
- apiGroups: ["storage.k8s.io"]
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
# 生成 JWT Secret
JWT_SECRET=$(openssl rand -base64 32)

# 生成密码哈希 (使用 bcrypt)
# 可以使用在线工具或 KubeVision 的 /api/admin/password/hash 接口

# 创建 Secret
kubectl create secret generic k8svision-secret \
  -n k8svision \
  --from-literal=jwt-secret="$JWT_SECRET" \
  --from-literal=admin-password='your-hashed-password'
```

### 3. 部署应用

```yaml
# k8svision-deployment.yaml
apiVersion: apps/v1
kind: Deployment
metadata:
  name: k8svision
  namespace: k8svision
  labels:
    app: k8svision
spec:
  replicas: 2
  selector:
    matchLabels:
      app: k8svision
  template:
    metadata:
      labels:
        app: k8svision
      annotations:
        prometheus.io/scrape: "true"
        prometheus.io/port: "8080"
        prometheus.io/path: "/metrics"
    spec:
      serviceAccountName: k8svision-sa
      securityContext:
        runAsNonRoot: true
        runAsUser: 1000
      containers:
      - name: k8svision
        image: k8svision:latest
        imagePullPolicy: Always
        ports:
        - containerPort: 8080
          name: http
          protocol: TCP
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
        - name: K8SVISION_LOG_LEVEL
          value: "info"
        - name: K8SVISION_LOG_FORMAT
          value: "json"
        livenessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 10
          periodSeconds: 30
          timeoutSeconds: 5
          failureThreshold: 3
        readinessProbe:
          httpGet:
            path: /health
            port: 8080
          initialDelaySeconds: 5
          periodSeconds: 10
          timeoutSeconds: 3
          failureThreshold: 3
        resources:
          requests:
            cpu: "100m"
            memory: "128Mi"
          limits:
            cpu: "500m"
            memory: "512Mi"
        securityContext:
          allowPrivilegeEscalation: false
          readOnlyRootFilesystem: true
          capabilities:
            drop:
              - ALL
```

应用：

```bash
kubectl apply -f k8svision-deployment.yaml
```

### 4. 创建服务

```yaml
# k8svision-service.yaml
apiVersion: v1
kind: Service
metadata:
  name: k8svision
  namespace: k8svision
  labels:
    app: k8svision
spec:
  type: ClusterIP
  ports:
  - port: 80
    targetPort: 8080
    protocol: TCP
    name: http
  selector:
    app: k8svision
```

### 5. 配置 Ingress (可选)

```yaml
# k8svision-ingress.yaml
apiVersion: networking.k8s.io/v1
kind: Ingress
metadata:
  name: k8svision
  namespace: k8svision
  annotations:
    nginx.ingress.kubernetes.io/ssl-redirect: "true"
    nginx.ingress.kubernetes.io/proxy-read-timeout: "3600"
    nginx.ingress.kubernetes.io/proxy-send-timeout: "3600"
spec:
  ingressClassName: nginx
  tls:
  - hosts:
    - k8svision.example.com
    secretName: k8svision-tls
  rules:
  - host: k8svision.example.com
    http:
      paths:
      - path: /
        pathType: Prefix
        backend:
          service:
            name: k8svision
            port:
              number: 80
```

应用：

```bash
kubectl apply -f k8svision-service.yaml
kubectl apply -f k8svision-ingress.yaml
```

### 6. 验证部署

```bash
# 查看 Pod 状态
kubectl get pods -n k8svision

# 查看日志
kubectl logs -f deployment/k8svision -n k8svision

# 健康检查
kubectl run -it --rm --restart=Never test --image=curlimages/curl -- curl http://k8svision.k8svision.svc/health

# 端口转发访问
kubectl port-forward svc/k8svision -n k8svision 8080:80
```

---

## 配置说明

### 生产环境推荐配置

```yaml
server:
  port: "8080"
  host: "0.0.0.0"
  allowedOrigin:
    - "https://k8svision.your-company.com"  # 生产域名

kubernetes:
  timeout: 30s
  qps: 100        # 根据集群规模调整
  burst: 200      # 根据集群规模调整

jwt:
  secret: "使用 32+ 字符的随机密钥"
  expiration: 24h

log:
  level: "info"   # 生产环境建议 info，调试用 debug
  format: "json"  # 生产环境建议 json，方便日志收集

auth:
  username: "admin"
  password: "$2a$12$..."  # 必须使用 bcrypt 哈希
  maxLoginFail: 5
  lockDuration: 10m
  sessionTimeout: 24h
  enableRateLimit: true
  rateLimit: 100

cache:
  enabled: true
  ttl: 5m
  maxSize: 1000
  cleanupInterval: 10m
```

### 环境变量优先级

环境变量会覆盖配置文件中的值：

```bash
# 高优先级
K8SVISION_SERVER_PORT=8080
K8SVISION_LOG_LEVEL=info
K8SVISION_AUTH_MAX_FAIL=5
```

---

## 监控告警

### Prometheus 配置

```yaml
# prometheus-config.yaml
scrape_configs:
- job_name: 'k8svision'
  kubernetes_sd_configs:
  - role: pod
    namespaces:
      names:
      - k8svision
  relabel_configs:
  - source_labels: [__meta_kubernetes_pod_label_app]
    action: keep
    regex: k8svision
  - source_labels: [__meta_kubernetes_pod_annotation_prometheus_io_scrape]
    action: keep
    regex: true
  - source_labels: [__meta_kubernetes_pod_annotation_prometheus_io_port]
    action: replace
    target_label: __address__
    regex: (.+)
    replacement: ${1}
```

### 关键指标

| 指标 | 说明 | 告警阈值 |
|------|------|----------|
| `http_requests_total` | 请求总数 | - |
| `http_request_duration_seconds` | 请求延迟 | p99 > 5s |
| `cache_hit_rate` | 缓存命中率 | < 50% |
| `k8s_api_calls_total` | K8s API 调用 | - |
| `active_websocket_connections` | WebSocket 连接数 | > 80% 限制 |

### Grafana 面板

导入 Dashboard JSON（待提供）

---

## 故障排查

### Pod 无法启动

```bash
# 查看 Pod 状态
kubectl describe pod <pod-name> -n k8svision

# 查看日志
kubectl logs <pod-name> -n k8svision

# 常见问题：
# 1. RBAC 权限不足 - 检查 ClusterRoleBinding
# 2. 配置错误 - 检查 Secret 和 ConfigMap
# 3. 资源不足 - 检查 requests/limits
```

### WebSocket 连接失败

```bash
# 检查 Ingress 配置
kubectl get ingress k8svision -n k8svision

# 检查网络策略
kubectl get networkpolicy -n k8svision

# 测试连接
kubectl port-forward svc/k8svision -n k8svision 8080:80
# 访问 http://localhost:8080 测试 WebSocket
```

### K8s API 调用失败

```bash
# 检查 ServiceAccount
kubectl get sa k8svision-sa -n k8svision

# 检查权限
kubectl auth can-i get pods --as=system:serviceaccount:k8svision:k8svision-sa

# 检查 kubeconfig (如使用外部集群)
kubectl exec -it <pod-name> -n k8svision -- cat /root/.kube/config
```

### 性能问题

```bash
# 查看资源使用
kubectl top pod -n k8svision

# 查看慢查询日志
kubectl logs deployment/k8svision -n k8svision | grep "duration"

# 调整 QPS/Burst
kubectl set env deployment/k8svision -n k8svision \
  K8SVISION_KUBERNETES_QPS=200 \
  K8SVISION_KUBERNETES_BURST=400
```

---

## 升级指南

### 版本升级

```bash
# 1. 备份配置
kubectl get secret k8svision-secret -n k8svision -o yaml > secret-backup.yaml

# 2. 更新镜像
kubectl set image deployment/k8svision -n k8svision \
  k8svision=k8svision:latest

# 3. 滚动更新
kubectl rollout status deployment/k8svision -n k8svision

# 4. 验证
kubectl get pods -n k8svision
```

### 回滚

```bash
# 回滚到上一版本
kubectl rollout undo deployment/k8svision -n k8svision

# 回滚到指定版本
kubectl rollout undo deployment/k8svision -n k8svision --to-revision=2
```

---

## 备份恢复

### 配置备份

```bash
# 导出所有配置
kubectl get secret k8svision-secret -n k8svision -o yaml > secret.yaml
kubectl get configmap k8svision-config -n k8svision -o yaml > configmap.yaml
kubectl get deployment k8svision -n k8svision -o yaml > deployment.yaml
```

### 配置恢复

```bash
kubectl apply -f secret.yaml
kubectl apply -f configmap.yaml
kubectl apply -f deployment.yaml
```

---

## 安全加固

### 1. 启用 HTTPS

```yaml
# Ingress 配置 TLS
spec:
  tls:
  - hosts:
    - k8svision.example.com
    secretName: k8svision-tls
```

### 2. 网络策略

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
  policyTypes:
  - Ingress
  - Egress
  ingress:
  - from:
    - namespaceSelector:
        matchLabels:
          name: ingress-nginx
    ports:
    - protocol: TCP
      port: 8080
  egress:
  - to:
    - namespaceSelector: {}
    ports:
    - protocol: TCP
      port: 443
```

### 3. Pod 安全策略

已在 Deployment 中配置：
- 非 root 用户运行
- 只读文件系统
- 禁用特权升级
- 禁用所有 capabilities

---

## 联系支持

遇到问题请提交 Issue：https://github.com/nick0323/K8sVision/issues
