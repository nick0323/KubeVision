import type { K8sResource, K8sMetadata, MetricSpec, MetricStatus, LabelSelector } from './core';

export interface Node extends K8sResource {
  spec: {
    providerID?: string;
    unschedulable?: boolean;
    taints?: { key: string; value?: string; effect: string }[];
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

export interface K8sNamespace extends K8sResource {
  status?: { phase: string };
}

export interface NamespaceListItem {
  name: string;
  status: string;
  age: string;
}

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
  series?: EventSeries;
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

export interface EventListItem {
  name: string;
  namespace: string;
  type: string;
  reason: string;
  message: string;
  count: number;
  lastSeen: string;
}

export interface HPAListItem {
  name: string;
  namespace: string;
  minReplicas: number;
  maxReplicas: number;
  currentReplicas: number;
  desiredReplicas: number;
  metrics: string;
  age: string;
  _origin?: HorizontalPodAutoscaler;
}

export interface HorizontalPodAutoscaler extends K8sResource {
  metadata: K8sMetadata & { namespace: string };
  spec: {
    scaleTargetRef: {
      apiVersion: string;
      kind: string;
      name: string;
    };
    minReplicas?: number;
    maxReplicas: number;
    metrics: MetricSpec[];
  };
  status: {
    currentReplicas: number;
    desiredReplicas: number;
    currentMetrics?: MetricStatus[];
    lastScaleTime?: string;
    conditions?: {
      type: string;
      status: string;
      lastTransitionTime: string;
      reason: string;
      message: string;
    }[];
  };
}

export interface ResourceQuotaListItem {
  name: string;
  namespace: string;
  requests: string;
  limits: string;
  age: string;
  _origin?: ResourceQuota;
}

export interface ResourceQuota extends K8sResource {
  metadata: K8sMetadata & { namespace: string };
  spec: {
    hard?: Record<string, string>;
    scopeSelector?: { matchExpressions: { scopeName: string; operator: string; values?: string[] }[] };
    scopes?: string[];
  };
  status: {
    hard?: Record<string, string>;
    used?: Record<string, string>;
  };
}

export interface LimitRangeItem {
  type: string;
  max?: Record<string, string>;
  min?: Record<string, string>;
  default?: Record<string, string>;
  defaultRequest?: Record<string, string>;
  maxLimitRequestRatio?: Record<string, string>;
}

export interface LimitRangeListItem {
  name: string;
  namespace: string;
  limits: string;
  age: string;
  _origin?: LimitRange;
}

export interface LimitRange extends K8sResource {
  metadata: K8sMetadata & { namespace: string };
  spec: {
    limits: LimitRangeItem[];
  };
}

export interface PodDisruptionBudgetListItem {
  name: string;
  namespace: string;
  minAvailable: string;
  maxUnavailable: string;
  currentHealthy: number;
  desiredHealthy: number;
  age: string;
  _origin?: PodDisruptionBudget;
}

export interface PodDisruptionBudget extends K8sResource {
  spec: {
    minAvailable?: number | string;
    maxUnavailable?: number | string;
    selector?: LabelSelector;
  };
  status: {
    currentHealthy: number;
    desiredHealthy: number;
    expectedPods: number;
    conditions?: { type: string; status: string; lastTransitionTime?: string; reason?: string; message?: string }[];
    observedGeneration?: number;
  };
}


