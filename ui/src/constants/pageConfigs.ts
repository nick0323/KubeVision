/**
 * 页面配置
 */
import { Column } from '../types';

// Pods 页面配置
export const PODS_CONFIG = {
  title: 'Pods',
  apiEndpoint: '/api/pods',
  resourceType: 'pods',
  columns: [
    { title: 'Name', dataIndex: 'name', width: '25%' },
    { title: 'Namespace', dataIndex: 'namespace', width: '15%' },
    { title: 'Status', dataIndex: 'status', width: '10%' },
    { title: 'CPU Usage', dataIndex: 'cpuUsage', width: '10%' },
    { title: 'Memory Usage', dataIndex: 'memoryUsage', width: '12%' },
    { title: 'PodIP', dataIndex: 'podIP', width: '13%' },
    { title: 'Node', dataIndex: 'nodeName', width: '15%' },
  ] as Column<any>[],
  namespaceFilter: true,
};

// Deployments 页面配置
export const DEPLOYMENTS_CONFIG = {
  title: 'Deployments',
  apiEndpoint: '/api/deployments',
  resourceType: 'deployments',
  columns: [
    { title: 'Name', dataIndex: 'name', width: '250' },
    { title: 'Namespace', dataIndex: 'namespace', width: '150' },
    { title: 'Available', dataIndex: 'availableReplicas', width: '100' },
    { title: 'Desired', dataIndex: 'desiredReplicas', width: '100' },
    { title: 'Status', dataIndex: 'status', width: '100' },
  ] as Column<any>[],
  namespaceFilter: true,
};

// StatefulSets 页面配置
export const STATEFULSETS_CONFIG = {
  title: 'StatefulSets',
  apiEndpoint: '/api/statefulsets',
  resourceType: 'statefulsets',
  columns: [
    { title: 'Name', dataIndex: 'name', width: 250 },
    { title: 'Namespace', dataIndex: 'namespace', width: 150 },
    { title: 'Available', dataIndex: 'availableReplicas', width: 100 },
    { title: 'Desired', dataIndex: 'desiredReplicas', width: 100 },
    { title: 'Status', dataIndex: 'status', width: 100 },
  ] as Column<any>[],
  namespaceFilter: true,
};

// DaemonSets 页面配置
export const DAEMONSETS_CONFIG = {
  title: 'DaemonSets',
  apiEndpoint: '/api/daemonsets',
  resourceType: 'daemonsets',
  columns: [
    { title: 'Name', dataIndex: 'name', width: 250 },
    { title: 'Namespace', dataIndex: 'namespace', width: 150 },
    { title: 'Available', dataIndex: 'availableReplicas', width: 100 },
    { title: 'Desired', dataIndex: 'desiredReplicas', width: 100 },
    { title: 'Status', dataIndex: 'status', width: 100 },
  ] as Column<any>[],
  namespaceFilter: true,
};

// Jobs 页面配置
export const JOBS_CONFIG = {
  title: 'Jobs',
  apiEndpoint: '/api/jobs',
  resourceType: 'jobs',
  columns: [
    { title: 'Name', dataIndex: 'name', width: 250 },
    { title: 'Namespace', dataIndex: 'namespace', width: 150 },
    { title: 'Completions', dataIndex: 'completions', width: 100 },
    { title: 'Succeeded', dataIndex: 'succeeded', width: 100 },
    { title: 'Failed', dataIndex: 'failed', width: 100 },
    { title: 'Status', dataIndex: 'status', width: 100 },
  ] as Column<any>[],
  namespaceFilter: true,
};

// CronJobs 页面配置
export const CRONJOBS_CONFIG = {
  title: 'CronJobs',
  apiEndpoint: '/api/cronjobs',
  resourceType: 'cronjobs',
  columns: [
    { title: 'Name', dataIndex: 'name', width: 250 },
    { title: 'Namespace', dataIndex: 'namespace', width: 150 },
    { title: 'Schedule', dataIndex: 'schedule', width: 120 },
    { title: 'Suspend', dataIndex: 'suspend', width: 80 },
    { title: 'Active', dataIndex: 'active', width: 80 },
    { title: 'Status', dataIndex: 'status', width: 100 },
  ] as Column<any>[],
  namespaceFilter: true,
};

// Services 页面配置
export const SERVICES_CONFIG = {
  title: 'Services',
  apiEndpoint: '/api/services',
  resourceType: 'services',
  columns: [
    { title: 'Name', dataIndex: 'name', width: 250 },
    { title: 'Namespace', dataIndex: 'namespace', width: 150 },
    { title: 'Type', dataIndex: 'type', width: 100 },
    { title: 'ClusterIP', dataIndex: 'clusterIP', width: 140 },
    { title: 'Ports', dataIndex: 'ports', width: 150 },
  ] as Column<any>[],
  namespaceFilter: true,
};

// Ingress 页面配置
export const INGRESS_CONFIG = {
  title: 'Ingress',
  apiEndpoint: '/api/ingress',
  resourceType: 'ingress',
  columns: [
    { title: 'Name', dataIndex: 'name', width: 250 },
    { title: 'Namespace', dataIndex: 'namespace', width: 150 },
    { title: 'Class', dataIndex: 'class', width: 100 },
    { title: 'Hosts', dataIndex: 'hosts', width: 200 },
    { title: 'Address', dataIndex: 'address', width: 150 },
    { title: 'Status', dataIndex: 'status', width: 100 },
  ] as Column<any>[],
  namespaceFilter: true,
};

// Events 页面配置
export const EVENTS_CONFIG = {
  title: 'Events',
  apiEndpoint: '/api/events',
  resourceType: 'events',
  columns: [
    { title: 'Type', dataIndex: 'type', width: 200 },
    { title: 'Reason', dataIndex: 'reason', width: 200 },
    { title: 'Message', dataIndex: 'message', width: 200 },
    { title: 'Name', dataIndex: 'name', width: 200 },
    { title: 'Namespace', dataIndex: 'namespace', width: 150 },
    { title: 'LastSeen', dataIndex: 'lastSeen', width: 150 },
    { title: 'Count', dataIndex: 'count', width: 100 },
  ] as Column<any>[],
  namespaceFilter: true,
};

// PVCs 页面配置
export const PVCS_CONFIG = {
  title: 'PersistentVolumeClaims',
  apiEndpoint: '/api/pvcs',
  resourceType: 'pvcs',
  columns: [
    { title: 'Name', dataIndex: 'name', width: 250 },
    { title: 'Namespace', dataIndex: 'namespace', width: 150 },
    { title: 'Status', dataIndex: 'status', width: 100 },
    { title: 'Volume', dataIndex: 'volumeName', width: 150 },
    { title: 'Capacity', dataIndex: 'capacity', width: 100 },
    { title: 'StorageClass', dataIndex: 'storageClass', width: 150 },
  ] as Column<any>[],
  namespaceFilter: true,
};

// PVs 页面配置
export const PVS_CONFIG = {
  title: 'PersistentVolumes',
  apiEndpoint: '/api/pvs',
  resourceType: 'pvs',
  columns: [
    { title: 'Name', dataIndex: 'name', width: 250 },
    { title: 'Status', dataIndex: 'status', width: 100 },
    { title: 'Capacity', dataIndex: 'capacity', width: 100 },
    { title: 'AccessMode', dataIndex: 'accessMode', width: 120 },
    { title: 'StorageClass', dataIndex: 'storageClass', width: 150 },
    { title: 'Claim', dataIndex: 'claimRef', width: 200 },
  ] as Column<any>[],
  namespaceFilter: false,
};

// StorageClasses 页面配置
export const STORAGECLASSES_CONFIG = {
  title: 'StorageClasses',
  apiEndpoint: '/api/storageclasses',
  resourceType: 'storageclasses',
  columns: [
    { title: 'Name', dataIndex: 'name', width: 250 },
    { title: 'Provisioner', dataIndex: 'provisioner', width: 200 },
    { title: 'ReclaimPolicy', dataIndex: 'reclaimPolicy', width: 150 },
    { title: 'BindingMode', dataIndex: 'volumeBindingMode', width: 150 },
    { title: 'Default', dataIndex: 'isDefault', width: 80 },
  ] as Column<any>[],
  namespaceFilter: false,
};

// ConfigMaps 页面配置
export const CONFIGMAPS_CONFIG = {
  title: 'ConfigMaps',
  apiEndpoint: '/api/configmaps',
  resourceType: 'configmaps',
  columns: [
    { title: 'Name', dataIndex: 'name', width: 250 },
    { title: 'Namespace', dataIndex: 'namespace', width: 150 },
    { title: 'DataCount', dataIndex: 'dataCount', width: 100 },
    { title: 'Keys', dataIndex: 'keys', width: 200 },
  ] as Column<any>[],
  namespaceFilter: true,
};

// Secrets 页面配置
export const SECRETS_CONFIG = {
  title: 'Secrets',
  apiEndpoint: '/api/secrets',
  resourceType: 'secrets',
  columns: [
    { title: 'Name', dataIndex: 'name', width: 250 },
    { title: 'Namespace', dataIndex: 'namespace', width: 150 },
    { title: 'Type', dataIndex: 'type', width: 150 },
    { title: 'DataCount', dataIndex: 'dataCount', width: 100 },
  ] as Column<any>[],
  namespaceFilter: true,
};

// Namespaces 页面配置
export const NAMESPACES_CONFIG = {
  title: 'Namespaces',
  apiEndpoint: '/api/namespaces',
  resourceType: 'namespaces',
  columns: [
    { title: 'Name', dataIndex: 'name', width: 300 },
    { title: 'Status', dataIndex: 'status', width: 100 },
  ] as Column<any>[],
  namespaceFilter: false,
};

// Nodes 页面配置
export const NODES_CONFIG = {
  title: 'Nodes',
  apiEndpoint: '/api/nodes',
  resourceType: 'nodes',
  columns: [
    { title: 'Name', dataIndex: 'name', width: 250 },
    { title: 'IP', dataIndex: 'ip', width: 150 },
    { title: 'Role', dataIndex: 'role', width: 150 },
    { title: 'CPU', dataIndex: 'cpuUsage', width: 120 },
    { title: 'Memory', dataIndex: 'memoryUsage', width: 120 },
    { title: 'Pods', dataIndex: 'pods', width: 100 },
    { title: 'Status', dataIndex: 'status', width: 100 },
  ] as Column<any>[],
  namespaceFilter: false,
};
