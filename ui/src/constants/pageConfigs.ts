import { ColumnDef } from '../types';

interface ExtendedColumn extends ColumnDef<any> {
  sortable?: boolean;
}

// ==================== Workloads ====================

export const PODS_CONFIG = {
  title: 'Pods',
  apiEndpoint: '/api/pod',
  resourceType: 'pod',
  columns: [
    { title: 'Name', dataIndex: 'name', width: '25%', sortable: true },
    { title: 'Namespace', dataIndex: 'namespace', width: '15%', sortable: false },
    { title: 'Status', dataIndex: 'status', width: '15%', sortable: true },
    { title: 'Ready', dataIndex: 'ready', width: '8%' },
    { title: 'Restarts', dataIndex: 'restarts', width: '8%', sortable: true },
    { title: 'IP', dataIndex: 'podIP', width: '11%' },
    { title: 'Node', dataIndex: 'nodeName', width: '10%' },
    { title: 'Age', dataIndex: 'age', width: '8%', sortable: true },
  ] as ExtendedColumn[],
  namespaceFilter: true,
  defaultSort: { field: 'name', order: 'asc' as const },
};

export const DEPLOYMENTS_CONFIG = {
  title: 'Deployments',
  apiEndpoint: '/api/deployment',
  resourceType: 'deployment',
  columns: [
    { title: 'Name', dataIndex: 'name', width: '25%', sortable: true },
    { title: 'Namespace', dataIndex: 'namespace', width: '15%' },
    { title: 'Status', dataIndex: 'status', width: '15%', sortable: true },
    {
      title: 'Ready',
      dataIndex: 'readyReplicas',
      width: '10%',
      render: (v: any, r: any) => `${v}/${r.desiredReplicas || 0}`,
    },
    { title: 'UpToDate', dataIndex: 'updatedReplicas', width: '10%' },
    { title: 'Available', dataIndex: 'availableReplicas', width: '10%' },
    { title: 'Age', dataIndex: 'age', width: '10%' },
  ] as ExtendedColumn[],
  namespaceFilter: true,
};

export const STATEFULSETS_CONFIG = {
  title: 'StatefulSets',
  apiEndpoint: '/api/statefulset',
  resourceType: 'statefulset',
  columns: [
    { title: 'Name', dataIndex: 'name', width: '40%', sortable: true },
    { title: 'Namespace', dataIndex: 'namespace', width: '15%' },
    { title: 'Status', dataIndex: 'status', width: '15%', sortable: true },
    {
      title: 'Ready',
      dataIndex: 'readyReplicas',
      width: '15%',
      render: (v: any, r: any) => `${v}/${r.desiredReplicas || 0}`,
    },
    { title: 'Age', dataIndex: 'age', width: '15%' },
  ] as ExtendedColumn[],
  namespaceFilter: true,
};

export const DAEMONSETS_CONFIG = {
  title: 'DaemonSets',
  apiEndpoint: '/api/daemonset',
  resourceType: 'daemonset',
  columns: [
    { title: 'Name', dataIndex: 'name', width: '25%', sortable: true },
    { title: 'Namespace', dataIndex: 'namespace', width: '15%' },
    { title: 'Status', dataIndex: 'status', width: '12%', sortable: true },
    { title: 'Desired', dataIndex: 'desiredReplicas', width: '10%' },
    { title: 'Ready', dataIndex: 'readyReplicas', width: '10%' },
    { title: 'Age', dataIndex: 'age', width: '8%' },
  ] as ExtendedColumn[],
  namespaceFilter: true,
};

export const JOBS_CONFIG = {
  title: 'Jobs',
  apiEndpoint: '/api/job',
  resourceType: 'job',
  columns: [
    { title: 'Name', dataIndex: 'name', width: '25%', sortable: true },
    { title: 'Namespace', dataIndex: 'namespace', width: '15%' },
    { title: 'Status', dataIndex: 'status', width: '12%', sortable: true },
    {
      title: 'Completions',
      dataIndex: 'succeeded',
      width: '12%',
      render: (v: any, r: any) => `${v || 0}/${r.completions || 1}`,
    },
    { title: 'Duration', dataIndex: 'duration', width: '12%' },
    { title: 'Age', dataIndex: 'age', width: '12%' },
  ] as ExtendedColumn[],
  namespaceFilter: true,
};

export const CRONJOBS_CONFIG = {
  title: 'CronJobs',
  apiEndpoint: '/api/cronjob',
  resourceType: 'cronjob',
  columns: [
    { title: 'Name', dataIndex: 'name', width: '25%', sortable: true },
    { title: 'Namespace', dataIndex: 'namespace', width: '15%' },
    { title: 'Status', dataIndex: 'status', width: '12%', sortable: true },
    { title: 'Suspend', dataIndex: 'suspend', width: '10%' },
    { title: 'Schedule', dataIndex: 'schedule', width: '15%' },
    { title: 'Active', dataIndex: 'active', width: '8%' },
    { title: 'Age', dataIndex: 'age', width: '7%' },
  ] as ExtendedColumn[],
  namespaceFilter: true,
};

// ==================== Network ====================

export const SERVICES_CONFIG = {
  title: 'Services',
  apiEndpoint: '/api/service',
  resourceType: 'service',
  columns: [
    { title: 'Name', dataIndex: 'name', width: '25%', sortable: true },
    { title: 'Namespace', dataIndex: 'namespace', width: '18%' },
    { title: 'Type', dataIndex: 'type', width: '15%' },
    { title: 'Cluster IP', dataIndex: 'clusterIP', width: '18%' },
    { title: 'Ports', dataIndex: 'ports', width: '16%' },
    { title: 'Age', dataIndex: 'age', width: '8%' },
  ] as ExtendedColumn[],
  namespaceFilter: true,
};

export const INGRESS_CONFIG = {
  title: 'Ingress',
  apiEndpoint: '/api/ingress',
  resourceType: 'ingress',
  columns: [
    { title: 'Name', dataIndex: 'name', width: '20%', sortable: true },
    { title: 'Namespace', dataIndex: 'namespace', width: '10%' },
    { title: 'Class', dataIndex: 'class', width: '10%' },
    { title: 'Hosts', dataIndex: 'hosts', width: '15%' },
    {
      title: 'Target Service',
      dataIndex: 'targetService',
      width: '15%',
      render: (v: any) => (Array.isArray(v) ? [...new Set(v)].join(', ') : v || '-'),
    },
    { title: 'Path', dataIndex: 'path', width: '25%' },
    { title: 'Age', dataIndex: 'age', width: '5%' },
  ] as ExtendedColumn[],
  namespaceFilter: true,
};

// ==================== Storage ====================

export const PVCS_CONFIG = {
  title: 'PersistentVolumeClaim',
  apiEndpoint: '/api/pvcs',
  resourceType: 'pvc',
  columns: [
    { title: 'Name', dataIndex: 'name', width: '20%', sortable: true },
    { title: 'Namespace', dataIndex: 'namespace', width: '15%' },
    { title: 'Status', dataIndex: 'status', width: '10%', sortable: true },
    { title: 'AccessMode', dataIndex: 'accessMode', width: '15%' },
    { title: 'Volume', dataIndex: 'volumeName', width: '10%' },
    { title: 'Capacity', dataIndex: 'capacity', width: '10%' },
    { title: 'StorageClass', dataIndex: 'storageClass', width: '12%' },
    { title: 'Age', dataIndex: 'age', width: '8%' },
  ] as ExtendedColumn[],
  namespaceFilter: true,
};

export const PVS_CONFIG = {
  title: 'PersistentVolume',
  apiEndpoint: '/api/pvs',
  resourceType: 'pv',
  columns: [
    { title: 'Name', dataIndex: 'name', width: '20%', sortable: true },
    { title: 'Status', dataIndex: 'status', width: '10%', sortable: true },
    { title: 'Capacity', dataIndex: 'capacity', width: '10%' },
    { title: 'AccessMode', dataIndex: 'accessMode', width: '15%' },
    { title: 'StorageClass', dataIndex: 'storageClass', width: '10%' },
    { title: 'Claim', dataIndex: 'claimRef', width: '15%' },
    { title: 'ReclaimPolicy', dataIndex: 'reclaimPolicy', width: '15%' },
    { title: 'Age', dataIndex: 'age', width: '5%' },
  ] as ExtendedColumn[],
  namespaceFilter: false,
};

export const STORAGECLASSES_CONFIG = {
  title: 'StorageClasses',
  apiEndpoint: '/api/storageclass',
  resourceType: 'storageclass',
  columns: [
    { title: 'Name', dataIndex: 'name', width: '22%', sortable: true },
    { title: 'Provisioner', dataIndex: 'provisioner', width: '25%' },
    { title: 'ReclaimPolicy', dataIndex: 'reclaimPolicy', width: '15%' },
    { title: 'BindingMode', dataIndex: 'volumeBindingMode', width: '15%' },
    { title: 'Default', dataIndex: 'isDefault', width: '8%' },
    { title: 'Age', dataIndex: 'age', width: '15%' },
  ] as ExtendedColumn[],
  namespaceFilter: false,
};

// ==================== Configuration ====================

export const CONFIGMAPS_CONFIG = {
  title: 'ConfigMaps',
  apiEndpoint: '/api/configmap',
  resourceType: 'configmap',
  columns: [
    { title: 'Name', dataIndex: 'name', width: '30%', sortable: true },
    { title: 'Namespace', dataIndex: 'namespace', width: '20%' },
    { title: 'Data Count', dataIndex: 'dataCount', width: '10%' },
    { title: 'Keys', dataIndex: 'keys', width: '30%' },
    { title: 'Age', dataIndex: 'age', width: '10%' },
  ] as ExtendedColumn[],
  namespaceFilter: true,
};

export const SECRETS_CONFIG = {
  title: 'Secrets',
  apiEndpoint: '/api/secret',
  resourceType: 'secret',
  columns: [
    { title: 'Name', dataIndex: 'name', width: '25%', sortable: true },
    { title: 'Namespace', dataIndex: 'namespace', width: '20%' },
    { title: 'Type', dataIndex: 'type', width: '20%' },
    { title: 'Data Count', dataIndex: 'dataCount', width: '10%' },
    { title: 'Age', dataIndex: 'age', width: '15%' },
  ] as ExtendedColumn[],
  namespaceFilter: true,
};

// ==================== Cluster ====================

export const NAMESPACES_CONFIG = {
  title: 'Namespaces',
  apiEndpoint: '/api/namespace',
  resourceType: 'namespace',
  columns: [
    { title: 'Name', dataIndex: 'name', width: '60%', sortable: true },
    { title: 'Status', dataIndex: 'status', width: '25%', sortable: true },
    { title: 'Age', dataIndex: 'age', width: '15%' },
  ] as ExtendedColumn[],
  namespaceFilter: false,
};

export const NODES_CONFIG = {
  title: 'Nodes',
  apiEndpoint: '/api/node',
  resourceType: 'node',
  columns: [
    { title: 'Name', dataIndex: 'name', width: '15%', sortable: true },
    { title: 'IP', dataIndex: 'ip', width: '15%' },
    { title: 'Role', dataIndex: 'role', width: '20%', sortable: true },
    { title: 'CPU', dataIndex: 'cpuUsage', width: '10%' },
    { title: 'Memory', dataIndex: 'memoryUsage', width: '10%' },
    {
      title: 'Pods',
      dataIndex: 'podsUsed',
      width: '10%',
      render: (value: any, record: any) => `${value}/${record.podsCapacity || 0}`,
    },
    { title: 'Status', dataIndex: 'status', width: '10%', sortable: true },
    { title: 'Age', dataIndex: 'age', width: '10%' },
  ] as ExtendedColumn[],
  namespaceFilter: false,
  defaultSort: { field: 'name', order: 'asc' as const },
};

export const EVENTS_CONFIG = {
  title: 'Events',
  apiEndpoint: '/api/event',
  resourceType: 'event',
  columns: [
    { title: 'Type', dataIndex: 'type', width: '8%', sortable: true },
    { title: 'Reason', dataIndex: 'reason', width: '12%' },
    { title: 'Message', dataIndex: 'message', width: '25%' },
    { title: 'Object', dataIndex: 'object', width: '20%' },
    { title: 'Namespace', dataIndex: 'namespace', width: '10%' },
    { title: 'Last Seen', dataIndex: 'lastSeen', width: '20%', sortable: true },
    { title: 'Count', dataIndex: 'count', width: '5%' },
  ] as ExtendedColumn[],
  namespaceFilter: true,
  defaultSort: { field: 'lastSeen', order: 'asc' as const },
};

// ==================== 导出所有配置 ====================

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
