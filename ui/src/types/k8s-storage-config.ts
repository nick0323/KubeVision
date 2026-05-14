import type { K8sResource, K8sMetadata, ResourceRequirements } from './core';

export interface ConfigMap extends K8sResource {
  data?: Record<string, string>;
  binaryData?: Record<string, string>;
  immutable?: boolean;
}

export interface ConfigMapListItem {
  name: string;
  namespace: string;
  age: string;
  dataCount: number;
}

export interface Secret extends K8sResource {
  type?: string;
  data?: Record<string, string>;
  stringData?: Record<string, string>;
  immutable?: boolean;
}

export interface SecretListItem {
  name: string;
  namespace: string;
  type: string;
  age: string;
  dataCount: number;
}

export interface PersistentVolumeClaim extends K8sResource {
  spec: {
    accessModes?: string[];
    volumeName?: string;
    storageClassName?: string;
    resources?: ResourceRequirements;
  };
  status?: { phase: string; capacity?: Record<string, string> };
}

export interface PVCListItem {
  name: string;
  namespace: string;
  status: string;
  volume: string;
  storageClass: string;
  capacity: string;
  age: string;
}

export interface PersistentVolume extends K8sResource {
  spec: {
    capacity?: Record<string, string>;
    accessModes?: string[];
    persistentVolumeReclaimPolicy?: string;
    storageClassName?: string;
    claimRef?: { namespace: string; name: string };
    nodeAffinity?: { required?: { nodeSelectorTerms: { matchExpressions: { key: string; operator: string; values?: string[] }[] }[] } };
  };
  status?: { phase: string; message?: string };
}

export interface PVListItem {
  name: string;
  capacity: string;
  accessModes: string;
  reclaimPolicy: string;
  status: string;
  storageClass: string;
  claim: string;
  age: string;
}

export interface StorageClass extends K8sResource {
  provisioner: string;
  reclaimPolicy?: string;
  volumeBindingMode?: string;
  allowVolumeExpansion?: boolean;
  parameters?: Record<string, string>;
}

export interface StorageClassListItem {
  name: string;
  provisioner: string;
  reclaimPolicy: string;
  volumeBindingMode: string;
  age: string;
}
