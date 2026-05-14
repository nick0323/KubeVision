import type { K8sResource, K8sMetadata, K8sLocalObjectReference } from './core';

export interface PolicyRule {
  verbs: string[];
  apiGroups?: string[];
  resources?: string[];
  resourceNames?: string[];
  nonResourceURLs?: string[];
}

export interface Subject {
  kind: string;
  apiGroup?: string;
  name: string;
  namespace?: string;
}

export interface RoleRef {
  apiGroup: string;
  kind: string;
  name: string;
}

export interface RoleListItem {
  name: string;
  namespace: string;
  rules: number;
  age: string;
  _origin?: Role;
}

export interface Role extends K8sResource {
  metadata: K8sMetadata & { namespace: string };
  rules?: PolicyRule[];
}

export interface RoleBindingListItem {
  name: string;
  namespace: string;
  roleRef: string;
  subjects: string;
  age: string;
  _origin?: RoleBinding;
}

export interface RoleBinding extends K8sResource {
  metadata: K8sMetadata & { namespace: string };
  subjects?: Subject[];
  roleRef: RoleRef;
}

export interface ClusterRoleListItem {
  name: string;
  rules: number;
  age: string;
  _origin?: ClusterRole;
}

export interface ClusterRole extends K8sResource {
  metadata: K8sMetadata;
  rules?: PolicyRule[];
}

export interface ClusterRoleBindingListItem {
  name: string;
  roleRef: string;
  subjects: string;
  age: string;
  _origin?: ClusterRoleBinding;
}

export interface ClusterRoleBinding extends K8sResource {
  metadata: K8sMetadata;
  subjects?: Subject[];
  roleRef: RoleRef;
}

export interface ServiceAccountListItem {
  name: string;
  namespace: string;
  secrets: number;
  age: string;
  _origin?: ServiceAccount;
}

export interface ServiceAccount extends K8sResource {
  metadata: K8sMetadata & { namespace: string };
  secrets?: K8sLocalObjectReference[];
  automountServiceAccountToken?: boolean;
  imagePullSecrets?: K8sLocalObjectReference[];
}
