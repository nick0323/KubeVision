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
import { LoadingSpinner } from '../common/LoadingSpinner';
import { ErrorDisplay } from '../common/ErrorDisplay';
import { authFetch } from '../utils/auth';
import { notification } from '../common/Notification';
import '../styles/detail-page.css';

// 导入resource特定 Tabs
import { PodsTab } from '../tabs/PodsTab';

// Pod 特has Tabs（复use现hasComponent）
import { LogsTab } from '../tabs/LogsTab';
import { TerminalTab } from '../tabs/TerminalTab';

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

  /**
   * ProcessDelete
   */
  const handleDelete = useCallback(async () => {
    try {
      const response = await authFetch(`/api/${resourceType}/${namespace}/${resourceName}`, {
        method: 'DELETE',
      });
      const result = await response.json();

      if (result.code === 0) {
        notification.success(`${config.title} deleted`);
        // BackList页
        navigate(`/${resourceType}s`);
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
        // settings current_tab (Use JSON.stringify Keep with useLocalStorage consistent)
        localStorage.setItem('current_tab', JSON.stringify(path));

        // Trigger event component
        window.dispatchEvent(new CustomEvent('tab-change', { detail: path }));

        // jump totoList页（根路径）
        navigate('/');
      } else {
        // path for空，notProcess（namespace and name 's路径for空）
        return;
      }
    },
    [navigate]
  );

  // determineis否forclusterresource
  const isClusterResource = ['node', 'pv', 'storageclass', 'namespace'].includes(resourceType);

  // 面包屑Config - with Pod detail pagekeepconsistent：resourceType > namespace(not可Click) > name
  const breadcrumbs = useMemo(
    () => [
      { label: config.title, path: getResourceListPath(resourceType) },
      ...(isClusterResource ? [] : [{ label: namespace, path: '' }]),
      { label: resourceName, path: '' },
    ],
    [config.title, resourceType, namespace, resourceName, isClusterResource]
  );

  /**
   * according toresourceTypeGetList页路径（tab key，not带斜杠）
   */
  function getResourceListPath(type: string): string {
    const typeToPath: Record<string, string> = {
      pod: 'pods',
      deployment: 'deployments',
      statefulset: 'statefulsets',
      daemonset: 'daemonsets',
      job: 'jobs',
      cronjob: 'cronjobs',
      service: 'services',
      ingress: 'ingress',
      configmap: 'configmaps',
      secret: 'secrets',
      pvc: 'pvcs',
      pv: 'pvs',
      storageclass: 'storageclasses',
      namespace: 'namespaces',
      node: 'nodes',
    };
    return typeToPath[type] || 'overview';
  }

  // Tab Config
  const tabItems: TabItem[] = useMemo(
    () =>
      config.tabs.map(key => ({
        key,
        label: key.charAt(0).toUpperCase() + key.slice(1),
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
          <TerminalTab
            namespace={namespace}
            name={resourceName}
            containers={(data as Record<string, any>)?.spec?.containers || []}
          />
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
        name={(data as any)?.metadata?.name || resourceName}
        namespace={isClusterResource ? undefined : namespace}
        onRefresh={refresh}
        onDelete={handleDelete}
      />

      {/* Tab navigation */}
      <TabNavigation tabs={tabItems} activeTab={activeTab} onTabChange={handleTabChange} />

      {/* Tab content */}
      {renderTabContent()}
    </div>
  );
};

export default ResourceDetailPage;
