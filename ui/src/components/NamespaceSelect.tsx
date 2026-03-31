import React, { useState, useRef, useEffect } from 'react';
import { FaChevronDown, FaChevronUp } from 'react-icons/fa';
import './NamespaceSelect.css';

interface NamespaceSelectProps {
  value: string;
  onChange: (value: string) => void;
  placeholder?: string;
  disabled?: boolean;
  options?: string[]; // 父组件传入的命名空间列表
  className?: string; // 自定义 className
  width?: string; // 自定义宽度
}

/**
 * 自定义 Namespace 选择器组件 - 支持完全样式控制
 */
export const NamespaceSelect: React.FC<NamespaceSelectProps> = ({
  value,
  onChange,
  placeholder = '选择命名空间',
  disabled = false,
  options = [],
  className = '',
  width,
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

  const handleSelect = (ns: string) => {
    onChange(ns);
    setIsOpen(false);
  };

  const toggleDropdown = () => {
    if (!disabled) {
      setIsOpen(!isOpen);
    }
  };

  const selectedLabel = value || placeholder;

  return (
    <div className={`namespace-select-custom ${className || ''}`} ref={containerRef} style={width ? { width, minWidth: width } : {}}>
      {/* 选择框主体 */}
      <div
        className={`namespace-select-value ${isOpen ? 'open' : ''} ${disabled ? 'disabled' : ''}`}
        onClick={toggleDropdown}
      >
        <span className="namespace-select-text">{selectedLabel}</span>
        <span className="namespace-select-arrow">
          {isOpen ? <FaChevronUp /> : <FaChevronDown />}
        </span>
      </div>

      {/* 下拉选项 */}
      {isOpen && (
        <div className="namespace-select-dropdown">
          {/* 全部选项 - 只在 placeholder 有值时显示 */}
          {placeholder && (
            <div
              className={`namespace-select-option ${!value ? 'selected' : ''}`}
              onClick={() => handleSelect('')}
            >
              {placeholder}
            </div>
          )}

          {/* 命名空间列表 */}
          {options.length === 0 ? (
            <div className="namespace-select-option disabled">
              暂无数据
            </div>
          ) : (
            options.map((ns) => (
              <div
                key={ns}
                className={`namespace-select-option ${value === ns ? 'selected' : ''}`}
                onClick={() => handleSelect(ns)}
              >
                {ns}
              </div>
            ))
          )}
        </div>
      )}
    </div>
  );
};

export default NamespaceSelect;
