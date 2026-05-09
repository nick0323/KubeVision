import React from 'react';
import { ErrorDisplayProps } from '../types';
import './ErrorDisplay.css';

/**
 * Error display component
 */
export const ErrorDisplay: React.FC<ErrorDisplayProps> = ({
  message,
  type = 'error',
  onRetry,
  showRetry = false,
}) => {
  return (
    <div className={`error-display ${type}`}>
      <div className="error-display-content">
        <div className="error-display-title">
           {type === 'warning' ? 'Warning' : type === 'info' ? 'Info' : 'Error'}
        </div>
        <div className="error-display-message">{message}</div>
        {showRetry && onRetry && (
          <button className="error-display-retry" onClick={onRetry}>
            Retry
          </button>
        )}
      </div>
    </div>
  );
};

export default ErrorDisplay;
