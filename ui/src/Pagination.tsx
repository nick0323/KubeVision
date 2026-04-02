import React, { useEffect } from 'react';
import { PaginationProps } from './types';
import { PAGE_SIZE_OPTIONS } from './constants';
import './Pagination.css';

/**
 * 分页组件
 * 保持与 Pagination.jsx 完全一致的功能
 */
export const Pagination: React.FC<PaginationProps> = ({
  currentPage,
  total,
  pageSize,
  onPageChange,
  onPageSizeChange,
  pageSizeOptions = PAGE_SIZE_OPTIONS,
  fixed = false,
  fixedBottom = false,
}) => {
  if (total <= pageSize) return null;
  const totalPages = Math.ceil(total / pageSize);

  // 键盘快捷键支持（仅在固定模式下启用）
  useEffect(() => {
    if (!fixedBottom && !fixed) return;

    const handleKeyDown = (e: KeyboardEvent) => {
      // 只在没有聚焦输入框时启用快捷键
      const target = e.target as HTMLElement;
      if (target.tagName === 'INPUT' || target.tagName === 'TEXTAREA') return;

      if (e.key === 'ArrowLeft' && currentPage > 1) {
        e.preventDefault();
        onPageChange(currentPage - 1);
      } else if (e.key === 'ArrowRight' && currentPage < totalPages) {
        e.preventDefault();
        onPageChange(currentPage + 1);
      }
    };

    document.addEventListener('keydown', handleKeyDown);
    return () => document.removeEventListener('keydown', handleKeyDown);
  }, [fixedBottom, fixed, currentPage, totalPages, onPageChange]);

  // 构建 CSS 类名
  let className = 'table-pagination-area';
  if (fixedBottom) {
    className += ' fixed-bottom';
  } else if (fixed) {
    className += ' fixed';
  }

  return (
    <div className={className}>
      {/* 左侧：总行数和每页行数选择器 */}
      <div className="pagination-total">
        <span>{total} row(s) total</span>
        <span className="pagination-separator">|</span>
        <span className="page-size-selector">
          <span>Show </span>
          <select
            value={pageSize}
            onChange={e => {
              const newPageSize = Number(e.target.value);
              if (onPageSizeChange) {
                onPageSizeChange(newPageSize);
              }
            }}
            className="page-size-select"
          >
            {pageSizeOptions.map(size => (
              <option key={size} value={size}>
                {size}
              </option>
            ))}
          </select>
          <span> per page</span>
        </span>
      </div>

      {/* 右侧：页码信息和导航按钮 */}
      <div className="pagination-controls">
        <span className="pagination-info">
          Page {currentPage} of {totalPages}
        </span>
        <button
          className="pagination-btn"
          onClick={() => currentPage > 1 && onPageChange(currentPage - 1)}
          disabled={currentPage === 1}
        >
          ←
        </button>
        <button
          className="pagination-btn"
          onClick={() => currentPage < totalPages && onPageChange(currentPage + 1)}
          disabled={currentPage === totalPages}
        >
          →
        </button>
      </div>
    </div>
  );
};

export default Pagination;
