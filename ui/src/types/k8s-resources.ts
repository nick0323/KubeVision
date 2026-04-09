/**
 * Kubernetes 资源类型定义
 * 提供完整的类型安全支持
 */

import React from 'react';

// ==================== 基础类型 ====================

export interface K8sMetadata {
  name: string;
  namespace?: string;
  uid?: string;
  resourceVersion?: string;
  generation?: number;
  creationTimestamp?: string;
  deletionTimestamp?: string;
  labels?: Record<string, string>;
  annotations?: Record<string, string>;
  ownerReferences?: K8sOwnerReference[];
  finalizers?: string[];
}

export interface K8sOwnerReference {
  apiVersion: string;
  kind: string;
  name: string;
  uid: string;
  controller?: boolean;
  blockOwnerDeletion?: boolean;
}

export interface K8sLocalObjectReference {
  name: string;
}

export interface K8sResource {
  apiVersion: string;
  kind: string;
  metadata: K8sMetadata;
}

export interface K8sResourceList<T extends K8sResource> {
  apiVersion: string;
  kind: string;
  metadata?: {
    resourceVersion?: string;
    continue?: string;
    remainingItemCount?: number;
  };
  items: T[];
}

// ==================== API 响应类型 ====================

export interface PageMeta {
  total: number;
  limit: number;
  offset: number;
}

export interface APIResponse<T = any> {
  code: number;
  message: string;
  data: T;
  traceId?: string;
  timestamp?: number;
  page?: PageMeta;
}

export interface APIErrorResponse {
  code: number;
  message: string;
  details?: { field?: string; reason?: string }[];
  traceId?: string;
  timestamp?: number;
}

export interface ListQueryParams {
  namespace?: string;
  search?: string;
  limit: number;
  offset: number;
  sortBy?: string;
  sortOrder?: 'asc' | 'desc';
  labelSelector?: string;
  fieldSelector?: string;
}

export interface PaginatedResponse<T> {
  data: T[];
  page?: PageMeta;
}

// ==================== 表格列定义 ====================

export interface ColumnDef<T> {
  title: string;
  dataIndex: keyof T | string;
  width?: number | string;
  sortable?: boolean;
  render?: (value: T[keyof T], record: T, index: number) => React.ReactNode;
  className?: string;
  hidden?: boolean;
}

export interface StatusColumnDef<T> extends ColumnDef<T> {
  dataIndex: 'status' | 'phase' | 'state';
  statusMap?: Record<
    string,
    { color: 'success' | 'error' | 'warning' | 'default' | 'processing'; text?: string }
  >;
}

// 兼容旧代码的 Column 别名
export type Column<T = any> = ColumnDef<T>;

// ==================== 容器相关类型 ====================

export interface ContainerPort {
  name?: string;
  containerPort: number;
  protocol?: 'TCP' | 'UDP' | 'SCTP';
  hostIP?: string;
  hostPort?: number;
}

export interface ResourceRequirements {
  limits?: Record<string, string>;
  requests?: Record<string, string>;
}

export interface EnvVar {
  name: string;
  value?: string;
  valueFrom?: {
    configMapKeyRef?: { name: string; key: string; optional?: boolean };
    secretKeyRef?: { name: string; key: string; optional?: boolean };
    fieldRef?: { apiVersion?: string; fieldPath: string };
    resourceFieldRef?: { containerName?: string; resource: string; divisor?: string };
  };
}

export interface EnvFromSource {
  prefix?: string;
  configMapRef?: K8sLocalObjectReference;
  secretRef?: K8sLocalObjectReference;
}

export interface VolumeMount {
  name: string;
  mountPath: string;
  readOnly?: boolean;
  mountPropagation?: 'None' | 'HostToContainer' | 'Bidirectional';
  subPath?: string;
  subPathExpr?: string;
}

export interface Probe {
  httpGet?: {
    path?: string;
    port: number | string;
    host?: string;
    scheme?: 'HTTP' | 'HTTPS';
    httpHeaders?: { name: string; value: string }[];
  };
  tcpSocket?: { port: number | string; host?: string };
  exec?: { command?: string[] };
  initialDelaySeconds?: number;
  timeoutSeconds?: number;
  periodSeconds?: number;
  successThreshold?: number;
  failureThreshold?: number;
  terminationGracePeriodSeconds?: number;
}

export interface SecurityContext {
  privileged?: boolean;
  runAsNonRoot?: boolean;
  runAsUser?: number;
  runAsGroup?: number;
  readOnlyRootFilesystem?: boolean;
  allowPrivilegeEscalation?: boolean;
  capabilities?: { add?: string[]; drop?: string[] };
  seccompProfile?: {
    type: 'RuntimeDefault' | 'Localhost' | 'Unconfined';
    localhostProfile?: string;
  };
}

export interface Container {
  name: string;
  image: string;
  command?: string[];
  args?: string[];
  workingDir?: string;
  ports?: ContainerPort[];
  env?: EnvVar[];
  envFrom?: EnvFromSource[];
  volumeMounts?: VolumeMount[];
  resources?: ResourceRequirements;
  readinessProbe?: Probe;
  livenessProbe?: Probe;
  startupProbe?: Probe;
  lifecycle?: {
    postStart?: { exec?: { command?: string[] }; httpGet?: any; tcpSocket?: any };
    preStop?: { exec?: { command?: string[] }; httpGet?: any; tcpSocket?: any };
  };
  terminationMessagePath?: string;
  terminationMessagePolicy?: 'File' | 'FallbackToLogsOnError';
  imagePullPolicy?: 'Always' | 'Never' | 'IfNotPresent';
  securityContext?: SecurityContext;
  stdin?: boolean;
  stdinOnce?: boolean;
  tty?: boolean;
}

// ==================== 卷相关类型 ====================

export interface Volume {
  name: string;
  hostPath?: { path: string; type?: string };
  emptyDir?: { sizeLimit?: string; medium?: 'Memory' | '' };
  configMap?: { name: string; items?: { key: string; path: string }[]; optional?: boolean };
  secret?: { secretName: string; items?: { key: string; path: string }[]; optional?: boolean };
  persistentVolumeClaim?: { claimName: string; readOnly?: boolean };
  downwardAPI?: any;
  projected?: any;
  nfs?: { server: string; path: string; readOnly?: boolean };
  awsElasticBlockStore?: any;
  azureDisk?: any;
  azureFile?: any;
  cephfs?: any;
  cinder?: any;
  csi?: any;
  ephemeral?: any;
  fc?: any;
  flexVolume?: any;
  flocker?: any;
  gcePersistentDisk?: any;
  gitRepo?: any;
  glusterfs?: any;
  iscsi?: any;
  local?: any;
  photonPersistentDisk?: any;
  portworxVolume?: any;
  quobyte?: any;
  rbd?: any;
  scaleIO?: any;
  storageos?: any;
  vsphereVolume?: any;
}

// ==================== Pod 相关类型 ====================

export type PodPhase = 'Pending' | 'Running' | 'Succeeded' | 'Failed' | 'Unknown';

export interface PodCondition {
  type: 'Initialized' | 'Ready' | 'ContainersReady' | 'PodScheduled' | string;
  status: 'True' | 'False' | 'Unknown';
  lastProbeTime?: string;
  lastTransitionTime?: string;
  reason?: string;
  message?: string;
}

export interface ContainerStatusState {
  waiting?: { reason?: string; message?: string };
  running?: { startedAt?: string };
  terminated?: {
    exitCode?: number;
    reason?: string;
    message?: string;
    startedAt?: string;
    finishedAt?: string;
    containerID?: string;
  };
}

export interface ContainerStatus {
  name: string;
  image: string;
  imageID: string;
  ready: boolean;
  restartCount: number;
  state?: ContainerStatusState;
  lastState?: ContainerStatusState;
  started?: boolean;
  containerID?: string;
}

export interface PodSpec {
  nodeName?: string;
  nodeSelector?: Record<string, string>;
  serviceAccountName?: string;
  serviceAccount?: string;
  containers: Container[];
  initContainers?: Container[];
  restartPolicy?: 'Always' | 'OnFailure' | 'Never';
  terminationGracePeriodSeconds?: number;
  activeDeadlineSeconds?: number;
  dnsPolicy?: 'ClusterFirst' | 'Default' | 'ClusterFirstWithHostNet' | 'None';
  automountServiceAccountToken?: boolean;
  hostNetwork?: boolean;
  hostPID?: boolean;
  hostIPC?: boolean;
  securityContext?: {
    runAsNonRoot?: boolean;
    runAsUser?: number;
    runAsGroup?: number;
    fsGroup?: number;
    supplementalGroups?: number[];
    seccompProfile?: any;
    sysctls?: any[];
  };
  imagePullSecrets?: K8sLocalObjectReference[];
  hostname?: string;
  subdomain?: string;
  affinity?: any;
  tolerations?: any[];
  priorityClassName?: string;
  priority?: number;
  preemptionPolicy?: 'PreemptLowerPriority' | 'Never';
  volumes?: Volume[];
}

export interface PodStatus {
  phase: PodPhase;
  conditions: PodCondition[];
  message?: string;
  reason?: string;
  hostIP?: string;
  hostIPs?: { ip: string }[];
  podIP?: string;
  podIPs?: { ip: string }[];
  startTime?: string;
  initContainerStatuses?: ContainerStatus[];
  containerStatuses: ContainerStatus[];
  qosClass?: 'Guaranteed' | 'Burstable' | 'BestEffort';
  nominatedNodeName?: string;
}

export interface Pod extends K8sResource {
  metadata: K8sMetadata & { namespace: string };
  spec: PodSpec;
  status: PodStatus;
}

export interface PodListItem {
  name: string;
  namespace: string;
  status: PodPhase;
  ready: string;
  restarts: number;
  ip: string;
  node: string;
  age: string;
  containers: number;
  _origin?: Pod;
}

// ==================== Deployment 相关类型 ====================

export interface Deployment extends K8sResource {
  metadata: K8sMetadata & { namespace: string };
  spec: {
    replicas?: number;
    selector: { matchLabels?: Record<string, string>; matchExpressions?: any[] };
    template: { metadata: K8sMetadata; spec: PodSpec };
    strategy?: { type?: 'Recreate' | 'RollingUpdate'; rollingUpdate?: any };
    minReadySeconds?: number;
    revisionHistoryLimit?: number;
    paused?: boolean;
    progressDeadlineSeconds?: number;
  };
  status: {
    replicas: number;
    updatedReplicas: number;
    readyReplicas: number;
    availableReplicas: number;
    unavailableReplicas?: number;
    observedGeneration?: number;
    conditions?: { type: string; status: 'True' | 'False' | 'Unknown'; reason?: string }[];
    collisionCount?: number;
  };
}

export interface DeploymentListItem {
  name: string;
  namespace: string;
  status: string;
  readyReplicas: string;
  updatedReplicas: number;
  availableReplicas: number;
  age: string;
  _origin?: Deployment;
}

// ==================== StatefulSet 相关类型 ====================

export interface StatefulSet extends K8sResource {
  metadata: K8sMetadata & { namespace: string };
  spec: {
    replicas?: number;
    serviceName: string;
    selector: { matchLabels?: Record<string, string> };
    template: { metadata: K8sMetadata; spec: PodSpec };
    podManagementPolicy?: 'OrderedReady' | 'Parallel';
    updateStrategy?: any;
    volumeClaimTemplates?: any[];
    minReadySeconds?: number;
    revisionHistoryLimit?: number;
  };
  status: {
    replicas: number;
    readyReplicas: number;
    currentReplicas: number;
    updatedReplicas: number;
    availableReplicas: number;
    collisionCount?: number;
    observedGeneration?: number;
  };
}

export interface StatefulSetListItem {
  name: string;
  namespace: string;
  status: string;
  readyReplicas: string;
  age: string;
  _origin?: StatefulSet;
}

// ==================== DaemonSet 相关类型 ====================

export interface DaemonSet extends K8sResource {
  metadata: K8sMetadata & { namespace: string };
  spec: {
    selector: { matchLabels?: Record<string, string> };
    template: { metadata: K8sMetadata; spec: PodSpec };
    updateStrategy?: any;
    minReadySeconds?: number;
    revisionHistoryLimit?: number;
  };
  status: {
    currentNumberScheduled: number;
    numberMisscheduled: number;
    desiredNumberScheduled: number;
    numberReady: number;
    updatedNumberScheduled?: number;
    numberAvailable?: number;
    numberUnavailable?: number;
    collisionCount?: number;
    observedGeneration?: number;
  };
}

export interface DaemonSetListItem {
  name: string;
  namespace: string;
  status: string;
  readyReplicas: string;
  age: string;
  _origin?: DaemonSet;
}

// ==================== Service 相关类型 ====================

export type ServiceType = 'ClusterIP' | 'NodePort' | 'LoadBalancer' | 'ExternalName';

export interface ServicePort {
  name?: string;
  protocol?: 'TCP' | 'UDP' | 'SCTP';
  port: number;
  targetPort?: number | string;
  nodePort?: number;
  appProtocol?: string;
}

export interface Service extends K8sResource {
  metadata: K8sMetadata & { namespace: string };
  spec: {
    type?: ServiceType;
    selector?: Record<string, string>;
    ports: ServicePort[];
    clusterIP?: string;
    clusterIPs?: string[];
    externalIPs?: string[];
    loadBalancerIP?: string;
    externalName?: string;
    externalTrafficPolicy?: 'Cluster' | 'Local';
    sessionAffinity?: 'None' | 'ClientIP';
  };
  status: {
    loadBalancer?: { ingress?: { ip?: string; hostname?: string }[] };
  };
}

export interface ServiceListItem {
  name: string;
  namespace: string;
  type: ServiceType;
  clusterIP: string;
  externalIP: string;
  ports: string;
  age: string;
  _origin?: Service;
}

// ==================== Node 相关类型 ====================

export interface Node extends K8sResource {
  spec: {
    providerID?: string;
    unschedulable?: boolean;
    taints?: any[];
  };
  status: {
    conditions: {
      type: string;
      status: 'True' | 'False' | 'Unknown';
      lastHeartbeatTime?: string;
    }[];
    addresses: { type: string; address: string }[];
    nodeInfo: {
      machineID: string;
      systemUUID: string;
      bootID: string;
      kernelVersion: string;
      osImage: string;
      containerRuntimeVersion: string;
      kubeletVersion: string;
      kubeProxyVersion: string;
      operatingSystem: string;
      architecture: string;
    };
    capacity?: { cpu?: string; memory?: string; pods?: string; [key: string]: string | undefined };
    allocatable?: {
      cpu?: string;
      memory?: string;
      pods?: string;
      [key: string]: string | undefined;
    };
  };
}

export interface NodeListItem {
  name: string;
  status: string;
  roles: string;
  age: string;
  version: string;
  internalIP: string;
  osImage: string;
  kernelVersion: string;
  containerRuntime: string;
  cpu: string;
  memory: string;
  pods: string;
  _origin?: Node;
}

// ==================== K8s 事件类型 ====================

export interface K8sEvent extends K8sResource {
  involvedObject: ObjectReference;
  reason: string;
  message: string;
  source: EventSource;
  firstTimestamp: string;
  lastTimestamp: string;
  count: number;
  type: 'Normal' | 'Warning';
  eventTime?: string;
  series?: any;
  action?: string;
  related?: ObjectReference;
  reportingController?: string;
  reportingInstance?: string;
}

export interface ObjectReference {
  apiVersion?: string;
  kind?: string;
  namespace?: string;
  name: string;
  uid?: string;
  resourceVersion?: string;
  fieldPath?: string;
}

export interface EventSource {
  component?: string;
  host?: string;
}

export interface EventSeries {
  count: number;
  lastObservedTime: string;
}

// ==================== 通用类型 ====================

export interface GenericResourceItem {
  name: string;
  namespace?: string;
  status?: string;
  age: string;
  [key: string]: any;
}

export interface ResourceListItem {
  name: string;
  namespace?: string;
  status?: string;
  age?: string;
  [key: string]: any;
}

export interface ActionButton {
  label: string;
  icon?: React.ReactNode;
  onClick: (record: any) => void;
  confirm?: boolean;
  confirmMessage?: string;
  disabled?: boolean;
  danger?: boolean;
}

// ==================== 类型映射 ====================

export type ResourceListItemMap = {
  pods: PodListItem;
  deployments: DeploymentListItem;
  statefulsets: StatefulSetListItem;
  daemonsets: DaemonSetListItem;
  services: ServiceListItem;
  nodes: NodeListItem;
  generic: GenericResourceItem;
};

export type ResourceMap = {
  pods: Pod;
  deployments: Deployment;
  statefulsets: StatefulSet;
  daemonsets: DaemonSet;
  services: Service;
  nodes: Node;
  generic: K8sResource;
};

export type ResourceType = keyof ResourceMap;
export type GetListItem<T extends ResourceType> = ResourceListItemMap[T];
export type GetResource<T extends ResourceType> = ResourceMap[T];

// ==================== 辅助类型 ====================

export interface LabelSelector {
  matchLabels?: Record<string, string>;
  matchExpressions?: { key: string; operator: string; values?: string[] }[];
}

export interface Toleration {
  key?: string;
  operator?: 'Equal' | 'Exists';
  value?: string;
  effect?: 'NoSchedule' | 'PreferNoSchedule' | 'NoExecute';
  tolerationSeconds?: number;
}

export interface Affinity {
  nodeAffinity?: any;
  podAffinity?: any;
  podAntiAffinity?: any;
}
