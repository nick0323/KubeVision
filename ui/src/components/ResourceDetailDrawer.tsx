import React, { useState, useEffect } from 'react';
import { StatusCards } from './StatusCards';
import { ResourceBar } from './ResourceBar';
import { ContainerList } from './ContainerList';
import { EventTimeline } from './EventTimeline';
import { YamlEditor } from './YamlEditor';
import { LabelList } from './LabelList';
import './ResourceDetailDrawer.css';

interface ResourceDetailDrawerProps {
  visible: boolean;
  resourceType: string;
  data: any;
  loading: boolean;
  onClose: () => void;
  onRefresh: () => void;
  onResourceClick?: (type: string, namespace: string, name: string) => void;  // 资源点击回调
}

export const ResourceDetailDrawer: React.FC<ResourceDetailDrawerProps> = ({
  visible,
  resourceType,
  data,
  loading,
  onClose,
  onRefresh,
  onResourceClick,
}) => {
  const [activeTab, setActiveTab] = useState('overview');

  // 获取状态值（根据不同资源类型）
  const getStatusValue = (data: any): string => {
    if (!data) return 'Unknown';
    
    // Pod 有 phase
    if (data.status?.phase) return data.status.phase;
    
    // Deployment/StatefulSet/DaemonSet 检查条件
    if (data.status?.conditions) {
      const available = data.status.conditions.find((c: any) => c.type === 'Available');
      if (available) return available.status === 'True' ? 'Available' : 'Unavailable';
    }
    
    // Node 检查条件
    if (data.status?.conditions) {
      const ready = data.status.conditions.find((c: any) => c.type === 'Ready');
      if (ready) return ready.status === 'True' ? 'Ready' : 'NotReady';
    }
    
    // 默认返回 status 字符串或 Unknown
    if (typeof data.status === 'string') return data.status;
    
    return 'Unknown';
  };

  // 获取状态子信息
  const getStatusSub = (data: any): string => {
    if (!data || !data.status?.conditions) return '';
    
    const ready = data.status.conditions.find((c: any) => c.type === 'Ready');
    if (ready) return ready.status === 'True' ? 'Ready' : 'NotReady';
    
    const available = data.status.conditions.find((c: any) => c.type === 'Available');
    if (available) return available.status === 'True' ? 'Available' : 'Unavailable';
    
    return '';
  };

  // 关闭时重置标签页
  useEffect(() => {
    if (!visible) {
      setActiveTab('overview');
    }
  }, [visible]);

  // 处理资源跳转
  const handleResourceClick = (type: string, namespace: string, name: string) => {
    if (onResourceClick) {
      onResourceClick(type, namespace, name);
    }
  };

  if (!visible) {
    return null;
  }

  return (
    <div className="resource-detail-drawer-overlay" onClick={onClose}>
      <div className="resource-detail-drawer" onClick={(e) => e.stopPropagation()}>
        {/* 抽屉头部 */}
        <div className="drawer-header">
          <div className="drawer-title">
            <h2>{data?.metadata?.name || resourceType}</h2>
            <span className="drawer-subtitle">
              {resourceType} / {data?.metadata?.namespace || 'N/A'}
            </span>
          </div>
          <button className="drawer-close" onClick={onClose}>
            ✕
          </button>
        </div>

        {/* 标签页导航 */}
        <div className="drawer-tabs">
          <div
            className={`drawer-tab ${activeTab === 'overview' ? 'active' : ''}`}
            onClick={() => setActiveTab('overview')}
          >
            概览
          </div>
          {(resourceType === 'pod' || resourceType === 'pods') && (
            <div
              className={`drawer-tab ${activeTab === 'containers' ? 'active' : ''}`}
              onClick={() => setActiveTab('containers')}
            >
              容器
            </div>
          )}
          <div
            className={`drawer-tab ${activeTab === 'events' ? 'active' : ''}`}
            onClick={() => setActiveTab('events')}
          >
            事件
          </div>
          <div
            className={`drawer-tab ${activeTab === 'yaml' ? 'active' : ''}`}
            onClick={() => setActiveTab('yaml')}
          >
            YAML
          </div>
        </div>

        {/* 抽屉内容区 */}
        <div className="drawer-content">
          {loading ? (
            <div className="drawer-loading">
              <div className="loading-spinner">加载中...</div>
            </div>
          ) : data ? (
            <>
              {/* 概览标签页 */}
              {activeTab === 'overview' && (
                <>
                  <div className="drawer-section">
                    <h3 className="section-title">状态概览</h3>
                    <StatusCards
                      cards={[
                        {
                          title: '状态',
                          value: getStatusValue(data),
                          sub: getStatusSub(data),
                        },
                        {
                          title: '资源使用',
                          value: 'CPU: 100m',
                          sub: 'Mem: 128Mi',
                        },
                        {
                          title: '运行时间',
                          value: data.metadata?.creationTimestamp ?
                            new Date(data.metadata.creationTimestamp).toLocaleDateString() : 'N/A',
                          sub: `创建：${data.metadata?.creationTimestamp || ''}`,
                        },
                      ]}
                    />
                    {/* Node 特有：资源进度条 */}
                    {(resourceType === 'node' || resourceType === 'nodes') && (
                      <ResourceBar
                        items={[
                          { label: 'CPU 使用率', value: '5.2 / 8 核 (65%)', percentage: 65, type: 'cpu' },
                          { label: '内存使用率', value: '25.0 / 32 Gi (78%)', percentage: 78, type: 'memory' },
                          { label: 'Pod 使用率', value: '22 / 110 (20%)', percentage: 20, type: 'pods' },
                        ]}
                      />
                    )}
                  </div>

                  <div className="drawer-section">
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
                      {data.spec?.nodeName && (
                        <div className="info-item" onClick={() => handleResourceClick('node', '', data.spec.nodeName)}>
                          <div className="info-label">节点</div>
                          <div className="info-value link">{data.spec.nodeName}</div>
                        </div>
                      )}
                      {data.status?.podIP && (
                        <div className="info-item">
                          <div className="info-label">IP 地址</div>
                          <div className="info-value">{data.status.podIP}</div>
                        </div>
                      )}
                      <div className="info-item">
                        <div className="info-label">创建时间</div>
                        <div className="info-value">{data.metadata?.creationTimestamp}</div>
                      </div>
                      <div className="info-item">
                        <div className="info-label">UID</div>
                        <div className="info-value">{data.metadata?.uid}</div>
                      </div>
                    </div>

                    {/* 标签 */}
                    {data.metadata?.labels && Object.keys(data.metadata.labels).length > 0 && (
                      <div className="labels-section">
                        <div className="info-label">标签</div>
                        <LabelList 
                          labels={data.metadata.labels}
                          resourceType={resourceType}
                          namespace={data.metadata?.namespace}
                          onLabelClick={handleResourceClick}
                        />
                      </div>
                    )}

                    {/* 注解 */}
                    {data.metadata?.annotations && Object.keys(data.metadata.annotations).length > 0 && (
                      <div className="labels-section">
                        <div className="info-label">注解</div>
                        <LabelList labels={data.metadata.annotations} />
                      </div>
                    )}
                  </div>
                </>
              )}

              {/* 容器标签页 */}
              {activeTab === 'containers' && data.spec?.containers && (
                <div className="drawer-section">
                  <ContainerList 
                    containers={data.spec.containers}
                    onContainerClick={() => {}}  // TODO: 容器详情
                  />
                </div>
              )}

              {/* 事件标签页 */}
              {activeTab === 'events' && (
                <div className="drawer-section">
                  <EventTimeline 
                    events={data.events || []}
                    onEventClick={() => {}}  // TODO: 事件详情
                  />
                </div>
              )}

              {/* YAML 标签页 */}
              {activeTab === 'yaml' && (
                <div className="drawer-section">
                  <YamlEditor 
                    data={data} 
                    onRefresh={onRefresh}
                    onSave={(yamlStr) => console.log('Save YAML:', yamlStr)}  // TODO: 实现保存
                    onCompare={() => console.log('Compare YAML')}  // TODO: 实现对比
                  />
                </div>
              )}
            </>
          ) : (
            <div className="drawer-empty">
              <div className="empty-icon">📭</div>
              <p>暂无数据</p>
            </div>
          )}
        </div>
      </div>
    </div>
  );
};

export default ResourceDetailDrawer;
