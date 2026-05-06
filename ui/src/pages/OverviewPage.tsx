import React from 'react';
import InfoCard from './InfoCard.tsx';
import ResourceSummary from './ResourceSummary.tsx';
import EventCard from './EventCard.tsx';
import PageHeader from '../common/PageHeader.tsx';
import RefreshButton from '../common/RefreshButton.tsx';
import { OverviewPageProps, OverviewData } from '../types';
import { apiClient } from '../utils/apiClient';
import { FaServer, FaCube, FaNetworkWired } from 'react-icons/fa';
import { FaThLarge } from 'react-icons/fa';

/**
 * 简化's useFetch Hook（SupportmanualRefresh）
 */
function useFetch<T>(url: string) {
  const [data, setData] = React.useState<T | null>(null);
  const [loading, setLoading] = React.useState<boolean>(true);
  const [error, setError] = React.useState<string | null>(null);
  const [refreshKey, setRefreshKey] = React.useState<number>(0);

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
  }, [url, refreshKey]);

  const refresh = React.useCallback(() => {
    setLoading(true);
    setRefreshKey(prev => prev + 1);
  }, []);

  return { data, loading, error, refresh };
}

/**
 * clusterOverview page
 */
export const OverviewPage: React.FC<OverviewPageProps> = ({ collapsed, onToggleCollapsed }) => {
  const { data, loading, error, refresh } = useFetch<OverviewData>('/api/overview');
  const safeData: Partial<OverviewData> = data || {};

  if (loading) {
    return <div className="overview-loading">Loading....</div>;
  }

  if (error) {
    return <div className="overview-error">Error: {error}</div>;
  }

  return (
    <div className="overview-page">
      <PageHeader title="Overview" collapsed={collapsed} onToggleCollapsed={onToggleCollapsed}>
        <RefreshButton onClick={refresh} loading={loading} />
      </PageHeader>

      {/* Core resource statistics - 4 cards */}
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

      {/* Resource usage - CPU and Memory side by side, events below */}
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

      {/* Event list - full width */}
      <div className="overview-events-section">
        <EventCard events={safeData.events || []} limit={5} />
      </div>
    </div>
  );
};

export default OverviewPage;
