import React, { useState, useCallback, useEffect, useRef } from 'react';
import PageHeader from '../common/PageHeader';
import SearchInput from '../common/SearchInput';
import RefreshButton from '../common/RefreshButton';
import { LoadingSpinner } from '../common/LoadingSpinner';
import { ErrorDisplay } from '../common/ErrorDisplay';
import { apiClient } from '../utils/apiClient';
import { notification } from '../common/NotificationContext';
import { usePageTitle } from '../hooks/usePageTitle';
import { ConfirmModal } from '../common/ConfirmModal';
import { useConfirm } from '../hooks/useConfirm';
import { FaCheck, FaTimes, FaPlus, FaTrash, FaServer, FaWifi } from 'react-icons/fa';
import '../styles/argocd-page.css';

interface ClusterItem {
  name: string;
  apiServer: string;
  version: string;
  healthy: boolean;
  nodeCount: number;
  lastCheck: number;
}

interface AddClusterForm {
  name: string;
  apiServer: string;
  token: string;
  kubeconfig: string;
  caFile: string;
  insecure: boolean;
}

const initialForm: AddClusterForm = {
  name: '',
  apiServer: '',
  token: '',
  kubeconfig: '',
  caFile: '',
  insecure: false,
};

export const ClusterManagementPage: React.FC<{
  collapsed: boolean;
  onToggleCollapsed: () => void;
}> = ({ collapsed, onToggleCollapsed }) => {
  usePageTitle('Clusters');
  const [clusters, setClusters] = useState<ClusterItem[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [refreshKey, setRefreshKey] = useState(0);
  const [showAddForm, setShowAddForm] = useState(false);
  const [form, setForm] = useState<AddClusterForm>(initialForm);
  const [testing, setTesting] = useState(false);
  const [saving, setSaving] = useState(false);
  const { confirm, confirming, config, onConfirm, onCancel } = useConfirm();
  const [deleteTarget, setDeleteTarget] = useState<string | null>(null);
  const [searchQuery, setSearchQuery] = useState('');

  const fetchClusters = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const result = await apiClient.get<ClusterItem[]>('/api/v1/clusters');
      setClusters(result.data || []);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load clusters');
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    fetchClusters();
  }, [fetchClusters, refreshKey]);

  const handleRefresh = () => {
    setRefreshKey(prev => prev + 1);
  };

  const filteredClusters = searchQuery.trim()
    ? clusters.filter(c =>
        c.name.toLowerCase().includes(searchQuery.toLowerCase()) ||
        c.apiServer.toLowerCase().includes(searchQuery.toLowerCase()) ||
        c.version.toLowerCase().includes(searchQuery.toLowerCase())
      )
    : clusters;

  const handleInputChange = (field: keyof AddClusterForm, value: string | boolean) => {
    setForm(prev => ({ ...prev, [field]: value }));
  };

  const handleTestConnection = async () => {
    setTesting(true);
    try {
      await apiClient.post('/api/v1/clusters/test', {
        apiServer: form.apiServer || undefined,
        token: form.token || undefined,
        kubeconfig: form.kubeconfig || undefined,
        caFile: form.caFile || undefined,
        insecure: form.insecure,
      });
      notification.success('Connection successful');
    } catch (err) {
      notification.error(`Connection failed: ${err instanceof Error ? err.message : 'Unknown error'}`);
    } finally {
      setTesting(false);
    }
  };

  const handleAddCluster = async () => {
    if (!form.name.trim()) {
      notification.warning('Cluster name is required');
      return;
    }
    if (!form.apiServer.trim() && !form.kubeconfig.trim()) {
      notification.warning('API server URL or kubeconfig path is required');
      return;
    }

    setSaving(true);
    try {
      await apiClient.post('/api/v1/clusters', {
        name: form.name.trim(),
        apiServer: form.apiServer || undefined,
        token: form.token || undefined,
        kubeconfig: form.kubeconfig || undefined,
        caFile: form.caFile || undefined,
        insecure: form.insecure,
      });
      notification.success(`Cluster "${form.name}" added`);
      setShowAddForm(false);
      setForm(initialForm);
      handleRefresh();
    } catch (err) {
      notification.error(`Failed to add cluster: ${err instanceof Error ? err.message : 'Unknown error'}`);
    } finally {
      setSaving(false);
    }
  };

  const handleDeleteCluster = async (name: string) => {
    const result = await confirm({
      title: 'Delete Cluster',
      message: `Are you sure you want to remove cluster "${name}"?`,
      confirmText: 'Delete',
      danger: true,
    });
    if (!result.confirmed) return;

    try {
      await apiClient.delete(`/api/v1/clusters/${name}`);
      notification.success(`Cluster "${name}" removed`);
      handleRefresh();
    } catch (err) {
      notification.error(`Failed to remove cluster: ${err instanceof Error ? err.message : 'Unknown error'}`);
    }
  };

  const formatTime = (ts: number): string => {
    if (!ts) return 'Never';
    const d = new Date(ts * 1000);
    return d.toLocaleString();
  };

  if (loading && clusters.length === 0) {
    return (
      <div className="argocd-page">
        <PageHeader title="Cluster Management" collapsed={collapsed} onToggleCollapsed={onToggleCollapsed}>
          <RefreshButton onClick={handleRefresh} loading />
        </PageHeader>
        <LoadingSpinner text="Loading clusters..." size="lg" />
      </div>
    );
  }

  if (error && clusters.length === 0) {
    return (
      <div className="argocd-page">
        <PageHeader title="Cluster Management" collapsed={collapsed} onToggleCollapsed={onToggleCollapsed}>
          <RefreshButton onClick={handleRefresh} loading={loading} />
        </PageHeader>
        <ErrorDisplay message={error} type="error" showRetry onRetry={handleRefresh} />
      </div>
    );
  }

  return (
    <div className="argocd-page">
      <PageHeader title="Cluster Management" collapsed={collapsed} onToggleCollapsed={onToggleCollapsed}>
        <button
          className="create-resource-btn"
          onClick={() => {
            setShowAddForm(true);
            setForm(initialForm);
          }}
        >
          <FaPlus /> Add Cluster
        </button>
        <SearchInput
          placeholder="Search clusters..."
          value={searchQuery}
          onChange={(e: React.ChangeEvent<HTMLInputElement>) => setSearchQuery(e.target.value)}
          onSubmit={() => {}}
          onClear={() => setSearchQuery('')}
        />
        <RefreshButton onClick={handleRefresh} loading={loading || clusters.length === 0} />
      </PageHeader>

      {showAddForm && (
        <div className="cluster-form-overlay" onClick={() => setShowAddForm(false)}>
          <div className="cluster-form" onClick={e => e.stopPropagation()}>
            <h3>Add Cluster</h3>
            <div className="form-field">
              <label>Cluster Name *</label>
              <input
                type="text"
                value={form.name}
                onChange={e => handleInputChange('name', e.target.value)}
                placeholder="my-cluster"
              />
            </div>
            <div className="form-field">
              <label>API Server URL</label>
              <input
                type="text"
                value={form.apiServer}
                onChange={e => handleInputChange('apiServer', e.target.value)}
                placeholder="https://kubernetes.example.com:6443"
              />
            </div>
            <div className="form-field">
              <label>Token</label>
              <input
                type="password"
                value={form.token}
                onChange={e => handleInputChange('token', e.target.value)}
                placeholder="Service account token (if using token auth)"
              />
            </div>
            <div className="form-field">
              <label>Kubeconfig Path</label>
              <input
                type="text"
                value={form.kubeconfig}
                onChange={e => handleInputChange('kubeconfig', e.target.value)}
                placeholder="/path/to/kubeconfig.yaml"
              />
            </div>
            <div className="form-field">
              <label>CA File Path</label>
              <input
                type="text"
                value={form.caFile}
                onChange={e => handleInputChange('caFile', e.target.value)}
                placeholder="/path/to/ca.crt"
              />
            </div>
            <div className="form-field checkbox-field">
              <label>
                <input
                  type="checkbox"
                  checked={form.insecure}
                  onChange={e => handleInputChange('insecure', e.target.checked)}
                />
                Skip TLS verification (insecure)
              </label>
            </div>
            <div className="form-buttons">
              <div>
                <button
                  className="action-btn"
                  onClick={handleTestConnection}
                  disabled={testing || (!form.apiServer && !form.kubeconfig)}
                >
                  {testing ? 'Testing...' : 'Test Connection'}
                </button>
              </div>
              <div>
                <button
                  className="create-resource-btn"
                  onClick={handleAddCluster}
                  disabled={saving || !form.name.trim()}
                >
                  {saving ? 'Adding...' : 'Add Cluster'}
                </button>
                <button className="action-btn" onClick={() => setShowAddForm(false)} style={{ marginLeft: 8 }}>
                  Cancel
                </button>
              </div>
            </div>
          </div>
        </div>
      )}

      {filteredClusters.length === 0 ? (
        <div className="resource-empty">
          <div className="resource-empty-icon"><FaServer /></div>
          <p className="resource-empty-text">
            {searchQuery ? 'No clusters match your search' : 'No clusters configured'}
          </p>
          {!searchQuery && (
            <button className="create-resource-btn" onClick={() => setShowAddForm(true)}>
              Add your first cluster
            </button>
          )}
        </div>
      ) : (
        <div className="table-container" style={{ marginTop: 16 }}>
          <table className="resource-table">
            <thead>
              <tr>
                <th style={{ width: '18%' }}>Name</th>
                <th style={{ width: '28%' }}>API Server</th>
                <th style={{ width: '12%' }}>Status</th>
                <th style={{ width: '12%' }}>Version</th>
                <th style={{ width: '10%' }}>Nodes</th>
                <th style={{ width: '14%' }}>Last Check</th>
                <th style={{ width: '6%' }}>Actions</th>
              </tr>
            </thead>
            <tbody>
              {filteredClusters.map(cluster => (
                <tr key={cluster.name} className={cluster.name === 'default' ? 'default-row' : ''}>
                  <td>
                    <span className="resource-name-link">{cluster.name}</span>
                    {cluster.name === 'default' && (
                      <span className="status-badge" style={{ marginLeft: 8, fontSize: 11, padding: '2px 6px', background: '#e8f5e9', color: '#2e7d32' }}>default</span>
                    )}
                  </td>
                  <td style={{ fontSize: 13, color: '#666' }}>{cluster.apiServer || '-'}</td>
                  <td>
                    <span className={`status-pill ${cluster.healthy ? 'healthy' : 'unhealthy'}`}>
                      {cluster.healthy ? <FaCheck size={12} /> : <FaTimes size={12} />}
                      <span style={{ marginLeft: 4 }}>{cluster.healthy ? 'Healthy' : 'Unhealthy'}</span>
                    </span>
                  </td>
                  <td>{cluster.version || '-'}</td>
                  <td>{cluster.nodeCount}</td>
                  <td style={{ fontSize: 13, color: '#888' }}>{formatTime(cluster.lastCheck)}</td>
                  <td className="actions-cell">
                    {cluster.name !== 'default' && (
                      <button
                        className="actions-btn danger"
                        onClick={() => handleDeleteCluster(cluster.name)}
                        title="Remove cluster"
                      >
                        <FaTrash />
                      </button>
                    )}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      )}

      <ConfirmModal
        open={confirming && !!config}
        title={config?.title}
        message={config?.message || ''}
        confirmText={config?.confirmText}
        cancelText={config?.cancelText}
        danger={config?.danger}
        onConfirm={onConfirm}
        onCancel={onCancel}
      />

      <style>{`
        .cluster-form-overlay {
          position: fixed;
          top: 0; left: 0; right: 0; bottom: 0;
          background: rgba(0,0,0,0.4);
          z-index: 1000;
          display: flex;
          align-items: center;
          justify-content: center;
        }
        .cluster-form {
          background: #fff;
          border-radius: 8px;
          padding: 24px;
          min-width: 480px;
          max-width: 600px;
          box-shadow: 0 8px 32px rgba(0,0,0,0.2);
        }
        .cluster-form h3 {
          margin: 0 0 16px;
          font-size: 18px;
          color: #1a1a2e;
        }
        .form-field {
          margin-bottom: 12px;
        }
        .form-field label {
          display: block;
          margin-bottom: 4px;
          font-size: 13px;
          font-weight: 500;
          color: #555;
        }
        .form-field input[type="text"],
        .form-field input[type="password"] {
          width: 100%;
          padding: 8px 12px;
          border: 1px solid #ddd;
          border-radius: 4px;
          font-size: 14px;
          box-sizing: border-box;
        }
        .checkbox-field label {
          display: flex;
          align-items: center;
          gap: 8px;
          cursor: pointer;
        }
        .checkbox-field input[type="checkbox"] {
          width: 16px;
          height: 16px;
        }
        .form-buttons {
          display: flex;
          justify-content: space-between;
          align-items: center;
          margin-top: 16px;
        }
        .status-pill {
          display: inline-flex;
          align-items: center;
          padding: 4px 10px;
          border-radius: 12px;
          font-size: 12px;
          font-weight: 500;
        }
        .status-pill.healthy {
          background: #e8f5e9;
          color: #2e7d32;
        }
        .status-pill.unhealthy {
          background: #fbe9e7;
          color: #c62828;
        }
        .default-row {
          background: #fafafa;
        }
        .actions-cell {
          text-align: center;
        }
        .actions-btn {
          width: 32px;
          height: 32px;
          display: inline-flex;
          align-items: center;
          justify-content: center;
          border: none;
          border-radius: 6px;
          background: transparent;
          color: #999;
          cursor: pointer;
          font-size: 16px;
          transition: all 0.2s ease;
        }
        .actions-btn.danger:hover {
          background: #fbe9e7;
          color: #c62828;
        }
      `}</style>
    </div>
  );
};

export default ClusterManagementPage;
