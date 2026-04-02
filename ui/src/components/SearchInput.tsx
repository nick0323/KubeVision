import React, { useCallback } from 'react';
import { FaSearch, FaTimes } from 'react-icons/fa';
import { SearchInputProps } from '../types';
import './SearchInput.css';

/**
 * 搜索输入框组件
 */
export const SearchInput: React.FC<SearchInputProps> = ({
  placeholder = 'Search...',
  value,
  onChange,
  onSubmit,
  onClear,
  isSearching = false,
  hasSearchResults = false,
  showSearchButton = false,
  showClearButton = true,
}) => {
  const handleSubmit = useCallback(
    (e: React.FormEvent) => {
      e.preventDefault();
      if (onSubmit) {
        onSubmit(e);
      }
    },
    [onSubmit]
  );

  const handleKeyDown = useCallback(
    (e: React.KeyboardEvent<HTMLInputElement>) => {
      // 只在按 Enter 键时触发搜索
      if (e.key === 'Enter' && onSubmit) {
        e.preventDefault();
        onSubmit(e as any);
      }
    },
    [onSubmit]
  );

  return (
    <div className="search-input-wrapper">
      <form onSubmit={handleSubmit} className="search-input-form">
        <input
          type="text"
          className="search-input"
          placeholder={placeholder}
          value={value}
          onChange={onChange}
          onKeyDown={handleKeyDown}
        />
        {showSearchButton && (
          <button type="submit" className="search-button">
            <FaSearch />
          </button>
        )}
      </form>
      {showClearButton && value && (
        <button className="clear-button" onClick={onClear} title="清除搜索">
          <FaTimes />
        </button>
      )}
      {isSearching && (
        <div className="search-loading">
          <div className="loading-dot" />
        </div>
      )}
      {hasSearchResults && !isSearching && <div className="search-results-indicator" />}
    </div>
  );
};

export default SearchInput;
