/**
 * Service 详情页面组件
 * 专为 Service 设计，展示端点、端口映射、选择器匹配等信息
 */
import React from 'react';
import { StatusCards } from '../StatusCards';
import { LabelList } from '../LabelList';
import { EventTimeline } from '../EventTimeline';
import './ResourceDetail.css';

interface ServiceDetailProps {
  data: any;
  endpoints?: any;
  pods?: any[];
  events?: any[];
  onResourceClick?: (type: string, namespace: string, name: string) => void;
}

export const ServiceDetail: React.FC<ServiceDetailProps> = ({
  data,
  endpoints,
  pods = [],
  events = [],
  onResourceClick,
}) => {
  if (!data) return null;

  const { metadata, spec, status } = data;
  const ports = spec.ports || [];
  const selector = spec.selector || {};
  const clusterIP = spec.clusterIP;
  const externalIPs = spec.externalIPs || [];
  const loadBalancerIP = status?.loadBalancer?.ingress?.[0]?.ip;
  const loadBalancerHostname = status?.loadBalancer?.ingress?.[0]?.hostname;

  // Service 类型
  const type = spec.type || 'ClusterIP';
  const typeDisplay = {
    ClusterIP: '🔒 集群内部',
    NodePort: '🌐 节点端口',
    LoadBalancer: '☁️ 负载均衡',
    ExternalName: '🔗 外部名称',
  }[type] || type;

  // 状态卡片
  const statusCards = [
    {
      title: '类型',
      value: type,
      sub: typeDisplay,
    },
    {
      title: 'Cluster IP',
      value: clusterIP || 'None',
      sub: spec.clusterIP === 'None' ? 'Headless' : '标准',
    },
    {
      title: '端点',
      value: endpoints?.subsets?.reduce(
        (sum: number, s: any) => sum + (s.addresses?.length || 0),
        0
      ) || 0,
      sub: `端口：${ports.length}`,
    },
  ];

  // 渲染端口映射
  const renderPorts = () => {
    if (ports.length === 0) {
      return (
        <div className="empty-state">
          <span>未定义端口</span>
        </div>
      );
    }

    return (
      <div className="ports-table">
        <table className="resource-table">
          <thead>
            <tr>
              <th>名称</th>
              <th>协议</th>
              <th>端口</th>
              <th>目标端口</th>
              {type === 'NodePort' && <th>节点端口</th>}
            </tr>
          </thead>
          <tbody>
            {ports.map((port: any, idx: number) => (
              <tr key={idx}>
                <td>{port.name || '-'}</td>
                <td>{port.protocol || 'TCP'}</td>
                <td className="code">{port.port}</td>
                <td className="code">{port.targetPort || port.port}</td>
                {type === 'NodePort' && (
                  <td className="code">{port.nodePort || '-'}</td>
                )}
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    );
  };

  // 渲染端点
  const renderEndpoints = () => {
    const subsets = endpoints?.subsets || [];
    const addresses = subsets.flatMap((s: any) => s.addresses || []);
    const notReadyAddresses = subsets.flatMap((s: any) => s.notReadyAddresses || []);

    if (addresses.length === 0 && notReadyAddresses.length === 0) {
      return (
        <div className="empty-state warning">
          <span className="empty-icon">⚠️</span>
          <p>没有可用的端点</p>
          <p className="empty-desc">
            {Object.keys(selector).length > 0
              ? '可能是 Pod 选择器没有匹配到任何 Pod'
              : 'Service 没有定义选择器'}
          </p>
        </div>
      );
    }

    return (
      <div className="endpoints-section">
        {/* 就绪端点 */}
        {addresses.length > 0 && (
          <div className="endpoint-subset">
            <h4 className="subsection-title">✅ 就绪端点 ({addresses.length})</h4>
            <div className="endpoints-table">
              <table className="resource-table">
                <thead>
                  <tr>
                    <th>IP</th>
                    <th>端口</th>
                    <th>节点</th>
                    <th>目标引用</th>
                  </tr>
                </thead>
                <tbody>
                  {addresses.map((addr: any, idx: number) => (
                    <tr key={idx}>
                      <td className="code">{addr.ip}</td>
                      <td className="code">
                        {subsets[0]?.ports?.map((p: any) => p.port).join(', ') || '-'}
                      </td>
                      <td>{addr.nodeName || '-'}</td>
                      <td>
                        {addr.targetRef ? (
                          <span
                            className="link"
                            onClick={() =>
                              onResourceClick?.(
                                addr.targetRef.kind?.toLowerCase() || 'pod',
                                addr.targetRef.namespace || metadata.namespace,
                                addr.targetRef.name
                              )
                            }
                          >
                            {addr.targetRef.kind}:{addr.targetRef.name}
                          </span>
                        ) : (
                          '-'
                        )}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </div>
        )}

        {/* 未就绪端点 */}
        {notReadyAddresses.length > 0 && (
          <div className="endpoint-subset">
            <h4 className="subsection-title warning">⏳ 未就绪端点 ({notReadyAddresses.length})</h4>
            <div className="endpoints-table">
              <table className="resource-table">
                <thead>
                  <tr>
                    <th>IP</th>
                    <th>端口</th>
                    <th>节点</th>
                    <th>目标引用</th>
                  </tr>
                </thead>
                <tbody>
                  {notReadyAddresses.map((addr: any, idx: number) => (
                    <tr key={idx}>
                      <td className="code">{addr.ip}</td>
                      <td className="code">
                        {subsets[0]?.ports?.map((p: any) => p.port).join(', ') || '-'}
                      </td>
                      <td>{addr.nodeName || '-'}</td>
                      <td>
                        {addr.targetRef ? (
                          <span
                            className="link"
                            onClick={() =>
                              onResourceClick?.(
                                addr.targetRef.kind?.toLowerCase() || 'pod',
                                addr.targetRef.namespace || metadata.namespace,
                                addr.targetRef.name
                              )
                            }
                          >
                            {addr.targetRef.kind}:{addr.targetRef.name}
                          </span>
                        ) : (
                          '-'
                        )}
                      </td>
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </div>
        )}
      </div>
    );
  };

  // 渲染外部访问信息
  const renderExternalAccess = () => {
    const hasExternal =
      type === 'LoadBalancer' ||
      type === 'NodePort' ||
      externalIPs.length > 0 ||
      loadBalancerIP ||
      loadBalancerHostname;

    if (!hasExternal) return null;

    return (
      <div className="info-card">
        <h4 className="card-title">🌐 外部访问</h4>
        <div className="info-grid">
          {externalIPs.length > 0 && (
            <div className="info-item full-width">
              <span className="info-label">外部 IP</span>
              <span className="info-value code">
                {externalIPs.join(', ')}
              </span>
            </div>
          )}

          {loadBalancerIP && (
            <div className="info-item">
              <span className="info-label">负载均衡 IP</span>
              <span className="info-value code">{loadBalancerIP}</span>
            </div>
          )}

          {loadBalancerHostname && (
            <div className="info-item">
              <span className="info-label">负载均衡主机名</span>
              <span className="info-value code">{loadBalancerHostname}</span>
            </div>
          )}

          {type === 'NodePort' && (
            <div className="info-item full-width">
              <span className="info-label">访问地址</span>
              <span className="info-value code">
                {'<NodeIP>:'}
                {ports.map((p: any) => p.nodePort).join(', ')}
              </span>
            </div>
          )}

          {type === 'LoadBalancer' && (loadBalancerIP || loadBalancerHostname) && (
            <div className="info-item full-width">
              <span className="info-label">访问地址</span>
              <span className="info-value code">
                {loadBalancerIP || loadBalancerHostname}:
                {ports.map((p: any) => p.port).join(', ')}
              </span>
            </div>
          )}
        </div>
      </div>
    );
  };

  // 渲染 Pod 选择器匹配
  const renderSelectorMatch = () => {
    const selectorKeys = Object.keys(selector);

    if (selectorKeys.length === 0) {
      return (
        <div className="info-card">
          <h4 className="card-title">🎯 Pod 选择器</h4>
          <div className="selector-empty">
            <span className="empty-icon">⚠️</span>
            <p>此 Service 没有定义 Pod 选择器</p>
            <p className="empty-desc">
              可能是一个 ExternalName 服务或手动管理端点
            </p>
          </div>
        </div>
      );
    }

    return (
      <div className="info-card">
        <h4 className="card-title">🎯 Pod 选择器</h4>
        <div className="selector-labels">
          <LabelList labels={selector} />
        </div>
        <div className="match-info">
          <span className="match-count">
            匹配到 {pods.length} 个 Pod
          </span>
          {pods.length === 0 && (
            <span className="match-warning">
              ⚠️ 没有 Pod 匹配此选择器
            </span>
          )}
        </div>
      </div>
    );
  };

  // 渲染关联 Pod
  const renderPodList = () => {
    if (pods.length === 0) return null;

    return (
      <div className="pod-table">
        <table className="resource-table">
          <thead>
            <tr>
              <th>名称</th>
              <th>状态</th>
              <th>IP</th>
              <th>节点</th>
            </tr>
          </thead>
          <tbody>
            {pods.map((pod: any) => (
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
                <td className="code">{pod.status.podIP || '-'}</td>
                <td>{pod.spec.nodeName || '-'}</td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    );
  };

  return (
    <div className="service-detail">
      {/* 状态概览 */}
      <div className="detail-section">
        <h3 className="section-title">📊 状态概览</h3>
        <StatusCards cards={statusCards} />

        {/* 外部访问信息 */}
        {renderExternalAccess()}
      </div>

      {/* 端口配置 */}
      <div className="detail-section">
        <h3 className="section-title">🔌 端口配置</h3>
        {renderPorts()}
      </div>

      {/* 端点 */}
      <div className="detail-section">
        <h3 className="section-title">🎯 端点</h3>
        {renderEndpoints()}
      </div>

      {/* Pod 选择器 */}
      <div className="detail-section">
        <h3 className="section-title">🔗 选择器匹配</h3>
        {renderSelectorMatch()}
        {pods.length > 0 && renderPodList()}
      </div>

      {/* 会话保持 */}
      {spec.sessionAffinity && (
        <div className="detail-section">
          <h3 className="section-title">🔄 会话保持</h3>
          <div className="info-card">
            <div className="info-grid">
              <div className="info-item">
                <span className="info-label">会话亲和性</span>
                <span className="info-value">
                  {spec.sessionAffinity === 'ClientIP'
                    ? '基于客户端 IP'
                    : spec.sessionAffinity}
                </span>
              </div>
              {spec.sessionAffinityConfig?.clientIP?.timeoutSeconds && (
                <div className="info-item">
                  <span className="info-label">超时时间</span>
                  <span className="info-value">
                    {spec.sessionAffinityConfig.clientIP.timeoutSeconds} 秒
                  </span>
                </div>
              )}
            </div>
          </div>
        </div>
      )}

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
          </div>
        </div>
      </div>

      {/* 标签和注解 */}
      <div className="detail-section">
        <h3 className="section-title">🏷️ 标签和注解</h3>
        {metadata?.labels && Object.keys(metadata.labels).length > 0 && (
          <div className="labels-subsection">
            <h4 className="subsection-title">标签</h4>
            <LabelList labels={metadata.labels} />
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

export default ServiceDetail;
