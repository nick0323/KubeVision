import React, { useEffect } from 'react';
import { FaCheckCircle, FaExclamationCircle, FaInfoCircle, FaTimes } from 'react-icons/fa';
import './Notification.css';

export type NotificationType = 'success' | 'error' | 'info' | 'warning';

export interface NotificationProps {
  type: NotificationType;
  message: string;
  duration?: number;
  onClose: () => void;
}

// 全局通知状态
let notificationListeners: Array<(notifications: NotificationState[]) => void> = [];
let notifications: NotificationState[] = [];

interface NotificationState {
  id: string;
  type: NotificationType;
  message: string;
  duration?: number;
}

function addNotification(type: NotificationType, message: string, duration?: number) {
  const id = `${Date.now()}-${Math.random()}`;
  notifications = [...notifications, { id, type, message, duration }];
  notificationListeners.forEach((listener) => listener(notifications));
}

/**
 * 使用通知的工具函数
 */
export const notification = {
  success: (message: string, duration?: number) => {
    addNotification('success', message, duration);
  },
  error: (message: string, duration?: number) => {
    addNotification('error', message, duration);
  },
  info: (message: string, duration?: number) => {
    addNotification('info', message, duration);
  },
  warning: (message: string, duration?: number) => {
    addNotification('warning', message, duration);
  },
};

export function useNotification() {
  const [localNotifications, setLocalNotifications] = React.useState<NotificationState[]>([]);

  React.useEffect(() => {
    const listener = (newNotifications: NotificationState[]) => {
      setLocalNotifications(newNotifications);
    };
    notificationListeners.push(listener);
    return () => {
      notificationListeners = notificationListeners.filter((l) => l !== listener);
    };
  }, []);

  const removeNotification = (id: string) => {
    notifications = notifications.filter((n) => n.id !== id);
    notificationListeners.forEach((listener) => listener(notifications));
  };

  return {
    notifications: localNotifications,
    removeNotification,
    notify: notification,
  };
}

/**
 * Notification 通知组件
 * 用于替代 alert() 提供友好的用户提示
 */
export const Notification: React.FC<NotificationProps> = ({
  type,
  message,
  duration = 3000,
  onClose,
}) => {
  useEffect(() => {
    if (duration > 0) {
      const timer = setTimeout(() => {
        onClose();
      }, duration);
      return () => clearTimeout(timer);
    }
  }, [duration, onClose]);

  const icons = {
    success: <FaCheckCircle />,
    error: <FaExclamationCircle />,
    info: <FaInfoCircle />,
    warning: <FaExclamationCircle />,
  };

  return (
    <div className={`notification notification-${type}`}>
      <span className="notification-icon">{icons[type]}</span>
      <span className="notification-message">{message}</span>
      <button className="notification-close" onClick={onClose}>
        <FaTimes />
      </button>
    </div>
  );
};

export interface NotificationContainerProps {
  notifications: Array<{
    id: string;
    type: NotificationType;
    message: string;
  }>;
  onRemove: (id: string) => void;
}

/**
 * NotificationContainer 通知容器
 * 管理多个通知的显示
 */
export const NotificationContainer: React.FC<NotificationContainerProps> = ({
  notifications,
  onRemove,
}) => {
  return (
    <div className="notification-container">
      {notifications.map((notification) => (
        <Notification
          key={notification.id}
          type={notification.type}
          message={notification.message}
          onClose={() => onRemove(notification.id)}
        />
      ))}
    </div>
  );
};
