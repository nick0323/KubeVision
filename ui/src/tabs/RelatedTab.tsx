import React, { useState, useCallback, useEffect } from 'react';
import { RelatedTabProps } from '../pages/ResourceDetailPage.types';
import { LoadingSpinner } from '../common/LoadingSpinner';
import { ErrorDisplay } from '../common/ErrorDisplay';
import { authFetch } from '../utils/auth';
import { isClusterResource as isClusterScopeResource } from '../constants/config';

interface RelatedResource {
  kind: string;
  name: string;
  relation?: string;
}

// 关系Typelabel（英文）
const relationLabels: Record<string, string> = {
  owner: 'Owner',
  child: 'Child',
  selectedBy: 'Selected By',
  exposedBy: 'Exposed By',
  autoscaled: 'Autoscaled',
  protected: 'Protected By',
  routedBy: 'Routed By',
  volume: 'Volume',
  volumeClaim: 'Volume Claim',
  scheduledOn: 'Scheduled On',
  scheduled: 'Scheduled',
  headlessService: 'Headless Service',
  selects: 'Selects',
  endpoints: 'Endpoints',
  usedBy: 'Used By',
  boundPV: 'Bound PV',
  boundPVC: 'Bound PVC',
  storageClass: 'Storage Class',
  provisionedPVC: 'Provisioned PVC',
  provisionedPV: 'Provisioned PV',
  tlsSecret: 'TLS Secret',
  routesTo: 'Routes To',
  quota: 'Quota',
  contains: 'Contains',
};

/**
 * Related Tab - 关联resource
 */
export const RelatedTab: React.FC<RelatedTabProps> = ({ namespace, name, resourceType, ownerReferences }) => {
  const [relatedResources, setRelatedResources] = useState<RelatedResource[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  // Determine if cluster-level resource
  const isClusterResource = isClusterScopeResource(resourceType);

  // Loading...源
  useEffect(() => {
    const loadRelated = async () => {
      setLoading(true);
      setError(null);

      try {
        // 动态Build API 路径：cluster级resourceUse _cluster_，其他Use namespace
        const apiPath = isClusterResource
          ? `/api/${resourceType}/_cluster_/${name}/related`
          : `/api/${resourceType}/${namespace}/${name}/related`;
        const response = await authFetch(apiPath);
        const result = await response.json();

        if (result.code === 0 && result.data) {
          setRelatedResources(result.data || []);
        } else {
          // if没has关联resource API，Use ownerReferences
          const owners: RelatedResource[] = (ownerReferences || []).map(ref => ({
            kind: ref.kind,
            name: ref.name,
          }));
          setRelatedResources(owners);
        }
      } catch {
        // 降级Process：onlyDisplay ownerReferences
        const owners: RelatedResource[] = (ownerReferences || []).map(ref => ({
          kind: ref.kind,
          name: ref.name,
        }));
        setRelatedResources(owners);
      } finally {
        setLoading(false);
      }
    };

    loadRelated();
  }, [namespace, name, resourceType, ownerReferences, isClusterResource]);

  // Get关系labeltext
  const getRelationLabel = (relation?: string) => {
    if (!relation) return '';
    return relationLabels[relation] || relation;
  };

  if (loading) {
    return <LoadingSpinner text="Loading resources..." size="lg" />;
  }

  if (error && relatedResources.length === 0) {
    return (
      <ErrorDisplay
        message={error}
        type="error"
        showRetry
        onRetry={() => window.location.reload()}
      />
    );
  }

  return (
    <div className="related-tab">
      <div className="detail-card">
        <h3 className="detail-card-title">Related</h3>
        {relatedResources.length === 0 ? (
          <div className="empty-state">
            <span className="empty-state-text">No related resources found</span>
          </div>
        ) : (
          <table className="detail-table">
            <thead>
              <tr>
                <th>Kind</th>
                <th>Name</th>
                <th>Relation</th>
              </tr>
            </thead>
            <tbody>
              {relatedResources.map((resource, index) => (
                <tr key={index} className="table-row">
                  <td>
                    <span className="kind-badge">{resource.kind}</span>
                  </td>
                  <td>
                    <span className="related-resource-name">
                      {resource.name}
                    </span>
                  </td>
                  <td>
                    {resource.relation && (
                      <span className="relation-badge">{getRelationLabel(resource.relation)}</span>
                    )}
                  </td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </div>
    </div>
  );
};

export default RelatedTab;
