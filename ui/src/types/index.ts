/**
 * 通用类型定义
 */

import React from 'react';

// ==================== 从 k8s-resources 导出核心类型 ====================

export type {
  // K8s 基础类型
  K8sResource,
  K8sMetadata,
  K8sResourceList,
  
  // API 响应类型
  APIResponse,
  APIErrorResponse,
  PageMeta,
  PaginatedResponse,
  ListQueryParams,
  
  // 表格列类型
  ColumnDef,
  StatusColumnDef,
  
  // 资源类型
  Pod,
  Deployment,
  StatefulSet,
  DaemonSet,
  Service,
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
} from './k8s-resources';

// ==================== 本地类型定义 ====================

// 基础 Props
export interface BaseProps {
  className?: string;
  style?: React.CSSProperties;
  children?: React.ReactNode;
}

// 菜单项
export interface MenuItem {
  key: string;
  label: string;
  icon: string;
}

// 菜单组
export interface MenuGroup {
  group: string;
  items: MenuItem[];
}

// K8s 事件（保留，用于概览页）
export interface K8sEvent {
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

// 概览数据
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
  events: K8sEvent[];
}

// InfoCard Props
export interface InfoCardProps {
  icon: React.ReactNode;
  title: string;
  value: number;
  status?: React.ReactNode;
}

// Loading Props
export interface LoadingProps {
  text?: string;
  size?: 'sm' | 'md' | 'lg';
  overlay?: boolean;
}

// ErrorDisplay Props
export interface ErrorDisplayProps {
  message: string;
  type?: 'error' | 'warning' | 'info';
  onRetry?: () => void;
  showRetry?: boolean;
}

// PageHeader Props
export interface PageHeaderProps {
  title: string;
  collapsed: boolean;
  onToggleCollapsed: () => void;
  children?: React.ReactNode;
}

// CommonTable Props（保留向后兼容，建议使用 ColumnDef）
export interface Column<T = any> {
  title: string;
  dataIndex: string;
  width?: number | string;
  sortable?: boolean;
  render?: (text: any, record: T, index: number) => React.ReactNode;
}

export interface CommonTableProps<T> {
  columns: Column<T>[];
  data: T[];
  emptyText?: string;
  onRowClick?: (record: T) => void;
}

// Pagination Props
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

// ResourceSummary Props
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

// NamespaceSelect Props
export interface NamespaceSelectProps {
  value: string;
  onChange: (value: string) => void;
  placeholder?: string;
  disabled?: boolean;
  options?: string[];
}

// SearchInput Props
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

// RefreshButton Props
export interface RefreshButtonProps {
  onClick: () => void;
  loading?: boolean;
  title?: string;
  showLastUpdated?: boolean;
}

// LoginPage Props
export interface LoginPageProps {
  onLogin: () => void;
}

// OverviewPage Props
export interface OverviewPageProps {
  collapsed: boolean;
  onToggleCollapsed: () => void;
}

// ResourcePage Props（保留向后兼容，建议使用泛型版本）
export interface ResourcePageProps {
  title: string;
  apiEndpoint: string;
  resourceType: string;
  columns: Column<any>[];
  statusMap?: Record<string, string>;
  namespaceFilter?: boolean;
}

// ==================== 页面配置 ====================

// 页面配置
export interface PageConfig {
  title: string;
  apiEndpoint: string;
  resourceType: string;
  columns: Column<any>[];
  namespaceFilter?: boolean;
  statusFilter?: string[];
  typeFilter?: string[];
  defaultSort?: {
    field: string;
    order: 'asc' | 'desc';
  };
}

// StatusBadge Props
export interface StatusBadgeProps {
  status: string;
  resourceType?: string;
}

// SearchFilter Props
export interface SearchFilterProps {
  value: string;
  onSearch: (value: string) => void;
  placeholder?: string;
  debounceMs?: number;
}

// NamespaceFilter Props
export interface NamespaceFilterProps {
  namespaces: string[];
  value: string;
  onChange: (value: string) => void;
}

// StatusFilter Props
export interface StatusFilterProps {
  statuses: string[];
  value: string;
  onChange: (value: string) => void;
}
