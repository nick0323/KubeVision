import React, { useState, useCallback, useEffect } from 'react';
import { EventsTabProps } from '../pages/ResourceDetailPage.types';
import { LoadingSpinner } from '../common/LoadingSpinner';
import { ErrorDisplay } from '../common/ErrorDisplay';
import { authFetch, withCluster } from '../utils/auth';
import { formatTime } from '../utils/time';
import { isClusterResource } from '../constants/config';
import './EventsTab.css';

/**
 * Events Tab - Event list
 */
export const EventsTab: React.FC<EventsTabProps> = ({
  namespace,
  podName,
  name,
  resourceKind,
  onRefresh,
}) => {
  const [events, setEvents] = useState<any[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  // Load事 component
  const loadEvents = useCallback(async () => {
    setLoading(true);
    setError(null);

    try {
      const resourceName = name || podName;
      const involvedObject = resourceKind ? `${resourceKind}/${resourceName}` : resourceName;

      // cluster-scoped resources (Node, PV, etc.) have no namespace
      const eventNamespace = resourceKind && isClusterResource(resourceKind) ? '' : namespace;

      const response = await authFetch(withCluster(
        `/api/event?namespace=${eventNamespace}&involvedObject=${involvedObject}&force=true`
      ));
      const result = await response.json();

      if (result.code === 0 && result.data) {
        // backendalready经filter，直接UseBack'sdata
        setEvents(Array.isArray(result.data) ? result.data : []);
      } else {
        setError(result.message || 'Load failed');
      }
    } catch {
      setError('Load Failed');
    } finally {
      setLoading(false);
    }
  }, [namespace, name, podName, resourceKind]);

  useEffect(() => {
    loadEvents();
  }, [loadEvents, name, resourceKind]);

  if (loading) {
    return <LoadingSpinner text="Loading....." size="lg" />;
  }

  if (error && events.length === 0) {
    return <ErrorDisplay message={error} type="error" showRetry onRetry={loadEvents} />;
  }

  return (
    <div className="events-tab">
      <div className="detail-card">
        <h3 className="detail-card-title">Events</h3>
        {events.length === 0 ? (
          <div className="empty-state">No events</div>
        ) : (
          <table className="detail-table">
            <thead>
              <tr>
                <th className="col-w-100px">Type</th>
                <th className="col-w-150px">Reason</th>
                <th>Message</th>
                <th className="col-w-180px">Last Seen</th>
              </tr>
            </thead>
            <tbody>
              {events.map((event: any, index: number) => (
                <tr key={index} className="table-row">
                  <td>
                    <span
                      className={`event-type ${event.type === 'Warning' ? 'warning' : 'normal'}`}
                    >
                      {event.type}
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
