/**
 * Pod Detail Component
 * Displays comprehensive pod information
 */
import React from 'react';
import { ContainerList } from '../ContainerList';
import { LabelList } from '../LabelList';
import { EventTimeline } from '../EventTimeline';
import { CollapsibleSection } from '../CollapsibleSection';
import { RelatedResources } from '../RelatedResources';
import './ResourceDetail.css';

interface PodDetailProps {
  data: any;
  events?: any[];
  onResourceClick?: (type: string, namespace: string, name: string) => void;
  onViewLog?: (containerName: string) => void;
  onExec?: (containerName: string) => void;
}

export const PodDetail: React.FC<PodDetailProps> = ({
  data,
  events = [],
  onResourceClick,
  onViewLog,
  onExec,
}) => {
  if (!data) return null;

  const { metadata, spec, status } = data;
  const containers = spec?.containers || [];
  const initContainers = spec?.initContainers || [];
  const conditions = status?.conditions || [];
  const containerStatuses = status?.containerStatuses || [];

  // Pod info
  const phase = status?.phase || 'Unknown';
  const podIP = status?.podIP;
  const hostIP = status?.hostIP;
  const nodeName = spec?.nodeName;
  const qosClass = status?.qosClass;

  return (
    <div className="resource-detail-content">
      {/* Pod Info Section */}
      <CollapsibleSection title="Pod Information" icon="📦" defaultExpanded={true}>
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
              <span className={`status-badge status-${phase?.toLowerCase()}`}>
                {phase}
              </span>
            </div>
          </div>
          <div className="info-item">
            <div className="info-label">QoS Class</div>
            <div className="info-value">{qosClass || 'BestEffort'}</div>
          </div>
          <div className="info-item">
            <div className="info-label">Node</div>
            <div className="info-value">{nodeName || '-'}</div>
          </div>
          <div className="info-item">
            <div className="info-label">Pod IP</div>
            <div className="info-value">{podIP || '-'}</div>
          </div>
          <div className="info-item">
            <div className="info-label">Host IP</div>
            <div className="info-value">{hostIP || '-'}</div>
          </div>
          <div className="info-item">
            <div className="info-label">Service Account</div>
            <div className="info-value">{spec?.serviceAccountName || 'default'}</div>
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
                </div>
              ))}
            </div>
          </div>
        )}
      </CollapsibleSection>

      {/* Containers Section */}
      <CollapsibleSection title="Containers" icon="📦" defaultExpanded={true}>
        {/* Init Containers */}
        {initContainers.length > 0 && (
          <div className="subsection">
            <div className="subsection-title">Init Containers</div>
            <ContainerList
              containers={initContainers}
              statuses={status?.initContainerStatuses || []}
              onViewLog={onViewLog}
              onExec={onExec}
            />
          </div>
        )}

        {/* Main Containers */}
        {containers.length > 0 && (
          <div className="subsection">
            <div className="subsection-title">Containers</div>
            <ContainerList
              containers={containers}
              statuses={containerStatuses}
              onViewLog={onViewLog}
              onExec={onExec}
            />
          </div>
        )}
      </CollapsibleSection>

      {/* Volumes Section */}
      {spec?.volumes && spec.volumes.length > 0 && (
        <CollapsibleSection title="Volumes" icon="💾" defaultExpanded={false}>
          <div className="volumes-grid">
            {spec.volumes.map((volume: any, idx: number) => {
              const volumeType = volume.persistentVolumeClaim
                ? 'PVC'
                : volume.configMap
                ? 'ConfigMap'
                : volume.secret
                ? 'Secret'
                : volume.emptyDir
                ? 'EmptyDir'
                : volume.hostPath
                ? 'HostPath'
                : 'Unknown';

              return (
                <div key={idx} className="volume-card">
                  <div className="volume-header">
                    <span className="volume-name">{volume.name}</span>
                    <span className="volume-type">{volumeType}</span>
                  </div>
                  <div className="volume-details">
                    {volume.persistentVolumeClaim && (
                      <span>PVC: {volume.persistentVolumeClaim.claimName}</span>
                    )}
                    {volume.configMap && (
                      <span>ConfigMap: {volume.configMap.name}</span>
                    )}
                    {volume.secret && (
                      <span>Secret: {volume.secret.secretName}</span>
                    )}
                    {volume.emptyDir && (
                      <span>EmptyDir{volume.emptyDir.medium ? ` (${volume.emptyDir.medium})` : ''}</span>
                    )}
                    {volume.hostPath && (
                      <span>HostPath: {volume.hostPath.path}</span>
                    )}
                  </div>
                </div>
              );
            })}
          </div>
        </CollapsibleSection>
      )}

      {/* Labels and Annotations */}
      <CollapsibleSection title="Labels & Annotations" icon="🏷️" defaultExpanded={false}>
        {metadata?.labels && Object.keys(metadata.labels).length > 0 && (
          <div className="labels-subsection">
            <div className="subsection-title">Labels</div>
            <LabelList
              labels={metadata.labels}
              resourceType="pod"
              namespace={metadata.namespace}
            />
          </div>
        )}
        {metadata?.annotations && Object.keys(metadata.annotations).length > 0 && (
          <div className="labels-subsection">
            <div className="subsection-title">Annotations</div>
            <LabelList labels={metadata.annotations} />
          </div>
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

export default PodDetail;
