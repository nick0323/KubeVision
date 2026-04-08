import React, { useState, useCallback, useEffect } from 'react';
import { LoadingSpinner } from '../../LoadingSpinner';
import { ErrorDisplay } from '../../ErrorDisplay';
import { authFetch } from '../../../utils/auth';
import './ReplicaSetsTab.css';

interface ReplicaSet {
  metadata: {
    name: string;
    namespace: string;
    creationTimestamp: string;
  };
  spec: {
    replicas: number;
  };
  status: {
    readyReplicas: number;
    availableReplicas: number;
  };
}

interface ReplicaSetsTabProps {
  namespace: string;
  name: string;
  deployment: any | null;
}

/**
 * ReplicaSets Tab - 显示 Deployment 管理的 ReplicaSets
 */
export const ReplicaSetsTab: React.FC<ReplicaSetsTabProps> = ({ namespace, name, deployment }) => {
  const [replicaSets, setReplicaSets] = useState<ReplicaSet[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  // 加载 ReplicaSets
  const loadReplicaSets = useCallback(async () => {
    setLoading(true);
    setError(null);

    try {
      const response = await authFetch(`/api/replicasets/${namespace}?deployment=${name}`);
      const result = await response.json();

      if (result.code === 0 && result.data) {
        setReplicaSets(result.data.data || []);
      } else {
        setError(result.message || '加载失败');
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : '网络错误');
    } finally {
      setLoading(false);
    }
  }, [namespace, name]);

  useEffect(() => {
    loadReplicaSets();
  }, [loadReplicaSets]);

  if (loading) {
    return <LoadingSpinner text="加载 ReplicaSets..." size="lg" />;
  }

  if (error && replicaSets.length === 0) {
    return <ErrorDisplay message={error} type="error" showRetry onRetry={loadReplicaSets} />;
  }

  return (
    <div className="replicasets-tab">
      <div className="detail-card">
        <h3 className="detail-card-title">ReplicaSets</h3>
        {replicaSets.length === 0 ? (
          <div className="empty-state">
            <span className="empty-state-text">暂无 ReplicaSets</span>
          </div>
        ) : (
          <table className="detail-table">
            <thead>
              <tr>
                <th>名称</th>
                <th>副本数</th>
                <th>就绪副本</th>
                <th>可用副本</th>
                <th>创建时间</th>
              </tr>
            </thead>
            <tbody>
              {replicaSets.map((rs) => (
                <tr key={rs.metadata.name}>
                  <td>{rs.metadata.name}</td>
                  <td>{rs.spec.replicas}</td>
                  <td>{rs.status.readyReplicas || 0}</td>
                  <td>{rs.status.availableReplicas || 0}</td>
                  <td>{new Date(rs.metadata.creationTimestamp).toLocaleString()}</td>
                </tr>
              ))}
            </tbody>
          </table>
        )}
      </div>
    </div>
  );
};

export default ReplicaSetsTab;
