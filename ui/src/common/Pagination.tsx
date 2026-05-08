import React, { useEffect, useState, useRef } from 'react';
import { FaChevronDown, FaChevronUp } from 'react-icons/fa';
import { PaginationProps } from '../types';
import { PAGE_SIZE_OPTIONS } from '../constants';
import './Pagination.css';

/**
 * Pagination Component
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
  const [isOpen, setIsOpen] = useState(false);
  const containerRef = useRef<HTMLDivElement>(null);

  // 键盘快捷键 Support
  useEffect(() => {
    if (!fixedBottom && !fixed) return;

    const handleKeyDown = (e: KeyboardEvent) => {
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

  // Click outside 关闭 dropdown
  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (containerRef.current && !containerRef.current.contains(event.target as Node)) {
        setIsOpen(false);
      }
    };

    document.addEventListener('mousedown', handleClickOutside);
    return () => document.removeEventListener('mousedown', handleClickOutside);
  }, []);

  if (total <= pageSize) return null;

  let className = 'table-pagination-area';
  if (fixedBottom) {
    className += ' fixed-bottom';
  } else if (fixed) {
    className += ' fixed';
  }

  const handlePageSizeSelect = (newPageSize: number) => {
    if (onPageSizeChange) {
      onPageSizeChange(newPageSize);
    }
    setIsOpen(false);
  };

  return (
    <div className={className}>
      <div className="pagination-total">
        <span>{total} row(s) total</span>
        <span className="pagination-separator">|</span>
        <span className="page-size-selector">
          <span>Show </span>
          <div className="page-size-select-wrapper" ref={containerRef}>
            <div
              className={`page-size-select-value ${isOpen ? 'open' : ''}`}
              onClick={() => setIsOpen(!isOpen)}
            >
              <span className="page-size-select-text">{pageSize}</span>
              <span className="page-size-select-arrow">
                {isOpen ? <FaChevronUp /> : <FaChevronDown />}
              </span>
            </div>
            {isOpen && (
              <div className="page-size-select-dropdown">
                {pageSizeOptions.map(size => (
                  <div
                    key={size}
                    className={`page-size-option ${pageSize === size ? 'selected' : ''}`}
                    onClick={() => handlePageSizeSelect(size)}
                  >
                    {size}
                  </div>
                ))}
              </div>
            )}
          </div>
          <span> per page</span>
        </span>
      </div>

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
