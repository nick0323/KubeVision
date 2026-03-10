import React from 'react';
import './LabelList.css';

interface LabelListProps {
  labels: Record<string, string>;
  resourceType?: string;      // 当前资源类型
  namespace?: string;         // 当前命名空间
  onLabelClick?: (key: string, value: string) => void;  // 点击回调
}

export const LabelList: React.FC<LabelListProps> = ({ 
  labels, 
  resourceType,
  namespace,
  onLabelClick 
}) => {
  // 处理标签点击
  const handleClick = (key: string, value: string) => {
    // 如果是资源关联标签，触发跳转
    if (onLabelClick) {
      onLabelClick(key, value);
    }
  };

  // 判断是否为可点击的资源关联标签
  const isClickableLabel = (key: string, value: string) => {
    // 节点标签
    if (key === 'kubernetes.io/hostname') {
      return true;
    }
    return false;
  };

  return (
    <div className="label-list">
      {Object.entries(labels).map(([key, value]) => {
        const clickable = isClickableLabel(key, value);
        return (
          <div 
            key={key} 
            className={`label-item ${clickable ? 'clickable' : ''}`}
            onClick={() => clickable && handleClick(key, value)}
          >
            <span className="label-key">{key}</span>
            <span className="label-sep">=</span>
            <span className="label-value">{value}</span>
          </div>
        );
      })}
    </div>
  );
};

export default LabelList;
