/**
 * CronJob Detail Component
 */
import React from 'react';
import { LabelList } from '../LabelList';
import { EventTimeline } from '../EventTimeline';
import { CollapsibleSection } from '../CollapsibleSection';
import './ResourceDetail.css';

interface CronJobDetailProps {
  data: any;
  events?: any[];
  onResourceClick?: (type: string, namespace: string, name: string) => void;
}

export const CronJobDetail: React.FC<CronJobDetailProps> = ({
  data,
  events = [],
  onResourceClick,
}) => {
  if (!data) return null;

  const { metadata, spec, status } = data;

  return (
    <div className="resource-detail-content">
      <CollapsibleSection title="⏰ CronJob Information" defaultExpanded={true}>
        <div className="info-grid-4">
          <div className="info-item">
            <div className="info-label">Name</div>
            <div className="info-value">{metadata?.name}</div>
          </div>
          <div className="info-item">
            <div className="info-label">Namespace</div>
            <div className="info-value">{metadata?.namespace}</div>
          </div>
          <div className="info-item">
            <div className="info-label">Schedule</div>
            <div className="info-value code">{spec.schedule || '-'}</div>
          </div>
          <div className="info-item">
            <div className="info-label">Suspend</div>
            <div className="info-value">
              <span className={`status-badge ${spec.suspend ? 'status-bad' : 'status-good'}`}>
                {spec.suspend ? 'Suspended' : 'Active'}
              </span>
            </div>
          </div>
        </div>

        <div className="replica-stats">
          <div className="stat-item">
            <div className="stat-label">Active</div>
            <div className="stat-value">{status?.active?.length || 0}</div>
          </div>
          <div className="stat-item">
            <div className="stat-label">Last Schedule</div>
            <div className="stat-value">
              {status?.lastScheduleTime ? new Date(status.lastScheduleTime).toLocaleString() : '-'}
            </div>
          </div>
          <div className="stat-item">
            <div className="stat-label">Concurrency</div>
            <div className="stat-value">{spec.concurrencyPolicy || 'Allow'}</div>
          </div>
          <div className="stat-item">
            <div className="stat-label">Starting Deadline</div>
            <div className="stat-value">{spec.startingDeadlineSeconds ? `${spec.startingDeadlineSeconds}s` : '-'}</div>
          </div>
        </div>
      </CollapsibleSection>

      <CollapsibleSection title="🏷️ Labels & Annotations" defaultExpanded={false}>
        {metadata?.labels && Object.keys(metadata.labels).length > 0 && (
          <div className="labels-subsection">
            <div className="subsection-title">Labels</div>
            <LabelList labels={metadata.labels} />
          </div>
        )}
        {metadata?.annotations && Object.keys(metadata.annotations).length > 0 && (
          <div className="labels-subsection">
            <div className="subsection-title">Annotations</div>
            <LabelList labels={metadata.annotations} />
          </div>
        )}
      </CollapsibleSection>

      {events.length > 0 && (
        <CollapsibleSection title="📅 Events" defaultExpanded={false}>
          <EventTimeline events={events} />
        </CollapsibleSection>
      )}
    </div>
  );
};

export default CronJobDetail;
