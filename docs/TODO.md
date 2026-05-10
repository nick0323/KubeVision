# KubeVision 功能添加计划

> 最后检查日期: 2026-05-09
>
> ✅ = 已实现  ⚠️ = 部分实现  ❌ = 未实现

## Phase 1 — 核心运维能力（高优先级）

- [x] 1.1 扩缩容 — Deployment/StatefulSet 增减副本 ✅
  - PUT 路由 + ScaleResource 服务 + 前端 +/- 按钮 (ResourceActionBar)
- [x] 1.2 滚动重启 — Deployment/StatefulSet/DaemonSet annotation restart ✅
  - POST 路由 + RestartResource 服务 + 前端重启按钮
- [ ] 1.3 中文i18n（新增）
- [x] 1.4 启用登出 — POST /api/logout + 前端 Sign out ✅
  - 已实现 JWT blacklist + logout 按钮 (Sidebar)
- [ ] 1.5 HPA 列表/详情 — 补全常量 + Getter + 前端页面 ❌
  - 仅 Deployment/StatefulSet 详情页关联资源中显示

## Phase 2 — 常见缺失资源支持

- [x] 2.1 NetworkPolicy ✅
- [x] 2.2 ServiceAccount ✅
- [x] 2.3 RBAC 四件套 — Role/RoleBinding/ClusterRole/ClusterRoleBinding ✅
- [x] 2.4 ResourceQuota / LimitRange ✅
- [x] 2.5 PodDisruptionBudget ✅

## Phase 3 — 架构增强

- [ ] 3.1 动态客户端 — dynamic.Interface 替代 typed client ❌ (仅 ArgoCD Client 使用)
- [ ] 3.2 多集群 UI — 前端集群切换 + 后端配置加载 ❌ (仅有 clientPool/GetClient 骨架)
- [ ] 3.3 实时 Watch — Informer WebSocket 推送到前端 ❌
- [ ] 3.4 CRD 浏览器 — 动态列出 CRD 及其实例 ❌

## Phase 4 — 体验与可观测性

- [x] 4.1 Pod/Container CPU/Mem 指标 + 趋势图 ✅
  - Metrics API 已集成, 集群级 + 容器级展示
- [x] 4.2 YAML 下载 + Diff 编辑器 ✅
- [ ] 4.3 ConfigMap/Secret 表单编辑器 ❌ (仅有通用 YAML 编辑)
- [ ] 4.4 审计日志 — 用户操作记录 ❌
- [ ] 4.5 资源拓扑图 — Owner/Service/Ingress/Pod 可视化 ❌
