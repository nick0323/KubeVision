import React, { useState } from 'react';
import { StatusCards } from './StatusCards';
import { ResourceBar } from './ResourceBar';
import { LabelList } from './LabelList';
import { ContainerList } from './ContainerList';
import { EventTimeline } from './EventTimeline';
import { YamlViewer } from './YamlViewer';
import './ResourceDetailTabs.css';

interface ResourceDetailTabsProps {
  resourceType: string;
  activeTab: string;
  onTabChange: (tab: string) => void;
  data: any;
  onRefresh: () => void;
}

export const ResourceDetailTabs: React.FC<ResourceDetailTabsProps> = ({
  resourceType,
  activeTab,
  onTabChange,
  data,
  onRefresh,
}) => {
  const [yamlData, setYamlData] = useState<any>(null);

  // 渲染概览内容
  const renderOverview = () => {
    switch (resourceType) {
      case 'pod':
      case 'pods':
        return renderPodOverview();
      case 'node':
      case 'nodes':
        return renderNodeOverview();
      case 'deployment':
      case 'deployments':
        return renderDeploymentOverview();
      default:
        return renderDefaultOverview();
    }
  };

  // Pod 概览
  const renderPodOverview = () => {
    const statusCards = [
      {
        title: '状态',
        value: data.status?.phase || 'Unknown',
        sub: `Ready: ${data.status?.conditions?.find((c: any) => c.type === 'Ready')?.status || 'N/A'}`,
      },
      {
        title: '资源使用',
        value: 'CPU: 100m',
        sub: 'Mem: 128Mi',
      },
      {
        title: '运行时间',
        value: '2h 15m',
        sub: `创建：${data.metadata?.creationTimestamp || 'N/A'}`,
      },
    ];

    return (
      <>
        <div className="overview-section">
          <h3 className="section-title">状态概览</h3>
          <StatusCards cards={statusCards} />
        </div>

        <div className="overview-section">
          <h3 className="section-title">基本信息</h3>
          <div className="info-grid">
            <div className="info-item">
              <div className="info-label">名称</div>
              <div className="info-value">{data.metadata?.name}</div>
            </div>
            <div className="info-item">
              <div className="info-label">命名空间</div>
              <div className="info-value">{data.metadata?.namespace}</div>
            </div>
            <div className="info-item">
              <div className="info-label">节点</div>
              <div className="info-value link">{data.spec?.nodeName}</div>
            </div>
            <div className="info-item">
              <div className="info-label">IP 地址</div>
              <div className="info-value">{data.status?.podIP}</div>
            </div>
            <div className="info-item">
              <div className="info-label">创建时间</div>
              <div className="info-value">{data.metadata?.creationTimestamp}</div>
            </div>
            <div className="info-item">
              <div className="info-label">UID</div>
              <div className="info-value">{data.metadata?.uid}</div>
            </div>
          </div>

          {data.metadata?.labels && (
            <div style={{ marginTop: '20px' }}>
              <div className="info-label">标签</div>
              <LabelList labels={data.metadata.labels} />
            </div>
          )}

          {data.metadata?.annotations && (
            <div style={{ marginTop: '20px' }}>
              <div className="info-label">注解</div>
              <LabelList labels={data.metadata.annotations} />
            </div>
          )}
        </div>
      </>
    );
  };

  // Node 概览
  const renderNodeOverview = () => {
    const statusCards = [
      {
        title: 'Kubelet 版本',
        value: data.status?.nodeInfo?.kubeletVersion || 'N/A',
        sub: `Containerd: ${data.status?.nodeInfo?.containerRuntimeVersion || 'N/A'}`,
      },
      {
        title: '操作系统',
        value: data.status?.nodeInfo?.osImage || 'Linux',
        sub: data.status?.nodeInfo?.architecture || 'amd64',
      },
      {
        title: '运行时间',
        value: '30 天',
        sub: `启动：${data.metadata?.creationTimestamp || 'N/A'}`,
      },
    ];

    const resourceItems = [
      {
        label: 'CPU 使用率',
        value: '5.2 / 8 核 (65%)',
        percentage: 65,
        type: 'cpu' as const,
      },
      {
        label: '内存使用率',
        value: '25.0 / 32 Gi (78%)',
        percentage: 78,
        type: 'memory' as const,
      },
      {
        label: 'Pod 使用率',
        value: '22 / 110 (20%)',
        percentage: 20,
        type: 'pods' as const,
      },
    ];

    return (
      <>
        <div className="overview-section">
          <h3 className="section-title">状态概览</h3>
          <StatusCards cards={statusCards} />
          <ResourceBar items={resourceItems} />
        </div>

        <div className="overview-section">
          <h3 className="section-title">基本信息</h3>
          <div className="info-grid">
            <div className="info-item">
              <div className="info-label">名称</div>
              <div className="info-value">{data.metadata?.name}</div>
            </div>
            <div className="info-item">
              <div className="info-label">角色</div>
              <div className="info-value">
                {Object.keys(data.metadata?.labels || {})
                  .filter((k) => k.includes('node-role.kubernetes.io/'))
                  .map((k) => k.replace('node-role.kubernetes.io/', ''))
                  .join(', ') || 'worker'}
              </div>
            </div>
            <div className="info-item">
              <div className="info-label">内部 IP</div>
              <div className="info-value">
                {data.status?.addresses?.find((a: any) => a.type === 'InternalIP')?.address}
              </div>
            </div>
            <div className="info-item">
              <div className="info-label">外部 IP</div>
              <div className="info-value">
                {data.status?.addresses?.find((a: any) => a.type === 'ExternalIP')?.address || '-'}
              </div>
            </div>
            <div className="info-item">
              <div className="info-label">架构</div>
              <div className="info-value">{data.status?.nodeInfo?.architecture}</div>
            </div>
            <div className="info-item">
              <div className="info-label">内核版本</div>
              <div className="info-value">{data.status?.nodeInfo?.kernelVersion}</div>
            </div>
          </div>

          {data.metadata?.labels && (
            <div style={{ marginTop: '20px' }}>
              <div className="info-label">标签</div>
              <LabelList labels={data.metadata.labels} />
            </div>
          )}
        </div>
      </>
    );
  };

  // Deployment 概览
  const renderDeploymentOverview = () => {
    // 检查条件状态
    const conditions = data.status?.conditions || [];
    const available = conditions.find((c: any) => c.type === 'Available');
    const progressing = conditions.find((c: any) => c.type === 'Progressing');
    
    const statusValue = available?.status === 'True' ? 'Available' : 'Unavailable';
    const statusSub = progressing?.status === 'True' ? 'Progressing' : '';

    const statusCards = [
      {
        title: '状态',
        value: statusValue,
        sub: statusSub,
      },
      {
        title: '副本',
        value: `${data.status?.readyReplicas || 0} / ${data.spec?.replicas || 0}`,
        sub: '就绪/期望',
      },
      {
        title: '更新',
        value: data.status?.updatedReplicas || 0,
        sub: '已更新',
      },
      {
        title: '可用',
        value: data.status?.availableReplicas || 0,
        sub: '可用副本',
      },
    ];

    const strategy = data.spec?.strategy?.type || 'RollingUpdate';
    const maxSurge = data.spec?.strategy?.rollingUpdate?.maxSurge || '25%';
    const maxUnavailable =
      data.spec?.strategy?.rollingUpdate?.maxUnavailable || '25%';

    return (
      <>
        <div className="overview-section">
          <h3 className="section-title">状态概览</h3>
          <StatusCards cards={statusCards} />

          <div className="strategy-info">
            <div className="strategy-title">部署策略</div>
            <div className="strategy-grid">
              <div className="strategy-item">
                <span className="strategy-icon">🔄</span>
                <div className="strategy-text">{strategy}</div>
              </div>
              <div className="strategy-item">
                <span className="strategy-icon">⬆️</span>
                <div className="strategy-text">最大激增：{maxSurge}</div>
              </div>
              <div className="strategy-item">
                <span className="strategy-icon">⬇️</span>
                <div className="strategy-text">最大不可用：{maxUnavailable}</div>
              </div>
            </div>
          </div>
        </div>

        <div className="overview-section">
          <h3 className="section-title">基本信息</h3>
          <div className="info-grid">
            <div className="info-item">
              <div className="info-label">名称</div>
              <div className="info-value">{data.metadata?.name}</div>
            </div>
            <div className="info-item">
              <div className="info-label">命名空间</div>
              <div className="info-value">{data.metadata?.namespace}</div>
            </div>
            <div className="info-item">
              <div className="info-label">镜像</div>
              <div className="info-value">
                {data.spec?.template?.spec?.containers?.[0]?.image || 'N/A'}
              </div>
            </div>
            <div className="info-item">
              <div className="info-label">创建时间</div>
              <div className="info-value">{data.metadata?.creationTimestamp}</div>
            </div>
          </div>

          {data.spec?.selector?.matchLabels && (
            <div className="selector-box">
              <div className="selector-title">Pod 选择器</div>
              <div className="selector-labels">
                <LabelList labels={data.spec.selector.matchLabels} />
              </div>
            </div>
          )}

          {data.metadata?.labels && (
            <div style={{ marginTop: '20px' }}>
              <div className="info-label">标签</div>
              <LabelList labels={data.metadata.labels} />
            </div>
          )}
        </div>
      </>
    );
  };

  // 默认概览
  const renderDefaultOverview = () => {
    return (
      <div className="overview-section">
        <h3 className="section-title">基本信息</h3>
        <div className="info-grid">
          <div className="info-item">
            <div className="info-label">名称</div>
            <div className="info-value">{data.metadata?.name}</div>
          </div>
          <div className="info-item">
            <div className="info-label">命名空间</div>
            <div className="info-value">{data.metadata?.namespace}</div>
          </div>
          <div className="info-item">
            <div className="info-label">创建时间</div>
            <div className="info-value">{data.metadata?.creationTimestamp}</div>
          </div>
          <div className="info-item">
            <div className="info-label">UID</div>
            <div className="info-value">{data.metadata?.uid}</div>
          </div>
        </div>

        {data.metadata?.labels && (
          <div style={{ marginTop: '20px' }}>
            <div className="info-label">标签</div>
            <LabelList labels={data.metadata.labels} />
          </div>
        )}
      </div>
    );
  };

  // 渲染容器标签页
  const renderContainers = () => {
    const containers = data.spec?.containers || [];
    return (
      <div className="tab-section">
        <ContainerList
          containers={containers}
          onViewLog={(name) => console.log('View log for', name)}
          onExec={(name) => console.log('Exec into', name)}
        />
      </div>
    );
  };

  // 渲染事件标签页
  const renderEvents = () => {
    const events = data.events || [];
    return (
      <div className="tab-section">
        <EventTimeline events={events} />
      </div>
    );
  };

  // 渲染 YAML 标签页
  const renderYaml = () => {
    return (
      <div className="tab-section">
        <YamlViewer
          data={data}
          onRefresh={onRefresh}
          onEdit={(yaml) => console.log('Edit YAML', yaml)}
          onCompare={() => console.log('Compare')}
        />
      </div>
    );
  };

  return (
    <div className="resource-detail-tabs">
      <div className="tabs-header">
        <div
          className={`tab-item ${activeTab === 'overview' ? 'active' : ''}`}
          onClick={() => onTabChange('overview')}
        >
          概览
        </div>
        {(resourceType === 'pod' || resourceType === 'pods') && (
          <div
            className={`tab-item ${activeTab === 'containers' ? 'active' : ''}`}
            onClick={() => onTabChange('containers')}
          >
            容器
          </div>
        )}
        <div
          className={`tab-item ${activeTab === 'events' ? 'active' : ''}`}
          onClick={() => onTabChange('events')}
        >
          事件
        </div>
        <div
          className={`tab-item ${activeTab === 'yaml' ? 'active' : ''}`}
          onClick={() => onTabChange('yaml')}
        >
          YAML
        </div>
      </div>

      <div className="tab-content">
        {activeTab === 'overview' && renderOverview()}
        {activeTab === 'containers' && renderContainers()}
        {activeTab === 'events' && renderEvents()}
        {activeTab === 'yaml' && renderYaml()}
      </div>
    </div>
  );
};

export default ResourceDetailTabs;
