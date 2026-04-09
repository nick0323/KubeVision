import React from 'react';
import { ErrorDisplayProps } from '../types';
import './ErrorDisplay.css';

/**
 * 错误显示组件
 */
export const ErrorDisplay: React.FC<ErrorDisplayProps> = ({
  message,
  type = 'error',
  onRetry,
  showRetry = false,
}) => {
  return (
    <div className={`error-display ${type}`}>
      <div className="error-display-icon">
        {type === 'warning' ? '⚠️' : type === 'info' ? 'ℹ️' : '❌'}
      </div>
      <div className="error-display-content">
        <div className="error-display-title">
          {type === 'warning' ? '警告' : type === 'info' ? '提示' : '错误'}
        </div>
        <div className="error-display-message">{message}</div>
        {showRetry && onRetry && (
          <button className="error-display-retry" onClick={onRetry}>
            重试
          </button>
        )}
      </div>
    </div>
  );
};

export default ErrorDisplay;
