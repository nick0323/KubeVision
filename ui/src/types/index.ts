// types/index.ts

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
  icon: React.ReactNode;
  title: string;
  value: number | string;
  status?: React.ReactNode;
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

export interface ErrorDisplayProps {
  message: string;
  type?: 'error' | 'warning' | 'info';
  onRetry?: () => void;
  showRetry?: boolean;
}
