import React, { useState, useMemo } from 'react';
import Tippy from '@tippyjs/react';
import 'tippy.js/dist/tippy.css';
import 'tippy.js/themes/light.css';
import { OverviewTabProps } from '../resources/types';
import { StatusBadge } from '../common/StatusBadge';
import './OverviewTab.css';
import '../styles/detail-page.css';

/**
 * 资源类型配置
 */
const RESOURCE_CONFIG: Record<
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

/**
 * 各资源类型特有字段配置
 */
const RESOURCE_FIELDS: Record<
  string,
  Array<{
    key: string;
    label: string;
    getValue: (data: any) => any;
    condition?: (data: any) => boolean;
    render?: (value: any, data: any) => React.ReactNode;
  }>
> = {
  // Deployment
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

  // StatefulSet
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

  // DaemonSet
  daemonset: [
    {
      key: 'selector',
      label: 'Selector',
      getValue: data => data.spec?.selector?.matchLabels,
      render: (value: any) => renderLabelsInline(value),
    },
  ],

  // Job
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

  // CronJob
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

  // Service
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

  // Ingress
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

  // ConfigMap
  configmap: [
    {
      key: 'dataCount',
      label: 'Data Count',
      getValue: data => (data.data ? Object.keys(data.data).length : 0),
    },
  ],

  // Secret
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

  // PVC
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

  // PV
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

  // StorageClass
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

  // Namespace
  namespace: [
    {
      key: 'phase',
      label: 'Phase',
      getValue: data => data.status?.phase,
    },
  ],

  // Node
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

/**
 * 获取节点角色
 * K8s 节点角色标签格式：node-role.kubernetes.io/<role>: "" (空字符串)
 */
function getNodeRoles(labels?: Record<string, string>): string[] {
  if (!labels) return [];

  const roles: string[] = [];

  // 提取所有 node-role.kubernetes.io/ 开头的标签
  Object.keys(labels).forEach(key => {
    if (key.startsWith('node-role.kubernetes.io/')) {
      const role = key.replace('node-role.kubernetes.io/', '');
      if (role) {
        roles.push(role);
      }
    }
  });

  // 兼容旧版本标签（空值标签）
  if (labels['node-role.kubernetes.io/control-plane']) roles.push('control-plane');
  if (labels['node-role.kubernetes.io/master']) roles.push('master');
  if (labels['node-role.kubernetes.io/worker']) roles.push('worker');
  if (labels['node-role.kubernetes.io/infra']) roles.push('infra');

  // 去重
  const uniqueRoles = [...new Set(roles)];

  return uniqueRoles.length > 0 ? uniqueRoles : [];
}

/**
 * 渲染内联标签（用于 Selector 展示，格式：key: value）
 */
function renderLabelsInline(labels?: Record<string, string>): React.ReactNode {
  if (!labels || Object.keys(labels).length === 0) return null;

  return (
    <div className="label-list">
      {Object.entries(labels).map(([key, value]) => {
        const fullText = `${key}: ${value}`;
        const displayKey = truncateText(key as string, 20);
        const displayValue = truncateText(value as string, 20);

        // 只有当文本被截断时才显示 tooltip
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
            <Tippy
              key={key}
              content={fullText}
              theme="light"
              placement="top"
              arrow={true}
              duration={200}
            >
              {labelElement}
            </Tippy>
          );
        }

        return <span key={key}>{labelElement}</span>;
      })}
    </div>
  );
}

/**
 * 截断文本
 */
const truncateText = (text: string, maxLength: number) => {
  if (!text) return '';
  if (text.length <= maxLength) return text;
  return text.substring(0, maxLength) + '...';
};

/**
 * 格式化相对时间
 */
const formatRelativeTime = (timestamp?: string) => {
  if (!timestamp) return '-';
  const date = new Date(timestamp);
  const now = new Date();
  const diff = now.getTime() - date.getTime();
  const minutes = Math.floor(diff / 60000);
  const hours = Math.floor(minutes / 60);
  const days = Math.floor(hours / 24);

  if (days > 0) return `${days}d ago`;
  if (hours > 0) return `${hours}h ago`;
  if (minutes > 0) return `${minutes}m ago`;
  return 'Just now';
};

/**
 * 通用资源概览 Tab
 */
export const OverviewTab: React.FC<OverviewTabProps> = ({
  data,
  loading,
  resourceType = 'pod',
}) => {
  const [containersExpanded, setContainersExpanded] = useState<Record<string, boolean>>({});

  const resourceInfo = RESOURCE_CONFIG[resourceType] || { title: resourceType };

  // 提取 metadata、spec、status
  const metadata = data?.metadata || {};
  const spec = data?.spec || {};
  const status = data?.status || {};

  // Pod 特有的容器状态
  const containerStatuses = useMemo(() => {
    if (resourceType !== 'pod' || !data?.status?.containerStatuses) return [];
    return data.status.containerStatuses.map((cs: any) => ({
      name: cs.name,
      ready: cs.ready,
      restartCount: cs.restartCount,
      state: cs.state || {},
      image: cs.image,
    }));
  }, [data, resourceType]);

  // Workload 资源的容器信息（从 Pod Template 中提取）
  const workloadContainers = useMemo(() => {
    if (!resourceInfo.hasContainers) return [];

    // 从 spec.template.spec.containers 中提取（适用于 Deployment、StatefulSet、DaemonSet、Job）
    const podSpec = spec?.template?.spec || spec;
    const containers = podSpec?.containers || [];

    return containers.map((c: any) => ({
      name: c.name,
      image: c.image,
      ports: c.ports,
      env: c.env,
      resources: c.resources,
      livenessProbe: c.livenessProbe,
      readinessProbe: c.readinessProbe,
      volumeMounts: c.volumeMounts,
      imagePullPolicy: c.imagePullPolicy,
    }));
  }, [data, spec, resourceInfo, resourceType]);

  // 计算 Ready Containers
  const readyContainers = useMemo(() => {
    if (resourceType !== 'pod') return null;
    const ready = containerStatuses.filter((c: any) => c.ready).length;
    const total = containerStatuses.length || spec?.containers?.length || 0;
    return { ready, total };
  }, [containerStatuses, spec, resourceType]);

  // 计算 Restart Count
  const restartCount = useMemo(() => {
    if (resourceType !== 'pod') return 0;
    return containerStatuses.reduce((sum: number, c: any) => sum + c.restartCount, 0);
  }, [containerStatuses, resourceType]);

  // Conditions
  const conditions = useMemo(() => {
    return status?.conditions || [];
  }, [status]);

  // 渲染容器详情
  const renderContainerDetails = (container: any) => {
    return (
      <div className="container-card-body">
        {/* PORTS */}
        <div className="sub-module">
          <div className="sub-module-title">Ports</div>
          <div className="info-grid">
            {container?.ports?.map((port: any, idx: number) => (
              <div key={idx} className="info-item">
                <span className="info-value">
                  {port.name ? `${port.name}: ` : ''}
                  {port.containerPort}/{port.protocol || 'TCP'}
                </span>
              </div>
            )) || (
              <div className="empty-state">
                <span className="empty-state-text">No ports defined</span>
              </div>
            )}
          </div>
        </div>

        {/* ENVIRONMENT VARIABLES */}
        <div className="sub-module">
          <div className="sub-module-title">Environment Variables</div>
          <div className="info-grid">
            {container?.env?.map((env: any, idx: number) => (
              <div key={idx} className="info-item">
                <span className="info-value">
                  <span className="env-key">{env.name}</span>
                  <span className="env-separator">: </span>
                  <span className="env-value">
                    {env.value || (env.valueFrom ? '(From ConfigMap/Secret)' : '-')}
                  </span>
                </span>
              </div>
            )) || (
              <div className="empty-state">
                <span className="empty-state-text">No environment variables</span>
              </div>
            )}
          </div>
        </div>

        {/* RESOURCES */}
        <div className="sub-module">
          <div className="sub-module-title">Resources</div>
          <div className="info-grid">
            <div className="info-item">
              <span className="info-label">Requests (CPU)</span>
              <span className="info-value">{container?.resources?.requests?.cpu || '-'}</span>
            </div>
            <div className="info-item">
              <span className="info-label">Requests (Memory)</span>
              <span className="info-value">{container?.resources?.requests?.memory || '-'}</span>
            </div>
            <div className="info-item">
              <span className="info-label">Limits (CPU)</span>
              <span className="info-value">{container?.resources?.limits?.cpu || '-'}</span>
            </div>
            <div className="info-item">
              <span className="info-label">Limits (Memory)</span>
              <span className="info-value">{container?.resources?.limits?.memory || '-'}</span>
            </div>
          </div>
        </div>

        {/* HEALTH CHECKS */}
        <div className="sub-module">
          <div className="sub-module-title">Health Checks</div>
          <div className="info-grid">
            <div className="info-item">
              <span className="info-label">Liveness Probe</span>
              <span className="info-value">
                {container?.livenessProbe ? 'Configured' : 'Not Configured'}
              </span>
            </div>
            <div className="info-item">
              <span className="info-label">Readiness Probe</span>
              <span className="info-value">
                {container?.readinessProbe ? 'Configured' : 'Not Configured'}
              </span>
            </div>
          </div>
        </div>

        {/* VOLUMES */}
        <div className="sub-module">
          <div className="sub-module-title">Volumes</div>
          <div className="info-grid">
            {container?.volumeMounts?.map((mount: any, idx: number) => (
              <div key={idx} className="info-item">
                <span className="info-label">{mount.name}</span>
                <span className="info-value">
                  {mount.mountPath} {mount.readOnly ? '(RO)' : '(RW)'}
                </span>
              </div>
            )) || (
              <div className="empty-state">
                <span className="empty-state-text">No volumes</span>
              </div>
            )}
          </div>
        </div>
      </div>
    );
  };

  // 判断是否需要显示 Status Overview
  const hasStatusOverview = useMemo(() => {
    // 只有 Pod、Deployment、StatefulSet 显示状态概览
    return ['pod', 'deployment', 'statefulset'].includes(resourceType);
  }, [resourceType]);

  if (loading || !data) {
    return <div className="overview-tab-loading">加载中...</div>;
  }

  return (
    <div className="overview-tab">
      {/* STATUS OVERVIEW - 只有有状态的资源才显示 */}
      {hasStatusOverview && (
        <div className="detail-card">
          <h3 className="detail-card-title">Status Overview</h3>
          <div className="detail-card-body">
            <div className="stats-grid">
              {/* Pod 特有的状态 */}
              {resourceType === 'pod' ? (
                <>
                  <div className="stat-card">
                    <div className="stat-value">
                      <StatusBadge status={status.phase || 'Unknown'} resourceType={resourceType} />
                    </div>
                    <div className="stat-label">Phase</div>
                  </div>
                  {readyContainers && (
                    <div className="stat-card">
                      <div className="stat-value">
                        {readyContainers.ready} / {readyContainers.total}
                      </div>
                      <div className="stat-label">Ready Containers</div>
                    </div>
                  )}
                  <div className="stat-card">
                    <div className="stat-value">{restartCount}</div>
                    <div className="stat-label">Restart Count</div>
                  </div>
                  <div className="stat-card">
                    <div className="stat-value">{spec?.nodeName || '-'}</div>
                    <div className="stat-label">Node</div>
                  </div>
                </>
              ) : (
                /* Workload 资源的状态 */
                <>
                  <div className="stat-card">
                    <div className="stat-value">
                      {status.readyReplicas !== undefined ? (
                        <StatusBadge
                          status={
                            status.readyReplicas === status.replicas
                              ? 'Available'
                              : status.readyReplicas > 0
                                ? 'Partial'
                                : 'Unavailable'
                          }
                          resourceType={resourceType}
                        />
                      ) : status.phase ? (
                        <StatusBadge status={status.phase} resourceType={resourceType} />
                      ) : (
                        '-'
                      )}
                    </div>
                    <div className="stat-label">Status</div>
                  </div>
                  {status.replicas !== undefined && (
                    <div className="stat-card">
                      <div className="stat-value">{String(status.replicas)}</div>
                      <div className="stat-label">Replicas</div>
                    </div>
                  )}
                  {status.readyReplicas !== undefined && (
                    <div className="stat-card">
                      <div className="stat-value">{String(status.readyReplicas)}</div>
                      <div className="stat-label">Ready</div>
                    </div>
                  )}
                  {status.availableReplicas !== undefined && (
                    <div className="stat-card">
                      <div className="stat-value">{String(status.availableReplicas)}</div>
                      <div className="stat-label">Available</div>
                    </div>
                  )}
                </>
              )}
            </div>
          </div>
        </div>
      )}

      {/* RESOURCE INFORMATION */}
      <div className="detail-card">
        <h3 className="detail-card-title">{resourceInfo.title} Information</h3>
        <div className="detail-card-body">
          {/* 通用字段 */}
          <div className="info-grid">
            <div className="info-item">
              <span className="info-label">Created</span>
              <span className="info-value">
                {metadata.creationTimestamp
                  ? `${new Date(metadata.creationTimestamp).toLocaleString()} (${formatRelativeTime(metadata.creationTimestamp)})`
                  : '-'}
              </span>
            </div>
            {metadata.namespace && (
              <div className="info-item">
                <span className="info-label">Namespace</span>
                <span className="info-value">{metadata.namespace}</span>
              </div>
            )}
            <div className="info-item">
              <span className="info-label">UID</span>
              <span className="info-value">{metadata.uid || '-'}</span>
            </div>
            <div className="info-item">
              <span className="info-label">Resource Version</span>
              <span className="info-value">{metadata.resourceVersion || '-'}</span>
            </div>

            {metadata.ownerReferences && metadata.ownerReferences.length > 0 && (
              <div className="info-item">
                <span className="info-label">Owner</span>
                <span className="info-value clickable">
                  {metadata.ownerReferences[0].kind}/{metadata.ownerReferences[0].name}
                </span>
              </div>
            )}

            {/* 资源特有字段 */}
            {RESOURCE_FIELDS[resourceType]?.map(field => {
              // 检查条件
              if (field.condition && !field.condition(data)) return null;

              // 获取值
              const value = field.getValue(data);
              if (value === null || value === undefined || value === '') return null;

              // 渲染
              return (
                <div key={field.key} className="info-item">
                  <span className="info-label">{field.label}</span>
                  <span className="info-value">
                    {field.render ? field.render(value, data) : value}
                  </span>
                </div>
              );
            })}
          </div>

          {/* Labels 和 Annotations - 同一行平分宽度，最多显示 5 行 */}
          {(metadata.labels && Object.keys(metadata.labels).length > 0) ||
          (metadata.annotations && Object.keys(metadata.annotations).length > 0) ? (
            <div className="info-section">
              <div className="info-grid info-grid-2col">
                {/* Labels - 最多显示 5 个 */}
                {metadata.labels && Object.keys(metadata.labels).length > 0 && (
                  <div className="info-item">
                    <span className="info-label">Labels</span>
                    <div className="label-list">
                      {Object.entries(metadata.labels)
                        .slice(0, 5)
                        .map(([key, value]) => {
                          const fullText = `${key}: ${value}`;
                          const displayKey = truncateText(key as string, 30);
                          const displayValue = truncateText(value as string, 30);

                          // 只有当文本被截断时才显示 tooltip
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
                              <Tippy
                                key={key}
                                content={fullText}
                                theme="light"
                                placement="top"
                                arrow={true}
                                duration={200}
                              >
                                {labelElement}
                              </Tippy>
                            );
                          }

                          return <span key={key}>{labelElement}</span>;
                        })}
                      {/* 显示更多提示 */}
                      {Object.keys(metadata.labels).length > 5 && (
                        <Tippy
                          content={
                            <div style={{ maxHeight: '200px', overflow: 'auto' }}>
                              {Object.entries(metadata.labels).map(([key, value]) => (
                                <div key={key}>
                                  {key}: {value}
                                </div>
                              ))}
                            </div>
                          }
                          theme="light"
                          placement="top"
                          arrow={true}
                          duration={200}
                          interactive={true}
                        >
                          <span className="label-tag label-more">
                            +{Object.keys(metadata.labels).length - 5} more
                          </span>
                        </Tippy>
                      )}
                    </div>
                  </div>
                )}

                {/* Annotations - 最多显示 5 个 */}
                {metadata.annotations && Object.keys(metadata.annotations).length > 0 && (
                  <div className="info-item">
                    <span className="info-label">Annotations</span>
                    <div className="annotation-list">
                      {Object.entries(metadata.annotations)
                        .slice(0, 5)
                        .map(([key, value]) => {
                          const fullText = `${key}: ${value}`;
                          const displayKey = truncateText(key as string, 30);
                          const displayValue = truncateText(value as string, 30);

                          // 只有当文本被截断时才显示 tooltip
                          const isTruncated = displayKey !== key || displayValue !== value;

                          const labelElement = (
                            <span className="annotation-tag">
                              <span className="annotation-key">{displayKey}</span>
                              <span className="annotation-separator">: </span>
                              <span className="annotation-value">{displayValue}</span>
                            </span>
                          );

                          if (isTruncated) {
                            return (
                              <Tippy
                                key={key}
                                content={fullText}
                                theme="light"
                                placement="top"
                                arrow={true}
                                duration={200}
                              >
                                {labelElement}
                              </Tippy>
                            );
                          }

                          return <span key={key}>{labelElement}</span>;
                        })}
                      {/* 显示更多提示 */}
                      {Object.keys(metadata.annotations).length > 5 && (
                        <Tippy
                          content={
                            <div style={{ maxHeight: '200px', overflow: 'auto' }}>
                              {Object.entries(metadata.annotations).map(([key, value]) => (
                                <div key={key}>
                                  {key}: {value}
                                </div>
                              ))}
                            </div>
                          }
                          theme="light"
                          placement="top"
                          arrow={true}
                          duration={200}
                          interactive={true}
                        >
                          <span className="annotation-tag label-more">
                            +{Object.keys(metadata.annotations).length - 5} more
                          </span>
                        </Tippy>
                      )}
                    </div>
                  </div>
                )}
              </div>
            </div>
          ) : null}
        </div>
      </div>

      {/* POD: CONTAINERS */}
      {containerStatuses.length > 0 && (
        <div className="detail-card">
          <h3 className="detail-card-title">Containers</h3>
          <div className="detail-card-body">
            {containerStatuses.map((container: any) => {
              const isExpanded = containersExpanded[container.name] || false;
              const containerSpec = spec?.containers?.find((c: any) => c.name === container.name);

              return (
                <div key={container.name} className="container-card">
                  <div
                    className="container-card-header"
                    onClick={() =>
                      setContainersExpanded(prev => ({
                        ...prev,
                        [container.name]: !isExpanded,
                      }))
                    }
                  >
                    <div>
                      <span className="collapse-btn">{isExpanded ? '▼' : '▶'}</span>
                      <span className="container-card-title">{container.name}</span>
                      <span className="container-card-image">{container.image}</span>
                    </div>
                    {containerSpec?.imagePullPolicy && (
                      <span className="container-card-pull-policy">
                        {containerSpec.imagePullPolicy}
                      </span>
                    )}
                  </div>

                  {isExpanded && renderContainerDetails(containerSpec)}
                </div>
              );
            })}
          </div>
        </div>
      )}

      {/* WORKLOAD: CONTAINERS (Deployment, StatefulSet, DaemonSet, Job) */}
      {resourceInfo.hasContainers && resourceType !== 'pod' && workloadContainers.length > 0 && (
        <div className="detail-card">
          <h3 className="detail-card-title">Containers</h3>
          <div className="detail-card-body">
            {workloadContainers.map((container: any, index: number) => {
              const isExpanded = containersExpanded[container.name] || false;

              return (
                <div key={index} className="container-card">
                  <div
                    className="container-card-header"
                    onClick={() =>
                      setContainersExpanded(prev => ({
                        ...prev,
                        [container.name]: !isExpanded,
                      }))
                    }
                  >
                    <div>
                      <span className="collapse-btn">{isExpanded ? '▼' : '▶'}</span>
                      <span className="container-card-title">{container.name}</span>
                      <span className="container-card-image">{container.image}</span>
                    </div>
                    {container.imagePullPolicy && (
                      <span className="container-card-pull-policy">
                        {container.imagePullPolicy}
                      </span>
                    )}
                  </div>

                  {isExpanded && renderContainerDetails(container)}
                </div>
              );
            })}
          </div>
        </div>
      )}

      {/* CONDITIONS */}
      {conditions.length > 0 && (
        <div className="detail-card">
          <h3 className="detail-card-title">Conditions</h3>
          <div className="detail-card-body">
            <div className="condition-list">
              {conditions.map((condition: any) => (
                <div key={condition.type} className="condition-item">
                  <span className="condition-type">{condition.type}</span>
                  <span
                    className={`condition-status status-badge ${
                      condition.status === 'True'
                        ? 'success'
                        : condition.status === 'False'
                          ? 'error'
                          : 'warning'
                    }`}
                  >
                    {condition.status === 'True' ? '✓' : condition.status === 'False' ? '✗' : '?'}{' '}
                    {condition.status}
                  </span>
                  {condition.lastTransitionTime && (
                    <span className="condition-time">
                      {formatRelativeTime(condition.lastTransitionTime)}
                    </span>
                  )}
                </div>
              ))}
            </div>
          </div>
        </div>
      )}
    </div>
  );
};

export default OverviewTab;
