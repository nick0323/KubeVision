import React, { Component, ReactNode } from 'react';

interface ErrorBoundaryState {
  hasError: boolean;
  error: Error | null;
}

interface ErrorBoundaryProps {
  children: ReactNode;
  fallback?: React.ComponentType<{
    error: Error | null;
    onRetry: () => void;
  }>;
}

/**
 * 简化的全局错误边界组件
 * 改进：
 * 1. 移除过度复杂的错误报告
 * 2. 简化 UI 显示
 * 3. 移除未使用的功能
 */
export class ErrorBoundary extends Component<ErrorBoundaryProps, ErrorBoundaryState> {
  constructor(props: ErrorBoundaryProps) {
    super(props);
    this.state = {
      hasError: false,
      error: null,
    };
  }

  static getDerivedStateFromError(error: Error): ErrorBoundaryState {
    return {
      hasError: true,
      error,
    };
  }

  componentDidCatch(error: Error) {
    console.error('ErrorBoundary caught an error:', error);
  }

  handleRetry = () => {
    this.setState({
      hasError: false,
      error: null,
    });
  };

  handleReload = () => {
    window.location.reload();
  };

  render() {
    if (this.state.hasError) {
      const { error } = this.state;
      const { fallback: Fallback } = this.props;

      if (Fallback) {
        return <Fallback error={error} onRetry={this.handleRetry} />;
      }

      return (
        <div className="error-boundary">
          <div className="error-content">
            <h3>⚠️ 应用程序出现错误</h3>
            <p>抱歉，应用程序遇到了一个意外错误。</p>

            <div className="error-actions">
              <button onClick={this.handleRetry} className="retry-btn">
                🔄 重试
              </button>
              <button onClick={this.handleReload} className="reload-btn">
                🔄 重新加载
              </button>
            </div>

            {error && (
              <details className="error-details">
                <summary>错误详情</summary>
                <pre>{error.toString()}</pre>
              </details>
            )}
          </div>
        </div>
      );
    }

    return this.props.children;
  }
}

export default ErrorBoundary;
