/**
 * 通用资源详情页类型定义 - 兼容旧代码
 * 所有类型已迁移到 ../ResourceDetail/types.ts
 */

// 重新导出通用类型，保持向后兼容
export type {
  ResourceConfig,
  ResourceDetailPageProps,
  OverviewTabProps,
  YamlTabProps,
  EventsTabProps,
  RelatedTabProps,
} from '../../ResourceDetail/types';

// 重新导出常量（不是类型）
export { RESOURCE_CONFIGS } from '../../ResourceDetail/types';

// 重新导出 K8s 类型
export type { K8sResource } from '../../types/k8s-resources';
