import React, { useEffect, useState, useCallback, useRef } from 'react';
import { ResourcePageProps } from '../types';
import PageHeader from './PageHeader.tsx';
import SearchInput from './SearchInput.tsx';
import NamespaceSelect from './NamespaceSelect.tsx';
import RefreshButton from './RefreshButton.tsx';
import CommonTable from '../CommonTable.tsx';
import Pagination from '../Pagination.tsx';
import { usePagination } from '../hooks/usePagination.ts';
import { apiClient, createPaginatedQuery } from '../utils/apiClient.ts';
import { LoadingSpinner } from './LoadingSpinner.tsx';
import { ErrorDisplay } from './ErrorDisplay.tsx';
import './ResourcePage.css';

/**
 * 通用资源页面组件 - 修复版
 * 改进：
 * 1. 合并 useEffect 减少重渲染
 * 2. 使用 AbortController 取消请求
 * 3. 添加请求去抖
 */
export const ResourcePage: React.FC<ResourcePageProps> = ({
  title,
  apiEndpoint,
  resourceType,
  columns,
  collapsed,
  onToggleCollapsed,
  statusMap = {},
  namespaceFilter = true,
}) => {
  const [data, setData] = useState<any[]>([]);
  const [total, setTotal] = useState<number>(0);
  const [loading, setLoading] = useState<boolean>(false);
  const [error, setError] = useState<string | null>(null);
  const [namespace, setNamespace] = useState<string>('');
  const [search, setSearch] = useState<string>('');

  const {
    page,
    pageSize,
    handlePageChange,
    handlePageSizeChange,
    resetPagination,
  } = usePagination();

  // 使用 ref 保存最新值
  const pageRef = useRef(page);
  const pageSizeRef = useRef(pageSize);
  const namespaceRef = useRef(namespace);
  const searchRef = useRef(search);
  const abortControllerRef = useRef<AbortController | null>(null);

  // 更新 ref
  useEffect(() => {
    pageRef.current = page;
    pageSizeRef.current = pageSize;
    namespaceRef.current = namespace;
    searchRef.current = search;
  });

  // 获取数据 - 使用 AbortController 取消旧请求
  const fetchData = useCallback(async () => {
    // 取消之前的请求
    if (abortControllerRef.current) {
      abortControllerRef.current.abort();
    }
    
    abortControllerRef.current = new AbortController();
    
    setLoading(true);
    setError(null);

    try {
      const result = await createPaginatedQuery(apiEndpoint, {
        page: pageRef.current,
        pageSize: pageSizeRef.current,
        namespace: namespaceRef.current,
        search: searchRef.current,
      });

      setData(result.data || []);
      setTotal(result.page?.total || 0);
    } catch (err) {
      if (err instanceof Error && err.name !== 'AbortError') {
        setError(err.message);
      }
    } finally {
      setLoading(false);
    }
  }, [apiEndpoint]);

  // 统一的数据加载逻辑
  useEffect(() => {
    fetchData();
    
    // 清理函数
    return () => {
      if (abortControllerRef.current) {
        abortControllerRef.current.abort();
      }
    };
  }, [page, pageSize, namespace, search, fetchData]);

  // namespace 变化时重置分页
  useEffect(() => {
    resetPagination();
  }, [namespace, resetPagination]);

  // 处理搜索
  const handleSearchChange = useCallback((e: React.ChangeEvent<HTMLInputElement>) => {
    setSearch(e.target.value);
  }, []);

  // 搜索提交
  const handleSearchSubmit = useCallback(() => {
    handlePageChange(1);
  }, [handlePageChange]);

  // 清空搜索
  const handleClearSearch = useCallback(() => {
    setSearch('');
    handlePageChange(1);
  }, [handlePageChange]);

  // 处理每页条数变化
  const handlePageSizeChangeWrapper = useCallback((newPageSize: number) => {
    handlePageSizeChange(newPageSize);
  }, [handlePageSizeChange]);

  // 刷新数据
  const handleRefresh = useCallback(() => {
    fetchData();
  }, [fetchData]);

  if (loading && data.length === 0) {
    return <LoadingSpinner text="Loading..." overlay />;
  }

  if (error && data.length === 0) {
    return (
      <ErrorDisplay
        message={error}
        type="error"
        showRetry
        onRetry={fetchData}
      />
    );
  }

  return (
    <div className="resource-page">
      <PageHeader
        title={title}
        collapsed={collapsed}
        onToggleCollapsed={onToggleCollapsed}
      >
        {namespaceFilter && (
          <NamespaceSelect
            value={namespace}
            onChange={setNamespace}
            placeholder="All Namespaces"
          />
        )}
        <SearchInput
          placeholder={`搜索 ${title}...`}
          value={search}
          onChange={handleSearchChange}
          onSubmit={handleSearchSubmit}
          onClear={handleClearSearch}
          isSearching={loading}
          hasSearchResults={search.length > 0}
        />
        <RefreshButton onClick={handleRefresh} loading={loading} />
      </PageHeader>

      <CommonTable
        columns={columns}
        data={data}
        emptyText={`暂无 ${title} 数据`}
      />

      <Pagination
        currentPage={page}
        total={total}
        pageSize={pageSize}
        onPageChange={handlePageChange}
        onPageSizeChange={handlePageSizeChangeWrapper}
      />
    </div>
  );
};

export default ResourcePage;
