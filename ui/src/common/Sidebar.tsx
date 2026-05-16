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

      {/* Bottom: settings */}
      <div className="sider-bottom">
        {!collapsed && (
          <div className="settings-area">
            <div className="settings-trigger" onClick={() => setSettingsOpen(true)}>
              <span className="icon"><FiSettings /></span>
              <span>Settings</span>
            </div>
          </div>
        )}
        {collapsed && (
          <button className="logout-btn" onClick={() => setSettingsOpen(true)} title="Settings">
            <span className="icon"><FiSettings /></span>
          </button>
        )}
      </div>

      {/* Settings Overlay */}
      {settingsOpen && (
        <div className="settings-overlay" onClick={() => setSettingsOpen(false)}>
          <div className="settings-page" onClick={e => e.stopPropagation()}>
            <div className="settings-page-header">
              <h2>Settings</h2>
              <button className="settings-close-btn" onClick={() => setSettingsOpen(false)}>✕</button>
            </div>

            <div className="settings-page-body">
              {/* Current Cluster */}
              <div className="settings-section-label">Cluster</div>
              <div className="settings-current-cluster">
                <FaCloud size={18} />
                <span className={`cluster-status-dot ${clusterHealth[currentCluster]?.healthy ? 'healthy' : 'unhealthy'}`} />
                <span className="settings-cluster-name">{currentCluster}</span>
                <span className={`status-text ${clusterHealth[currentCluster]?.healthy ? 'healthy' : 'unhealthy'}`}>
                  {clusterHealth[currentCluster] ? (clusterHealth[currentCluster].healthy ? 'Healthy' : 'Unhealthy') : 'Unknown'}
                </span>
              </div>
              {clusterHealth[currentCluster] && (
                <div className="settings-cluster-info">
                  <span>{clusterHealth[currentCluster].host}</span>
                  <span className="sep">|</span>
                  <span>v{clusterHealth[currentCluster].version}</span>
                  <span className="sep">|</span>
                  <span>{clusterHealth[currentCluster].nodeCount} nodes</span>
                </div>
              )}

              {/* Switch Cluster */}
              {clusters.length > 0 && (
                <>
                  <div className="settings-section-label" style={{ marginTop: 20 }}>Switch Cluster</div>
                  <div className="settings-cluster-list">
                    {clusters.map(name => {
                      const health = clusterHealth[name];
                      return (
                        <div
                          key={name}
                          className={`settings-cluster-option ${name === currentCluster ? 'active' : ''}`}
                          onClick={() => handleClusterChange(name)}
                        >
                          <span className={`cluster-status-dot ${health?.healthy ? 'healthy' : 'unhealthy'}`} />
                          <span>{name}</span>
                          {name === currentCluster && <span className="settings-check">✓</span>}
                          {health && (
                            <span className="settings-cluster-meta">
                              {health.host} · v{health.version}
                            </span>
                          )}
                        </div>
                      );
                    })}
                  </div>
                </>
              )}

              <hr className="settings-divider" />

              {/* Administration */}
              <div className="settings-section-label">Administration</div>
              <div className="settings-action-list">
                <div className="settings-action" onClick={() => { setSettingsOpen(false); handleMenuClick('clusters'); }}>
                  <FaServer size={18} />
                  <div className="settings-action-text">
                    <span className="settings-action-title">Manage Clusters</span>
                    <span className="settings-action-desc">Add, edit, or remove cluster connections</span>
                  </div>
                </div>
                <div className="settings-action" onClick={() => setShowPasswordModal(true)}>
                  <FaLock size={18} />
                  <div className="settings-action-text">
                    <span className="settings-action-title">Change Password</span>
                    <span className="settings-action-desc">Update your login password</span>
                  </div>
                </div>
                <div className="settings-action disabled" title="Coming soon">
                  <FaUserCheck size={18} />
                  <div className="settings-action-text">
                    <span className="settings-action-title">User Management</span>
                    <span className="settings-action-desc">Manage user accounts and permissions</span>
                  </div>
                </div>
              </div>

              <hr className="settings-divider" />

              {/* Sign Out */}
              <button className="settings-signout" onClick={handleLogout}>
                <FiLogOut size={18} />
                <span>Sign Out</span>
              </button>
            </div>
          </div>
        </div>
      )}

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
