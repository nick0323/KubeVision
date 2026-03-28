/**
 * Pod 详情页类型定义
 */

import { Pod, K8sOwnerReference, Container, K8sEvent } from '../../types/k8s-resources';

/**
 * Pod 详情页 Props
 */
export interface PodDetailPageProps {
  namespace: string;
  name: string;
}

/**
 * Overview Tab Props
 */
export interface OverviewTabProps {
  pod: Pod | null;
  loading: boolean;
  onRefresh: () => void;
}

/**
 * YAML Tab Props
 */
export interface YamlTabProps {
  namespace: string;
  name: string;
  pod: Pod | null;
}

/**
 * Logs Tab Props
 */
export interface LogsTabProps {
  namespace: string;
  name: string;
  containers: Container[];
}

/**
 * Terminal Tab Props
 */
export interface TerminalTabProps {
  namespace: string;
  name: string;
  containers: Container[];
}

/**
 * Related Tab Props
 */
export interface RelatedTabProps {
  namespace: string;
  name: string;
  ownerReferences: K8sOwnerReference[];
}

/**
 * Events Tab Props
 */
export interface EventsTabProps {
  namespace: string;
  podName: string;
}

/**
 * 容器状态
 */
export interface ContainerState {
  name: string;
  waiting?: { reason?: string };
  running?: { startedAt?: string };
  terminated?: { reason?: string; exitCode?: number };
}

/**
 * 容器状态摘要
 */
export interface ContainerStatusSummary {
  name: string;
  ready: boolean;
  restartCount: number;
  state: ContainerState;
  image: string;
}

/**
 * Pod 条件
 */
export interface PodCondition {
  type: string;
  status: 'True' | 'False' | 'Unknown';
  lastTransitionTime?: string;
  reason?: string;
  message?: string;
}

/**
 * 关联资源
 */
export interface RelatedResource {
  kind: string;
  name: string;
  namespace?: string;
  apiVersion?: string;
}

/**
 * 日志选项
 */
export interface LogOptions {
  container?: string;
  since?: string;
  previous?: boolean;
  timestamps?: boolean;
  tailLines?: number;
}

/**
 * 终端选项
 */
export interface TerminalOptions {
  container?: string;
  shell: string;
}

/**
 * 事件统计
 */
export interface EventStats {
  total: number;
  normal: number;
  warning: number;
}
