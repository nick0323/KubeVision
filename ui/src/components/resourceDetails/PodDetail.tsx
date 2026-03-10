/**
 * Pod 详情页面组件
 * 专为 Pod 资源设计，展示容器、探针、环境变量、日志等信息
 */
import React from 'react';
import { StatusCards } from '../StatusCards';
import { ResourceBar } from '../ResourceBar';
import { ContainerList } from '../ContainerList';
import { LabelList } from '../LabelList';
import { EventTimeline } from '../EventTimeline';
import './ResourceDetail.css';

interface PodDetailProps {
  data: any;
  events?: any[];
  onResourceClick?: (type: string, namespace: string, name: string) => void;
  onViewLog?: (containerName: string) => void;
  onExec?: (containerName: string) => void;
}

export const PodDetail: React.FC<PodDetailProps> = ({
  data,
  events = [],
  onResourceClick,
  onViewLog,
  onExec,
}) => {
  if (!data) return null;

  const { metadata, spec, status } = data;
  const containers = spec?.containers || [];
  const initContainers = spec?.initContainers || [];
  const conditions = status?.conditions || [];

  // Pod 状态卡片
  const phase = status?.phase || 'Unknown';
  const readyCondition = conditions.find((c: any) => c.type === 'Ready');
  const podIP = status?.podIP;
  const hostIP = status?.hostIP;

  // 计算容器状态
  const containerStatuses = status?.containerStatuses || [];
  const readyCount = containerStatuses.filter((c: any) => c.ready).length;
  const totalContainers = containers.length;
  const restartCount = containerStatuses.reduce(
    (sum: number, c: any) => sum + (c.restartCount || 0),
    0
  );

  const statusCards = [
    {
      title: '状态',
      value: phase,
      sub: readyCondition?.status === 'True' ? 'Ready' : 'Not Ready',
    },
    {
      title: '容器',
      value: `${readyCount}/${totalContainers}`,
      sub: `重启：${restartCount} 次`,
    },
    {
      title: 'QoS 等级',
      value: data.status?.qosClass || 'BestEffort',
      sub: spec?.priorityClassName || '默认优先级',
    },
  ];

  // 渲染容器详情卡片
  const renderContainerCard = (container: any, isInit = false) => {
    const containerStatus = containerStatuses.find(
      (cs: any) => cs.name === container.name
    );
    const state = containerStatus?.state || {};
    const stateKey = Object.keys(state)[0] || 'waiting';
    const stateValue = state[stateKey] || {};

    return {
      title: isInit ? `Init: ${container.name}` : container.name,
      items: [
        { label: '镜像', value: container.image },
        {
          label: '状态',
          value: containerStatus?.state
            ? Object.keys(containerStatus.state)[0]
            : 'Unknown',
        },
        {
          label: '就绪',
          value: containerStatus?.ready ? '✅' : '❌',
        },
        {
          label: '重启次数',
          value: containerStatus?.restartCount || 0,
        },
      ],
    };
  };

  // 渲染环境变量
  const renderEnvVars = (container: any) => {
    if (!container.env || container.env.length === 0) return null;

    return (
      <div className="env-section">
        <h4 className="subsection-title">{container.name} - 环境变量</h4>
        <div className="env-grid">
          {container.env.map((env: any, idx: number) => (
            <div key={idx} className="env-item">
              <span className="env-name">{env.name}</span>
              <span className="env-value">
                {env.value !== undefined
                  ? env.value
                  : env.valueFrom
                  ? JSON.stringify(env.valueFrom)
                  : '-'}
              </span>
            </div>
          ))}
        </div>
      </div>
    );
  };

  // 渲染探针配置
  const renderProbe = (probe: any, title: string) => {
    if (!probe) return null;

    const probeType = probe.httpGet
      ? `HTTP: ${probe.httpGet.path || '/'}:${probe.httpGet.port || ''}`
      : probe.exec
      ? `Exec: ${probe.exec.command?.join(' ') || ''}`
      : probe.tcpSocket
      ? `TCP: ${probe.tcpSocket.port || ''}`
      : null;

    if (!probeType) return null;

    return (
      <div className="probe-item">
        <span className="probe-title">{title}</span>
        <span className="probe-value">{probeType}</span>
      </div>
    );
  };

  return (
    <div className="pod-detail">
      {/* 状态概览 */}
      <div className="detail-section">
        <h3 className="section-title">📊 状态概览</h3>
        <StatusCards cards={statusCards} />

        {/* 网络信息 */}
        <div className="network-info info-card">
          <h4 className="card-title">🌐 网络信息</h4>
          <div className="info-grid">
            <div className="info-item">
              <span className="info-label">Pod IP</span>
              <span className="info-value">{podIP || '-'}</span>
            </div>
            <div className="info-item">
              <span className="info-label">Host IP</span>
              <span className="info-value">{hostIP || '-'}</span>
            </div>
            {status?.podIPs?.length > 0 && (
              <div className="info-item">
                <span className="info-label">IP 列表</span>
                <span className="info-value">
                  {status.podIPs.map((ip: any) => ip.ip).join(', ')}
                </span>
              </div>
            )}
          </div>
        </div>
      </div>

      {/* 调度信息 */}
      <div className="detail-section">
        <h3 className="section-title">📍 调度信息</h3>
        <div className="info-card">
          <div className="info-grid">
            <div className="info-item">
              <span className="info-label">节点</span>
              <span
                className="info-value link"
                onClick={() =>
                  onResourceClick?.('node', '', spec.nodeName)
                }
              >
                {spec.nodeName || '-'}
              </span>
            </div>
            <div className="info-item">
              <span className="info-label">命名空间</span>
              <span className="info-value">{metadata.namespace || '-'}</span>
            </div>
            <div className="info-item">
              <span className="info-label">服务账户</span>
              <span className="info-value">
                {spec.serviceAccountName || 'default'}
              </span>
            </div>
            <div className="info-item">
              <span className="info-label">节点选择器</span>
              <span className="info-value">
                {spec.nodeSelector
                  ? Object.entries(spec.nodeSelector)
                      .map(([k, v]) => `${k}=${v}`)
                      .join(', ')
                  : '-'}
              </span>
            </div>
          </div>
        </div>
      </div>

      {/* 容器列表 */}
      <div className="detail-section">
        <h3 className="section-title">📦 容器</h3>

        {/* Init 容器 */}
        {initContainers.length > 0 && (
          <div className="subsection">
            <h4 className="subsection-title">Init 容器</h4>
            <ContainerList
              containers={initContainers}
              statuses={status?.initContainerStatuses || []}
              onViewLog={onViewLog}
              onExec={onExec}
            />
          </div>
        )}

        {/* 主容器 */}
        {containers.length > 0 && (
          <div className="subsection">
            <h4 className="subsection-title">应用容器</h4>
            <ContainerList
              containers={containers}
              statuses={containerStatuses}
              onViewLog={onViewLog}
              onExec={onExec}
            />
          </div>
        )}

        {/* 容器详情 */}
        {containers.map((container: any) => (
          <div key={container.name} className="container-detail-card">
            <h4 className="card-title">{container.name}</h4>

            {/* 探针配置 */}
            {(container.readinessProbe || container.livenessProbe || container.startupProbe) && (
              <div className="probe-section">
                <span className="subsection-label">探针配置</span>
                <div className="probe-list">
                  {renderProbe(container.readinessProbe, '就绪探针')}
                  {renderProbe(container.livenessProbe, '存活探针')}
                  {renderProbe(container.startupProbe, '启动探针')}
                </div>
              </div>
            )}

            {/* 资源请求/限制 */}
            {container.resources && (
              <div className="resources-section">
                <span className="subsection-label">资源配置</span>
                <div className="resources-grid">
                  <div className="resource-box requests">
                    <span className="resource-label">请求</span>
                    <div className="resource-values">
                      <span>CPU: {container.resources.requests?.cpu || '-'}</span>
                      <span>内存：{container.resources.requests?.memory || '-'}</span>
                    </div>
                  </div>
                  <div className="resource-box limits">
                    <span className="resource-label">限制</span>
                    <div className="resource-values">
                      <span>CPU: {container.resources.limits?.cpu || '-'}</span>
                      <span>内存：{container.resources.limits?.memory || '-'}</span>
                    </div>
                  </div>
                </div>
              </div>
            )}

            {/* 端口映射 */}
            {container.ports && container.ports.length > 0 && (
              <div className="ports-section">
                <span className="subsection-label">端口</span>
                <div className="ports-grid">
                  {container.ports.map((port: any, idx: number) => (
                    <div key={idx} className="port-item">
                      <span className="port-name">{port.name || 'N/A'}</span>
                      <span className="port-value">
                        {port.containerPort}/{port.protocol || 'TCP'}
                        {port.hostPort ? ` → ${port.hostPort}` : ''}
                      </span>
                    </div>
                  ))}
                </div>
              </div>
            )}

            {/* 环境变量 */}
            {renderEnvVars(container)}

            {/* 卷挂载 */}
            {container.volumeMounts && container.volumeMounts.length > 0 && (
              <div className="volume-mounts-section">
                <span className="subsection-label">卷挂载</span>
                <div className="mounts-grid">
                  {container.volumeMounts.map((mount: any, idx: number) => (
                    <div key={idx} className="mount-item">
                      <span className="mount-name">{mount.name}</span>
                      <span className="mount-path">{mount.mountPath}</span>
                      <span className="mount-readonly">
                        {mount.readOnly ? '🔒 只读' : '✏️ 读写'}
                      </span>
                    </div>
                  ))}
                </div>
              </div>
            )}
          </div>
        ))}
      </div>

      {/* 卷信息 */}
      {spec?.volumes && spec.volumes.length > 0 && (
        <div className="detail-section">
          <h3 className="section-title">💾 卷</h3>
          <div className="volumes-grid">
            {spec.volumes.map((volume: any, idx: number) => {
              const volumeType = volume.persistentVolumeClaim
                ? 'PVC'
                : volume.configMap
                ? 'ConfigMap'
                : volume.secret
                ? 'Secret'
                : volume.emptyDir
                ? 'EmptyDir'
                : volume.hostPath
                ? 'HostPath'
                : 'Unknown';

              return (
                <div key={idx} className="volume-card">
                  <div className="volume-header">
                    <span className="volume-name">{volume.name}</span>
                    <span className="volume-type">{volumeType}</span>
                  </div>
                  <div className="volume-details">
                    {volume.persistentVolumeClaim && (
                      <span>
                        PVC: {volume.persistentVolumeClaim.claimName}
                      </span>
                    )}
                    {volume.configMap && (
                      <span>ConfigMap: {volume.configMap.name}</span>
                    )}
                    {volume.secret && (
                      <span>Secret: {volume.secret.secretName}</span>
                    )}
                    {volume.emptyDir && (
                      <span>
                        EmptyDir
                        {volume.emptyDir.medium && ` (${volume.emptyDir.medium})`}
                      </span>
                    )}
                    {volume.hostPath && (
                      <span>
                        HostPath: {volume.hostPath.path}
                        {volume.hostPath.type && ` (${volume.hostPath.type})`}
                      </span>
                    )}
                  </div>
                </div>
              );
            })}
          </div>
        </div>
      )}

      {/* 标签和注解 */}
      <div className="detail-section">
        <h3 className="section-title">🏷️ 标签和注解</h3>
        {metadata?.labels && Object.keys(metadata.labels).length > 0 && (
          <div className="labels-subsection">
            <h4 className="subsection-title">标签</h4>
            <LabelList
              labels={metadata.labels}
              resourceType="pod"
              namespace={metadata.namespace}
            />
          </div>
        )}
        {metadata?.annotations && Object.keys(metadata.annotations).length > 0 && (
          <div className="labels-subsection">
            <h4 className="subsection-title">注解</h4>
            <LabelList labels={metadata.annotations} />
          </div>
        )}
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

export default PodDetail;
