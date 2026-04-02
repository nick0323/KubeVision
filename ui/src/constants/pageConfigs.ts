/**
 * 页面配置 - 精简版
 */

import { ColumnDef } from '../types';

interface ExtendedColumn extends ColumnDef<any> {
  sortable?: boolean;
}

// ==================== Workloads ====================

export const PODS_CONFIG = {
  title: 'Pods',
  apiEndpoint: '/api/pods',
  resourceType: 'pods',
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
  defaultSort: { field: 'age', order: 'desc' as const },
};

export const DEPLOYMENTS_CONFIG = {
  title: 'Deployments',
  apiEndpoint: '/api/deployments',
  resourceType: 'deployments',
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
    { title: 'Up To Date', dataIndex: 'updatedReplicas', width: '10%' },
    { title: 'Available', dataIndex: 'availableReplicas', width: '10%' },
    { title: 'Restarts', dataIndex: 'restarts', width: '5%', sortable: true },
    { title: 'Age', dataIndex: 'age', width: '10%' },
  ] as ExtendedColumn[],
  namespaceFilter: true,
};

export const STATEFULSETS_CONFIG = {
  title: 'StatefulSets',
  apiEndpoint: '/api/statefulsets',
  resourceType: 'statefulsets',
  columns: [
    { title: 'Name', dataIndex: 'name', width: '25%', sortable: true },
    { title: 'Namespace', dataIndex: 'namespace', width: '15%' },
    { title: 'Status', dataIndex: 'status', width: '12%', sortable: true },
    {
      title: 'Ready',
      dataIndex: 'readyReplicas',
      width: '12%',
      render: (v: any, r: any) => `${v}/${r.desiredReplicas || 0}`,
    },
    { title: 'Restarts', dataIndex: 'restarts', width: '12%', sortable: true },
    { title: 'Age', dataIndex: 'age', width: '12%' },
  ] as ExtendedColumn[],
  namespaceFilter: true,
};

export const DAEMONSETS_CONFIG = {
  title: 'DaemonSets',
  apiEndpoint: '/api/daemonsets',
  resourceType: 'daemonsets',
  columns: [
    { title: 'Name', dataIndex: 'name', width: '25%', sortable: true },
    { title: 'Namespace', dataIndex: 'namespace', width: '15%' },
    { title: 'Status', dataIndex: 'status', width: '12%', sortable: true },
    { title: 'Desired', dataIndex: 'desiredReplicas', width: '10%' },
    { title: 'Ready', dataIndex: 'readyReplicas', width: '10%' },
    { title: 'Restarts', dataIndex: 'restarts', width: '10%', sortable: true },
    { title: 'Age', dataIndex: 'age', width: '8%' },
  ] as ExtendedColumn[],
  namespaceFilter: true,
};

export const JOBS_CONFIG = {
  title: 'Jobs',
  apiEndpoint: '/api/jobs',
  resourceType: 'jobs',
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
    { title: 'Restarts', dataIndex: 'restarts', width: '12%', sortable: true },
    { title: 'Duration', dataIndex: 'duration', width: '12%' },
    { title: 'Age', dataIndex: 'age', width: '12%' },
  ] as ExtendedColumn[],
  namespaceFilter: true,
};

export const CRONJOBS_CONFIG = {
  title: 'CronJobs',
  apiEndpoint: '/api/cronjobs',
  resourceType: 'cronjobs',
  columns: [
    { title: 'Name', dataIndex: 'name', width: '25%', sortable: true },
    { title: 'Namespace', dataIndex: 'namespace', width: '15%' },
    { title: 'Status', dataIndex: 'status', width: '12%', sortable: true },
    { title: 'Suspend', dataIndex: 'suspend', width: '10%' },
    { title: 'Schedule', dataIndex: 'schedule', width: '15%' },
    { title: 'Active', dataIndex: 'active', width: '8%' },
    { title: 'Restarts', dataIndex: 'restarts', width: '8%', sortable: true },
    { title: 'Age', dataIndex: 'age', width: '7%' },
  ] as ExtendedColumn[],
  namespaceFilter: true,
};

// ==================== Network ====================

export const SERVICES_CONFIG = {
  title: 'Services',
  apiEndpoint: '/api/services',
  resourceType: 'services',
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
  title: 'PersistentVolumeClaims',
  apiEndpoint: '/api/pvcs',
  resourceType: 'pvcs',
  columns: [
    { title: 'Name', dataIndex: 'name', width: '20%', sortable: true },
    { title: 'Namespace', dataIndex: 'namespace', width: '15%' },
    { title: 'Status', dataIndex: 'status', width: '10%', sortable: true },
    { title: 'Access Mode', dataIndex: 'accessMode', width: '15%' },
    { title: 'Volume', dataIndex: 'volumeName', width: '10%' },
    { title: 'Capacity', dataIndex: 'capacity', width: '10%' },
    { title: 'Storage Class', dataIndex: 'storageClass', width: '12%' },
    { title: 'Age', dataIndex: 'age', width: '8%' },
  ] as ExtendedColumn[],
  namespaceFilter: true,
};

export const PVS_CONFIG = {
  title: 'PersistentVolumes',
  apiEndpoint: '/api/pvs',
  resourceType: 'pvs',
  columns: [
    { title: 'Name', dataIndex: 'name', width: '25%', sortable: true },
    { title: 'Status', dataIndex: 'status', width: '10%', sortable: true },
    { title: 'Capacity', dataIndex: 'capacity', width: '10%' },
    { title: 'Access Mode', dataIndex: 'accessMode', width: '15%' },
    { title: 'Storage Class', dataIndex: 'storageClass', width: '10%' },
    { title: 'Claim', dataIndex: 'claimRef', width: '15%' },
    { title: 'Reclaim Policy', dataIndex: 'reclaimPolicy', width: '10%' },
    { title: 'Age', dataIndex: 'age', width: '5%' },
  ] as ExtendedColumn[],
  namespaceFilter: false,
};

export const STORAGECLASSES_CONFIG = {
  title: 'StorageClasses',
  apiEndpoint: '/api/storageclasses',
  resourceType: 'storageclasses',
  columns: [
    { title: 'Name', dataIndex: 'name', width: '22%', sortable: true },
    { title: 'Provisioner', dataIndex: 'provisioner', width: '25%' },
    { title: 'Reclaim Policy', dataIndex: 'reclaimPolicy', width: '15%' },
    { title: 'Binding Mode', dataIndex: 'volumeBindingMode', width: '15%' },
    { title: 'Default', dataIndex: 'isDefault', width: '8%' },
    { title: 'Age', dataIndex: 'age', width: '15%' },
  ] as ExtendedColumn[],
  namespaceFilter: false,
};

// ==================== Configuration ====================

export const CONFIGMAPS_CONFIG = {
  title: 'ConfigMaps',
  apiEndpoint: '/api/configmaps',
  resourceType: 'configmaps',
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
  apiEndpoint: '/api/secrets',
  resourceType: 'secrets',
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
  apiEndpoint: '/api/namespaces',
  resourceType: 'namespaces',
  columns: [
    { title: 'Name', dataIndex: 'name', width: '60%', sortable: true },
    { title: 'Status', dataIndex: 'status', width: '25%', sortable: true },
    { title: 'Age', dataIndex: 'age', width: '15%' },
  ] as ExtendedColumn[],
  namespaceFilter: false,
};

export const NODES_CONFIG = {
  title: 'Nodes',
  apiEndpoint: '/api/nodes',
  resourceType: 'nodes',
  columns: [
    { title: 'Name', dataIndex: 'name', width: '15%', sortable: true },
    { title: 'IP', dataIndex: 'ip', width: '15%' },
    { title: 'Role', dataIndex: 'role', width: '20%', sortable: true },
    { title: 'CPU', dataIndex: 'cpuUsage', width: '10%' },
    { title: 'Memory', dataIndex: 'memoryUsage', width: '10%' },
    { title: 'Pods', dataIndex: 'podsUsed', width: '10%' },
    { title: 'Status', dataIndex: 'status', width: '10%', sortable: true },
    { title: 'Age', dataIndex: 'age', width: '10%' },
  ] as ExtendedColumn[],
  namespaceFilter: false,
};

export const EVENTS_CONFIG = {
  title: 'Events',
  apiEndpoint: '/api/events',
  resourceType: 'events',
  columns: [
    { title: 'Type', dataIndex: 'type', width: '8%', sortable: true },
    { title: 'Reason', dataIndex: 'reason', width: '12%' },
    { title: 'Message', dataIndex: 'message', width: '25%' },
    { title: 'Object', dataIndex: 'object', width: '20%' },
    { title: 'Last Seen', dataIndex: 'lastSeen', width: '15%', sortable: true },
    { title: 'Count', dataIndex: 'count', width: '8%' },
  ] as ExtendedColumn[],
  namespaceFilter: true,
  defaultSort: { field: 'lastSeen', order: 'desc' as const },
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
