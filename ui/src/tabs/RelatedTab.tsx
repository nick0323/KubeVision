import React, { useState, useCallback, useEffect } from 'react';
import { RelatedTabProps } from '../resources/types';
import { LoadingSpinner } from '../common/LoadingSpinner';
import { ErrorDisplay } from '../common/ErrorDisplay';
import { authFetch } from '../utils/auth';
import { useNavigate } from 'react-router-dom';

interface RelatedResource {
  kind: string;
  name: string;
  namespace?: string;
  apiVersion?: string;
}

/**
 * Related Tab - 关联资源
 */
export const RelatedTab: React.FC<RelatedTabProps> = ({ namespace, name, ownerReferences }) => {
  const navigate = useNavigate();
  const [relatedResources, setRelatedResources] = useState<RelatedResource[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  // 加载关联资源
  useEffect(() => {
    const loadRelated = async () => {
      setLoading(true);
      setError(null);

      try {
        // 后端只支持单数形式：/api/pod/ns/name/related
        const response = await authFetch(`/api/pod/${namespace}/${name}/related`);
        const result = await response.json();

        if (result.code === 0 && result.data) {
          setRelatedResources(result.data || []);
        } else {
          // 如果没有关联资源 API，使用 ownerReferences
          const owners: RelatedResource[] = (ownerReferences || []).map(ref => ({
            kind: ref.kind,
            name: ref.name,
            apiVersion: ref.apiVersion,
          }));
          setRelatedResources(owners);
        }
      } catch {
        // 降级处理：只显示 ownerReferences
        const owners: RelatedResource[] = (ownerReferences || []).map(ref => ({
          kind: ref.kind,
          name: ref.name,
          apiVersion: ref.apiVersion,
        }));
        setRelatedResources(owners);
      } finally {
        setLoading(false);
      }
    };

    loadRelated();
  }, [namespace, name, ownerReferences]);

  // 跳转到资源详情
  const handleResourceClick = useCallback(
    (kind: string, name: string) => {
      const resourceType = kind.toLowerCase() + 's';
      navigate(`/${resourceType}/${namespace}/${name}`);
    },
    [namespace, navigate]
  );

  if (loading) {
    return <LoadingSpinner text="加载关联资源..." size="lg" />;
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
                <th style={{ width: '200px' }}>Kind</th>
                <th>Name</th>
              </tr>
            </thead>
            <tbody>
              {relatedResources.map((resource, index) => (
                <tr key={index} className="table-row">
                  <td>
                    <span className="kind-badge">{resource.kind}</span>
                  </td>
                  <td>
                    <span
                      className="related-resource-name"
                      onClick={() => handleResourceClick(resource.kind, resource.name)}
                    >
                      {resource.name}
                    </span>
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
