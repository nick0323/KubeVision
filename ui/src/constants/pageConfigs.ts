import { Column } from '../types';

// 扩展 Column 类型以支持 sortable
interface ExtendedColumn<T = any> extends Column<T> {
  sortable?: boolean;
}

// ==================== Workloads ====================

// Pods 页面配置 - 保持所有排序 ✅
export const PODS_CONFIG = {
  title: 'Pods',
  apiEndpoint: '/api/pods',
  resourceType: 'pods',
  columns: [
    { title: 'Name', dataIndex: 'name', width: '25%', sortable: true },
    { title: 'Namespace', dataIndex: 'namespace', width: '15%', sortable: false },
    { title: 'Status', dataIndex: 'status', width: '15%', sortable: true },
    { title: 'Ready', dataIndex: 'ready', width: '8%', sortable: false },
    { title: 'Restarts', dataIndex: 'restarts', width: '8%', sortable: true },
    { title: 'IP', dataIndex: 'podIP', width: '11%', sortable: false },
    { title: 'Node', dataIndex: 'nodeName', width: '10%', sortable: false },
    { title: 'Age', dataIndex: 'age', width: '8%', sortable: true },
  ] as ExtendedColumn<any>[],
  namespaceFilter: true,
  defaultSort: { field: 'age', order: 'desc' as const },
};

// Deployments 页面配置 - 简化排序（7→3）
export const DEPLOYMENTS_CONFIG = {
  title: 'Deployments',
  apiEndpoint: '/api/deployments',
  resourceType: 'deployments',
  columns: [
    { title: 'Name', dataIndex: 'name', width: '25%', sortable: true },
    { title: 'Namespace', dataIndex: 'namespace', width: '15%', sortable: false },
    { title: 'Status', dataIndex: 'status', width: '15%', sortable: true },
    {
      title: 'Ready',
      dataIndex: 'readyReplicas',
      width: '10%',
      sortable: false,
      render: (value: any, record: any) => `${value}/${record.desiredReplicas || 0}`
    },
    { title: 'Up To Date', dataIndex: 'updatedReplicas', width: '10%', sortable: false },
    { title: 'Available', dataIndex: 'availableReplicas', width: '10%', sortable: false },
    { title: 'Restarts', dataIndex: 'restarts', width: '5%', sortable: true },
    { title: 'Age', dataIndex: 'age', width: '10%', sortable: false },
  ] as ExtendedColumn<any>[],
  namespaceFilter: true,
};

// StatefulSets 页面配置 - 简化排序（7→3）
export const STATEFULSETS_CONFIG = {
  title: 'StatefulSets',
  apiEndpoint: '/api/statefulsets',
  resourceType: 'statefulsets',
  columns: [
    { title: 'Name', dataIndex: 'name', width: '25%', sortable: true },
    { title: 'Namespace', dataIndex: 'namespace', width: '15%', sortable: false },
    { title: 'Status', dataIndex: 'status', width: '12%', sortable: true },
    {
      title: 'Ready',
      dataIndex: 'readyReplicas',
      width: '12%',
      sortable: false,
      render: (value: any, record: any) => `${value}/${record.desiredReplicas || 0}`
    },
    { title: 'Restarts', dataIndex: 'restarts', width: '12%', sortable: true },
    { title: 'Age', dataIndex: 'age', width: '12%', sortable: false },
  ] as ExtendedColumn<any>[],
  namespaceFilter: true,
};

// DaemonSets 页面配置 - 简化排序（7→3）
export const DAEMONSETS_CONFIG = {
  title: 'DaemonSets',
  apiEndpoint: '/api/daemonsets',
  resourceType: 'daemonsets',
  columns: [
    { title: 'Name', dataIndex: 'name', width: '25%', sortable: true },
    { title: 'Namespace', dataIndex: 'namespace', width: '15%', sortable: false },
    { title: 'Status', dataIndex: 'status', width: '12%', sortable: true },
    { title: 'Desired', dataIndex: 'desiredReplicas', width: '10%', sortable: false },
    { title: 'Ready', dataIndex: 'readyReplicas', width: '10%', sortable: false },
    { title: 'Restarts', dataIndex: 'restarts', width: '10%', sortable: true },
    { title: 'Age', dataIndex: 'age', width: '8%', sortable: false },
  ] as ExtendedColumn<any>[],
  namespaceFilter: true,
};

// Jobs 页面配置 - 简化排序（7→5）
export const JOBS_CONFIG = {
  title: 'Jobs',
  apiEndpoint: '/api/jobs',
  resourceType: 'jobs',
  columns: [
    { title: 'Name', dataIndex: 'name', width: '25%', sortable: true },
    { title: 'Namespace', dataIndex: 'namespace', width: '15%', sortable: false },
    { title: 'Status', dataIndex: 'status', width: '12%', sortable: true },
    {
      title: 'Completions',
      dataIndex: 'succeeded',
      width: '12%',
      sortable: false,
      render: (value: any, record: any) => {
        const completions = record.completions || 1;
        return `${value || 0}/${completions}`;
      }
    },
    { title: 'Restarts', dataIndex: 'restarts', width: '12%', sortable: true },
    { title: 'Duration', dataIndex: 'duration', width: '12%', sortable: false },
    { title: 'Age', dataIndex: 'age', width: '12%', sortable: false },
  ] as ExtendedColumn<any>[],
  namespaceFilter: true,
};

// CronJobs 页面配置 - 简化排序（7→4）
export const CRONJOBS_CONFIG = {
  title: 'CronJobs',
  apiEndpoint: '/api/cronjobs',
  resourceType: 'cronjobs',
  columns: [
    { title: 'Name', dataIndex: 'name', width: '25%', sortable: true },
    { title: 'Namespace', dataIndex: 'namespace', width: '15%', sortable: false},
    { title: 'Status', dataIndex: 'status', width: '12%', sortable: true },
    { title: 'Suspend', dataIndex: 'suspend', width: '10%', sortable: false },
    { title: 'Schedule', dataIndex: 'schedule', width: '15%', sortable: false },
    { title: 'Active', dataIndex: 'active', width: '8%', sortable: false },
    { title: 'Restarts', dataIndex: 'restarts', width: '8%', sortable: true },
    { title: 'Age', dataIndex: 'age', width: '7%', sortable: false },
  ] as ExtendedColumn<any>[],
  namespaceFilter: true,
};

// ==================== Network ====================

// Services 页面配置 - 简化排序（4→3）
export const SERVICES_CONFIG = {
  title: 'Services',
  apiEndpoint: '/api/services',
  resourceType: 'services',
  columns: [
    { title: 'Name', dataIndex: 'name', width: '25%', sortable: true },
    { title: 'Namespace', dataIndex: 'namespace', width: '18%', sortable: false },
    { title: 'Type', dataIndex: 'type', width: '15%', sortable: false },
    { title: 'Cluster IP', dataIndex: 'clusterIP', width: '18%', sortable: false },
    { title: 'Ports', dataIndex: 'ports', width: '16%', sortable: false },
    { title: 'Age', dataIndex: 'age', width: '8%', sortable: false },
  ] as ExtendedColumn<any>[],
  namespaceFilter: true,
};

// Ingress 页面配置 - 简化排序（4→2）
export const INGRESS_CONFIG = {
  title: 'Ingress',
  apiEndpoint: '/api/ingress',
  resourceType: 'ingress',
  columns: [
    { title: 'Name', dataIndex: 'name', width: '20%', sortable: true },
    { title: 'Namespace', dataIndex: 'namespace', width: '10%', sortable: false },
    { title: 'Class', dataIndex: 'class', width: '10%', sortable: false },
    { title: 'Hosts', dataIndex: 'hosts', width: '15%', sortable: false },
    { 
      title: 'Target Service', 
      dataIndex: 'targetService', 
      width: '15%', 
      sortable: false,
      render: (value: any) => {
        // 去重展示
        if (!value) return '-';
        if (Array.isArray(value)) {
          const unique = [...new Set(value)];
          return unique.join(', ');
        }
        return value;
      }
    },
    { title: 'Path', dataIndex: 'path', width: '25%', sortable: false },
    { title: 'Age', dataIndex: 'age', width: '5%', sortable: false },
  ] as ExtendedColumn<any>[],
  namespaceFilter: true,
};

// ==================== Storage ====================

// PVCs 页面配置 - 简化排序（7→3）
export const PVCS_CONFIG = {
  title: 'PersistentVolumeClaims',
  apiEndpoint: '/api/pvcs',
  resourceType: 'pvcs',
  columns: [
    { title: 'Name', dataIndex: 'name', width: '20%', sortable: true },
    { title: 'Namespace', dataIndex: 'namespace', width: '15%', sortable: false },
    { title: 'Status', dataIndex: 'status', width: '10%', sortable: true },
    { title: 'Access Mode', dataIndex: 'accessMode', width: '15%', sortable: false },
    { title: 'Volume', dataIndex: 'volumeName', width: '10%', sortable: false },
    { title: 'Capacity', dataIndex: 'capacity', width: '10%', sortable: false },
    { title: 'Storage Class', dataIndex: 'storageClass', width: '12%', sortable: false },
    { title: 'Age', dataIndex: 'age', width: '8%', sortable: false },
  ] as ExtendedColumn<any>[],
  namespaceFilter: true,
};

// PVs 页面配置 - 简化排序（6→2）
export const PVS_CONFIG = {
  title: 'PersistentVolumes',
  apiEndpoint: '/api/pvs',
  resourceType: 'pvs',
  columns: [
    { title: 'Name', dataIndex: 'name', width: '25%', sortable: true },
    { title: 'Status', dataIndex: 'status', width: '10%', sortable: true },
    { title: 'Capacity', dataIndex: 'capacity', width: '10%', sortable: false },
    { title: 'Access Mode', dataIndex: 'accessMode', width: '15%', sortable: false },
    { title: 'Storage Class', dataIndex: 'storageClass', width: '10%', sortable: false },
    { title: 'Claim', dataIndex: 'claimRef', width: '15%', sortable: false },
    { title: 'Reclaim Policy', dataIndex: 'reclaimPolicy', width: '10%', sortable: false },
    { title: 'Age', dataIndex: 'age', width: '5%', sortable: false },
  ] as ExtendedColumn<any>[],
  namespaceFilter: false,
};

// StorageClasses 页面配置 - 简化排序（4→2）
export const STORAGECLASSES_CONFIG = {
  title: 'StorageClasses',
  apiEndpoint: '/api/storageclasses',
  resourceType: 'storageclasses',
  columns: [
    { title: 'Name', dataIndex: 'name', width: '22%', sortable: true },
    { title: 'Provisioner', dataIndex: 'provisioner', width: '25%', sortable: false },
    { title: 'Reclaim Policy', dataIndex: 'reclaimPolicy', width: '15%', sortable: false },
    { title: 'Binding Mode', dataIndex: 'volumeBindingMode', width: '15%', sortable: false },
    { title: 'Default', dataIndex: 'isDefault', width: '8%', sortable: false },
    { title: 'Age', dataIndex: 'age', width: '15%', sortable: false },
  ] as ExtendedColumn<any>[],
  namespaceFilter: false,
};

// ==================== Configuration ====================

// ConfigMaps 页面配置 - 简化排序（4→2）
export const CONFIGMAPS_CONFIG = {
  title: 'ConfigMaps',
  apiEndpoint: '/api/configmaps',
  resourceType: 'configmaps',
  columns: [
    { title: 'Name', dataIndex: 'name', width: '30%', sortable: true },
    { title: 'Namespace', dataIndex: 'namespace', width: '20%', sortable: false },
    { title: 'Data Count', dataIndex: 'dataCount', width: '10%', sortable: false },
    { title: 'Keys', dataIndex: 'keys', width: '30%', sortable: false },
    { title: 'Age', dataIndex: 'age', width: '10%', sortable: false },
  ] as ExtendedColumn<any>[],
  namespaceFilter: true,
};

// Secrets 页面配置 - 简化排序（5→3）
export const SECRETS_CONFIG = {
  title: 'Secrets',
  apiEndpoint: '/api/secrets',
  resourceType: 'secrets',
  columns: [
    { title: 'Name', dataIndex: 'name', width: '25%', sortable: true },
    { title: 'Namespace', dataIndex: 'namespace', width: '20%', sortable: false },
    { title: 'Type', dataIndex: 'type', width: '20%', sortable: false },
    { title: 'Data Count', dataIndex: 'dataCount', width: '10%', sortable: false },
    { title: 'Age', dataIndex: 'age', width: '15%', sortable: false },
  ] as ExtendedColumn<any>[],
  namespaceFilter: true,
};

// ==================== Cluster ====================

// Namespaces 页面配置 - 简化排序（3→2）
export const NAMESPACES_CONFIG = {
  title: 'Namespaces',
  apiEndpoint: '/api/namespaces',
  resourceType: 'namespaces',
  columns: [
    { title: 'Name', dataIndex: 'name', width: '60%', sortable: true },
    { title: 'Status', dataIndex: 'status', width: '25%', sortable: true },
    { title: 'Age', dataIndex: 'age', width: '15%', sortable: false },
  ] as ExtendedColumn<any>[],
  namespaceFilter: false,
};

// Nodes 页面配置 - 简化排序（7→3）
export const NODES_CONFIG = {
  title: 'Nodes',
  apiEndpoint: '/api/nodes',
  resourceType: 'nodes',
  columns: [
    { title: 'Name', dataIndex: 'name', width: '15%', sortable: true },
    { title: 'IP', dataIndex: 'ip', width: '15%', sortable: false },
    { title: 'Role', dataIndex: 'role', width: '20%', sortable: true },
    { title: 'CPU', dataIndex: 'cpuUsage', width: '10%', sortable: false },
    { title: 'Memory', dataIndex: 'memoryUsage', width: '10%', sortable: false },
    { title: 'Pods', dataIndex: 'podsUsed', width: '10%', sortable: false },
    { title: 'Status', dataIndex: 'status', width: '10%', sortable: true },
    { title: 'Age', dataIndex: 'age', width: '10%', sortable: false },
  ] as ExtendedColumn<any>[],
  namespaceFilter: false,
};

// Events 页面配置 - 简化排序（6→3）
export const EVENTS_CONFIG = {
  title: 'Events',
  apiEndpoint: '/api/events',
  resourceType: 'events',
  columns: [
    { title: 'Type', dataIndex: 'type', width: '8%', sortable: true },
    { title: 'Reason', dataIndex: 'reason', width: '12%', sortable: false },
    { title: 'Message', dataIndex: 'message', width: '25%', sortable: false },
    { title: 'Object', dataIndex: 'object', width: '20%', sortable: false },
    { title: 'Last Seen', dataIndex: 'lastSeen', width: '15%', sortable: true },
    { title: 'Count', dataIndex: 'count', width: '8%', sortable: false },
  ] as ExtendedColumn<any>[],
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
