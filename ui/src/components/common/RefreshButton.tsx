import React from 'react';
import { FaSync } from 'react-icons/fa';
import { RefreshButtonProps } from '../types';
import './RefreshButton.css';

/**
 * 刷新按钮组件
 */
export const RefreshButton: React.FC<RefreshButtonProps> = ({
  onClick,
  loading = false,
  title = '刷新',
}) => {
  return (
    <button className="refresh-button" onClick={onClick} title={title} disabled={loading}>
      <FaSync className={loading ? 'spinning' : ''} />
    </button>
  );
};

export default RefreshButton;
