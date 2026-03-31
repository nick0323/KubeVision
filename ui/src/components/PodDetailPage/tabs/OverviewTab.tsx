import React, { useState, useMemo } from 'react';
import { Pod } from '../../types/k8s-resources';
import { StatusBadge } from '../../StatusBadge';
import { CollapsibleSection } from '../../CollapsibleSection';
import { OverviewTabProps, ContainerStatusSummary, PodCondition } from '../types';
import './OverviewTab.css';

/**
 * Overview Tab - Pod 概览
 */
export const OverviewTab: React.FC<OverviewTabProps> = ({ pod, loading, onRefresh }) => {
  const [containersExpanded, setContainersExpanded] = useState<Record<string, boolean>>({});

  /**
   * 容器状态摘要
   */
  const containerStatuses = useMemo<ContainerStatusSummary[]>(() => {
    if (!pod?.status?.containerStatuses) return [];
    return pod.status.containerStatuses.map((cs) => ({
      name: cs.name,
      ready: cs.ready,
      restartCount: cs.restartCount,
      state: cs.state || {},
      image: cs.image,
    }));
  }, [pod?.status?.containerStatuses]);

  /**
   * 计算 Ready Containers
   * 如果容器状态为空，显示 0 / 期望容器数
   */
  const readyContainers = useMemo(() => {
    const ready = containerStatuses.filter((c) => c.ready).length;
    const total = containerStatuses.length || pod?.spec?.containers?.length || 0;
    return { ready, total };
  }, [containerStatuses, pod?.spec?.containers?.length]);

  /**
   * 计算 Restart Count
   */
  const restartCount = useMemo(() => {
    return containerStatuses.reduce((sum, c) => sum + c.restartCount, 0);
  }, [containerStatuses]);

  /**
   * 判断 Pod 是否正在初始化
   */
  const isInitializing = useMemo(() => {
    // 如果容器状态为空但 spec 中有容器定义，说明正在创建
    return (!pod?.status?.containerStatuses || pod.status.containerStatuses.length === 0) &&
           pod?.spec?.containers && pod.spec.containers.length > 0;
  }, [pod?.status?.containerStatuses, pod?.spec?.containers]);

  /**
   * 获取 Pod 显示状态
   */
  const getPodStatus = () => {
    if (isInitializing) {
      return 'Pending';
    }
    return pod.status?.phase || 'Unknown';
  };

  /**
   * Pod 条件
   */
  const conditions = useMemo<PodCondition[]>(() => {
    return pod?.status?.conditions || [];
  }, [pod?.status?.conditions]);

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

  if (loading || !pod) {
    return <div className="overview-tab-loading">加载中...</div>;
  }

  return (
    <div className="overview-tab">
      {/* STATUS OVERVIEW */}
      <div className="detail-card">
        <h3 className="detail-card-title">Status Overview</h3>
        <div className="detail-card-body">
          <div className="stats-grid">
            <div className="stat-card">
              <div className="stat-value">
                <StatusBadge status={getPodStatus()} resourceType="pods" />
              </div>
              <div className="stat-label">Phase</div>
            </div>
            <div className="stat-card">
              <div className="stat-value">
                {readyContainers.ready} / {readyContainers.total}
              </div>
              <div className="stat-label">Ready Containers</div>
            </div>
            <div className="stat-card">
              <div className="stat-value">
                {restartCount}
              </div>
              <div className="stat-label">Restart Count</div>
            </div>
            <div className="stat-card">
              <div className="stat-value">{pod.spec?.nodeName || '-'}</div>
              <div className="stat-label">Node</div>
            </div>
          </div>
        </div>
      </div>

      {/* POD INFORMATION */}
      <div className="detail-card">
        <h3 className="detail-card-title">Pod Information</h3>
        <div className="detail-card-body">
          <div className="info-grid">
            <div className="info-item">
              <span className="info-label">Created</span>
              <span className="info-value">
                {pod.metadata?.creationTimestamp
                  ? `${new Date(pod.metadata.creationTimestamp).toLocaleString()} (${formatRelativeTime(pod.metadata.creationTimestamp)})`
                  : '-'}
              </span>
            </div>
            <div className="info-item">
              <span className="info-label">Started</span>
              <span className="info-value">
                {containerStatuses[0]?.state?.running?.startedAt
                  ? `${new Date(containerStatuses[0].state.running.startedAt).toLocaleString()} (${formatRelativeTime(containerStatuses[0].state.running.startedAt)})`
                  : '-'}
              </span>
            </div>
            <div className="info-item">
              <span className="info-label">Pod IP</span>
              <span className="info-value">{pod.status?.podIP || '-'}</span>
            </div>
            <div className="info-item">
              <span className="info-label">Host IP</span>
              <span className="info-value">{pod.status?.hostIP || '-'}</span>
            </div>
            {pod.metadata?.ownerReferences && pod.metadata.ownerReferences.length > 0 && (
              <div className="info-item">
                <span className="info-label">Owner</span>
                <span className="info-value clickable">
                  {pod.metadata.ownerReferences[0].kind}/{pod.metadata.ownerReferences[0].name}
                </span>
              </div>
            )}
            {pod.spec?.containers?.[0]?.ports?.[0] && (
              <div className="info-item">
                <span className="info-label">Ports</span>
                <span className="info-value">
                  {pod.spec.containers[0].ports.map((p) => `${p.containerPort}/${p.protocol || 'TCP'}`).join(', ')}
                </span>
              </div>
            )}
            {pod.metadata?.labels && Object.keys(pod.metadata.labels).length > 0 && (
              <div className="info-item">
                <span className="info-label">Labels</span>
                <div className="label-list">
                  {Object.entries(pod.metadata.labels).map(([key, value]) => (
                    <span key={key} className="label-item">
                      <span className="label-key">{key}</span>
                      <span className="label-separator">: </span>
                      <span className="label-value">{value}</span>
                    </span>
                  ))}
                </div>
              </div>
            )}
            {pod.metadata?.annotations && Object.keys(pod.metadata.annotations).length > 0 && (
              <div className="info-item">
                <span className="info-label">Annotations</span>
                <div className="label-list">
                  {Object.entries(pod.metadata.annotations).map(([key, value]) => (
                    <span key={key} className="label-item">
                      <span className="label-key">{key}</span>
                      <span className="label-separator">: </span>
                      <span className="label-value">{value}</span>
                    </span>
                  ))}
                </div>
              </div>
            )}
          </div>
        </div>
      </div>

      {/* CONTAINERS */}
      {containerStatuses.length > 0 && (
        <div className="detail-card">
          <h3 className="detail-card-title">Containers</h3>
          <div className="detail-card-body">
            {containerStatuses.map((container) => {
              const isExpanded = containersExpanded[container.name] || false;
              const containerSpec = pod.spec?.containers?.find(c => c.name === container.name);

              return (
                <div key={container.name} className="container-card">
                  <div
                    className="container-card-header"
                    onClick={() => setContainersExpanded(prev => ({
                      ...prev,
                      [container.name]: !isExpanded,
                    }))}
                  >
                    <div>
                      <span className="collapse-btn">{isExpanded ? '▼' : '▶'}</span>
                      <span className="container-card-title">{container.name}</span>
                      <span className="container-card-image">{container.image}</span>
                    </div>
                    {containerSpec?.imagePullPolicy && (
                      <span className="container-card-pull-policy">{containerSpec.imagePullPolicy}</span>
                    )}
                  </div>
                  
                  {isExpanded && (
                    <div className="container-card-body">
                      {/* PORTS */}
                      <div className="sub-module">
                        <div className="sub-module-title">Ports</div>
                        <div className="info-grid">
                          {pod.spec?.containers?.find(c => c.name === container.name)?.ports?.map((port, idx) => (
                            <div key={idx} className="info-item">
                              <span className="info-value">
                                {port.name ? `${port.name}: ` : ''}{port.containerPort}/{port.protocol || 'TCP'}
                              </span>
                            </div>
                          )) || <div className="empty-state"><span className="empty-state-text">No ports defined</span></div>}
                        </div>
                      </div>

                      {/* ENVIRONMENT VARIABLES */}
                      <div className="sub-module">
                        <div className="sub-module-title">Environment Variables</div>
                        <div className="info-grid">
                          {pod.spec?.containers?.find(c => c.name === container.name)?.env?.map((env, idx) => (
                            <div key={idx} className="info-item">
                              <span className="info-value">
                                <span className="env-key">{env.name}</span>
                                <span className="env-separator">: </span>
                                <span className="env-value">{env.value || (env.valueFrom ? '(From ConfigMap/Secret)' : '-')}</span>
                              </span>
                            </div>
                          )) || <div className="empty-state"><span className="empty-state-text">No environment variables</span></div>}
                        </div>
                      </div>

                      {/* RESOURCES */}
                      <div className="sub-module">
                        <div className="sub-module-title">Resources</div>
                        <div className="info-grid">
                          <div className="info-item">
                            <span className="info-label">Requests (CPU)</span>
                            <span className="info-value">
                              {pod.spec.containers.find(c => c.name === container.name)?.resources?.requests?.cpu || '-'}
                            </span>
                          </div>
                          <div className="info-item">
                            <span className="info-label">Requests (Memory)</span>
                            <span className="info-value">
                              {pod.spec.containers.find(c => c.name === container.name)?.resources?.requests?.memory || '-'}
                            </span>
                          </div>
                          <div className="info-item">
                            <span className="info-label">Limits (CPU)</span>
                            <span className="info-value">
                              {pod.spec.containers.find(c => c.name === container.name)?.resources?.limits?.cpu || '-'}
                            </span>
                          </div>
                          <div className="info-item">
                            <span className="info-label">Limits (Memory)</span>
                            <span className="info-value">
                              {pod.spec.containers.find(c => c.name === container.name)?.resources?.limits?.memory || '-'}
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
                              {pod.spec?.containers?.find(c => c.name === container.name)?.livenessProbe ? 'Configured' : 'Not Configured'}
                            </span>
                          </div>
                          <div className="info-item">
                            <span className="info-label">Readiness Probe</span>
                            <span className="info-value">
                              {pod.spec?.containers?.find(c => c.name === container.name)?.readinessProbe ? 'Configured' : 'Not Configured'}
                            </span>
                          </div>
                        </div>
                      </div>

                      {/* VOLUMES */}
                      <div className="sub-module">
                        <div className="sub-module-title">Volumes</div>
                        <div className="info-grid">
                          {pod.spec?.containers?.find(c => c.name === container.name)?.volumeMounts?.map((mount, idx) => (
                            <div key={idx} className="info-item">
                              <span className="info-label">{mount.name}</span>
                              <span className="info-value">
                                {mount.mountPath} {mount.readOnly ? '(RO)' : '(RW)'}
                              </span>
                            </div>
                          )) || <div className="empty-state"><span className="empty-state-text">No volumes</span></div>}
                        </div>
                      </div>
                    </div>
                  )}
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
              {conditions.map((condition) => (
                <div key={condition.type} className="condition-item">
                  <span className="condition-type">{condition.type}</span>
                  <span className={`condition-status status-badge ${condition.status === 'True' ? 'success' : condition.status === 'False' ? 'error' : 'warning'}`}>
                    {condition.status === 'True' ? '✓' : condition.status === 'False' ? '✗' : '?'} {condition.status}
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
