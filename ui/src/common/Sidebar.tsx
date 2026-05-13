import React, { useState, useCallback, useEffect, useRef } from 'react';
import { useNavigate } from 'react-router-dom';
import { useLocalStorage } from '../hooks/useLocalStorage';
import { authUtils, authFetch } from '../utils/auth';
import apiClient from '../utils/apiClient';
import { MENU_LIST, STORAGE_KEYS } from '../constants';
import k8sLogo from '../assets/kubernetes-logo.svg';
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
  FaShieldAlt,
  FaUserSecret,
  FaUserTag,
  FaUserCheck,
  FaTachometerAlt,
  FaSlidersH,
  FaBalanceScale,
  FaPuzzlePiece,
  FaHdd,
  FaDatabase,
  FaListAlt,
  FaFileAlt,
  FaLock,
  FaBell,
  FaThLarge,
  FaChevronDown,
  FaChevronRight,
  FaSync,
  FaGitAlt,
  FaArrowsAltV,
  FaCloud,
} from 'react-icons/fa';
import { FiLogOut } from 'react-icons/fi';

/**
 * iconMapping
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
  FaSync: <FaSync />,
  FaGitAlt: <FaGitAlt />,
  FaArrowsAltV: <FaArrowsAltV />,
  FaShieldAlt: <FaShieldAlt />,
  FaUserSecret: <FaUserSecret />,
  FaUserTag: <FaUserTag />,
  FaUserCheck: <FaUserCheck />,
  FaTachometerAlt: <FaTachometerAlt />,
  FaSlidersH: <FaSlidersH />,
  FaBalanceScale: <FaBalanceScale />,
  FaPuzzlePiece: <FaPuzzlePiece />,
};

/**
 * Sidebar Props
 */
export interface SidebarProps {
  activeTab?: string;
  onMenuClick?: (key: string) => void;
  showActiveState?: boolean;
  collapsed: boolean;
  onToggleCollapsed: () => void;
}

/**
 * SidebarComponent
 *
 * UnifiedAlternative App.tsx 's 3 处Sidebarcode
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

  const [clusters, setClusters] = useState<string[]>([]);
  const [clusterOpen, setClusterOpen] = useState(false);
  const clusterRef = useRef<HTMLDivElement>(null);
  const currentCluster = localStorage.getItem(STORAGE_KEYS.CURRENT_CLUSTER) || 'default';

  useEffect(() => {
    apiClient.get<string[]>('/api/clusters').then(res => {
      const list = res.data || [];
      if (list.length > 1) setClusters(list);
    }).catch(() => {});
  }, []);

  useEffect(() => {
    const handleClick = (e: MouseEvent) => {
      if (clusterRef.current && !clusterRef.current.contains(e.target as Node)) {
        setClusterOpen(false);
      }
    };
    document.addEventListener('mousedown', handleClick);
    return () => document.removeEventListener('mousedown', handleClick);
  }, []);

  const handleClusterChange = useCallback((name: string) => {
    if (name === currentCluster) return;
    localStorage.setItem(STORAGE_KEYS.CURRENT_CLUSTER, name);
    window.location.reload();
  }, [currentCluster]);

  /**
   * toggle分组展开/收起
   */
  const toggleGroup = useCallback((group: string) => {
    setOpenGroups(prev => ({ ...prev, [group]: !prev[group] }));
  }, []);

  /**
   * ProcessMenuClick
   */
  const handleMenuClick = useCallback(
    (key: string) => {
      localStorage.setItem('current_tab', JSON.stringify(key));
      navigate('/?tab=' + encodeURIComponent(key));
    },
    [navigate]
  );

  /**
   * Process登出
   */
  const handleLogout = useCallback(async () => {
    try {
      await authFetch('/api/logout', { method: 'POST' });
    } catch {
      // ignore server errors, still logout locally
    }
    authUtils.clearToken();
    localStorage.removeItem('sider_collapsed');
    localStorage.removeItem('current_tab');
    window.dispatchEvent(new CustomEvent('logout'));
    navigate('/login', { replace: true });
  }, [navigate]);

  return (
    <div className={`sider-menu ${collapsed ? 'collapsed' : ''}`} data-testid="sidebar">
      {/* Logo */}
      <div className="logo-area">
        <span className="logo-text-full">KubeVision</span>
        <img src={k8sLogo} alt="Kubernetes" className="logo-compact" />
      </div>

      {/* Menu */}
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

          {/* Menu Group */}
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

      {/* Bottom: cluster + logout */}
      <div className="sider-bottom">
        {clusters.length > 0 && !collapsed && (
          <div className="cluster-selector" ref={clusterRef}>
            <div className="cluster-selector-trigger" onClick={() => setClusterOpen(o => !o)}>
              <span className="icon"><FaCloud /></span>
              <span className="cluster-name">{currentCluster}</span>
              <span className="cluster-arrow">{clusterOpen ? '▲' : '▼'}</span>
            </div>
            {clusterOpen && (
              <div className="cluster-dropdown">
                {clusters.map(name => (
                  <div
                    key={name}
                    className={`cluster-option ${name === currentCluster ? 'active' : ''}`}
                    onClick={() => handleClusterChange(name)}
                  >
                    <span className="cluster-check">{name === currentCluster ? '✓' : ''}</span>
                    {name}
                  </div>
                ))}
              </div>
            )}
          </div>
        )}
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
