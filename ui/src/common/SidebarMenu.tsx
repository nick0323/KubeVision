import React from 'react';
import { MENU_LIST } from '../constants';
import {
  FaChartPie, FaCube, FaRocket, FaTree, FaServer, FaCogs, FaBriefcase,
  FaClock, FaNetworkWired, FaDoorOpen, FaShieldAlt, FaUserSecret, FaUserTag,
  FaUserCheck, FaTachometerAlt, FaSlidersH, FaBalanceScale, FaPuzzlePiece,
  FaHdd, FaDatabase, FaListAlt, FaFileAlt, FaLock, FaBell, FaThLarge,
  FaChevronDown, FaChevronRight, FaSync, FaGitAlt, FaArrowsAltV,
} from 'react-icons/fa';
import k8sLogo from '../assets/kubernetes-logo.svg';

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

interface SidebarMenuProps {
  collapsed: boolean;
  openGroups: Record<string, boolean>;
  activeTab?: string;
  showActiveState: boolean;
  onMenuClick: (key: string) => void;
  onToggleGroup: (group: string) => void;
}

const SidebarMenu: React.FC<SidebarMenuProps> = ({
  collapsed, openGroups, activeTab, showActiveState, onMenuClick, onToggleGroup,
}) => (
  <>
    <div className="logo-area">
      <span className="logo-text-full">KubeVision</span>
      <img src={k8sLogo} alt="Kubernetes" className="logo-compact" />
    </div>

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
              onClick={() => onMenuClick(item.key)}
              onKeyDown={e => {
                if (e.key === 'Enter' || e.key === ' ') {
                  e.preventDefault();
                  onMenuClick(item.key);
                }
              }}
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
                    onToggleGroup(group.group);
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
                    onClick={() => onMenuClick(item.key)}
                    onKeyDown={e => {
                      if (e.key === 'Enter' || e.key === ' ') {
                        e.preventDefault();
                        onMenuClick(item.key);
                      }
                    }}
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
  </>
);

export default React.memo(SidebarMenu);
