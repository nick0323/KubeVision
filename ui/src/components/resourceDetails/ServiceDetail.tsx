/**
 * Service Detail Component
 */
import React from 'react';
import { FaNetworkWired, FaPlug, FaLink, FaBullseye, FaBook, FaCalendarAlt } from 'react-icons/fa';
import { LabelList } from '../LabelList';
import { EventTimeline } from '../EventTimeline';
import { CollapsibleSection } from '../CollapsibleSection';
import { RelatedResources } from '../RelatedResources';
import './ResourceDetail.css';

interface ServiceDetailProps {
  data: any;
  endpoints?: any;
  pods?: any[];
  onResourceClick?: (type: string, namespace: string, name: string) => void;
}

export const ServiceDetail: React.FC<ServiceDetailProps> = ({
  data,
  endpoints,
  pods = [],
  onResourceClick,
}) => {
  if (!data) return null;

  const { metadata, spec, status } = data;
  const type = spec.type || 'ClusterIP';
  const clusterIP = spec.clusterIP;
  const ports = spec.ports || [];

  return (
    <div className="resource-detail-content">
      <CollapsibleSection title={<><FaNetworkWired className="section-icon" /> Service Information</>} defaultExpanded={true}>
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
            <div className="info-label">Type</div>
            <div className="info-value">{type}</div>
          </div>
          <div className="info-item">
            <div className="info-label">Cluster IP</div>
            <div className="info-value">{clusterIP || 'None'}</div>
          </div>
        </div>

        {ports.length > 0 && (
          <div className="ports-section">
            <div className="subsection-title"><FaPlug className="subsection-icon" /> Ports</div>
            <div className="ports-table">
              <table className="resource-table">
                <thead>
                  <tr>
                    <th>Name</th>
                    <th>Protocol</th>
                    <th>Port</th>
                    <th>Target Port</th>
                    {type === 'NodePort' && <th>Node Port</th>}
                  </tr>
                </thead>
                <tbody>
                  {ports.map((port: any, idx: number) => (
                    <tr key={idx}>
                      <td>{port.name || '-'}</td>
                      <td>{port.protocol || 'TCP'}</td>
                      <td className="code">{port.port}</td>
                      <td className="code">{port.targetPort || port.port}</td>
                      {type === 'NodePort' && <td className="code">{port.nodePort || '-'}</td>}
                    </tr>
                  ))}
                </tbody>
              </table>
            </div>
          </div>
        )}
      </CollapsibleSection>

      <CollapsibleSection title={<><FaBullseye className="section-icon" /> Pod Selector</>} defaultExpanded={true}>
        {spec.selector ? (
          <LabelList labels={spec.selector} />
        ) : (
          <div className="empty-state">No selector defined</div>
        )}
      </CollapsibleSection>

      <CollapsibleSection title={<><FaLink className="section-icon" /> Endpoints</>} defaultExpanded={false}>
        {endpoints?.subsets?.length > 0 ? (
          <div className="endpoints-list">
            {endpoints.subsets.flatMap((subset: any) =>
              (subset.addresses || []).map((addr: any, idx: number) => (
                <div key={idx} className="endpoint-item">
                  <span className="endpoint-ip">{addr.ip}</span>
                  <span className="endpoint-ports">
                    {(subset.ports || []).map((p: any) => `${p.port}/${p.protocol || 'TCP'}`).join(', ')}
                  </span>
                </div>
              ))
            )}
          </div>
        ) : (
          <div className="empty-state">No endpoints</div>
        )}
      </CollapsibleSection>

      <CollapsibleSection title={<><FaLink className="section-icon" /> Related Pods</>} defaultExpanded={false}>
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

export default ServiceDetail;
