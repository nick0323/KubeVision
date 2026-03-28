// Pod 详情页组件导出
export { PodDetailPage } from './PodDetailPage';
export { ResourceActionBar } from './components/ResourceActionBar';
export { TabNavigation } from './components/TabNavigation';
export type { TabItem } from './components/TabNavigation';

// Tabs
export { OverviewTab } from './tabs/OverviewTab';
export { YamlTab } from './tabs/YamlTab';
export { LogsTab } from './tabs/LogsTab';
export { TerminalTab } from './tabs/TerminalTab';
export { RelatedTab } from './tabs/RelatedTab';
export { EventsTab } from './tabs/EventsTab';

// Hooks
export { usePodDetail } from './hooks/usePodDetail';

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
