import React, { useState, useCallback, useEffect, useRef } from 'react';
import { useNavigate } from 'react-router-dom';
import { authUtils, authFetch } from '../utils/auth';
import apiClient from '../utils/apiClient';
import { STORAGE_KEYS } from '../constants';
import { ClusterHealth } from '../types';
import { FiSettings } from 'react-icons/fi';
import SidebarMenu from './SidebarMenu';
import ClusterSelector from './ClusterSelector';
import SettingsModal from './SettingsModal';

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

export const Sidebar: React.FC<SidebarProps> = ({
  activeTab,
  onMenuClick,
  showActiveState = true,
  collapsed,
}) => {
  const navigate = useNavigate();
  const [openGroups, setOpenGroups] = useState<Record<string, boolean>>({});
  const [clusters, setClusters] = useState<string[]>([]);
  const [clusterHealth, setClusterHealth] = useState<Record<string, ClusterHealth>>({});
  const [clusterOpen, setClusterOpen] = useState(false);
  const [settingsOpen, setSettingsOpen] = useState(false);
  const clusterRef = useRef<HTMLDivElement>(null);
  const currentCluster = localStorage.getItem(STORAGE_KEYS.CURRENT_CLUSTER) || 'default';

  useEffect(() => {
    apiClient.get<ClusterHealth[]>('/api/v1/clusters/health').then(res => {
      const list = res.data || [];
      const healthMap: Record<string, ClusterHealth> = {};
      list.forEach(h => { healthMap[h.name] = h; });
      setClusterHealth(healthMap);
      setClusters(list.map(h => h.name));
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

  const toggleGroup = useCallback((group: string) => {
    setOpenGroups(prev => ({ ...prev, [group]: !prev[group] }));
  }, []);

  const handleMenuClick = useCallback(
    (key: string) => {
      localStorage.setItem('current_tab', JSON.stringify(key));
      navigate('/?tab=' + encodeURIComponent(key));
    },
    [navigate]
  );

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
      <SidebarMenu
        collapsed={collapsed}
        openGroups={openGroups}
        activeTab={activeTab}
        showActiveState={showActiveState}
        onMenuClick={handleMenuClick}
        onToggleGroup={toggleGroup}
      />

      <div className="sider-bottom">
        {!collapsed && (
          <>
            <ClusterSelector
              clusters={clusters}
              clusterHealth={clusterHealth}
              currentCluster={currentCluster}
              clusterOpen={clusterOpen}
              onToggle={() => setClusterOpen(o => !o)}
              onChange={handleClusterChange}
              innerRef={clusterRef}
            />
            <div className="settings-area">
              <div className="settings-trigger" onClick={() => setSettingsOpen(true)}>
                <span className="icon"><FiSettings /></span>
                <span>Settings</span>
              </div>
            </div>
          </>
        )}
        {collapsed && (
          <button className="logout-btn" onClick={() => setSettingsOpen(true)} title="Settings">
            <span className="icon"><FiSettings /></span>
          </button>
        )}
      </div>

      <SettingsModal
        open={settingsOpen}
        onClose={() => setSettingsOpen(false)}
        onManageClusters={() => { setSettingsOpen(false); handleMenuClick('clusters'); }}
        onLogout={handleLogout}
      />
    </div>
  );
};

export default Sidebar;
