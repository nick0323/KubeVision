/**
 * 可折叠区块组件
 * 允许用户折叠/展开信息区块
 */
import React, { useState } from 'react';
import './CollapsibleSection.css';

interface CollapsibleSectionProps {
  title: string;
  defaultExpanded?: boolean;
  children: React.ReactNode;
  icon?: string;
  className?: string;
}

export const CollapsibleSection: React.FC<CollapsibleSectionProps> = ({
  title,
  defaultExpanded = true,
  children,
  icon,
  className = '',
}) => {
  const [isExpanded, setIsExpanded] = useState(defaultExpanded);

  return (
    <div className={`collapsible-section ${className}`}>
      <div
        className="collapsible-header"
        onClick={() => setIsExpanded(!isExpanded)}
      >
        <div className="collapsible-title">
          {icon && <span className="collapsible-icon">{icon}</span>}
          <span className="collapsible-text">{title}</span>
        </div>
        <div className="collapsible-actions">
          <span className={`collapse-arrow ${isExpanded ? 'expanded' : ''}`}>
            {isExpanded ? '▼' : '▶'}
          </span>
        </div>
      </div>
      {isExpanded && (
        <div className="collapsible-content">
          {children}
        </div>
      )}
    </div>
  );
};

export default CollapsibleSection;
