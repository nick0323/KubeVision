import React, { useState, useCallback, useEffect } from 'react';
import { EventsTabProps } from '../types';
import { LoadingSpinner } from '../../LoadingSpinner';
import { ErrorDisplay } from '../../ErrorDisplay';
import { authFetch } from '../../../utils/auth';
import './EventsTab.css';

/**
 * Events Tab - 事件列表
 */
export const EventsTab: React.FC<EventsTabProps> = ({ namespace, podName }) => {
  const [events, setEvents] = useState<any[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  // 加载事件
  const loadEvents = useCallback(async () => {
    setLoading(true);
    setError(null);

    try {
      const response = await authFetch(
        `/api/events?namespace=${namespace}&involvedObject=Pod/${podName}`
      );
      const result = await response.json();

      if (result.code === 0 && result.data) {
        setEvents(Array.isArray(result.data) ? result.data : []);
      } else {
        setError(result.message || '加载事件失败');
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : '加载失败');
    } finally {
      setLoading(false);
    }
  }, [namespace, podName]);

  useEffect(() => {
    loadEvents();
  }, [loadEvents]);

  // 格式化时间
  const formatTime = (timestamp?: string) => {
    if (!timestamp) return '-';
    // 尝试解析多种时间格式
    try {
      const date = new Date(timestamp);
      // 格式化为：2026-03-31 20:42:11
      const year = date.getFullYear();
      const month = String(date.getMonth() + 1).padStart(2, '0');
      const day = String(date.getDate()).padStart(2, '0');
      const hours = String(date.getHours()).padStart(2, '0');
      const minutes = String(date.getMinutes()).padStart(2, '0');
      const seconds = String(date.getSeconds()).padStart(2, '0');
      return `${year}-${month}-${day} ${hours}:${minutes}:${seconds}`;
    } catch (e) {
      return timestamp;
    }
  };

  if (loading) {
    return <LoadingSpinner text="加载事件..." size="lg" />;
  }

  if (error && events.length === 0) {
    return <ErrorDisplay message={error} type="error" showRetry onRetry={loadEvents} />;
  }

  return (
    <div className="events-tab">
      <div className="detail-card">
        <h3 className="detail-card-title">Events</h3>
        {events.length === 0 ? (
          <div className="empty-state">
            <span className="empty-state-text">No events</span>
          </div>
        ) : (
          <table className="detail-table">
            <thead>
              <tr>
                <th style={{ width: '100px' }}>Type</th>
                <th style={{ width: '150px' }}>Reason</th>
                <th>Message</th>
                <th style={{ width: '180px' }}>Last Seen</th>
              </tr>
            </thead>
            <tbody>
              {events.map((event: any, index: number) => (
                <tr
                  key={index}
                  className={`table-row ${event.type === 'Warning' ? 'warning' : 'normal'}`}
                >
                  <td>
                    <span
                      className={`event-type ${event.type === 'Warning' ? 'warning' : 'normal'}`}
                    >
                      {event.type === 'Warning' ? '⚠️' : '✓'} {event.type}
                    </span>
                  </td>
                  <td className="event-reason">{event.reason || '-'}</td>
                  <td className="event-message">{event.message || '-'}</td>
                  <td className="event-time">
                    {formatTime(event.lastSeen || event.lastTimestamp || event.eventTime)}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </div>
    </div>
  );
};

export default EventsTab;
