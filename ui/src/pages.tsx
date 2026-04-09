import React from 'react';

// 懒加载概览页
const LazyOverview = React.lazy(() => import('./pages/OverviewPage.tsx'));

// 导入 ResourceListPage（不懒加载，因为需要传递 props）
import { ResourceListPage } from './pages/ResourceListPage.tsx';
import {
  PODS_CONFIG,
  DEPLOYMENTS_CONFIG,
  STATEFULSETS_CONFIG,
  DAEMONSETS_CONFIG,
  JOBS_CONFIG,
  CRONJOBS_CONFIG,
  SERVICES_CONFIG,
  INGRESS_CONFIG,
  PVCS_CONFIG,
  PVS_CONFIG,
  STORAGECLASSES_CONFIG,
  CONFIGMAPS_CONFIG,
  SECRETS_CONFIG,
  NAMESPACES_CONFIG,
  NODES_CONFIG,
  EVENTS_CONFIG,
} from './constants/pageConfigs';

interface PageComponentProps {
  collapsed: boolean;
  onToggleCollapsed: () => void;
}

/**
 * 创建资源列表页面组件
 */
const createResourcePage = (config: any) => {
  return ({ collapsed, onToggleCollapsed }: PageComponentProps) => (
    <ResourceListPage config={config} collapsed={collapsed} onToggleCollapsed={onToggleCollapsed} />
  );
};

/**
 * 所有资源页面组件映射
 */
export const PAGE_COMPONENTS = {
  // 概览页
  overview: LazyOverview,

  // Pods
  pods: createResourcePage(PODS_CONFIG),

  // Deployments
  deployments: createResourcePage(DEPLOYMENTS_CONFIG),

  // StatefulSets
  statefulsets: createResourcePage(STATEFULSETS_CONFIG),

  // DaemonSets
  daemonsets: createResourcePage(DAEMONSETS_CONFIG),

  // Jobs
  jobs: createResourcePage(JOBS_CONFIG),

  // CronJobs
  cronjobs: createResourcePage(CRONJOBS_CONFIG),

  // Services
  services: createResourcePage(SERVICES_CONFIG),

  // Ingress
  ingress: createResourcePage(INGRESS_CONFIG),

  // PVCs
  pvcs: createResourcePage(PVCS_CONFIG),

  // PVs
  pvs: createResourcePage(PVS_CONFIG),

  // StorageClasses
  storageclasses: createResourcePage(STORAGECLASSES_CONFIG),

  // ConfigMaps
  configmaps: createResourcePage(CONFIGMAPS_CONFIG),

  // Secrets
  secrets: createResourcePage(SECRETS_CONFIG),

  // Namespaces
  namespaces: createResourcePage(NAMESPACES_CONFIG),

  // Nodes
  nodes: createResourcePage(NODES_CONFIG),

  // Events
  events: createResourcePage(EVENTS_CONFIG),
};

export default PAGE_COMPONENTS;
