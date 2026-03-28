import React, { useState, useCallback, useEffect, useMemo } from 'react';
import { EventsTabProps } from '../types';
import { LoadingSpinner } from '../../LoadingSpinner';
import { ErrorDisplay } from '../../ErrorDisplay';
import { authFetch } from '../../../utils/auth';
import { EventStats } from '../types';
import './EventsTab.css';

/**
 * Events Tab - 事件列表
 */
export const EventsTab: React.FC<EventsTabProps> = ({ namespace, podName }) => {
  const [events, setEvents] = useState<any[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [stats, setStats] = useState<EventStats>({ total: 0, normal: 0, warning: 0 });

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
        const eventList = Array.isArray(result.data) ? result.data : [];
        setEvents(eventList);
        
        // 统计
        const normal = eventList.filter((e: any) => e.type === 'Normal').length;
        const warning = eventList.filter((e: any) => e.type === 'Warning').length;
        setStats({
          total: eventList.length,
          normal,
          warning,
        });
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

  // 复制事件
  const handleCopy = useCallback(() => {
    const content = events.map((e: any) => 
      `[${e.type}] ${e.reason}: ${e.message}`
    ).join('\n');
    navigator.clipboard.writeText(content);
    alert('事件已复制');
  }, [events]);

  // 下载事件
  const handleDownload = useCallback(() => {
    const content = events.map((e: any) => 
      `[${e.lastTimestamp || e.eventTime}] ${e.type} ${e.reason}: ${e.message}`
    ).join('\n');
    const blob = new Blob([content], { type: 'text/plain' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = `${podName}-events.txt`;
    a.click();
    URL.revokeObjectURL(url);
  }, [events, podName]);

  // 格式化时间
  const formatTime = (timestamp?: string) => {
    if (!timestamp) return '-';
    return new Date(timestamp).toLocaleString('zh-CN');
  };

  if (loading) {
    return <LoadingSpinner text="加载事件..." size="lg" />;
  }

  if (error && events.length === 0) {
    return <ErrorDisplay message={error} type="error" showRetry onRetry={loadEvents} />;
  }

  return (
    <div className="events-tab">
      {/* 统计卡片 */}
      <div className="stats-grid">
        <div className="stat-card">
          <div className="stat-value">{stats.total}</div>
          <div className="stat-label">Total</div>
        </div>
        <div className="stat-card">
          <div className="stat-value" style={{ color: '#52c41a' }}>{stats.normal}</div>
          <div className="stat-label">Normal</div>
        </div>
        <div className={`stat-card ${stats.warning > 0 ? 'warning' : ''}`}>
          <div className="stat-value" style={{ color: '#faad14' }}>{stats.warning}</div>
          <div className="stat-label">Warning</div>
        </div>
      </div>

      {/* 操作按钮 */}
      <div className="events-actions">
        <button className="toolbar-btn" onClick={handleCopy}>📋 复制</button>
        <button className="toolbar-btn" onClick={handleDownload}>⬇️ 下载</button>
        <button className="toolbar-btn" onClick={loadEvents}>🔄 刷新</button>
      </div>

      {/* 事件列表 */}
      {events.length === 0 ? (
        <div className="empty-state">
          <span className="empty-state-icon">⚡</span>
          <span className="empty-state-text">暂无事件</span>
        </div>
      ) : (
        <div className="events-list">
          {events.map((event: any, index: number) => (
            <div
              key={index}
              className={`event-item ${event.type === 'Warning' ? 'warning' : 'normal'}`}
            >
              <div className="event-time">
                {formatTime(event.lastTimestamp || event.eventTime)}
              </div>
              <div className="event-content">
                <div className="event-header">
                  <span className={`event-type-icon ${event.type === 'Warning' ? 'warning' : 'normal'}`}>
                    {event.type === 'Warning' ? '⚠️' : '✓'}
                  </span>
                  <span className="event-reason">{event.reason}</span>
                  {event.count && event.count > 1 && (
                    <span className="event-count">x{event.count}</span>
                  )}
                </div>
                <div className="event-message">{event.message}</div>
                <div className="event-meta">
                  <span>Source: {event.source?.component || event.reportingController || '-'}</span>
                  {event.involvedObject && (
                    <span>Object: {event.involvedObject.kind}/{event.involvedObject.name}</span>
                  )}
                </div>
              </div>
            </div>
          ))}
        </div>
      )}
    </div>
  );
};

export default EventsTab;
