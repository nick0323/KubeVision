import React, { useState, useRef, useEffect } from 'react';
import { FaChevronDown, FaChevronUp } from 'react-icons/fa';
import './StatusFilter.css';

interface StatusFilterProps {
  statuses: string[];
  value: string;
  onChange: (value: string) => void;
}

/**
 * 自定义状态选择器组件 - 支持完全样式控制
 */
export const StatusFilter: React.FC<StatusFilterProps> = ({
  statuses,
  value,
  onChange,
}) => {
  const [isOpen, setIsOpen] = useState(false);
  const containerRef = useRef<HTMLDivElement>(null);

  // 点击外部关闭下拉
  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (containerRef.current && !containerRef.current.contains(event.target as Node)) {
        setIsOpen(false);
      }
    };

    document.addEventListener('mousedown', handleClickOutside);
    return () => document.removeEventListener('mousedown', handleClickOutside);
  }, []);

  const handleSelect = (status: string) => {
    onChange(status);
    setIsOpen(false);
  };

  const toggleDropdown = () => {
    setIsOpen(!isOpen);
  };

  const selectedLabel = value || '全部状态';

  return (
    <div className="status-filter-custom" ref={containerRef}>
      {/* 选择框主体 */}
      <div
        className={`status-filter-value ${isOpen ? 'open' : ''}`}
        onClick={toggleDropdown}
      >
        <span className="status-filter-text">{selectedLabel}</span>
        <span className="status-filter-arrow">
          {isOpen ? <FaChevronUp /> : <FaChevronDown />}
        </span>
      </div>

      {/* 下拉选项 */}
      {isOpen && (
        <div className="status-filter-dropdown">
          {/* 全部选项 */}
          <div
            className={`status-filter-option ${!value ? 'selected' : ''}`}
            onClick={() => handleSelect('')}
          >
            全部状态
          </div>
          
          {/* 状态列表 */}
          {statuses.length === 0 ? (
            <div className="status-filter-option disabled">
              暂无数据
            </div>
          ) : (
            statuses.map((status) => (
              <div
                key={status}
                className={`status-filter-option ${value === status ? 'selected' : ''}`}
                onClick={() => handleSelect(status)}
              >
                {status}
              </div>
            ))
          )}
        </div>
      )}
    </div>
  );
};

export default StatusFilter;
