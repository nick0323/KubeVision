/**
 * Node 详情页面组件
 * 专为 Node 设计，展示资源使用、分配情况、条件状态、关联 Pod 等信息
 */
import React from 'react';
import { StatusCards } from '../StatusCards';
import { ResourceBar } from '../ResourceBar';
import { LabelList } from '../LabelList';
import { EventTimeline } from '../EventTimeline';
import './ResourceDetail.css';

interface NodeDetailProps {
  data: any;
  events?: any[];
  pods?: any[];
  onResourceClick?: (type: string, namespace: string, name: string) => void;
}

export const NodeDetail: React.FC<NodeDetailProps> = ({
  data,
  events = [],
  pods = [],
  onResourceClick,
}) => {
  if (!data) return null;

  const { metadata, status } = data;
  
  // 添加空值检查
  if (!metadata || !status) {
    return (
      <div className="empty-state">
        <span className="empty-icon">⚠️</span>
        <p>数据不完整</p>
      </div>
    );
  }

  const conditions = status?.conditions || [];
  const addresses = status?.addresses || [];
  const allocatable = status?.allocatable || {};
  const capacity = status?.capacity || {};

  // 节点信息
  const nodeInfo = status?.nodeInfo || {};
  const internalIP = addresses?.find((a: any) => a.type === 'InternalIP')?.address;
  const externalIP = addresses?.find((a: any) => a.type === 'ExternalIP')?.address;
  const hostName = addresses?.find((a: any) => a.type === 'Hostname')?.address;

  // 角色
  const roles = Object.keys(metadata?.labels || {})
    .filter((k) => k.includes('node-role.kubernetes.io/'))
    .map((k) => k.replace('node-role.kubernetes.io/', ''));
  const roleDisplay = roles.length > 0 ? roles.join(', ') : 'worker';

  // 条件状态
  const readyCondition = conditions.find((c: any) => c.type === 'Ready');
  const memoryPressure = conditions.find((c: any) => c.type === 'MemoryPressure');
  const diskPressure = conditions.find((c: any) => c.type === 'DiskPressure');
  const pidPressure = conditions.find((c: any) => c.type === 'PIDPressure');
  const networkUnavailable = conditions.find((c: any) => c.type === 'NetworkUnavailable');

  const isReady = readyCondition?.status === 'True';
  const hasMemoryPressure = memoryPressure?.status === 'True';
  const hasDiskPressure = diskPressure?.status === 'True';
  const hasPidPressure = pidPressure?.status === 'True';
  const hasNetworkIssue = networkUnavailable?.status === 'True';

  // 状态卡片
  const statusCards = [
    {
      title: '状态',
      value: isReady ? 'Ready' : 'Not Ready',
      sub: hasNetworkIssue ? '⚠️ 网络不可用' : '-',
    },
    {
      title: '角色',
      value: roleDisplay,
      sub: nodeInfo.kubeletVersion || '-',
    },
    {
      title: 'Pod 容量',
      value: `${pods.length}/${allocatable.pods || capacity.pods || '110'}`,
      sub: `可调度：${isReady && !hasMemoryPressure && !hasDiskPressure && !hasPidPressure ? '✅' : '❌'}`,
    },
  ];

  // 资源使用（如果有 metrics 数据）
  const renderResourceBars = () => {
    // 这里使用模拟数据，实际应从 metrics API 获取
    const cpuCapacity = parseInt(allocatable.cpu || capacity.cpu || '0');
    const memCapacity = parseInt(allocatable.memory || capacity.memory || '0') / (1024 * 1024 * 1024);
    const podCapacity = parseInt(allocatable.pods || capacity.pods || '110');

    const cpuPercent = pods.length > 0 ? 45 : 0; // 模拟数据
    const memPercent = pods.length > 0 ? 62 : 0; // 模拟数据
    const podPercent = (pods.length / podCapacity) * 100;

    return (
      <div className="resource-bars">
        <ResourceBar
          items={[
            {
              label: 'CPU',
              value: `${(cpuCapacity * cpuPercent / 100).toFixed(1)} / ${cpuCapacity} 核`,
              percentage: cpuPercent,
              type: 'cpu',
            },
            {
              label: '内存',
              value: `${(memCapacity * memPercent / 100).toFixed(1)} / ${memCapacity.toFixed(1)} Gi`,
              percentage: memPercent,
              type: 'memory',
            },
            {
              label: 'Pod',
              value: `${pods.length} / ${podCapacity}`,
              percentage: podPercent,
              type: 'pods',
            },
          ]}
        />
      </div>
    );
  };

  // 渲染条件状态
  const renderConditions = () => {
    return (
      <div className="conditions-grid">
        <div className={`condition-card ${isReady ? 'good' : 'bad'}`}>
          <span className="condition-icon">{isReady ? '✅' : '❌'}</span>
          <span className="condition-name">Ready</span>
          <span className="condition-value">{isReady ? 'True' : 'False'}</span>
          {readyCondition?.message && (
            <span className="condition-message">{readyCondition.message}</span>
          )}
        </div>

        <div className={`condition-card ${!hasMemoryPressure ? 'good' : 'bad'}`}>
          <span className="condition-icon">{!hasMemoryPressure ? '✅' : '⚠️'}</span>
          <span className="condition-name">Memory Pressure</span>
          <span className="condition-value">{hasMemoryPressure ? 'True' : 'False'}</span>
        </div>

        <div className={`condition-card ${!hasDiskPressure ? 'good' : 'bad'}`}>
          <span className="condition-icon">{!hasDiskPressure ? '✅' : '⚠️'}</span>
          <span className="condition-name">Disk Pressure</span>
          <span className="condition-value">{hasDiskPressure ? 'True' : 'False'}</span>
        </div>

        <div className={`condition-card ${!hasPidPressure ? 'good' : 'bad'}`}>
          <span className="condition-icon">{!hasPidPressure ? '✅' : '⚠️'}</span>
          <span className="condition-name">PID Pressure</span>
          <span className="condition-value">{hasPidPressure ? 'True' : 'False'}</span>
        </div>

        <div className={`condition-card ${!hasNetworkIssue ? 'good' : 'bad'}`}>
          <span className="condition-icon">{!hasNetworkIssue ? '✅' : '❌'}</span>
          <span className="condition-name">Network</span>
          <span className="condition-value">{hasNetworkIssue ? 'Unavailable' : 'Available'}</span>
        </div>
      </div>
    );
  };

  // 渲染 Pod 列表
  const renderPodList = () => {
    if (pods.length === 0) {
      return (
        <div className="empty-state">
          <span className="empty-icon">📭</span>
          <p>此节点上没有运行 Pod</p>
        </div>
      );
    }

    return (
      <div className="pod-table">
        <table className="resource-table">
          <thead>
            <tr>
              <th>名称</th>
              <th>命名空间</th>
              <th>状态</th>
              <th>就绪</th>
              <th>重启</th>
              <th>IP</th>
              <th>年龄</th>
            </tr>
          </thead>
          <tbody>
            {pods.map((pod: any) => {
              const containerStatuses = pod.status?.containerStatuses || [];
              const readyCount = containerStatuses.filter((c: any) => c.ready).length;
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
                  <td>{pod.metadata.namespace}</td>
                  <td>
                    <span className={`status-badge status-${pod.status?.phase?.toLowerCase()}`}>
                      {pod.status?.phase || 'Unknown'}
                    </span>
                  </td>
                  <td>
                    {readyCount}/{containerStatuses.length}
                  </td>
                  <td>{restartCount}</td>
                  <td>{pod.status.podIP || '-'}</td>
                  <td>
                    {pod.metadata.creationTimestamp
                      ? new Date(pod.metadata.creationTimestamp).toLocaleDateString()
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
    <div className="node-detail">
      {/* 状态概览 */}
      <div className="detail-section">
        <h3 className="section-title">📊 状态概览</h3>
        <StatusCards cards={statusCards} />

        {/* 资源使用条 */}
        {renderResourceBars()}

        {/* 条件状态 */}
        <div className="info-card" style={{ marginTop: '16px' }}>
          <h4 className="card-title">❤️ 健康状态</h4>
          {renderConditions()}
        </div>
      </div>

      {/* 节点信息 */}
      <div className="detail-section">
        <h3 className="section-title">🖥️ 节点信息</h3>
        <div className="info-card">
          <div className="info-grid">
            <div className="info-item">
              <span className="info-label">主机名</span>
              <span className="info-value">{hostName || '-'}</span>
            </div>
            <div className="info-item">
              <span className="info-label">内部 IP</span>
              <span className="info-value">{internalIP || '-'}</span>
            </div>
            <div className="info-item">
              <span className="info-label">外部 IP</span>
              <span className="info-value">{externalIP || '-'}</span>
            </div>
            <div className="info-item">
              <span className="info-label">角色</span>
              <span className="info-value">{roleDisplay}</span>
            </div>
            <div className="info-item">
              <span className="info-label">Kubelet 版本</span>
              <span className="info-value">{nodeInfo.kubeletVersion || '-'}</span>
            </div>
            <div className="info-item">
              <span className="info-label">容器运行时</span>
              <span className="info-value">{nodeInfo.containerRuntimeVersion || '-'}</span>
            </div>
            <div className="info-item">
              <span className="info-label">操作系统</span>
              <span className="info-value">{nodeInfo.osImage || '-'}</span>
            </div>
            <div className="info-item">
              <span className="info-label">架构</span>
              <span className="info-value">{nodeInfo.architecture || '-'}</span>
            </div>
            <div className="info-item">
              <span className="info-label">内核版本</span>
              <span className="info-value">{nodeInfo.kernelVersion || '-'}</span>
            </div>
          </div>
        </div>
      </div>

      {/* 资源分配 */}
      <div className="detail-section">
        <h3 className="section-title">📈 资源分配</h3>
        <div className="allocatable-card info-card">
          <div className="allocatable-grid">
            <div className="allocatable-item">
              <span className="allocatable-label">CPU 容量</span>
              <span className="allocatable-value">{capacity.cpu || '-'}</span>
            </div>
            <div className="allocatable-item">
              <span className="allocatable-label">CPU 可分配</span>
              <span className="allocatable-value">{allocatable.cpu || '-'}</span>
            </div>
            <div className="allocatable-item">
              <span className="allocatable-label">内存容量</span>
              <span className="allocatable-value">
                {capacity.memory ? (parseInt(capacity.memory) / (1024 * 1024 * 1024)).toFixed(2) + ' Gi' : '-'}
              </span>
            </div>
            <div className="allocatable-item">
              <span className="allocatable-label">内存可分配</span>
              <span className="allocatable-value">
                {allocatable.memory ? (parseInt(allocatable.memory) / (1024 * 1024 * 1024)).toFixed(2) + ' Gi' : '-'}
              </span>
            </div>
            <div className="allocatable-item">
              <span className="allocatable-label">Pod 容量</span>
              <span className="allocatable-value">{capacity.pods || '-'}</span>
            </div>
            <div className="allocatable-item">
              <span className="allocatable-label">Pod 可分配</span>
              <span className="allocatable-value">{allocatable.pods || '-'}</span>
            </div>
            <div className="allocatable-item">
              <span className="allocatable-label">Ephemeral Storage</span>
              <span className="allocatable-value">
                {allocatable['ephemeral-storage'] 
                  ? (parseInt(allocatable['ephemeral-storage']) / (1024 * 1024 * 1024)).toFixed(2) + ' Gi' 
                  : '-'}
              </span>
            </div>
          </div>
        </div>
      </div>

      {/* 系统信息 */}
      <div className="detail-section">
        <h3 className="section-title">🔧 系统信息</h3>
        <div className="info-card">
          <div className="info-grid">
            <div className="info-item">
              <span className="info-label">机器 ID</span>
              <span className="info-value code">{nodeInfo.machineID || '-'}</span>
            </div>
            <div className="info-item">
              <span className="info-label">系统 UUID</span>
              <span className="info-value code">{nodeInfo.systemUUID || '-'}</span>
            </div>
            <div className="info-item">
              <span className="info-label">启动 ID</span>
              <span className="info-value code">{nodeInfo.bootID || '-'}</span>
            </div>
          </div>
        </div>
      </div>

      {/* 标签 */}
      <div className="detail-section">
        <h3 className="section-title">🏷️ 标签</h3>
        {metadata?.labels && Object.keys(metadata.labels).length > 0 && (
          <LabelList
            labels={metadata.labels}
            resourceType="node"
            onLabelClick={(key, value) => {
              // 处理节点角色标签点击
              if (key.includes('node-role.kubernetes.io/')) {
                console.log('Role:', value);
              }
            }}
          />
        )}
      </div>

      {/* 注解 */}
      <div className="detail-section">
        <h3 className="section-title">📝 注解</h3>
        {metadata?.annotations && Object.keys(metadata.annotations).length > 0 && (
          <LabelList labels={metadata.annotations} />
        )}
      </div>

      {/* 关联 Pod */}
      <div className="detail-section">
        <h3 className="section-title">🔗 运行中的 Pod ({pods.length})</h3>
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

export default NodeDetail;
