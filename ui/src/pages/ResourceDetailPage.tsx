import React, { useState, useCallback, useMemo, useEffect } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import PageHeader from '../common/PageHeader.tsx';
import { ResourceActionBar } from '../common/ResourceActionBar';
import { TabNavigation, TabItem } from '../common/TabNavigation';
import { OverviewTab } from '../tabs/OverviewTab';
import { YamlTab } from '../tabs/YamlTab';
import { EventsTab } from '../tabs/EventsTab';
import { RelatedTab } from '../tabs/RelatedTab';
import { EndpointsTab } from '../tabs/EndpointsTab';
import { useResourceDetail } from '../hooks/useResourceDetail';
import { ResourceDetailPageProps, RESOURCE_CONFIGS } from './ResourceDetailPage.types';
import { capitalize } from '../utils/string';
import { RESOURCE_TYPE_MAP } from '../constants/config';
import { usePageTitle } from '../hooks/usePageTitle';
import { LoadingSpinner } from '../common/LoadingSpinner';
import { ErrorDisplay } from '../common/ErrorDisplay';
import { authFetch, withCluster } from '../utils/auth';
import { isClusterResource as isClusterScopeResource } from '../constants/config';
import { notification } from '../common/NotificationContext';
import '../styles/detail-page.css';

// 导入resource特定 Tabs
import { PodsTab } from '../tabs/PodsTab';

// Pod 特has Tabs（复use现hasComponent）
import { LogsTab } from '../tabs/LogsTab';
const TerminalTab = React.lazy(() => import('../tabs/TerminalTab'));

/**
 * CommonResource detail pageComponent
 */
export const ResourceDetailPage: React.FC<ResourceDetailPageProps> = ({
  resourceType,
  namespace: namespaceFromProps,
  name,
  collapsed,
  onToggleCollapsed,
}) => {
  const params = useParams<{ namespace?: string; name?: string }>();
  const navigate = useNavigate();

  // preferUse URL 参数
  const namespace = params.namespace || namespaceFromProps || 'default';
  const resourceName = params.name || name || '';

  // GetresourceConfig
  const config = RESOURCE_CONFIGS[resourceType] || {
    title: resourceType,
    tabs: ['overview', 'yaml', 'events'],
  };

  usePageTitle(`${config.title}: ${resourceName}`);

  // whenbefore激活's Tab
  const [activeTab, setActiveTab] = useState<string>(config.tabs[0] || 'overview');

  // whenresourceTypefor Pod 时，ensure activeTab notis 'pods'
  useEffect(() => {
    if (resourceType === 'pod' && activeTab === 'pods') {
      setActiveTab('overview');
    }
  }, [resourceType, activeTab]);

  // UseCommon Hook Getdata
  const { data, loading, error, refresh } = useResourceDetail({
    resourceType,
    namespace,
    name: resourceName,
    // autoRefresh: true,
    // refreshInterval: 30000,
  });

  /**
   * Process Tab toggle
   */
  const handleTabChange = useCallback((tabKey: string) => {
    setActiveTab(tabKey);
  }, []);

  const scalableTypes = ['deployment', 'statefulset'];
  const restartableTypes = ['deployment', 'statefulset', 'daemonset'];
  const canScale = scalableTypes.includes(resourceType);
  const canRestart = restartableTypes.includes(resourceType);
  const resource = data as Record<string, unknown> | null;
  const spec = resource?.spec as Record<string, unknown> | undefined;
  const meta = resource?.metadata as Record<string, unknown> | undefined;
  const currentReplicas = canScale ? (spec?.replicas as number) ?? 0 : undefined;

  const scaleReplicas = useCallback(async (delta: number) => {
    const d = data as Record<string, unknown> | null;
    const s = d?.spec as Record<string, unknown> | undefined;
    const current = (s?.replicas as number) ?? 0;
    const replicas = Math.max(0, current + delta);
    try {
      const response = await authFetch(withCluster(`/api/${resourceType}/${namespace}/${resourceName}/scale`), {
        method: 'PUT',
        body: JSON.stringify({ replicas }),
      });
      const result = await response.json();
      if (result.code === 0) {
        notification.success(`Scaled to ${replicas} replicas`);
        refresh();
      } else {
        notification.error(`Scale failed: ${result.message}`);
      }
    } catch (err) {
      notification.error(`Scale failed: ${err instanceof Error ? err.message : 'Unknown error'}`);
    }
  }, [resourceType, namespace, resourceName, data, refresh]);

  const handleScaleUp = useCallback(() => scaleReplicas(1), [scaleReplicas]);
  const handleScaleDown = useCallback(() => scaleReplicas(-1), [scaleReplicas]);

  const handleRestart = useCallback(async () => {
    try {
      const response = await authFetch(withCluster(`/api/${resourceType}/${namespace}/${resourceName}/restart`), {
        method: 'POST',
      });
      const result = await response.json();
      if (result.code === 0) {
        notification.success('Rolling restart initiated');
      } else {
        notification.error(`Restart failed: ${result.message}`);
      }
    } catch (err) {
      notification.error(`Restart failed: ${err instanceof Error ? err.message : 'Unknown error'}`);
    }
  }, [resourceType, namespace, resourceName]);

  /**
   * ProcessDelete
   */
  const handleDelete = useCallback(async () => {
    try {
      const response = await authFetch(withCluster(`/api/${resourceType}/${namespace}/${resourceName}`), {
        method: 'DELETE',
      });
      const result = await response.json();

      if (result.code === 0) {
        notification.success(`${config.title} deleted`);
        // BackList页
        navigate(`/${getResourceListPath(resourceType)}`);
      } else {
        notification.error(`Delete failed: ${result.message}`);
      }
    } catch (err) {
      notification.error(`Delete failed: ${err instanceof Error ? err.message : 'Unknown error'}`);
    }
  }, [resourceType, namespace, resourceName, config.title, navigate]);

  /**
   * Process面包屑jump to
   */
  const handleBreadcrumbClick = useCallback(
    (path: string) => {
      if (path) {
        localStorage.setItem('current_tab', JSON.stringify(path));
        navigate('/?tab=' + encodeURIComponent(path));
      }
    },
    [navigate]
  );

  // determineis否forclusterresource
  const isClusterResource = isClusterScopeResource(resourceType);

  const getResourceListPath = (type: string): string =>
    (RESOURCE_TYPE_MAP as Record<string, string>)[type] || 'overview';

  const breadcrumbs = useMemo(
    () => [
      { label: config.title, path: getResourceListPath(resourceType) },
      ...(isClusterResource ? [] : [{ label: namespace, path: '' }]),
      { label: resourceName, path: '' },
    ],
    [config.title, resourceType, namespace, resourceName, isClusterResource]
  );

  // Tab Config
  const tabItems: TabItem[] = useMemo(
    () =>
      config.tabs.map(key => ({
        key,
        label: capitalize(key),
      })),
    [config.tabs]
  );

  // Render Tab content
  const renderTabContent = useCallback(() => {
    if (loading) {
      return <LoadingSpinner text="Loading...." size="lg" />;
    }

    if (error) {
      return <ErrorDisplay message={error} type="error" showRetry onRetry={refresh} />;
    }

    switch (activeTab) {
      case 'overview':
        return (
          <OverviewTab
            data={data}
            loading={loading}
            onRefresh={refresh}
            resourceType={resourceType}
          />
        );

      case 'yaml':
        return (
          <YamlTab
            namespace={namespace}
            name={resourceName}
            resourceType={resourceType}
            data={data}
          />
        );

      case 'events':
        return <EventsTab namespace={namespace} name={resourceName} resourceKind={config.title} />;

      case 'endpoints':
        return <EndpointsTab namespace={namespace} serviceName={resourceName} />;

       case 'related':
        return (
          <RelatedTab
            namespace={namespace}
            name={resourceName}
            resourceType={resourceType}
            ownerReferences={(data as Record<string, any>)?.metadata?.ownerReferences}
          />
        );

      case 'logs':
        return (
          <LogsTab
            namespace={namespace}
            name={resourceName}
            containers={(data as Record<string, any>)?.spec?.containers || []}
          />
        );

      case 'terminal':
        return (
          <React.Suspense fallback={<LoadingSpinner text="Loading terminal..." size="lg" />}>
            <TerminalTab
              namespace={namespace}
              name={resourceName}
              containers={(data as Record<string, any>)?.spec?.containers || []}
            />
          </React.Suspense>
        );

      case 'pods':
        const resourceData = data as Record<string, any>;
        return (
          <PodsTab
            namespace={namespace}
            resourceName={resourceName}
            resourceKind={config.title}
            resourceLabels={
              resourceData?.spec?.selector?.matchLabels || resourceData?.metadata?.labels || {}
            }
            resourceUid={resourceData?.metadata?.uid}
          />
        );

      default:
        return <div>Tab content not found</div>;
    }
  }, [
    activeTab,
    loading,
    error,
    data,
    namespace,
    resourceName,
    resourceType,
    config.title,
    refresh,
  ]);

  // RenderLoadStatus
  if (loading && !data) {
    return (
      <div className="resource-detail-page">
        <LoadingSpinner text={`Loading ${config.title} details...`} size="lg" overlay />
      </div>
    );
  }

  // RenderError status
  if (error && !data) {
    return (
      <div className="resource-detail-page">
        <ErrorDisplay
          message={error}
          type="error"
          showRetry
          onRetry={() => window.location.reload()}
        />
      </div>
    );
  }

  return (
    <div className="resource-detail-page">
      {/* Page header - Using PageHeader component */}
      <PageHeader
        title={`${config.title} Details`}
        breadcrumbs={breadcrumbs}
        onBreadcrumbClick={handleBreadcrumbClick}
        collapsed={collapsed}
        onToggleCollapsed={onToggleCollapsed}
      >
        {/* Right action buttons can be placed here */}
      </PageHeader>

      {/* Resource info bar */}
      <ResourceActionBar
        name={(meta?.name as string) || resourceName}
        namespace={isClusterResource ? undefined : namespace}
        onRefresh={refresh}
        onDelete={handleDelete}
        onScaleUp={canScale ? handleScaleUp : undefined}
        onScaleDown={canScale ? handleScaleDown : undefined}
        currentReplicas={currentReplicas}
        onRestart={canRestart ? handleRestart : undefined}
      />

      {/* Tab navigation */}
      <TabNavigation tabs={tabItems} activeTab={activeTab} onTabChange={handleTabChange} />

      {/* Tab content */}
      {renderTabContent()}
    </div>
  );
};

export default ResourceDetailPage;
