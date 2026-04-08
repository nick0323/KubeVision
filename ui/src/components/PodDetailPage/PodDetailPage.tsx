import React, { useState, useCallback, useMemo } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import { Pod } from '../../types/k8s-resources';
import PageHeader from '../PageHeader';
import { ResourceActionBar } from '../common/ResourceActionBar';
import { TabNavigation, TabItem } from '../common/TabNavigation';
import { OverviewTab } from './tabs/OverviewTab';
import { YamlTab } from '../ResourceDetail/tabs/YamlTab';
import { LogsTab } from './tabs/LogsTab';
import { TerminalTab } from './tabs/TerminalTab';
import { RelatedTab } from './tabs/RelatedTab';
import { EventsTab } from './tabs/EventsTab';
import { useResourceDetail } from '../ResourceDetailPage/hooks/useResourceDetail';
import { LoadingSpinner } from '../LoadingSpinner';
import { ErrorDisplay } from '../ErrorDisplay';
import { authFetch } from '../../utils/auth';
import '../../styles/detail-page.css';

/**
 * Tab 配置
 */
const TAB_CONFIG: TabItem[] = [
  { key: 'overview', label: 'Overview' },
  { key: 'yaml', label: 'YAML' },
  { key: 'logs', label: 'Logs' },
  { key: 'terminal', label: 'Terminal' },
  { key: 'related', label: 'Related' },
  { key: 'events', label: 'Events' },
];

export interface PodDetailPageProps {
  collapsed: boolean;
  onToggleCollapsed: () => void;
}

/**
 * Pod 详情页组件
 */
export const PodDetailPage: React.FC<PodDetailPageProps> = ({ collapsed, onToggleCollapsed }) => {
  const params = useParams<{ namespace: string; name: string }>();
  const navigate = useNavigate();

  const namespace = params.namespace || 'default';
  const podName = params.name || '';

  // 当前激活的 Tab
  const [activeTab, setActiveTab] = useState<string>('overview');

  // 使用通用 Hook 获取 Pod 数据
  const { data: pod, loading, error, refresh } = useResourceDetail<Pod>({ 
    resourceType: 'pod',
    namespace, 
    name: podName 
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
    if (window.confirm(`确定要删除 Pod "${podName}" 吗？此操作不可恢复。`)) {
      try {
        // 后端只支持单数形式：/api/pod/ns/name
      const response = await authFetch(`/api/pod/${namespace}/${podName}`, {
          method: 'DELETE',
        });
        const result = await response.json();

        if (result.code === 0) {
          alert('Pod 已删除');
          navigate('/pods');
        } else {
          alert(`删除失败：${result.message}`);
        }
      } catch (err) {
        alert(`删除失败：${err instanceof Error ? err.message : '未知错误'}`);
      }
    }
  }, [namespace, podName, navigate]);

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

        // 跳转到列表页
        navigate(path);
      }
    },
    [navigate]
  );

  // 面包屑配置 - 资源类型 > namespace(不可点击) > name
  const breadcrumbs = useMemo(
    () => [
      { label: 'Pods', path: '/pods' },
      { label: namespace, path: '' }, // namespace 不可点击
      { label: podName, path: '' },
    ],
    [namespace, podName]
  );

  // 渲染加载状态
  if (loading && !pod) {
    return (
      <div className="pod-detail-page">
        <LoadingSpinner text="加载 Pod 详情..." size="lg" overlay />
      </div>
    );
  }

  // 渲染错误状态
  if (error && !pod) {
    return (
      <div className="pod-detail-page">
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
    <div className="pod-detail-page">
      {/* 页面头部 - 使用 PageHeader 组件 */}
      <PageHeader
        title="Pod 详情"
        breadcrumbs={breadcrumbs}
        onBreadcrumbClick={handleBreadcrumbClick}
        collapsed={collapsed}
        onToggleCollapsed={onToggleCollapsed}
      >
        {/* 右侧操作按钮可以放在这里 */}
      </PageHeader>

      {/* 资源信息栏 */}
      <ResourceActionBar
        name={pod?.metadata?.name || podName}
        namespace={namespace}
        onRefresh={refresh}
        onDelete={handleDelete}
        onDescribe={() => {}}
      />

      {/* Tab 导航 */}
      <TabNavigation tabs={TAB_CONFIG} activeTab={activeTab} onTabChange={handleTabChange} />

      {/* Tab 内容 */}
      {activeTab === 'overview' && <OverviewTab pod={pod} loading={loading} onRefresh={refresh} resourceType="pod" />}
      {activeTab === 'yaml' && <YamlTab namespace={namespace} name={podName} resourceType="pod" data={pod} />}
      {activeTab === 'logs' && (
        <LogsTab namespace={namespace} name={podName} containers={pod?.spec?.containers || []} />
      )}
      {activeTab === 'terminal' && (
        <TerminalTab
          namespace={namespace}
          name={podName}
          containers={pod?.spec?.containers || []}
        />
      )}
      {activeTab === 'related' && (
        <RelatedTab
          namespace={namespace}
          name={podName}
          ownerReferences={pod?.metadata?.ownerReferences || []}
        />
      )}
      {activeTab === 'events' && (
        <EventsTab
          namespace={namespace}
          podName={podName}
          name={podName}
          resourceKind="Pod"
          onRefresh={refresh}
        />
      )}
    </div>
  );
};

export default PodDetailPage;
