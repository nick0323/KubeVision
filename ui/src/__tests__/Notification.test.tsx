/**
 * Notification Component Test
 */

import { describe, it, expect, jest, beforeEach } from '@jest/globals';
import { render, screen, fireEvent, act } from '@testing-library/react';
import { Notification, NotificationContainer, useNotification, notification } from '../common/Notification';

describe('Notification Component', () => {
  const mockOnClose = jest.fn();

  beforeEach(() => {
    mockOnClose.mockClear();
    jest.useFakeTimers();
  });

  afterEach(() => {
    jest.useRealTimers();
  });

  describe('Notification', () => {
    it('should render success notification', () => {
      render(
        <Notification
          type="success"
          message="Operation successful"
          onClose={mockOnClose}
        />
      );

      expect(screen.getByText('Operation successful')).toBeInTheDocument();
      expect(screen.getByRole('button')).toBeInTheDocument();
    });

    it('should render error notification', () => {
      render(
        <Notification
          type="error"
          message="Operation failed"
          onClose={mockOnClose}
        />
      );

      expect(screen.getByText('Operation failed')).toBeInTheDocument();
      expect(screen.getByRole('button')).toBeInTheDocument();
    });

    it('should render info notification', () => {
      render(
        <Notification
          type="info"
          message="Information message"
          onClose={mockOnClose}
        />
      );

      expect(screen.getByText('Information message')).toBeInTheDocument();
    });

    it('should render warning notification', () => {
      render(
        <Notification
          type="warning"
          message="Warning message"
          onClose={mockOnClose}
        />
      );

      expect(screen.getByText('Warning message')).toBeInTheDocument();
    });

    it('should call onClose when clicking close button', () => {
      render(
        <Notification
          type="info"
          message="Test message"
          onClose={mockOnClose}
        />
      );

      const closeButton = screen.getByRole('button');
      fireEvent.click(closeButton);

      expect(mockOnClose).toHaveBeenCalledTimes(1);
    });

    it('should auto-close after duration', () => {
      render(
        <Notification
          type="info"
          message="Auto-close message"
          duration={3000}
          onClose={mockOnClose}
        />
      );

      act(() => {
        jest.advanceTimersByTime(3000);
      });

      expect(mockOnClose).toHaveBeenCalledTimes(1);
    });

    it('should not auto-close when duration is 0', () => {
      render(
        <Notification
          type="info"
          message="Persistent message"
          duration={0}
          onClose={mockOnClose}
        />
      );

      act(() => {
        jest.advanceTimersByTime(5000);
      });

      expect(mockOnClose).not.toHaveBeenCalled();
    });

    it('should have correct CSS class for type', () => {
      const { container, rerender } = render(
        <Notification
          type="success"
          message="Test"
          onClose={mockOnClose}
        />
      );

      expect(container.firstChild).toHaveClass('notification-success');

      rerender(
        <Notification
          type="error"
          message="Test"
          onClose={mockOnClose}
        />
      );

      expect(container.firstChild).toHaveClass('notification-error');
    });
  });

  describe('NotificationContainer', () => {
    it('should render multiple notifications', () => {
      const notifications = [
        { id: '1', type: 'success' as const, message: 'Success 1' },
        { id: '2', type: 'error' as const, message: 'Error 2' },
        { id: '3', type: 'info' as const, message: 'Info 3' },
      ];

      render(
        <NotificationContainer
          notifications={notifications}
          onRemove={mockOnClose}
        />
      );

      expect(screen.getByText('Success 1')).toBeInTheDocument();
      expect(screen.getByText('Error 2')).toBeInTheDocument();
      expect(screen.getByText('Info 3')).toBeInTheDocument();
    });

    it('should call onRemove when closing notification', () => {
      const notifications = [
        { id: '1', type: 'success' as const, message: 'Success 1' },
      ];

      render(
        <NotificationContainer
          notifications={notifications}
          onRemove={mockOnClose}
        />
      );

      const closeButton = screen.getByRole('button');
      fireEvent.click(closeButton);

      expect(mockOnClose).toHaveBeenCalledWith('1');
    });

    it('should render empty container when no notifications', () => {
      render(
        <NotificationContainer
          notifications={[]}
          onRemove={mockOnClose}
        />
      );

      expect(screen.queryByRole('alert')).not.toBeInTheDocument();
    });
  });

  describe('notification helper', () => {
    it('should have success method', () => {
      expect(notification.success).toBeDefined();
      expect(typeof notification.success).toBe('function');
    });

    it('should have error method', () => {
      expect(notification.error).toBeDefined();
      expect(typeof notification.error).toBe('function');
    });

    it('should have info method', () => {
      expect(notification.info).toBeDefined();
      expect(typeof notification.info).toBe('function');
    });

    it('should have warning method', () => {
      expect(notification.warning).toBeDefined();
      expect(typeof notification.warning).toBe('function');
    });
  });

  describe('useNotification hook', () => {
    it('should return notification state and methods', () => {
      const TestComponent = () => {
        const { notifications, removeNotification, notify } = useNotification();
        
        return (
          <div>
            <span data-testid="count">{notifications.length}</span>
            <button onClick={() => notify.success('Test')}>Add</button>
            <button onClick={() => removeNotification('1')}>Remove</button>
          </div>
        );
      };

      render(<TestComponent />);

      expect(screen.getByTestId('count')).toHaveTextContent('0');
      
      fireEvent.click(screen.getByText('Add'));
      
      // 通知已添加到全局状态
      expect(screen.getByTestId('count')).toHaveTextContent('1');
    });
  });
});
