/**
 * 通用组件导出
 */
export { ResourceActionBar } from './ResourceActionBar';
export type { ResourceActionBarProps } from './ResourceActionBar';

export { TabNavigation } from './TabNavigation';
export type { TabNavigationProps } from './TabNavigation';

// 类型导出
export type {
  // 通用 Tab Props
  OverviewTabProps,
  YamlTabProps,
  EventsTabProps,
  RelatedTabProps,
  LogsTabProps,
  TerminalTabProps,
  
  // Pod 特有类型
  PodDetailPageProps,
  ContainerState,
  ContainerStatusSummary,
  PodCondition,
  LogOptions,
  TerminalOptions,
  EventStats,
  
  // 通用资源类型
  ResourceConfig,
  ResourceDetailPageProps,
  RelatedResource,
  TabItem,
} from '../ResourceDetail/types';

// 导出常量
export { RESOURCE_CONFIGS } from '../ResourceDetail/types';
