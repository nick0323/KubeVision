import React from 'react';
import Tippy from '@tippyjs/react';
import { truncateText } from '../../utils/string';

export interface FieldDef {
  key: string;
  label: string;
  getValue: (data: any) => any;
  condition?: (data: any) => boolean;
  render?: (value: any, data: any) => React.ReactNode;
}

export const RESOURCE_CONFIG: Record<
  string,
  { title: string; hasContainers?: boolean; hasPorts?: boolean }
> = {
  pod: { title: 'Pod', hasContainers: true },
  deployment: { title: 'Deployment', hasContainers: true },
  statefulset: { title: 'StatefulSet', hasContainers: true },
  daemonset: { title: 'DaemonSet', hasContainers: true },
  service: { title: 'Service', hasPorts: true },
  configmap: { title: 'ConfigMap' },
  secret: { title: 'Secret' },
  ingress: { title: 'Ingress' },
  job: { title: 'Job', hasContainers: true },
  cronjob: { title: 'CronJob' },
  pvc: { title: 'PersistentVolumeClaim' },
  pv: { title: 'PersistentVolume' },
  storageclass: { title: 'StorageClass' },
  namespace: { title: 'Namespace' },
  node: { title: 'Node' },
};

function getNodeRoles(labels?: Record<string, string>): string[] {
  if (!labels) return [];
  const roles: string[] = [];
  Object.keys(labels).forEach(key => {
    if (key.startsWith('node-role.kubernetes.io/')) {
      const role = key.replace('node-role.kubernetes.io/', '');
      if (role) roles.push(role);
    }
  });
  if (labels['node-role.kubernetes.io/control-plane']) roles.push('control-plane');
  if (labels['node-role.kubernetes.io/master']) roles.push('master');
  if (labels['node-role.kubernetes.io/worker']) roles.push('worker');
  if (labels['node-role.kubernetes.io/infra']) roles.push('infra');
  return [...new Set(roles)];
}

function renderLabelsInline(labels?: Record<string, string>): React.ReactNode {
  if (!labels || Object.keys(labels).length === 0) return null;
  return (
    <div className="label-list">
      {Object.entries(labels).map(([key, value]) => {
        const fullText = `${key}: ${value}`;
        const displayKey = truncateText(key as string, 20);
        const displayValue = truncateText(value as string, 20);
        const isTruncated = displayKey !== key || displayValue !== value;
        const labelElement = (
          <span className="label-tag">
            <span className="label-key">{displayKey}</span>
            <span className="label-separator">: </span>
            <span className="label-value">{displayValue}</span>
          </span>
        );
        if (isTruncated) {
          return (
            <Tippy key={key} content={fullText} theme="light" placement="top" arrow={true} duration={200}>
              {labelElement}
            </Tippy>
          );
        }
        return <span key={key}>{labelElement}</span>;
      })}
    </div>
  );
}

export const RESOURCE_FIELDS: Record<string, FieldDef[]> = {
  deployment: [
    {
      key: 'strategy',
      label: 'Strategy',
      getValue: data => data.spec?.strategy?.type,
    },
    {
      key: 'selector',
      label: 'Selector',
      getValue: data => data.spec?.selector?.matchLabels,
      render: (value: any) => renderLabelsInline(value),
    },
  ],
  statefulset: [
    {
      key: 'serviceName',
      label: 'Service Name',
      getValue: data => data.spec?.serviceName,
    },
    {
      key: 'selector',
      label: 'Selector',
      getValue: data => data.spec?.selector?.matchLabels,
      render: (value: any) => renderLabelsInline(value),
    },
  ],
  daemonset: [
    {
      key: 'selector',
      label: 'Selector',
      getValue: data => data.spec?.selector?.matchLabels,
      render: (value: any) => renderLabelsInline(value),
    },
  ],
  job: [
    {
      key: 'completions',
      label: 'Completions',
      getValue: data => data.spec?.completions || 1,
    },
    {
      key: 'parallelism',
      label: 'Parallelism',
      getValue: data => data.spec?.parallelism || 1,
    },
  ],
  cronjob: [
    {
      key: 'schedule',
      label: 'Schedule',
      getValue: data => data.spec?.schedule,
    },
    {
      key: 'suspend',
      label: 'Suspend',
      getValue: data => data.spec?.suspend,
      render: (value: boolean) => (value ? 'Yes' : 'No'),
    },
  ],
  service: [
    {
      key: 'type',
      label: 'Type',
      getValue: data => data.spec?.type,
    },
    {
      key: 'clusterIP',
      label: 'Cluster IP',
      getValue: data => data.spec?.clusterIP,
    },
    {
      key: 'selector',
      label: 'Selector',
      getValue: data => data.spec?.selector,
      condition: data => data.spec?.selector && Object.keys(data.spec.selector).length > 0,
      render: (value: any) => renderLabelsInline(value),
    },
  ],
  ingress: [
    {
      key: 'ingressClassName',
      label: 'Class',
      getValue: data => data.spec?.ingressClassName,
    },
    {
      key: 'rules',
      label: 'Rules',
      getValue: data => data.spec?.rules?.length || 0,
      render: (value: number) => `${value} rule${value !== 1 ? 's' : ''}`,
    },
  ],
  configmap: [
    {
      key: 'dataCount',
      label: 'Data Count',
      getValue: data => (data.data ? Object.keys(data.data).length : 0),
    },
  ],
  secret: [
    {
      key: 'type',
      label: 'Type',
      getValue: data => data.type,
    },
    {
      key: 'dataCount',
      label: 'Data Count',
      getValue: data => (data.data ? Object.keys(data.data).length : 0),
    },
  ],
  pvc: [
    {
      key: 'accessModes',
      label: 'Access Modes',
      getValue: data => data.spec?.accessModes?.[0],
    },
    {
      key: 'storageClass',
      label: 'Storage Class',
      getValue: data => data.spec?.storageClassName,
    },
    {
      key: 'volume',
      label: 'Volume',
      getValue: data => data.spec?.volumeName,
      condition: data => !!data.spec?.volumeName,
    },
  ],
  pv: [
    {
      key: 'accessModes',
      label: 'Access Modes',
      getValue: data => data.spec?.accessModes?.[0],
    },
    {
      key: 'storageClass',
      label: 'Storage Class',
      getValue: data => data.spec?.storageClassName,
    },
    {
      key: 'claim',
      label: 'Claim',
      getValue: data => {
        const claimRef = data.spec?.claimRef;
        if (claimRef?.namespace && claimRef?.name) {
          return `${claimRef.namespace}/${claimRef.name}`;
        }
        return claimRef?.name;
      },
      condition: data => !!data.spec?.claimRef?.name,
    },
  ],
  storageclass: [
    {
      key: 'provisioner',
      label: 'Provisioner',
      getValue: data => data.provisioner,
    },
    {
      key: 'bindingMode',
      label: 'Binding Mode',
      getValue: data => data.volumeBindingMode,
    },
    {
      key: 'isDefault',
      label: 'Default',
      getValue: data =>
        data.metadata?.annotations?.['storageclass.kubernetes.io/is-default-class'] === 'true',
      render: (value: boolean) => (value ? 'Yes' : 'No'),
      condition: data =>
        data.metadata?.annotations?.['storageclass.kubernetes.io/is-default-class'] !== undefined,
    },
  ],
  namespace: [
    {
      key: 'phase',
      label: 'Phase',
      getValue: data => data.status?.phase,
    },
  ],
  node: [
    {
      key: 'kernelVersion',
      label: 'Kernel Version',
      getValue: data => data.status?.nodeInfo?.kernelVersion,
    },
    {
      key: 'kubeletVersion',
      label: 'Kubelet Version',
      getValue: data => data.status?.nodeInfo?.kubeletVersion,
    },
    {
      key: 'kubeProxyVersion',
      label: 'KubeProxy Version',
      getValue: data => data.status?.nodeInfo?.kubeProxyVersion,
    },
    {
      key: 'os',
      label: 'OS',
      getValue: data => data.status?.nodeInfo?.osImage,
    },
    {
      key: 'architecture',
      label: 'Architecture',
      getValue: data => data.status?.nodeInfo?.architecture,
    },
    {
      key: 'operatingSystem',
      label: 'Operating System',
      getValue: data => data.status?.nodeInfo?.operatingSystem,
    },
    {
      key: 'roles',
      label: 'Roles',
      getValue: data => getNodeRoles(data.metadata?.labels),
      condition: data => {
        const roles = getNodeRoles(data.metadata?.labels);
        return roles && roles.length > 0;
      },
      render: (value: string[]) => value.join(', '),
    },
  ],
};
