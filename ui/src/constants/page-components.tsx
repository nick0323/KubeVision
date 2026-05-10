import React from 'react';

// LazyLoading...
const LazyOverview = React.lazy(() => import('../pages/OverviewPage.tsx'));
const LazyArgoCD = React.lazy(() => import('../pages/ArgoCDPage.tsx'));

// 导入 ResourceListPage（notLazyLoading...need传递 props）
import { ResourceListPage } from '../pages/ResourceListPage.tsx';
import {
  PODS_CONFIG,
  DEPLOYMENTS_CONFIG,
  STATEFULSETS_CONFIG,
  DAEMONSETS_CONFIG,
  JOBS_CONFIG,
  CRONJOBS_CONFIG,
  SERVICES_CONFIG,
  INGRESS_CONFIG,
  NETWORKPOLICIES_CONFIG,
  SERVICEACCOUNTS_CONFIG,
  ROLES_CONFIG,
  ROLEBINDINGS_CONFIG,
  CLUSTERROLES_CONFIG,
  CLUSTERROLEBINDINGS_CONFIG,
  RESOURCEQUOTAS_CONFIG,
  LIMITRANGES_CONFIG,
  PODDISRUPTIONBUDGETS_CONFIG,
  PVCS_CONFIG,
  PVS_CONFIG,
  STORAGECLASSES_CONFIG,
  CONFIGMAPS_CONFIG,
  SECRETS_CONFIG,
  NAMESPACES_CONFIG,
  NODES_CONFIG,
  EVENTS_CONFIG,
  HPAS_CONFIG,
} from './pageConfigs';

interface PageComponentProps {
  collapsed: boolean;
  onToggleCollapsed: () => void;
}

/**
 * CreateResource list page面Component
 */
const createResourcePage = (config: any) => {
  return ({ collapsed, onToggleCollapsed }: PageComponentProps) => (
    <ResourceListPage config={config} collapsed={collapsed} onToggleCollapsed={onToggleCollapsed} />
  );
};

/**
 * 所hasresourcepageComponentMapping
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

  // NetworkPolicies
  networkpolicies: createResourcePage(NETWORKPOLICIES_CONFIG),

  // ServiceAccounts
  serviceaccounts: createResourcePage(SERVICEACCOUNTS_CONFIG),

  // Roles
  roles: createResourcePage(ROLES_CONFIG),

  // RoleBindings
  rolebindings: createResourcePage(ROLEBINDINGS_CONFIG),

  // ClusterRoles
  clusterroles: createResourcePage(CLUSTERROLES_CONFIG),

  // ClusterRoleBindings
  clusterrolebindings: createResourcePage(CLUSTERROLEBINDINGS_CONFIG),

  // ResourceQuotas
  resourcequotas: createResourcePage(RESOURCEQUOTAS_CONFIG),

  // LimitRanges
  limitranges: createResourcePage(LIMITRANGES_CONFIG),

  // PodDisruptionBudgets
  poddisruptionbudgets: createResourcePage(PODDISRUPTIONBUDGETS_CONFIG),

  // HPA
  horizontalpodautoscalers: createResourcePage(HPAS_CONFIG),

  // ArgoCD (GitOps)
  argocd: LazyArgoCD,
};

export default PAGE_COMPONENTS;
