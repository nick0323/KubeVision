import React from 'react';
import { FaChevronRight } from 'react-icons/fa';
import './Breadcrumb.css';

export interface BreadcrumbItem {
  label: string;
  path?: string;
  icon?: React.ReactNode;
}

export interface BreadcrumbProps {
  items?: BreadcrumbItem[];
  namespace?: string;
  resourceType?: string;
  resourceName?: string;
  onNavigate?: (path: string) => void;
}

export const Breadcrumb: React.FC<BreadcrumbProps> = ({
  items = [],
  namespace,
  resourceType,
  resourceName,
  onNavigate,
}) => {
  // 构建面包屑项
  const buildBreadcrumbItems = (): BreadcrumbItem[] => {
    const builtItems: BreadcrumbItem[] = [];

    // 添加命名空间
    if (namespace && namespace !== '_all') {
      builtItems.push({
        label: namespace,
        path: `/namespaces/${namespace}`,
      });
    }

    // 添加资源类型
    if (resourceType) {
      builtItems.push({
        label: formatResourceType(resourceType),
        path: `/${resourceType}${namespace && namespace !== '_all' ? `/${namespace}` : ''}`,
      });
    }

    // 添加资源名称（最后一项，不可点击）
    if (resourceName && resourceName !== '_all') {
      builtItems.push({
        label: resourceName,
      });
    }

    // 合并自定义项
    if (items && items.length > 0) {
      return [...builtItems, ...items];
    }

    return builtItems;
  };

  const breadcrumbItems = buildBreadcrumbItems();

  // 格式化资源类型名称
  function formatResourceType(type: string): string {
    const typeMap: Record<string, string> = {
      pods: 'Pods',
      deployments: 'Deployments',
      statefulsets: 'StatefulSets',
      daemonsets: 'DaemonSets',
      services: 'Services',
      configmaps: 'ConfigMaps',
      secrets: 'Secrets',
      persistentvolumeclaims: 'PVCs',
      persistentvolumes: 'PVs',
      storageclasses: 'StorageClasses',
      ingresses: 'Ingresses',
      jobs: 'Jobs',
      cronjobs: 'CronJobs',
      nodes: 'Nodes',
      namespaces: 'Namespaces',
      events: 'Events',
    };
    return typeMap[type] || type;
  }

  return (
    <ol className="breadcrumb-list">
      {breadcrumbItems.map((item, index) => (
        <li key={index} className="breadcrumb-item">
          {item.path && onNavigate ? (
            <button className="breadcrumb-link" onClick={() => onNavigate(item.path!)}>
              {item.icon && <span className="breadcrumb-icon">{item.icon}</span>}
              {item.label}
            </button>
          ) : (
            <span className="breadcrumb-text">
              {item.icon && <span className="breadcrumb-icon">{item.icon}</span>}
              {item.label}
            </span>
          )}
          {index < breadcrumbItems.length - 1 && (
            <FaChevronRight className="breadcrumb-separator" />
          )}
        </li>
      ))}
    </ol>
  );
};

export default Breadcrumb;
