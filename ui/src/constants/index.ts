/**
 * Constants Definition
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

// Menu Configuration
export const MENU_LIST: MenuGroup[] = [
  {
    group: '',
    items: [{ key: 'overview', label: 'Overview', icon: 'FaChartPie' }],
  },
  {
    group: 'Workloads',
    items: [
      { key: 'pods', label: 'Pods', icon: 'FaCube' },
      { key: 'deployments', label: 'Deployments', icon: 'FaRocket' },
      { key: 'statefulsets', label: 'StatefulSets', icon: 'FaTree' },
      { key: 'daemonsets', label: 'DaemonSets', icon: 'FaCogs' },
      { key: 'jobs', label: 'Jobs', icon: 'FaBriefcase' },
      { key: 'cronjobs', label: 'CronJobs', icon: 'FaClock' },
    ],
  },
  {
    group: 'Network',
    items: [
      { key: 'services', label: 'Services', icon: 'FaNetworkWired' },
      { key: 'ingress', label: 'Ingress', icon: 'FaDoorOpen' },
    ],
  },
  {
    group: 'Config',
    items: [
      { key: 'configmaps', label: 'ConfigMaps', icon: 'FaFileAlt' },
      { key: 'secrets', label: 'Secrets', icon: 'FaLock' },
    ],
  },
  {
    group: 'Storage',
    items: [
      { key: 'pvcs', label: 'PVCs', icon: 'FaHdd' },
      { key: 'pvs', label: 'PVs', icon: 'FaDatabase' },
      { key: 'storageclasses', label: 'StorageClasses', icon: 'FaListAlt' },
    ],
  },
  {
    group: 'Cluster',
    items: [
      { key: 'nodes', label: 'Nodes', icon: 'FaServer' },
      { key: 'namespaces', label: 'Namespaces', icon: 'FaThLarge' },
      { key: 'events', label: 'Events', icon: 'FaBell' },
    ],
  },
];

// Pagination Configuration
export const DEFAULT_PAGE_SIZE = 10;
export const PAGE_SIZE_OPTIONS = [10, 15, 20, 50];
