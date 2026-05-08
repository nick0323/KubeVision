# KubeVision 功能添加计划

## Phase 1 — 核心运维能力（高优先级）

- [ ] 1.1 扩缩容 — Deployment/StatefulSet 增减副本按钮
- [ ] 1.2 滚动重启 — Deployment/StatefulSet/DaemonSet annotation restart
- [ ] 1.4 启用登出 — 取消 server.go /api/logout 注释
- [ ] 1.5 HPA 列表/详情 — 补全常量 + Getter + 前端页面

## Phase 2 — 常见缺失资源支持

- [ ] 2.1 NetworkPolicy 支持
- [ ] 2.2 ServiceAccount 支持
- [ ] 2.3 RBAC 四件套 — Role/RoleBinding/ClusterRole/ClusterRoleBinding
- [ ] 2.4 ResourceQuota / LimitRange 支持
- [ ] 2.5 PodDisruptionBudget 补全页面

## Phase 3 — 架构增强

- [ ] 3.1 动态客户端 — dynamic.Interface 替代 typed client
- [ ] 3.2 多集群 UI — 前端集群切换 + 后端配置加载
- [ ] 3.3 实时 Watch — Informer WebSocket 推送到前端
- [ ] 3.4 CRD 浏览器 — 动态列出 CRD 及其实例

## Phase 4 — 体验与可观测性

- [ ] 4.1 Pod/Container CPU/Mem 实时指标 + 趋势图
- [ ] 4.2 YAML 下载 + Diff 编辑器
- [ ] 4.3 ConfigMap/Secret 表单编辑器
- [ ] 4.4 审计日志 — 用户操作记录
- [ ] 4.5 资源拓扑图 — Owner/Service/Ingress/Pod 可视化
