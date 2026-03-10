/**
 * StorageClass Detail Component
 */
import React from 'react';
import { LabelList } from '../LabelList';
import { CollapsibleSection } from '../CollapsibleSection';
import './ResourceDetail.css';

interface StorageClassDetailProps {
  data: any;
  onResourceClick?: (type: string, namespace: string, name: string) => void;
}

export const StorageClassDetail: React.FC<StorageClassDetailProps> = ({
  data,
  onResourceClick,
}) => {
  if (!data) return null;

  const { metadata, provisioner, parameters, reclaimPolicy, volumeBindingMode, allowVolumeExpansion } = data;
  const isDefault = metadata?.annotations?.['storageclass.kubernetes.io/is-default-class'] === 'true';

  return (
    <div className="resource-detail-content">
      <CollapsibleSection title="🗄️ StorageClass Information" defaultExpanded={true}>
        <div className="info-grid-4">
          <div className="info-item">
            <div className="info-label">Name</div>
            <div className="info-value">{metadata?.name}</div>
          </div>
          <div className="info-item">
            <div className="info-label">Provisioner</div>
            <div className="info-value">{provisioner || '-'}</div>
          </div>
          <div className="info-item">
            <div className="info-label">Default</div>
            <div className="info-value">
              <span className={`status-badge ${isDefault ? 'status-good' : 'status-bad'}`}>
                {isDefault ? 'Yes' : 'No'}
              </span>
            </div>
          </div>
          <div className="info-item">
            <div className="info-label">Reclaim Policy</div>
            <div className="info-value">{reclaimPolicy || 'Delete'}</div>
          </div>
        </div>

        <div className="sc-details">
          <div className="detail-row">
            <span className="detail-label">Volume Binding Mode:</span>
            <span className="detail-value">{volumeBindingMode || 'Immediate'}</span>
          </div>
          <div className="detail-row">
            <span className="detail-label">Allow Volume Expansion:</span>
            <span className="detail-value">{allowVolumeExpansion ? 'Yes' : 'No'}</span>
          </div>
          {parameters && Object.keys(parameters).length > 0 && (
            <div className="parameters-section">
              <div className="subsection-title">Parameters</div>
              <div className="parameters-grid">
                {Object.entries(parameters).map(([key, value]) => (
                  <div key={key} className="parameter-item">
                    <span className="parameter-key">{key}</span>
                    <span className="parameter-value">{value}</span>
                  </div>
                ))}
              </div>
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
    </div>
  );
};

export default StorageClassDetail;
