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
      { key: 'clusters', label: 'Clusters', icon: 'FaProjectDiagram' },
      { key: 'nodes', label: 'Nodes', icon: 'FaServer' },
      { key: 'namespaces', label: 'Namespaces', icon: 'FaThLarge' },
      { key: 'events', label: 'Events', icon: 'FaBell' },
    ],
  },
  {
    group: 'Policy',
    items: [
      { key: 'networkpolicies', label: 'NetworkPolicies', icon: 'FaShieldAlt' },
      { key: 'resourcequotas', label: 'ResourceQuotas', icon: 'FaTachometerAlt' },
      { key: 'limitranges', label: 'LimitRanges', icon: 'FaSlidersH' },
      { key: 'poddisruptionbudgets', label: 'PDBs', icon: 'FaBalanceScale' },
    ],
  },
  {
    group: 'Others',
    items: [
      { key: 'horizontalpodautoscalers', label: 'HPA', icon: 'FaArrowsAltV' },
      { key: 'serviceaccounts', label: 'ServiceAccounts', icon: 'FaUserSecret' },
      { key: 'roles', label: 'Roles', icon: 'FaUserTag' },
      { key: 'rolebindings', label: 'RoleBindings', icon: 'FaUserCheck' },
      { key: 'clusterroles', label: 'ClusterRoles', icon: 'FaUserTag' },
      { key: 'clusterrolebindings', label: 'ClusterRoleBindings', icon: 'FaUserCheck' },
      { key: 'crds', label: 'CRDs', icon: 'FaPuzzlePiece' },
    ],
  },
  {
    group: 'GitOps',
    items: [
      { key: 'argocd', label: 'ArgoCD', icon: 'FaGitAlt' },
    ],
  },
];

// Re-export from config.ts
export * from './config';
