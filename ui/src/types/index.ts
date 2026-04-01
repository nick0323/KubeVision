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

// ==================== 页面配置 ====================

// 页面配置
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

// StatusBadge Props
export interface StatusBadgeProps {
  status: string;
  resourceType?: string;
}
