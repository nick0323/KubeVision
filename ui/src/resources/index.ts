// 通用资源组件导出 - 从 pages 导入
export { ResourceDetailPage } from '../pages/ResourceDetailPage';
export type { ResourceDetailPageProps } from '../pages/ResourceDetailPage.types';

// Hook
export { useResourceDetail } from './useResourceDetail';

// Tabs 从统一目录导出
export { OverviewTab } from '../tabs/OverviewTab';
export { YamlTab } from '../tabs/YamlTab';
export { EventsTab } from '../tabs/EventsTab';
export { RelatedTab } from '../tabs/RelatedTab';
export { PodsTab } from '../tabs/PodsTab';
export { EndpointsTab } from '../tabs/EndpointsTab';

// Hooks
export { useResourceDetail } from './useResourceDetail';

// 类型
export type {
  ResourceConfig,
  OverviewTabProps,
  YamlTabProps,
  EventsTabProps,
  RelatedTabProps,
} from './types';
