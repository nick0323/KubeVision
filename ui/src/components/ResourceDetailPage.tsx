import React, { useState, useEffect, useCallback } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { YamlEditor } from './YamlEditor';
import { LabelList } from './LabelList';
import { authFetch } from '../utils/auth';
import { PodDetail } from './resourceDetails/PodDetail';
import { DeploymentDetail } from './resourceDetails/DeploymentDetail';
import { NodeDetail } from './resourceDetails/NodeDetail';
import { ServiceDetail } from './resourceDetails/ServiceDetail';
import { ConfigDetail } from './resourceDetails/ConfigDetail';
import { Breadcrumb } from './Breadcrumb';
import { QuickActions } from './QuickActions';
import './ResourceDetailPage.css';

interface ResourceDetailPageProps {
  resourceType: string;
}

export const ResourceDetailPage: React.FC<ResourceDetailPageProps> = ({ resourceType }) => {
  const { namespace, name } = useParams<{ namespace: string; name: string }>();
  const navigate = useNavigate();
  const [data, setData] = useState<any>(null);
  const [relatedData, setRelatedData] = useState<any>({});
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [activeTab, setActiveTab] = useState('overview');
  const [lastRefresh, setLastRefresh] = useState<Date>(new Date());

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

  // 加载数据
  const loadData = useCallback(async () => {
    setLoading(true);
    setError(null);

    try {
      const ns = namespace || '';
      const response = await authFetch(
        `/api/${resourceType}/${ns}/${name}`
      );
      const result = await response.json();

      if (result.code === 0 && result.data) {
        setData(result.data);
        setError(null);
        setLastRefresh(new Date());
        
        // 加载关联数据
        await loadRelatedData(result.data, ns);
      } else {
        setError(result.message || '加载失败');
      }
    } catch (err) {
      setError('加载失败');
    } finally {
      setLoading(false);
    }
  }, [resourceType, namespace, name]);

  // 加载关联数据（Pod、事件等）
  const loadRelatedData = async (resourceData: any, ns: string) => {
    try {
      const related: any = {};

      // 加载事件 - 修复：处理 404 和空命名空间
      try {
        const resourceName = resourceData.metadata?.name;
        if (ns && resourceName) {
          const eventResponse = await authFetch(
            `/api/events/${ns}/${resourceName}?limit=50`
          );
          if (eventResponse.ok) {
            const eventResult = await eventResponse.json();
            if (eventResult.code === 0 && eventResult.data) {
              related.events = Array.isArray(eventResult.data) 
                ? eventResult.data.filter((e: any) => e.involvedObject?.name === resourceName)
                : [];
            }
          }
        }
      } catch (e) {
        console.warn('加载事件失败:', e);
        related.events = [];
      }

      // 根据资源类型加载关联数据
      if (resourceType === 'pod' || resourceType === 'pods') {
        // Pod 不需要额外加载
      } else if (resourceType === 'deployment' || resourceType === 'deployments') {
        // 加载关联的 Pod
        try {
          const selector = resourceData.spec?.selector?.matchLabels;
          if (selector) {
            const labelSelector = Object.entries(selector)
              .map(([k, v]) => `${k}=${v}`)
              .join(',');
            const podResponse = await authFetch(
              `/api/pods/${ns}?labelSelector=${encodeURIComponent(labelSelector)}&limit=100`
            );
            const podResult = await podResponse.json();
            if (podResult.code === 0 && podResult.data) {
              related.pods = podResult.data;
            }
          }
        } catch (e) {
          console.warn('加载 Pod 列表失败:', e);
        }
      } else if (resourceType === 'node' || resourceType === 'nodes') {
        // 加载节点上的 Pod
        try {
          const podResponse = await authFetch(
            `/api/pods/?fieldSelector=${encodeURIComponent(`spec.nodeName=${resourceData.metadata?.name}`)}&limit=100`
          );
          const podResult = await podResponse.json();
          if (podResult.code === 0 && podResult.data) {
            related.pods = podResult.data;
          }
        } catch (e) {
          console.warn('加载节点 Pod 失败:', e);
        }
      } else if (resourceType === 'service' || resourceType === 'services') {
        // 加载 Endpoints
        try {
          const epResponse = await authFetch(
            `/api/endpoints/${ns}/${resourceData.metadata?.name}`
          );
          const epResult = await epResponse.json();
          if (epResult.code === 0 && epResult.data) {
            related.endpoints = epResult.data;
          }
        } catch (e) {
          console.warn('加载 Endpoints 失败:', e);
        }

        // 加载匹配的 Pod
        const selector = resourceData.spec?.selector;
        if (selector) {
          try {
            const labelSelector = Object.entries(selector)
              .map(([k, v]) => `${k}=${v}`)
              .join(',');
            const podResponse = await authFetch(
              `/api/pods/${ns}?labelSelector=${encodeURIComponent(labelSelector)}&limit=100`
            );
            const podResult = await podResponse.json();
            if (podResult.code === 0 && podResult.data) {
              related.pods = podResult.data;
            }
          } catch (e) {
            console.warn('加载 Pod 列表失败:', e);
          }
        }
      }

      setRelatedData(related);
    } catch (e) {
      console.warn('加载关联数据失败:', e);
    }
  };

  useEffect(() => {
    loadData();
  }, [loadData]);

  // 返回列表
  const handleBack = () => {
    navigate(-1); // 返回上一页
  };

  // 资源跳转
  const handleResourceClick = (type: string, ns: string, resourceName: string) => {
    navigate(`/${type}s/${ns}/${resourceName}`);
  };

  // 查看日志
  const handleViewLog = (containerName: string) => {
    console.log('查看日志:', containerName);
    // TODO: 实现日志查看
  };

  // 进入容器
  const handleExec = (containerName: string) => {
    console.log('进入容器:', containerName);
    // TODO: 实现 exec 功能
  };

  if (loading) {
    return (
      <div className="resource-detail-page">
        <div className="detail-loading">
          <div className="loading-spinner">加载中...</div>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="resource-detail-page">
        <div className="detail-error">
          <h3>加载失败</h3>
          <p>{error}</p>
          <button className="btn btn-primary" onClick={handleBack}>
            返回列表
          </button>
        </div>
      </div>
    );
  }

  if (!data) {
    return null;
  }

  // 根据资源类型渲染不同的详情组件
  const renderResourceDetail = () => {
    const commonProps = {
      data,
      events: relatedData.events || [],
      onResourceClick: handleResourceClick,
    };

    switch (resourceType) {
      case 'pod':
      case 'pods':
        return (
          <PodDetail
            {...commonProps}
            onViewLog={handleViewLog}
            onExec={handleExec}
          />
        );

      case 'deployment':
      case 'deployments':
        return (
          <DeploymentDetail
            {...commonProps}
            relatedPods={relatedData.pods || []}
          />
        );

      case 'node':
      case 'nodes':
        return (
          <NodeDetail
            {...commonProps}
            pods={relatedData.pods || []}
          />
        );

      case 'service':
      case 'services':
        return (
          <ServiceDetail
            {...commonProps}
            endpoints={relatedData.endpoints}
            pods={relatedData.pods || []}
          />
        );

      case 'configmap':
      case 'configmaps':
        return (
          <ConfigDetail
            {...commonProps}
            isSecret={false}
          />
        );

      case 'secret':
      case 'secrets':
        return (
          <ConfigDetail
            {...commonProps}
            isSecret={true}
          />
        );

      case 'statefulset':
      case 'statefulsets':
      case 'daemonset':
      case 'daemonsets':
      case 'job':
      case 'jobs':
      case 'cronjob':
      case 'cronjobs':
      case 'ingress':
      case 'ingresses':
      case 'pv':
      case 'pvs':
      case 'pvc':
      case 'pvcs':
      case 'storageclass':
      case 'storageclasses':
      case 'namespace':
      case 'namespaces':
      default:
        // 使用通用详情视图
        return (
          <div className="generic-detail">
            <div className="detail-section">
              <h3 className="section-title">📊 状态概览</h3>
              <div className="status-cards-placeholder">
                <div className="status-card">
                  <span className="card-label">状态</span>
                  <span className="card-value">{getStatusValue(data)}</span>
                  <span className="card-sub">{getStatusSub(data)}</span>
                </div>
                <div className="status-card">
                  <span className="card-label">名称</span>
                  <span className="card-value">{data.metadata?.name}</span>
                </div>
                <div className="status-card">
                  <span className="card-label">命名空间</span>
                  <span className="card-value">{data.metadata?.namespace || 'N/A'}</span>
                </div>
              </div>
            </div>

            <div className="detail-section">
              <h3 className="section-title">📋 基本信息</h3>
              <div className="info-card">
                <div className="info-grid">
                  <div className="info-item">
                    <span className="info-label">名称</span>
                    <span className="info-value">{data.metadata?.name}</span>
                  </div>
                  <div className="info-item">
                    <span className="info-label">命名空间</span>
                    <span className="info-value">{data.metadata?.namespace || 'N/A'}</span>
                  </div>
                  <div className="info-item">
                    <span className="info-label">创建时间</span>
                    <span className="info-value">{data.metadata?.creationTimestamp}</span>
                  </div>
                  <div className="info-item">
                    <span className="info-label">UID</span>
                    <span className="info-value">{data.metadata?.uid}</span>
                  </div>
                </div>
              </div>
            </div>

            {data.metadata?.labels && Object.keys(data.metadata.labels).length > 0 && (
              <div className="detail-section">
                <h3 className="section-title">🏷️ 标签</h3>
                <LabelList labels={data.metadata.labels} />
              </div>
            )}

            {data.metadata?.annotations && Object.keys(data.metadata.annotations).length > 0 && (
              <div className="detail-section">
                <h3 className="section-title">📝 注解</h3>
                <LabelList labels={data.metadata.annotations} />
              </div>
            )}
          </div>
        );
    }
  };

  // 获取标签页配置
  const getTabConfig = () => {
    const tabs = [
      { id: 'overview', label: '概览' },
      { id: 'yaml', label: 'YAML' },
    ];

    // Pod 有容器标签页
    if (resourceType === 'pod' || resourceType === 'pods') {
      tabs.splice(1, 0, { id: 'containers', label: '容器' });
    }

    // 所有资源都有事件标签页
    tabs.splice(tabs.length - 1, 0, { id: 'events', label: '事件' });

    return tabs;
  };

  const tabs = getTabConfig();

  return (
    <div className="resource-detail-page">
      {/* 面包屑导航 */}
      <Breadcrumb
        items={[
          { label: 'Cluster', href: '/' },
          { label: data.metadata?.namespace || 'Cluster-Scoped', type: 'NS' },
          { label: resourceType, href: `/${resourceType}` },
          { label: data.metadata?.name },
        ]}
      />

      {/* 页面头部 - 简化版 */}
      <div className="detail-header">
        <div className="header-left">
          <button className="btn-back" onClick={handleBack}>
            ← 返回
          </button>
          <div className="header-status">
            <span className={`status-badge status-${getStatusValue(data).toLowerCase()}`}>
              {getStatusValue(data)}
            </span>
          </div>
        </div>
        <div className="header-actions">
          {/* 快速操作区 */}
          <QuickActions
            resourceType={resourceType}
            resourceName={data.metadata?.name}
            namespace={data.metadata?.namespace}
            data={data}
            onRefresh={loadData}
          />

          <span className="last-refresh">
            最后刷新：{lastRefresh.toLocaleTimeString()}
          </span>
          <button className="btn btn-default" onClick={loadData}>
            🔄 刷新
          </button>
        </div>
      </div>

      {/* 标签页导航 */}
      <div className="detail-tabs">
        {tabs.map((tab) => (
          <div
            key={tab.id}
            className={`tab ${activeTab === tab.id ? 'active' : ''}`}
            onClick={() => setActiveTab(tab.id)}
          >
            {tab.label}
          </div>
        ))}
      </div>

      {/* 内容区 */}
      <div className="detail-content">
        {/* 概览标签页 - 渲染专业详情组件 */}
        {activeTab === 'overview' && renderResourceDetail()}

        {/* 容器标签页（仅 Pod） */}
        {activeTab === 'containers' && data.spec?.containers && (
          <div className="tab-content-section">
            <div className="detail-section">
              <h3 className="section-title">📦 容器列表</h3>
              {/* 这里可以复用 PodDetail 中的容器部分 */}
              {renderResourceDetail()}
            </div>
          </div>
        )}

        {/* 事件标签页 */}
        {activeTab === 'events' && (
          <div className="tab-content-section">
            <div className="detail-section">
              <h3 className="section-title">📅 事件</h3>
              {relatedData.events && relatedData.events.length > 0 ? (
                <div className="events-list">
                  {relatedData.events.map((event: any, idx: number) => (
                    <div key={idx} className="event-item">
                      <div className="event-header">
                        <span className={`event-type ${event.type?.toLowerCase()}`}>
                          {event.type || 'Normal'}
                        </span>
                        <span className="event-reason">{event.reason || '-'}</span>
                        <span className="event-time">
                          {event.lastTimestamp
                            ? new Date(event.lastTimestamp).toLocaleString()
                            : '-'}
                        </span>
                      </div>
                      <div className="event-message">{event.message || '-'}</div>
                      <div className="event-object">
                        对象：{event.involvedObject?.kind}/{event.involvedObject?.name}
                      </div>
                    </div>
                  ))}
                </div>
              ) : (
                <div className="empty-state">
                  <span className="empty-icon">📭</span>
                  <p>暂无事件</p>
                </div>
              )}
            </div>
          </div>
        )}

        {/* YAML 标签页 */}
        {activeTab === 'yaml' && (
          <div className="tab-content-section">
            <YamlEditor
              data={data}
              onRefresh={loadData}
              onSave={(yamlStr) => console.log('Save YAML:', yamlStr)}
              onCompare={() => console.log('Compare YAML')}
            />
          </div>
        )}
      </div>
    </div>
  );
};

export default ResourceDetailPage;
