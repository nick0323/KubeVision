/**
 * CommonResource detail pageType definitions
 * Type definitions for Resource Detail Page
 */

import { K8sOwnerReference, Container as K8sContainer } from '../types/k8s-resources';

// ============================================================================
// Common Tab Props
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
  podName?: string; // Compatible with old Pod detail page
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
// page级 Props
// ============================================================================

/**
 * CommonResource detail page Props
 */
export interface ResourceDetailPageProps {
  resourceType: string;
  namespace: string;
  name: string;
  collapsed: boolean;
  onToggleCollapsed: () => void;
}

/**
 * YAML Tab Props (扩展版，兼容旧code)
 */
export interface YamlTabPropsExtended extends YamlTabProps {
  pod?: unknown | null; // Compatible with old Pod detail page
}

/**
 * Resource type config
 */
export interface ResourceConfig {
  title: string;
  tabs: string[];
  hasLogs?: boolean;
  hasTerminal?: boolean;
  hasPods?: boolean;
  hasEndpoints?: boolean;
  [key: string]: unknown; // Supports dynamic property access
}

// ============================================================================
// Resource config factoryfunction
// forreduceduplicateConfigcode
// ============================================================================

/**
 * CreateresourceConfig
 * @param title resource标题
 * @param tabs Tab List
 * @param options 额outside选项
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

// Common Tab Config常量
const COMMON_TABS = {
  BASIC: ['overview', 'yaml', 'related', 'events'] as const,
  WITH_PODS: ['overview', 'yaml', 'pods', 'related', 'events'] as const,
  POD_FULL: ['overview', 'yaml', 'logs', 'terminal', 'related', 'events'] as const,
} as const;

// ============================================================================
// resourceConfigMapping
// 定义every种 K8s resource'sDisplayConfig
// ============================================================================

export const RESOURCE_CONFIGS: Record<string, ResourceConfig> = {
  // Pod - 完整functionality
  pod: createResourceConfig('Pod', COMMON_TABS.POD_FULL, {
    hasLogs: true,
    hasTerminal: true,
  }),

  // Workloads - Support Pods
  deployment: createResourceConfig('Deployment', COMMON_TABS.WITH_PODS, { hasPods: true }),
  statefulset: createResourceConfig('StatefulSet', COMMON_TABS.WITH_PODS, { hasPods: true }),
  daemonset: createResourceConfig('DaemonSet', COMMON_TABS.WITH_PODS, { hasPods: true }),
  job: createResourceConfig('Job', COMMON_TABS.WITH_PODS, { hasPods: true }),

  // CronJob - notSupport Pods
  cronjob: createResourceConfig('CronJob', COMMON_TABS.BASIC),

  // Service - Support Endpoints
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
// helperType
// ============================================================================

/**
 * 关联resource
 */
export interface RelatedResource {
  kind: string;
  name: string;
  namespace?: string;
  apiVersion?: string;
}

/**
 * resourceType键名
 */
export type ResourceTypeKey = keyof typeof RESOURCE_CONFIGS;

/**
 * Tab 键名Type
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
 * Getresource's Tab List
 */
export function getResourceTabs(resourceType: string): string[] {
  const config = RESOURCE_CONFIGS[resourceType as keyof typeof RESOURCE_CONFIGS];
  return config?.tabs || COMMON_TABS.BASIC;
}

/**
 * 检查resourceis否Support特定functionality
 */
export function hasResourceFeature(
  resourceType: string,
  feature: 'logs' | 'terminal' | 'pods' | 'endpoints'
): boolean {
  const config = RESOURCE_CONFIGS[resourceType as keyof typeof RESOURCE_CONFIGS] as ResourceConfig | undefined;
  return (config?.[feature as keyof ResourceConfig] as boolean) ?? false;
}
