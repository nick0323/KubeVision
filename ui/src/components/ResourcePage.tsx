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
 * 通用资源页面组件
 * 支持所有 K8s 资源类型的列表展示
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
  const [loading, setLoading] = useState<boolean>(true);
  const [error, setError] = useState<string | null>(null);
  const [namespace, setNamespace] = useState<string>('');
  const [search, setSearch] = useState<string>('');

  const {
    page,
    pageSize,
    handlePageChange,
    handlePageSizeChange: handlePageSizeChangeFromHook,
    resetPagination,
  } = usePagination();

  // 使用 ref 保存状态，避免闭包问题
  const pageRef = useRef(page);
  const pageSizeRef = useRef(pageSize);
  const namespaceRef = useRef(namespace);
  const searchRef = useRef(search);

  // 更新 ref
  useEffect(() => {
    pageRef.current = page;
  }, [page]);

  useEffect(() => {
    pageSizeRef.current = pageSize;
  }, [pageSize]);

  useEffect(() => {
    namespaceRef.current = namespace;
  }, [namespace]);

  useEffect(() => {
    searchRef.current = search;
  }, [search]);

  // 获取数据 - 不依赖状态，从 ref 读取最新值
  const fetchData = useCallback(async () => {
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
      setError(err instanceof Error ? err.message : '加载失败');
    } finally {
      setLoading(false);
    }
  }, [apiEndpoint]);

  // 初始加载
  useEffect(() => {
    fetchData();
  }, [fetchData]);

  // namespace 变化时重置
  useEffect(() => {
    resetPagination();
  }, [namespace, resetPagination]);

  // namespace 变化后刷新数据
  useEffect(() => {
    fetchData();
  }, [namespace]);

  // page 变化时刷新数据
  useEffect(() => {
    fetchData();
  }, [page]);

  // pageSize 变化后刷新数据
  useEffect(() => {
    fetchData();
  }, [pageSize]);

  // 处理搜索
  const handleSearchChange = (e: React.ChangeEvent<HTMLInputElement>) => {
    setSearch(e.target.value);
  };

  // 搜索提交（回车或点击搜索按钮）
  const handleSearchSubmit = useCallback(() => {
    handlePageChange(1);
    fetchData();
  }, [handlePageChange, fetchData]);

  // 清空搜索
  const handleClearSearch = useCallback(() => {
    setSearch('');
    handlePageChange(1);
    fetchData();
  }, [handlePageChange, fetchData]);

  // 处理每页条数变化
  const handlePageSizeChange = useCallback((newPageSize: number) => {
    handlePageSizeChangeFromHook(newPageSize);
  }, [handlePageSizeChangeFromHook]);

  // pageSize 变化后刷新数据
  useEffect(() => {
    fetchData();
  }, [pageSize]);

  if (loading) {
    return <LoadingSpinner text="Loading..." overlay />;
  }

  if (error) {
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
        <RefreshButton onClick={fetchData} loading={loading} />
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
        onPageSizeChange={handlePageSizeChange}
      />
    </div>
  );
};

export default ResourcePage;
