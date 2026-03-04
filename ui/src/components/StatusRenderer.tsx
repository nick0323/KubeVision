import React from 'react';

interface StatusMap {
  [key: string]: string;
}

interface StatusTagProps {
  value: string;
  className?: string;
  size?: 'default' | 'small' | 'large';
}

/**
 * 创建状态渲染器
 */
export function createStatusRenderer(statusMap: StatusMap = {}) {
  return (value: string, row?: any, index?: number, isTooltip?: boolean) => {
    if (isTooltip) return value;

    if (!value) return '-';

    // 使用配置的状态映射
    if (statusMap[value]) {
      return (
        <span className={`status-tag ${statusMap[value]}`}>
          {value}
        </span>
      );
    }

    // 默认状态判断逻辑
    const isHealthy = value === 'Running' || value === 'Succeeded' || value === 'Ready' ||
                     value === 'Healthy' || value === 'Normal' || value === 'Active' ||
                     value === 'Bound' || value === 'Available';
    const isFailed = value === 'Failed' || value === 'Error' || value === 'CrashLoopBackOff' ||
                     value === 'Unhealthy' || value === 'Warning';
    const isPending = value === 'Pending' || value === 'ContainerCreating' ||
                     value === 'PodInitializing' || value === 'Creating';

    let statusClass = 'status-running';
    if (isHealthy) {
      statusClass = 'status-ready';
    } else if (isFailed) {
      statusClass = 'status-failed';
    } else if (isPending) {
      statusClass = 'status-pending';
    }

    return (
      <span className={`status-tag ${statusClass}`}>
        {value}
      </span>
    );
  };
}

/**
 * 通用状态标签组件
 */
export const StatusTag: React.FC<StatusTagProps> = ({ value, className = '', size = 'default' }) => {
  if (!value) return null;

  const isHealthy = value === 'Running' || value === 'Succeeded' || value === 'Ready' ||
                   value === 'Healthy' || value === 'Normal' || value === 'Active' ||
                   value === 'Bound' || value === 'Available';
  const isFailed = value === 'Failed' || value === 'Error' || value === 'CrashLoopBackOff' ||
                   value === 'Unhealthy' || value === 'Warning';
  const isPending = value === 'Pending' || value === 'ContainerCreating' ||
                   value === 'PodInitializing' || value === 'Creating';

  let statusClass = 'status-running';
  if (isHealthy) {
    statusClass = 'status-ready';
  } else if (isFailed) {
    statusClass = 'status-failed';
  } else if (isPending) {
    statusClass = 'status-pending';
  }

  return (
    <span className={`status-tag ${statusClass} ${className} ${size}`}>
      {value}
    </span>
  );
};

export default createStatusRenderer;
