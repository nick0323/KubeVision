/**
 * 快速操作区组件
 * 提供常用操作入口
 */
import React, { useState } from 'react';
import './QuickActions.css';

interface QuickAction {
  label: string;
  icon: string;
  action: () => void;
  danger?: boolean;
  disabled?: boolean;
  confirm?: boolean;
  confirmMessage?: string;
}

interface QuickActionsProps {
  resourceType: string;
  resourceName: string;
  namespace: string;
  data?: any;
  onRefresh?: () => void;
}

export const QuickActions: React.FC<QuickActionsProps> = ({
  resourceType,
  resourceName,
  namespace,
  data,
  onRefresh,
}) => {
  const [showDropdown, setShowDropdown] = useState(false);

  // 定义不同资源类型的操作
  const getActions = (): QuickAction[] => {
    const actions: QuickAction[] = [];

    switch (resourceType) {
      case 'pod':
      case 'pods':
        actions.push(
          {
            label: '查看日志',
            icon: '📜',
            action: () => console.log('查看日志', resourceName),
          },
          {
            label: '进入终端',
            icon: '💻',
            action: () => console.log('进入终端', resourceName),
          },
          {
            label: '重启 Pod',
            icon: '🔄',
            action: () => console.log('重启 Pod', resourceName),
            confirm: true,
            confirmMessage: `确定要重启 Pod "${resourceName}" 吗？`,
          },
          {
            label: '删除 Pod',
            icon: '🗑️',
            action: () => console.log('删除 Pod', resourceName),
            danger: true,
            confirm: true,
            confirmMessage: `确定要删除 Pod "${resourceName}" 吗？此操作不可恢复！`,
          }
        );
        break;

      case 'deployment':
      case 'deployments':
        actions.push(
          {
            label: '扩缩容',
            icon: '📈',
            action: () => console.log('扩缩容', resourceName),
          },
          {
            label: '重启',
            icon: '🔄',
            action: () => console.log('重启 Deployment', resourceName),
            confirm: true,
            confirmMessage: `确定要重启 Deployment "${resourceName}" 吗？`,
          },
          {
            label: '回滚',
            icon: '⏮️',
            action: () => console.log('回滚', resourceName),
          }
        );
        break;

      case 'node':
      case 'nodes':
        actions.push(
          {
            label: '隔离节点',
            icon: '🚫',
            action: () => console.log('隔离节点', resourceName),
            confirm: true,
            confirmMessage: `确定要隔离节点 "${resourceName}" 吗？`,
          },
          {
            label: '排水节点',
            icon: '💧',
            action: () => console.log('排水节点', resourceName),
            danger: true,
            confirm: true,
            confirmMessage: `确定要排水节点 "${resourceName}" 吗？这将迁移所有 Pod！`,
          }
        );
        break;

      default:
        actions.push({
          label: '刷新',
          icon: '🔄',
          action: onRefresh || (() => {}),
        });
    }

    return actions;
  };

  const actions = getActions();

  const handleAction = (action: QuickAction) => {
    if (action.confirm) {
      const message = action.confirmMessage || `确定要执行"${action.label}"吗？`;
      if (window.confirm(message)) {
        action.action();
      }
    } else {
      action.action();
    }
    setShowDropdown(false);
  };

  return (
    <div className="quick-actions">
      {/* 主操作按钮 */}
      <div className="quick-actions-main">
        {actions.slice(0, 2).map((action, index) => (
          <button
            key={index}
            className={`quick-action-btn ${action.danger ? 'btn-danger' : ''}`}
            onClick={() => handleAction(action)}
            disabled={action.disabled}
          >
            <span className="action-icon">{action.icon}</span>
            <span className="action-label">{action.label}</span>
          </button>
        ))}

        {/* 更多操作下拉菜单 */}
        {actions.length > 2 && (
          <div className="quick-actions-more">
            <button
              className="quick-action-btn more-btn"
              onClick={() => setShowDropdown(!showDropdown)}
            >
              <span className="action-icon">⋮</span>
              <span className="action-label">更多</span>
            </button>

            {showDropdown && (
              <div className="quick-actions-dropdown">
                {actions.slice(2).map((action, index) => (
                  <button
                    key={index}
                    className={`dropdown-item ${action.danger ? 'danger' : ''}`}
                    onClick={() => handleAction(action)}
                  >
                    <span className="dropdown-icon">{action.icon}</span>
                    <span className="dropdown-label">{action.label}</span>
                  </button>
                ))}
              </div>
            )}
          </div>
        )}
      </div>
    </div>
  );
};

export default QuickActions;
