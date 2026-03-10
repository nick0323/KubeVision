/**
 * 资源引用检测组件
 * 显示 ConfigMap/Secret 被哪些资源引用
 */
import React, { useState, useEffect } from 'react';
import { authFetch } from '../utils/auth';
import './ResourceReferences.css';

interface ReferenceInfo {
  kind: string;
  name: string;
  namespace: string;
  refType: string;
  field: string;
}

interface ResourceReferencesProps {
  name: string;
  namespace: string;
  type: 'configmap' | 'secret';
  onResourceClick?: (kind: string, namespace: string, name: string) => void;
}

export const ResourceReferences: React.FC<ResourceReferencesProps> = ({
  name,
  namespace,
  type,
  onResourceClick,
}) => {
  const [references, setReferences] = useState<ReferenceInfo[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    const loadReferences = async () => {
      setLoading(true);
      setError(null);

      try {
        const endpoint = `/api/${type}s/${namespace}/${name}/references`;
        const response = await authFetch(endpoint);
        const result = await response.json();

        if (result.code === 0 && result.data) {
          setReferences(result.data.references || []);
        } else {
          setError(result.message || '加载引用失败');
        }
      } catch (err) {
        setError('加载引用失败');
        console.error('加载引用失败:', err);
      } finally {
        setLoading(false);
      }
    };

    loadReferences();
  }, [name, namespace, type]);

  // 获取资源类型的图标
  const getResourceIcon = (kind: string): string => {
    const icons: Record<string, string> = {
      Pod: '📦',
      Deployment: '🚀',
      StatefulSet: '📋',
      DaemonSet: '📊',
      ServiceAccount: '🔑',
      Ingress: '🌐',
    };
    return icons[kind] || '📄';
  };

  // 获取引用类型的显示文本
  const getRefTypeLabel = (refType: string): string => {
    const labels: Record<string, string> = {
      volume: '📁 卷挂载',
      configMapKeyRef: '🔑 环境变量 (KeyRef)',
      configMapRef: '📝 环境变量 (All)',
      secretKeyRef: '🔑 密钥引用',
      imagePullSecret: '🖼️ 镜像拉取密钥',
      secret: '🔒 Secret',
      tls: '🔒 TLS 证书',
    };
    return labels[refType] || refType;
  };

  if (loading) {
    return (
      <div className="resource-references">
        <div className="info-card">
          <h4 className="card-title">🔗 被引用情况</h4>
          <div className="loading-state">加载中...</div>
        </div>
      </div>
    );
  }

  if (error) {
    return (
      <div className="resource-references">
        <div className="info-card">
          <h4 className="card-title">🔗 被引用情况</h4>
          <div className="error-state">⚠️ {error}</div>
        </div>
      </div>
    );
  }

  if (references.length === 0) {
    return (
      <div className="resource-references">
        <div className="info-card">
          <h4 className="card-title">🔗 被引用情况</h4>
          <div className="empty-state">
            <span className="empty-icon">🔍</span>
            <p>暂无引用</p>
            <p className="empty-desc">
              没有资源引用此 {type === 'configmap' ? 'ConfigMap' : 'Secret'}
            </p>
          </div>
        </div>
      </div>
    );
  }

  // 按资源类型分组
  const groupedByKind = references.reduce((acc, ref) => {
    if (!acc[ref.kind]) {
      acc[ref.kind] = [];
    }
    acc[ref.kind].push(ref);
    return acc;
  }, {} as Record<string, ReferenceInfo[]>);

  return (
    <div className="resource-references">
      <div className="info-card">
        <h4 className="card-title">
          🔗 被引用情况 ({references.length} 个引用)
        </h4>

        <div className="references-summary">
          {Object.entries(groupedByKind).map(([kind, refs]) => (
            <div key={kind} className="kind-group">
              <div className="kind-header">
                <span className="kind-icon">{getResourceIcon(kind)}</span>
                <span className="kind-name">{kind}</span>
                <span className="kind-count">{refs.length}</span>
              </div>
              <div className="ref-list">
                {refs.map((ref, idx) => (
                  <div key={idx} className="ref-item">
                    <div className="ref-header">
                      <span
                        className="ref-name link"
                        onClick={() =>
                          onResourceClick?.(
                            ref.kind.toLowerCase() + 's',
                            ref.namespace,
                            ref.name
                          )
                        }
                      >
                        {ref.name}
                      </span>
                      <span className="ref-namespace">{ref.namespace}</span>
                    </div>
                    <div className="ref-type">
                      {getRefTypeLabel(ref.refType)}
                    </div>
                    {ref.field && (
                      <div className="ref-field code">{ref.field}</div>
                    )}
                  </div>
                ))}
              </div>
            </div>
          ))}
        </div>
      </div>
    </div>
  );
};

export default ResourceReferences;
