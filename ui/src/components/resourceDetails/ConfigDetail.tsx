/**
 * ConfigMap/Secret Detail Component
 */
import React, { useState } from 'react';
import { FaFileAlt, FaLock, FaBook, FaCalendarAlt } from 'react-icons/fa';
import { LabelList } from '../LabelList';
import { EventTimeline } from '../EventTimeline';
import { CollapsibleSection } from '../CollapsibleSection';
import { ResourceReferences } from '../ResourceReferences';
import './ResourceDetail.css';

interface ConfigDetailProps {
  data: any;
  isSecret?: boolean;
  events?: any[];
  onResourceClick?: (type: string, namespace: string, name: string) => void;
}

export const ConfigDetail: React.FC<ConfigDetailProps> = ({
  data,
  isSecret = false,
  events = [],
  onResourceClick,
}) => {
  if (!data) return null;

  const { metadata, data: configData, binaryData } = data;
  const [revealedKeys, setRevealedKeys] = useState<Set<string>>(new Set());
  const [copiedKey, setCopiedKey] = useState<string | null>(null);

  const toggleReveal = (key: string) => {
    setRevealedKeys((prev) => {
      const next = new Set(prev);
      if (next.has(key)) {
        next.delete(key);
      } else {
        next.add(key);
      }
      return next;
    });
  };

  const copyValue = async (key: string, value: string) => {
    try {
      await navigator.clipboard.writeText(value);
      setCopiedKey(key);
      setTimeout(() => setCopiedKey(null), 2000);
    } catch (err) {
      console.error('Copy failed:', err);
    }
  };

  const decodeValue = (value: string) => {
    try {
      return atob(value);
    } catch {
      return value;
    }
  };

  const getDisplayValue = (key: string, value: string) => {
    if (isSecret && !revealedKeys.has(key)) {
      return '••••••••';
    }
    return isSecret ? decodeValue(value) : value;
  };

  const renderDataItems = (items: Record<string, string>, title: string) => {
    if (!items || Object.keys(items).length === 0) return null;

    return (
      <div className="data-section">
        <div className="subsection-title">{title}</div>
        <div className="data-list">
          {Object.entries(items).map(([key, value]: [string, any]) => {
            const displayValue = getDisplayValue(key, String(value));
            const isBinary = !isSecret && typeof value !== 'string';
            const isLong = displayValue.length > 100;

            return (
              <div key={key} className="data-item">
                <div className="data-header">
                  <span className="data-key">{key}</span>
                  <div className="data-actions">
                    {isSecret && (
                      <button
                        className="btn-icon"
                        onClick={() => toggleReveal(key)}
                        title={revealedKeys.has(key) ? 'Hide' : 'Show'}
                      >
                        {revealedKeys.has(key) ? '🙈' : '👁️'}
                      </button>
                    )}
                    <button
                      className="btn-icon"
                      onClick={() => copyValue(key, displayValue)}
                      title="Copy"
                    >
                      {copiedKey === key ? '✅' : '📋'}
                    </button>
                  </div>
                </div>
                <div className={`data-value ${isLong ? 'data-value-long' : ''}`}>
                  {isBinary ? (
                    <span className="binary-indicator">
                      📦 Binary data ({value.length} bytes)
                    </span>
                  ) : (
                    <pre className="code-block">
                      {displayValue}
                    </pre>
                  )}
                </div>
              </div>
            );
          })}
        </div>
      </div>
    );
  };

  return (
    <div className="resource-detail-content">
      <CollapsibleSection title={<>{isSecret ? <FaLock className="section-icon" /> : <FaFileAlt className="section-icon" />} {isSecret ? 'Secret' : 'ConfigMap'} Information</>} defaultExpanded={true}>
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
            <div className="info-value">{isSecret ? (data.type || 'Opaque') : 'ConfigMap'}</div>
          </div>
          <div className="info-item">
            <div className="info-label">Data Items</div>
            <div className="info-value">{Object.keys(configData || {}).length}</div>
          </div>
        </div>
      </CollapsibleSection>

      <CollapsibleSection title={<><FaBook className="section-icon" /> Data</>} defaultExpanded={true}>
        {renderDataItems(configData, isSecret ? 'Secret Data' : 'Data')}
        {binaryData && Object.keys(binaryData).length > 0 && (
          <div style={{ marginTop: '16px' }}>
            {renderDataItems(binaryData, 'Binary Data')}
          </div>
        )}
        {(!configData || Object.keys(configData).length === 0) &&
          (!binaryData || Object.keys(binaryData).length === 0) && (
            <div className="empty-state">No data</div>
          )}
      </CollapsibleSection>

      <CollapsibleSection title={<><FaBook className="section-icon" /> References</>} defaultExpanded={false}>
        <ResourceReferences
          name={metadata?.name}
          namespace={metadata?.namespace}
          type={isSecret ? 'secret' : 'configmap'}
          onResourceClick={onResourceClick}
        />
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

export default ConfigDetail;
