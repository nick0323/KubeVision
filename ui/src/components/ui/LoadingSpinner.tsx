import React from 'react';
import './LoadingSpinner.css';

/**
 * 加载动画组件 Props
 */
interface LoadingSpinnerProps {
  type?: 'spinner' | 'skeleton' | 'pulse' | 'progress';
  text?: string;
  size?: 'sm' | 'md' | 'lg';
  className?: string;
  overlay?: boolean;
}

/**
 * 骨架屏组件 Props
 */
interface SkeletonLoaderProps {
  rows?: number;
  className?: string;
}

/**
 * 表格骨架屏组件 Props
 */
interface TableSkeletonProps {
  rows?: number;
  columns?: number;
  className?: string;
}

/**
 * 卡片骨架屏组件 Props
 */
interface CardSkeletonProps {
  className?: string;
}

/**
 * 加载动画组件
 * 保持与 LoadingSpinner.jsx 完全一致的功能
 */
export const LoadingSpinner: React.FC<LoadingSpinnerProps> = ({
  type = 'spinner',
  text = 'Loading...',
  size = 'md',
  className = '',
  overlay = false,
}) => {
  const sizeClasses = {
    sm: 'loading-sm',
    md: 'loading-md',
    lg: 'loading-lg',
  };

  const renderLoadingContent = () => {
    const content = (() => {
      switch (type) {
        case 'skeleton':
          return (
            <div className={`skeleton-container ${sizeClasses[size]} ${className}`}>
              <div className="skeleton skeleton-title"></div>
              <div className="skeleton skeleton-text"></div>
              <div className="skeleton skeleton-text"></div>
              <div className="skeleton skeleton-text"></div>
            </div>
          );

        case 'pulse':
          return (
            <div className={`loading-container ${className}`}>
              <div className={`pulse ${sizeClasses[size]}`}>
                <div className="loading-spinner"></div>
              </div>
              {text && <div className="loading-text">{text}</div>}
            </div>
          );

        case 'progress':
          return (
            <div className={`loading-container ${className}`}>
              <div className="progress-bar">
                <div className="progress-fill"></div>
              </div>
              {text && <div className="loading-text">{text}</div>}
            </div>
          );

        case 'spinner':
        default:
          return (
            <div className={`loading-container ${className}`}>
              <div className={`loading-spinner ${sizeClasses[size]}`}></div>
              {text && <div className="loading-text">{text}</div>}
            </div>
          );
      }
    })();

    if (overlay) {
      return <div className="loading-overlay">{content}</div>;
    }
    return content;
  };

  return renderLoadingContent();
};

/**
 * 骨架屏加载组件
 * 保持与原始组件完全一致的功能
 */
export const SkeletonLoader: React.FC<SkeletonLoaderProps> = ({ rows = 3, className = '' }) => {
  return (
    <div className={`skeleton-loader ${className}`}>
      {Array.from({ length: rows }).map((_, index) => (
        <div key={index} className="skeleton skeleton-text"></div>
      ))}
    </div>
  );
};

/**
 * 表格骨架屏组件
 * 保持与原始组件完全一致的功能
 */
export const TableSkeleton: React.FC<TableSkeletonProps> = ({
  rows = 5,
  columns = 4,
  className = '',
}) => {
  return (
    <div className={`table-skeleton ${className}`}>
      {/* 表头骨架 */}
      <div className="skeleton-row skeleton-header">
        {Array.from({ length: columns }).map((_, index) => (
          <div key={index} className="skeleton skeleton-text"></div>
        ))}
      </div>

      {/* 表行骨架 */}
      {Array.from({ length: rows }).map((_, rowIndex) => (
        <div key={rowIndex} className="skeleton-row">
          {Array.from({ length: columns }).map((_, colIndex) => (
            <div key={colIndex} className="skeleton skeleton-text"></div>
          ))}
        </div>
      ))}
    </div>
  );
};

/**
 * 卡片骨架屏组件
 * 保持与原始组件完全一致的功能
 */
export const CardSkeleton: React.FC<CardSkeletonProps> = ({ className = '' }) => {
  return (
    <div className={`card-skeleton ${className}`}>
      <div className="skeleton skeleton-title"></div>
      <div className="skeleton skeleton-text"></div>
      <div className="skeleton skeleton-text"></div>
      <div className="skeleton skeleton-button"></div>
    </div>
  );
};

export default LoadingSpinner;
