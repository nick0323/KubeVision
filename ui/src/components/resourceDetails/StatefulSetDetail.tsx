/**
 * StatefulSet Detail Component
 */
import React from 'react';
import { FaRocket, FaBullseye, FaLink, FaBook, FaCalendarAlt } from 'react-icons/fa';
import { LabelList } from '../LabelList';
import { EventTimeline } from '../EventTimeline';
import { CollapsibleSection } from '../CollapsibleSection';
import { RelatedResources } from '../RelatedResources';
import './ResourceDetail.css';

interface StatefulSetDetailProps {
  data: any;
  events?: any[];
  relatedPods?: any[];
  onResourceClick?: (type: string, namespace: string, name: string) => void;
}

export const StatefulSetDetail: React.FC<StatefulSetDetailProps> = ({
  data,
  events = [],
  relatedPods = [],
  onResourceClick,
}) => {
  if (!data) return null;

  const { metadata, spec, status } = data;
  const replicas = spec.replicas || 1;
  const readyReplicas = status?.readyReplicas || 0;

  return (
    <div className="resource-detail-content">
      <CollapsibleSection title={<><FaRocket className="section-icon" /> StatefulSet Information</>} defaultExpanded={true}>
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
            <div className="info-label">Replicas</div>
            <div className="info-value">
              <span className={`status-badge ${readyReplicas === replicas ? 'status-good' : 'status-warning'}`}>
                {readyReplicas}/{replicas}
              </span>
            </div>
          </div>
          <div className="info-item">
            <div className="info-label">Service</div>
            <div className="info-value">{spec.serviceName || '-'}</div>
          </div>
        </div>

        <div className="replica-stats">
          <div className="stat-item">
            <div className="stat-label">Replicas</div>
            <div className="stat-value">{replicas}</div>
          </div>
          <div className="stat-item">
            <div className="stat-label">Ready</div>
            <div className="stat-value status-good-text">{readyReplicas}</div>
          </div>
          <div className="stat-item">
            <div className="stat-label">Updated</div>
            <div className="stat-value">{status?.updatedReplicas || 0}</div>
          </div>
          <div className="stat-item">
            <div className="stat-label">Available</div>
            <div className="stat-value">{status?.availableReplicas || 0}</div>
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

      <CollapsibleSection title={<><FaLink className="section-icon" /> Related Pods</>} defaultExpanded={false}>
        {relatedPods.length > 0 ? (
          <RelatedResources
            resources={relatedPods.map((pod: any) => ({
              kind: 'Pod',
              name: pod.metadata?.name,
              namespace: pod.metadata?.namespace,
              status: pod.status?.phase,
            }))}
            onResourceClick={onResourceClick}
          />
        ) : (
          <div className="empty-state">No related pods found</div>
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

export default StatefulSetDetail;
