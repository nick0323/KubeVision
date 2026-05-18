import React from 'react';
import './TabNavigation.css';

export interface TabItem {
  key: string;
  label: string;
}

export interface TabNavigationProps {
  tabs: TabItem[];
  activeTab: string;
  onTabChange: (tabKey: string) => void;
}

/**
 * Tab navigationComponent - Common版本
 */
const TabNavigationImpl: React.FC<TabNavigationProps> = ({ tabs, activeTab, onTabChange }) => {
  return (
    <div className="tab-list">
      {tabs.map(tab => (
        <button
          key={tab.key}
          className={`tab-item ${activeTab === tab.key ? 'active' : ''}`}
          onClick={() => onTabChange(tab.key)}
        >
          <span className="tab-label">{tab.label}</span>
        </button>
      ))}
    </div>
  );
};

export const TabNavigation = React.memo(TabNavigationImpl);
export default TabNavigation;
