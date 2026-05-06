import React, { useState } from 'react';
import { FaSync, FaClipboardList, FaTrash } from 'react-icons/fa';
import './ResourceActionBar.css';

export interface ResourceActionBarProps {
  name: string;
  namespace?: string; // Optional, cluster resources have no namespace
  onRefresh: () => void;
  onDelete: () => void;
  onDescribe?: () => void; // Optional, needed for Pod detail page
}

/**
 * Resource action bar - CommonComponent
 */
export const ResourceActionBar: React.FC<ResourceActionBarProps> = ({
  name,
  namespace,
  onRefresh,
  onDelete,
  onDescribe,
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
              title="Refresh"
              disabled={refreshing}
            >
              <FaSync />
            </button>
            {onDescribe && (
              <button className="action-btn" onClick={onDescribe} title="View details">
                <FaClipboardList />
              </button>
            )}
            <button className="action-btn danger" onClick={onDelete} title="Delete">
              <FaTrash />
            </button>
          </div>
        </div>
        {namespace && (
          <div className="resource-namespace">
            <span className="namespace-label">Namespace:</span>
            <span className="namespace-value">{namespace}</span>
          </div>
        )}
      </div>
    </div>
  );
};

export default ResourceActionBar;
