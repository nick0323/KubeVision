/**
 * Pod 详情页类型定义 - 兼容旧代码
 * 所有类型已迁移到 ../ResourceDetail/types.ts
 */

// 重新导出通用类型，保持向后兼容
export type {
  OverviewTabProps,
  YamlTabProps,
  LogsTabProps,
  TerminalTabProps,
  RelatedTabProps,
  EventsTabProps,
  ContainerState,
  ContainerStatusSummary,
  PodCondition,
  LogOptions,
  TerminalOptions,
  EventStats,
  PodDetailPageProps,
} from '../../ResourceDetail/types';

// 重新导出 K8s 类型
export type { Pod, K8sOwnerReference } from '../../types/k8s-resources';
