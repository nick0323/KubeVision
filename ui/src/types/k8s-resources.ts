/**
 * Kubernetes 资源类型定义
 * 提供完整的类型安全支持，用于 ResourcePage 泛型组件
 */

import React from 'react';

// ==================== 基础接口 ====================

/**
 * K8s 资源元数据
 */
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

/**
 * 所有者引用
 */
export interface K8sOwnerReference {
  apiVersion: string;
  kind: string;
  name: string;
  uid: string;
  controller?: boolean;
  blockOwnerDeletion?: boolean;
}

/**
 * K8s 资源基础接口
 * 所有 K8s 资源都必须实现此接口
 */
export interface K8sResource {
  apiVersion: string;
  kind: string;
  metadata: K8sMetadata;
}

/**
 * K8s 资源列表
 */
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

/**
 * 分页元数据
 */
export interface PageMeta {
  total: number;
  limit: number;
  offset: number;
}

/**
 * 统一 API 响应结构
 */
export interface APIResponse<T = any> {
  code: number;           // 0 表示成功，非 0 表示失败
  message: string;        // 响应消息
  data: T;                // 数据（数组或对象）
  traceId?: string;       // 追踪 ID
  timestamp?: number;     // 时间戳
  page?: PageMeta;        // 分页信息（列表查询时）
}

/**
 * API 错误响应
 */
export interface APIErrorResponse {
  code: number;
  message: string;
  details?: {
    field?: string;
    reason?: string;
  }[];
  traceId?: string;
  timestamp?: number;
}

/**
 * 分页查询参数
 */
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

/**
 * 分页响应数据
 */
export interface PaginatedResponse<T> {
  data: T[];
  page?: PageMeta;
}

// ==================== 表格列定义 ====================

/**
 * 表格列定义（泛型版本）
 * @template T - 记录类型
 */
export interface ColumnDef<T> {
  /** 列标题 */
  title: string;
  /** 数据字段索引 */
  dataIndex: keyof T | string;
  /** 列宽度 */
  width?: number | string;
  /** 是否可排序 */
  sortable?: boolean;
  /** 自定义渲染函数 */
  render?: (value: T[keyof T], record: T, index: number) => React.ReactNode;
  /** CSS 类名 */
  className?: string;
  /** 是否隐藏 */
  hidden?: boolean;
}

/**
 * 状态列配置
 */
export interface StatusColumnDef<T> extends ColumnDef<T> {
  dataIndex: 'status' | 'phase' | 'state' | string;
  statusMap?: Record<string, {
    color: 'success' | 'error' | 'warning' | 'default' | 'processing';
    text?: string;
  }>;
}

// ==================== Pod 相关类型 ====================

/**
 * Pod 状态阶段
 */
export type PodPhase = 'Pending' | 'Running' | 'Succeeded' | 'Failed' | 'Unknown';

/**
 * Pod 容器状态
 */
export interface ContainerStatus {
  name: string;
  image: string;
  imageID: string;
  ready: boolean;
  restartCount: number;
  state?: {
    waiting?: { reason?: string; message?: string };
    running?: { startedAt?: string };
    terminated?: {
      exitCode?: number;
      reason?: string;
      message?: string;
      startedAt?: string;
      finishedAt?: string;
    };
  };
  lastState?: {
    waiting?: { reason?: string };
    running?: { startedAt?: string };
    terminated?: { reason?: string; exitCode?: number };
  };
  started?: boolean;
  containerID?: string;
}

/**
 * Pod 条件
 */
export interface PodCondition {
  type: 'Initialized' | 'Ready' | 'ContainersReady' | 'PodScheduled' | string;
  status: 'True' | 'False' | 'Unknown';
  lastProbeTime?: string;
  lastTransitionTime?: string;
  reason?: string;
  message?: string;
}

/**
 * Pod 资源
 */
export interface Pod extends K8sResource {
  metadata: K8sMetadata & {
    namespace: string;
  };
  spec: {
    nodeName?: string;
    nodeSelector?: Record<string, string>;
    serviceAccountName?: string;
    containers: {
      name: string;
      image: string;
      ports?: { containerPort: number; protocol?: string }[];
      resources?: {
        requests?: { cpu?: string; memory?: string };
        limits?: { cpu?: string; memory?: string };
      };
    }[];
    initContainers?: { name: string; image: string }[];
    restartPolicy?: 'Always' | 'OnFailure' | 'Never';
    terminationGracePeriodSeconds?: number;
  };
  status: {
    phase: PodPhase;
    conditions: PodCondition[];
    hostIP?: string;
    podIP?: string;
    podIPs?: { ip: string }[];
    startTime?: string;
    containerStatuses: ContainerStatus[];
    initContainerStatuses?: ContainerStatus[];
    qosClass?: 'Guaranteed' | 'Burstable' | 'BestEffort';
    nominatedNodeName?: string;
  };
}

/**
 * Pod 列表项（UI 展示用）
 */
export interface PodListItem {
  name: string;
  namespace: string;
  status: PodPhase;
  ready: string;           // e.g., "1/3"
  restarts: number;
  ip: string;
  node: string;
  age: string;
  containers: number;
  // 原始对象
  _origin?: Pod;
}

// ==================== Deployment 相关类型 ====================

/**
 * Deployment 策略
 */
export interface DeploymentStrategy {
  type?: 'Recreate' | 'RollingUpdate';
  rollingUpdate?: {
    maxUnavailable?: number | string;
    maxSurge?: number | string;
  };
}

/**
 * Deployment 条件
 */
export interface DeploymentCondition {
  type: 'Available' | 'Progressing' | 'ReplicaFailure' | string;
  status: 'True' | 'False' | 'Unknown';
  lastUpdateTime?: string;
  lastTransitionTime?: string;
  reason?: string;
  message?: string;
}

/**
 * Deployment 资源
 */
export interface Deployment extends K8sResource {
  metadata: K8sMetadata & {
    namespace: string;
  };
  spec: {
    replicas?: number;
    selector: {
      matchLabels?: Record<string, string>;
      matchExpressions?: {
        key: string;
        operator: string;
        values?: string[];
      }[];
    };
    template: {
      metadata: K8sMetadata;
      spec: Pod['spec'];
    };
    strategy?: DeploymentStrategy;
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
    conditions?: DeploymentCondition[];
    collisionCount?: number;
  };
}

/**
 * Deployment 列表项（UI 展示用）
 */
export interface DeploymentListItem {
  name: string;
  namespace: string;
  status: string;          // Healthy, Partial, Unavailable
  readyReplicas: string;   // e.g., "3/5"
  updatedReplicas: number;
  availableReplicas: number;
  age: string;
  // 原始对象
  _origin?: Deployment;
}

// ==================== StatefulSet 相关类型 ====================

/**
 * StatefulSet 资源
 */
export interface StatefulSet extends K8sResource {
  metadata: K8sMetadata & {
    namespace: string;
  };
  spec: {
    replicas?: number;
    serviceName: string;
    selector: {
      matchLabels?: Record<string, string>;
    };
    template: {
      metadata: K8sMetadata;
      spec: Pod['spec'];
    };
    podManagementPolicy?: 'OrderedReady' | 'Parallel';
    updateStrategy?: {
      type?: 'RollingUpdate' | 'OnDelete';
      rollingUpdate?: { partition?: number };
    };
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

/**
 * StatefulSet 列表项
 */
export interface StatefulSetListItem {
  name: string;
  namespace: string;
  status: string;
  readyReplicas: string;
  age: string;
  _origin?: StatefulSet;
}

// ==================== DaemonSet 相关类型 ====================

/**
 * DaemonSet 资源
 */
export interface DaemonSet extends K8sResource {
  metadata: K8sMetadata & {
    namespace: string;
  };
  spec: {
    selector: {
      matchLabels?: Record<string, string>;
    };
    template: {
      metadata: K8sMetadata;
      spec: Pod['spec'];
    };
    updateStrategy?: {
      type?: 'RollingUpdate' | 'OnDelete';
    };
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

/**
 * DaemonSet 列表项
 */
export interface DaemonSetListItem {
  name: string;
  namespace: string;
  status: string;
  readyReplicas: string;
  age: string;
  _origin?: DaemonSet;
}

// ==================== Service 相关类型 ====================

/**
 * Service 类型
 */
export type ServiceType = 'ClusterIP' | 'NodePort' | 'LoadBalancer' | 'ExternalName';

/**
 * Service 端口
 */
export interface ServicePort {
  name?: string;
  protocol?: 'TCP' | 'UDP' | 'SCTP';
  port: number;
  targetPort?: number | string;
  nodePort?: number;
  appProtocol?: string;
}

/**
 * Service 资源
 */
export interface Service extends K8sResource {
  metadata: K8sMetadata & {
    namespace: string;
  };
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
    loadBalancer?: {
      ingress?: { ip?: string; hostname?: string }[];
    };
  };
}

/**
 * Service 列表项
 */
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

/**
 * Node 条件
 */
export interface NodeCondition {
  type: 'Ready' | 'MemoryPressure' | 'DiskPressure' | 'PIDPressure' | 'NetworkUnavailable';
  status: 'True' | 'False' | 'Unknown';
  lastHeartbeatTime?: string;
  lastTransitionTime?: string;
  reason?: string;
  message?: string;
}

/**
 * Node 资源
 */
export interface Node extends K8sResource {
  spec: {
    providerID?: string;
    unschedulable?: boolean;
    taints?: {
      key: string;
      value?: string;
      effect: 'NoSchedule' | 'PreferNoSchedule' | 'NoExecute';
    }[];
  };
  status: {
    conditions: NodeCondition[];
    addresses: {
      type: 'Hostname' | 'ExternalIP' | 'InternalIP' | 'ExternalDNS' | 'InternalDNS';
      address: string;
    }[];
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
    capacity?: {
      cpu?: string;
      memory?: string;
      pods?: string;
      [key: string]: string | undefined;
    };
    allocatable?: {
      cpu?: string;
      memory?: string;
      pods?: string;
      [key: string]: string | undefined;
    };
  };
}

/**
 * Node 列表项
 */
export interface NodeListItem {
  name: string;
  status: string;          // Ready, NotReady, Unknown
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

// ==================== 通用列表项类型 ====================

/**
 * 通用资源列表项
 * 用于不确定具体类型时的兜底方案
 */
export interface GenericResourceItem {
  name: string;
  namespace?: string;
  status?: string;
  age: string;
  [key: string]: any;
}

// ==================== 资源类型映射 ====================

/**
 * 资源类型到列表项的映射
 */
export type ResourceListItemMap = {
  pods: PodListItem;
  deployments: DeploymentListItem;
  statefulsets: StatefulSetListItem;
  daemonsets: DaemonSetListItem;
  services: ServiceListItem;
  nodes: NodeListItem;
  generic: GenericResourceItem;
};

/**
 * 资源类型到原始资源的映射
 */
export type ResourceMap = {
  pods: Pod;
  deployments: Deployment;
  statefulsets: StatefulSet;
  daemonsets: DaemonSet;
  services: Service;
  nodes: Node;
  generic: K8sResource;
};

/**
 * 资源类型联合
 */
export type ResourceType = keyof ResourceMap;

/**
 * 根据资源类型获取列表项类型
 */
export type GetListItem<T extends ResourceType> = ResourceListItemMap[T];

/**
 * 根据资源类型获取原始资源类型
 */
export type GetResource<T extends ResourceType> = ResourceMap[T];
