import React from 'react';
import { Sidebar } from './Sidebar';

/**
 * 侧边栏布局 Props
 */
export interface SidebarLayoutProps {
  children: React.ReactNode | ((props: { collapsed: boolean; onToggleCollapsed: () => void }) => React.ReactNode);
  activeTab?: string;
  onMenuClick?: (key: string) => void;
  showActiveState?: boolean;
}

/**
 * 侧边栏布局组件
 *
 * 统一替代 App.tsx 中的 3 处 layout 代码
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

  // 支持 Render Props 模式
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
