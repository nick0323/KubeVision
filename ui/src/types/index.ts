/**
 * 通用类型定义
 * 统一导出所有类型，简化导入路径
 */

// ==================== 从 k8s-resources 导出核心类型 ====================

export type {
  // K8s 基础类型
  K8sResource,
  K8sMetadata,
  K8sResourceList,
  K8sOwnerReference,
  K8sLocalObjectReference,

  // API 响应类型
  APIResponse,
  APIErrorResponse,
  PageMeta,
  PaginatedResponse,
  ListQueryParams,

  // 表格列类型
  StatusColumnDef,

  // 容器相关
  Container,
  ContainerPort,
  ContainerStatus,
  ContainerStatusState,
  ResourceRequirements,
  Probe,
  Volume,
  VolumeMount,
  EnvVar,
  EnvFromSource,
  SecurityContext,
  PodSpec,
  PodStatus,
  PodCondition,
  LabelSelector,
  Toleration,
  Affinity,

  // 资源类型
  Pod,
  PodPhase,
  Deployment,
  StatefulSet,
  DaemonSet,
  Service,
  ServiceType,
  ServicePort,
  Node,

  // 列表项类型
  PodListItem,
  DeploymentListItem,
  StatefulSetListItem,
  DaemonSetListItem,
  ServiceListItem,
  NodeListItem,
  GenericResourceItem,

  // 类型映射
  ResourceListItemMap,
  ResourceMap,
  ResourceType,
  GetListItem,
  GetResource,

  // K8s 事件
  K8sEvent,
  ObjectReference,
  EventSource,
  EventSeries,

  // UI 类型
  ResourceListItem,
  ActionButton,
} from './k8s-resources';

// ==================== 菜单类型 ====================

export interface MenuItem {
  key: string;
  label: string;
  icon: string;
}

export interface MenuGroup {
  group: string;
  items: MenuItem[];
}

// ==================== 表格列定义（本地定义） ====================

export interface ColumnDef<T> {
  title: string;
  dataIndex: keyof T | string;
  width?: number | string;
  sortable?: boolean;
  render?: (value: T[keyof T], record: T, index: number) => React.ReactNode;
  className?: string;
  hidden?: boolean;
}

export type Column<T = any> = ColumnDef<T>;

// ==================== 概览页类型 ====================

export interface K8sEventSimple {
  namespace: string;
  name: string;
  type: string;
  reason: string;
  message: string;
  count?: number;
  firstSeen: string;
  lastSeen: string;
  duration?: string;
  pod?: string;
  reporter?: string;
  cloneset?: string;
}

export interface OverviewData {
  nodeCount: number;
  nodeReady: number;
  podCount: number;
  podNotReady: number;
  namespaceCount: number;
  serviceCount: number;
  cpuCapacity: number;
  cpuRequests: number;
  cpuLimits: number;
  memoryCapacity: number;
  memoryRequests: number;
  memoryLimits: number;
  events: K8sEventSimple[];
}

// ==================== 组件 Props 类型 ====================

export interface InfoCardProps {
  icon: React.ReactNode;
  title: string;
  value: number;
  status?: React.ReactNode;
  children?: React.ReactNode;
}

export interface ErrorDisplayProps {
  message: string;
  type?: 'error' | 'warning' | 'info';
  onRetry?: () => void;
  showRetry?: boolean;
}

export interface PageHeaderProps {
  title: string;
  collapsed: boolean;
  onToggleCollapsed: () => void;
  breadcrumbs?: { label: string; path: string }[];
  onBreadcrumbClick?: (path: string) => void;
  children?: React.ReactNode;
}

export interface PaginationProps {
  currentPage: number;
  total: number;
  pageSize: number;
  onPageChange: (page: number) => void;
  onPageSizeChange?: (pageSize: number) => void;
  pageSizeOptions?: number[];
  fixed?: boolean;
  fixedBottom?: boolean;
  showQuickJumper?: boolean;
}

export interface ResourceSummaryProps {
  title: string;
  requestsValue: number | string;
  requestsPercent: number | string;
  limitsValue: number | string;
  limitsPercent: number | string;
  totalValue: number | string;
  availableValue: number | string;
  unit: string;
}

export interface NamespaceSelectProps {
  value: string;
  onChange: (value: string) => void;
  placeholder?: string;
  disabled?: boolean;
  options?: string[];
  className?: string;
  width?: string;
}

export interface SearchInputProps {
  placeholder?: string;
  value: string;
  onChange: (e: React.ChangeEvent<HTMLInputElement>) => void;
  onSubmit?: (e: React.FormEvent) => void;
  onClear?: () => void;
  isSearching?: boolean;
  hasSearchResults?: boolean;
  showSearchButton?: boolean;
  showClearButton?: boolean;
  disabled?: boolean;
}

export interface RefreshButtonProps {
  onClick: () => void;
  loading?: boolean;
  title?: string;
  showLastUpdated?: boolean;
}

export interface StatusBadgeProps {
  status: string;
  resourceType?: string;
}

export interface LoginPageProps {
  onLogin: () => void;
}

export interface OverviewPageProps {
  collapsed: boolean;
  onToggleCollapsed: () => void;
}

// ==================== 页面配置 ====================

export interface PageConfig {
  title: string;
  apiEndpoint: string;
  resourceType: string;
  columns: Column<any>[];
  namespaceFilter?: boolean;
  defaultSort?: {
    field: string;
    order: 'asc' | 'desc';
  };
}
