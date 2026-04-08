import React from 'react';
import { ResourceSummaryProps } from './types';
import './ResourceSummary.css';

/**
 * 资源概览组件
 */
export const ResourceSummary: React.FC<ResourceSummaryProps> = ({
  title = '',
  requestsValue = 0,
  limitsValue = 0,
  totalValue = 0,
  availableValue = 0,
  requestsPercent = 0,
  limitsPercent = 0,
  unit = '',
}) => {
  return (
    <div className="resource-summary-card">
      <div className="resource-summary-title">{title}</div>
      <div className="resource-summary-info">
        Requests: <span className="summary-num">{requestsValue}</span> / Limits:{' '}
        <span className="summary-num">{limitsValue}</span> / Total:{' '}
        <span className="summary-num">{totalValue}</span> {unit}
      </div>
      <div className="resource-summary-row">
        <div className="resource-block">
          <div className="resource-header">
            <span className="label requests">Requests</span>
            <span className="value">
              {requestsValue} {unit}
            </span>
          </div>
          <div className="progress-bar requests">
            <div className="progress" style={{ width: `${requestsPercent}%` }} />
          </div>
          <div className="percent">{requestsPercent}% of capacity</div>
        </div>
        <div className="resource-block">
          <div className="resource-header">
            <span className="label limits">Limits</span>
            <span className="value">
              {limitsValue} {unit}
            </span>
          </div>
          <div className="progress-bar limits">
            <div className="progress" style={{ width: `${limitsPercent}%` }} />
          </div>
          <div className="percent">{limitsPercent}% of capacity</div>
        </div>
      </div>
      <div className="resource-summary-available">
        Available: {availableValue} {unit}
      </div>
    </div>
  );
};

export default ResourceSummary;
