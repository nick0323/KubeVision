import type { K8sResource, K8sMetadata, K8sLocalObjectReference, ResourceRequirements } from './core';

export interface ContainerPort {
  name?: string;
  containerPort: number;
  protocol?: 'TCP' | 'UDP' | 'SCTP';
  hostIP?: string;
  hostPort?: number;
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

export interface CronJob extends K8sResource {
  spec: {
    schedule: string;
    suspend?: boolean;
    jobTemplate?: any;
    successfulJobsHistoryLimit?: number;
    failedJobsHistoryLimit?: number;
  };
  status?: {
    active?: number;
    lastScheduleTime?: string;
    lastSuccessfulTime?: string;
  };
}

export interface CronJobListItem {
  name: string;
  namespace: string;
  schedule: string;
  suspend: boolean;
  active: number;
  lastScheduleTime: string;
  age: string;
}

export interface K8sJob extends K8sResource {
  spec: {
    completions?: number;
    parallelism?: number;
    backoffLimit?: number;
    activeDeadlineSeconds?: number;
  };
  status?: {
    active?: number;
    succeeded?: number;
    failed?: number;
    conditions?: any[];
    startTime?: string;
    completionTime?: string;
  };
}

export interface JobListItem {
  name: string;
  namespace: string;
  completions: string;
  duration: string;
  status: string;
  age: string;
}
