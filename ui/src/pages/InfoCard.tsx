import React from 'react';
import { InfoCardProps } from '../types';
import './InfoCard.css';

export const InfoCard: React.FC<InfoCardProps> = ({ icon, title, value, status, children }) => {
  return (
    <div className="overview-card">
      {icon && (
        <div className="overview-card-icon-col">
          <span className="overview-icon">{icon}</span>
        </div>
      )}
      <div className="overview-card-content-col">
        <div className="overview-title">{title}</div>
        {value !== undefined && <div className="overview-value">{value}</div>}
        {status && <div className="overview-status">{status}</div>}
        {children}
      </div>
    </div>
  );
};

export default InfoCard;
