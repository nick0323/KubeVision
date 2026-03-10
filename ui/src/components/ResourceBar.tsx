import React from 'react';
import './ResourceBar.css';

interface ResourceBarItem {
  label: string;
  value: string;
  percentage: number;
  type?: 'cpu' | 'memory' | 'pods';
}

interface ResourceBarProps {
  items: ResourceBarItem[];
}

export const ResourceBar: React.FC<ResourceBarProps> = ({ items }) => {
  return (
    <div className="resource-bar">
      {items.map((item, index) => (
        <div key={index} className="resource-bar-item">
          <div className="resource-bar-header">
            <span className="resource-bar-label">{item.label}</span>
            <span className="resource-bar-value">{item.value}</span>
          </div>
          <div className="progress-bar">
            <div
              className={`progress-fill ${item.type || ''} ${
                item.percentage > 80 ? 'warning' : item.percentage > 90 ? 'danger' : ''
              }`}
              style={{ width: `${item.percentage}%` }}
            />
          </div>
        </div>
      ))}
    </div>
  );
};

export default ResourceBar;
