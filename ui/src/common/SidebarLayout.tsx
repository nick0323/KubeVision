import React from 'react';
import { Sidebar } from './Sidebar';

/**
 * Sidebarlayout Props
 */
export interface SidebarLayoutProps {
  children: React.ReactNode | ((props: { collapsed: boolean; onToggleCollapsed: () => void }) => React.ReactNode);
  activeTab?: string;
  onMenuClick?: (key: string) => void;
  showActiveState?: boolean;
}

/**
 * SidebarlayoutComponent
 *
 * UnifiedAlternative App.tsx 's 3 处 layout code
 */
export const SidebarLayout: React.FC<SidebarLayoutProps> = ({
  children,
  activeTab,
  onMenuClick,
  showActiveState = true,
}) => {
  const [collapsed, setCollapsed] = React.useState(false);

  const onToggleCollapsed = React.useCallback(() => {
    setCollapsed(prev => {
      const next = !prev;
      localStorage.setItem('sider_collapsed', JSON.stringify(next));
      return next;
    });
  }, []);

  // Support Render Props Mode
  const content = typeof children === 'function'
    ? children({ collapsed, onToggleCollapsed })
    : children;

  return (
    <div className="layout-root" data-sider-collapsed={collapsed}>
      <Sidebar
        activeTab={activeTab}
        onMenuClick={onMenuClick}
        showActiveState={showActiveState}
        collapsed={collapsed}
        onToggleCollapsed={onToggleCollapsed}
      />
      <div className="main-content">{content}</div>
    </div>
  );
};

export default SidebarLayout;
