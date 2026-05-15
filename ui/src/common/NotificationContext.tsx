import React, { createContext, useContext, useState, useCallback, ReactNode, useRef, useEffect } from 'react';

export type NotificationType = 'success' | 'error' | 'info' | 'warning';

interface NotificationState {
  id: string;
  type: NotificationType;
  message: string;
  duration?: number;
}

interface NotificationContextType {
  notifications: NotificationState[];
  addNotification: (type: NotificationType, message: string, duration?: number) => void;
  removeNotification: (id: string) => void;
}

const NotificationContext = createContext<NotificationContextType | undefined>(undefined);

// globalNotificationfunction引use（forin非ComponentinUse）
let globalAddNotification: ((type: NotificationType, message: string, duration?: number) => void) | null = null;

export function NotificationProvider({ children }: { children: ReactNode }) {
  const [notifications, setNotifications] = useState<NotificationState[]>([]);
  const timeoutRef = useRef<{ [key: string]: number }>({});

  const removeNotification = useCallback((id: string) => {
    setNotifications(prev => prev.filter(n => n.id !== id));
    if (timeoutRef.current[id]) {
      clearTimeout(timeoutRef.current[id]);
      delete timeoutRef.current[id];
    }
  }, []);

  const addNotification = useCallback((type: NotificationType, message: string, duration?: number) => {
    const id = `${Date.now()}-${Math.random()}`;
    setNotifications(prev => [...prev, { id, type, message, duration }]);
    
    // Auto移除Notification
    const autoRemoveDuration = duration || 5000;
    timeoutRef.current[id] = window.setTimeout(() => {
      removeNotification(id);
    }, autoRemoveDuration);
  }, [removeNotification]);

  useEffect(() => {
    globalAddNotification = addNotification;
    return () => {
      globalAddNotification = null;
    };
  }, [addNotification]);

  return (
    <NotificationContext.Provider value={{ notifications, addNotification, removeNotification }}>
      {children}
    </NotificationContext.Provider>
  );
}

export function useNotificationContext() {
  const context = useContext(NotificationContext);
  if (!context) {
    throw new Error('useNotificationContext must be used within NotificationProvider');
  }
  return context;
}

// 兼容旧code's工具function - globalNotification接口
export const notification = {
  success: (message: string, duration?: number) => {
    if (globalAddNotification) {
      globalAddNotification('success', message, duration);
    }
  },
  error: (message: string, duration?: number) => {
    if (globalAddNotification) {
      globalAddNotification('error', message, duration);
    }
  },
  info: (message: string, duration?: number) => {
    if (globalAddNotification) {
      globalAddNotification('info', message, duration);
    }
  },
  warning: (message: string, duration?: number) => {
    if (globalAddNotification) {
      globalAddNotification('warning', message, duration);
    }
  },
};
