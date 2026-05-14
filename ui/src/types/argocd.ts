/**
 * ArgoCD 资源类型定义
 */

export interface ArgoCDApplication {
  metadata: {
    name: string;
    namespace: string;
    uid: string;
    resourceVersion: string;
    creationTimestamp: string;
    labels?: Record<string, string>;
    annotations?: Record<string, string>;
  };
  spec: {
    project: string;
    source: {
      repoURL: string;
      path?: string;
      targetRevision?: string;
      helm?: Record<string, unknown>;
      kustomize?: Record<string, unknown>;
    };
    destination: {
      server: string;
      namespace: string;
    };
    syncPolicy?: {
      automated?: {
        prune?: boolean;
        selfHeal?: boolean;
      };
      syncOptions?: string[];
    };
    ignoreDifferences?: Array<{
      group?: string;
      kind: string;
      name?: string;
      namespace?: string;
      jsonPointers?: string[];
      jqPathExpressions?: string[];
    }>;
  };
  status: {
    sync: {
      status: 'Synced' | 'OutOfSync' | 'Unknown';
      revision?: string;
      comparedTo: {
        source: {
          repoURL: string;
          path?: string;
          targetRevision?: string;
        };
        destination: {
          server: string;
          namespace: string;
        };
      };
    };
    health: {
      status: 'Healthy' | 'Progressing' | 'Suspended' | 'Degraded' | 'Missing' | 'Unknown';
      message?: string;
    };
    summary?: {
      externalURLs?: string[];
      images?: string[];
    };
    history?: Array<{
      revision: string;
      deployedAt: string;
      id: number;
      source: {
        repoURL: string;
        path?: string;
        targetRevision?: string;
      };
    }>;
    conditions?: Array<{
      type: string;
      message: string;
      lastTransitionTime?: string;
    }>;
    operationState?: {
      phase: 'Running' | 'Succeeded' | 'Failed' | 'Error';
      message?: string;
      startedAt: string;
      finishedAt?: string;
    };
  };
}

export interface ArgoCDAppProject {
  metadata: {
    name: string;
    namespace: string;
  };
  spec: {
    description?: string;
    sourceRepos: string[];
    destinations: Array<{
      server: string;
      namespace?: string;
    }>;
    clusterResourceWhitelist?: Array<{
      group: string;
      kind: string;
    }>;
    namespaceResourceWhitelist?: Array<{
      group: string;
      kind: string;
    }>;
    orphanedResources?: {
      warn?: boolean;
    };
    roles?: Array<{
      name: string;
      policies?: string[];
      description?: string;
    }>;
  };
}

export interface ArgoCDRepo {
  type: string;
  name: string;
  repo: string;
  connectionState: {
    status: 'Successful' | 'Failed';
    message?: string;
    attemptedAt?: string;
  };
  server: string;
}

// API 响应格式
export interface ArgoCDListResponse {
  code: number;
  message: string;
  data: ArgoCDApplication[];
}

export interface ArgoCDAppResponse {
  code: number;
  message: string;
  data: ArgoCDApplication;
}

export interface ArgoCDActionResponse {
  code: number;
  message: string;
  data?: Record<string, unknown>;
}
