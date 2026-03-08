/**
 * 通用类型定义 - 修复版
 * 改进：
 * 1. 与后端 APIResponse 保持一致
 * 2. 添加缺失的字段
 * 3. 移除冗余类型
 */

import React from 'react';

// API 响应基础结构 - 与后端保持一致
export interface ApiResponse<T = any> {
  code: number;
  message: string;
  data: T;  // 直接是数据数组或对象
  traceId?: string;
  timestamp?: number;
  page?: PageMeta;
}

// API 错误响应
export interface ApiErrorResponse {
  code: number;
  message: string;
  details?: any;
  traceId?: string;
  timestamp?: number;
}

// 分页元数据
export interface PageMeta {
  total: number;
  limit: number;
  offset: number;
}

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

// K8s 事件
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

// 概览数据 - 与后端 OverviewStatus 保持一致
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

// CommonTable Props
export interface Column<T> {
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
}

// RefreshButton Props
export interface RefreshButtonProps {
  onClick: () => void;
  loading?: boolean;
  title?: string;
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

// ResourcePage Props
export interface ResourcePageProps extends OverviewPageProps {
  title: string;
  apiEndpoint: string;
  resourceType: string;
  columns: Column<any>[];
  statusMap?: Record<string, string>;
  namespaceFilter?: boolean;
}

// ==================== 新增类型定义 ====================

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

// ResourceListPage Props
export interface ResourceListPageProps<T = any> {
  config: PageConfig;
}
