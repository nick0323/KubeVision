import React, { useRef, useEffect, useState, memo, useCallback, ReactElement } from 'react';
import Tippy from '@tippyjs/react';
import 'tippy.js/dist/tippy.css';
import 'tippy.js/themes/light.css';

// ==================== TooltipCell 组件 ====================

interface TooltipCellProps {
  content: React.ReactNode;
  text: string;
  maxWidth?: number;
}

/**
 * 带 Tooltip 的单元格组件
 * 只有当内容溢出时才显示 Tooltip
 */
export const TooltipCell: React.FC<TooltipCellProps> = memo(({ content, text, maxWidth = 600 }) => {
  const [showTooltip, setShowTooltip] = useState(false);
  const spanRef = useRef<HTMLSpanElement>(null);

  // 检测内容是否溢出
  useEffect(() => {
    if (spanRef.current) {
      const isOverflowing = spanRef.current.scrollWidth > spanRef.current.clientWidth;
      setShowTooltip(isOverflowing);
    }
  }, [text]);

  const displayContent = Array.isArray(content) ? content.join(', ') : content;

  const cellElement = (
    <span ref={spanRef} className="table-cell-text">
      {displayContent}
    </span>
  );

  if (showTooltip) {
    return (
      <Tippy
        content={text}
        theme="light"
        placement="top"
        arrow={true}
        maxWidth={maxWidth}
        duration={200}
        allowHTML={false}
        interactive={false}
      >
        {cellElement}
      </Tippy>
    );
  }

  return cellElement;
});

TooltipCell.displayName = 'TooltipCell';

// ==================== 表格列定义 ====================

export interface TableColumn<T = any> {
  title: string;
  dataIndex: string;
  width?: number | string;
  sortable?: boolean;
  render?: (value: any, record: T, index: number) => React.ReactNode;
  className?: string;
}

// ==================== TableHeaderCell 组件 ====================

interface TableHeaderCellProps {
  column: TableColumn;
  sortField: string;
  sortOrder: 'asc' | 'desc';
  onSort: (field: string) => void;
}

/**
 * 表头单元格组件
 */
export const TableHeaderCell = memo(
  ({ column, sortField, sortOrder, onSort }: TableHeaderCellProps) => {
    const isSorted = sortField === column.dataIndex;
    const sortIcon = isSorted ? (sortOrder === 'asc' ? ' ↑' : ' ↓') : '';
    
    return (
      <th style={{ width: column.width }} className={isSorted ? 'sorted' : ''}>
        <div
          className={`table-header-cell ${column.sortable ? 'sortable' : ''}`}
          onClick={() => {
            if (column.sortable) {
              onSort(column.dataIndex);
            }
          }}
          role={column.sortable ? 'button' : undefined}
          tabIndex={column.sortable ? 0 : undefined}
          onKeyDown={e => {
            if (column.sortable && (e.key === 'Enter' || e.key === ' ')) {
              e.preventDefault();
              onSort(column.dataIndex);
            }
          }}
        >
          <span>{column.title}</span>
          {column.sortable && <span className="sort-icon">{sortIcon}</span>}
        </div>
      </th>
    );
  }
) as React.FC<TableHeaderCellProps>;

(TableHeaderCell as any).displayName = 'TableHeaderCell';

// ==================== TableCell 组件 ====================

interface TableCellProps {
  value: unknown;
  column: TableColumn;
  record: any;
  index: number;
  isStatusColumn: boolean;
  statusColumnIndex?: string;
  StatusBadge?: React.ComponentType<{ status: string; resourceType?: string }>;
  resourceType?: string;
}

/**
 * 数据单元格组件
 */
export const TableCell = memo(
  ({
    value,
    column,
    record,
    index,
    isStatusColumn,
    statusColumnIndex,
    StatusBadge,
    resourceType,
  }: TableCellProps) => {
  // 状态列使用 StatusBadge
      if (isStatusColumn && column.dataIndex === statusColumnIndex) {
        const statusValue = typeof value === 'string' ? value : String(value ?? '');
        return (
          <td key={column.dataIndex} className={`table-cell ${column.className || ''}`}>
            {StatusBadge && <StatusBadge status={statusValue} resourceType={resourceType} />}
          </td>
        );
      }

  // 使用自定义渲染函数
      if (column.render) {
        const content = column.render(value, record, index);
        // 将渲染结果转换为字符串用于 Tooltip
        const tooltipText = typeof content === 'string' ? content : String(value ?? '');
        return (
          <td key={String(column.dataIndex)} className={`table-cell ${column.className || ''}`}>
            <TooltipCell content={content} text={tooltipText} />
          </td>
        );
      }

  // 数组显示
      if (Array.isArray(value)) {
        const text = (value as unknown[]).join(', ');
        return (
          <td key={String(column.dataIndex)} className={`table-cell ${column.className || ''}`}>
            <TooltipCell content={text} text={text} />
          </td>
        );
      }

  // 对象显示
      if (typeof value === 'object' && value !== null) {
        const text = JSON.stringify(value);
        return (
          <td key={String(column.dataIndex)} className={`table-cell ${column.className || ''}`}>
            <TooltipCell content={text} text={text} />
          </td>
        );
      }

  // 普通文本
      const text = String(value ?? '');
      return (
        <td key={String(column.dataIndex)} className={`table-cell ${column.className || ''}`}>
          <TooltipCell content={text} text={text} />
        </td>
      );
  }
) as React.FC<TableCellProps>;

(TableCell as any).displayName = 'TableCell';

// ==================== TableRow 组件 ====================

interface TableRowProps<T> {
  record: T;
  rowIndex: number;
  columns: TableColumn<T>[];
  isStatusColumn: boolean;
  statusColumnIndex?: string;
  StatusBadge?: React.ComponentType<{ status: string; resourceType?: string }>;
  resourceType?: string;
  onClick?: (record: T) => void;
}

/**
 * 表格行组件
 */
export const TableRow = memo(
  ({
    record,
    rowIndex,
    columns,
    isStatusColumn,
    statusColumnIndex,
    StatusBadge,
    resourceType,
    onClick,
  }: TableRowProps<any>) => {
    const handleClick = useCallback(() => {
      onClick?.(record);
    }, [onClick, record]);

  return (
    <tr
      key={rowIndex}
      className={`table-row ${onClick ? 'clickable' : ''}`}
      onClick={handleClick}
    >
        {columns.map((column, colIndex) => (
          <TableCell
            key={colIndex}
            value={(record as any)[column.dataIndex]}
            column={column}
            record={record}
            index={rowIndex}
            isStatusColumn={isStatusColumn}
            statusColumnIndex={statusColumnIndex}
            StatusBadge={StatusBadge}
            resourceType={resourceType}
          />
        ))}
      </tr>
    );
  }
) as React.FC<TableRowProps<any>>;

(TableRow as any).displayName = 'TableRow';

// ==================== Table Skeleton 组件 ====================

interface TableSkeletonProps {
  columns: number;
  rows?: number;
}

/**
 * 表格骨架屏
 */
export const TableSkeleton: React.FC<TableSkeletonProps> = memo(({ columns, rows = 10 }) => {
  return (
    <>
      {Array.from({ length: rows }).map((_, rowIndex) => (
        <tr key={rowIndex} className="table-row skeleton-row">
          {Array.from({ length: columns }).map((_, colIndex) => (
            <td key={colIndex} className="table-cell">
              <div className="skeleton-placeholder" />
            </td>
          ))}
        </tr>
      ))}
    </>
  );
});

TableSkeleton.displayName = 'TableSkeleton';

// ==================== Empty State 组件 ====================

interface EmptyStateProps {
  message?: string;
  description?: string;
  icon?: React.ReactNode;
  colSpan: number;
  action?: React.ReactNode;
}

/**
 * 空状态组件
 */
export const EmptyState: React.FC<EmptyStateProps> = memo(
  ({ message = '暂无数据', description, icon = '📭', colSpan, action }) => {
    return (
      <tr className="empty-row">
        <td colSpan={colSpan}>
          <div className="empty-state">
            <span className="empty-icon">{icon}</span>
            <p className="empty-message">{message}</p>
            {description && <p className="empty-description">{description}</p>}
            {action && <div className="empty-action">{action}</div>}
          </div>
        </td>
      </tr>
    );
  }
);

EmptyState.displayName = 'EmptyState';

// ==================== Error State 组件 ====================

interface ErrorStateProps {
  message: string;
  colSpan: number;
  onRetry?: () => void;
  retryText?: string;
}

/**
 * 错误状态组件
 */
export const ErrorState: React.FC<ErrorStateProps> = memo(
  ({ message, colSpan, onRetry, retryText = '重试' }) => {
    return (
      <tr className="error-row">
        <td colSpan={colSpan}>
          <div className="error-state">
            <span className="error-icon">⚠️</span>
            <p className="error-message">{message}</p>
            {onRetry && (
              <button onClick={onRetry} className="retry-btn">
                {retryText}
              </button>
            )}
          </div>
        </td>
      </tr>
    );
  }
);

ErrorState.displayName = 'ErrorState';
