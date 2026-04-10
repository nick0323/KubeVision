import React, { useState, useCallback, useEffect } from 'react';
import { LoadingSpinner } from '../../../ui/LoadingSpinner';
import { ErrorDisplay } from '../../../ui/ErrorDisplay';
import { StatusBadge } from '../../../ui/StatusBadge';
import { authFetch } from '../../../../utils/auth';
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
  resourceLabels?: Record<string, string>; // 资源的 Labels
  ownerReferences?: any[];
}

/**
 * Pods Tab - 显示资源关联的 Pods
 * 使用资源的 Label Selector 查询关联的 Pod
 */
export const PodsTab: React.FC<PodsTabProps> = ({
  namespace,
  resourceName,
  resourceKind,
  resourceLabels,
  ownerReferences,
}) => {
  const [pods, setPods] = useState<Pod[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  // 从资源 Labels 构建 Label Selector
  const buildLabelSelector = useCallback(() => {
    if (!resourceLabels || Object.keys(resourceLabels).length === 0) {
      return null;
    }

    // 将 resourceLabels 转换为 label selector 格式
    // 例如：{ app: 'nginx', tier: 'frontend' } => "app=nginx,tier=frontend"
    const selectors = Object.entries(resourceLabels)
      .filter(([key, value]) => value !== undefined && value !== null)
      .map(([key, value]) => `${key}=${value}`);

    return selectors.length > 0 ? selectors.join(',') : null;
  }, [resourceLabels]);

  // 加载 Pods
  const loadPods = useCallback(async () => {
    setLoading(true);
    setError(null);

    try {
      // 优先使用 Label Selector 查询
      const labelSelector = buildLabelSelector();

      // 构建 URL - namespace 作为 query 参数传递
      let url = `/api/pod?limit=1000`;

      // Node 页面：使用 fieldSelector=spec.nodeName=node1
      if (resourceKind === 'Node' || resourceKind === 'node') {
        url += `&fieldSelector=spec.nodeName=${encodeURIComponent(resourceName)}`;
      } else if (labelSelector) {
        // Workload (Deployment/StatefulSet/DaemonSet/Job) 和 Service:
        // 使用 labelSelector 参数查询（支持多个 selector，逗号分隔）
        url += `&namespace=${namespace}&labelSelector=${encodeURIComponent(labelSelector)}`;
      }

      const response = await authFetch(url);
      const result = await response.json();

      if (result.code === 0 && result.data) {
        // 后端返回格式：{ code: 0, data: [...], page: {...} }
        // result.data 直接是数组
        let allPods = Array.isArray(result.data) ? result.data : result.data.data || [];

        // 如果没有 label selector，使用 ownerReferences 二次过滤
        if (!labelSelector && ownerReferences && ownerReferences.length > 0) {
          allPods = allPods.filter((pod: any) => {
            const owners = pod.metadata?.ownerReferences || [];
            return owners.some(
              (owner: any) => owner.kind === resourceKind && owner.name === resourceName
            );
          });
        }

        setPods(allPods);
      } else {
        setError(result.message || '加载失败');
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : '网络错误');
    } finally {
      setLoading(false);
    }
  }, [namespace, resourceName, resourceKind, ownerReferences, buildLabelSelector]);

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
        <h3 className="detail-card-title">Pods</h3>
        {pods.length === 0 ? (
          <div className="empty-state">
            <span className="empty-state-text">暂无 Pods</span>
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
                <tr key={pod.name}>
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
