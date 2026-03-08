import React, { useRef, useEffect, useState } from 'react';
import Tippy from '@tippyjs/react';
import 'tippy.js/dist/tippy.css';
import 'tippy.js/themes/light.css';
import { Column, CommonTableProps } from '../types';
import './CommonTable.css';

// 单元格组件（判断是否显示省略号）
const TableCell: React.FC<{
  content: React.ReactNode;
  text: string;
  maxWidth: number;
}> = ({ content, text, maxWidth }) => {
  const spanRef = useRef<HTMLSpanElement>(null);
  const [showTooltip, setShowTooltip] = useState(false);

  useEffect(() => {
    // 判断是否显示省略号（scrollWidth > clientWidth 表示内容溢出）
    if (spanRef.current) {
      setShowTooltip(spanRef.current.scrollWidth > spanRef.current.clientWidth);
    }
  }, [text]);

  const cellElement = (
    <span
      ref={spanRef}
      style={{
        display: 'block',
        maxWidth: `${maxWidth}px`,
        overflow: 'hidden',
        textOverflow: 'ellipsis',
        whiteSpace: 'nowrap',
        cursor: showTooltip ? 'pointer' : 'default'
      }}
    >
      {content}
    </span>
  );

  // 只有长内容才显示 Tooltip
  if (showTooltip) {
    return (
      <Tippy
        content={text}
        theme="light"
        placement="top"
        arrow={true}
        maxWidth={600} // 增加最大宽度
        duration={200}
        allowHTML={false}
        interactive={false}
      >
        {cellElement}
      </Tippy>
    );
  }

  return cellElement;
};

/**
 * 通用表格组件
 */
export const CommonTable: React.FC<CommonTableProps<any>> = ({
  columns = [],
  data = [],
  pageSize = 20,
  currentPage = 1,
  total = 0,
  onPageChange,
  emptyText = '暂无数据',
  className = '',
  hasFixedPagination = false
}) => {
  const totalPages = Math.ceil(total / pageSize);

  // 根据是否有固定分页来决定表格卡片的类名
  const tableCardClass = hasFixedPagination ? 'table-card has-fixed-pagination' : 'table-card';

  // 根据数据量决定表格内容区域的类名 - 数据少时使用 compact 模式
  const isCompact = data.length <= 3; // 3 行或更少时使用紧凑模式
  const tableContentClass = isCompact ? 'table-content-area compact' : 'table-content-area';

  return (
    <div className={tableCardClass}>
      <div className={tableContentClass}>
        <table className={`table ${className}`}>
          <thead>
            <tr>
              {columns.map((col: Column<any>) => (
                <th key={col.key || col.dataIndex}>
                  {col.label || col.title}
                </th>
              ))}
            </tr>
          </thead>
          <tbody>
            {data.length === 0 ? (
              <tr>
                <td colSpan={columns.length} style={{ textAlign: 'center', color: '#c0c4cc', fontSize: 'var(--font-size-sm)' }}>
                  {emptyText}
                </td>
              </tr>
            ) : (
              data.map((row, i) => (
                <tr key={i}>
                  {columns.map((col: Column<any>) => {
                    // 渲染内容 - 优先使用 render 函数，否则直接访问字段
                    let cellContent: React.ReactNode;
                    
                    if (col.render) {
                      // 使用自定义渲染函数
                      cellContent = col.render(row[col.dataIndex], row, i, false);
                    } else {
                      // 直接访问字段，处理 undefined/null 情况
                      const value = row[col.dataIndex];
                      cellContent = (value === undefined || value === null) ? '' : value;
                    }

                    // 如果内容是数组，转换为逗号分隔的字符串
                    if (Array.isArray(cellContent)) {
                      cellContent = cellContent.join(', ');
                    }

                    // 如果内容仍为空，显示'-'
                    const displayContent = cellContent === '' ? '-' : cellContent;
                    const cellContentStr = String(displayContent);

                    // 使用 TableCell 组件（自动判断是否显示 Tooltip）
                    return (
                      <td
                        key={col.key || col.dataIndex}
                        data-type={col.dataIndex}
                      >
                        <TableCell
                          content={displayContent}
                          text={cellContentStr}
                          maxWidth={250}
                        />
                      </td>
                    );
                  })}
                </tr>
              ))
            )}
          </tbody>
        </table>
      </div>
      {/* 分页区域已移除，由外部页面负责渲染 */}
    </div>
  );
};

export default CommonTable;
