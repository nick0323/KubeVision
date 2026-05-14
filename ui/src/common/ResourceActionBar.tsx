import React, { useState, useCallback } from 'react';
import { FaSync, FaClipboardList, FaTrash, FaPowerOff, FaMinus, FaPlus, FaTimes, FaExclamationTriangle } from 'react-icons/fa';
import { useConfirm } from '../hooks/useConfirm';
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
  const { confirm, confirming, config, onConfirm, onCancel } = useConfirm();

  const handleRefresh = async () => {
    setRefreshing(true);
    try {
      await onRefresh();
    } finally {
      setRefreshing(false);
    }
  };

  const handleDelete = useCallback(async () => {
    const result = await confirm({
      title: 'Confirm Delete',
      message: `Are you sure you want to delete "${name}"? This action cannot be undone.`,
      confirmText: 'Delete',
      cancelText: 'Cancel',
      danger: true,
    });

    if (result.confirmed) {
      onDelete();
    }
  }, [confirm, name, onDelete]);

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
            <button className="action-btn danger" onClick={handleDelete} title="Delete">
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

      {/* Confirm Dialog Modal */}
      {confirming && config && (
        <div className="confirm-overlay" onClick={onCancel}>
          <div className="confirm-dialog" onClick={e => e.stopPropagation()}>
            <div className="confirm-header">
              <h3 className="confirm-title">
                <FaExclamationTriangle className="confirm-icon" />
                {config.title || 'Confirm'}
              </h3>
              <button className="confirm-close" onClick={onCancel}>
                <FaTimes />
              </button>
            </div>
            <div className="confirm-body">
              <p className="confirm-message">{config.message}</p>
            </div>
            <div className="confirm-footer">
              <button className="confirm-btn cancel" onClick={onCancel}>
                {config.cancelText || 'Cancel'}
              </button>
              <button
                className={`confirm-btn ${config.danger ? 'danger' : 'primary'}`}
                onClick={onConfirm}
              >
                {config.confirmText || 'Confirm'}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
};

export default ResourceActionBar;
