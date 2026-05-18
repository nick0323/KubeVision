import React from 'react';
import './StatusBadge.css';

interface StatusBadgeProps {
  status: string;
  resourceType?: string;
}

const StatusBadgeImpl: React.FC<StatusBadgeProps> = ({ status, resourceType }) => {
  const className = getStatusClass(status, resourceType);

  return (
    <span className={`status-badge ${className}`}>
      <span className="status-text">{status}</span>
    </span>
  );
};

function getStatusClass(status: string, resourceType?: string): string {
  if (resourceType === 'pod') {
    switch (status) {
      case 'Running':
        return 'status-running';
      case 'Pending':
        return 'status-pending';
      case 'Failed':
        return 'status-failed';
      case 'Succeeded':
        return 'status-succeeded';
      case 'Unknown':
        return 'status-unknown';
      case 'CrashLoopBackOff':
        return 'status-crashloop';
      case 'Error':
        return 'status-error';
      case 'Terminating':
        return 'status-terminating';
      default:
        return '';
    }
  }

  if (['deployment', 'statefulset', 'daemonset'].includes(resourceType || '')) {
    switch (status) {
      case 'Available':
        return 'status-healthy';
      case 'Partial':
        return 'status-partial';
      case 'Unavailable':
        return 'status-unavailable';
      case 'ScaledToZero':
        return 'status-scaled';
      case 'Progressing':
        return 'status-progressing';
      default:
        return '';
    }
  }

  if (resourceType === 'job') {
    switch (status) {
      case 'Running':
        return 'status-running';
      case 'Completed':
        return 'status-succeeded';
      case 'Failed':
        return 'status-failed';
      case 'Pending':
        return 'status-pending';
      default:
        return '';
    }
  }

  if (resourceType === 'cronjob') {
    switch (status) {
      case 'Active':
        return 'status-running';
      case 'Succeeded':
      case 'Suspended':
        return 'status-succeeded';
      case 'Pending':
      default:
        return 'status-pending';
    }
  }

  if (resourceType === 'node') {
    switch (status) {
      case 'Ready':
        return 'status-ready';
      case 'NotReady':
        return 'status-notready';
      case 'SchedulingDisabled':
        return 'status-disabled';
      case 'Unknown':
        return 'status-unknown';
      default:
        return '';
    }
  }

  if (resourceType === 'pvc') {
    switch (status) {
      case 'Bound':
        return 'status-bound';
      case 'Pending':
        return 'status-pending';
      case 'Lost':
        return 'status-lost';
      default:
        return '';
    }
  }

  if (resourceType === 'pv') {
    switch (status) {
      case 'Available':
        return 'status-available';
      case 'Bound':
        return 'status-bound';
      case 'Released':
        return 'status-released';
      case 'Failed':
        return 'status-failed';
      default:
        return '';
    }
  }

  if (resourceType === 'namespace') {
    switch (status) {
      case 'Active':
        return 'status-active';
      case 'Terminating':
        return 'status-terminating';
      default:
        return '';
    }
  }

  if (resourceType === 'event') {
    switch (status) {
      case 'Normal':
        return 'status-normal';
      case 'Warning':
        return 'status-warning';
      default:
        return '';
    }
  }

  if (resourceType === 'service') {
    switch (status) {
      case 'ClusterIP':
        return 'status-clusterip';
      case 'NodePort':
        return 'status-nodeport';
      case 'LoadBalancer':
        return 'status-loadbalancer';
      case 'ExternalName':
        return 'status-external';
      default:
        return '';
    }
  }

  return '';
}

export const StatusBadge = React.memo(StatusBadgeImpl);
export default StatusBadge;
