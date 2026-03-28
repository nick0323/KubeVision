import React, { useState, useCallback } from 'react';
import './CollapsibleSection.css';

export interface CollapsibleSectionProps {
  title: string;
  defaultCollapsed?: boolean;
  children: React.ReactNode;
  className?: string;
}

/**
 * 可折叠面板组件
 */
export const CollapsibleSection: React.FC<CollapsibleSectionProps> = ({
  title,
  defaultCollapsed = false,
  children,
  className = '',
}) => {
  const [collapsed, setCollapsed] = useState(defaultCollapsed);

  const handleToggle = useCallback(() => {
    setCollapsed(prev => !prev);
  }, []);

  return (
    <div className={`collapsible-section ${className} ${collapsed ? 'collapsed' : ''}`}>
      <div className="collapsible-header" onClick={handleToggle}>
        <span className="collapsible-title">{title}</span>
        <span className="collapsible-icon">{collapsed ? '▶' : '▼'}</span>
      </div>
      {!collapsed && (
        <div className="collapsible-content">
          {children}
        </div>
      )}
    </div>
  );
};

export default CollapsibleSection;
