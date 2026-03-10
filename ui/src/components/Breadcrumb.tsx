/**
 * 面包屑导航组件
 * 显示资源层级关系
 */
import React from 'react';
import { useNavigate } from 'react-router-dom';
import './Breadcrumb.css';

interface BreadcrumbItem {
  label: string;
  href?: string;
  type?: string;
}

interface BreadcrumbProps {
  items: BreadcrumbItem[];
}

export const Breadcrumb: React.FC<BreadcrumbProps> = ({ items }) => {
  const navigate = useNavigate();

  const handleClick = (item: BreadcrumbItem) => {
    if (item.href) {
      navigate(item.href);
    }
  };

  return (
    <div className="breadcrumb">
      <div className="breadcrumb-list">
        {items.map((item, index) => (
          <React.Fragment key={index}>
            {index > 0 && <span className="breadcrumb-separator">/</span>}
            <span
              className={`breadcrumb-item ${item.href ? 'clickable' : ''}`}
              onClick={() => item.href && handleClick(item)}
            >
              {item.type && <span className="breadcrumb-type">{item.type}</span>}
              <span className="breadcrumb-label">{item.label}</span>
            </span>
          </React.Fragment>
        ))}
      </div>
    </div>
  );
};

export default Breadcrumb;
