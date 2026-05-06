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
  // Build breadcrumb items
  const buildBreadcrumbItems = (): BreadcrumbItem[] => {
    const builtItems: BreadcrumbItem[] = [];

    // Add namespace
    if (namespace && namespace !== '_all') {
      builtItems.push({
        label: namespace,
        path: `/namespaces/${namespace}`,
      });
    }

    // Add resource type
    if (resourceType) {
      builtItems.push({
        label: formatResourceType(resourceType),
        path: `/${resourceType}${namespace && namespace !== '_all' ? `/${namespace}` : ''}`,
      });
    }

    // AddResource name(last item, not clickable)
    if (resourceName && resourceName !== '_all') {
      builtItems.push({
        label: resourceName,
      });
    }

    // Merge custom items
    if (items && items.length > 0) {
      return [...builtItems, ...items];
    }

    return builtItems;
  };

  const breadcrumbItems = buildBreadcrumbItems();

  // Format resource type name
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
