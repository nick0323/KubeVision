import React, { useState, useCallback } from 'react';
import { useNavigate } from 'react-router-dom';
import { useLocalStorage } from '../hooks/useLocalStorage';
import { authUtils } from '../utils/auth';
import { MENU_LIST } from '../constants';
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
import { FiLogOut } from 'react-icons/fi';

/**
 * 图标映射
 */
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

/**
 * 侧边栏 Props
 */
export interface SidebarProps {
  activeTab?: string;
  onMenuClick?: (key: string) => void;
  showActiveState?: boolean;
  collapsed: boolean;
  onToggleCollapsed: () => void;
}

/**
 * 侧边栏组件
 *
 * 统一替代 App.tsx 中的 3 处侧边栏代码
 */
export const Sidebar: React.FC<SidebarProps> = ({
  activeTab,
  onMenuClick,
  showActiveState = true,
  collapsed,
  onToggleCollapsed,
}) => {
  const navigate = useNavigate();
  const [openGroups, setOpenGroups] = useState<Record<string, boolean>>(() => {
    const state: Record<string, boolean> = {};
    MENU_LIST.forEach(g => {
      state[g.group] = true;
    });
    return state;
  });

  /**
   * 切换分组展开/收起
   */
  const toggleGroup = useCallback((group: string) => {
    setOpenGroups(prev => ({ ...prev, [group]: !prev[group] }));
  }, []);

  /**
   * 处理菜单点击
   */
  const handleMenuClick = useCallback(
    (key: string) => {
      // 先设置 current_tab
      localStorage.setItem('current_tab', JSON.stringify(key));
      // 触发事件（通知已挂载的组件）
      window.dispatchEvent(new CustomEvent('tab-change', { detail: key }));
      // 跳转到列表页（ListPage 挂载时会读取最新的 current_tab）
      navigate('/');
      onMenuClick?.(key);
    },
    [navigate, onMenuClick]
  );

  /**
   * 处理登出
   */
  const handleLogout = useCallback(() => {
    authUtils.clearToken();
    localStorage.removeItem('sider_collapsed');
    localStorage.removeItem('current_tab');
    window.location.reload();
  }, []);

  return (
    <div className={`sider-menu ${collapsed ? 'collapsed' : ''}`} data-testid="sidebar">
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
              className={showActiveState && activeTab === item.key ? 'active' : ''}
              onClick={() => handleMenuClick(item.key)}
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
                    style={{ marginLeft: 8, cursor: 'pointer' }}
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
                    className={showActiveState && activeTab === item.key ? 'active' : ''}
                    onClick={() => handleMenuClick(item.key)}
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
  );
};

export default Sidebar;
