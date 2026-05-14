import React, { useState, useCallback, useEffect, useMemo } from 'react';
import jsyaml from 'js-yaml';
import PageHeader from '../common/PageHeader';
import SearchInput from '../common/SearchInput';
import RefreshButton from '../common/RefreshButton';
import { LoadingSpinner } from '../common/LoadingSpinner';
import { ErrorDisplay } from '../common/ErrorDisplay';
import Pagination from '../common/Pagination.tsx';
import { apiClient } from '../utils/apiClient';
import { downloadFile } from '../utils/download';
import { usePageTitle } from '../hooks/usePageTitle';
import { PAGINATION_CONFIG } from '../constants/config';
import { filterHiddenFields } from '../utils/filterHiddenFields';
import { FaArrowLeft, FaCopy, FaDownload } from 'react-icons/fa';
import { notification } from '../common/NotificationContext';
import { EmptyState, ErrorState, TableSkeleton } from '../common/Table';
import '../pages/ResourceListPage.css';
import '../tabs/YamlTab.css';
import '../styles/crd-page.css';

interface CRDSummary {
  name: string;
  group: string;
  version: string;
  kind: string;
  plural: string;
  scope: string;
  instanceCnt: number;
}

type ViewType = 'list' | 'instances' | 'yaml';

export const CRDPage: React.FC<{ collapsed: boolean; onToggleCollapsed: () => void }> = ({
  collapsed,
  onToggleCollapsed,
}) => {
  const [view, setView] = useState<ViewType>('list');
  const [crds, setCrds] = useState<CRDSummary[]>([]);
  const [instances, setInstances] = useState<Record<string, any>[]>([]);
  const [instanceYaml, setInstanceYaml] = useState<string>('');
  const [selectedCRD, setSelectedCRD] = useState<CRDSummary | null>(null);
  const [selectedInstance, setSelectedInstance] = useState<Record<string, any> | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [refreshKey, setRefreshKey] = useState(0);
  const [searchQuery, setSearchQuery] = useState('');
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(PAGINATION_CONFIG.DEFAULT_PAGE_SIZE);

  const crdTitle = view === 'list' ? 'CRDs' : view === 'instances' ? `${selectedCRD?.kind || 'CRD'} Instances` : `${selectedInstance?.metadata?.name || 'Instance'} YAML`;
  usePageTitle(crdTitle);

  const fetchCRDs = useCallback(async () => {
    setLoading(true);
    setError(null);
    try {
      const result = await apiClient.get<CRDSummary[]>('/api/v1/crds');
      setCrds(result.data || []);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load CRDs');
    } finally {
      setLoading(false);
    }
  }, []);

  useEffect(() => {
    if (view === 'list') {
      fetchCRDs();
    }
  }, [fetchCRDs, refreshKey, view]);

  const fetchInstances = useCallback(async (crd: CRDSummary) => {
    setLoading(true);
    setError(null);
    try {
      const group = crd.group || '_';
      const result = await apiClient.get<Record<string, any>[]>(
        `/api/v1/crds/${group}/${crd.version}/${crd.plural}`
      );
      setInstances(result.data || []);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load CRD instances');
    } finally {
      setLoading(false);
    }
  }, []);

  const fetchInstanceYaml = useCallback(async (crd: CRDSummary, instance: Record<string, any>) => {
    setLoading(true);
    setError(null);
    try {
      const group = crd.group || '_';
      const name = instance?.metadata?.name || '';
      const namespace = instance?.metadata?.namespace || '_';
      const result = await apiClient.get<Record<string, any>>(
        `/api/v1/crds/${group}/${crd.version}/${crd.plural}/${namespace}/${name}`
      );
      const filtered = filterHiddenFields(result.data);
      const yaml = jsyaml.dump(filtered, {
        indent: 2,
        lineWidth: -1,
        noRefs: true,
        quotingType: '"',
        forceQuotes: false,
      });
      setInstanceYaml(yaml);
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Failed to load instance YAML');
    } finally {
      setLoading(false);
    }
  }, []);

  const handleCRDClick = useCallback((crd: CRDSummary) => {
    setSelectedCRD(crd);
    setView('instances');
    setSearchQuery('');
  }, []);

  const handleInstanceClick = useCallback(async (instance: Record<string, any>) => {
    if (!selectedCRD) return;
    setSelectedInstance(instance);
    setView('yaml');
    await fetchInstanceYaml(selectedCRD, instance);
  }, [selectedCRD, fetchInstanceYaml]);

  const handleBack = useCallback(() => {
    if (view === 'instances') {
      setView('list');
      setSelectedCRD(null);
      setInstances([]);
    } else if (view === 'yaml') {
      setView('instances');
      setSelectedInstance(null);
      setInstanceYaml('');
    }
  }, [view]);

  const handleRefresh = useCallback(() => {
    setRefreshKey(prev => prev + 1);
  }, []);

  const handleCopyYaml = useCallback(async () => {
    try {
      await navigator.clipboard.writeText(instanceYaml);
      notification.success('YAML copied to clipboard');
    } catch {
      notification.error('Copy failed');
    }
  }, [instanceYaml]);

  const handleDownloadYaml = useCallback(() => {
    const name = selectedInstance?.metadata?.name || 'instance';
    downloadFile(instanceYaml, `${name}.yaml`, 'text/yaml;charset=utf-8');
  }, [instanceYaml, selectedInstance]);

  useEffect(() => {
    if (view === 'instances' && selectedCRD) {
      fetchInstances(selectedCRD);
    }
  }, [view, selectedCRD, fetchInstances, refreshKey]);

  useEffect(() => {
    setPage(1);
  }, [searchQuery, view]);

  const handlePageChange = useCallback((newPage: number) => {
    setPage(newPage);
  }, []);

  const handlePageSizeChange = useCallback((newSize: number) => {
    setPageSize(newSize);
    setPage(1);
  }, []);

  const filteredCRDs = useMemo(() => {
    if (!searchQuery.trim()) return crds;
    const q = searchQuery.toLowerCase();
    return crds.filter(
      c =>
        c.name.toLowerCase().includes(q) ||
        c.kind.toLowerCase().includes(q) ||
        c.group.toLowerCase().includes(q) ||
        c.plural.toLowerCase().includes(q)
    );
  }, [crds, searchQuery]);

  const filteredInstances = useMemo(() => {
    if (!searchQuery.trim()) return instances;
    const q = searchQuery.toLowerCase();
    return instances.filter(
      i =>
        (i.metadata?.name || '').toLowerCase().includes(q) ||
        (i.metadata?.namespace || '').toLowerCase().includes(q) ||
        (i.kind || '').toLowerCase().includes(q)
    );
  }, [instances, searchQuery]);

  const paginatedCRDs = useMemo(() => {
    const start = (page - 1) * pageSize;
    return filteredCRDs.slice(start, start + pageSize);
  }, [filteredCRDs, page, pageSize]);

  const paginatedInstances = useMemo(() => {
    const start = (page - 1) * pageSize;
    return filteredInstances.slice(start, start + pageSize);
  }, [filteredInstances, page, pageSize]);

  const breadcrumbs = useMemo(() => {
    if (view === 'list') {
      return [{ label: 'CRDs', path: 'crds' }];
    }
    if (view === 'instances' && selectedCRD) {
      return [
        { label: 'CRDs', path: 'crds' },
        { label: selectedCRD.kind, path: '' },
      ];
    }
    if (view === 'yaml' && selectedCRD && selectedInstance) {
      return [
        { label: 'CRDs', path: 'crds' },
        { label: selectedCRD.kind, path: '' },
        { label: selectedInstance?.metadata?.name || '', path: '' },
      ];
    }
    return [{ label: 'CRDs', path: 'crds' }];
  }, [view, selectedCRD, selectedInstance]);

  const pageTitle =
    view === 'list' ? 'Custom Resource Definitions' :
    view === 'instances' ? `${selectedCRD?.kind || 'CRD'} Instances` :
    `${selectedInstance?.metadata?.name || 'Instance'} YAML`;

  const headerActions = (
    <div className="crd-header-actions">
      {view !== 'list' && (
        <button className="crd-back-btn" onClick={handleBack} title="Go back">
          <FaArrowLeft />
        </button>
      )}
      <SearchInput
        placeholder={
          view === 'list' ? 'Search CRDs...' :
          view === 'instances' ? 'Search instances...' :
          'Search...'
        }
        value={searchQuery}
        onChange={e => setSearchQuery(e.target.value)}
        onClear={() => setSearchQuery('')}
        showClearButton={true}
      />
      <RefreshButton onClick={handleRefresh} loading={loading} />
    </div>
  );

  const renderBody = () => {
    if (view === 'yaml') {
      if (loading && !instanceYaml) {
        return (
          <div className="yaml-tab">
            <LoadingSpinner text="Loading YAML..." size="lg" />
          </div>
        );
      }
      if (error && !instanceYaml) {
        return (
          <div className="yaml-tab">
            <ErrorDisplay message={error} type="error" showRetry onRetry={handleRefresh} />
          </div>
        );
      }
      return (
        <div className="yaml-tab">
          <div className="yaml-toolbar">
            <div className="yaml-toolbar-left">
              <span className="yaml-title">Instance YAML</span>
              <span className="yaml-title-separator">|</span>
              <span className="yaml-status-inline">
                Lines: {instanceYaml.split('\n').length} | Chars: {instanceYaml.length}
              </span>
            </div>
            <div className="yaml-toolbar-actions">
              <button className="toolbar-btn" onClick={handleCopyYaml} title="Copy to clipboard">
                <FaCopy />
              </button>
              <button className="toolbar-btn" onClick={handleDownloadYaml} title="Download YAML">
                <FaDownload />
              </button>
            </div>
          </div>
          <div className="yaml-editor-container">
            <pre className="yaml-editor-pre">
              <code>{instanceYaml}</code>
            </pre>
          </div>
        </div>
      );
    }

    const isInstanceView = view === 'instances';
    const items = isInstanceView ? paginatedInstances : paginatedCRDs;
    const total = isInstanceView ? filteredInstances.length : filteredCRDs.length;
    const isClusterScoped = selectedCRD?.scope === 'Cluster';
    const colSpan = isInstanceView ? (isClusterScoped ? 3 : 4) : 6;

    return (
      <>
        <div className="table-container">
          <table className="resource-table">
            <thead>
              {isInstanceView ? (
                <tr>
                  <th style={{ width: '35%' }}>Name</th>
                  {!isClusterScoped && <th style={{ width: '25%' }}>Namespace</th>}
                  <th style={{ width: '15%' }}>Kind</th>
                  <th style={{ width: '25%' }}>API Version</th>
                </tr>
              ) : (
                <tr>
                  <th style={{ width: '30%' }}>Name</th>
                  <th style={{ width: '20%' }}>Group</th>
                  <th style={{ width: '12%' }}>Version</th>
                  <th style={{ width: '13%' }}>Kind</th>
                  <th style={{ width: '13%' }}>Scope</th>
                  <th style={{ width: '12%' }}>Instances</th>
                </tr>
              )}
            </thead>
            <tbody>
              {loading && items.length === 0 ? (
                <TableSkeleton columns={colSpan} rows={8} />
              ) : error && items.length === 0 ? (
                <ErrorState message={error} colSpan={colSpan} onRetry={handleRefresh} retryText="Reload" />
              ) : items.length === 0 ? (
                <EmptyState
                  message={
                    isInstanceView
                      ? `No ${selectedCRD?.kind || 'CRD'} instances found`
                      : searchQuery
                        ? 'No CRDs match your search'
                        : 'No Custom Resource Definitions found'
                  }
                  description={!isInstanceView && !searchQuery ? 'CRDs extend the Kubernetes API with custom resources.' : undefined}
                  colSpan={colSpan}
                />
              ) : isInstanceView ? (
                (items as Record<string, any>[]).map((inst, idx) => (
                  <tr key={inst.metadata?.uid || idx} onClick={() => handleInstanceClick(inst)} className="clickable">
                    <td>
                      <span className="resource-name-link">{inst.metadata?.name || '-'}</span>
                    </td>
                    {!isClusterScoped && <td>{inst.metadata?.namespace || '-'}</td>}
                    <td>{inst.kind || '-'}</td>
                    <td><span className="table-cell-text">{inst.apiVersion || '-'}</span></td>
                  </tr>
                ))
              ) : (
                (items as CRDSummary[]).map(crd => (
                  <tr key={crd.name} onClick={() => handleCRDClick(crd)} className="clickable">
                    <td>
                      <span className="resource-name-link">{crd.name}</span>
                    </td>
                    <td><span className="table-cell-text">{crd.group}</span></td>
                    <td>{crd.version}</td>
                    <td><span className="kind-badge">{crd.kind}</span></td>
                    <td>
                      <span className={`scope-badge ${crd.scope === 'Cluster' ? 'scope-cluster' : 'scope-namespaced'}`}>
                        {crd.scope || 'Namespaced'}
                      </span>
                    </td>
                    <td>{crd.instanceCnt}</td>
                  </tr>
                ))
              )}
            </tbody>
          </table>
        </div>
        {total > pageSize && (
          <Pagination
            currentPage={page}
            total={total}
            pageSize={pageSize}
            onPageChange={handlePageChange}
            onPageSizeChange={handlePageSizeChange}
            showQuickJumper
          />
        )}
      </>
    );
  };

  return (
    <div className="resource-list-page">
      <PageHeader
        title={pageTitle}
        breadcrumbs={breadcrumbs}
        onBreadcrumbClick={(path) => {
          if (path === 'crds') {
            setView('list');
            setSelectedCRD(null);
            setSelectedInstance(null);
          }
        }}
        collapsed={collapsed}
        onToggleCollapsed={onToggleCollapsed}
      >
        {headerActions}
      </PageHeader>
      {renderBody()}
    </div>
  );
};

export default CRDPage;
