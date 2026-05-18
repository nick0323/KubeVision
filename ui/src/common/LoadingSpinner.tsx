import React from 'react';
import './LoadingSpinner.css';

/**
 * Loading... component Props
 */
interface LoadingSpinnerProps {
  type?: 'spinner' | 'skeleton' | 'pulse' | 'progress';
  text?: string;
  size?: 'sm' | 'md' | 'lg';
  className?: string;
  overlay?: boolean;
}

/**
 * Skeleton screenComponent Props
 */
interface SkeletonLoaderProps {
  rows?: number;
  className?: string;
}

/**
 * TableSkeleton screenComponent Props
 */
interface TableSkeletonProps {
  rows?: number;
  columns?: number;
  className?: string;
}

/**
 * CardSkeleton screenComponent Props
 */
interface CardSkeletonProps {
  className?: string;
}

/**
 * Loading... component
 * Keep with LoadingSpinner.jsx exactly the same functionality
 */
const LoadingSpinnerImpl: React.FC<LoadingSpinnerProps> = ({
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
 * Skeleton screenLoadComponent
 * Keep withoriginalComponentexactly the same functionality
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
 * Common DIV Skeleton screenComponent（for非Table场景）
 */
export const DivSkeleton: React.FC<TableSkeletonProps> = ({
  rows = 5,
  columns = 4,
  className = '',
}) => {
  return (
    <div className={`table-skeleton ${className}`}>
      {/* Table header skeleton */}
      <div className="skeleton-row skeleton-header">
        {Array.from({ length: columns }).map((_, index) => (
          <div key={index} className="skeleton skeleton-text"></div>
        ))}
      </div>

      {/* Table row skeleton */}
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
 * CardSkeleton screenComponent
 * Keep withoriginalComponentexactly the same functionality
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

export const LoadingSpinner = React.memo(LoadingSpinnerImpl);
export default LoadingSpinner;
