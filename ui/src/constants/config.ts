/**
 * Application Configuration Constants
 * 集中管理应用配置和魔法数字
 */

// ==================== API 配置 ====================

export const API_CONFIG = {
  BASE_URL: '/api',
  WS_URL: 'ws://localhost:8080/api/ws',
  WSS_URL: 'wss://localhost:8080/api/ws',
  TIMEOUT: 30000, // 30 秒请求超时
} as const;

// ==================== 分页配置 ====================

export const PAGINATION_CONFIG = {
  DEFAULT_PAGE_SIZE: 20,
  PAGE_SIZE_OPTIONS: [20, 50, 100, 500],
  MIN_PAGE_SIZE: 10,
  MAX_PAGE_SIZE: 1000,
} as const;

// ==================== 缓存配置 ====================

export const CACHE_CONFIG = {
  STALE_TIME: 30000, // 30 秒缓存过期时间
  RETRY_COUNT: 3,
  RETRY_DELAY: 1000,
} as const;

// ==================== 集群资源类型 ====================

/**
 * 集群级资源（不需要 namespace）
 */
export const CLUSTER_SCOPE_RESOURCES = new Set([
  'node',
  'nodes',
  'pv',
  'pvs',
  'persistentvolume',
  'persistentvolumes',
  'storageclass',
  'storageclasses',
  'namespace',
  'namespaces',
]) as ReadonlySet<string>;

/**
 * 判断是否为集群级资源
 */
export function isClusterResource(resourceType: string): boolean {
  return CLUSTER_SCOPE_RESOURCES.has(resourceType.toLowerCase());
}

// ==================== 资源类型映射 ====================

/**
 * 资源类型单复数映射
 */
export const RESOURCE_TYPE_MAP = {
  // Workloads
  pod: 'pods',
  deployment: 'deployments',
  statefulset: 'statefulsets',
  daemonset: 'daemonsets',
  job: 'jobs',
  cronjob: 'cronjobs',
  // Network
  service: 'services',
  ingress: 'ingress',
  // Storage
  pvc: 'pvcs',
  pv: 'pvs',
  storageclass: 'storageclasses',
  // Config
  configmap: 'configmaps',
  secret: 'secrets',
  // Cluster
  namespace: 'namespaces',
  node: 'nodes',
  // Events
  event: 'events',
} as const;

/**
 * 资源类型显示名称
 */
export const RESOURCE_DISPLAY_NAMES: Record<string, string> = {
  pods: 'Pods',
  deployments: 'Deployments',
  statefulsets: 'StatefulSets',
  daemonsets: 'DaemonSets',
  jobs: 'Jobs',
  cronjobs: 'CronJobs',
  services: 'Services',
  ingress: 'Ingress',
  pvcs: 'PVCs',
  pvs: 'PVs',
  storageclasses: 'StorageClasses',
  configmaps: 'ConfigMaps',
  secrets: 'Secrets',
  namespaces: 'Namespaces',
  nodes: 'Nodes',
  events: 'Events',
} as const;

// ==================== 日志配置 ====================

export const LOG_CONFIG = {
  MAX_LOG_LINES: 1000,
  LINE_HEIGHT: 23,
  OVERSCAN_ROWS: 20,
  DEFAULT_TAIL_LINES: 100,
  TAIL_LINES_OPTIONS: [100, 200, 500],
} as const;

// ==================== Shell 配置 ====================

export const SHELL_CONFIG = {
  OPTIONS: ['bash', 'sh', 'zsh'],
  DEFAULT: 'bash',
} as const;

// ==================== 时间配置 ====================

export const TIME_CONFIG = {
  REFRESH_INTERVAL: 30000, // 30 秒自动刷新
  HEARTBEAT_INTERVAL: 15000, // 15 秒心跳
  TOAST_DURATION: 3000, // 提示持续时间
} as const;

// ==================== 状态配置 ====================

/**
 * Pod 状态映射
 */
export const POD_STATUS_MAP = {
  Running: { color: 'success' as const, icon: '🟢' },
  Pending: { color: 'warning' as const, icon: '🟡' },
  Failed: { color: 'error' as const, icon: '🔴' },
  Succeeded: { color: 'default' as const, icon: '🔵' },
  Unknown: { color: 'default' as const, icon: '⚪' },
  CrashLoopBackOff: { color: 'error' as const, icon: '🔴' },
  Error: { color: 'error' as const, icon: '🔴' },
  Terminating: { color: 'warning' as const, icon: '🟡' },
} as const;

/**
 * Node 状态映射
 */
export const NODE_STATUS_MAP = {
  Ready: { color: 'success' as const, icon: '🟢' },
  NotReady: { color: 'error' as const, icon: '🔴' },
  SchedulingDisabled: { color: 'warning' as const, icon: '🟡' },
  Unknown: { color: 'default' as const, icon: '⚪' },
} as const;

// ==================== 本地存储键名 ====================

export const STORAGE_KEYS = {
  TOKEN: 'token',
  TOKEN_EXPIRY: 'token_expiry',
  TOKEN_USERNAME: 'token_username',
  REMEMBERED_USERNAME: 'remembered_username',
  SIDER_COLLAPSED: 'sider_collapsed',
  CURRENT_TAB: 'current_tab',
} as const;

// ==================== 事件类型 ====================

export const CUSTOM_EVENTS = {
  TAB_CHANGE: 'tab-change',
  AUTH_UNAUTHORIZED: 'auth-unauthorized',
} as const;
