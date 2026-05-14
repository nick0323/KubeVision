import React, { useState, Suspense, useEffect } from 'react';
import { BrowserRouter, Routes, Route, Navigate, useNavigate, useSearchParams } from 'react-router-dom';
import './App.css';
import LoadingSpinner from './common/LoadingSpinner.tsx';
import ErrorBoundary from './common/ErrorBoundary.tsx';
import { NotificationProvider } from './common/NotificationContext';
import { NotificationContainerWrapper } from './common/NotificationContainerWrapper';
import { SidebarLayout } from './common/SidebarLayout';
import LoginPage from './pages/LoginPage.tsx';
import { authUtils } from './utils/auth';
import { PAGE_COMPONENTS } from './constants/page-components.tsx';
import { RESOURCE_TYPE_MAP } from './constants';
import { ResourceDetailPage as ImportedResourceDetail } from './pages/ResourceDetailPage';

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
  const [searchParams] = useSearchParams();
  const urlTab = searchParams.get('tab') || 'overview';
  const [tab, setTab] = useState(urlTab);

  useEffect(() => {
    setTab(urlTab);
    localStorage.setItem('current_tab', JSON.stringify(urlTab));
  }, [urlTab]);

  // Cross-tab sync
  useEffect(() => {
    const handleStorageChange = (e: StorageEvent) => {
      if (e.key === 'current_tab' && e.newValue) {
        const newTab = JSON.parse(e.newValue);
        if (newTab !== tab) {
          setTab(newTab);
        }
      }
    };
    window.addEventListener('storage', handleStorageChange);
    return () => window.removeEventListener('storage', handleStorageChange);
  }, [tab]);

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

function sidebarKeyForResource(resourceType: string): string {
  return RESOURCE_TYPE_MAP[resourceType as keyof typeof RESOURCE_TYPE_MAP] || 'overview';
}

/**
 * Generic Resource Detail Page Wrapper
 */
function GenericResourceDetail({ resourceType }: { resourceType: string }) {
  const tab = sidebarKeyForResource(resourceType);

  return (
    <SidebarLayout activeTab={tab}>
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
 * Resource detail route config（simplify重复定义）
 */
const RESOURCE_DETAIL_ROUTES = [
  { path: '/pod', resourceType: 'pod' },
  { path: '/deployment', resourceType: 'deployment' },
  { path: '/statefulset', resourceType: 'statefulset' },
  { path: '/daemonset', resourceType: 'daemonset' },
  { path: '/service', resourceType: 'service' },
  { path: '/configmap', resourceType: 'configmap' },
  { path: '/secret', resourceType: 'secret' },
  { path: '/ingress', resourceType: 'ingress' },
  { path: '/job', resourceType: 'job' },
  { path: '/cronjob', resourceType: 'cronjob' },
  { path: '/pvc', resourceType: 'pvc' },
  { path: '/pv', resourceType: 'pv' },
  { path: '/storageclass', resourceType: 'storageclass' },
  { path: '/namespace', resourceType: 'namespace' },
  { path: '/node', resourceType: 'node' },
  { path: '/horizontalpodautoscaler', resourceType: 'horizontalpodautoscaler' },
  { path: '/networkpolicy', resourceType: 'networkpolicy' },
  { path: '/serviceaccount', resourceType: 'serviceaccount' },
  { path: '/role', resourceType: 'role' },
  { path: '/rolebinding', resourceType: 'rolebinding' },
  { path: '/resourcequota', resourceType: 'resourcequota' },
  { path: '/limitrange', resourceType: 'limitrange' },
  { path: '/poddisruptionbudget', resourceType: 'poddisruptionbudget' },
  { path: '/clusterrole', resourceType: 'clusterrole' },
  { path: '/clusterrolebinding', resourceType: 'clusterrolebinding' },
] as const;

/**
 * Main App Component with Notification support
 */
const AppWithNotification: React.FC = () => {
  const [login, setLogin] = useState<boolean>(() => authUtils.isLoggedIn());

  useEffect(() => {
    const checkLogin = () => {
      setLogin(authUtils.isLoggedIn());
    };

    const handleLogout = () => {
      setLogin(false);
    };

    window.addEventListener('storage', checkLogin);
    window.addEventListener('logout', handleLogout);
    return () => {
      window.removeEventListener('storage', checkLogin);
      window.removeEventListener('logout', handleLogout);
    };
  }, []);

  return (
    <NotificationProvider>
      <NotificationContainerWrapper />
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
                  <LoginPage onLogin={() => setLogin(true)} />
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

            {/* 详情页路由（动态生成） */}
            {RESOURCE_DETAIL_ROUTES.map(({ path, resourceType }) => (
              <Route
                key={resourceType}
                path={`${path}/:namespace/:name`}
                element={
                  <RequireAuth>
                    <GenericResourceDetail resourceType={resourceType} />
                  </RequireAuth>
                }
              />
            ))}

            {/* 根路径重定向 */}
            <Route path="/" element={<RequireAuth><ListPage /></RequireAuth>} />

            {/* 未匹配重定向 */}
            <Route path="*" element={<Navigate to="/" replace />} />
          </Routes>
        </BrowserRouter>
      </>
    </NotificationProvider>
  );
};

export default AppWithNotification;
