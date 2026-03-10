/**
 * Node Detail Component
 * Displays comprehensive node information
 */
import React from 'react';
import { ResourceBar } from '../ResourceBar';
import { LabelList } from '../LabelList';
import { EventTimeline } from '../EventTimeline';
import { CollapsibleSection } from '../CollapsibleSection';
import { RelatedResources } from '../RelatedResources';
import './ResourceDetail.css';

interface NodeDetailProps {
  data: any;
  events?: any[];
  pods?: any[];
  onResourceClick?: (type: string, namespace: string, name: string) => void;
}

export const NodeDetail: React.FC<NodeDetailProps> = ({
  data,
  events = [],
  pods = [],
  onResourceClick,
}) => {
  if (!data) return null;

  const { metadata, status } = data;
  const conditions = status?.conditions || [];
  const addresses = status?.addresses || [];
  const allocatable = status?.allocatable || {};
  const capacity = status?.capacity || {};
  const nodeInfo = status?.nodeInfo || {};

  // Node info
  const internalIP = addresses?.find((a: any) => a.type === 'InternalIP')?.address;
  const externalIP = addresses?.find((a: any) => a.type === 'ExternalIP')?.address;
  const hostName = addresses?.find((a: any) => a.type === 'Hostname')?.address;

  // Roles
  const roles = Object.keys(metadata?.labels || {})
    .filter((k) => k.includes('node-role.kubernetes.io/'))
    .map((k) => k.replace('node-role.kubernetes.io/', ''));
  const roleDisplay = roles.length > 0 ? roles.join(', ') : 'worker';

  // Conditions
  const readyCondition = conditions.find((c: any) => c.type === 'Ready');
  const isReady = readyCondition?.status === 'True';

  return (
    <div className="resource-detail-content">
      {/* Node Info */}
      <CollapsibleSection title="Node Information" icon="🖥️" defaultExpanded={true}>
        <div className="info-grid-4">
          <div className="info-item">
            <div className="info-label">Name</div>
            <div className="info-value">{metadata?.name}</div>
          </div>
          <div className="info-item">
            <div className="info-label">Status</div>
            <div className="info-value">
              <span className={`status-badge ${isReady ? 'status-good' : 'status-bad'}`}>
                {isReady ? 'Ready' : 'Not Ready'}
              </span>
            </div>
          </div>
          <div className="info-item">
            <div className="info-label">Role</div>
            <div className="info-value">{roleDisplay}</div>
          </div>
          <div className="info-item">
            <div className="info-label">Kubelet Version</div>
            <div className="info-value">{nodeInfo.kubeletVersion || '-'}</div>
          </div>
          <div className="info-item">
            <div className="info-label">Internal IP</div>
            <div className="info-value">{internalIP || '-'}</div>
          </div>
          <div className="info-item">
            <div className="info-label">External IP</div>
            <div className="info-value">{externalIP || '-'}</div>
          </div>
          <div className="info-item">
            <div className="info-label">OS Image</div>
            <div className="info-value">{nodeInfo.osImage || '-'}</div>
          </div>
          <div className="info-item">
            <div className="info-label">Architecture</div>
            <div className="info-value">{nodeInfo.architecture || '-'}</div>
          </div>
        </div>

        {/* Conditions */}
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

        {/* Resource Bars */}
        <div className="resource-bars-section">
          <div className="subsection-title">Resource Allocation</div>
          <ResourceBar
            items={[
              {
                label: 'CPU',
                value: `${allocatable.cpu || capacity.cpu || '0'} cores`,
                percentage: 0,
                type: 'cpu',
              },
              {
                label: 'Memory',
                value: `${allocatable.memory ? (parseInt(allocatable.memory) / (1024 * 1024 * 1024)).toFixed(1) : '0'} Gi`,
                percentage: 0,
                type: 'memory',
              },
              {
                label: 'Pods',
                value: `${pods.length}/${allocatable.pods || capacity.pods || '110'}`,
                percentage: (pods.length / parseInt(allocatable.pods || capacity.pods || '110')) * 100,
                type: 'pods',
              },
            ]}
          />
        </div>
      </CollapsibleSection>

      {/* System Info */}
      <CollapsibleSection title="System Information" icon="🔧" defaultExpanded={false}>
        <div className="info-grid-2">
          <div className="info-item">
            <div className="info-label">Kernel Version</div>
            <div className="info-value">{nodeInfo.kernelVersion}</div>
          </div>
          <div className="info-item">
            <div className="info-label">Container Runtime</div>
            <div className="info-value">{nodeInfo.containerRuntimeVersion}</div>
          </div>
          <div className="info-item">
            <div className="info-label">Machine ID</div>
            <div className="info-value code">{nodeInfo.machineID || '-'}</div>
          </div>
          <div className="info-item">
            <div className="info-label">System UUID</div>
            <div className="info-value code">{nodeInfo.systemUUID || '-'}</div>
          </div>
        </div>
      </CollapsibleSection>

      {/* Allocatable Resources */}
      <CollapsibleSection title="Allocatable Resources" icon="📈" defaultExpanded={false}>
        <div className="allocatable-grid">
          <div className="allocatable-item">
            <div className="allocatable-label">CPU Capacity</div>
            <div className="allocatable-value">{capacity.cpu || '-'}</div>
          </div>
          <div className="allocatable-item">
            <div className="allocatable-label">CPU Allocatable</div>
            <div className="allocatable-value">{allocatable.cpu || '-'}</div>
          </div>
          <div className="allocatable-item">
            <div className="allocatable-label">Memory Capacity</div>
            <div className="allocatable-value">
              {capacity.memory ? (parseInt(capacity.memory) / (1024 * 1024 * 1024)).toFixed(2) + ' Gi' : '-'}
            </div>
          </div>
          <div className="allocatable-item">
            <div className="allocatable-label">Memory Allocatable</div>
            <div className="allocatable-value">
              {allocatable.memory ? (parseInt(allocatable.memory) / (1024 * 1024 * 1024)).toFixed(2) + ' Gi' : '-'}
            </div>
          </div>
          <div className="allocatable-item">
            <div className="allocatable-label">Pod Capacity</div>
            <div className="allocatable-value">{capacity.pods || '-'}</div>
          </div>
          <div className="allocatable-item">
            <div className="allocatable-label">Pod Allocatable</div>
            <div className="allocatable-value">{allocatable.pods || '-'}</div>
          </div>
        </div>
      </CollapsibleSection>

      {/* Labels */}
      <CollapsibleSection title="Labels" icon="🏷️" defaultExpanded={false}>
        {metadata?.labels && Object.keys(metadata.labels).length > 0 ? (
          <LabelList
            labels={metadata.labels}
            resourceType="node"
          />
        ) : (
          <div className="empty-state">No labels</div>
        )}
      </CollapsibleSection>

      {/* Annotations */}
      <CollapsibleSection title="Annotations" icon="📝" defaultExpanded={false}>
        {metadata?.annotations && Object.keys(metadata.annotations).length > 0 ? (
          <LabelList labels={metadata.annotations} />
        ) : (
          <div className="empty-state">No annotations</div>
        )}
      </CollapsibleSection>

      {/* Running Pods */}
      <CollapsibleSection title={`Running Pods (${pods.length})`} icon="🔗" defaultExpanded={false}>
        {pods.length > 0 ? (
          <RelatedResources
            resources={pods.map((pod: any) => ({
              kind: 'Pod',
              name: pod.metadata?.name,
              namespace: pod.metadata?.namespace,
              status: pod.status?.phase,
            }))}
            onResourceClick={onResourceClick}
          />
        ) : (
          <div className="empty-state">No pods running on this node</div>
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

export default NodeDetail;
