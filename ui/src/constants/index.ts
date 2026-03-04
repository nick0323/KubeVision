/**
 * 菜单配置
 */

export interface MenuItem {
  key: string;
  label: string;
  icon: string;
}

export interface MenuGroup {
  group: string;
  items: MenuItem[];
}

export const MENU_LIST: MenuGroup[] = [
  {
    group: '',
    items: [
      { key: 'overview', label: 'Overview', icon: 'FaChartPie' },
    ],
  },
  {
    group: 'Workloads',
    items: [
      { key: 'pods', label: 'Pods', icon: 'FaCubes' },
      { key: 'deployments', label: 'Deployments', icon: 'FaLayerGroup' },
      { key: 'statefulsets', label: 'StatefulSets', icon: 'FaBoxes' },
      { key: 'daemonsets', label: 'DaemonSets', icon: 'FaBoxOpen' },
      { key: 'jobs', label: 'Jobs', icon: 'FaTasks' },
      { key: 'cronjobs', label: 'CronJobs', icon: 'FaRegClock' },
    ],
  },
  {
    group: 'Service',
    items: [
      { key: 'services', label: 'Services', icon: 'FaProjectDiagram' },
      { key: 'ingress', label: 'Ingress', icon: 'FaNetworkWired' },
    ],
  },
  {
    group: 'Storage',
    items: [
      { key: 'pvcs', label: 'PVCs', icon: 'FaHdd' },
      { key: 'pvs', label: 'PVs', icon: 'FaDatabase' },
      { key: 'storageclasses', label: 'StorageClasses', icon: 'FaCog' },
    ],
  },
  {
    group: 'Config',
    items: [
      { key: 'configmaps', label: 'ConfigMaps', icon: 'FaCog' },
      { key: 'secrets', label: 'Secrets', icon: 'FaKey' },
    ],
  },
  {
    group: 'Cluster',
    items: [
      { key: 'nodes', label: 'Nodes', icon: 'FaDesktop' },
      { key: 'namespaces', label: 'Namespaces', icon: 'LuSquareDashed' },
      { key: 'events', label: 'Events', icon: 'FaBell' },
    ],
  },
];

// API 端点映射
export const API_MAP: Record<string, string> = {
  overview: '/api/overview',
  pods: '/api/pods',
  deployments: '/api/deployments',
  statefulsets: '/api/statefulsets',
  daemonsets: '/api/daemonsets',
  jobs: '/api/jobs',
  cronjobs: '/api/cronjobs',
  services: '/api/services',
  ingress: '/api/ingress',
  pvcs: '/api/pvcs',
  pvs: '/api/pvs',
  storageclasses: '/api/storageclasses',
  configmaps: '/api/configmaps',
  secrets: '/api/secrets',
  nodes: '/api/nodes',
  namespaces: '/api/namespaces',
  events: '/api/events',
};

// 常量
export const PAGE_SIZE = 10;
export const PAGE_SIZE_OPTIONS = [10, 15, 20];
export const SEARCH_PLACEHOLDER = '搜索名称、标签...';
export const EMPTY_TEXT = '暂无数据';
export const LOADING_TEXT = '加载中...';
export const ERROR_TEXT = '错误：';
