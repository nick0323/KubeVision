import React, { useState } from 'react';
import { FaSync, FaClipboardList, FaTrash } from 'react-icons/fa';
import './ResourceActionBar.css';

export interface ResourceActionBarProps {
  name: string;
  namespace?: string;  // 可选，集群资源没有 namespace
  onRefresh: () => void;
  onDelete: () => void;
  onDescribe?: () => void;  // 可选，Pod 详情页需要
}

/**
 * 资源操作栏 - 通用组件
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
              title="刷新"
              disabled={refreshing}
            >
              <FaSync />
            </button>
            {onDescribe && (
              <button className="action-btn" onClick={onDescribe} title="查看详情">
                <FaClipboardList />
              </button>
            )}
            <button className="action-btn danger" onClick={onDelete} title="删除">
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
