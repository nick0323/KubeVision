import React, { useState, useCallback, useEffect } from 'react';
import { LoadingSpinner } from '../../LoadingSpinner';
import { ErrorDisplay } from '../../ErrorDisplay';
import { authFetch } from '../../../utils/auth';
import './PodsTab.css';

interface Pod {
  metadata: {
    name: string;
    namespace: string;
  };
  status: {
    phase: string;
  };
}

interface PodsTabProps {
  namespace: string;
  resourceName: string;
  resourceKind: string;
  ownerReferences?: any[];
}

/**
 * Pods Tab - 显示资源关联的 Pods
 */
export const PodsTab: React.FC<PodsTabProps> = ({ namespace, resourceName, resourceKind, ownerReferences }) => {
  const [pods, setPods] = useState<Pod[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  // 加载 Pods
  const loadPods = useCallback(async () => {
    setLoading(true);
    setError(null);

    try {
      const response = await authFetch(`/api/pods/${namespace}`);
      const result = await response.json();

      if (result.code === 0 && result.data) {
        // 过滤出关联的 Pods
        const allPods = result.data.data || [];
        const relatedPods = allPods.filter((pod: any) => {
          const owners = pod.metadata?.ownerReferences || [];
          return owners.some((owner: any) => owner.kind === resourceKind && owner.name === resourceName);
        });
        setPods(relatedPods);
      } else {
        setError(result.message || '加载失败');
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : '网络错误');
    } finally {
      setLoading(false);
    }
  }, [namespace, resourceName, resourceKind]);

  useEffect(() => {
    loadPods();
  }, [loadPods]);

  if (loading) {
    return <LoadingSpinner text="加载 Pods..." size="lg" />;
  }

  if (error && pods.length === 0) {
    return <ErrorDisplay message={error} type="error" showRetry onRetry={loadPods} />;
  }

  return (
    <div className="pods-tab">
      <div className="detail-card">
        <h3 className="detail-card-title">关联的 Pods</h3>
        {pods.length === 0 ? (
          <div className="empty-state">
            <span className="empty-state-text">暂无 Pods</span>
          </div>
        ) : (
          <table className="detail-table">
            <thead>
              <tr>
                <th>名称</th>
                <th>命名空间</th>
                <th>状态</th>
              </tr>
            </thead>
            <tbody>
              {pods.map((pod) => (
                <tr key={pod.metadata.name}>
                  <td>{pod.metadata.name}</td>
                  <td>{pod.metadata.namespace}</td>
                  <td>{pod.status.phase}</td>
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
