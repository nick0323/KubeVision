/**
 * 通用资源详情页类型定义
 */

import { K8sOwnerReference } from '../types/k8s-resources';

// ==================== 通用 Tab Props ====================

/**
 * Overview Tab Props - 通用版本
 */
export interface OverviewTabProps<T = any> {
  data?: T | null;
  pod?: T | null; // 兼容 Pod 详情页
  loading: boolean;
  onRefresh?: () => void;
  resourceType?: string;
}

/**
 * YAML Tab Props
 */
export interface YamlTabProps {
  namespace: string;
  name: string;
  resourceType: string;
  data?: any | null;
  pod?: any | null; // 兼容 Pod 详情页
}

/**
 * Events Tab Props
 */
export interface EventsTabProps {
  namespace: string;
  podName?: string; // 兼容 Pod 详情页
  name: string;
  resourceKind?: string;
  onRefresh?: () => void;
}

/**
 * Related Tab Props
 */
export interface RelatedTabProps {
  namespace: string;
  name: string;
  resourceType: string;
  ownerReferences?: K8sOwnerReference[];
}

/**
 * Logs Tab Props
 */
export interface LogsTabProps {
  namespace: string;
  name: string;
  containers: any[];
}

/**
 * Terminal Tab Props
 */
export interface TerminalTabProps {
  namespace: string;
  name: string;
  containers: any[];
}

// ==================== Pod 特有类型 ====================

/**
 * Pod 详情页 Props
 */
export interface PodDetailPageProps {
  namespace: string;
  name: string;
  collapsed: boolean;
  onToggleCollapsed: () => void;
}

/**
 * 关联资源
 */
export interface RelatedResource {
  kind: string;
  name: string;
  namespace?: string;
  apiVersion?: string;
}

// ==================== 通用资源类型 ====================

/**
 * 资源类型配置
 */
export interface ResourceConfig {
  title: string;
  tabs: string[];
  hasLogs?: boolean;
  hasTerminal?: boolean;
  hasPods?: boolean;
  hasEndpoints?: boolean;
}

/**
 * 通用资源详情 Props
 */
export interface ResourceDetailPageProps {
  resourceType: string;
  namespace: string;
  name: string;
  collapsed: boolean;
  onToggleCollapsed: () => void;
}

/**
 * 资源配置映射
 */
export const RESOURCE_CONFIGS: Record<string, ResourceConfig> = {
  pod: {
    title: 'Pod',
    tabs: ['overview', 'yaml', 'logs', 'terminal', 'related', 'events'],
    hasLogs: true,
    hasTerminal: true,
  },
  deployment: {
    title: 'Deployment',
    tabs: ['overview', 'yaml', 'pods', 'related', 'events'],
    hasPods: true,
  },
  statefulset: {
    title: 'StatefulSet',
    tabs: ['overview', 'yaml', 'pods', 'related', 'events'],
    hasPods: true,
  },
  daemonset: {
    title: 'DaemonSet',
    tabs: ['overview', 'yaml', 'pods', 'related', 'events'],
    hasPods: true,
  },
  job: {
    title: 'Job',
    tabs: ['overview', 'yaml', 'pods', 'related', 'events'],
    hasPods: true,
  },
  cronjob: {
    title: 'CronJob',
    tabs: ['overview', 'yaml', 'related', 'events'],
  },
  service: {
    title: 'Service',
    tabs: ['overview', 'yaml', 'endpoints', 'related', 'events'],
  },
  ingress: {
    title: 'Ingress',
    tabs: ['overview', 'yaml', 'related', 'events'],
  },
  configmap: {
    title: 'ConfigMap',
    tabs: ['overview', 'yaml', 'related', 'events'],
  },
  secret: {
    title: 'Secret',
    tabs: ['overview', 'yaml', 'related', 'events'],
  },
  pvc: {
    title: 'PersistentVolumeClaim',
    tabs: ['overview', 'yaml', 'related', 'events'],
  },
  pv: {
    title: 'PersistentVolume',
    tabs: ['overview', 'yaml', 'related', 'events'],
  },
  storageclass: {
    title: 'StorageClass',
    tabs: ['overview', 'yaml', 'related', 'events'],
  },
  namespace: {
    title: 'Namespace',
    tabs: ['overview', 'yaml', 'related', 'events'],
  },
  node: {
    title: 'Node',
    tabs: ['overview', 'yaml', 'pods', 'events'],
    hasPods: true,
  },
};
