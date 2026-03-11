/**
 * Ingress Detail Component
 */
import React from 'react';
import { FaExchangeAlt, FaGlobe, FaLock, FaBook, FaCalendarAlt } from 'react-icons/fa';
import { LabelList } from '../LabelList';
import { EventTimeline } from '../EventTimeline';
import { CollapsibleSection } from '../CollapsibleSection';
import './ResourceDetail.css';

interface IngressDetailProps {
  data: any;
  events?: any[];
  onResourceClick?: (type: string, namespace: string, name: string) => void;
}

export const IngressDetail: React.FC<IngressDetailProps> = ({
  data,
  events = [],
  onResourceClick,
}) => {
  if (!data) return null;

  const { metadata, spec, status } = data;
  const rules = spec.rules || [];
  const tls = spec.tls || [];
  const ingressClass = spec.ingressClassName || data.metadata?.annotations?.['kubernetes.io/ingress.class'];
  const loadBalancerIngress = status?.loadBalancer?.ingress || [];

  return (
    <div className="resource-detail-content">
      <CollapsibleSection title={<><FaExchangeAlt className="section-icon" /> Ingress Information</>} defaultExpanded={true}>
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
            <div className="info-label">Class</div>
            <div className="info-value">{ingressClass || '-'}</div>
          </div>
          <div className="info-item">
            <div className="info-label">Rules</div>
            <div className="info-value">{rules.length}</div>
          </div>
        </div>

        {loadBalancerIngress.length > 0 && (
          <div className="strategy-config">
            <div className="subsection-title">Load Balancer</div>
            <div className="info-grid-2">
              {loadBalancerIngress.map((lb: any, idx: number) => (
                <div key={idx} className="strategy-item">
                  <div className="strategy-label">Address</div>
                  <div className="strategy-value">{lb.ip || lb.hostname}</div>
                </div>
              ))}
            </div>
          </div>
        )}
      </CollapsibleSection>

      {rules.length > 0 && (
        <CollapsibleSection title={<><FaGlobe className="section-icon" /> Routing Rules</>} defaultExpanded={true}>
          <div className="rules-table">
            <table className="resource-table">
              <thead>
                <tr>
                  <th>Host</th>
                  <th>Path</th>
                  <th>Path Type</th>
                  <th>Service</th>
                  <th>Port</th>
                </tr>
              </thead>
              <tbody>
                {rules.flatMap((rule: any, ruleIdx: number) =>
                  (rule.http?.paths || []).map((path: any, pathIdx: number) => (
                    <tr key={`${ruleIdx}-${pathIdx}`}>
                      <td>{rule.host || '*'}</td>
                      <td className="code">{path.path || '/'}</td>
                      <td>{path.pathType || 'ImplementationSpecific'}</td>
                      <td>{path.backend?.service?.name || '-'}</td>
                      <td>{path.backend?.service?.port?.number || path.backend?.service?.port?.name || '-'}</td>
                    </tr>
                  ))
                )}
              </tbody>
            </table>
          </div>
        </CollapsibleSection>
      )}

      {tls.length > 0 && (
        <CollapsibleSection title={<><FaLock className="section-icon" /> TLS Configuration</>} defaultExpanded={false}>
          <div className="tls-list">
            {tls.map((tlsConfig: any, idx: number) => (
              <div key={idx} className="tls-item">
                <div className="tls-hosts">
                  <span className="tls-label">Hosts:</span>
                  <span className="tls-value">{tlsConfig.hosts?.join(', ') || '*'}</span>
                </div>
                <div className="tls-secret">
                  <span className="tls-label">Secret:</span>
                  <span className="tls-value">{tlsConfig.secretName || '-'}</span>
                </div>
              </div>
            ))}
          </div>
        </CollapsibleSection>
      )}

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

export default IngressDetail;
