import React, { useState, useMemo } from 'react';
import Tippy from '@tippyjs/react';
import 'tippy.js/dist/tippy.css';
import 'tippy.js/themes/light.css';
import { OverviewTabProps } from '../pages/ResourceDetailPage.types';
import { StatusBadge } from '../common/StatusBadge';
import { formatRelativeTime } from '../utils/time';
import { truncateText } from '../utils/string';
import { RESOURCE_CONFIG, RESOURCE_FIELDS } from './overview/overviewConfig';
import { ContainerDetails } from './overview/ContainerDetails';
import './OverviewTab.css';
import '../styles/detail-page.css';

export const OverviewTab: React.FC<OverviewTabProps> = ({
  data,
  loading,
  resourceType = 'pod',
}) => {
  const [containersExpanded, setContainersExpanded] = useState<Record<string, boolean>>({});

  const resourceInfo = RESOURCE_CONFIG[resourceType] || { title: resourceType };

  const typedData = data as Record<string, any>;
  const metadata = typedData?.metadata || {};
  const spec = typedData?.spec || {};
  const status = typedData?.status || {};

  const containerStatuses = useMemo(() => {
    if (resourceType !== 'pod' || !typedData?.status?.containerStatuses) return [];
    return typedData.status.containerStatuses.map((cs: any) => ({
      name: cs.name,
      ready: cs.ready,
      restartCount: cs.restartCount,
      state: cs.state || {},
      image: cs.image,
    }));
  }, [data, resourceType]);

  const workloadContainers = useMemo(() => {
    if (!resourceInfo.hasContainers) return [];
    const podSpec = spec?.template?.spec || spec;
    const containers = podSpec?.containers || [];
    return containers.map((c: any) => ({
      name: c.name,
      image: c.image,
      ports: c.ports,
      env: c.env,
      resources: c.resources,
      livenessProbe: c.livenessProbe,
      readinessProbe: c.readinessProbe,
      volumeMounts: c.volumeMounts,
      imagePullPolicy: c.imagePullPolicy,
    }));
  }, [data, spec, resourceInfo, resourceType]);

  const readyContainers = useMemo(() => {
    if (resourceType !== 'pod') return null;
    const ready = containerStatuses.filter((c: any) => c.ready).length;
    const total = containerStatuses.length || spec?.containers?.length || 0;
    return { ready, total };
  }, [containerStatuses, spec, resourceType]);

  const restartCount = useMemo(() => {
    if (resourceType !== 'pod') return 0;
    return containerStatuses.reduce((sum: number, c: any) => sum + c.restartCount, 0);
  }, [containerStatuses, resourceType]);

  const conditions = useMemo(() => {
    return status?.conditions || [];
  }, [status]);

  const hasStatusOverview = useMemo(() => {
    return ['pod', 'deployment', 'statefulset'].includes(resourceType);
  }, [resourceType]);

  if (loading || !data) {
    return <div className="overview-tab-loading">Loading....</div>;
  }

  return (
    <div className="overview-tab">
      {hasStatusOverview && (
        <div className="detail-card">
          <h3 className="detail-card-title">Status Overview</h3>
          <div className="detail-card-body">
            <div className="stats-grid">
              {resourceType === 'pod' ? (
                <>
                  <div className="stat-card">
                    <div className="stat-value">
                      <StatusBadge status={status.phase || 'Unknown'} resourceType={resourceType} />
                    </div>
                    <div className="stat-label">Phase</div>
                  </div>
                  {readyContainers && (
                    <div className="stat-card">
                      <div className="stat-value">
                        {readyContainers.ready} / {readyContainers.total}
                      </div>
                      <div className="stat-label">Ready Containers</div>
                    </div>
                  )}
                  <div className="stat-card">
                    <div className="stat-value">{restartCount}</div>
                    <div className="stat-label">Restart Count</div>
                  </div>
                  <div className="stat-card">
                    <div className="stat-value">{spec?.nodeName || '-'}</div>
                    <div className="stat-label">Node</div>
                  </div>
                </>
              ) : (
                <>
                  <div className="stat-card">
                    <div className="stat-value">
                      {status.readyReplicas !== undefined ? (
                        <StatusBadge
                          status={
                            status.readyReplicas === status.replicas
                              ? 'Available'
                              : status.readyReplicas > 0
                                ? 'Partial'
                                : 'Unavailable'
                          }
                          resourceType={resourceType}
                        />
                      ) : status.phase ? (
                        <StatusBadge status={status.phase} resourceType={resourceType} />
                      ) : (
                        '-'
                      )}
                    </div>
                    <div className="stat-label">Status</div>
                  </div>
                  {status.replicas !== undefined && (
                    <div className="stat-card">
                      <div className="stat-value">{String(status.replicas)}</div>
                      <div className="stat-label">Replicas</div>
                    </div>
                  )}
                  {status.readyReplicas !== undefined && (
                    <div className="stat-card">
                      <div className="stat-value">{String(status.readyReplicas)}</div>
                      <div className="stat-label">Ready</div>
                    </div>
                  )}
                  {status.availableReplicas !== undefined && (
                    <div className="stat-card">
                      <div className="stat-value">{String(status.availableReplicas)}</div>
                      <div className="stat-label">Available</div>
                    </div>
                  )}
                </>
              )}
            </div>
          </div>
        </div>
      )}

      <div className="detail-card">
        <h3 className="detail-card-title">{resourceInfo.title} Information</h3>
        <div className="detail-card-body">
          <div className="info-grid">
            <div className="info-item">
              <span className="info-label">Created</span>
              <span className="info-value">
                {metadata.creationTimestamp
                  ? `${new Date(metadata.creationTimestamp).toLocaleString()} (${formatRelativeTime(metadata.creationTimestamp)})`
                  : '-'}
              </span>
            </div>
            {metadata.namespace && (
              <div className="info-item">
                <span className="info-label">Namespace</span>
                <span className="info-value">{metadata.namespace}</span>
              </div>
            )}
            <div className="info-item">
              <span className="info-label">UID</span>
              <span className="info-value">{metadata.uid || '-'}</span>
            </div>
            <div className="info-item">
              <span className="info-label">Resource Version</span>
              <span className="info-value">{metadata.resourceVersion || '-'}</span>
            </div>

            {metadata.ownerReferences && metadata.ownerReferences.length > 0 && (
              <div className="info-item">
                <span className="info-label">Owner</span>
                <span className="info-value clickable">
                  {metadata.ownerReferences[0].kind}/{metadata.ownerReferences[0].name}
                </span>
              </div>
            )}

            {RESOURCE_FIELDS[resourceType]?.map(field => {
              if (field.condition && !field.condition(data)) return null;
              const value = field.getValue(data);
              if (value === null || value === undefined || value === '') return null;
              return (
                <div key={field.key} className="info-item">
                  <span className="info-label">{field.label}</span>
                  <span className="info-value">
                    {field.render ? field.render(value, data) : value}
                  </span>
                </div>
              );
            })}
          </div>

          {(metadata.labels && Object.keys(metadata.labels).length > 0) ||
          (metadata.annotations && Object.keys(metadata.annotations).length > 0) ? (
            <div className="info-section">
              <div className="info-grid info-grid-2col">
                {metadata.labels && Object.keys(metadata.labels).length > 0 && (
                  <div className="info-item">
                    <span className="info-label">Labels</span>
                    <div className="label-list">
                      {Object.entries(metadata.labels)
                        .slice(0, 5)
                        .map(([key, value]) => {
                          const fullText = `${key}: ${value}`;
                          const displayKey = truncateText(key as string, 30);
                          const displayValue = truncateText(value as string, 30);
                          const isTruncated = displayKey !== key || displayValue !== value;
                          const labelElement = (
                            <span className="label-tag">
                              <span className="label-key">{displayKey}</span>
                              <span className="label-separator">: </span>
                              <span className="label-value">{displayValue}</span>
                            </span>
                          );
                          if (isTruncated) {
                            return (
                              <Tippy
                                key={key}
                                content={fullText}
                                theme="light"
                                placement="top"
                                arrow={true}
                                duration={200}
                              >
                                {labelElement}
                              </Tippy>
                            );
                          }
                          return <span key={key}>{labelElement}</span>;
                        })}
                      {Object.keys(metadata.labels).length > 5 && (
                        <Tippy
                          content={
                            <div style={{ maxHeight: '200px', overflow: 'auto' }}>
                              {Object.entries(metadata.labels).map(([key, value]) => (
                                <div key={key}>
                                  {key}: {String(value)}
                                </div>
                              ))}
                            </div>
                          }
                          theme="light"
                          placement="top"
                          arrow={true}
                          duration={200}
                          interactive={true}
                        >
                          <span className="label-tag label-more">
                            +{Object.keys(metadata.labels).length - 5} more
                          </span>
                        </Tippy>
                      )}
                    </div>
                  </div>
                )}

                {metadata.annotations && Object.keys(metadata.annotations).length > 0 && (
                  <div className="info-item">
                    <span className="info-label">Annotations</span>
                    <div className="annotation-list">
                      {Object.entries(metadata.annotations)
                        .slice(0, 5)
                        .map(([key, value]) => {
                          const fullText = `${key}: ${value}`;
                          const displayKey = truncateText(key as string, 30);
                          const displayValue = truncateText(value as string, 30);
                          const isTruncated = displayKey !== key || displayValue !== value;
                          const labelElement = (
                            <span className="annotation-tag">
                              <span className="annotation-key">{displayKey}</span>
                              <span className="annotation-separator">: </span>
                              <span className="annotation-value">{displayValue}</span>
                            </span>
                          );
                          if (isTruncated) {
                            return (
                              <Tippy
                                key={key}
                                content={fullText}
                                theme="light"
                                placement="top"
                                arrow={true}
                                duration={200}
                              >
                                {labelElement}
                              </Tippy>
                            );
                          }
                          return <span key={key}>{labelElement}</span>;
                        })}
                      {Object.keys(metadata.annotations).length > 5 && (
                        <Tippy
                          content={
                            <div style={{ maxHeight: '200px', overflow: 'auto' }}>
                              {Object.entries(metadata.annotations).map(([key, value]) => (
                                <div key={key}>
                                  {key}: {String(value)}
                                </div>
                              ))}
                            </div>
                          }
                          theme="light"
                          placement="top"
                          arrow={true}
                          duration={200}
                          interactive={true}
                        >
                          <span className="annotation-tag label-more">
                            +{Object.keys(metadata.annotations).length - 5} more
                          </span>
                        </Tippy>
                      )}
                    </div>
                  </div>
                )}
              </div>
            </div>
          ) : null}
        </div>
      </div>

      {containerStatuses.length > 0 && (
        <div className="detail-card">
          <h3 className="detail-card-title">Containers</h3>
          <div className="detail-card-body">
            {containerStatuses.map((container: any) => {
              const isExpanded = containersExpanded[container.name] || false;
              const containerSpec = spec?.containers?.find((c: any) => c.name === container.name);
              return (
                <div key={container.name} className="container-card">
                  <div
                    className="container-card-header"
                    onClick={() =>
                      setContainersExpanded(prev => ({
                        ...prev,
                        [container.name]: !isExpanded,
                      }))
                    }
                  >
                    <span className="collapse-btn">{isExpanded ? '▼' : '▶'}</span>
                    <span className="container-card-title">{container.name}</span>
                    <span className="container-card-image">{container.image}</span>
                    {containerSpec?.imagePullPolicy && (
                      <span className="container-card-pull-policy">
                        {containerSpec.imagePullPolicy}
                      </span>
                    )}
                  </div>
                  {isExpanded && <ContainerDetails container={containerSpec} />}
                </div>
              );
            })}
          </div>
        </div>
      )}

      {resourceInfo.hasContainers && resourceType !== 'pod' && workloadContainers.length > 0 && (
        <div className="detail-card">
          <h3 className="detail-card-title">Containers</h3>
          <div className="detail-card-body">
            {workloadContainers.map((container: any, index: number) => {
              const isExpanded = containersExpanded[container.name] || false;
              return (
                <div key={index} className="container-card">
                  <div
                    className="container-card-header"
                    onClick={() =>
                      setContainersExpanded(prev => ({
                        ...prev,
                        [container.name]: !isExpanded,
                      }))
                    }
                  >
                    <span className="collapse-btn">{isExpanded ? '▼' : '▶'}</span>
                    <span className="container-card-title">{container.name}</span>
                    <span className="container-card-image">{container.image}</span>
                    {container.imagePullPolicy && (
                      <span className="container-card-pull-policy">
                        {container.imagePullPolicy}
                      </span>
                    )}
                  </div>
                  {isExpanded && <ContainerDetails container={container} />}
                </div>
              );
            })}
          </div>
        </div>
      )}

      {conditions.length > 0 && (
        <div className="detail-card">
          <h3 className="detail-card-title">Conditions</h3>
          <div className="detail-card-body">
            <div className="condition-list">
              {conditions.map((condition: any) => (
                <div key={condition.type} className="condition-item">
                  <span className="condition-type">{condition.type}</span>
                  <span
                    className={`condition-status status-badge ${
                      condition.status === 'True'
                        ? 'success'
                        : condition.status === 'False'
                          ? 'error'
                          : 'warning'
                    }`}
                  >
                    {condition.status}
                  </span>
                  {condition.lastTransitionTime && (
                    <span className="condition-time">
                      {formatRelativeTime(condition.lastTransitionTime)}
                    </span>
                  )}
                </div>
              ))}
            </div>
          </div>
        </div>
      )}
    </div>
  );
};

export default OverviewTab;
