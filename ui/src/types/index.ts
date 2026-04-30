// types/index.ts

// ==================== API 响应类型 ====================

/**
 * 通用 API 响应结构
 */
export interface APIResponse<T = unknown> {
  code: number;
  message: string;
  data: T;
  traceId?: string;
  timestamp?: number;
  page?: PageMeta;
}

/**
 * 分页元数据
 */
export interface PageMeta {
  total: number;
  limit: number;
  offset: number;
}

/**
 * 分页查询参数
 */
export interface ListQueryParams {
  limit: number;
  offset: number;
  namespace?: string;
  search?: string;
  sortBy?: string;
  sortOrder?: 'asc' | 'desc';
  labelSelector?: string;
  fieldSelector?: string;
}

/**
 * 分页响应数据
 */
export interface PaginatedResponse<T> {
  code: number;
  message: string;
  data: T[];
  page?: PageMeta;
}

// ==================== UI 组件 Props ====================

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
  requestsValue: string;
  requestsPercent: string;
  limitsValue: string;
  limitsPercent: string;
  totalValue: string;
  availableValue: string;
  unit: string;
}

export interface InfoCardProps {
  icon?: React.ReactNode;
  title: string;
  value: number | string;
  status?: React.ReactNode;
  children?: React.ReactNode;
}

export interface OverviewPageProps {
  collapsed: boolean;
  onToggleCollapsed: () => void;
}

export interface K8sEventSimple {
  type: string;
  reason: string;
  message: string;
  lastSeen: string;
  pod?: string;
  cloneset?: string;
  namespace?: string;
  name?: string;
  reporter?: string;
}

export interface OverviewData {
  nodeCount: number;
  nodeReadyCount: number;
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

export interface ErrorDisplayProps {
  message: string;
  type?: 'error' | 'warning' | 'info';
  onRetry?: () => void;
  showRetry?: boolean;
}

export interface PageHeaderProps {
  title?: string;
  children?: React.ReactNode;
  collapsed: boolean;
  onToggleCollapsed: () => void;
  breadcrumbs?: Array<{ label: string; path: string }>;
  onBreadcrumbClick?: (path: string) => void;
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

export interface LoginPageProps {
  onLogin: () => void;
}

export interface PageConfig {
  title: string;
  apiEndpoint: string;
  resourceType: string;
  columns: Array<{
    title: string;
    dataIndex: string;
    width?: string;
    sortable?: boolean;
    render?: (value: any, record: any) => React.ReactNode;
  }>;
  namespaceFilter: boolean;
  defaultSort?: {
    field: string;
    order: 'asc' | 'desc';
  };
}

// ==================== 通用 Hooks 类型 ====================

/**
 * 列表查询 Hook 返回值
 */
export interface UseListReturn<T> {
  data: T[];
  loading: boolean;
  isValidating: boolean;
  error: string | null;
  total: number;
  page: number;
  pageSize: number;
  setPage: (page: number) => void;
  setPageSize: (size: number) => void;
  namespace: string;
  search: string;
  setNamespace: (ns: string) => void;
  setSearch: (s: string) => void;
  sortField: string;
  sortOrder: 'asc' | 'desc';
  handleSort: (field: string) => void;
  refresh: () => Promise<void>;
  mutate: (newData?: T[]) => void;
  handleSubmit: () => void;
  clearSearch: () => void;
  namespaces: string[];
  namespacesLoading: boolean;
}

/**
 * 详情查询 Hook 返回值
 */
export interface UseDetailReturn<T> {
  data: T | null;
  loading: boolean;
  error: string | null;
  refresh: () => Promise<void>;
  mutate: (data: T) => void;
}
