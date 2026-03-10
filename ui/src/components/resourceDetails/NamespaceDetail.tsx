/**
 * Namespace Detail Component
 */
import React from 'react';
import { LabelList } from '../LabelList';
import { EventTimeline } from '../EventTimeline';
import { CollapsibleSection } from '../CollapsibleSection';
import './ResourceDetail.css';

interface NamespaceDetailProps {
  data: any;
  events?: any[];
  onResourceClick?: (type: string, namespace: string, name: string) => void;
}

export const NamespaceDetail: React.FC<NamespaceDetailProps> = ({
  data,
  events = [],
  onResourceClick,
}) => {
  if (!data) return null;

  const { metadata, status, spec } = data;
  const phase = status?.phase;

  return (
    <div className="resource-detail-content">
      <CollapsibleSection title="📦 Namespace Information" defaultExpanded={true}>
        <div className="info-grid-4">
          <div className="info-item">
            <div className="info-label">Name</div>
            <div className="info-value">{metadata?.name}</div>
          </div>
          <div className="info-item">
            <div className="info-label">Status</div>
            <div className="info-value">
              <span className={`status-badge ${phase === 'Active' ? 'status-good' : 'status-warning'}`}>
                {phase || 'Unknown'}
              </span>
            </div>
          </div>
          <div className="info-item">
            <div className="info-label">Created</div>
            <div className="info-value">
              {metadata?.creationTimestamp ? new Date(metadata.creationTimestamp).toLocaleDateString() : '-'}
            </div>
          </div>
          <div className="info-item">
            <div className="info-label">Resource Quota</div>
            <div className="info-value">{spec?.resourceQuota ? 'Yes' : 'No'}</div>
          </div>
        </div>

        {spec?.finalizers && spec.finalizers.length > 0 && (
          <div className="sc-details">
            <div className="detail-row">
              <span className="detail-label">Finalizers:</span>
              <span className="detail-value">{spec.finalizers.join(', ')}</span>
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

export default NamespaceDetail;
