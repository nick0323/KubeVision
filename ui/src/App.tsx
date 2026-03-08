import React, { useState, Suspense, useEffect, useCallback } from 'react';
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

// 导入页面组件映射
import { PAGE_COMPONENTS } from './pages.tsx';

// 图标映射 - 使用 useMemo 优化
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

// 页面组件映射
const CURRENT_PAGE_COMPONENTS = {
  ...PAGE_COMPONENTS
};

// 预加载策略：当用户 hover 到菜单项时预加载对应组件
const preloadComponent = (componentKey: string) => {
  const Component = CURRENT_PAGE_COMPONENTS[componentKey as keyof typeof CURRENT_PAGE_COMPONENTS];
  if (Component && (Component as any).preload) {
    (Component as any).preload();
  }
};

/**
 * 主应用组件 - 修复登录跳转问题
 */
export default function App() {
  // 使用函数初始化状态，确保读取最新状态
  const [login, setLogin] = useState<boolean>(() => authUtils.isLoggedIn());
  const [tab, setTab] = useLocalStorage<string>('current_tab', 'overview');
  const [collapsed, setCollapsed] = useLocalStorage<boolean>('sider_collapsed', false);
  const [openGroups, setOpenGroups] = useState<Record<string, boolean>>(() => {
    const state: Record<string, boolean> = {};
    MENU_LIST.forEach(g => { state[g.group] = true; });
    return state;
  });

  // 初始化认证拦截器 - 监听未授权事件
  useEffect(() => {
    const handleUnauthorized = () => {
      authUtils.clearToken();
      setLogin(false);
    };

    window.addEventListener('auth-unauthorized', handleUnauthorized);

    return () => {
      window.removeEventListener('auth-unauthorized', handleUnauthorized);
    };
  }, []);

  // 登录成功处理函数 - 使用回调确保状态更新
  const handleLoginSuccess = useCallback(() => {
    // 强制重新检查登录状态
    const isLoggedIn = authUtils.isLoggedIn();
    setLogin(isLoggedIn);
    
    // 触发页面刷新（确保状态同步）
    if (isLoggedIn) {
      window.location.href = '/';
    }
  }, []);

  // 折叠侧边栏 - 不使用 useCallback，直接传递函数
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

  const handleTabChange = (newTab: string) => {
    setTab(newTab);
    // 预加载相邻的组件
    const menuItems = MENU_LIST.flatMap(g => g.items);
    const currentIndex = menuItems.findIndex(item => item.key === newTab);
    if (currentIndex !== -1) {
      // 预加载下一个组件
      if (currentIndex + 1 < menuItems.length) {
        preloadComponent(menuItems[currentIndex + 1].key);
      }
      // 预加载上一个组件
      if (currentIndex - 1 >= 0) {
        preloadComponent(menuItems[currentIndex - 1].key);
      }
    }
  };

  // 渲染页面 - 直接渲染
  const renderPage = () => {
    const PageComponent = CURRENT_PAGE_COMPONENTS[tab as keyof typeof CURRENT_PAGE_COMPONENTS];
    if (PageComponent) {
      return <PageComponent collapsed={collapsed} onToggleCollapsed={toggleCollapsed} />;
    }
    return null;
  };

  if (!login) {
    return (
      <Suspense fallback={<LoadingSpinner text="加载中..." overlay />}>
        <LoginPage onLogin={handleLoginSuccess} />
      </Suspense>
    );
  }

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
                onClick={() => handleTabChange(item.key)}
                onMouseEnter={(e) => {
                  preloadComponent(item.key);
                }}
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
                    onClick={() => handleTabChange(item.key)}
                    onMouseEnter={(e) => {
                      preloadComponent(item.key);
                    }}
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
          <button className="logout-btn" onClick={() => { authUtils.clearToken(); setLogin(false); }}>
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
}
