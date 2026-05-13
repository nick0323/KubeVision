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
  onNavigate?: (path: string) => void;
}

export const Breadcrumb: React.FC<BreadcrumbProps> = ({
  items = [],
  onNavigate,
}) => {

  return (
    <ol className="breadcrumb-list">
      {items.map((item, index) => (
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
          {index < items.length - 1 && (
            <FaChevronRight className="breadcrumb-separator" />
          )}
        </li>
      ))}
    </ol>
  );
};

export default Breadcrumb;
