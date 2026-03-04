/**
 * 通用类型定义
 */

import React from 'react';

// API 响应基础结构
export interface ApiResponse<T = any> {
  code: number;
  message: string;
  data: T;
  page?: PageMeta;
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
  type: string;
  reason: string;
  message: string;
  lastSeen: string;
  pod?: string;
  reporter?: string;
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
  memoryCapacity: number;
  memoryRequests: number;
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
