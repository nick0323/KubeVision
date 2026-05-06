import React, { useEffect } from 'react';
import { PaginationProps } from '../types';
import { PAGE_SIZE_OPTIONS } from '../constants';
import './Pagination.css';

/**
 * PaginationComponent
 * Keep with Pagination.jsx exactly the same functionality
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
  const totalPages = Math.ceil(total / pageSize);

  // 键盘快捷键Support（仅in固定Modeunder启use）
  useEffect(() => {
    // onlyin固定Modeunder启use快捷键
    if (!fixedBottom && !fixed) return;

    const handleKeyDown = (e: KeyboardEvent) => {
      // onlyin没has聚焦Input field时启use快捷键
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

  // if总数小forevery页数量，notDisplayPagination
  if (total <= pageSize) return null;

  // Build CSS 类名
  let className = 'table-pagination-area';
  if (fixedBottom) {
    className += ' fixed-bottom';
  } else if (fixed) {
    className += ' fixed';
  }

  return (
    <div className={className}>
      {/* Left: Total rows and rows per page selector */}
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

      {/* Right: Page info and navigation buttons */}
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
