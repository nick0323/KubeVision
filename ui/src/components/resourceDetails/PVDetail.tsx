/**
 * PV Detail Component
 */
import React from 'react';
import { FaDatabase, FaBook, FaCalendarAlt } from 'react-icons/fa';
import { LabelList } from '../LabelList';
import { EventTimeline } from '../EventTimeline';
import { CollapsibleSection } from '../CollapsibleSection';
import './ResourceDetail.css';

interface PVDetailProps {
  data: any;
  events?: any[];
  onResourceClick?: (type: string, namespace: string, name: string) => void;
}

export const PVDetail: React.FC<PVDetailProps> = ({
  data,
  events = [],
  onResourceClick,
}) => {
  if (!data) return null;

  const { metadata, spec, status } = data;
  const capacity = spec.capacity?.storage;
  const accessModes = spec.accessModes || [];

  return (
    <div className="resource-detail-content">
      <CollapsibleSection title={<><FaDatabase className="section-icon" /> PV Information</>} defaultExpanded={true}>
        <div className="info-grid-4">
          <div className="info-item">
            <div className="info-label">Name</div>
            <div className="info-value">{metadata?.name}</div>
          </div>
          <div className="info-item">
            <div className="info-label">Status</div>
            <div className="info-value">
              <span className={`status-badge ${status?.phase === 'Available' ? 'status-good' : status?.phase === 'Bound' ? 'status-warning' : 'status-bad'}`}>
                {status?.phase || 'Unknown'}
              </span>
            </div>
          </div>
          <div className="info-item">
            <div className="info-label">Capacity</div>
            <div className="info-value">{capacity || '-'}</div>
          </div>
          <div className="info-item">
            <div className="info-label">Storage Class</div>
            <div className="info-value">{spec.storageClassName || '-'}</div>
          </div>
        </div>

        <div className="pv-details">
          <div className="detail-row">
            <span className="detail-label">Access Modes:</span>
            <span className="detail-value">{accessModes.join(', ') || '-'}</span>
          </div>
          <div className="detail-row">
            <span className="detail-label">Reclaim Policy:</span>
            <span className="detail-value">{spec.persistentVolumeReclaimPolicy || '-'}</span>
          </div>
          <div className="detail-row">
            <span className="detail-label">Volume Mode:</span>
            <span className="detail-value">{spec.volumeMode || '-'}</span>
          </div>
          {spec.claimRef && (
            <div className="detail-row">
              <span className="detail-label">Claim:</span>
              <span className="detail-value">{spec.claimRef.namespace}/{spec.claimRef.name}</span>
            </div>
          )}
          {spec.local && (
            <div className="detail-row">
              <span className="detail-label">Local Path:</span>
              <span className="detail-value">{spec.local.path}</span>
            </div>
          )}
          {spec.hostPath && (
            <div className="detail-row">
              <span className="detail-label">Host Path:</span>
              <span className="detail-value">{spec.hostPath.path}</span>
            </div>
          )}
          {spec.nfs && (
            <div className="detail-row">
              <span className="detail-label">NFS:</span>
              <span className="detail-value">{spec.nfs.server}:{spec.nfs.path}</span>
            </div>
          )}
        </div>
      </CollapsibleSection>

      <CollapsibleSection title={<><FaBook className="section-icon" /> Labels & Annotations</>} defaultExpanded={false}>
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
        <CollapsibleSection title={<><FaCalendarAlt className="section-icon" /> Events</>} defaultExpanded={false}>
          <EventTimeline events={events} />
        </CollapsibleSection>
      )}
    </div>
  );
};

export default PVDetail;
