/**
 * Kubernetes 资源类型定义
 * 提供完整的类型安全支持
 */

// ==================== 基础类型 ====================

/**
 * K8s 对象元数据
 */
export interface K8sObjectMeta {
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
 * 本地对象引用
 */
export interface K8sLocalObjectReference {
  name: string;
}

// ==================== 容器相关类型 ====================

/**
 * 容器镜像拉取策略
 */
export type ImagePullPolicy = 'Always' | 'Never' | 'IfNotPresent';

/**
 * 容器终止策略
 */
export type ContainerTerminationMessagePolicy = 'File' | 'FallbackToLogsOnError';

/**
 * 容器端口
 */
export interface ContainerPort {
  name?: string;
  containerPort: number;
  protocol?: 'TCP' | 'UDP' | 'SCTP';
  hostIP?: string;
  hostPort?: number;
}

/**
 * 资源需求
 */
export interface ResourceRequirements {
  limits?: Record<string, string>;
  requests?: Record<string, string>;
}

/**
 * 容器状态
 */
export interface ContainerState {
  name: string;
  state?: 'waiting' | 'running' | 'terminated';
  reason?: string;
  message?: string;
  startedAt?: string;
  finishedAt?: string;
  exitCode?: number;
  signal?: number;
  containerID?: string;
}

/**
 * 容器定义
 */
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
  lifecycle?: Lifecycle;
  terminationMessagePath?: string;
  terminationMessagePolicy?: ContainerTerminationMessagePolicy;
  imagePullPolicy?: ImagePullPolicy;
  securityContext?: SecurityContext;
  stdin?: boolean;
  stdinOnce?: boolean;
  tty?: boolean;
}

/**
 * 环境变量
 */
export interface EnvVar {
  name: string;
  value?: string;
  valueFrom?: EnvVarSource;
}

/**
 * 环境变量来源
 */
export interface EnvVarSource {
  configMapKeyRef?: ConfigMapKeySelector;
  secretKeyRef?: SecretKeySelector;
  fieldRef?: ObjectFieldSelector;
  resourceFieldRef?: ResourceFieldSelector;
}

/**
 * ConfigMap 引用
 */
export interface ConfigMapKeySelector {
  name: string;
  key: string;
  optional?: boolean;
}

/**
 * Secret 引用
 */
export interface SecretKeySelector {
  name: string;
  key: string;
  optional?: boolean;
}

/**
 * 对象字段选择器
 */
export interface ObjectFieldSelector {
  apiVersion?: string;
  fieldPath: string;
}

/**
 * 资源字段选择器
 */
export interface ResourceFieldSelector {
  containerName?: string;
  resource: string;
  divisor?: string;
}

/**
 * 探针（健康检查）
 */
export interface Probe {
  httpGet?: HTTPGetAction;
  tcpSocket?: TCPSocketAction;
  exec?: ExecAction;
  initialDelaySeconds?: number;
  timeoutSeconds?: number;
  periodSeconds?: number;
  successThreshold?: number;
  failureThreshold?: number;
  terminationGracePeriodSeconds?: number;
}

/**
 * HTTP 探测动作
 */
export interface HTTPGetAction {
  path?: string;
  port: number | string;
  host?: string;
  scheme?: 'HTTP' | 'HTTPS';
  httpHeaders?: HTTPHeader[];
}

/**
 * HTTP 头
 */
export interface HTTPHeader {
  name: string;
  value: string;
}

/**
 * TCP 探测动作
 */
export interface TCPSocketAction {
  port: number | string;
  host?: string;
}

/**
 * 执行探测动作
 */
export interface ExecAction {
  command?: string[];
}

/**
 * 生命周期钩子
 */
export interface Lifecycle {
  postStart?: LifecycleHandler;
  preStop?: LifecycleHandler;
}

/**
 * 生命周期处理器
 */
export interface LifecycleHandler {
  exec?: ExecAction;
  httpGet?: HTTPGetAction;
  tcpSocket?: TCPSocketAction;
}

/**
 * 安全上下文
 */
export interface SecurityContext {
  privileged?: boolean;
  runAsNonRoot?: boolean;
  runAsUser?: number;
  runAsGroup?: number;
  readOnlyRootFilesystem?: boolean;
  allowPrivilegeEscalation?: boolean;
  capabilities?: Capabilities;
  seccompProfile?: SeccompProfile;
}

/**
 * 能力
 */
export interface Capabilities {
  add?: string[];
  drop?: string[];
}

/**
 * Seccomp 配置
 */
export interface SeccompProfile {
  type: 'RuntimeDefault' | 'Localhost' | 'Unconfined';
  localhostProfile?: string;
}

// ==================== 卷相关类型 ====================

/**
 * 卷挂载
 */
export interface VolumeMount {
  name: string;
  mountPath: string;
  readOnly?: boolean;
  mountPropagation?: 'None' | 'HostToContainer' | 'Bidirectional';
  subPath?: string;
  subPathExpr?: string;
}

/**
 * 卷来源
 */
export interface EnvFromSource {
  prefix?: string;
  configMapRef?: K8sLocalObjectReference;
  secretRef?: K8sLocalObjectReference;
}

// ==================== Pod 相关类型 ====================

/**
 * Pod 状态阶段
 */
export type PodPhase = 'Pending' | 'Running' | 'Succeeded' | 'Failed' | 'Unknown';

/**
 * Pod 状态
 */
export interface PodStatus {
  phase: PodPhase;
  conditions: PodCondition[];
  message?: string;
  reason?: string;
  hostIP?: string;
  hostIPs?: HostIP[];
  podIP?: string;
  podIPs?: PodIP[];
  startTime?: string;
  initContainerStatuses?: ContainerStatus[];
  containerStatuses?: ContainerStatus[];
  qosClass?: 'Guaranteed' | 'Burstable' | 'BestEffort';
  nominatedNodeName?: string;
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
 * 容器状态
 */
export interface ContainerStatus {
  name: string;
  state?: ContainerStateInfo;
  lastState?: ContainerStateInfo;
  ready: boolean;
  restartCount: number;
  image: string;
  imageID: string;
  containerID?: string;
  started?: boolean;
}

/**
 * 容器状态信息
 */
export interface ContainerStateInfo {
  waiting?: ContainerStateWaiting;
  running?: ContainerStateRunning;
  terminated?: ContainerStateTerminated;
}

/**
 * 容器等待状态
 */
export interface ContainerStateWaiting {
  reason?: string;
  message?: string;
}

/**
 * 容器运行状态
 */
export interface ContainerStateRunning {
  startedAt?: string;
}

/**
 * 容器终止状态
 */
export interface ContainerStateTerminated {
  exitCode: number;
  signal?: number;
  reason?: string;
  message?: string;
  startedAt?: string;
  finishedAt?: string;
  containerID?: string;
}

/**
 * Host IP
 */
export interface HostIP {
  ip: string;
}

/**
 * Pod IP
 */
export interface PodIP {
  ip: string;
}

/**
 * Pod 规格
 */
export interface PodSpec {
  containers: Container[];
  initContainers?: Container[];
  volumes?: Volume[];
  restartPolicy?: 'Always' | 'OnFailure' | 'Never';
  terminationGracePeriodSeconds?: number;
  activeDeadlineSeconds?: number;
  dnsPolicy?: 'ClusterFirst' | 'Default' | 'ClusterFirstWithHostNet' | 'None';
  nodeSelector?: Record<string, string>;
  serviceAccountName?: string;
  serviceAccount?: string;
  automountServiceAccountToken?: boolean;
  nodeName?: string;
  hostNetwork?: boolean;
  hostPID?: boolean;
  hostIPC?: boolean;
  securityContext?: PodSecurityContext;
  imagePullSecrets?: K8sLocalObjectReference[];
  hostname?: string;
  subdomain?: string;
  affinity?: Affinity;
  tolerations?: Toleration[];
  priorityClassName?: string;
  priority?: number;
  preemptionPolicy?: 'PreemptLowerPriority' | 'Never';
}

/**
 * Pod 安全上下文
 */
export interface PodSecurityContext {
  runAsNonRoot?: boolean;
  runAsUser?: number;
  runAsGroup?: number;
  fsGroup?: number;
  supplementalGroups?: number[];
  seccompProfile?: SeccompProfile;
  sysctls?: Sysctl[];
}

/**
 * Sysctl 配置
 */
export interface Sysctl {
  name: string;
  value: string;
}

/**
 * 亲和性
 */
export interface Affinity {
  nodeAffinity?: NodeAffinity;
  podAffinity?: PodAffinity;
  podAntiAffinity?: PodAntiAffinity;
}

/**
 * 节点亲和性
 */
export interface NodeAffinity {
  requiredDuringSchedulingIgnoredDuringExecution?: NodeSelector;
  preferredDuringSchedulingIgnoredDuringExecution?: PreferredSchedulingTerm[];
}

/**
 * 节点选择器
 */
export interface NodeSelector {
  nodeSelectorTerms: NodeSelectorTerm[];
}

/**
 * 节点选择器项
 */
export interface NodeSelectorTerm {
  matchExpressions?: NodeSelectorRequirement[];
  matchFields?: NodeSelectorRequirement[];
}

/**
 * 节点选择器要求
 */
export interface NodeSelectorRequirement {
  key: string;
  operator: 'In' | 'NotIn' | 'Exists' | 'DoesNotExist' | 'Gt' | 'Lt';
  values?: string[];
}

/**
 * 首选调度项
 */
export interface PreferredSchedulingTerm {
  weight: number;
  preference: NodeSelectorTerm;
}

/**
 * Pod 亲和性
 */
export interface PodAffinity {
  requiredDuringSchedulingIgnoredDuringExecution?: PodAffinityTerm[];
  preferredDuringSchedulingIgnoredDuringExecution?: WeightedPodAffinityTerm[];
}

/**
 * Pod 反亲和性
 */
export interface PodAntiAffinity {
  requiredDuringSchedulingIgnoredDuringExecution?: PodAffinityTerm[];
  preferredDuringSchedulingIgnoredDuringExecution?: WeightedPodAffinityTerm[];
}

/**
 * Pod 亲和性项
 */
export interface PodAffinityTerm {
  labelSelector?: LabelSelector;
  namespaces?: string[];
  topologyKey: string;
  namespaceSelector?: LabelSelector;
}

/**
 * 带权重的 Pod 亲和性项
 */
export interface WeightedPodAffinityTerm {
  weight: number;
  podAffinityTerm: PodAffinityTerm;
}

/**
 * 标签选择器
 */
export interface LabelSelector {
  matchLabels?: Record<string, string>;
  matchExpressions?: LabelSelectorRequirement[];
}

/**
 * 标签选择器要求
 */
export interface LabelSelectorRequirement {
  key: string;
  operator: 'In' | 'NotIn' | 'Exists' | 'DoesNotExist';
  values?: string[];
}

/**
 * 容忍
 */
export interface Toleration {
  key?: string;
  operator?: 'Equal' | 'Exists';
  value?: string;
  effect?: 'NoSchedule' | 'PreferNoSchedule' | 'NoExecute';
  tolerationSeconds?: number;
}

/**
 * 卷
 */
export interface Volume {
  name: string;
  hostPath?: HostPathVolumeSource;
  emptyDir?: EmptyDirVolumeSource;
  configMap?: ConfigMapVolumeSource;
  secret?: SecretVolumeSource;
  persistentVolumeClaim?: PersistentVolumeClaimVolumeSource;
  downwardAPI?: DownwardAPIVolumeSource;
  projected?: ProjectedVolumeSource;
  nfs?: NFSVolumeSource;
  awsElasticBlockStore?: AWSElasticBlockStoreVolumeSource;
  azureDisk?: AzureDiskVolumeSource;
  azureFile?: AzureFileVolumeSource;
  cephfs?: CephFSVolumeSource;
  cinder?: CinderVolumeSource;
  csi?: CSIVolumeSource;
  ephemeral?: EphemeralVolumeSource;
  fc?: FCVolumeSource;
  flexVolume?: FlexVolumeSource;
  flocker?: FlockerVolumeSource;
  gcePersistentDisk?: GCEPersistentDiskVolumeSource;
  gitRepo?: GitRepoVolumeSource;
  glusterfs?: GlusterfsVolumeSource;
  iscsi?: ISCSIVolumeSource;
  local?: LocalVolumeSource;
  photonPersistentDisk?: PhotonPersistentDiskVolumeSource;
  portworxVolume?: PortworxVolumeSource;
  quobyte?: QuobyteVolumeSource;
  rbd?: RBDVolumeSource;
  scaleIO?: ScaleIOVolumeSource;
  storageos?: StorageOSVolumeSource;
  vsphereVolume?: VsphereVirtualDiskVolumeSource;
}

// ==================== 工作负载类型 ====================

/**
 * Deployment 规格
 */
export interface DeploymentSpec {
  replicas?: number;
  selector: LabelSelector;
  template: PodTemplateSpec;
  strategy?: DeploymentStrategy;
  minReadySeconds?: number;
  revisionHistoryLimit?: number;
  paused?: boolean;
  progressDeadlineSeconds?: number;
}

/**
 * Deployment 策略
 */
export interface DeploymentStrategy {
  type?: 'Recreate' | 'RollingUpdate';
  rollingUpdate?: RollingUpdateDeployment;
}

/**
 * 滚动更新配置
 */
export interface RollingUpdateDeployment {
  maxUnavailable?: number | string;
  maxSurge?: number | string;
}

/**
 * Pod 模板
 */
export interface PodTemplateSpec {
  metadata: K8sObjectMeta;
  spec: PodSpec;
}

/**
 * Deployment 状态
 */
export interface DeploymentStatus {
  replicas: number;
  updatedReplicas: number;
  readyReplicas: number;
  availableReplicas: number;
  unavailableReplicas?: number;
  observedGeneration?: number;
  conditions?: DeploymentCondition[];
  collisionCount?: number;
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
 * StatefulSet 规格
 */
export interface StatefulSetSpec {
  replicas?: number;
  selector: LabelSelector;
  template: PodTemplateSpec;
  serviceName: string;
  podManagementPolicy?: 'OrderedReady' | 'Parallel';
  updateStrategy?: StatefulSetUpdateStrategy;
  volumeClaimTemplates?: PersistentVolumeClaim[];
  minReadySeconds?: number;
  revisionHistoryLimit?: number;
}

/**
 * StatefulSet 更新策略
 */
export interface StatefulSetUpdateStrategy {
  type?: 'RollingUpdate' | 'OnDelete';
  rollingUpdate?: RollingUpdateStatefulSetStrategy;
}

/**
 * StatefulSet 滚动更新
 */
export interface RollingUpdateStatefulSetStrategy {
  partition?: number;
  maxUnavailable?: number | string;
}

/**
 * DaemonSet 规格
 */
export interface DaemonSetSpec {
  selector: LabelSelector;
  template: PodTemplateSpec;
  updateStrategy?: DaemonSetUpdateStrategy;
  minReadySeconds?: number;
  revisionHistoryLimit?: number;
}

/**
 * DaemonSet 更新策略
 */
export interface DaemonSetUpdateStrategy {
  type?: 'RollingUpdate' | 'OnDelete';
  rollingUpdate?: RollingUpdateDaemonSet;
}

/**
 * DaemonSet 滚动更新
 */
export interface RollingUpdateDaemonSet {
  maxUnavailable?: number | string;
  maxSurge?: number | string;
}

/**
 * DaemonSet 状态
 */
export interface DaemonSetStatus {
  currentNumberScheduled: number;
  numberMisscheduled: number;
  desiredNumberScheduled: number;
  numberReady: number;
  updatedNumberScheduled?: number;
  numberAvailable?: number;
  numberUnavailable?: number;
  collisionCount?: number;
  observedGeneration?: number;
  conditions?: DaemonSetCondition[];
}

/**
 * DaemonSet 条件
 */
export interface DaemonSetCondition {
  type: 'DisruptionRestricted' | string;
  status: 'True' | 'False' | 'Unknown';
  lastTransitionTime?: string;
  reason?: string;
  message?: string;
}

// ==================== Service 相关类型 ====================

/**
 * Service 类型
 */
export type ServiceType = 'ClusterIP' | 'NodePort' | 'LoadBalancer' | 'ExternalName';

/**
 * Service 协议
 */
export type ServiceProtocol = 'TCP' | 'UDP' | 'SCTP';

/**
 * Service 规格
 */
export interface ServiceSpec {
  type?: ServiceType;
  selector?: Record<string, string>;
  ports: ServicePort[];
  clusterIP?: string;
  clusterIPs?: string[];
  externalIPs?: string[];
  loadBalancerIP?: string;
  loadBalancerSourceRanges?: string[];
  externalName?: string;
  externalTrafficPolicy?: 'Cluster' | 'Local';
  sessionAffinity?: 'None' | 'ClientIP';
  healthCheckNodePort?: number;
  publishNotReadyAddresses?: boolean;
  ipFamilies?: 'IPv4' | 'IPv6'[];
  ipFamilyPolicy?: 'SingleStack' | 'PreferDualStack' | 'RequireDualStack';
  allocateLoadBalancerNodePorts?: boolean;
}

/**
 * Service 端口
 */
export interface ServicePort {
  name?: string;
  protocol?: ServiceProtocol;
  port: number;
  targetPort?: number | string;
  nodePort?: number;
  appProtocol?: string;
}

/**
 * Service 状态
 */
export interface ServiceStatus {
  loadBalancer?: LoadBalancerStatus;
}

/**
 * 负载均衡器状态
 */
export interface LoadBalancerStatus {
  ingress?: LoadBalancerIngress[];
}

/**
 * 负载均衡器入口
 */
export interface LoadBalancerIngress {
  ip?: string;
  hostname?: string;
  ports?: PortStatus[];
}

/**
 * 端口状态
 */
export interface PortStatus {
  port: number;
  protocol: ServiceProtocol;
  error?: string;
}

// ==================== 存储相关类型 ====================

/**
 * PVC 状态
 */
export interface PersistentVolumeClaimSpec {
  accessModes?: ('ReadWriteOnce' | 'ReadOnlyMany' | 'ReadWriteMany' | 'ReadWriteOncePod')[];
  selector?: LabelSelector;
  resources?: ResourceRequirements;
  volumeName?: string;
  storageClassName?: string;
  volumeMode?: 'Filesystem' | 'Block';
  dataSource?: K8sTypedLocalObjectReference;
  dataSourceRef?: K8sTypedLocalObjectReference;
}

/**
 * PVC 状态
 */
export interface PersistentVolumeClaimStatus {
  phase: 'Pending' | 'Bound' | 'Lost';
  accessModes?: ('ReadWriteOnce' | 'ReadOnlyMany' | 'ReadWriteMany' | 'ReadWriteOncePod')[];
  capacity?: Record<string, string>;
  conditions?: PVCCondition[];
  allocatedResources?: Record<string, string>;
  resizeStatus?: string;
}

/**
 * PVC 条件
 */
export interface PVCCondition {
  type: 'Resizing' | 'FileSystemResizePending' | string;
  status: 'True' | 'False' | 'Unknown';
  lastTransitionTime?: string;
  reason?: string;
  message?: string;
}

/**
 * 带类型的本地对象引用
 */
export interface K8sTypedLocalObjectReference {
  apiGroup?: string;
  kind: string;
  name: string;
}

// ==================== K8s 事件类型 ====================

/**
 * K8s 事件
 */
export interface K8sEvent {
  apiVersion: string;
  kind: string;
  metadata: K8sObjectMeta;
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

/**
 * 对象引用
 */
export interface ObjectReference {
  apiVersion?: string;
  kind?: string;
  namespace?: string;
  name: string;
  uid?: string;
  resourceVersion?: string;
  fieldPath?: string;
}

/**
 * 事件来源
 */
export interface EventSource {
  component?: string;
  host?: string;
}

/**
 * 事件序列
 */
export interface EventSeries {
  count: number;
  lastObservedTime: string;
}

// ==================== 通用 UI 类型 ====================

/**
 * 资源列表项（UI 层）
 */
export interface ResourceListItem {
  name: string;
  namespace?: string;
  status?: string;
  age?: string;
  [key: string]: any;
}

/**
 * 操作按钮配置
 */
export interface ActionButton {
  label: string;
  icon?: React.ReactNode;
  onClick: (record: any) => void;
  confirm?: boolean;
  confirmMessage?: string;
  disabled?: boolean;
  permission?: string;
  danger?: boolean;
}
