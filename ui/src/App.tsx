import React, { useState, Suspense, useEffect, useCallback } from 'react';
import { BrowserRouter, Routes, Route, Navigate, useParams, useNavigate } from 'react-router-dom';
import './App.css';
import LoadingSpinner from './components/LoadingSpinner.tsx';
import ErrorBoundary from './components/ErrorBoundary.tsx';
import { useLocalStorage } from './hooks/useLocalStorage.ts';
import { FaChartPie, FaCube, FaRocket, FaTree, FaServer, FaCogs, FaBriefcase, FaClock, FaNetworkWired, FaDoorOpen, FaHdd, FaDatabase, FaListAlt, FaFileAlt, FaLock, FaBell, FaThLarge, FaChevronDown, FaChevronRight } from 'react-icons/fa';
import { LuSquareDashed } from 'react-icons/lu';
import { MENU_LIST } from './constants';
import LoginPage from './LoginPage.tsx';
import { FiLogOut } from 'react-icons/fi';
import { authUtils } from './utils/auth.ts';
import { PAGE_COMPONENTS } from './pages.tsx';
import ResourceDetailPage from './components/ResourceDetailPage';

// 图标映射
const ICON_MAP: Record<string, React.ReactNode> = {
  FaChartPie: <FaChartPie />,
  FaCube: <FaCube />,
  FaRocket: <FaRocket />,
  FaTree: <FaTree />,
  FaServer: <FaServer />,
  FaCogs: <FaCogs />,
  FaBriefcase: <FaBriefcase />,
  FaClock: <FaClock />,
  FaNetworkWired: <FaNetworkWired />,
  FaDoorOpen: <FaDoorOpen />,
  FaHdd: <FaHdd />,
  FaDatabase: <FaDatabase />,
  FaListAlt: <FaListAlt />,
  FaFileAlt: <FaFileAlt />,
  FaLock: <FaLock />,
  FaBell: <FaBell />,
  FaThLarge: <FaThLarge />,
  FaChevronDown: <FaChevronDown />,
  FaChevronRight: <FaChevronRight />,
};

// 列表页面组件
const ListPage: React.FC = () => {
  const [tab, setTab] = useLocalStorage<string>('current_tab', 'overview');
  const [collapsed, setCollapsed] = useLocalStorage<boolean>('sider_collapsed', false);
  const [openGroups, setOpenGroups] = useState<Record<string, boolean>>(() => {
    const state: Record<string, boolean> = {};
    MENU_LIST.forEach(g => { state[g.group] = true; });
    return state;
  });

  const toggleCollapsed = () => {
    setCollapsed(prev => {
      const next = !prev;
      localStorage.setItem('sider_collapsed', JSON.stringify(next));
      return next;
    });
  };

  const toggleGroup = (group: string) => {
    setOpenGroups(prev => ({ ...prev, [group]: !prev[group] }));
  };

  const renderPage = () => {
    const PageComponent = PAGE_COMPONENTS[tab as keyof typeof PAGE_COMPONENTS];
    if (PageComponent) {
      return <PageComponent collapsed={collapsed} onToggleCollapsed={toggleCollapsed} />;
    }
    return null;
  };

  return (
    <div className="layout-root" data-sider-collapsed={collapsed}>
      {/* 侧边栏 */}
      <div className={`sider-menu ${collapsed ? 'collapsed' : ''}`}>
        {/* Logo */}
        <div className="logo-area">
          <span className="logo-text-full">KubeVision</span>
          <span className="logo-text-compact">KV</span>
        </div>

        {/* 菜单 */}
        <div className="sider-scroll">
          <ul>
            {/* Overview */}
            {MENU_LIST[0].items.map(item => (
              <li
                key={item.key}
                className={tab === item.key ? 'active' : ''}
                onClick={() => setTab(item.key)}
                data-tip={item.label}
              >
                <span className="icon">{ICON_MAP[item.icon]}</span>
                <span>{item.label}</span>
              </li>
            ))}

            {/* 分组菜单 */}
            {MENU_LIST.slice(1).map(group => (
              <React.Fragment key={group.group}>
                {!collapsed && (
                  <li className="menu-group-title">
                    <span>{group.group}</span>
                    <span
                      style={{marginLeft:8,cursor:'pointer',display:'flex',alignItems:'center'}}
                      onClick={(e) => { e.stopPropagation(); toggleGroup(group.group); }}
                    >
                      {openGroups[group.group] ? <FaChevronDown size={12}/> : <FaChevronRight size={12}/>}
                    </span>
                  </li>
                )}
                {(collapsed || openGroups[group.group]) && group.items.map(item => (
                  <li
                    key={item.key}
                    className={tab === item.key ? 'active' : ''}
                    onClick={() => setTab(item.key)}
                    data-tip={item.label}
                  >
                    <span className="icon">{ICON_MAP[item.icon]}</span>
                    <span>{item.label}</span>
                  </li>
                ))}
              </React.Fragment>
            ))}
          </ul>
        </div>

        {/* 退出按钮 */}
        <div className="sider-bottom">
          <button className="logout-btn" onClick={() => { authUtils.clearToken(); }}>
            <span className="icon"><FiLogOut /></span>
            <span>Sign out</span>
          </button>
        </div>
      </div>

      {/* 主内容区 */}
      <div className="main-content">
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
            {renderPage()}
          </Suspense>
        </ErrorBoundary>
      </div>
    </div>
  );
};

// 详情页面包装器
const ResourceDetailWrapper: React.FC = () => {
  const { resourceType, namespace, name } = useParams<{ resourceType: string; namespace?: string; name?: string }>();
  const navigate = useNavigate();

  if (!resourceType || !name) {
    return <Navigate to="/" replace />;
  }

  return <ResourceDetailPage resourceType={resourceType} />;
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

  // 登出处理
  const handleLogout = useCallback(() => {
    authUtils.clearToken();
    setLogin(false);
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
        
        {/* 详情页面 */}
        <Route path="/:resourceType/:namespace/:name" element={<ResourceDetailWrapper />} />
        <Route path="/:resourceType/:name" element={<ResourceDetailWrapper />} />
        
        {/* 重定向 */}
        <Route path="*" to="/" />
      </Routes>
    </BrowserRouter>
  );
}
