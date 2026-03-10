/**
 * 关联资源面板组件
 * 显示与当前资源关联的其他资源
 */
import React from 'react';
import { useNavigate } from 'react-router-dom';
import './RelatedResources.css';

interface RelatedResource {
  kind: string;
  name: string;
  namespace: string;
  status?: string;
  icon?: string;
}

interface RelatedResourcesProps {
  resources: RelatedResource[];
  title?: string;
  onResourceClick?: (kind: string, namespace: string, name: string) => void;
}

export const RelatedResources: React.FC<RelatedResourcesProps> = ({
  resources,
  title = '关联资源',
  onResourceClick,
}) => {
  const navigate = useNavigate();

  const handleClick = (resource: RelatedResource) => {
    if (onResourceClick) {
      onResourceClick(resource.kind, resource.namespace, resource.name);
    } else {
      const path = `/${resource.kind.toLowerCase()}s/${resource.namespace}/${resource.name}`;
      navigate(path);
    }
  };

  const getResourceIcon = (kind: string): string => {
    const icons: Record<string, string> = {
      Pod: '📦',
      Deployment: '🚀',
      StatefulSet: '📋',
      DaemonSet: '📊',
      Service: '🌐',
      ConfigMap: '📄',
      Secret: '🔒',
      PVC: '💾',
      PV: '🗄️',
      Ingress: '🔀',
      ServiceAccount: '🔑',
    };
    return icons[kind] || '📄';
  };

  const getStatusClass = (status?: string): string => {
    if (!status) return '';
    const statusLower = status.toLowerCase();
    if (statusLower.includes('running') || statusLower.includes('ready') || statusLower.includes('available')) {
      return 'status-good';
    }
    if (statusLower.includes('pending') || statusLower.includes('progressing')) {
      return 'status-warning';
    }
    if (statusLower.includes('failed') || statusLower.includes('error')) {
      return 'status-error';
    }
    return '';
  };

  if (resources.length === 0) {
    return (
      <div className="related-resources">
        <h4 className="related-title">{title}</h4>
        <div className="related-empty">
          <span className="empty-icon">🔍</span>
          <p>暂无关联资源</p>
        </div>
      </div>
    );
  }

  return (
    <div className="related-resources">
      <h4 className="related-title">{title}</h4>
      <div className="related-list">
        {resources.map((resource, index) => (
          <div
            key={index}
            className="related-item"
            onClick={() => handleClick(resource)}
          >
            <div className="related-icon">
              {getResourceIcon(resource.kind)}
            </div>
            <div className="related-info">
              <div className="related-header">
                <span className="related-kind">{resource.kind}</span>
                <span className="related-name">{resource.name}</span>
              </div>
              <div className="related-meta">
                <span className="related-namespace">{resource.namespace}</span>
                {resource.status && (
                  <span className={`related-status ${getStatusClass(resource.status)}`}>
                    {resource.status}
                  </span>
                )}
              </div>
            </div>
            <div className="related-arrow">→</div>
          </div>
        ))}
      </div>
    </div>
  );
};

export default RelatedResources;
