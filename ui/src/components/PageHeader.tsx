import React from 'react';
import { FaBars } from 'react-icons/fa';
import { PageHeaderProps } from '../types';
import './PageHeader.css';

/**
 * 页面头部组件
 */
export const PageHeader: React.FC<PageHeaderProps> = ({
  title,
  children,
  collapsed,
  onToggleCollapsed
}) => {
  return (
    <div className="page-header">
      <div className="page-header-left">
        <h1 className="page-title">{title}</h1>
        <span className="page-separator">|</span>
        <button
          className="page-collapse-btn"
          onClick={onToggleCollapsed}
          aria-label={collapsed ? '展开侧边栏' : '折叠侧边栏'}
          title={collapsed ? '展开侧边栏' : '折叠侧边栏'}
        >
          <FaBars />
        </button>
      </div>
      <div className="page-header-right">
        {children}
      </div>
    </div>
  );
};

export default PageHeader;
