import React from 'react';
import type { APIResponse as APIResponseType, PageMeta as PageMetaType } from './index';

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

export interface APIErrorResponse {
  code: number;
  message: string;
  details?: { field?: string; reason?: string }[];
  traceId?: string;
  timestamp?: number;
}

export interface PaginatedResponse<T> {
  data: T[];
  page?: PageMetaType;
}

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

export type Column<T = unknown> = ColumnDef<T>;

export interface GenericResourceItem {
  name: string;
  namespace?: string;
  status?: string;
  age: string;
  [key: string]: string | undefined;
}

export interface ResourceListItem {
  name: string;
  namespace?: string;
  status?: string;
  age?: string;
  [key: string]: string | undefined;
}

export interface ActionButton<T = Record<string, unknown>> {
  label: string;
  icon?: React.ReactNode;
  onClick: (record: T) => void;
  confirm?: boolean;
  confirmMessage?: string;
  disabled?: boolean;
  danger?: boolean;
}

export interface ResourceRequirements {
  limits?: Record<string, string>;
  requests?: Record<string, string>;
}

export interface MetricTarget {
  type: 'Utilization' | 'Value' | 'AverageValue';
  averageUtilization?: number;
  averageValue?: string;
  value?: string;
}

export interface MetricValueStatus {
  averageUtilization?: number;
  averageValue?: string;
  value?: string;
}

export interface ResourceMetricSource {
  name: string;
  target: MetricTarget;
}

export interface ResourceMetricStatus {
  name: string;
  current: MetricValueStatus;
}

export interface MetricSpec {
  type: 'Resource' | 'Pods' | 'Object' | 'External';
  resource?: ResourceMetricSource;
  pods?: { metric: { name: string }; target: MetricTarget };
  object?: { metric: { name: string }; target: MetricTarget };
  external?: { metric: { name: string }; target: MetricTarget };
}

export interface MetricStatus {
  type: 'Resource' | 'Pods' | 'Object' | 'External';
  resource?: ResourceMetricStatus;
  pods?: { metric: { name: string }; current: MetricValueStatus };
  object?: { metric: { name: string }; current: { value?: string; averageValue?: string } };
  external?: { metric: { name: string }; current: { value?: string; averageValue?: string } };
}

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
