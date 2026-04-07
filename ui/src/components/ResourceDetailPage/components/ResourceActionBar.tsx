import React, { useState } from 'react';
import { FaSync, FaTrash } from 'react-icons/fa';
import './ResourceActionBar.css';

export interface ResourceActionBarProps {
  name: string;
  namespace: string;
  onRefresh: () => void;
  onDelete: () => void;
}

/**
 * 资源操作栏
 */
export const ResourceActionBar: React.FC<ResourceActionBarProps> = ({
  name,
  namespace,
  onRefresh,
  onDelete,
}) => {
  const [refreshing, setRefreshing] = useState(false);

  const handleRefresh = async () => {
    setRefreshing(true);
    try {
      await onRefresh();
    } finally {
      setRefreshing(false);
    }
  };

  return (
    <div className="resource-action-bar">
      <div className="resource-info">
        <div className="resource-name-row">
          <h2 className="resource-name">{name}</h2>
          <div className="resource-actions">
            <button
              className={`action-btn ${refreshing ? 'spinning' : ''}`}
              onClick={handleRefresh}
              title="刷新"
              disabled={refreshing}
            >
              <FaSync />
            </button>
            <button className="action-btn danger" onClick={onDelete} title="删除">
              <FaTrash />
            </button>
          </div>
        </div>
        <div className="resource-namespace">
          <span className="namespace-label">Namespace:</span>
          <span className="namespace-value">{namespace}</span>
        </div>
      </div>
    </div>
  );
};

export default ResourceActionBar;
