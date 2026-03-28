/**
 * ResourcePage 泛型组件使用示例
 * 
 * 展示如何使用类型安全的 ResourcePage 组件
 */

import React from 'react';
import { ResourcePage } from '../components/ResourcePage';
import type { 
  Pod, 
  Deployment, 
  Service,
  ColumnDef,
  PodListItem,
  DeploymentListItem,
} from '../types';
import { StatusBadge } from './StatusBadge';

// ==================== 示例 1: Pod 列表页面 ====================

/**
 * Pod 列定义 - 类型安全
 * dataIndex 会自动提示 Pod 的字段
 * render 函数的 value 和 record 参数类型自动推断
 */
const POD_COLUMNS: ColumnDef<Pod>[] = [
  {
    title: 'Name',
    dataIndex: 'metadata.name',
    width: '25%',
    sortable: true,
  },
  {
    title: 'Namespace',
    dataIndex: 'metadata.namespace',
    width: '15%',
  },
  {
    title: 'Status',
    dataIndex: 'status.phase',
    width: '12%',
    // value 类型自动推断为 PodPhase
    render: (value, record) => <StatusBadge status={value} resourceType="pod" />,
  },
  {
    title: 'Ready',
    dataIndex: 'status.containerStatuses',
    width: '10%',
    // record 类型为 Pod
    render: (_, record) => {
      const ready = record.status.containerStatuses.filter(c => c.ready).length;
      const total = record.status.containerStatuses.length;
      return <span>{ready}/{total}</span>;
    },
  },
  {
    title: 'Restarts',
    dataIndex: 'status.containerStatuses',
    width: '10%',
    render: (_, record) => {
      const restarts = record.status.containerStatuses.reduce(
        (sum, c) => sum + c.restartCount,
        0
      );
      return <span>{restarts}</span>;
    },
  },
  {
    title: 'IP',
    dataIndex: 'status.podIP',
    width: '12%',
  },
  {
    title: 'Node',
    dataIndex: 'spec.nodeName',
    width: '12%',
  },
  {
    title: 'Age',
    dataIndex: 'metadata.creationTimestamp',
    width: '8%',
    render: (value) => {
      const date = new Date(value);
      const now = new Date();
      const diff = now.getTime() - date.getTime();
      const minutes = Math.floor(diff / 60000);
      const hours = Math.floor(minutes / 60);
      const days = Math.floor(hours / 24);
      
      if (days > 0) return `${days}d`;
      if (hours > 0) return `${hours}h`;
      return `${minutes}m`;
    },
  },
];

/**
 * Pod 列表页面组件
 */
export const PodListPage: React.FC = () => {
  return (
    <ResourcePage<Pod>
      title="Pods"
      apiEndpoint="/api/pods"
      resourceType="pod"
      columns={POD_COLUMNS}
      collapsed={false}
      onToggleCollapsed={() => {}}
      namespaceFilter={true}
      onRowClick={(record) => {
        // record 类型为 Pod
        console.log('Clicked pod:', record.metadata.name);
        console.log('Namespace:', record.metadata.namespace);
      }}
    />
  );
};

// ==================== 示例 2: Deployment 列表页面 ====================

const DEPLOYMENT_COLUMNS: ColumnDef<Deployment>[] = [
  {
    title: 'Name',
    dataIndex: 'metadata.name',
    width: '25%',
    sortable: true,
  },
  {
    title: 'Namespace',
    dataIndex: 'metadata.namespace',
    width: '15%',
  },
  {
    title: 'Status',
    dataIndex: 'status.conditions',
    width: '12%',
    render: (_, record) => {
      const available = record.status.conditions?.some(
        c => c.type === 'Available' && c.status === 'True'
      );
      const progressing = record.status.conditions?.some(
        c => c.type === 'Progressing' && c.status === 'True'
      );
      
      let status = 'Unknown';
      if (available) status = 'Healthy';
      else if (progressing) status = 'Progressing';
      else status = 'Unavailable';
      
      return <StatusBadge status={status} resourceType="deployment" />;
    },
  },
  {
    title: 'Ready',
    dataIndex: 'status.readyReplicas',
    width: '12%',
    render: (_, record) => {
      const ready = record.status.readyReplicas || 0;
      const total = record.spec.replicas || 0;
      return <span>{ready}/{total}</span>;
    },
  },
  {
    title: 'Up-to-date',
    dataIndex: 'status.updatedReplicas',
    width: '12%',
  },
  {
    title: 'Available',
    dataIndex: 'status.availableReplicas',
    width: '12%',
  },
  {
    title: 'Age',
    dataIndex: 'metadata.creationTimestamp',
    width: '12%',
  },
];

export const DeploymentListPage: React.FC = () => {
  return (
    <ResourcePage<Deployment>
      title="Deployments"
      apiEndpoint="/api/deployments"
      resourceType="deployment"
      columns={DEPLOYMENT_COLUMNS}
      collapsed={false}
      onToggleCollapsed={() => {}}
      namespaceFilter={true}
    />
  );
};

// ==================== 示例 3: Service 列表页面 ====================

const SERVICE_COLUMNS: ColumnDef<Service>[] = [
  {
    title: 'Name',
    dataIndex: 'metadata.name',
    width: '25%',
    sortable: true,
  },
  {
    title: 'Namespace',
    dataIndex: 'metadata.namespace',
    width: '15%',
  },
  {
    title: 'Type',
    dataIndex: 'spec.type',
    width: '12%',
  },
  {
    title: 'Cluster IP',
    dataIndex: 'spec.clusterIP',
    width: '15%',
  },
  {
    title: 'Ports',
    dataIndex: 'spec.ports',
    width: '15%',
    render: (value) => {
      if (!value || value.length === 0) return '-';
      return value.map(p => `${p.port}/${p.protocol || 'TCP'}`).join(', ');
    },
  },
  {
    title: 'Age',
    dataIndex: 'metadata.creationTimestamp',
    width: '12%',
  },
];

export const ServiceListPage: React.FC = () => {
  return (
    <ResourcePage<Service>
      title="Services"
      apiEndpoint="/api/services"
      resourceType="service"
      columns={SERVICE_COLUMNS}
      collapsed={false}
      onToggleCollapsed={() => {}}
      namespaceFilter={true}
    />
  );
};

// ==================== 示例 4: 使用自定义渲染器 ====================

interface CustomPodRenderers {
  'metadata.name': (value: string, record: Pod) => React.ReactNode;
  'actions': (value: null, record: Pod) => React.ReactNode;
}

export const PodListPageWithCustomRenderers: React.FC = () => {
  const customRenderers: CustomPodRenderers = {
    'metadata.name': (value, record) => (
      <a href={`/pods/${record.metadata.namespace}/${value}`}>
        {value}
      </a>
    ),
    'actions': (_, record) => (
      <div style={{ display: 'flex', gap: '8px' }}>
        <button onClick={() => console.log('Edit', record)}>编辑</button>
        <button onClick={() => console.log('Delete', record)}>删除</button>
      </div>
    ),
  };

  return (
    <ResourcePage<Pod>
      title="Pods"
      apiEndpoint="/api/pods"
      resourceType="pod"
      columns={[
        ...POD_COLUMNS,
        {
          title: 'Actions',
          dataIndex: 'actions',
          width: '120px',
        },
      ]}
      collapsed={false}
      onToggleCollapsed={() => {}}
      customRenderers={customRenderers}
    />
  );
};

// ==================== 示例 5: 类型推断演示 ====================

/**
 * 类型推断示例
 * 
 * 1. columns 的 dataIndex 会提示 Pod 的字段路径
 * 2. render 函数的 value 类型根据 dataIndex 自动推断
 * 3. record 类型始终为 Pod
 */
const TYPED_COLUMNS: ColumnDef<Pod>[] = [
  {
    title: 'Name',
    dataIndex: 'metadata.name',  // ✅ 自动提示
    render: (value, record) => {
      // value 类型：string（根据 metadata.name 推断）
      // record 类型：Pod
      return <span>{value}</span>;
    },
  },
  {
    title: 'Phase',
    dataIndex: 'status.phase',  // ✅ 自动提示
    render: (value, record) => {
      // value 类型：PodPhase（'Pending' | 'Running' | ...）
      // record 类型：Pod
      return <span>{value}</span>;
    },
  },
  {
    title: 'Container Count',
    dataIndex: 'spec.containers',  // ✅ 自动提示
    render: (value, record) => {
      // value 类型：Container[]
      // record 类型：Pod
      return <span>{value?.length || 0}</span>;
    },
  },
];
