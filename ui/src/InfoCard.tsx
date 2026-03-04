import React from 'react';
import { InfoCardProps } from '../types';
import './App.css';

/**
 * 信息卡片组件
 * 保持与 InfoCard.jsx 完全一致的功能
 */
export const InfoCard: React.FC<InfoCardProps> = ({ icon, title, value, status, children }) => {
  return (
    <div className="overview-card">
      {icon && (
        <div className="overview-card-icon-col">
          <span className="overview-icon">{icon}</span>
        </div>
      )}
      <div className="overview-card-content-col" style={{minHeight: 110, display: 'flex', flexDirection: 'column', justifyContent: 'center'}}>
        <div className="overview-title">{title}</div>
        {value !== undefined && (
          <div className="overview-value">{value}</div>
        )}
        {status && (
          <div className="overview-status">{status}</div>
        )}
        {children}
      </div>
    </div>
  );
};

export default InfoCard;
