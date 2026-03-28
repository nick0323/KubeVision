import React, { useState, useCallback, useEffect } from 'react';
import { RelatedTabProps, RelatedResource } from '../types';
import { LoadingSpinner } from '../../LoadingSpinner';
import { ErrorDisplay } from '../../ErrorDisplay';
import { authFetch } from '../../../utils/auth';
import { useNavigate } from 'react-router-dom';
import './RelatedTab.css';

const RESOURCE_ICONS: Record<string, string> = {
  ReplicaSet: '📦',
  Deployment: '🚀',
  StatefulSet: '📋',
  DaemonSet: '📱',
  Service: '🌐',
  Ingress: '🌍',
  PersistentVolumeClaim: '💾',
  ConfigMap: '⚙️',
  Secret: '🔐',
  Node: '🖥️',
  Namespace: '📁',
};

const RESOURCE_CATEGORY: Record<string, string> = {
  ReplicaSet: 'ownership',
  Deployment: 'ownership',
  StatefulSet: 'ownership',
  DaemonSet: 'ownership',
  Service: 'network',
  Ingress: 'network',
  PersistentVolumeClaim: 'storage',
  ConfigMap: 'configuration',
  Secret: 'configuration',
  Node: 'infrastructure',
  Namespace: 'infrastructure',
};

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
        const response = await authFetch(`/api/pods/${namespace}/${name}/related`);
        const result = await response.json();
        
        if (result.code === 0 && result.data) {
          setRelatedResources(result.data || []);
        } else {
          // 如果没有关联资源 API，使用 ownerReferences
          const owners: RelatedResource[] = ownerReferences.map(ref => ({
            kind: ref.kind,
            name: ref.name,
            apiVersion: ref.apiVersion,
          }));
          setRelatedResources(owners);
        }
      } catch (err) {
        // 降级处理：只显示 ownerReferences
        const owners: RelatedResource[] = ownerReferences.map(ref => ({
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
  const handleResourceClick = useCallback((resource: RelatedResource) => {
    const resourceType = resource.kind.toLowerCase() + 's';
    navigate(`/${resourceType}/${namespace}/${resource.name}`);
  }, [namespace, navigate]);

  // 导出资源列表
  const handleExport = useCallback(() => {
    const content = relatedResources.map(r => `${r.kind}: ${r.name}`).join('\n');
    const blob = new Blob([content], { type: 'text/plain' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = `${name}-related-resources.txt`;
    a.click();
    URL.revokeObjectURL(url);
  }, [relatedResources, name]);

  // 按类别分组
  const groupedResources = relatedResources.reduce((acc, resource) => {
    const category = RESOURCE_CATEGORY[resource.kind] || 'other';
    if (!acc[category]) {
      acc[category] = [];
    }
    acc[category].push(resource);
    return acc;
  }, {} as Record<string, RelatedResource[]>);

  if (loading) {
    return <LoadingSpinner text="加载关联资源..." size="lg" />;
  }

  if (error && relatedResources.length === 0) {
    return <ErrorDisplay message={error} type="error" showRetry onRetry={() => window.location.reload()} />;
  }

  if (relatedResources.length === 0) {
    return (
      <div className="related-tab">
        <div className="empty-state">
          <span className="empty-state-icon">🔗</span>
          <span className="empty-state-text">暂无关联资源</span>
        </div>
      </div>
    );
  }

  return (
    <div className="related-tab">
      {/* 操作按钮 */}
      <div className="related-actions">
        <button className="toolbar-btn" onClick={handleExport}>📥 导出</button>
        <button className="toolbar-btn" onClick={() => window.location.reload()}>🔄 刷新</button>
      </div>

      {/* 资源列表 */}
      {Object.entries(groupedResources).map(([category, resources]) => (
        <div key={category} className="related-section">
          <div className="related-section-title">
            {category === 'ownership' && '📦 OWNERSHIP'}
            {category === 'network' && '🌐 NETWORK'}
            {category === 'storage' && '💾 STORAGE'}
            {category === 'configuration' && '⚙️ CONFIGURATION'}
            {category === 'infrastructure' && '🖥️ INFRASTRUCTURE'}
            {category === 'other' && '🔗 OTHER'}
          </div>
          
          <div className="related-list">
            {resources.map((resource, index) => (
              <div
                key={index}
                className="related-item"
                onClick={() => handleResourceClick(resource)}
              >
                <span className="related-item-icon">
                  {RESOURCE_ICONS[resource.kind] || '📄'}
                </span>
                <span className="related-item-type">{resource.kind}</span>
                <span className="related-item-name">{resource.name}</span>
              </div>
            ))}
          </div>
        </div>
      ))}
    </div>
  );
};

export default RelatedTab;
