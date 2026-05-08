# KubeVision 接入 ArgoCD/GitOps 技术方案

## 1. 概述

### 1.1 目标
将 GitOps 工作流引入 KubeVision，利用 ArgoCD 实现：
- Kubernetes 资源的声明式持续交付
- 应用部署状态的统一可视化
- 多集群应用管理的集中化

### 1.2 为什么选择 ArgoCD
- CNCF 毕业项目，生态成熟
- 声明式 GitOps，Git 作为唯一事实源
- 支持 Helm、Kustomize、YAML 多种清单格式
- 自带 UI，但可通过 API 深度集成到 KubeVision

---

## 2. 架构设计

### 2.1 集成模式（推荐：集中式）
```
┌─────────────┐     ┌─────────────┐     ┌─────────────┐
│  Git Repo  │     │ KubeVision │     │ ArgoCD      │
│ (配置仓库) │────>│ 后端 API  │────>│ API Server  │
└─────────────┘     │ (Go)      │     └──────┬──────┘
                      └─────────────┘            │
                                           │ REST/gRPC
                                           ▼
                                    ┌─────────────┐
                                    │ Kubernetes 集群    │
                                    │ (应用部署目标)    │
                                    └─────────────┘
```

**说明**：
- Git Repo：存放 Kubernetes 清单（Helm/Kustomize/YAML）
- KubeVision 后端：通过 ArgoCD Go 客户端或 REST API 与 ArgoCD 交互
- ArgoCD：负责同步 Git 状态到集群

### 2.2 多集群支持
- ArgoCD 可以管理多个集群（通过 `argocd cluster add`）
- KubeVision 现有的 `service/client_manager.go` 可扩展为同时管理 ArgoCD 和应用集群

---

## 3. 后端实现（Go）

### 3.1 添加 ArgoCD Go 客户端依赖
```bash
go get github.com/argoproj/argo-cd/v3/pkg/apiclient
```

### 3.2 创建 ArgoCD 客户端管理模块
新建文件：`service/argocd_client.go`

```go
package service

import (
    "context"
    "fmt"

    "github.com/argoproj/argo-cd/v3/pkg/apiclient"
    applicationpkg "github.com/argoproj/argo-cd/v3/pkg/apiclient/application"
)

type ArgoCDManager struct {
    client apiclient.Client
}

func NewArgoCDManager(serverAddr, token string) (*ArgoCDManager, error) {
    client, err := apiclient.NewClient(&apiclient.ClientOptions{
        ServerAddr: serverAddr,
        AuthToken:  token,
        Insecure:   true, // 生产环境应使用 false + TLS
    })
    if err != nil {
        return nil, err
    }
    return &ArgoCDManager{client: client}, nil
}

// 列出所有应用
func (m *ArgoCDManager) ListApplications(ctx context.Context, project string) ([]*v1alpha1.Application, error) {
    appClient, err := m.client.NewApplicationClient()
    if err != nil {
        return nil, err
    }
    defer appClient.Close()

    resp, err := appClient.List(ctx, &application.ApplicationQuery{
        Projects: []string{project},
    })
    if err != nil {
        return nil, err
    }
    return resp.Items, nil
}

// 触发应用同步
func (m *ArgoCDManager) SyncApplication(ctx context.Context, appName string) error {
    appClient, err := m.client.NewApplicationClient()
    if err != nil {
        return err
    }
    defer appClient.Close()

    _, err = appClient.Sync(ctx, &application.ApplicationSyncRequest{
        Name: &appName,
    })
    return err
}
```

### 3.3 扩展现有 client_manager.go
在 `service/client_manager.go` 中添加 ArgoCD 客户端管理：

```go
type ClientManager struct {
    // ... 现有字段
    argoCDClient *ArgoCDManager
}

func (cm *ClientManager) GetArgoCDClient() *ArgoCDManager {
    return cm.argoCDClient
}
```

### 3.4 新增 API 端点
在 `api/` 下新增 ArgoCD 相关路由：

```go
// api/argocd_handler.go

// 列出 ArgoCD 应用
func ListArgoCDApps(c *gin.Context) {
    manager := getClientManager() // 获取 ClientManager 单例
    apps, err := manager.GetArgoCDClient().ListApplications(c, "")
    if err != nil {
        c.JSON(500, gin.H{"error": err.Error()})
        return
    }
    c.JSON(200, apps)
}

// 同步应用
func SyncArgoCDApp(c *gin.Context) {
    appName := c.Param("name")
    manager := getClientManager()
    err := manager.GetArgoCDClient().SyncApplication(c, appName)
    if err != nil {
        c.JSON(500, gin.H{"error": err.Error()})
        return
    }
    c.JSON(200, gin.H{"status": "synced"})
}
```

注册路由：
```go
// server/server.go
r.GET("/api/argocd/apps", api.ListArgoCDApps)
r.POST("/api/argocd/apps/:name/sync", api.SyncArgoCDApp)
```

---

## 4. 前端实现（React + TypeScript）

### 4.1 新增 ArgoCD 页面
创建 `ui/src/pages/ArgoCDPage.tsx`：

```tsx
import React, { useState, useEffect } from 'react';
import { FaRocket } from 'react-icons/fa';
import PageHeader from '../common/PageHeader';

export const ArgoCDPage = () => {
  const [apps, setApps] = useState<any[]>([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    fetch('/api/argocd/apps')
      .then(res => res.json())
      .then(data => {
        setApps(data);
        setLoading(false);
      });
  }, []);

  const syncApp = async (name: string) => {
    await fetch(`/api/argocd/apps/${name}/sync`, { method: 'POST' });
    // 刷新列表
    const res = await fetch('/api/argocd/apps');
    setApps(await res.json());
  };

  if (loading) return <div>Loading...</div>;

  return (
    <div className="argocd-page">
      <PageHeader title="ArgoCD Applications" />
      <div className="apps-grid">
        {apps.map(app => (
          <div key={app.metadata.name} className="app-card">
            <h3>{app.metadata.name}</h3>
            <span className={`status ${app.status.health.status}`}>
              {app.status.health.status}
            </span>
            <button onClick={() => syncApp(app.metadata.name)}>同步</button>
          </div>
        ))}
      </div>
    </div>
  );
};
```

### 4.2 添加侧边栏入口
在 `common/Sidebar.tsx` 的 MENU_LIST 中添加：

```tsx
{
  group: 'GitOps',
  items: [
    { key: 'argocd', label: 'ArgoCD', icon: 'FaRocket' }
  ]
}
```

---

## 5. 部署方案

### 5.1 安装 ArgoCD（到目标集群）
```bash
kubectl create namespace argocd
kubectl apply -n argocd -f https://raw.githubusercontent.com/argoproj/argo-cd/stable/manifests/install.yaml
```

### 5.2 配置 KubeVision 连接 ArgoCD
在 `config.yaml` 中添加：

```yaml
argocd:
  server: "argocd-server.argocd.svc.cluster.local:443"
  token: "your-argocd-admin-token"  # 或从 Secret 读取
```

### 5.3 高可用（可选）
```bash
kubectl scale -n argocd deployment argocd-server --replicas=2
kubectl scale -n argocd deployment argocd-repo-server --replicas=2
```

---

## 6. 最佳实践（来自 ArgoCD 官方）

1. **分离配置仓库与应用代码仓库**
   - 好处：单独修改清单、不同权限控制、避免 CI 循环触发
2. **使用 ApplicationSet 管理大量应用**
   - 适合多团队、多集群场景
3. **配置同步策略**
   - 自动同步：`spec.syncPolicy.automated: {}`
   - 手动同步：需要人工审批
4. **保护清单不可变**
   - 使用 Git tag 或 commit SHA，而不是 `HEAD`
5. **忽略无关字段**
   - 在 `spec.ignoreDifferences` 中配置（如 HPA 管理的 replica）
6. **RBAC 与 SSO 集成**
   - 使用 Dex 或企业 OIDC 提供商

---

## 7. 后续步骤

| 阶段 | 任务 | 预计时间 |
|------|------|----------|
| 第1周 | 后端集成 ArgoCD Go 客户端，完成 API 端点 | 3天 |
| 第1周 | 前端新增 ArgoCD 页面和侧边栏入口 | 2天 |
| 第2周 | 在测试集群部署 ArgoCD，联调 API | 2天 |
| 第2周 | 添加多集群支持（ArgoCD cluster secret） | 2天 |
| 第3周 | 完善错误处理、日志、认证 | 3天 |
| 第3周 | 编写文档和测试 | 2天 |

---

## 8. 风险与挑战

1. **ArgoCD API 版本稳定性**：`v3` 可能变化，建议锁定版本
2. **Token 管理**：ArgoCD admin token 需要安全存储（用 K8s Secret + 挂载）
3. **性能**：大量 Application 时，ArgoCD API 可能慢，考虑分页
4. **与现有 K8s 客户端的冲突**：确保不干扰 KubeVision 原有的资源监听

---

## 9. 替代方案（可选）

如果不想深度集成 ArgoCD，也可以：
- **仅展示**：在 KubeVision 中嵌入 ArgoCD 原生 UI（iframe）
- **Flux CD**：另一个 GitOps 工具，但 ArgoCD 生态更成熟

---

**方案制定完成**：2026-05-01
**版本**：v1.0
**负责人**：KubeVision 后端团队
