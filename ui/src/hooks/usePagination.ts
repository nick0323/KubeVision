import { useState, useCallback } from 'react';
import { PAGE_SIZE, PAGE_SIZE_OPTIONS } from '../constants';

/**
 * 自定义分页 Hook
 */
export function usePagination(
  initialPageSize: number = PAGE_SIZE,
  pageSizeOptions: number[] = PAGE_SIZE_OPTIONS
) {
  const [page, setPage] = useState<number>(1);
  const [pageSize, setPageSize] = useState<number>(initialPageSize);

  // 处理页码变化
  const handlePageChange = useCallback((newPage: number) => {
    setPage(newPage);
  }, []);

  // 处理每页行数变化
  const handlePageSizeChange = useCallback((newPageSize: number) => {
    setPageSize(newPageSize);
    setPage(1); // 重置到第一页
  }, []);

  // 重置分页状态
  const resetPagination = useCallback(() => {
    setPage(1);
  }, []);

  // 计算总页数
  const getTotalPages = useCallback((total: number) => {
    return Math.ceil(total / pageSize);
  }, [pageSize]);

  // 计算偏移量
  const getOffset = useCallback(() => {
    return (page - 1) * pageSize;
  }, [page, pageSize]);

  return {
    // 状态
    page,
    pageSize,

    // 操作方法
    setPage,
    setPageSize,
    handlePageChange,
    handlePageSizeChange,
    resetPagination,

    // 计算属性
    getTotalPages,
    getOffset,

    // 配置
    pageSizeOptions
  };
}
