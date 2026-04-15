/**
 * Resource Configuration Factory
 * 用于生成资源列表页面配置，减少重复代码
 */

import type { ColumnDef } from '../types/k8s-resources';

export interface ExtendedColumn extends ColumnDef<any> {
  sortable?: boolean;
}

interface ResourcePageConfig {
  title: string;
  apiEndpoint: string;
  resourceType: string;
  columns: ExtendedColumn[];
  namespaceFilter: boolean;
  defaultSort?: {
    field: string;
    order: 'asc' | 'desc';
  };
}

/**
 * 创建通用的 Name 列
 */
export function createNameColumn(width: string = '25%'): ExtendedColumn {
  return {
    title: 'Name',
    dataIndex: 'name',
    width,
    sortable: true,
  };
}

/**
 * 创建 Namespace 列
 */
export function createNamespaceColumn(width: string = '15%'): ExtendedColumn {
  return {
    title: 'Namespace',
    dataIndex: 'namespace',
    width,
    sortable: false,
  };
}

/**
 * 创建 Status 列
 */
export function createStatusColumn(width: string = '15%'): ExtendedColumn {
  return {
    title: 'Status',
    dataIndex: 'status',
    width,
    sortable: true,
  };
}

/**
 * 创建 Age 列
 */
export function createAgeColumn(width: string = '8%'): ExtendedColumn {
  return {
    title: 'Age',
    dataIndex: 'age',
    width,
    sortable: true,
  };
}

/**
 * 创建 Ready 列（用于 Workload 资源）
 */
export function createReadyColumn(options?: {
  width?: string;
  desiredReplicasField?: string;
}): ExtendedColumn {
  const { width = '10%', desiredReplicasField = 'desiredReplicas' } = options || {};
  
  return {
    title: 'Ready',
    dataIndex: 'readyReplicas',
    width,
    render: (v: any, r: any) => {
      const desired = r[desiredReplicasField] || 0;
      return `${v}/${desired}`;
    },
  };
}

/**
 * 创建基础配置（包含通用列）
 */
export function createBaseConfig(options: {
  title: string;
  apiEndpoint: string;
  resourceType: string;
  hasNamespace?: boolean;
}): Partial<ResourcePageConfig> {
  const { title, apiEndpoint, resourceType, hasNamespace = true } = options;
  
  const columns: ExtendedColumn[] = [createNameColumn()];
  
  if (hasNamespace) {
    columns.push(createNamespaceColumn());
  }
  
  columns.push(createStatusColumn());
  
  return {
    title,
    apiEndpoint,
    resourceType,
    columns,
    namespaceFilter: hasNamespace,
  };
}

/**
 * 完成配置（添加默认值和排序）
 */
export function finalizeConfig(
  base: Partial<ResourcePageConfig>,
  options?: {
    defaultSort?: { field: string; order: 'asc' | 'desc' };
    additionalColumns?: ExtendedColumn[];
  }
): ResourcePageConfig {
  const { defaultSort = { field: 'name', order: 'asc' }, additionalColumns = [] } = options || {};
  
  if (!base.columns) {
    base.columns = [];
  }
  
  // 添加 Age 列作为最后一列
  if (!base.columns.some(col => col.dataIndex === 'age')) {
    base.columns.push(createAgeColumn());
  }
  
  return {
    title: base.title || 'Resource',
    apiEndpoint: base.apiEndpoint || '/api/resource',
    resourceType: base.resourceType || 'unknown',
    columns: [...base.columns, ...additionalColumns],
    namespaceFilter: base.namespaceFilter ?? true,
    defaultSort,
  };
}

/**
 * 创建工作负载类资源配置（Deployment, StatefulSet, DaemonSet 等）
 */
export function createWorkloadConfig(options: {
  title: string;
  apiEndpoint: string;
  resourceType: string;
  extraColumns?: ExtendedColumn[];
}): ResourcePageConfig {
  const base = createBaseConfig(options);
  
  const columns: ExtendedColumn[] = [
    createNameColumn('30%'),
    createNamespaceColumn('15%'),
    createStatusColumn('15%'),
    createReadyColumn(),
    ...(options.extraColumns || []),
  ];
  
  return finalizeConfig({ ...base, columns });
}

/**
 * 创建 Pod 资源配置
 */
export function createPodConfig(): ResourcePageConfig {
  const columns: ExtendedColumn[] = [
    createNameColumn('25%'),
    createNamespaceColumn('15%'),
    createStatusColumn('15%'),
    { title: 'Ready', dataIndex: 'ready', width: '8%' },
    { title: 'Restarts', dataIndex: 'restarts', width: '8%', sortable: true },
    { title: 'IP', dataIndex: 'podIP', width: '11%' },
    { title: 'Node', dataIndex: 'nodeName', width: '10%' },
    createAgeColumn('8%'),
  ];
  
  return finalizeConfig({
    title: 'Pods',
    apiEndpoint: '/api/pod',
    resourceType: 'pod',
    columns,
    namespaceFilter: true,
    defaultSort: { field: 'name', order: 'asc' },
  });
}

/**
 * 创建 Service 资源配置
 */
export function createServiceConfig(): ResourcePageConfig {
  const columns: ExtendedColumn[] = [
    createNameColumn('25%'),
    createNamespaceColumn('18%'),
    { title: 'Type', dataIndex: 'type', width: '15%' },
    { title: 'Cluster IP', dataIndex: 'clusterIP', width: '18%' },
    { title: 'Ports', dataIndex: 'ports', width: '16%' },
    createAgeColumn('8%'),
  ];
  
  return finalizeConfig({
    title: 'Services',
    apiEndpoint: '/api/service',
    resourceType: 'service',
    columns,
    namespaceFilter: true,
  });
}

/**
 * 创建 Node 资源配置（集群资源）
 */
export function createNodeConfig(): ResourcePageConfig {
  const columns: ExtendedColumn[] = [
    createNameColumn('15%'),
    { title: 'IP', dataIndex: 'ip', width: '15%' },
    { title: 'Role', dataIndex: 'role', width: '20%', sortable: true },
    {
      title: 'CPU',
      dataIndex: 'cpuUsage',
      width: '10%',
      render: (value: any) => value !== null && value !== undefined ? `${Math.round(value)}%` : 'N/A',
    },
    {
      title: 'Memory',
      dataIndex: 'memoryUsage',
      width: '10%',
      render: (value: any) => value !== null && value !== undefined ? `${Math.round(value)}%` : 'N/A',
    },
    {
      title: 'Pods',
      dataIndex: 'podsUsed',
      width: '10%',
      render: (value: any, record: any) => `${value}/${record.podsCapacity || 0}`,
    },
    createStatusColumn('10%'),
    createAgeColumn('10%'),
  ];
  
  return finalizeConfig({
    title: 'Nodes',
    apiEndpoint: '/api/node',
    resourceType: 'node',
    columns,
    namespaceFilter: false,
    defaultSort: { field: 'name', order: 'asc' },
  });
}

/**
 * 创建集群级资源配置（PV, StorageClass, Namespace 等）
 */
export function createClusterResourceConfig(options: {
  title: string;
  apiEndpoint: string;
  resourceType: string;
  columns: ExtendedColumn[];
}): ResourcePageConfig {
  return finalizeConfig({
    ...options,
    namespaceFilter: false,
  });
}
