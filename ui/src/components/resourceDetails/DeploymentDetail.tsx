/**
 * Deployment Detail Component
 * Displays comprehensive deployment information
 */
import React from 'react';
import { LabelList } from '../LabelList';
import { EventTimeline } from '../EventTimeline';
import { CollapsibleSection } from '../CollapsibleSection';
import { RelatedResources } from '../RelatedResources';
import './ResourceDetail.css';

interface DeploymentDetailProps {
  data: any;
  events?: any[];
  relatedPods?: any[];
  onResourceClick?: (type: string, namespace: string, name: string) => void;
}

export const DeploymentDetail: React.FC<DeploymentDetailProps> = ({
  data,
  events = [],
  relatedPods = [],
  onResourceClick,
}) => {
  if (!data) return null;

  const { metadata, spec, status } = data;
  const conditions = status?.conditions || [];

  // Replica info
  const replicas = spec.replicas || 1;
  const readyReplicas = status?.readyReplicas || 0;
  const updatedReplicas = status?.updatedReplicas || 0;
  const availableReplicas = status?.availableReplicas || 0;
  const unavailableReplicas = status?.unavailableReplicas || 0;

  // Strategy info
  const strategy = spec.strategy?.type || 'RollingUpdate';
  const rollingUpdate = spec.strategy?.rollingUpdate || {};

  // Containers
  const containers = spec.template?.spec?.containers || [];

  return (
    <div className="resource-detail-content">
      {/* Deployment Info */}
      <CollapsibleSection title="Deployment Information" icon="🚀" defaultExpanded={true}>
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
                {readyReplicas}/{replicas} ready
              </span>
            </div>
          </div>
          <div className="info-item">
            <div className="info-label">Strategy</div>
            <div className="info-value">{strategy}</div>
          </div>
        </div>

        {/* Replica Stats */}
        <div className="replica-stats">
          <div className="stat-item">
            <div className="stat-label">Desired</div>
            <div className="stat-value">{replicas}</div>
          </div>
          <div className="stat-item">
            <div className="stat-label">Ready</div>
            <div className="stat-value status-good-text">{readyReplicas}</div>
          </div>
          <div className="stat-item">
            <div className="stat-label">Updated</div>
            <div className="stat-value">{updatedReplicas}</div>
          </div>
          <div className="stat-item">
            <div className="stat-label">Available</div>
            <div className="stat-value status-good-text">{availableReplicas}</div>
          </div>
          <div className="stat-item">
            <div className="stat-label">Unavailable</div>
            <div className={`stat-value ${unavailableReplicas > 0 ? 'status-bad-text' : ''}`}>
              {unavailableReplicas}
            </div>
          </div>
        </div>

        {/* Conditions */}
        {conditions.length > 0 && (
          <div className="conditions-section">
            <div className="subsection-title">Conditions</div>
            <div className="conditions-grid">
              {conditions.map((condition: any, idx: number) => (
                <div
                  key={idx}
                  className={`condition-badge ${condition.status === 'True' ? 'status-good' : 'status-bad'}`}
                >
                  <span className="condition-type">{condition.type}</span>
                  <span className="condition-status">{condition.status === 'True' ? '✓' : '✗'}</span>
                  {condition.message && (
                    <span className="condition-message" title={condition.message}>
                      {condition.reason || ''}
                    </span>
                  )}
                </div>
              ))}
            </div>
          </div>
        )}

        {/* Strategy Config */}
        {strategy === 'RollingUpdate' && rollingUpdate && (
          <div className="strategy-config">
            <div className="subsection-title">Rolling Update Strategy</div>
            <div className="strategy-grid">
              <div className="strategy-item">
                <div className="strategy-label">Max Surge</div>
                <div className="strategy-value">{rollingUpdate.maxSurge || '25%'}</div>
              </div>
              <div className="strategy-item">
                <div className="strategy-label">Max Unavailable</div>
                <div className="strategy-value">{rollingUpdate.maxUnavailable || '25%'}</div>
              </div>
            </div>
          </div>
        )}
      </CollapsibleSection>

      {/* Pod Selector */}
      <CollapsibleSection title="Pod Selector" icon="🎯" defaultExpanded={true}>
        {spec.selector?.matchLabels ? (
          <LabelList labels={spec.selector.matchLabels} />
        ) : (
          <div className="empty-state">No selector defined</div>
        )}
      </CollapsibleSection>

      {/* Containers */}
      <CollapsibleSection title="Containers" icon="📦" defaultExpanded={false}>
        {containers.map((container: any, idx: number) => (
          <div key={idx} className="container-info-card">
            <div className="container-header">
              <span className="container-name">{container.name}</span>
              <span className="container-image">{container.image}</span>
            </div>
            <div className="container-details">
              {container.ports && container.ports.length > 0 && (
                <div className="detail-row">
                  <span className="detail-label">Ports:</span>
                  <span className="detail-value">
                    {container.ports.map((p: any) => `${p.containerPort}/${p.protocol || 'TCP'}`).join(', ')}
                  </span>
                </div>
              )}
              {container.resources && (
                <div className="detail-row">
                  <span className="detail-label">Resources:</span>
                  <span className="detail-value">
                    Requests: {container.resources.requests?.cpu || '-'} CPU, {container.resources.requests?.memory || '-'} Mem | 
                    Limits: {container.resources.limits?.cpu || '-'} CPU, {container.resources.limits?.memory || '-'} Mem
                  </span>
                </div>
              )}
            </div>
          </div>
        ))}
      </CollapsibleSection>

      {/* Labels and Annotations */}
      <CollapsibleSection title="Labels & Annotations" icon="🏷️" defaultExpanded={false}>
        {spec.template?.metadata?.labels && (
          <div className="labels-subsection">
            <div className="subsection-title">Pod Template Labels</div>
            <LabelList labels={spec.template.metadata.labels} />
          </div>
        )}
        {metadata?.labels && Object.keys(metadata.labels).length > 0 && (
          <div className="labels-subsection">
            <div className="subsection-title">Deployment Labels</div>
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

      {/* Related Pods */}
      <CollapsibleSection title="Related Pods" icon="🔗" defaultExpanded={false}>
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

      {/* Events */}
      {events.length > 0 && (
        <CollapsibleSection title="Events" icon="📅" defaultExpanded={false}>
          <EventTimeline events={events} />
        </CollapsibleSection>
      )}
    </div>
  );
};

export default DeploymentDetail;
