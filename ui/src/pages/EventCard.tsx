import React from 'react';
import { K8sEventSimple } from '../types';
import './EventCard.css';

interface EventCardProps {
  events?: K8sEventSimple[];
  title?: string;
  limit?: number;
}

/**
 * 格式化相对时间
 */
const formatRelativeTime = (dateString: string) => {
  const date = new Date(dateString);
  const now = new Date();
  const diffMs = now.getTime() - date.getTime();
  const diffSec = Math.floor(diffMs / 1000);
  const diffMin = Math.floor(diffSec / 60);
  const diffHour = Math.floor(diffMin / 60);
  const diffDay = Math.floor(diffHour / 24);

  if (diffSec < 60) return 'Just now';
  if (diffMin < 60) return `${diffMin}m ago`;
  if (diffHour < 24) return `${diffHour}h ago`;
  return `${diffDay}d ago`;
};

/**
 * 事件列表卡片组件
 */
export const EventCard: React.FC<EventCardProps> = ({
  events = [],
  title = 'Recent Events',
  limit = 5,
}) => {
  const sortedEvents = [...events]
    .sort(
      (a, b) => new Date(b.lastSeen).getTime() - new Date(a.lastSeen).getTime()
    )
    .slice(0, limit);

  return (
    <div className="event-card">
      <div className="event-card-title">{title}</div>
      {sortedEvents.length > 0 ? (
        <div className="event-list">
          {sortedEvents.map((event, index) => (
            <div
              key={event.name || index}
              className={`event-item ${event.type === 'Warning' ? 'event-warning' : 'event-normal'}`}
            >
              <div className="event-header">
                <span className={`event-type event-type-${event.type.toLowerCase()}`}>
                  {event.type}
                </span>
                <span className="event-reason">{event.reason}</span>
                <span className="event-time">
                  {formatRelativeTime(event.lastSeen)}
                </span>
              </div>
              <div className="event-message">{event.message}</div>
              <div className="event-meta">
                {event.pod && <span>Pod: {event.pod}</span>}
                {event.namespace && event.name && (
                  <span>
                    {' '}
                    {event.namespace}/{event.name}
                  </span>
                )}
                {event.reporter && <span> Reporter: {event.reporter}</span>}
              </div>
            </div>
          ))}
        </div>
      ) : (
        <div className="event-empty">No Events</div>
      )}
    </div>
  );
};

export default EventCard;
