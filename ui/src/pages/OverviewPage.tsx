import React, { useEffect, useState } from 'react';
import InfoCard from './InfoCard.tsx';
import ResourceSummary from './ResourceSummary.tsx';
import PageHeader from '../components/common/PageHeader.tsx';
import { OverviewPageProps, OverviewData, K8sEventSimple } from '../types';
import { apiClient } from '../utils/apiClient';
import { FaServer, FaCube, FaNetworkWired } from 'react-icons/fa';
import { FaThLarge } from 'react-icons/fa';

/**
 * 简化的 useFetch Hook
 */
function useFetch<T>(url: string) {
  const [data, setData] = useState<T | null>(null);
  const [loading, setLoading] = useState<boolean>(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const fetchData = async () => {
      try {
        const result = await apiClient.get<T>(url);
        setData(result.data || null);
      } catch (err) {
        setError(err instanceof Error ? err.message : 'Unknown error');
      } finally {
        setLoading(false);
      }
    };

    fetchData();
  }, [url]);

  return { data, loading, error };
}

export const OverviewPage: React.FC<OverviewPageProps> = ({ collapsed, onToggleCollapsed }) => {
  const { data, loading, error } = useFetch<OverviewData>('/api/overview');
  const safeData: Partial<OverviewData> = data || {};

  // 格式化时间
  const formatRelativeTime = (dateString: string) => {
    const date = new Date(dateString);
    const now = new Date();
    const diffMs = now.getTime() - date.getTime();
    const diffSec = Math.floor(diffMs / 1000);
    const diffMin = Math.floor(diffSec / 60);
    const diffHour = Math.floor(diffMin / 60);
    const diffDay = Math.floor(diffHour / 24);

    if (diffSec < 60) return 'Just now';
    if (diffMin < 60) return `${diffMin}m ago`;
    if (diffHour < 24) return `${diffHour}h ago`;
    return `${diffDay}d ago`;
  };

  if (loading) {
    return <div style={{ textAlign: 'center', color: '#888', padding: '32px 0' }}>加载中...</div>;
  }

  if (error) {
    return (
      <div style={{ textAlign: 'center', color: 'red', padding: '32px 0' }}>错误：{error}</div>
    );
  }

  return (
    <div>
      <PageHeader title="Overview" collapsed={collapsed} onToggleCollapsed={onToggleCollapsed} />

      <div className="overview-grid">
        <InfoCard
          icon={<FaServer />}
          title="Nodes"
          value={safeData.nodeCount || 0}
          status={
            safeData.nodeCount === 0 ? (
              <div className="center-empty">
                <span style={{ color: '#c0c4cc', fontSize: 'var(--font-size-sm)' }}>暂无数据</span>
              </div>
            ) : (
              <span
                className={
                  (safeData.nodeReady ?? 0) === (safeData.nodeCount ?? 0)
                    ? 'status-ready'
                    : 'status-failed'
                }
              >
                {(safeData.nodeReady ?? 0) === (safeData.nodeCount ?? 0)
                  ? 'All Ready'
                  : `${(safeData.nodeCount ?? 0) - (safeData.nodeReady ?? 0)} Not Ready`}
              </span>
            )
          }
        />
        <InfoCard
          icon={<FaCube />}
          title="Pods"
          value={safeData.podCount || 0}
          status={
            safeData.podCount === 0 ? (
              <div className="center-empty">
                <span style={{ color: '#c0c4cc', fontSize: 'var(--font-size-sm)' }}>暂无数据</span>
              </div>
            ) : (
              <span
                className={(safeData.podNotReady ?? 0) === 0 ? 'status-ready' : 'status-failed'}
              >
                {(safeData.podNotReady ?? 0) === 0
                  ? 'All Ready'
                  : `${safeData.podNotReady} Not Ready`}
              </span>
            )
          }
        />
        <InfoCard
          icon={<FaThLarge />}
          title="Namespaces"
          value={safeData.namespaceCount || 0}
          status={
            safeData.namespaceCount === 0 ? (
              <div className="center-empty">
                <span style={{ color: '#c0c4cc', fontSize: 'var(--font-size-sm)' }}>暂无数据</span>
              </div>
            ) : (
              <span className="status-ready">All Ready</span>
            )
          }
        />
        <InfoCard
          icon={<FaNetworkWired />}
          title="Services"
          value={safeData.serviceCount || 0}
          status={
            safeData.serviceCount === 0 ? (
              <div className="center-empty">
                <span style={{ color: '#c0c4cc', fontSize: 'var(--font-size-sm)' }}>暂无数据</span>
              </div>
            ) : (
              <span className="status-ready">All Ready</span>
            )
          }
        />
      </div>

      <div className="overview-row2">
        <div className="overview-left-col">
          <ResourceSummary
            title="CPU"
            requestsValue={(safeData.cpuRequests ?? 0).toFixed(1)}
            requestsPercent={(
              ((safeData.cpuRequests ?? 0) / (safeData.cpuCapacity ?? 1)) *
              100
            ).toFixed(1)}
            limitsValue={(safeData.cpuLimits ?? 0).toFixed(1)}
            limitsPercent={(
              ((safeData.cpuLimits ?? 0) / (safeData.cpuCapacity ?? 1)) *
              100
            ).toFixed(1)}
            totalValue={(safeData.cpuCapacity ?? 0).toFixed(1)}
            availableValue={((safeData.cpuCapacity ?? 0) - (safeData.cpuRequests ?? 0)).toFixed(1)}
            unit="cores"
          />
          <ResourceSummary
            title="Memory"
            requestsValue={(safeData.memoryRequests ?? 0).toFixed(1)}
            requestsPercent={(
              ((safeData.memoryRequests ?? 0) / (safeData.memoryCapacity ?? 1)) *
              100
            ).toFixed(1)}
            limitsValue={(safeData.memoryLimits ?? 0).toFixed(1)}
            limitsPercent={(
              ((safeData.memoryLimits ?? 0) / (safeData.memoryCapacity ?? 1)) *
              100
            ).toFixed(1)}
            totalValue={(safeData.memoryCapacity ?? 0).toFixed(1)}
            availableValue={(
              (safeData.memoryCapacity ?? 0) - (safeData.memoryRequests ?? 0)
            ).toFixed(1)}
            unit="GiB"
          />
        </div>

        <div className="overview-event-col">
          <div className="overview-event-card resource-summary-card">
            <div className="resource-summary-title">Recent Events</div>
            {safeData.events && safeData.events.length > 0 ? (
              safeData.events
                .slice()
                .sort(
                  (a: K8sEventSimple, b: K8sEventSimple) =>
                    new Date(b.lastSeen).getTime() - new Date(a.lastSeen).getTime()
                )
                .slice(0, 5)
                .map((e: K8sEventSimple, i: number) => (
                  <div
                    key={i}
                    style={{
                      display: 'flex',
                      alignItems: 'flex-start',
                      paddingBottom: '18px',
                      borderBottom: i !== 4 ? '1px solid #f0f0f0' : 'none',
                      marginBottom: i !== 4 ? 12 : 0,
                    }}
                  >
                    <div
                      style={{
                        width: 28,
                        display: 'flex',
                        justifyContent: 'center',
                        alignItems: 'flex-start',
                        marginTop: 2,
                      }}
                    >
                      <span
                        style={{
                          display: 'inline-block',
                          width: 16,
                          height: 16,
                          borderRadius: '50%',
                          border: '2px solid #a5b4fc',
                          background: '#fff',
                          marginTop: 2,
                        }}
                      ></span>
                    </div>
                    <div style={{ flex: 1 }}>
                      <div style={{ display: 'flex', alignItems: 'center', marginBottom: 2 }}>
                        <span
                          className={
                            e.type === 'Warning' ? 'event-type-warning' : 'event-type-normal'
                          }
                          style={{
                            background: e.type === 'Warning' ? '#ffeaea' : '#e6f7ff',
                            color: e.type === 'Warning' ? '#ff4d4f' : '#1890ff',
                            borderRadius: '10px',
                            padding: '2px 10px',
                            fontSize: 'var(--font-size-sm)',
                            fontWeight: 600,
                            marginRight: 8,
                          }}
                        >
                          {e.type}
                        </span>
                        <span
                          style={{
                            fontWeight: 600,
                            fontSize: 'var(--font-size-sm)',
                            color: '#222',
                            marginRight: 8,
                          }}
                        >
                          {e.reason}
                        </span>
                        <span
                          style={{
                            marginLeft: 'auto',
                            fontSize: 'var(--font-size-sm)',
                            color: '#888',
                            fontWeight: 400,
                          }}
                        >
                          {formatRelativeTime(e.lastSeen)}
                        </span>
                      </div>
                      <div
                        style={{
                          fontSize: 'var(--font-size-sm)',
                          color: '#444',
                          marginBottom: 2,
                          wordBreak: 'break-all',
                        }}
                      >
                        {e.message}
                      </div>
                      <div
                        style={{ fontSize: 'var(--font-size-sm)', color: '#888', fontWeight: 400 }}
                      >
                        {e.pod ? `Pod: ${e.pod}` : ''}
                        {e.cloneset ? `CloneSet: ${e.cloneset}` : ''}
                        {e.namespace && e.name ? `Pod: ${e.namespace}/${e.name}` : ''}
                        {e.reporter ? ` Reporter: ${e.reporter}` : ''}
                      </div>
                    </div>
                  </div>
                ))
            ) : (
              <div
                style={{
                  color: '#888',
                  fontSize: 'var(--font-size-sm)',
                  padding: '24px 0',
                  textAlign: 'center',
                }}
              >
                暂无数据
              </div>
            )}
          </div>
        </div>
      </div>
    </div>
  );
};

export default OverviewPage;
