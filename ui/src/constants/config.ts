/**
 * Application Configuration Constants
 * 集in管理App configand魔法数字
 */

// ==================== API Config ====================

export const API_CONFIG = {
  BASE_URL: '/api/v1',
  TIMEOUT: 30000,
} as const;

// ==================== PaginationConfig ====================

export const PAGINATION_CONFIG = {
  DEFAULT_PAGE_SIZE: 15,
  PAGE_SIZE_OPTIONS: [15, 30, 50, 100],
  MIN_PAGE_SIZE: 10,
  MAX_PAGE_SIZE: 1000,
} as const;

// ==================== Cache config ====================

export const CACHE_CONFIG = {
  STALE_TIME: 30000, // 30 seconds cache expiration time
  RETRY_COUNT: 3,
  RETRY_DELAY: 1000,
} as const;

// ==================== clusterresourceType ====================

/**
 * cluster级resource（not needed namespace）
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
  'clusterrole',
  'clusterroles',
  'clusterrolebinding',
  'clusterrolebindings',
]) as ReadonlySet<string>;

/**
 * Determine if cluster-level resource
 */
export function isClusterResource(resourceType: string): boolean {
  return CLUSTER_SCOPE_RESOURCES.has(resourceType.toLowerCase());
}

// ==================== resourceTypeMapping ====================

/**
 * resourceType单复数Mapping
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
  networkpolicy: 'networkpolicies',
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
  // Auth
  serviceaccount: 'serviceaccounts',
  role: 'roles',
  rolebinding: 'rolebindings',
  clusterrole: 'clusterroles',
  clusterrolebinding: 'clusterrolebindings',
  // Policy
  resourcequota: 'resourcequotas',
  limitrange: 'limitranges',
  poddisruptionbudget: 'poddisruptionbudgets',
  horizontalpodautoscaler: 'horizontalpodautoscalers',
  // Events
  event: 'events',
} as const;

/**
 * resourceTypeDisplay名称
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
  networkpolicies: 'NetworkPolicies',
  serviceaccounts: 'ServiceAccounts',
  roles: 'Roles',
  rolebindings: 'RoleBindings',
  clusterroles: 'ClusterRoles',
  clusterrolebindings: 'ClusterRoleBindings',
  resourcequotas: 'ResourceQuotas',
  limitranges: 'LimitRanges',
  poddisruptionbudgets: 'PDBs',
  horizontalpodautoscalers: 'HPA',
  events: 'Events',
} as const;

// ==================== 日志Config ====================

export const LOG_CONFIG = {
  MAX_LOG_LINES: 1000,
  LINE_HEIGHT: 23,
  OVERSCAN_ROWS: 20,
  DEFAULT_TAIL_LINES: 100,
  TAIL_LINES_OPTIONS: [100, 200, 500],
} as const;

// ==================== Shell Config ====================

export const SHELL_CONFIG = {
  OPTIONS: ['bash', 'sh', 'zsh'],
  DEFAULT: 'bash',
} as const;

export const ALWAYS_HIDDEN_FIELDS = ['managedFields', 'selfLink', 'clusterName'] as const;

export const LINES_OPTIONS = [
  { value: '100', label: '100' },
  { value: '200', label: '200' },
  { value: '500', label: '500' },
] as const;

// ==================== Time config ====================

export const TIME_CONFIG = {
  REFRESH_INTERVAL: 30000, // 30 seconds auto refresh
  HEARTBEAT_INTERVAL: 15000, // 15 seconds heartbeat
  TOAST_DURATION: 3000, // Toast duration
} as const;

// ==================== StatusConfig ====================

/**
 * Pod StatusMapping
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
 * Node StatusMapping
 */
export const NODE_STATUS_MAP = {
  Ready: { color: 'success' as const, icon: '🟢' },
  NotReady: { color: 'error' as const, icon: '🔴' },
  SchedulingDisabled: { color: 'warning' as const, icon: '🟡' },
  Unknown: { color: 'default' as const, icon: '⚪' },
} as const;

// ==================== 本地Storage keys名 ====================

export const STORAGE_KEYS = {
  TOKEN: 'token',
  TOKEN_EXPIRY: 'token_expiry',
  TOKEN_USERNAME: 'token_username',
  REFRESH_TOKEN: 'refresh_token',
  REMEMBERED_USERNAME: 'remembered_username',
  SIDER_COLLAPSED: 'sider_collapsed',
  CURRENT_TAB: 'current_tab',
  CURRENT_CLUSTER: 'current_cluster',
} as const;

// ==================== 事 componentType ====================

export const CUSTOM_EVENTS = {
  TAB_CHANGE: 'tab-change',
  AUTH_UNAUTHORIZED: 'auth-unauthorized',
} as const;
