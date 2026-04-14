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
import { useResourceDetail } from '../resources/useResourceDetail';
import { ResourceDetailPageProps, RESOURCE_CONFIGS } from './ResourceDetailPage.types';
import { LoadingSpinner } from '../common/LoadingSpinner';
import { ErrorDisplay } from '../common/ErrorDisplay';
import { authFetch } from '../utils/auth';
import '../styles/detail-page.css';

// 导入资源特定 Tabs
import { PodsTab } from '../tabs/PodsTab';

// Pod 特有 Tabs（复用现有组件）
import { LogsTab } from '../tabs/LogsTab';
import { TerminalTab } from '../tabs/TerminalTab';

/**
 * 通用资源详情页组件
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

  // 优先使用 URL 参数
  const namespace = params.namespace || namespaceFromProps || 'default';
  const resourceName = params.name || name || '';

  // 获取资源配置
  const config = RESOURCE_CONFIGS[resourceType] || {
    title: resourceType,
    tabs: ['overview', 'yaml', 'events'],
  };

  // 当前激活的 Tab
  const [activeTab, setActiveTab] = useState<string>(config.tabs[0] || 'overview');

  // 当资源类型为 Pod 时，确保 activeTab 不是 'pods'
  useEffect(() => {
    if (resourceType === 'pod' && activeTab === 'pods') {
      setActiveTab('overview');
    }
  }, [resourceType, activeTab]);

  // 使用通用 Hook 获取数据
  const { data, loading, error, refresh } = useResourceDetail({
    resourceType,
    namespace,
    name: resourceName,
    // autoRefresh: true,
    // refreshInterval: 30000,
  });

  /**
   * 处理 Tab 切换
   */
  const handleTabChange = useCallback((tabKey: string) => {
    setActiveTab(tabKey);
  }, []);

  /**
   * 处理删除
   */
  const handleDelete = useCallback(async () => {
    if (window.confirm(`确定要删除 ${config.title} "${resourceName}" 吗？此操作不可恢复。`)) {
      try {
        const response = await authFetch(`/api/${resourceType}/${namespace}/${resourceName}`, {
          method: 'DELETE',
        });
        const result = await response.json();

        if (result.code === 0) {
          alert(`${config.title} 已删除`);
          // 返回列表页
          navigate(`/${resourceType}s`);
        } else {
          alert(`删除失败：${result.message}`);
        }
      } catch (err) {
        alert(`删除失败：${err instanceof Error ? err.message : '未知错误'}`);
      }
    }
  }, [resourceType, namespace, resourceName, config.title, navigate]);

  /**
   * 处理面包屑跳转
   */
  const handleBreadcrumbClick = useCallback(
    (path: string) => {
      if (path) {
        // 设置 current_tab
        localStorage.setItem('current_tab', path);

        // 触发事件
        window.dispatchEvent(new CustomEvent('tab-change', { detail: path }));

        // 跳转到列表页（根路径）
        navigate('/');
      } else {
        // path 为空，不处理（namespace 和 name 的路径为空）
        return;
      }
    },
    [navigate]
  );

  // 判断是否为集群资源
  const isClusterResource = ['node', 'pv', 'storageclass', 'namespace'].includes(resourceType);

  // 面包屑配置 - 与 Pod 详情页保持一致：资源类型 > namespace(不可点击) > name
  const breadcrumbs = useMemo(
    () => [
      { label: config.title, path: getResourceListPath(resourceType) },
      ...(isClusterResource ? [] : [{ label: namespace, path: '' }]),
      { label: resourceName, path: '' },
    ],
    [config.title, resourceType, namespace, resourceName, isClusterResource]
  );

  /**
   * 根据资源类型获取列表页路径（tab key，不带斜杠）
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

  // Tab 配置
  const tabItems: TabItem[] = useMemo(
    () =>
      config.tabs.map(key => ({
        key,
        label: key.charAt(0).toUpperCase() + key.slice(1),
      })),
    [config.tabs]
  );

  // 渲染 Tab 内容
  const renderTabContent = useCallback(() => {
    if (loading) {
      return <LoadingSpinner text="加载中..." size="lg" />;
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
            ownerReferences={(data as any)?.metadata?.ownerReferences}
          />
        );

      case 'logs':
        return (
          <LogsTab
            namespace={namespace}
            name={resourceName}
            containers={(data as any)?.spec?.containers || []}
          />
        );

      case 'terminal':
        return (
          <TerminalTab
            namespace={namespace}
            name={resourceName}
            containers={(data as any)?.spec?.containers || []}
          />
        );

      case 'pods':
        return (
          <PodsTab
            namespace={namespace}
            resourceName={resourceName}
            resourceKind={config.title}
            resourceLabels={
              (data as any)?.spec?.selector?.matchLabels || (data as any)?.metadata?.labels || {}
            }
            ownerReferences={
              (data as any)?.metadata?.uid
                ? [{ uid: (data as any).metadata.uid, kind: config.title, name: resourceName }]
                : []
            }
          />
        );

      default:
        return <div>Tab 内容不存在</div>;
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

  // 渲染加载状态
  if (loading && !data) {
    return (
      <div className="resource-detail-page">
        <LoadingSpinner text={`加载 ${config.title} 详情...`} size="lg" overlay />
      </div>
    );
  }

  // 渲染错误状态
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
      {/* 页面头部 - 使用 PageHeader 组件 */}
      <PageHeader
        title={`${config.title} 详情`}
        breadcrumbs={breadcrumbs}
        onBreadcrumbClick={handleBreadcrumbClick}
        collapsed={collapsed}
        onToggleCollapsed={onToggleCollapsed}
      >
        {/* 右侧操作按钮可以放在这里 */}
      </PageHeader>

      {/* 资源信息栏 */}
      <ResourceActionBar
        name={data?.metadata?.name || resourceName}
        namespace={isClusterResource ? undefined : namespace}
        onRefresh={refresh}
        onDelete={handleDelete}
      />

      {/* Tab 导航 */}
      <TabNavigation tabs={tabItems} activeTab={activeTab} onTabChange={handleTabChange} />

      {/* Tab 内容 */}
      {renderTabContent()}
    </div>
  );
};

export default ResourceDetailPage;
