import React, { useState, Suspense, useEffect, useCallback } from 'react';
import { BrowserRouter, Routes, Route, Navigate, useNavigate, useParams } from 'react-router-dom';
import './App.css';
import LoadingSpinner from './common/LoadingSpinner.tsx';
import ErrorBoundary from './common/ErrorBoundary.tsx';
import { useLocalStorage } from './hooks/useLocalStorage.ts';
import { SidebarLayout } from './common/SidebarLayout';
import LoginPage from './pages/LoginPage.tsx';
import { authUtils } from './utils/auth.ts';
import { PAGE_COMPONENTS } from './constants/page-components.tsx';
import { ResourceDetailPage as ImportedResourceDetail } from './resources';

const ListPage: React.FC = () => {
  const getInitialTab = () => {
    try {
      const item = localStorage.getItem('current_tab');
      return item || 'overview';
    } catch {
      return 'overview';
    }
  };

  const [tab, setTab] = useState<string>(getInitialTab);

  useEffect(() => {
    const handleTabChange = (newTab: string) => {
      setTab(newTab);
    };

    const handleCustomEvent = (e: Event) => {
      const customEvent = e as CustomEvent<string>;
      handleTabChange(customEvent.detail);
    };

    const handleStorageChange = (e: StorageEvent) => {
      if (e.key === 'current_tab' && e.newValue) {
        handleTabChange(e.newValue);
      }
    };

    window.addEventListener('tab-change', handleCustomEvent);
    window.addEventListener('storage', handleStorageChange);
    return () => {
      window.removeEventListener('tab-change', handleCustomEvent);
      window.removeEventListener('storage', handleStorageChange);
    };
  }, []);

  const renderPage = () => {
    return null;
  };

  return (
    <SidebarLayout activeTab={tab}>
      {({ collapsed, onToggleCollapsed }) => (
        <ErrorBoundary
          fallback={({ error, onRetry }) => (
            <div className="error-fallback">
              <h3>页面加载失败</h3>
              <p>{error?.message || '未知错误'}</p>
              <button onClick={onRetry} className="retry-btn">
                重试
              </button>
            </div>
          )}
        >
          <Suspense fallback={<LoadingSpinner text="加载中..." size="lg" className="app-loading" />}>
            {(() => {
              const PageComponent = PAGE_COMPONENTS[tab as keyof typeof PAGE_COMPONENTS];
              return PageComponent ? (
                <PageComponent collapsed={collapsed} onToggleCollapsed={onToggleCollapsed} />
              ) : null;
            })()}
          </Suspense>
        </ErrorBoundary>
      )}
    </SidebarLayout>
  );
};

// 主应用组件
export default function App() {
  const [login, setLogin] = useState<boolean>(() => authUtils.isLoggedIn());

  // 监听登录状态变化
  useEffect(() => {
    const checkLogin = () => {
      setLogin(authUtils.isLoggedIn());
    };

    window.addEventListener('storage', checkLogin);
    return () => window.removeEventListener('storage', checkLogin);
  }, []);

  if (!login) {
    return (
      <Suspense fallback={<LoadingSpinner text="加载中..." overlay />}>
        <LoginPage onLogin={() => setLogin(true)} />
      </Suspense>
    );
  }

  return (
    <BrowserRouter>
      <Routes>
        {/* 列表页面 */}
        <Route path="/" element={<ListPage />} />

        {/* Pod 详情页 */}
        <Route
          path="/pod/:namespace/:name"
          element={<GenericResourceDetail resourceType="pod" />}
        />
        {/* 其他资源详情页 - 所有资源统一使用 /:resourceType/:namespace/:name 格式 */}
        <Route
          path="/deployment/:namespace/:name"
          element={<GenericResourceDetail resourceType="deployment" />}
        />
        <Route
          path="/statefulset/:namespace/:name"
          element={<GenericResourceDetail resourceType="statefulset" />}
        />
        <Route
          path="/daemonset/:namespace/:name"
          element={<GenericResourceDetail resourceType="daemonset" />}
        />
        <Route
          path="/service/:namespace/:name"
          element={<GenericResourceDetail resourceType="service" />}
        />
        <Route
          path="/configmap/:namespace/:name"
          element={<GenericResourceDetail resourceType="configmap" />}
        />
        <Route
          path="/secret/:namespace/:name"
          element={<GenericResourceDetail resourceType="secret" />}
        />
        <Route
          path="/ingress/:namespace/:name"
          element={<GenericResourceDetail resourceType="ingress" />}
        />
        <Route
          path="/job/:namespace/:name"
          element={<GenericResourceDetail resourceType="job" />}
        />
        <Route
          path="/cronjob/:namespace/:name"
          element={<GenericResourceDetail resourceType="cronjob" />}
        />
        <Route
          path="/pvc/:namespace/:name"
          element={<GenericResourceDetail resourceType="pvc" />}
        />
        <Route path="/pv/:namespace/:name" element={<GenericResourceDetail resourceType="pv" />} />
        <Route
          path="/storageclass/:namespace/:name"
          element={<GenericResourceDetail resourceType="storageclass" />}
        />
        <Route
          path="/namespace/:namespace/:name"
          element={<GenericResourceDetail resourceType="namespace" />}
        />
        <Route
          path="/node/:namespace/:name"
          element={<GenericResourceDetail resourceType="node" />}
        />

        {/* 重定向 */}
        <Route path="*" element={<Navigate to="/" replace />} />
      </Routes>
    </BrowserRouter>
  );
}

// 通用资源详情页包装器 - 使用 SidebarLayout 替代重复代码
function GenericResourceDetail({ resourceType }: { resourceType: string }) {
  const params = useParams<{ namespace?: string; name: string }>();
  const navigate = useNavigate();

  // 当前激活的 tab（用于侧边栏高亮）
  const [tab] = useState<string>(() => {
    try {
      const item = localStorage.getItem('current_tab');
      return item || 'overview';
    } catch {
      return 'overview';
    }
  });

  const handleMenuClick = (key: string) => {
    localStorage.setItem('current_tab', key);
    window.dispatchEvent(new CustomEvent('tab-change', { detail: key }));
    navigate('/');
  };

  return (
    <SidebarLayout activeTab={tab} onMenuClick={handleMenuClick}>
      {({ collapsed, onToggleCollapsed }) => (
        <ImportedResourceDetail
          resourceType={resourceType}
          namespace={params.namespace || 'default'}
          name={params.name || ''}
          collapsed={collapsed}
          onToggleCollapsed={onToggleCollapsed}
        />
      )}
    </SidebarLayout>
  );
}
