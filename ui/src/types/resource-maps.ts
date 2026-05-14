import type {
  PodListItem,
  DeploymentListItem,
  StatefulSetListItem,
  DaemonSetListItem,
  CronJobListItem,
  JobListItem,
  Pod,
  Deployment,
  StatefulSet,
  DaemonSet,
  CronJob,
  K8sJob,
} from './k8s-workloads';
import type {
  ServiceListItem,
  IngressListItem,
  NetworkPolicyListItem,
  Service,
  Ingress,
  NetworkPolicy,
} from './k8s-network';
import type {
  RoleListItem,
  RoleBindingListItem,
  ClusterRoleListItem,
  ClusterRoleBindingListItem,
  Role,
  RoleBinding,
  ClusterRole,
  ClusterRoleBinding,
  ServiceAccountListItem,
  ServiceAccount,
} from './k8s-rbac';
import type {
  ConfigMapListItem,
  SecretListItem,
  PVCListItem,
  PVListItem,
  StorageClassListItem,
  ConfigMap,
  Secret,
  PersistentVolumeClaim,
  PersistentVolume,
  StorageClass,
} from './k8s-storage-config';
import type {
  NodeListItem,
  NamespaceListItem,
  EventListItem,
  HPAListItem,
  ResourceQuotaListItem,
  LimitRangeListItem,
  PodDisruptionBudgetListItem,
  Node,
  K8sNamespace,
  K8sEvent,
  HorizontalPodAutoscaler,
  ResourceQuota,
  LimitRange,
  PodDisruptionBudget,
} from './k8s-other';
import type { GenericResourceItem, K8sResource } from './core';

export type ResourceListItemMap = {
  pods: PodListItem;
  deployments: DeploymentListItem;
  statefulsets: StatefulSetListItem;
  daemonsets: DaemonSetListItem;
  services: ServiceListItem;
  nodes: NodeListItem;
  horizontalpodautoscalers: HPAListItem;
  networkpolicies: NetworkPolicyListItem;
  serviceaccounts: ServiceAccountListItem;
  roles: RoleListItem;
  rolebindings: RoleBindingListItem;
  clusterroles: ClusterRoleListItem;
  clusterrolebindings: ClusterRoleBindingListItem;
  resourcequotas: ResourceQuotaListItem;
  limitranges: LimitRangeListItem;
  poddisruptionbudgets: PodDisruptionBudgetListItem;
  configmaps: ConfigMapListItem;
  secrets: SecretListItem;
  ingress: IngressListItem;
  pvcs: PVCListItem;
  pvs: PVListItem;
  storageclasses: StorageClassListItem;
  namespaces: NamespaceListItem;
  events: EventListItem;
  jobs: JobListItem;
  cronjobs: CronJobListItem;
  generic: GenericResourceItem;
};

export type ResourceMap = {
  pods: Pod;
  deployments: Deployment;
  statefulsets: StatefulSet;
  daemonsets: DaemonSet;
  services: Service;
  nodes: Node;
  horizontalpodautoscalers: HorizontalPodAutoscaler;
  networkpolicies: NetworkPolicy;
  serviceaccounts: ServiceAccount;
  roles: Role;
  rolebindings: RoleBinding;
  clusterroles: ClusterRole;
  clusterrolebindings: ClusterRoleBinding;
  resourcequotas: ResourceQuota;
  limitranges: LimitRange;
  poddisruptionbudgets: PodDisruptionBudget;
  configmaps: ConfigMap;
  secrets: Secret;
  ingress: Ingress;
  pvcs: PersistentVolumeClaim;
  pvs: PersistentVolume;
  storageclasses: StorageClass;
  namespaces: K8sNamespace;
  events: K8sEvent;
  jobs: K8sJob;
  cronjobs: CronJob;
  generic: K8sResource;
};

export type ResourceType = keyof ResourceMap;
export type GetListItem<T extends ResourceType> = ResourceListItemMap[T];
export type GetResource<T extends ResourceType> = ResourceMap[T];
