import React, { useState, useMemo } from 'react';
import { OverviewTabProps } from '../types';
import { StatusBadge } from '../../StatusBadge';
import './OverviewTab.css';

/**
 * 资源类型配置
 */
const RESOURCE_CONFIG: Record<string, { title: string; hasContainers?: boolean; hasPorts?: boolean }> = {
  pod: { title: 'Pod', hasContainers: true },
  deployment: { title: 'Deployment', hasContainers: true },
  statefulset: { title: 'StatefulSet', hasContainers: true },
  daemonset: { title: 'DaemonSet', hasContainers: true },
  service: { title: 'Service', hasPorts: true },
  configmap: { title: 'ConfigMap' },
  secret: { title: 'Secret' },
  ingress: { title: 'Ingress' },
  job: { title: 'Job', hasContainers: true },
  cronjob: { title: 'CronJob' }, // CronJob 不显示 Containers
  pvc: { title: 'PersistentVolumeClaim' },
  pv: { title: 'PersistentVolume' },
  storageclass: { title: 'StorageClass' },
  namespace: { title: 'Namespace' },
  node: { title: 'Node' },
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
export const OverviewTab: React.FC<OverviewTabProps> = ({ data, loading, resourceType = 'pod' }) => {
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

  // 渲染标签列表（限制显示数量和长度）
  const renderLabels = (labels: Record<string, string>, maxItems = 5) => {
    if (!labels || Object.keys(labels).length === 0) return null;
    
    const entries = Object.entries(labels);
    const visibleEntries = entries.slice(0, maxItems);
    const hiddenCount = entries.length - maxItems;
    
    return (
      <div className="info-item labels-item">
        <span className="info-label">Labels</span>
        <div className="label-list">
          {visibleEntries.map(([key, value]) => (
            <span key={key} className="label-tag" title={`${key}: ${value}`}>
              <span className="label-key">{truncateText(key, 30)}</span>
              <span className="label-separator">: </span>
              <span className="label-value">{truncateText(value, 30)}</span>
            </span>
          ))}
          {hiddenCount > 0 && (
            <span className="label-tag label-more">
              +{hiddenCount}
            </span>
          )}
        </div>
      </div>
    );
  };

  // 渲染注解列表（限制显示数量和长度，使用标签样式）
  const renderAnnotations = (annotations: Record<string, string>, maxItems = 5) => {
    if (!annotations || Object.keys(annotations).length === 0) return null;
    
    const entries = Object.entries(annotations);
    const visibleEntries = entries.slice(0, maxItems);
    const hiddenCount = entries.length - maxItems;
    
    return (
      <div className="info-item annotations-item">
        <span className="info-label">Annotations</span>
        <div className="annotation-list">
          {visibleEntries.map(([key, value]) => (
            <span key={key} className="annotation-tag" title={`${key}: ${value}`}>
              <span className="annotation-key">{truncateText(key, 30)}</span>
              <span className="annotation-separator">: </span>
              <span className="annotation-value">{truncateText(value, 30)}</span>
            </span>
          ))}
          {hiddenCount > 0 && (
            <span className="annotation-tag label-more">
              +{hiddenCount}
            </span>
          )}
        </div>
      </div>
    );
  };

  // 截断文本
  const truncateText = (text: string, maxLength: number) => {
    if (!text) return '';
    if (text.length <= maxLength) return text;
    return text.substring(0, maxLength) + '...';
  };

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
              <span className="info-value">
                {container?.resources?.requests?.cpu || '-'}
              </span>
            </div>
            <div className="info-item">
              <span className="info-label">Requests (Memory)</span>
              <span className="info-value">
                {container?.resources?.requests?.memory || '-'}
              </span>
            </div>
            <div className="info-item">
              <span className="info-label">Limits (CPU)</span>
              <span className="info-value">
                {container?.resources?.limits?.cpu || '-'}
              </span>
            </div>
            <div className="info-item">
              <span className="info-label">Limits (Memory)</span>
              <span className="info-value">
                {container?.resources?.limits?.memory || '-'}
              </span>
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
    // Pod 和 Workload 资源显示状态
    if (resourceType === 'pod') return true;
    if (['deployment', 'statefulset', 'daemonset', 'job'].includes(resourceType)) return true;
    // 其他资源如果有 status.phase 或 status.replicas 也显示
    if (status.phase || status.replicas !== undefined) return true;
    return false;
  }, [resourceType, status]);

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
                            ? 'Ready'
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

            {renderLabels(metadata.labels)}
            {renderAnnotations(metadata.annotations)}
          </div>
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
      {resourceInfo.hasContainers && workloadContainers.length > 0 && (
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
                    {condition.status === 'True'
                      ? '✓'
                      : condition.status === 'False'
                        ? '✗'
                        : '?'}{' '}
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
