/**
 * DaemonSet Detail Component
 */
import React from 'react';
import { FaServer, FaBullseye, FaBook, FaCalendarAlt } from 'react-icons/fa';
import { LabelList } from '../LabelList';
import { EventTimeline } from '../EventTimeline';
import { CollapsibleSection } from '../CollapsibleSection';
import './ResourceDetail.css';

interface DaemonSetDetailProps {
  data: any;
  events?: any[];
  onResourceClick?: (type: string, namespace: string, name: string) => void;
}

export const DaemonSetDetail: React.FC<DaemonSetDetailProps> = ({
  data,
  events = [],
  onResourceClick,
}) => {
  if (!data) return null;

  const { metadata, spec, status } = data;

  return (
    <div className="resource-detail-content">
      <CollapsibleSection title={<><FaServer className="section-icon" /> DaemonSet Information</>} defaultExpanded={true}>
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
            <div className="info-label">Desired</div>
            <div className="info-value">{status?.desiredNumberScheduled || 0}</div>
          </div>
          <div className="info-item">
            <div className="info-label">Ready</div>
            <div className="info-value">
              <span className={`status-badge ${status?.numberReady === status?.desiredNumberScheduled ? 'status-good' : 'status-warning'}`}>
                {status?.numberReady || 0}
              </span>
            </div>
          </div>
        </div>

        <div className="replica-stats">
          <div className="stat-item">
            <div className="stat-label">Desired</div>
            <div className="stat-value">{status?.desiredNumberScheduled || 0}</div>
          </div>
          <div className="stat-item">
            <div className="stat-label">Ready</div>
            <div className="stat-value status-good-text">{status?.numberReady || 0}</div>
          </div>
          <div className="stat-item">
            <div className="stat-label">Updated</div>
            <div className="stat-value">{status?.updatedNumberScheduled || 0}</div>
          </div>
          <div className="stat-item">
            <div className="stat-label">Available</div>
            <div className="stat-value">{status?.numberAvailable || 0}</div>
          </div>
        </div>
      </CollapsibleSection>

      <CollapsibleSection title={<><FaBullseye className="section-icon" /> Pod Selector</>} defaultExpanded={true}>
        {spec.selector?.matchLabels ? (
          <LabelList labels={spec.selector.matchLabels} />
        ) : (
          <div className="empty-state">No selector defined</div>
        )}
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

export default DaemonSetDetail;
