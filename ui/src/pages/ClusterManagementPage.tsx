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
import '../pages/ResourceListPage.css';
import './ClusterManagementPage.css';

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
              <button
                className="btn btn-secondary"
                onClick={() => setShowAddForm(false)}
              >
                Cancel
              </button>
              <button
                className="btn btn-outline"
                onClick={handleTestConnection}
                disabled={testing || (!form.apiServer && !form.kubeconfig)}
              >
                {testing ? 'Testing...' : 'Test Connection'}
              </button>
              <button
                className="btn btn-primary"
                onClick={handleAddCluster}
                disabled={saving || !form.name.trim()}
              >
                {saving ? 'Adding...' : 'Add Cluster'}
              </button>
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
        <div className="table-container table-container-mt">
          <table className="resource-table">
            <thead>
              <tr>
                <th className="col-w-15">Name</th>
                <th className="col-w-32">API Server</th>
                <th className="col-w-11">Status</th>
                <th className="col-w-10">Version</th>
                <th className="col-w-8">Nodes</th>
                <th className="col-w-18">Last Check</th>
                <th className="col-w-6">Actions</th>
              </tr>
            </thead>
            <tbody>
              {filteredClusters.map(cluster => (
                <tr key={cluster.name} className={cluster.name === 'default' ? 'default-row' : ''}>
                  <td>
                    <span className="resource-name-link">{cluster.name}</span>
                    {cluster.name === 'default' && (
                      <span className="default-badge">default</span>
                    )}
                  </td>
                  <td className="cell-api-server">{cluster.apiServer || '-'}</td>
                  <td>
                    <span className={`status-pill ${cluster.healthy ? 'healthy' : 'unhealthy'}`}>
                      {cluster.healthy ? <FaCheck size={12} /> : <FaTimes size={12} />}
                      <span className="status-text">{cluster.healthy ? 'Healthy' : 'Unhealthy'}</span>
                    </span>
                  </td>
                  <td>{cluster.version || '-'}</td>
                  <td>{cluster.nodeCount}</td>
                  <td className="cell-last-check">{formatTime(cluster.lastCheck)}</td>
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
    </div>
  );
};

export default ClusterManagementPage;
