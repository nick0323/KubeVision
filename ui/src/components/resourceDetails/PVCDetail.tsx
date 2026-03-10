/**
 * PersistentVolumeClaim Detail Component
 */
import React from 'react';
import { LabelList } from '../LabelList';
import { EventTimeline } from '../EventTimeline';
import { CollapsibleSection } from '../CollapsibleSection';
import './ResourceDetail.css';

interface PVCDetailProps {
  data: any;
  events?: any[];
  onResourceClick?: (type: string, namespace: string, name: string) => void;
}

export const PVCDetail: React.FC<PVCDetailProps> = ({
  data,
  events = [],
  onResourceClick,
}) => {
  if (!data) return null;

  const { metadata, spec, status } = data;
  const accessModes = spec.accessModes || [];
  const capacity = status?.capacity?.storage || spec.resources?.requests?.storage;

  return (
    <div className="resource-detail-content">
      <CollapsibleSection title="💾 PVC Information" defaultExpanded={true}>
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
              <span className={`status-badge ${status?.phase === 'Bound' ? 'status-good' : 'status-warning'}`}>
                {status?.phase || 'Pending'}
              </span>
            </div>
          </div>
          <div className="info-item">
            <div className="info-label">Capacity</div>
            <div className="info-value">{capacity || '-'}</div>
          </div>
        </div>

        <div className="pvc-details">
          <div className="detail-row">
            <span className="detail-label">Volume:</span>
            <span className="detail-value">{status?.volumeName || 'Not bound'}</span>
          </div>
          <div className="detail-row">
            <span className="detail-label">Storage Class:</span>
            <span className="detail-value">{spec.storageClassName || '-'}</span>
          </div>
          <div className="detail-row">
            <span className="detail-label">Access Modes:</span>
            <span className="detail-value">{accessModes.join(', ') || '-'}</span>
          </div>
          {spec.volumeMode && (
            <div className="detail-row">
              <span className="detail-label">Volume Mode:</span>
              <span className="detail-value">{spec.volumeMode}</span>
            </div>
          )}
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

export default PVCDetail;
