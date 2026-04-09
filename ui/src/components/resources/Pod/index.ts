// Pod 详情页组件导出
export { PodDetailPage } from './PodDetailPage';

// 通用组件从 common 导出
export { ResourceActionBar, TabNavigation } from '../../common';
export type { TabItem } from '../../common';

// Tabs
export { OverviewTab } from './tabs/OverviewTab';
export { LogsTab } from './tabs/LogsTab';
export { TerminalTab } from './tabs/TerminalTab';
export { RelatedTab } from './tabs/RelatedTab';
export { EventsTab } from './tabs/EventsTab';

// YamlTab 从通用目录导出
export { YamlTab } from '../../ResourceDetail/tabs/YamlTab';

// Hooks - 使用通用 useResourceDetail
// export { usePodDetail } from './hooks/usePodDetail';  // 已废弃，使用 useResourceDetail<Pod>

// 类型
export type {
  PodDetailPageProps,
  OverviewTabProps,
  YamlTabProps,
  LogsTabProps,
  TerminalTabProps,
  RelatedTabProps,
  EventsTabProps,
} from './types';
