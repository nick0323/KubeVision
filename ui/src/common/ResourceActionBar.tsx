import React, { useState } from 'react';
import { FaSync, FaClipboardList, FaTrash, FaPowerOff, FaMinus, FaPlus } from 'react-icons/fa';
import './ResourceActionBar.css';

export interface ResourceActionBarProps {
  name: string;
  namespace?: string;
  onRefresh: () => void;
  onDelete: () => void;
  onDescribe?: () => void;
  onScaleUp?: () => void;
  onScaleDown?: () => void;
  currentReplicas?: number;
  onRestart?: () => void;
}

export const ResourceActionBar: React.FC<ResourceActionBarProps> = ({
  name,
  namespace,
  onRefresh,
  onDelete,
  onDescribe,
  onScaleUp,
  onScaleDown,
  currentReplicas,
  onRestart,
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
            {onDescribe && (
              <button className="action-btn" onClick={onDescribe} title="View details">
                <FaClipboardList />
              </button>
            )}
            {onScaleUp && onScaleDown && currentReplicas !== undefined && (
              <span className="scale-controls">
                <button className="action-btn" onClick={onScaleDown} title="Scale down">
                  <FaMinus />
                </button>
                <span className="scale-value">{currentReplicas}</span>
                <button className="action-btn" onClick={onScaleUp} title="Scale up">
                  <FaPlus />
                </button>
              </span>
            )}
            {onRestart && (
              <button className="action-btn" onClick={onRestart} title="Rolling Restart">
                <FaPowerOff />
              </button>
            )}
            <button
              className={`action-btn ${refreshing ? 'spinning' : ''}`}
              onClick={handleRefresh}
              title="Refresh"
              disabled={refreshing}
            >
              <FaSync />
            </button>
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
