import React, { useState, useCallback, useEffect, useMemo } from 'react';
import PageHeader from '../common/PageHeader';
import SearchInput from '../common/SearchInput';
import RefreshButton from '../common/RefreshButton';
import { LoadingSpinner } from '../common/LoadingSpinner';
import { ErrorDisplay } from '../common/ErrorDisplay';
import { apiClient } from '../utils/apiClient';
import { notification } from '../common/Notification';
import { ArgoCDApplication } from '../types/argocd';
import { FaCheckCircle, FaExclamationTriangle, FaSync, FaHourglassHalf } from 'react-icons/fa';
import '../styles/argocd-page.css';

/**
 * ArgoCD 应用管理页面
 */
export const ArgoCDPage: React.FC<{ collapsed: boolean; onToggleCollapsed: () => void }> = ({
  collapsed,
  onToggleCollapsed,
}) => {
  const [apps, setApps] = useState<ArgoCDApplication[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [refreshKey, setRefreshKey] = useState(0);
  const [actionLoading, setActionLoading] = useState<Record<string, boolean>>({});
  const [searchQuery, setSearchQuery] = useState('');

  const fetchApps = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const result = await apiClient.get<{ data: ArgoCDApplication[] }>('/api/argocd/apps');
      setApps(result.data || []);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load ArgoCD applications');
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchApps();
  }, [fetchApps, refreshKey]);

  const handleRefresh = () => {
    setRefreshKey(prev => prev + 1);
  };

  // 搜索过滤
  const filteredApps = useMemo(() => {
    if (!searchQuery.trim()) return apps;
    const query = searchQuery.toLowerCase();
    return apps.filter(
      app =>
        app.metadata.name.toLowerCase().includes(query) ||
        app.spec.project.toLowerCase().includes(query) ||
        app.spec.source.repoURL.toLowerCase().includes(query) ||
        app.status.sync.status.toLowerCase().includes(query) ||
        app.status.health.status.toLowerCase().includes(query)
    );
  }, [apps, searchQuery]);

  const handleSync = async (name: string) => {
    setActionLoading(prev => ({ ...prev, [`${name}-sync`]: true }));
    try {
      await apiClient.post(`/api/argocd/apps/${name}/sync`);
      notification.success(`Application "${name}" sync triggered`);
      handleRefresh();
    } catch (err) {
      notification.error(`Sync failed: ${err instanceof Error ? err.message : 'Unknown error'}`);
    } finally {
      setActionLoading(prev => ({ ...prev, [`${name}-sync`]: false }));
    }
  };

  const handleRefreshApp = async (name: string) => {
    setActionLoading(prev => ({ ...prev, [`${name}-refresh`]: true }));
    try {
      await apiClient.post(`/api/argocd/apps/${name}/refresh`);
      notification.success(`Application "${name}" refresh triggered`);
      handleRefresh();
    } catch (err) {
      notification.error(`Refresh failed: ${err instanceof Error ? err.message : 'Unknown error'}`);
    } finally {
      setActionLoading(prev => ({ ...prev, [`${name}-refresh`]: false }));
    }
  };

  const handleDelete = async (name: string) => {
    if (!window.confirm(`Are you sure you want to delete application "${name}"?`)) {
      return;
    }
    setActionLoading(prev => ({ ...prev, [`${name}-delete`]: true }));
    try {
      await apiClient.delete(`/api/argocd/apps/${name}`);
      notification.success(`Application "${name}" deleted`);
      handleRefresh();
    } catch (err) {
      notification.error(`Delete failed: ${err instanceof Error ? err.message : 'Unknown error'}`);
    } finally {
      setActionLoading(prev => ({ ...prev, [`${name}-delete`]: false }));
    }
  };

  const getStatusColor = (status?: string) => {
    switch (status) {
      case 'Synced':
      case 'Healthy':
        return '#52c41a'; // --success
      case 'OutOfSync':
        return '#faad14'; // --warning
      case 'Degraded':
      case 'Failed':
        return '#ff4d4f'; // --danger
      case 'Progressing':
        return '#1890ff'; // --primary
      default:
        return '#666'; // --text-tertiary
    }
  };

  const getStatusBgColor = (status?: string) => {
    switch (status) {
      case 'Synced':
      case 'Healthy':
        return 'rgba(82, 196, 26, 0.1)'; // --bg-success
      case 'OutOfSync':
        return 'rgba(250, 173, 20, 0.1)'; // --bg-warning
      case 'Degraded':
      case 'Failed':
        return 'rgba(245, 34, 45, 0.1)'; // --bg-danger
      case 'Progressing':
        return 'rgba(24, 144, 255, 0.1)'; // --bg-info
      default:
        return 'rgba(102, 102, 102, 0.1)';
    }
  };

  const getStatusIcon = (status?: string) => {
    switch (status) {
      case 'Synced':
      case 'Healthy':
        return <FaCheckCircle />;
      case 'OutOfSync':
        return <FaExclamationTriangle />;
      case 'Degraded':
      case 'Failed':
        return <FaExclamationTriangle />;
      case 'Progressing':
        return <FaSync className="spinning" />;
      default:
        return <FaHourglassHalf />;
    }
  };

  const breadcrumbs = [
    { label: 'ArgoCD', path: 'argocd' },
  ];

  if (loading) {
    return (
      <div className="argocd-page">
        <LoadingSpinner text="Loading ArgoCD applications..." size="lg" />
      </div>
    );
  }

  if (error) {
    return (
      <div className="argocd-page">
        <ErrorDisplay message={error} type="error" showRetry onRetry={handleRefresh} />
      </div>
    );
  }

  return (
    <div className="argocd-page">
      <PageHeader
        title="ArgoCD Applications"
        breadcrumbs={breadcrumbs}
        onBreadcrumbClick={() => {}}
        collapsed={collapsed}
        onToggleCollapsed={onToggleCollapsed}
      >
        <div className="argocd-header-actions">
          <SearchInput
            placeholder="Search applications..."
            value={searchQuery}
            onChange={e => setSearchQuery(e.target.value)}
            onClear={() => setSearchQuery('')}
            showClearButton={true}
          />
          <RefreshButton onClick={handleRefresh} loading={loading} />
        </div>
      </PageHeader>

      <div className="argocd-content">
        {filteredApps.length === 0 ? (
          <div className="empty-state">
            <p className="empty-state-text">
              {searchQuery ? 'No applications match your search' : 'No ArgoCD applications found'}
            </p>
            {!searchQuery && (
              <p className="empty-state-hint">
                Configure ArgoCD connection in config.yaml to manage GitOps applications
              </p>
            )}
          </div>
        ) : (
          <div className="apps-grid">
            {filteredApps.map(app => (
              <div key={app.metadata.uid} className="app-card">
                <div className="app-card-header">
                  <h3 className="app-name">{app.metadata.name}</h3>
                  <div className="app-status-badges">
                    <span
                      className="status-badge"
                      style={{
                        backgroundColor: getStatusBgColor(app.status.sync.status),
                        color: getStatusColor(app.status.sync.status)
                      }}
                      title={`Sync: ${app.status.sync.status}`}
                    >
                      {getStatusIcon(app.status.sync.status)}
                      {app.status.sync.status}
                    </span>
                    <span
                      className="status-badge"
                      style={{
                        backgroundColor: getStatusBgColor(app.status.health.status),
                        color: getStatusColor(app.status.health.status)
                      }}
                      title={`Health: ${app.status.health.status}`}
                    >
                      {getStatusIcon(app.status.health.status)}
                      {app.status.health.status}
                    </span>
                  </div>
                </div>

                <div className="app-card-body">
                  <div className="app-info-row">
                    <span className="app-info-label">Project:</span>
                    <span className="app-info-value">{app.spec.project}</span>
                  </div>
                  <div className="app-info-row">
                    <span className="app-info-label">Repo:</span>
                    <span className="app-info-value" title={app.spec.source.repoURL}>
                      {app.spec.source.repoURL}
                    </span>
                  </div>
                  <div className="app-info-row">
                    <span className="app-info-label">Target:</span>
                    <span className="app-info-value">
                      {app.spec.destination.namespace}
                    </span>
                  </div>
                  {app.status.operationState && (
                    <div className="app-info-row">
                      <span className="app-info-label">Operation:</span>
                      <span
                        className="app-info-value operation-badge"
                        style={{
                          backgroundColor:
                            app.status.operationState.phase === 'Succeeded'
                              ? 'rgba(82, 196, 26, 0.1)' // --bg-success
                              : app.status.operationState.phase === 'Failed' ||
                                  app.status.operationState.phase === 'Error'
                                ? 'rgba(245, 34, 45, 0.1)' // --bg-danger
                                : 'rgba(24, 144, 255, 0.1)', // --bg-info
                          color:
                            app.status.operationState.phase === 'Succeeded'
                              ? '#52c41a' // --success
                              : app.status.operationState.phase === 'Failed' ||
                                  app.status.operationState.phase === 'Error'
                                ? '#ff4d4f' // --danger
                                : '#1890ff', // --primary
                        }}
                      >
                        {app.status.operationState.phase}
                      </span>
                    </div>
                  )}
                </div>

                <div className="app-card-actions">
                  <button
                    className="action-btn sync-btn"
                    onClick={() => handleSync(app.metadata.name)}
                    disabled={actionLoading[`${app.metadata.name}-sync`]}
                  >
                    {actionLoading[`${app.metadata.name}-sync`] ? 'Syncing...' : 'Sync'}
                  </button>
                  <button
                    className="action-btn refresh-btn"
                    onClick={() => handleRefreshApp(app.metadata.name)}
                    disabled={actionLoading[`${app.metadata.name}-refresh`]}
                  >
                    {actionLoading[`${app.metadata.name}-refresh`] ? 'Refreshing...' : 'Refresh'}
                  </button>
                  <button
                    className="action-btn delete-btn"
                    onClick={() => handleDelete(app.metadata.name)}
                    disabled={actionLoading[`${app.metadata.name}-delete`]}
                  >
                    {actionLoading[`${app.metadata.name}-delete`] ? 'Deleting...' : 'Delete'}
                  </button>
                </div>
              </div>
            ))}
          </div>
        )}
      </div>
    </div>
  );
};

export default ArgoCDPage;
