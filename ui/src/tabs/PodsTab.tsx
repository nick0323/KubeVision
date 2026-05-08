import React, { useState, useCallback, useEffect } from 'react';
import { useNavigate } from 'react-router-dom';
import { LoadingSpinner } from '../common/LoadingSpinner';
import { ErrorDisplay } from '../common/ErrorDisplay';
import { StatusBadge } from '../common/StatusBadge';
import { authFetch } from '../utils/auth';
import './PodsTab.css';

interface Pod {
  namespace: string;
  name: string;
  status: string;
  ready?: string;
  restarts?: number;
  age?: string;
  podIP?: string;
  nodeName?: string;
}

interface PodsTabProps {
  namespace: string;
  resourceName: string;
  resourceKind: string;
  resourceLabels?: Record<string, string>;
  resourceUid?: string;
}

/**
 * Pods Tab - Displayresource关联's Pods
 * Useresource's Label Selector query关联's Pod
 */
export const PodsTab: React.FC<PodsTabProps> = ({
  namespace,
  resourceName,
  resourceKind,
  resourceLabels,
  resourceUid,
}) => {
  const navigate = useNavigate();
  const [pods, setPods] = useState<Pod[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  // jump toto Pod detail page
  const handlePodClick = useCallback((podNamespace: string, podName: string) => {
    navigate(`/pod/${podNamespace}/${podName}`);
  }, [navigate]);

  // fromresource Labels Build Label Selector
  const buildLabelSelector = useCallback(() => {
    if (!resourceLabels || Object.keys(resourceLabels).length === 0) {
      return null;
    }

    // will resourceLabels 转换for label selector format
    // 例like：{ app: 'nginx', tier: 'frontend' } => "app=nginx,tier=frontend"
    const selectors = Object.entries(resourceLabels)
      .filter(([key, value]) => value !== undefined && value !== null)
      .map(([key, value]) => `${key}=${value}`);

    return selectors.length > 0 ? selectors.join(',') : null;
  }, [resourceLabels]);

  // Loading...ds
  const loadPods = useCallback(async () => {
    setLoading(true);
    setError(null);

    try {
      // preferUse Label Selector query
      const labelSelector = buildLabelSelector();

      // Build URL - namespace 作for query 参数传递
      let url = `/api/pod?limit=1000`;

      // Node page：Use fieldSelector=spec.nodeName=node1
      if (resourceKind === 'Node' || resourceKind === 'node') {
        url += `&fieldSelector=spec.nodeName=${encodeURIComponent(resourceName)}`;
      } else if (labelSelector) {
        url += `&namespace=${namespace}&labelSelector=${encodeURIComponent(labelSelector)}`;
      }

      // 后端按 owner UID 精确过滤，避免不同资源相同 label 导致误匹配
      if (resourceUid) {
        url += `&ownerUid=${encodeURIComponent(resourceUid)}`;
      }

      const response = await authFetch(url);
      const result = await response.json();

      if (result.code === 0 && result.data) {
        let allPods = Array.isArray(result.data) ? result.data : result.data.data || [];
        setPods(allPods);
      } else {
        setError(result.message || 'Load Failed');
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Network error');
    } finally {
      setLoading(false);
    }
  }, [namespace, resourceName, resourceKind, resourceUid, buildLabelSelector]);

  useEffect(() => {
    loadPods();
  }, [loadPods]);

  if (loading) {
    return <LoadingSpinner text="Loading...ds..." size="lg" />;
  }

  if (error && pods.length === 0) {
    return <ErrorDisplay message={error} type="error" showRetry onRetry={loadPods} />;
  }

  return (
    <div className="pods-tab">
      <div className="detail-card">
        <h3 className="detail-card-title">Pods</h3>
        {pods.length === 0 ? (
          <div className="empty-state">
            <span className="empty-state-text">No Pods</span>
          </div>
        ) : (
          <table className="detail-table">
            <thead>
              <tr>
                <th style={{ width: '25%' }}>Name</th>
                <th style={{ width: '15%' }}>Namespace</th>
                <th style={{ width: '15%' }}>Status</th>
                <th style={{ width: '8%' }}>Ready</th>
                <th style={{ width: '8%' }}>Restarts</th>
                <th style={{ width: '11%' }}>IP</th>
                <th style={{ width: '10%' }}>Node</th>
                <th style={{ width: '8%' }}>Age</th>
              </tr>
            </thead>
            <tbody>
              {pods.map(pod => (
                <tr key={pod.name} className="clickable-row" onClick={() => handlePodClick(pod.namespace, pod.name)}>
                  <td>
                    <span className="resource-name-link">{pod.name}</span>
                  </td>
                  <td>{pod.namespace}</td>
                  <td>
                    <StatusBadge status={pod.status} resourceType="pod" />
                  </td>
                  <td>{pod.ready || '-'}</td>
                  <td>{pod.restarts ?? '-'}</td>
                  <td>{pod.podIP || '-'}</td>
                  <td>{pod.nodeName || '-'}</td>
                  <td>{pod.age || '-'}</td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </div>
    </div>
  );
};

export default PodsTab;
