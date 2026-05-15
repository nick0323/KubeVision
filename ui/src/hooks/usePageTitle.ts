import { useEffect } from 'react';

const BASE_TITLE = 'KubeVision';

const TAB_TITLES: Record<string, string> = {
  overview: 'Overview',
  clusters: 'Cluster Management',
  pods: 'Pods',
  deployments: 'Deployments',
  statefulsets: 'StatefulSets',
  daemonsets: 'DaemonSets',
  jobs: 'Jobs',
  cronjobs: 'CronJobs',
  services: 'Services',
  ingress: 'Ingress',
  configmaps: 'ConfigMaps',
  secrets: 'Secrets',
  pvcs: 'PersistentVolumeClaims',
  pvs: 'PersistentVolumes',
  storageclasses: 'StorageClasses',
  nodes: 'Nodes',
  namespaces: 'Namespaces',
  events: 'Events',
  networkpolicies: 'NetworkPolicies',
  serviceaccounts: 'ServiceAccounts',
  roles: 'Roles',
  rolebindings: 'RoleBindings',
  clusterroles: 'ClusterRoles',
  clusterrolebindings: 'ClusterRoleBindings',
  resourcequotas: 'ResourceQuotas',
  limitranges: 'LimitRanges',
  poddisruptionbudgets: 'PodDisruptionBudgets',
  horizontalpodautoscalers: 'HPA',
  argocd: 'ArgoCD',
  crds: 'CRDs',
};

export function usePageTitle(title?: string) {
  useEffect(() => {
    document.title = title ? `${title} | ${BASE_TITLE}` : BASE_TITLE;
    return () => {
      document.title = BASE_TITLE;
    };
  }, [title]);
}

export function getTabTitle(tab: string): string {
  return TAB_TITLES[tab] || tab.charAt(0).toUpperCase() + tab.slice(1);
}
