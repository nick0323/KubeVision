/**
 * 通用资源详情页类型定义
 * Type definitions for Resource Detail Page
 */

import { K8sOwnerReference, Container as K8sContainer } from '../types/k8s-resources';

// ============================================================================
// 通用 Tab Props
// ============================================================================

/**
 * Overview Tab Props
 */
export interface OverviewTabProps<T = unknown> {
  data?: T | null;
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
  data?: unknown | null;
}

/**
 * Events Tab Props
 */
export interface EventsTabProps {
  namespace: string;
  podName?: string; // 兼容旧版 Pod 详情页
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
  containers: K8sContainer[];
}

/**
 * Terminal Tab Props
 */
export interface TerminalTabProps {
  namespace: string;
  name: string;
  containers: K8sContainer[];
}

/**
 * Pods Tab Props
 */
export interface PodsTabProps {
  namespace: string;
  resourceName: string;
  resourceKind: string;
  resourceLabels?: Record<string, string>;
  ownerReferences?: unknown[];
}

// ============================================================================
// 页面级 Props
// ============================================================================

/**
 * 通用资源详情页 Props
 */
export interface ResourceDetailPageProps {
  resourceType: string;
  namespace: string;
  name: string;
  collapsed: boolean;
  onToggleCollapsed: () => void;
}

/**
 * YAML Tab Props (扩展版，兼容旧代码)
 */
export interface YamlTabPropsExtended extends YamlTabProps {
  pod?: unknown | null; // 兼容旧版 Pod 详情页
}

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
  [key: string]: unknown; // 支持动态属性访问
}

// ============================================================================
// 资源配置工厂函数
// 用于减少重复配置代码
// ============================================================================

/**
 * 创建资源配置
 * @param title 资源标题
 * @param tabs Tab 列表
 * @param options 额外选项
 */
export function createResourceConfig(
  title: string,
  tabs: readonly string[],
  options: Partial<Omit<ResourceConfig, 'title' | 'tabs'>> = {}
): ResourceConfig {
  return {
    title,
    tabs: [...tabs],
    ...options,
  };
}

// 通用 Tab 配置常量
const COMMON_TABS = {
  BASIC: ['overview', 'yaml', 'related', 'events'] as const,
  WITH_PODS: ['overview', 'yaml', 'pods', 'related', 'events'] as const,
  POD_FULL: ['overview', 'yaml', 'logs', 'terminal', 'related', 'events'] as const,
} as const;

// ============================================================================
// 资源配置映射
// 定义每种 K8s 资源的显示配置
// ============================================================================

export const RESOURCE_CONFIGS: Record<string, ResourceConfig> = {
  // Pod - 完整功能
  pod: createResourceConfig('Pod', COMMON_TABS.POD_FULL, {
    hasLogs: true,
    hasTerminal: true,
  }),

  // Workloads - 支持 Pods
  deployment: createResourceConfig('Deployment', COMMON_TABS.WITH_PODS, { hasPods: true }),
  statefulset: createResourceConfig('StatefulSet', COMMON_TABS.WITH_PODS, { hasPods: true }),
  daemonset: createResourceConfig('DaemonSet', COMMON_TABS.WITH_PODS, { hasPods: true }),
  job: createResourceConfig('Job', COMMON_TABS.WITH_PODS, { hasPods: true }),

  // CronJob - 不支持 Pods
  cronjob: createResourceConfig('CronJob', COMMON_TABS.BASIC),

  // Service - 支持 Endpoints
  service: createResourceConfig('Service', ['overview', 'yaml', 'endpoints', 'related', 'events'], {
    hasEndpoints: true,
  }),

  // Network
  ingress: createResourceConfig('Ingress', COMMON_TABS.BASIC),

  // Configuration
  configmap: createResourceConfig('ConfigMap', COMMON_TABS.BASIC),
  secret: createResourceConfig('Secret', COMMON_TABS.BASIC),

  // Storage
  pvc: createResourceConfig('PersistentVolumeClaim', COMMON_TABS.BASIC),
  pv: createResourceConfig('PersistentVolume', COMMON_TABS.BASIC),
  storageclass: createResourceConfig('StorageClass', COMMON_TABS.BASIC),

  // Cluster
  namespace: createResourceConfig('Namespace', COMMON_TABS.BASIC),
  node: createResourceConfig('Node', ['overview', 'yaml', 'pods', 'events'], { hasPods: true }),
} as const;

// ============================================================================
// 辅助类型
// ============================================================================

/**
 * 关联资源
 */
export interface RelatedResource {
  kind: string;
  name: string;
  namespace?: string;
  apiVersion?: string;
}

/**
 * 资源类型键名
 */
export type ResourceTypeKey = keyof typeof RESOURCE_CONFIGS;

/**
 * Tab 键名类型
 */
export type TabKey =
  | 'overview'
  | 'yaml'
  | 'logs'
  | 'terminal'
  | 'related'
  | 'events'
  | 'pods'
  | 'endpoints';

/**
 * 获取资源的 Tab 列表
 */
export function getResourceTabs(resourceType: string): string[] {
  const config = RESOURCE_CONFIGS[resourceType as keyof typeof RESOURCE_CONFIGS];
  return config?.tabs || COMMON_TABS.BASIC;
}

/**
 * 检查资源是否支持特定功能
 */
export function hasResourceFeature(
  resourceType: string,
  feature: 'logs' | 'terminal' | 'pods' | 'endpoints'
): boolean {
  const config = RESOURCE_CONFIGS[resourceType as keyof typeof RESOURCE_CONFIGS] as ResourceConfig | undefined;
  return (config?.[feature as keyof ResourceConfig] as boolean) ?? false;
}
