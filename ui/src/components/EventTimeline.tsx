import React from 'react';
import './EventTimeline.css';

interface Event {
  type: 'Normal' | 'Warning';
  reason: string;
  message: string;
  count?: number;
  firstTimestamp?: string;
  lastTimestamp?: string;
  eventTime?: string;
}

interface EventTimelineProps {
  events: Event[];
}

export const EventTimeline: React.FC<EventTimelineProps> = ({ events }) => {
  if (!events || events.length === 0) {
    return (
      <div className="event-timeline-empty">
        <div className="empty-icon">📭</div>
        <p>暂无事件</p>
      </div>
    );
  }

  // 按时间倒序排序
  const sortedEvents = [...events].sort((a, b) => {
    const timeA = new Date(a.lastTimestamp || a.eventTime || 0).getTime();
    const timeB = new Date(b.lastTimestamp || b.eventTime || 0).getTime();
    return timeB - timeA;
  });

  return (
    <div className="event-timeline">
      {sortedEvents.map((event, index) => (
        <div
          key={index}
          className={`event-item ${event.type === 'Warning' ? 'warning' : 'normal'}`}
        >
          <div className="event-dot" />
          <div className="event-time">
            {event.lastTimestamp
              ? new Date(event.lastTimestamp).toLocaleString('zh-CN')
              : event.eventTime
              ? new Date(event.eventTime).toLocaleString('zh-CN')
              : 'N/A'}
          </div>
          <div className="event-reason">
            {event.reason}
            {event.count && event.count > 1 && (
              <span className="event-count">{event.count}</span>
            )}
          </div>
          <div className="event-message">{event.message}</div>
        </div>
      ))}
    </div>
  );
};

export default EventTimeline;
