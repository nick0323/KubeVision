import React from 'react';
import './StatusBadge.css';

interface StatusBadgeProps {
  status: string;
  resourceType?: string;
}

/**
 * Status标识Component
 * according to K8s resourceStatusDisplaynot同颜色andicon
 */
export const StatusBadge: React.FC<StatusBadgeProps> = ({ status, resourceType }) => {
  const statusInfo = getStatusInfo(status, resourceType);

  return (
    <span className={`status-badge ${statusInfo.className}`}>
      <span className="status-icon">{statusInfo.icon}</span>
      <span className="status-text">{status}</span>
    </span>
  );
};

/**
 * GetStatusinfo（颜色andicon）
 */
function getStatusInfo(status: string, resourceType?: string): { className: string; icon: string } {
  // Pod Status
  if (resourceType === 'pod') {
    switch (status) {
      case 'Running':
        return { className: 'status-running', icon: '🟢' };
      case 'Pending':
        return { className: 'status-pending', icon: '🟡' };
      case 'Failed':
        return { className: 'status-failed', icon: '🔴' };
      case 'Succeeded':
        return { className: 'status-succeeded', icon: '🔵' };
      case 'Unknown':
        return { className: 'status-unknown', icon: '⚪' };
      case 'CrashLoopBackOff':
        return { className: 'status-crashloop', icon: '🔴' };
      case 'Error':
        return { className: 'status-error', icon: '🔴' };
      case 'Terminating':
        return { className: 'status-terminating', icon: '🟡' };
      default:
        return { className: '', icon: '⚪' };
    }
  }

  // Workload Status（Deployment/StatefulSet/DaemonSet）
  if (['deployment', 'statefulset', 'daemonset'].includes(resourceType || '')) {
    switch (status) {
      case 'Available':
        return { className: 'status-healthy', icon: '🟢' };
      case 'Partial':
        return { className: 'status-partial', icon: '🟡' };
      case 'Unavailable':
        return { className: 'status-unavailable', icon: '🔴' };
      case 'ScaledToZero':
        return { className: 'status-scaled', icon: '🔵' };
      case 'Progressing':
        return { className: 'status-progressing', icon: '🟡' };
      default:
        return { className: '', icon: '⚪' };
    }
  }

  // Job Status
  if (resourceType === 'job') {
    switch (status) {
      case 'Running':
        return { className: 'status-running', icon: '🟡' };
      case 'Completed':
        return { className: 'status-succeeded', icon: '🟢' };
      case 'Failed':
        return { className: 'status-failed', icon: '🔴' };
      case 'Pending':
        return { className: 'status-pending', icon: '⚪' };
      default:
        return { className: '', icon: '⚪' };
    }
  }

  // CronJob Status
  if (resourceType === 'cronjob') {
    switch (status) {
      case 'Active':
        return { className: 'status-running', icon: '🟢' };
      case 'Succeeded':
      case 'Suspended':
        return { className: 'status-succeeded', icon: '🔵' };
      case 'Pending':
      default:
        return { className: 'status-pending', icon: '🟡' };
    }
  }

  // Node Status
  if (resourceType === 'node') {
    switch (status) {
      case 'Ready':
        return { className: 'status-ready', icon: '🟢' };
      case 'NotReady':
        return { className: 'status-notready', icon: '🔴' };
      case 'SchedulingDisabled':
        return { className: 'status-disabled', icon: '🟡' };
      case 'Unknown':
        return { className: 'status-unknown', icon: '⚪' };
      default:
        return { className: '', icon: '⚪' };
    }
  }

  // PVC Status
  if (resourceType === 'pvc') {
    switch (status) {
      case 'Bound':
        return { className: 'status-bound', icon: '🟢' };
      case 'Pending':
        return { className: 'status-pending', icon: '🟡' };
      case 'Lost':
        return { className: 'status-lost', icon: '🔴' };
      default:
        return { className: '', icon: '⚪' };
    }
  }

  // PV Status
  if (resourceType === 'pv') {
    switch (status) {
      case 'Available':
        return { className: 'status-available', icon: '🟢' };
      case 'Bound':
        return { className: 'status-bound', icon: '🟢' };
      case 'Released':
        return { className: 'status-released', icon: '🟡' };
      case 'Failed':
        return { className: 'status-failed', icon: '🔴' };
      default:
        return { className: '', icon: '⚪' };
    }
  }

  // Namespace Status
  if (resourceType === 'namespace') {
    switch (status) {
      case 'Active':
        return { className: 'status-active', icon: '🟢' };
      case 'Terminating':
        return { className: 'status-terminating', icon: '🟡' };
      default:
        return { className: '', icon: '⚪' };
    }
  }

  // Event Type
  if (resourceType === 'event') {
    switch (status) {
      case 'Normal':
        return { className: 'status-normal', icon: '🔵' };
      case 'Warning':
        return { className: 'status-warning', icon: '🟡' };
      default:
        return { className: '', icon: '⚪' };
    }
  }

  // Service Type
  if (resourceType === 'service') {
    switch (status) {
      case 'ClusterIP':
        return { className: 'status-clusterip', icon: '🔵' };
      case 'NodePort':
        return { className: 'status-nodeport', icon: '🟢' };
      case 'LoadBalancer':
        return { className: 'status-loadbalancer', icon: '🟣' };
      case 'ExternalName':
        return { className: 'status-external', icon: '🟡' };
      default:
        return { className: '', icon: '⚪' };
    }
  }

  // defaultStatus
  return { className: '', icon: '⚪' };
}

export default StatusBadge;
