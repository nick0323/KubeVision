import type { K8sResource, K8sMetadata } from './core';

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

export interface Ingress extends K8sResource {
  spec: {
    ingressClassName?: string;
    rules?: {
      host?: string;
      http?: {
        paths: {
          path: string;
          pathType: string;
          backend: any;
        }[];
      };
    }[];
    tls?: { hosts: string[]; secretName: string }[];
  };
  status?: { loadBalancer?: { ingress?: { ip?: string; hostname?: string }[] } };
}

export interface IngressListItem {
  name: string;
  namespace: string;
  hosts: string;
  age: string;
}

export interface NetworkPolicyListItem {
  name: string;
  namespace: string;
  podSelector: string;
  policyTypes: string[];
  age: string;
  _origin?: NetworkPolicy;
}

export interface NetworkPolicy extends K8sResource {
  metadata: K8sMetadata & { namespace: string };
  spec: {
    podSelector: { matchLabels?: Record<string, string> };
    policyTypes?: string[];
    ingress?: any[];
    egress?: any[];
  };
}
