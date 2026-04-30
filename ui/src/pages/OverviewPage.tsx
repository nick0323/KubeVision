import React from 'react';
import InfoCard from './InfoCard.tsx';
import ResourceSummary from './ResourceSummary.tsx';
import EventCard from './EventCard.tsx';
import PageHeader from '../common/PageHeader.tsx';
import { OverviewPageProps, OverviewData } from '../types';
import { apiClient } from '../utils/apiClient';
import { FaServer, FaCube, FaNetworkWired } from 'react-icons/fa';
import { FaThLarge } from 'react-icons/fa';

/**
 * 简化的 useFetch Hook
 */
function useFetch<T>(url: string) {
  const [data, setData] = React.useState<T | null>(null);
  const [loading, setLoading] = React.useState<boolean>(true);
  const [error, setError] = React.useState<string | null>(null);

  React.useEffect(() => {
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

/**
 * 集群概览页面
 */
export const OverviewPage: React.FC<OverviewPageProps> = ({ collapsed, onToggleCollapsed }) => {
  const { data, loading, error } = useFetch<OverviewData>('/api/overview');
  const safeData: Partial<OverviewData> = data || {};

  if (loading) {
    return <div className="overview-loading">加载中...</div>;
  }

  if (error) {
    return <div className="overview-error">错误：{error}</div>;
  }

  return (
    <div className="overview-page">
      <PageHeader title="Overview" collapsed={collapsed} onToggleCollapsed={onToggleCollapsed} />

      {/* 核心资源统计 - 4个卡片 */}
      <div className="overview-stats-grid">
        <InfoCard
          icon={<FaServer />}
          title="Nodes"
          value={safeData.nodeCount || 0}
          status={
            safeData.nodeCount === 0 ? (
              <span className="status-empty">No data</span>
            ) : (safeData.nodeReadyCount ?? 0) === (safeData.nodeCount ?? 0) ? (
              <span className="status-success">All Ready ({safeData.nodeReadyCount})</span>
            ) : (
              <span className="status-warning">
                {safeData.nodeReadyCount}/{safeData.nodeCount} Ready
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
              <span className="status-empty">No data</span>
            ) : (safeData.podNotReady ?? 0) === 0 ? (
              <span className="status-success">All Ready</span>
            ) : (
              <span className="status-warning">
                {safeData.podNotReady} Not Ready
              </span>
            )
          }
        />
        <InfoCard
          icon={<FaThLarge />}
          title="Namespaces"
          value={safeData.namespaceCount || 0}
          status={<span className="status-success">Available</span>}
        />
        <InfoCard
          icon={<FaNetworkWired />}
          title="Services"
          value={safeData.serviceCount || 0}
          status={<span className="status-success">Available</span>}
        />
      </div>

      {/* 资源使用情况 - CPU和Memory并排，事件在下方 */}
      <div className="overview-resources-section">
        <div className="overview-resources-grid">
          <ResourceSummary
            title="CPU"
            requestsValue={(safeData.cpuRequests ?? 0).toFixed(1)}
            requestsPercent={
              ((safeData.cpuRequests ?? 0) / (safeData.cpuCapacity ?? 1)) * 100
            }
            limitsValue={(safeData.cpuLimits ?? 0).toFixed(1)}
            limitsPercent={
              ((safeData.cpuLimits ?? 0) / (safeData.cpuCapacity ?? 1)) * 100
            }
            totalValue={(safeData.cpuCapacity ?? 0).toFixed(1)}
            availableValue={
              ((safeData.cpuCapacity ?? 0) - (safeData.cpuRequests ?? 0)).toFixed(1)
            }
            unit="cores"
          />
          <ResourceSummary
            title="Memory"
            requestsValue={(safeData.memoryRequests ?? 0).toFixed(1)}
            requestsPercent={
              ((safeData.memoryRequests ?? 0) / (safeData.memoryCapacity ?? 1)) * 100
            }
            limitsValue={(safeData.memoryLimits ?? 0).toFixed(1)}
            limitsPercent={
              ((safeData.memoryLimits ?? 0) / (safeData.memoryCapacity ?? 1)) * 100
            }
            totalValue={(safeData.memoryCapacity ?? 0).toFixed(1)}
            availableValue={
              ((safeData.memoryCapacity ?? 0) - (safeData.memoryRequests ?? 0)).toFixed(1)
            }
            unit="GiB"
          />
        </div>
      </div>

      {/* 事件列表 - 占满整行 */}
      <div className="overview-events-section">
        <EventCard events={safeData.events || []} limit={5} />
      </div>
    </div>
  );
};

export default OverviewPage;
