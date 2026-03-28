import React, { useState, useCallback, useMemo } from 'react';
import { useParams, useNavigate } from 'react-router-dom';
import PageHeader from '../PageHeader';
import { ResourceActionBar } from './components/ResourceActionBar';
import { TabNavigation, TabItem } from './components/TabNavigation';
import { OverviewTab } from './tabs/OverviewTab';
import { YamlTab } from './tabs/YamlTab';
import { LogsTab } from './tabs/LogsTab';
import { TerminalTab } from './tabs/TerminalTab';
import { RelatedTab } from './tabs/RelatedTab';
import { EventsTab } from './tabs/EventsTab';
import { usePodDetail } from './hooks/usePodDetail';
import { LoadingSpinner } from '../LoadingSpinner';
import { ErrorDisplay } from '../ErrorDisplay';
import './PodDetailPage.css';

/**
 * Tab 配置
 */
const TAB_CONFIG: TabItem[] = [
  { key: 'overview', label: 'Overview', icon: '📊' },
  { key: 'yaml', label: 'YAML', icon: '📄' },
  { key: 'logs', label: 'Logs', icon: '📝' },
  { key: 'terminal', label: 'Terminal', icon: '💻' },
  { key: 'related', label: 'Related', icon: '🔗' },
  { key: 'events', label: 'Events', icon: '⚡' },
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
  
  // 使用 Pod 数据 Hook
  const {
    data: pod,
    loading,
    error,
    refresh,
  } = usePodDetail({ namespace, name: podName });
  
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
        const response = await fetch(`/api/pods/${namespace}/${podName}`, {
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
  const handleBreadcrumbClick = useCallback((path: string) => {
    if (path) {
      navigate(path);
    }
  }, [navigate]);
  
  // 面包屑导航
  const breadcrumbs = useMemo(() => [
    { label: 'Pods', path: '/pods' },
    { label: namespace, path: `/pods?namespace=${namespace}` },
    { label: podName, path: '' },
  ], [namespace, podName]);
  
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
      <TabNavigation
        tabs={TAB_CONFIG}
        activeTab={activeTab}
        onTabChange={handleTabChange}
      />

      {/* Tab 内容 */}
      {activeTab === 'overview' && (
        <OverviewTab pod={pod} loading={loading} onRefresh={refresh} />
      )}
      {activeTab === 'yaml' && (
        <YamlTab
          namespace={namespace}
          name={podName}
          pod={pod}
        />
      )}
      {activeTab === 'logs' && (
        <LogsTab
          namespace={namespace}
          name={podName}
          containers={pod?.spec?.containers || []}
        />
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
        />
      )}
    </div>
  );
};

export default PodDetailPage;
