/**
 * Page Configurations
 * 使用工厂函数创建资源配置，减少重复代码
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
      render: (v: any, r: any) => `${v || 0}/${r.completions || 1}`,
    },
    { title: 'Duration', dataIndex: 'duration', width: '12%' },
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
    { title: 'Suspend', dataIndex: 'suspend', width: '12%' },
    { title: 'Schedule', dataIndex: 'schedule', width: '15%' },
    { title: 'Active', dataIndex: 'active', width: '8%' },
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
    { title: 'Class', dataIndex: 'class', width: '10%' },
    { title: 'Hosts', dataIndex: 'hosts', width: '15%' },
    {
      title: 'Target Service',
      dataIndex: 'targetService',
      width: '15%',
      render: (v: any) => (Array.isArray(v) ? [...new Set(v)].join(', ') : v || '-'),
    },
    { title: 'Path', dataIndex: 'path', width: '25%' },
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
    { title: 'Reason', dataIndex: 'reason', width: '10%' },
    { title: 'Message', dataIndex: 'message', width: '20%' },
    { title: 'Object', dataIndex: 'object', width: '20%' },
    createNamespaceColumn('15%'),
    { title: 'Last Seen', dataIndex: 'lastSeen', width: '20%', sortable: true },
    { title: 'Count', dataIndex: 'count', width: '5%' },
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
  pvcs: PVCS_CONFIG,
  pvs: PVS_CONFIG,
  storageclasses: STORAGECLASSES_CONFIG,
  configmaps: CONFIGMAPS_CONFIG,
  secrets: SECRETS_CONFIG,
  namespaces: NAMESPACES_CONFIG,
  nodes: NODES_CONFIG,
  events: EVENTS_CONFIG,
};
