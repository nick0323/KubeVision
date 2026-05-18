import React from 'react';
import { FaSync } from 'react-icons/fa';
import { RefreshButtonProps } from '../types';
import './RefreshButton.css';

/**
 * Refresh buttonComponent
 */
const RefreshButton: React.FC<RefreshButtonProps> = ({
  onClick,
  loading = false,
  title = 'Refresh',
}) => {
  return (
    <button className="refresh-button" onClick={onClick} title={title} disabled={loading}>
      <FaSync className={loading ? 'spinning' : ''} />
    </button>
  );
};

export default React.memo(RefreshButton);
