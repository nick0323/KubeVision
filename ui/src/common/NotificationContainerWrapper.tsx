import React from 'react';
import { useNotificationContext } from './NotificationContext';
import './Notification.css';

const NotificationContainerWrapperImpl: React.FC = () => {
  const { notifications, removeNotification } = useNotificationContext();

  return (
    <div className="notification-container">
      {notifications.map((notification) => (
        <div
          key={notification.id}
          className={`notification notification-${notification.type}`}
        >
          <span className="notification-message">{notification.message}</span>
          <button
            className="notification-close"
            onClick={() => removeNotification(notification.id)}
          >
            ×
          </button>
        </div>
      ))}
    </div>
  );
};

export const NotificationContainerWrapper = React.memo(NotificationContainerWrapperImpl);
