import React from 'react';
import { FaBars } from 'react-icons/fa';
import { Breadcrumb } from './Breadcrumb';
import { PageHeaderProps } from '../types';
import './PageHeader.css';

export interface PageHeaderWithBreadcrumbProps extends PageHeaderProps {
  breadcrumbs?: Array<{ label: string; path: string }>;
  onBreadcrumbClick?: (path: string) => void;
}

/**
 * 页面头部组件
 */
export const PageHeader: React.FC<PageHeaderWithBreadcrumbProps> = ({
  title,
  children,
  collapsed,
  onToggleCollapsed,
  breadcrumbs,
  onBreadcrumbClick,
}) => {
  return (
    <div className="page-header">
      <div className="page-header-left">
        <button
          className="page-collapse-btn"
          onClick={onToggleCollapsed}
          aria-label={collapsed ? '展开侧边栏' : '折叠侧边栏'}
          title={collapsed ? '展开侧边栏' : '折叠侧边栏'}
        >
          <FaBars />
        </button>
        <span className="page-separator">|</span>
        {breadcrumbs && breadcrumbs.length > 0 ? (
          <Breadcrumb items={breadcrumbs} onItemClick={onBreadcrumbClick || (() => {})} />
        ) : (
          <h1 className="page-title">{title}</h1>
        )}
      </div>
      <div className="page-header-right">
        {children}
      </div>
    </div>
  );
};

export default PageHeader;
