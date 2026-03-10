/**
 * Deployment 详情页面组件
 * 专为 Deployment 设计，展示副本状态、升级策略、关联 RS 和 Pod 等信息
 */
import React from 'react';
import { StatusCards } from '../StatusCards';
import { LabelList } from '../LabelList';
import { EventTimeline } from '../EventTimeline';
import './ResourceDetail.css';

interface DeploymentDetailProps {
  data: any;
  events?: any[];
  relatedPods?: any[];
  onResourceClick?: (type: string, namespace: string, name: string) => void;
}

export const DeploymentDetail: React.FC<DeploymentDetailProps> = ({
  data,
  events = [],
  relatedPods = [],
  onResourceClick,
}) => {
  if (!data) return null;

  const { metadata, spec, status } = data;
  const conditions = status?.conditions || [];

  // 计算状态
  const available = conditions.find((c: any) => c.type === 'Available');
  const progressing = conditions.find((c: any) => c.type === 'Progressing');
  const replicaFailure = conditions.find((c: any) => c.type === 'ReplicaFailure');

  const isAvailable = available?.status === 'True';
  const isProgressing = progressing?.status === 'True';
  const hasFailure = replicaFailure?.status === 'True';

  const statusValue = isAvailable ? 'Available' : 'Unavailable';
  const statusSub = isProgressing
    ? 'Progressing'
    : hasFailure
    ? 'Failure'
    : '';

  // 副本信息
  const replicas = spec.replicas || 1;
  const readyReplicas = status?.readyReplicas || 0;
  const updatedReplicas = status?.updatedReplicas || 0;
  const availableReplicas = status?.availableReplicas || 0;
  const unavailableReplicas = status?.unavailableReplicas || 0;

  // 状态卡片
  const statusCards = [
    {
      title: '状态',
      value: statusValue,
      sub: statusSub,
    },
    {
      title: '副本',
      value: `${readyReplicas}/${replicas}`,
      sub: `更新：${updatedReplicas} | 可用：${availableReplicas}`,
    },
    {
      title: '不可用',
      value: unavailableReplicas,
      sub: hasFailure ? '⚠️ 副本创建失败' : '-',
    },
  ];

  // 升级策略
  const strategy = spec.strategy?.type || 'RollingUpdate';
  const rollingUpdate = spec.strategy?.rollingUpdate || {};

  // 容器信息
  const containers = spec.template?.spec?.containers || [];
  const initContainers = spec.template?.spec?.initContainers || [];

  // 渲染条件状态
  const renderConditions = () => {
    return (
      <div className="conditions-list">
        {conditions.map((condition: any, idx: number) => (
          <div
            key={idx}
            className={`condition-item ${
              condition.status === 'True' ? 'status-true' : 'status-false'
            }`}
          >
            <span className="condition-type">{condition.type}</span>
            <span className="condition-status">
              {condition.status === 'True' ? '✅' : '❌'}
            </span>
            <span className="condition-reason">
              {condition.reason || '-'}
            </span>
            <span className="condition-message" title={condition.message}>
              {condition.message || '-'}
            </span>
          </div>
        ))}
      </div>
    );
  };

  // 渲染 Pod 列表
  const renderPodList = () => {
    if (relatedPods.length === 0) {
      return (
        <div className="empty-state">
          <span className="empty-icon">📭</span>
          <p>暂无关联 Pod</p>
        </div>
      );
    }

    return (
      <div className="pod-table">
        <table className="resource-table">
          <thead>
            <tr>
              <th>名称</th>
              <th>状态</th>
              <th>就绪</th>
              <th>重启</th>
              <th>节点</th>
              <th>IP</th>
              <th>年龄</th>
            </tr>
          </thead>
          <tbody>
            {relatedPods.map((pod: any) => {
              const containerStatuses = pod.status?.containerStatuses || [];
              const readyCount = containerStatuses.filter(
                (c: any) => c.ready
              ).length;
              const restartCount = containerStatuses.reduce(
                (sum: number, c: any) => sum + (c.restartCount || 0),
                0
              );

              return (
                <tr key={pod.metadata.uid}>
                  <td
                    className="link"
                    onClick={() =>
                      onResourceClick?.('pod', pod.metadata.namespace, pod.metadata.name)
                    }
                  >
                    {pod.metadata.name}
                  </td>
                  <td>
                    <span className={`status-badge status-${pod.status?.phase?.toLowerCase()}`}>
                      {pod.status?.phase || 'Unknown'}
                    </span>
                  </td>
                  <td>
                    {readyCount}/{containerStatuses.length}
                  </td>
                  <td>{restartCount}</td>
                  <td>{pod.spec.nodeName || '-'}</td>
                  <td>{pod.status.podIP || '-'}</td>
                  <td>
                    {pod.metadata.creationTimestamp
                      ? new Date(
                          pod.metadata.creationTimestamp
                        ).toLocaleDateString()
                      : '-'}
                  </td>
                </tr>
              );
            })}
          </tbody>
        </table>
      </div>
    );
  };

  return (
    <div className="deployment-detail">
      {/* 状态概览 */}
      <div className="detail-section">
        <h3 className="section-title">📊 状态概览</h3>
        <StatusCards cards={statusCards} />

        {/* 条件状态 */}
        <div className="info-card">
          <h4 className="card-title">📋 条件状态</h4>
          {renderConditions()}
        </div>
      </div>

      {/* 副本信息 */}
      <div className="detail-section">
        <h3 className="section-title">🔢 副本详情</h3>
        <div className="replicas-card info-card">
          <div className="replicas-visual">
            <div className="replica-bar">
              {Array.from({ length: replicas }).map((_, idx) => {
                const isReady = idx < readyReplicas;
                const isUpdated = idx < updatedReplicas;
                return (
                  <div
                    key={idx}
                    className={`replica-block ${
                      isReady ? 'ready' : 'not-ready'
                    } ${isUpdated ? 'updated' : 'old'}`}
                    title={`副本 ${idx + 1}: ${
                      isReady ? '就绪' : '未就绪'
                    }, ${isUpdated ? '已更新' : '旧版本'}`}
                  >
                    {idx + 1}
                  </div>
                );
              })}
            </div>
            <div className="replica-legend">
              <span className="legend-item">
                <span className="legend-color ready-updated"></span>
                已更新且就绪
              </span>
              <span className="legend-item">
                <span className="legend-color ready-old"></span>
                旧版本就绪
              </span>
              <span className="legend-item">
                <span className="legend-color not-ready"></span>
                未就绪
              </span>
            </div>
          </div>

          <div className="replicas-stats">
            <div className="stat-item">
              <span className="stat-label">期望副本</span>
              <span className="stat-value">{replicas}</span>
            </div>
            <div className="stat-item">
              <span className="stat-label">就绪副本</span>
              <span className="stat-value">{readyReplicas}</span>
            </div>
            <div className="stat-item">
              <span className="stat-label">更新副本</span>
              <span className="stat-value">{updatedReplicas}</span>
            </div>
            <div className="stat-item">
              <span className="stat-label">可用副本</span>
              <span className="stat-value">{availableReplicas}</span>
            </div>
            <div className="stat-item">
              <span className="stat-label">不可用</span>
              <span className="stat-value error">{unavailableReplicas}</span>
            </div>
          </div>
        </div>
      </div>

      {/* 部署策略 */}
      <div className="detail-section">
        <h3 className="section-title">🔄 部署策略</h3>
        <div className="strategy-card info-card">
          <div className="strategy-header">
            <span className="strategy-type">{strategy}</span>
          </div>

          {strategy === 'RollingUpdate' && rollingUpdate && (
            <div className="rolling-update-config">
              <div className="config-item">
                <span className="config-label">最大激增</span>
                <span className="config-value">{rollingUpdate.maxSurge || '25%'}</span>
                <span className="config-desc">
                  升级过程中允许超出期望副本数的最大数量
                </span>
              </div>
              <div className="config-item">
                <span className="config-label">最大不可用</span>
                <span className="config-value">
                  {rollingUpdate.maxUnavailable || '25%'}
                </span>
                <span className="config-desc">
                  升级过程中允许的最大不可用副本数
                </span>
              </div>
            </div>
          )}

          {strategy === 'Recreate' && (
            <div className="recreate-notice">
              <span className="notice-icon">⚠️</span>
              <span className="notice-text">
                重建策略：先删除所有旧 Pod，再创建新 Pod（会有服务中断）
              </span>
            </div>
          )}
        </div>
      </div>

      {/* 基本信息 */}
      <div className="detail-section">
        <h3 className="section-title">📋 基本信息</h3>
        <div className="info-card">
          <div className="info-grid">
            <div className="info-item">
              <span className="info-label">名称</span>
              <span className="info-value">{metadata.name}</span>
            </div>
            <div className="info-item">
              <span className="info-label">命名空间</span>
              <span className="info-value">{metadata.namespace}</span>
            </div>
            <div className="info-item">
              <span className="info-label">创建时间</span>
              <span className="info-value">
                {metadata.creationTimestamp
                  ? new Date(metadata.creationTimestamp).toLocaleString()
                  : '-'}
              </span>
            </div>
            <div className="info-item">
              <span className="info-label">UID</span>
              <span className="info-value">{metadata.uid}</span>
            </div>
            <div className="info-item full-width">
              <span className="info-label">选择器</span>
              <span className="info-value">
                {spec.selector?.matchLabels
                  ? Object.entries(spec.selector.matchLabels)
                      .map(([k, v]) => `${k}=${v}`)
                      .join(', ')
                  : '-'}
              </span>
            </div>
          </div>
        </div>
      </div>

      {/* 容器镜像 */}
      <div className="detail-section">
        <h3 className="section-title">📦 容器镜像</h3>
        <div className="containers-list">
          {containers.map((container: any, idx: number) => (
            <div key={idx} className="container-card">
              <h4 className="card-title">{container.name}</h4>
              <div className="container-info">
                <div className="info-row">
                  <span className="info-label">镜像</span>
                  <span className="info-value code">{container.image}</span>
                </div>
                {container.ports && container.ports.length > 0 && (
                  <div className="info-row">
                    <span className="info-label">端口</span>
                    <span className="info-value">
                      {container.ports
                        .map(
                          (p: any) =>
                            `${p.containerPort}/${p.protocol || 'TCP'}`
                        )
                        .join(', ')}
                    </span>
                  </div>
                )}
                {container.resources && (
                  <div className="info-row">
                    <span className="info-label">资源</span>
                    <span className="info-value">
                      请求：
                      {container.resources.requests?.cpu || '-'} CPU,{' '}
                      {container.resources.requests?.memory || '-'} Mem | 限制：
                      {container.resources.limits?.cpu || '-'} CPU,{' '}
                      {container.resources.limits?.memory || '-'} Mem
                    </span>
                  </div>
                )}
              </div>
            </div>
          ))}
        </div>
      </div>

      {/* 标签和注解 */}
      <div className="detail-section">
        <h3 className="section-title">🏷️ 标签和注解</h3>

        {/* Pod 模板标签 */}
        {spec.template?.metadata?.labels && (
          <div className="labels-subsection">
            <h4 className="subsection-title">Pod 模板标签</h4>
            <LabelList labels={spec.template.metadata.labels} />
          </div>
        )}

        {/* Deployment 标签 */}
        {metadata?.labels && Object.keys(metadata.labels).length > 0 && (
          <div className="labels-subsection">
            <h4 className="subsection-title">Deployment 标签</h4>
            <LabelList labels={metadata.labels} />
          </div>
        )}

        {/* Deployment 注解 */}
        {metadata?.annotations &&
          Object.keys(metadata.annotations).length > 0 && (
            <div className="labels-subsection">
              <h4 className="subsection-title">注解</h4>
              <LabelList labels={metadata.annotations} />
            </div>
          )}
      </div>

      {/* 关联 Pod 列表 */}
      <div className="detail-section">
        <h3 className="section-title">🔗 关联 Pod</h3>
        {renderPodList()}
      </div>

      {/* 事件 */}
      {events.length > 0 && (
        <div className="detail-section">
          <h3 className="section-title">📅 事件</h3>
          <EventTimeline events={events} />
        </div>
      )}
    </div>
  );
};

export default DeploymentDetail;
