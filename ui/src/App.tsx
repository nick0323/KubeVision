import React, { useState, Suspense, useEffect } from 'react';
import { BrowserRouter, Routes, Route, Navigate, useNavigate } from 'react-router-dom';
import './App.css';
import LoadingSpinner from './common/LoadingSpinner.tsx';
import ErrorBoundary from './common/ErrorBoundary.tsx';
import { NotificationProvider } from './common/NotificationContext';
import { NotificationContainerWrapper } from './common/NotificationContainerWrapper';
import { SidebarLayout } from './common/SidebarLayout';
import LoginPage from './pages/LoginPage.tsx';
import { authUtils } from './utils/auth';
import { PAGE_COMPONENTS } from './constants/page-components.tsx';
import { ResourceDetailPage as ImportedResourceDetail } from './pages/ResourceDetailPage';
import { useLocalStorage } from './hooks/useLocalStorage';

/**
 * 路由守卫组件：未登录时重定向到登录页
 */
const RequireAuth: React.FC<{ children: React.ReactNode }> = ({ children }) => {
  const navigate = useNavigate();
  const [ready, setReady] = useState(false);

  useEffect(() => {
    if (!authUtils.isLoggedIn()) {
      navigate('/login', { replace: true });
    } else {
      setReady(true);
    }
  }, [navigate]);

  if (!ready) {
    return null;
  }

  return <>{children}</>;
};

/**
 * Page Renderer Component (Alternative IIFE)
 */
const PageRenderer: React.FC<{ tab: string; collapsed: boolean; onToggleCollapsed: () => void }> = ({ tab, collapsed, onToggleCollapsed }) => {
  const PageComponent = PAGE_COMPONENTS[tab as keyof typeof PAGE_COMPONENTS];
  return PageComponent ? (
    <PageComponent collapsed={collapsed} onToggleCollapsed={onToggleCollapsed} />
  ) : null;
};

/**
 * List Page Component
 */
const ListPage: React.FC = () => {
  const [tab, setTab] = useLocalStorage<string>('current_tab', 'overview');

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

  return (
    <SidebarLayout activeTab={tab}>
      {({ collapsed, onToggleCollapsed }) => (
        <ErrorBoundary
          fallback={({ error, onRetry }) => (
            <div className="error-fallback">
              <h3 className="error-title">Page Load Failed</h3>
              <p>{error?.message || 'Unknown error'}</p>
              <button onClick={onRetry} className="retry-btn">
                Retry
              </button>
            </div>
          )}
        >
          <Suspense fallback={<LoadingSpinner text="Loading..." size="lg" className="app-loading" />}>
            <PageRenderer tab={tab} collapsed={collapsed} onToggleCollapsed={onToggleCollapsed} />
          </Suspense>
        </ErrorBoundary>
      )}
    </SidebarLayout>
  );
};

/**
 * Generic Resource Detail Page Wrapper
 */
function GenericResourceDetail({ resourceType }: { resourceType: string }) {
  const [tab] = useLocalStorage<string>('current_tab', 'overview');

  const handleMenuClick = () => {
    window.dispatchEvent(new CustomEvent('tab-change', { detail: tab }));
  };

  return (
    <SidebarLayout activeTab={tab} onMenuClick={handleMenuClick}>
      {({ collapsed, onToggleCollapsed }) => (
        <ImportedResourceDetail
          resourceType={resourceType}
          namespace="default"
          name=""
          collapsed={collapsed}
          onToggleCollapsed={onToggleCollapsed}
        />
      )}
    </SidebarLayout>
  );
}

/**
 * Main App Component with Notification support
 */
const AppWithNotification: React.FC = () => {
  const [login, setLogin] = useState<boolean>(() => authUtils.isLoggedIn());

  useEffect(() => {
    const checkLogin = () => {
      setLogin(authUtils.isLoggedIn());
    };

    window.addEventListener('storage', checkLogin);
    return () => window.removeEventListener('storage', checkLogin);
  }, []);

  return (
    <NotificationProvider>
      <>
        <BrowserRouter>
          <Routes>
            {/* 登录页 */}
            <Route
              path="/login"
              element={
                login ? (
                  <Navigate to="/" replace />
                ) : (
                  <>
                    <LoginPage onLogin={() => setLogin(true)} />
                    <NotificationContainerWrapper />
                  </>
                )
              }
            />

            {/* 受保护路由 */}
            <Route
              path="/*"
              element={
                <RequireAuth>
                  <ListPage />
                </RequireAuth>
              }
            />

            {/* 详情页路由 */}
            <Route
              path="/pod/:namespace/:name"
              element={
                <RequireAuth>
                  <GenericResourceDetail resourceType="pod" />
                </RequireAuth>
              }
            />
            <Route
              path="/deployment/:namespace/:name"
              element={
                <RequireAuth>
                  <GenericResourceDetail resourceType="deployment" />
                </RequireAuth>
              }
            />
            <Route
              path="/statefulset/:namespace/:name"
              element={
                <RequireAuth>
                  <GenericResourceDetail resourceType="statefulset" />
                </RequireAuth>
              }
            />
            <Route
              path="/daemonset/:namespace/:name"
              element={
                <RequireAuth>
                  <GenericResourceDetail resourceType="daemonset" />
                </RequireAuth>
              }
            />
            <Route
              path="/service/:namespace/:name"
              element={
                <RequireAuth>
                  <GenericResourceDetail resourceType="service" />
                </RequireAuth>
              }
            />
            <Route
              path="/configmap/:namespace/:name"
              element={
                <RequireAuth>
                  <GenericResourceDetail resourceType="configmap" />
                </RequireAuth>
              }
            />
            <Route
              path="/secret/:namespace/:name"
              element={
                <RequireAuth>
                  <GenericResourceDetail resourceType="secret" />
                </RequireAuth>
              }
            />
            <Route
              path="/ingress/:namespace/:name"
              element={
                <RequireAuth>
                  <GenericResourceDetail resourceType="ingress" />
                </RequireAuth>
              }
            />
            <Route
              path="/job/:namespace/:name"
              element={
                <RequireAuth>
                  <GenericResourceDetail resourceType="job" />
                </RequireAuth>
              }
            />
            <Route
              path="/cronjob/:namespace/:name"
              element={
                <RequireAuth>
                  <GenericResourceDetail resourceType="cronjob" />
                </RequireAuth>
              }
            />
            <Route
              path="/pvc/:namespace/:name"
              element={
                <RequireAuth>
                  <GenericResourceDetail resourceType="pvc" />
                </RequireAuth>
              }
            />
            <Route
              path="/pv/:namespace/:name"
              element={
                <RequireAuth>
                  <GenericResourceDetail resourceType="pv" />
                </RequireAuth>
              }
            />
            <Route
              path="/storageclass/:namespace/:name"
              element={
                <RequireAuth>
                  <GenericResourceDetail resourceType="storageclass" />
                </RequireAuth>
              }
            />
            <Route
              path="/namespace/:namespace/:name"
              element={
                <RequireAuth>
                  <GenericResourceDetail resourceType="namespace" />
                </RequireAuth>
              }
            />
            <Route
              path="/node/:namespace/:name"
              element={
                <RequireAuth>
                  <GenericResourceDetail resourceType="node" />
                </RequireAuth>
              }
            />

            {/* 根路径重定向 */}
            <Route path="/" element={<ListPage />} />

            {/* 未匹配重定向 */}
            <Route path="*" element={<Navigate to="/" replace />} />
          </Routes>
        </BrowserRouter>
      </>
    </NotificationProvider>
  );
};

export default AppWithNotification;
