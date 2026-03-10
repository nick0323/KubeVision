/**
 * Job Detail Component
 */
import React from 'react';
import { LabelList } from '../LabelList';
import { EventTimeline } from '../EventTimeline';
import { CollapsibleSection } from '../CollapsibleSection';
import './ResourceDetail.css';

interface JobDetailProps {
  data: any;
  events?: any[];
  onResourceClick?: (type: string, namespace: string, name: string) => void;
}

export const JobDetail: React.FC<JobDetailProps> = ({
  data,
  events = [],
  onResourceClick,
}) => {
  if (!data) return null;

  const { metadata, spec, status } = data;
  const succeeded = status?.succeeded || 0;
  const failed = status?.failed || 0;
  const active = status?.active || 0;
  const isComplete = succeeded > 0;
  const isFailed = failed > 0;

  return (
    <div className="resource-detail-content">
      <CollapsibleSection title="💼 Job Information" defaultExpanded={true}>
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
            <div className="info-label">Status</div>
            <div className="info-value">
              <span className={`status-badge ${isComplete ? 'status-good' : isFailed ? 'status-bad' : 'status-warning'}`}>
                {isComplete ? 'Completed' : isFailed ? 'Failed' : 'Running'}
              </span>
            </div>
          </div>
          <div className="info-item">
            <div className="info-label">Completions</div>
            <div className="info-value">{succeeded}/{spec.completions || 1}</div>
          </div>
        </div>

        <div className="replica-stats">
          <div className="stat-item">
            <div className="stat-label">Active</div>
            <div className="stat-value">{active}</div>
          </div>
          <div className="stat-item">
            <div className="stat-label">Succeeded</div>
            <div className="stat-value status-good-text">{succeeded}</div>
          </div>
          <div className="stat-item">
            <div className="stat-label">Failed</div>
            <div className={`stat-value ${failed > 0 ? 'status-bad-text' : ''}`}>{failed}</div>
          </div>
          <div className="stat-item">
            <div className="stat-label">Completions</div>
            <div className="stat-value">{spec.completions || 1}</div>
          </div>
        </div>

        {spec.parallelism && (
          <div className="strategy-config">
            <div className="subsection-title">Parallelism</div>
            <div className="strategy-grid">
              <div className="strategy-item">
                <div className="strategy-label">Parallel Pods</div>
                <div className="strategy-value">{spec.parallelism}</div>
              </div>
            </div>
          </div>
        )}
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

export default JobDetail;
