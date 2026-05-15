/**
 * Resource Configuration Factory
 * for生成resourceListPage config，reduceduplicatecode
 */

import type { ColumnDef } from '../types/k8s-resources';

export interface ExtendedColumn<T = Record<string, unknown>> extends ColumnDef<T> {
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
 * CreateCommon's Name 列
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
 * Create Namespace 列
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
 * Create Status 列
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
 * Create Age 列
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
 * Create Ready 列（for Workload resource）
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
    render: (v: unknown, r: Record<string, unknown>) => {
      const desired = r[desiredReplicasField] || 0;
      return `${v}/${desired}`;
    },
  };
}

/**
 * Create基础Config（includeCommon列）
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
 * 完成Config（Adddefault值andsort）
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
  
  // Add Age 列作for最after一列
  if (!base.columns.some(col => col.dataIndex === 'age')) {
    base.columns.push(createAgeColumn());
  }
  
  return {
    title: base.title || 'Resource',
    apiEndpoint: base.apiEndpoint || '/api/v1/resource',
    resourceType: base.resourceType || 'unknown',
    columns: [...base.columns, ...additionalColumns],
    namespaceFilter: base.namespaceFilter ?? true,
    defaultSort,
  };
}

/**
 * Create工作负载类resourceConfig（Deployment, StatefulSet, DaemonSet etc）
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
 * Create Pod resourceConfig
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
    apiEndpoint: '/api/v1/pod',
    resourceType: 'pod',
    columns,
    defaultSort: { field: 'name', order: 'asc' },
  });
}

export function createServiceConfig(): ResourcePageConfig {
  const columns: ExtendedColumn[] = [
    createNameColumn('20%'),
    createNamespaceColumn('12%'),
    { title: 'Type', dataIndex: 'type', width: '10%', sortable: true },
    { title: 'ClusterIP', dataIndex: 'clusterIP', width: '18%' },
    { title: 'Ports', dataIndex: 'ports', width: '26%', sortable: false },
    createAgeColumn('8%'),
  ];
  
  return finalizeConfig({
    title: 'Services',
    apiEndpoint: '/api/v1/service',
    resourceType: 'service',
    columns,
    defaultSort: { field: 'name', order: 'asc' },
  });
}

export function createNodeConfig(): ResourcePageConfig {
  const columns: ExtendedColumn[] = [
    createNameColumn('15%'),
    { title: 'IP', dataIndex: 'ip', width: '15%' },
    { title: 'Role', dataIndex: 'role', width: '20%', sortable: true },
    {
      title: 'CPU',
      dataIndex: 'cpuUsage',
      width: '10%',
      render: (value: unknown) => value !== null && value !== undefined ? `${Math.round(value as number)}%` : 'N/A',
    },
    {
      title: 'Memory',
      dataIndex: 'memoryUsage',
      width: '10%',
      render: (value: unknown) => value !== null && value !== undefined ? `${Math.round(value as number)}%` : 'N/A',
    },
    {
      title: 'Pods',
      dataIndex: 'podsUsed',
      width: '10%',
      render: (value: unknown, record: Record<string, unknown>) => `${value || 0}/${record.podsCapacity || 0}`,
    },
    createStatusColumn('10%'),
    createAgeColumn('10%'),
  ];
  
  return finalizeConfig({
    title: 'Nodes',
    apiEndpoint: '/api/v1/node',
    resourceType: 'node',
    columns,
    namespaceFilter: false,
    defaultSort: { field: 'name', order: 'asc' },
  });
}

/**
 * Createcluster级resourceConfig（PV, StorageClass, Namespace etc）
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
