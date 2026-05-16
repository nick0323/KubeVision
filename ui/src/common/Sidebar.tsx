import React, { useState, useCallback, useEffect, useRef } from 'react';
import { useNavigate } from 'react-router-dom';
import { useLocalStorage } from '../hooks/useLocalStorage';
import { authUtils, authFetch } from '../utils/auth';
import apiClient from '../utils/apiClient';
import { MENU_LIST, STORAGE_KEYS } from '../constants';
import { ClusterHealth } from '../types';
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
import { FiLogOut, FiSettings } from 'react-icons/fi';
import { notification } from '../common/NotificationContext';

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
  const [clusterHealth, setClusterHealth] = useState<Record<string, ClusterHealth>>({});
  const [settingsOpen, setSettingsOpen] = useState(false);
  const [showPasswordModal, setShowPasswordModal] = useState(false);
  const [passwordForm, setPasswordForm] = useState({ oldPassword: '', newPassword: '', confirmPassword: '' });
  const [changingPassword, setChangingPassword] = useState(false);
  const settingsRef = useRef<HTMLDivElement>(null);
  const currentCluster = localStorage.getItem(STORAGE_KEYS.CURRENT_CLUSTER) || 'default';

  useEffect(() => {
    apiClient.get<ClusterHealth[]>('/api/v1/clusters/health').then(res => {
      const list = res.data || [];
      const healthMap: Record<string, ClusterHealth> = {};
      list.forEach(h => { healthMap[h.name] = h; });
      setClusterHealth(healthMap);
      const names = list.map(h => h.name);
      setClusters(names);
    }).catch(() => {});
  }, []);

  useEffect(() => {
    const handleClick = (e: MouseEvent) => {
      if (settingsRef.current && !settingsRef.current.contains(e.target as Node)) {
        setSettingsOpen(false);
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

  const handlePasswordChange = useCallback(async () => {
    if (!passwordForm.oldPassword || !passwordForm.newPassword) {
      notification.warning('Please fill in all fields');
      return;
    }
    if (passwordForm.newPassword !== passwordForm.confirmPassword) {
      notification.warning('New passwords do not match');
      return;
    }
    if (passwordForm.newPassword.length < 8) {
      notification.warning('Password must be at least 8 characters');
      return;
    }
    setChangingPassword(true);
    try {
      await apiClient.post('/api/v1/admin/password/change', {
        oldPassword: passwordForm.oldPassword,
        newPassword: passwordForm.newPassword,
      });
      notification.success('Password changed successfully');
      setShowPasswordModal(false);
      setPasswordForm({ oldPassword: '', newPassword: '', confirmPassword: '' });
    } catch (err) {
      notification.error(`Failed to change password: ${err instanceof Error ? err.message : 'Unknown error'}`);
    } finally {
      setChangingPassword(false);
    }
  }, [passwordForm]);

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
    authUtils.clearTokens();
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
        <ul role="menubar">
          {MENU_LIST[0].items.map(item => {
            const isActive = showActiveState && activeTab === item.key;
            return (
              <li
                key={item.key}
                role="menuitem"
                tabIndex={0}
                className={isActive ? 'active' : ''}
                aria-current={isActive ? 'page' : undefined}
                onClick={() => handleMenuClick(item.key)}
                onKeyDown={e => { if (e.key === 'Enter' || e.key === ' ') { e.preventDefault(); handleMenuClick(item.key); } }}
                data-tip={item.label}
              >
                <span className="icon">{ICON_MAP[item.icon]}</span>
                <span>{item.label}</span>
              </li>
            );
          })}

          {MENU_LIST.slice(1).map(group => (
            <React.Fragment key={group.group}>
              {!collapsed && (
                <li className="menu-group-title" role="separator">
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
                group.items.map(item => {
                  const isActive = showActiveState && activeTab === item.key;
                  return (
                    <li
                      key={item.key}
                      role="menuitem"
                      tabIndex={0}
                      className={isActive ? 'active' : ''}
                      aria-current={isActive ? 'page' : undefined}
                      onClick={() => handleMenuClick(item.key)}
                      onKeyDown={e => { if (e.key === 'Enter' || e.key === ' ') { e.preventDefault(); handleMenuClick(item.key); } }}
                      data-tip={item.label}
                    >
                      <span className="icon">{ICON_MAP[item.icon]}</span>
                      <span>{item.label}</span>
                    </li>
                  );
                })}
            </React.Fragment>
          ))}
        </ul>
      </div>

      {/* Bottom: settings + cluster + logout */}
      <div className="sider-bottom">
        {!collapsed && (
          <div className="settings-area" ref={settingsRef}>
            <div className="settings-trigger" onClick={() => setSettingsOpen(o => !o)}>
              <span className="icon"><FiSettings /></span>
              <span>Settings</span>
              <span className="settings-arrow">{settingsOpen ? '▼' : '▲'}</span>
            </div>
            {settingsOpen && (
              <div className="settings-dropdown">
                <div className="settings-current-cluster">
                  <FaCloud size={14} />
                  <span className={`cluster-status-dot ${clusterHealth[currentCluster]?.healthy ? 'healthy' : 'unhealthy'}`} />
                  <span>{currentCluster}</span>
                  <span className="settings-host" title={(() => {
                    const h = clusterHealth[currentCluster];
                    return h ? `${h.host} | v${h.version} | ${h.nodeCount} nodes` : '';
                  })()}>
                    {clusterHealth[currentCluster]?.host || ''}
                  </span>
                </div>
                {clusters.length > 0 && (
                  <div className="settings-section">
                    {clusters.map(name => {
                      const health = clusterHealth[name];
                      return (
                        <div
                          key={name}
                          className={`settings-item ${name === currentCluster ? 'active' : ''}`}
                          onClick={() => handleClusterChange(name)}
                        >
                          <span className={`cluster-status-dot ${health?.healthy ? 'healthy' : 'unhealthy'}`} />
                          <span>{name}</span>
                          {name === currentCluster && <span className="settings-check">✓</span>}
                        </div>
                      );
                    })}
                  </div>
                )}
                <hr className="settings-divider" />
                <div className="settings-item" onClick={() => handleMenuClick('clusters')}>
                  <FaServer size={14} />
                  <span>Manage Clusters</span>
                </div>
                <div className="settings-item disabled" title="Coming soon">
                  <FaUserCheck size={14} />
                  <span>User Management</span>
                </div>
                <div className="settings-item" onClick={() => setShowPasswordModal(true)}>
                  <FaLock size={14} />
                  <span>Change Password</span>
                </div>
                <hr className="settings-divider" />
                <div className="settings-item danger" onClick={handleLogout}>
                  <FiLogOut size={14} />
                  <span>Sign Out</span>
                </div>
              </div>
            )}
          </div>
        )}
        {collapsed && (
          <button className="logout-btn" onClick={handleLogout} title="Sign out">
            <span className="icon"><FiLogOut /></span>
          </button>
        )}
      </div>

      {/* Password Change Modal */}
      {showPasswordModal && (
        <div className="password-modal-overlay" onClick={() => setShowPasswordModal(false)}>
          <div className="password-modal" onClick={e => e.stopPropagation()}>
            <h3>Change Password</h3>
            <div className="form-field">
              <label>Current Password</label>
              <input
                type="password"
                value={passwordForm.oldPassword}
                onChange={e => setPasswordForm(p => ({ ...p, oldPassword: e.target.value }))}
                placeholder="Enter current password"
              />
            </div>
            <div className="form-field">
              <label>New Password</label>
              <input
                type="password"
                value={passwordForm.newPassword}
                onChange={e => setPasswordForm(p => ({ ...p, newPassword: e.target.value }))}
                placeholder="Enter new password (min 8 chars)"
              />
            </div>
            <div className="form-field">
              <label>Confirm New Password</label>
              <input
                type="password"
                value={passwordForm.confirmPassword}
                onChange={e => setPasswordForm(p => ({ ...p, confirmPassword: e.target.value }))}
                placeholder="Confirm new password"
              />
            </div>
            <div className="form-buttons">
              <button className="action-btn" onClick={() => setShowPasswordModal(false)}>Cancel</button>
              <button
                className="create-resource-btn"
                onClick={handlePasswordChange}
                disabled={changingPassword}
              >
                {changingPassword ? 'Changing...' : 'Change Password'}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
};

export default Sidebar;
