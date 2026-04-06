import React, { useState, Suspense, useEffect, useCallback } from 'react';
import { BrowserRouter, Routes, Route, Navigate, useNavigate } from 'react-router-dom';
import './App.css';
import LoadingSpinner from './components/LoadingSpinner.tsx';
import ErrorBoundary from './components/ErrorBoundary.tsx';
import { useLocalStorage } from './hooks/useLocalStorage.ts';
import {
  FaChartPie,
  FaCube,
  FaRocket,
  FaTree,
  FaServer,
  FaCogs,
  FaBriefcase,
  FaClock,
  FaNetworkWired,
  FaDoorOpen,
  FaHdd,
  FaDatabase,
  FaListAlt,
  FaFileAlt,
  FaLock,
  FaBell,
  FaThLarge,
  FaChevronDown,
  FaChevronRight,
} from 'react-icons/fa';
import { MENU_LIST } from './constants';
import LoginPage from './LoginPage.tsx';
import { FiLogOut } from 'react-icons/fi';
import { authUtils } from './utils/auth.ts';
import { PAGE_COMPONENTS } from './pages.tsx';
import { PodDetailPage } from './components/PodDetailPage';

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
  // 直接从 localStorage 读取初始值
  const getInitialTab = () => {
    try {
      const item = localStorage.getItem('current_tab');
      // 直接返回字符串，不需要 JSON.parse
      return item || 'overview';
    } catch {
      return 'overview';
    }
  };

  const [tab, setTab] = useState<string>(getInitialTab);
  const [collapsed, setCollapsed] = useLocalStorage<boolean>('sider_collapsed', false);
  const [openGroups, setOpenGroups] = useState<Record<string, boolean>>(() => {
    const state: Record<string, boolean> = {};
    MENU_LIST.forEach(g => {
      state[g.group] = true;
    });
    return state;
  });

  // 监听 storage 事件和自定义事件，同步其他页面的 tab 变化
  useEffect(() => {
    const handleTabChange = (newTab: string) => {
      setTab(newTab);
    };

    // 监听自定义 tab-change 事件（同标签页内同步）
    const handleCustomEvent = (e: Event) => {
      const customEvent = e as CustomEvent<string>;
      handleTabChange(customEvent.detail);
    };

    // 监听 storage 事件（跨标签页同步）
    const handleStorageChange = (e: StorageEvent) => {
      if (e.key === 'current_tab' && e.newValue) {
        // 直接返回字符串，不需要 JSON.parse
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

  // 登出处理
  const handleLogout = useCallback(() => {
    authUtils.clearToken();
    localStorage.removeItem('sider_collapsed');
    localStorage.removeItem('current_tab');
    // 强制刷新页面
    window.location.reload();
  }, []);

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
                      style={{
                        marginLeft: 8,
                        cursor: 'pointer',
                        display: 'flex',
                        alignItems: 'center',
                      }}
                      onClick={e => {
                        e.stopPropagation();
                        toggleGroup(group.group);
                      }}
                    >
                      {openGroups[group.group] ? (
                        <FaChevronDown size={12} />
                      ) : (
                        <FaChevronRight size={12} />
                      )}
                    </span>
                  </li>
                )}
                {(collapsed || openGroups[group.group]) &&
                  group.items.map(item => (
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
          <button className="logout-btn" onClick={handleLogout}>
            <span className="icon">
              <FiLogOut />
            </span>
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
          <Suspense
            fallback={<LoadingSpinner text="加载中..." size="lg" className="app-loading" />}
          >
            {renderPage()}
          </Suspense>
        </ErrorBoundary>
      </div>
    </div>
  );
};

// 详情页面包装器
const PodDetailWrapper: React.FC = () => {
  const [collapsed, setCollapsed] = useLocalStorage<boolean>('sider_collapsed', false);
  const [openGroups, setOpenGroups] = useState<Record<string, boolean>>(() => {
    const state: Record<string, boolean> = {};
    MENU_LIST.forEach(g => {
      state[g.group] = true;
    });
    return state;
  });
  const navigate = useNavigate();

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

  // 登出处理
  const handleLogout = useCallback(() => {
    authUtils.clearToken();
    localStorage.removeItem('sider_collapsed');
    localStorage.removeItem('current_tab');
    window.dispatchEvent(new Event('storage'));
    window.location.reload();
  }, []);

  // 菜单点击处理 - 跳转回列表页并设置对应的 tab
  const handleMenuClick = (key: string) => {
    // 先设置 localStorage
    localStorage.setItem('current_tab', key);
    // 触发事件（在当前标签页内同步）
    window.dispatchEvent(new CustomEvent('tab-change', { detail: key }));
    // 跳转到列表页
    navigate('/');
  };

  return (
    <div className="layout-root" data-sider-collapsed={collapsed}>
      {/* 侧边栏 - 复用列表页面的实现 */}
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
                className=""
                data-tip={item.label}
                onClick={() => handleMenuClick(item.key)}
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
                      style={{
                        marginLeft: 8,
                        cursor: 'pointer',
                        display: 'flex',
                        alignItems: 'center',
                      }}
                      onClick={e => {
                        e.stopPropagation();
                        toggleGroup(group.group);
                      }}
                    >
                      {openGroups[group.group] ? (
                        <FaChevronDown size={12} />
                      ) : (
                        <FaChevronRight size={12} />
                      )}
                    </span>
                  </li>
                )}
                {(collapsed || openGroups[group.group]) &&
                  group.items.map(item => (
                    <li
                      key={item.key}
                      className=""
                      data-tip={item.label}
                      onClick={() => handleMenuClick(item.key)}
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
          <button className="logout-btn" onClick={handleLogout}>
            <span className="icon">
              <FiLogOut />
            </span>
            <span>Sign out</span>
          </button>
        </div>
      </div>

      {/* 主内容区 - Pod 详情页 */}
      <div className="main-content">
        <PodDetailPage collapsed={collapsed} onToggleCollapsed={toggleCollapsed} />
      </div>
    </div>
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

        {/* 资源详情页（单数形式） */}
        {/* Pod 详情页 */}
        <Route path="/pod/:namespace/:name" element={<PodDetailWrapper />} />
        {/* 兼容复数形式路由（重定向到单数） */}
        <Route path="/pods/:namespace/:name" element={<Navigate to="/pod/:namespace/:name" replace />} />

        {/* 其他资源详情页（待实现） */}
        {/* <Route path="/deployment/:namespace/:name" element={<DeploymentDetailWrapper />} /> */}
        {/* <Route path="/service/:namespace/:name" element={<ServiceDetailWrapper />} /> */}

        {/* 重定向 */}
        <Route path="*" element={<Navigate to="/" replace />} />
      </Routes>
    </BrowserRouter>
  );
}
