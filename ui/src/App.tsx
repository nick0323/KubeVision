import React, { useState, Suspense, useEffect } from 'react';
import { BrowserRouter, Routes, Route, Navigate } from 'react-router-dom';
import './App.css';
import LoadingSpinner from './common/LoadingSpinner.tsx';
import ErrorBoundary from './common/ErrorBoundary.tsx';
import { NotificationContainer, useNotification } from './common/Notification';
import { SidebarLayout } from './common/SidebarLayout';
import LoginPage from './pages/LoginPage.tsx';
import { authUtils } from './utils/auth.ts';
import { PAGE_COMPONENTS } from './constants/page-components.tsx';
import { ResourceDetailPage as ImportedResourceDetail } from './pages/ResourceDetailPage';

/**
 * List Page Component
 */
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

  return (
    <SidebarLayout activeTab={tab}>
      {({ collapsed, onToggleCollapsed }) => (
        <ErrorBoundary
          fallback={({ error, onRetry }) => (
            <div className="error-fallback">
              <h3>Page Load Failed</h3>
              <p>{error?.message || 'Unknown error'}</p>
              <button onClick={onRetry} className="retry-btn">
                Retry
              </button>
            </div>
          )}
        >
          <Suspense fallback={<LoadingSpinner text="Loading..." size="lg" className="app-loading" />}>
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

/**
 * Generic Resource Detail Page Wrapper
 */
function GenericResourceDetail({ resourceType }: { resourceType: string }) {
  const [tab] = useState<string>(() => {
    try {
      const item = localStorage.getItem('current_tab');
      return item || 'overview';
    } catch {
      return 'overview';
    }
  });

  const handleMenuClick = () => {
    window.dispatchEvent(new CustomEvent('tab-change', { detail: tab }));
  };

  return (
    <SidebarLayout activeTab={tab} onMenuClick={handleMenuClick}>
      {({ collapsed, onToggleCollapsed }) => (
        <ImportedResourceDetail
          resourceType={resourceType}
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
  const { notifications, removeNotification } = useNotification();
  const [login, setLogin] = useState<boolean>(() => authUtils.isLoggedIn());

  useEffect(() => {
    const checkLogin = () => {
      setLogin(authUtils.isLoggedIn());
    };

    window.addEventListener('storage', checkLogin);
    return () => window.removeEventListener('storage', checkLogin);
  }, []);

  if (!login) {
    return (
      <>
        <LoginPage onLogin={() => setLogin(true)} />
        <NotificationContainer
          notifications={notifications}
          onRemove={removeNotification}
        />
      </>
    );
  }

  return (
    <>
      <BrowserRouter>
        <Routes>
          <Route path="/" element={<ListPage />} />
          <Route path="/pod/:namespace/:name" element={<GenericResourceDetail resourceType="pod" />} />
          <Route path="/deployment/:namespace/:name" element={<GenericResourceDetail resourceType="deployment" />} />
          <Route path="/statefulset/:namespace/:name" element={<GenericResourceDetail resourceType="statefulset" />} />
          <Route path="/daemonset/:namespace/:name" element={<GenericResourceDetail resourceType="daemonset" />} />
          <Route path="/service/:namespace/:name" element={<GenericResourceDetail resourceType="service" />} />
          <Route path="/configmap/:namespace/:name" element={<GenericResourceDetail resourceType="configmap" />} />
          <Route path="/secret/:namespace/:name" element={<GenericResourceDetail resourceType="secret" />} />
          <Route path="/ingress/:namespace/:name" element={<GenericResourceDetail resourceType="ingress" />} />
          <Route path="/job/:namespace/:name" element={<GenericResourceDetail resourceType="job" />} />
          <Route path="/cronjob/:namespace/:name" element={<GenericResourceDetail resourceType="cronjob" />} />
          <Route path="/pvc/:namespace/:name" element={<GenericResourceDetail resourceType="pvc" />} />
          <Route path="/pv/:namespace/:name" element={<GenericResourceDetail resourceType="pv" />} />
          <Route path="/storageclass/:namespace/:name" element={<GenericResourceDetail resourceType="storageclass" />} />
          <Route path="/namespace/:namespace/:name" element={<GenericResourceDetail resourceType="namespace" />} />
          <Route path="/node/:namespace/:name" element={<GenericResourceDetail resourceType="node" />} />
          <Route path="*" element={<Navigate to="/" replace />} />
        </Routes>
      </BrowserRouter>
      <NotificationContainer
        notifications={notifications}
        onRemove={removeNotification}
      />
    </>
  );
};

export default AppWithNotification;
