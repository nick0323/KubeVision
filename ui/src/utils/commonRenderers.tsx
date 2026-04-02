import React from 'react';
import { createStatusRenderer as createStatusRendererWrapper } from '../components/StatusRenderer';

/**
 * 创建名称渲染器（带点击链接）
 */
export const createNameRenderer = (onClick: (record: any) => void) => {
  return (text: string, record: any) => (
    <span
      className="resource-name-link"
      onClick={() => onClick(record)}
      style={{ cursor: 'pointer', color: '#1890ff' }}
    >
      {text}
    </span>
  );
};

/**
 * 创建 Namespace 渲染器
 */
export const createNamespaceRenderer = () => {
  return (text: string) => (
    <span className="namespace-tag" style={{ color: '#722ed1' }}>
      {text}
    </span>
  );
};

/**
 * 创建标签渲染器
 */
export const createLabelsRenderer = () => {
  return (labels: Record<string, string>) => {
    if (!labels || Object.keys(labels).length === 0) return '-';

    return (
      <div className="labels-container">
        {Object.entries(labels).map(([key, value]) => (
          <span key={key} className="label-tag">
            {key}: {value}
          </span>
        ))}
      </div>
    );
  };
};

/**
 * 创建时间渲染器
 */
export const createTimeRenderer = (format: 'relative' | 'absolute' = 'relative') => {
  return (text: string) => {
    if (!text) return '-';

    if (format === 'absolute') {
      return <span className="timestamp">{text}</span>;
    }

    // 相对时间
    const date = new Date(text);
    const now = new Date();
    const diffMs = now.getTime() - date.getTime();
    const diffSec = Math.floor(diffMs / 1000);
    const diffMin = Math.floor(diffSec / 60);
    const diffHour = Math.floor(diffMin / 60);
    const diffDay = Math.floor(diffHour / 24);

    if (diffSec < 60) return <span className="timestamp">Just now</span>;
    if (diffMin < 60) return <span className="timestamp">{diffMin}m ago</span>;
    if (diffHour < 24) return <span className="timestamp">{diffHour}h ago</span>;
    return <span className="timestamp">{diffDay}d ago</span>;
  };
};

/**
 * 创建详细时间渲染器
 */
export const createDetailedTimeRenderer = () => {
  return (text: string) => {
    if (!text) return '-';
    return <span className="detailed-timestamp">{text}</span>;
  };
};

/**
 * 创建数字渲染器
 */
export const createNumberRenderer = (suffix = '', formatter?: (value: number) => string) => {
  return (value: number) => {
    if (value === undefined || value === null) return '-';

    if (formatter) {
      return (
        <span className="number-value">
          {formatter(value)} {suffix}
        </span>
      );
    }

    return (
      <span className="number-value">
        {value} {suffix}
      </span>
    );
  };
};

/**
 * 创建布尔值渲染器
 */
export const createBooleanRenderer = (trueText = 'Yes', falseText = 'No') => {
  return (value: boolean) => {
    if (value === undefined || value === null) return '-';

    const displayValue = value ? trueText : falseText;
    const className = value ? 'boolean-true' : 'boolean-false';

    return <span className={`boolean-value ${className}`}>{displayValue}</span>;
  };
};

/**
 * 创建数组渲染器
 */
export const createArrayRenderer = (separator = ', ', maxItems = 999) => {
  return (array: any[]) => {
    if (!Array.isArray(array) || array.length === 0) return '-';

    const displayItems = array.slice(0, maxItems);
    const text = displayItems.join(separator);

    return <span className="array-value">{text}</span>;
  };
};

/**
 * 创建唯一数组渲染器（去重）
 */
export const createUniqueArrayRenderer = (separator = ', ', maxItems = 999) => {
  return (array: any[]) => {
    if (!Array.isArray(array) || array.length === 0) return '-';

    const uniqueItems = [...new Set(array.filter(item => item && item.toString().trim()))];
    const displayItems = uniqueItems.slice(0, maxItems);
    const text = displayItems.join(separator);

    return <span className="array-value">{text}</span>;
  };
};

/**
 * 创建使用量渲染器
 */
export const createUsageRenderer = () => {
  return (value: string | number) => {
    if (!value) return '-';
    return <span className="usage-value">{value}</span>;
  };
};

/**
 * 创建状态渲染器（导出）
 */
export const createStatusRenderer = createStatusRendererWrapper;

export default {
  createNameRenderer,
  createNamespaceRenderer,
  createLabelsRenderer,
  createTimeRenderer,
  createDetailedTimeRenderer,
  createNumberRenderer,
  createBooleanRenderer,
  createArrayRenderer,
  createUniqueArrayRenderer,
  createUsageRenderer,
  createStatusRenderer,
};
