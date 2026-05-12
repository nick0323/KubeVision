import React from 'react';
import { K8sEventSimple } from '../types';
import { formatRelativeTime } from '../utils/time';
import './EventCard.css';

interface EventCardProps {
  events?: K8sEventSimple[];
  title?: string;
  limit?: number;
}

/**
 * Event listCardComponent
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
