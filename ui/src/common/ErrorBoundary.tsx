import React, { Component, ReactNode } from 'react';
import { logError } from '../utils/errorHandler';

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
    logError(error, 'ErrorBoundary');
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
             <h3 className="error-title">Application Error</h3>
             <p>Sorry, the application encountered an unexpected error.</p>

            <div className="error-actions">
              <button onClick={this.handleRetry} className="retry-btn">
                Retry
              </button>
              <button onClick={this.handleReload} className="reload-btn">
                Reload
              </button>
            </div>

            {error && (
              <details className="error-details">
                <summary>Error Details</summary>
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
