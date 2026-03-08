import React, { useState, useEffect, useCallback } from 'react';
import './SearchFilter.css';

interface SearchFilterProps {
  value: string;
  onSearch: (value: string) => void;
  placeholder?: string;
  debounceMs?: number;
}

/**
 * 搜索过滤组件
 * 支持防抖和清空功能
 */
export const SearchFilter: React.FC<SearchFilterProps> = ({
  value,
  onSearch,
  placeholder = '搜索...',
  debounceMs = 300,
}) => {
  const [inputValue, setInputValue] = useState(value);

  // 防抖处理
  useEffect(() => {
    const timer = setTimeout(() => {
      if (inputValue !== value) {
        onSearch(inputValue);
      }
    }, debounceMs);

    return () => clearTimeout(timer);
  }, [inputValue, value, onSearch, debounceMs]);

  // 同步外部值变化
  useEffect(() => {
    setInputValue(value);
  }, [value]);

  const handleClear = useCallback(() => {
    setInputValue('');
    onSearch('');
  }, [onSearch]);

  return (
    <div className="search-filter">
      <div className="search-input-wrapper">
        <span className="search-icon">🔍</span>
        <input
          type="text"
          className="search-input"
          value={inputValue}
          onChange={(e) => setInputValue(e.target.value)}
          placeholder={placeholder}
          autoComplete="off"
        />
        {inputValue && (
          <button className="clear-btn" onClick={handleClear} type="button">
            ✕
          </button>
        )}
      </div>
    </div>
  );
};

export default SearchFilter;
