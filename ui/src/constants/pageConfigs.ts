/**
 * Page Configurations
 * Use工厂functionCreateresourceConfig，reduceduplicatecode
 */

import {
  createPodConfig,
  createWorkloadConfig,
  createServiceConfig,
  createNodeConfig,
  createClusterResourceConfig,
  createBaseConfig,
  createNameColumn,
  createNamespaceColumn,
  createStatusColumn,
  createAgeColumn,
  finalizeConfig,
  type ExtendedColumn,
} from '../utils/resourceConfigFactory';

// ==================== Workloads ====================

export const PODS_CONFIG = createPodConfig();

export const DEPLOYMENTS_CONFIG = createWorkloadConfig({
  title: 'Deployments',
  apiEndpoint: '/api/deployment',
  resourceType: 'deployment',
  extraColumns: [
    { title: 'UpToDate', dataIndex: 'updatedReplicas', width: '10%' },
    { title: 'Available', dataIndex: 'availableReplicas', width: '10%' },
  ],
});

export const STATEFULSETS_CONFIG = createWorkloadConfig({
  title: 'StatefulSets',
  apiEndpoint: '/api/statefulset',
  resourceType: 'statefulset',
});

export const DAEMONSETS_CONFIG = createWorkloadConfig({
  title: 'DaemonSets',
  apiEndpoint: '/api/daemonset',
  resourceType: 'daemonset',
  extraColumns: [
    { title: 'Desired', dataIndex: 'desiredReplicas', width: '10%' },
  ],
});

export const JOBS_CONFIG = finalizeConfig({
  ...createBaseConfig({
    title: 'Jobs',
    apiEndpoint: '/api/job',
    resourceType: 'job',
  }),
  columns: [
    createNameColumn('30%'),
    createNamespaceColumn('15%'),
    createStatusColumn('15%'),
    {
      title: 'Completions',
      dataIndex: 'succeeded',
      width: '13%',
      render: (v: unknown, r: Record<string, unknown>) => `${v || 0}/${(r as { completions?: number }).completions || 1}`,
      sortable: false,
    },
    { title: 'Duration', dataIndex: 'duration', width: '12%', sortable: false },
    createAgeColumn('15%'),
  ],
});

export const CRONJOBS_CONFIG = finalizeConfig({
  ...createBaseConfig({
    title: 'CronJobs',
    apiEndpoint: '/api/cronjob',
    resourceType: 'cronjob',
  }),
  columns: [
    createNameColumn('30%'),
    createNamespaceColumn('15%'),
    createStatusColumn('15%'),
    { title: 'Suspend', dataIndex: 'suspend', width: '12%', sortable: false },
    { title: 'Schedule', dataIndex: 'schedule', width: '15%', sortable: false },
    { title: 'Active', dataIndex: 'active', width: '8%', sortable: false },
    createAgeColumn('5%'),
  ],
});

// ==================== Network ====================

export const SERVICES_CONFIG = createServiceConfig();

export const INGRESS_CONFIG = finalizeConfig({
  ...createBaseConfig({
    title: 'Ingress',
    apiEndpoint: '/api/ingress',
    resourceType: 'ingress',
  }),
  columns: [
    createNameColumn('20%'),
    createNamespaceColumn('10%'),
    { title: 'Class', dataIndex: 'class', width: '10%', sortable: false },
    { title: 'Hosts', dataIndex: 'hosts', width: '15%', sortable: false },
    {
      title: 'Target Service',
      dataIndex: 'targetService',
      width: '15%',
      render: (v: unknown) => (Array.isArray(v) ? [...new Set(v)].join(', ') : (v as string) || '-'),
      sortable: false,
    },
    { title: 'Path', dataIndex: 'path', width: '25%', sortable: false },
    createAgeColumn('5%'),
  ],
});

// ==================== Storage ====================

export const PVCS_CONFIG = finalizeConfig({
  ...createBaseConfig({
    title: 'PersistentVolumeClaim',
    apiEndpoint: '/api/pvc',
    resourceType: 'pvc',
  }),
  columns: [
    createNameColumn('20%'),
    createNamespaceColumn('15%'),
    createStatusColumn('10%'),
    { title: 'AccessMode', dataIndex: 'accessMode', width: '15%' },
    { title: 'Volume', dataIndex: 'volumeName', width: '10%' },
    { title: 'Capacity', dataIndex: 'capacity', width: '10%' },
    { title: 'StorageClass', dataIndex: 'storageClass', width: '12%' },
    createAgeColumn('8%'),
  ],
});

export const PVS_CONFIG = createClusterResourceConfig({
  title: 'PersistentVolume',
  apiEndpoint: '/api/pv',
  resourceType: 'pv',
  columns: [
    createNameColumn('20%'),
    createStatusColumn('10%'),
    { title: 'Capacity', dataIndex: 'capacity', width: '10%' },
    { title: 'AccessMode', dataIndex: 'accessMode', width: '15%' },
    { title: 'StorageClass', dataIndex: 'storageClass', width: '10%' },
    { title: 'Claim', dataIndex: 'claimRef', width: '15%' },
    { title: 'ReclaimPolicy', dataIndex: 'reclaimPolicy', width: '15%' },
    createAgeColumn('5%'),
  ],
});

export const STORAGECLASSES_CONFIG = createClusterResourceConfig({
  title: 'StorageClasses',
  apiEndpoint: '/api/storageclass',
  resourceType: 'storageclass',
  columns: [
    createNameColumn('22%'),
    { title: 'Provisioner', dataIndex: 'provisioner', width: '25%' },
    { title: 'ReclaimPolicy', dataIndex: 'reclaimPolicy', width: '15%' },
    { title: 'BindingMode', dataIndex: 'volumeBindingMode', width: '15%' },
    { title: 'Default', dataIndex: 'isDefault', width: '8%' },
    createAgeColumn('15%'),
  ],
});

// ==================== Configuration ====================

export const CONFIGMAPS_CONFIG = finalizeConfig({
  ...createBaseConfig({
    title: 'ConfigMaps',
    apiEndpoint: '/api/configmap',
    resourceType: 'configmap',
  }),
  columns: [
    createNameColumn('30%'),
    createNamespaceColumn('20%'),
    { title: 'Data Count', dataIndex: 'dataCount', width: '10%' },
    { title: 'Keys', dataIndex: 'keys', width: '30%' },
    createAgeColumn('10%'),
  ],
});

export const SECRETS_CONFIG = finalizeConfig({
  ...createBaseConfig({
    title: 'Secrets',
    apiEndpoint: '/api/secret',
    resourceType: 'secret',
  }),
  columns: [
    createNameColumn('30%'),
    createNamespaceColumn('20%'),
    { title: 'Type', dataIndex: 'type', width: '25%' },
    { title: 'Data Count', dataIndex: 'dataCount', width: '10%' },
    createAgeColumn('15%'),
  ],
});

// ==================== Auto Scaling ====================

export const HPAS_CONFIG = finalizeConfig({
  ...createBaseConfig({
    title: 'HPA',
    apiEndpoint: '/api/horizontalpodautoscaler',
    resourceType: 'horizontalpodautoscaler',
  }),
  columns: [
    createNameColumn('22%'),
    createNamespaceColumn('12%'),
    { title: 'Min', dataIndex: 'minReplicas', width: '6%', sortable: true },
    { title: 'Max', dataIndex: 'maxReplicas', width: '6%', sortable: true },
    { title: 'Current', dataIndex: 'currentReplicas', width: '8%', sortable: true },
    { title: 'Desired', dataIndex: 'desiredReplicas', width: '8%', sortable: true },
    { title: 'Metrics', dataIndex: 'metrics', width: '33%', sortable: false },
    createAgeColumn('5%'),
  ],
});

// ==================== Network Policies ====================

export const NETWORKPOLICIES_CONFIG = finalizeConfig({
  ...createBaseConfig({
    title: 'NetworkPolicies',
    apiEndpoint: '/api/networkpolicy',
    resourceType: 'networkpolicy',
  }),
  columns: [
    createNameColumn('22%'),
    createNamespaceColumn('12%'),
    { title: 'PodSelector', dataIndex: 'podSelector', width: '25%', sortable: false },
    { title: 'PolicyTypes', dataIndex: 'policyTypes', width: '18%', sortable: false, render: (v: unknown) => (Array.isArray(v) ? v.join(', ') : (v as string) || '-') },
    createAgeColumn('5%'),
  ],
});

// ==================== Service Accounts ====================

export const SERVICEACCOUNTS_CONFIG = finalizeConfig({
  ...createBaseConfig({
    title: 'ServiceAccounts',
    apiEndpoint: '/api/serviceaccount',
    resourceType: 'serviceaccount',
  }),
  columns: [
    createNameColumn('30%'),
    createNamespaceColumn('20%'),
    { title: 'Secrets', dataIndex: 'secrets', width: '15%', sortable: true },
    createAgeColumn('15%'),
  ],
});

// ==================== RBAC - Roles ====================

export const ROLES_CONFIG = finalizeConfig({
  ...createBaseConfig({
    title: 'Roles',
    apiEndpoint: '/api/role',
    resourceType: 'role',
  }),
  columns: [
    createNameColumn('30%'),
    createNamespaceColumn('20%'),
    { title: 'Rules', dataIndex: 'rules', width: '15%', sortable: true },
    createAgeColumn('15%'),
  ],
});

export const ROLEBINDINGS_CONFIG = finalizeConfig({
  ...createBaseConfig({
    title: 'RoleBindings',
    apiEndpoint: '/api/rolebinding',
    resourceType: 'rolebinding',
  }),
  columns: [
    createNameColumn('22%'),
    createNamespaceColumn('12%'),
    { title: 'RoleRef', dataIndex: 'roleRef', width: '22%', sortable: false },
    { title: 'Subjects', dataIndex: 'subjects', width: '22%', sortable: false },
    createAgeColumn('5%'),
  ],
});

export const CLUSTERROLES_CONFIG = createClusterResourceConfig({
  title: 'ClusterRoles',
  apiEndpoint: '/api/clusterrole',
  resourceType: 'clusterrole',
  columns: [
    createNameColumn('40%'),
    { title: 'Rules', dataIndex: 'rules', width: '20%', sortable: true },
    createAgeColumn('15%'),
  ],
});

export const CLUSTERROLEBINDINGS_CONFIG = createClusterResourceConfig({
  title: 'ClusterRoleBindings',
  apiEndpoint: '/api/clusterrolebinding',
  resourceType: 'clusterrolebinding',
  columns: [
    createNameColumn('30%'),
    { title: 'RoleRef', dataIndex: 'roleRef', width: '25%', sortable: false },
    { title: 'Subjects', dataIndex: 'subjects', width: '25%', sortable: false },
    createAgeColumn('10%'),
  ],
});

// ==================== Resource Quotas ====================

export const RESOURCEQUOTAS_CONFIG = finalizeConfig({
  ...createBaseConfig({
    title: 'ResourceQuotas',
    apiEndpoint: '/api/resourcequota',
    resourceType: 'resourcequota',
  }),
  columns: [
    createNameColumn('25%'),
    createNamespaceColumn('15%'),
    { title: 'Requests', dataIndex: 'requests', width: '20%', sortable: false },
    { title: 'Limits', dataIndex: 'limits', width: '20%', sortable: false },
    createAgeColumn('10%'),
  ],
});

// ==================== Limit Ranges ====================

export const LIMITRANGES_CONFIG = finalizeConfig({
  ...createBaseConfig({
    title: 'LimitRanges',
    apiEndpoint: '/api/limitrange',
    resourceType: 'limitrange',
  }),
  columns: [
    createNameColumn('30%'),
    createNamespaceColumn('20%'),
    { title: 'Limits', dataIndex: 'limits', width: '30%', sortable: false },
    createAgeColumn('10%'),
  ],
});

// ==================== Pod Disruption Budgets ====================

export const PODDISRUPTIONBUDGETS_CONFIG = finalizeConfig({
  ...createBaseConfig({
    title: 'PodDisruptionBudgets',
    apiEndpoint: '/api/poddisruptionbudget',
    resourceType: 'poddisruptionbudget',
  }),
  columns: [
    createNameColumn('18%'),
    createNamespaceColumn('12%'),
    { title: 'MinAvailable', dataIndex: 'minAvailable', width: '12%', sortable: false },
    { title: 'MaxUnavailable', dataIndex: 'maxUnavailable', width: '14%', sortable: false },
    { title: 'Current', dataIndex: 'currentHealthy', width: '8%', sortable: true },
    { title: 'Desired', dataIndex: 'desiredHealthy', width: '8%', sortable: true },
    createAgeColumn('5%'),
  ],
});

// ==================== Cluster ====================

export const NAMESPACES_CONFIG = createClusterResourceConfig({
  title: 'Namespaces',
  apiEndpoint: '/api/namespace',
  resourceType: 'namespace',
  columns: [
    createNameColumn('60%'),
    createStatusColumn('25%'),
    createAgeColumn('15%'),
  ],
});

export const NODES_CONFIG = createNodeConfig();

export const EVENTS_CONFIG = finalizeConfig({
  ...createBaseConfig({
    title: 'Events',
    apiEndpoint: '/api/event',
    resourceType: 'event',
  }),
  columns: [
    { title: 'Type', dataIndex: 'type', width: '10%', sortable: true },
    { title: 'Reason', dataIndex: 'reason', width: '10%', sortable: false },
    { title: 'Message', dataIndex: 'message', width: '20%', sortable: false },
    { title: 'Object', dataIndex: 'object', width: '20%', sortable: false },
    createNamespaceColumn('15%'),
    { title: 'Last Seen', dataIndex: 'lastSeen', width: '20%', sortable: true },
    { title: 'Count', dataIndex: 'count', width: '5%', sortable: false },
  ],
  defaultSort: { field: 'lastSeen', order: 'asc' },
});

// ==================== Export All Configs ====================

export const PAGE_CONFIGS = {
  pods: PODS_CONFIG,
  deployments: DEPLOYMENTS_CONFIG,
  statefulsets: STATEFULSETS_CONFIG,
  daemonsets: DAEMONSETS_CONFIG,
  jobs: JOBS_CONFIG,
  cronjobs: CRONJOBS_CONFIG,
  services: SERVICES_CONFIG,
  ingress: INGRESS_CONFIG,
  networkpolicies: NETWORKPOLICIES_CONFIG,
  serviceaccounts: SERVICEACCOUNTS_CONFIG,
  roles: ROLES_CONFIG,
  rolebindings: ROLEBINDINGS_CONFIG,
  clusterroles: CLUSTERROLES_CONFIG,
  clusterrolebindings: CLUSTERROLEBINDINGS_CONFIG,
  resourcequotas: RESOURCEQUOTAS_CONFIG,
  limitranges: LIMITRANGES_CONFIG,
  poddisruptionbudgets: PODDISRUPTIONBUDGETS_CONFIG,
  pvcs: PVCS_CONFIG,
  pvs: PVS_CONFIG,
  storageclasses: STORAGECLASSES_CONFIG,
  configmaps: CONFIGMAPS_CONFIG,
  secrets: SECRETS_CONFIG,
  namespaces: NAMESPACES_CONFIG,
  nodes: NODES_CONFIG,
  events: EVENTS_CONFIG,
  horizontalpodautoscalers: HPAS_CONFIG,
};
