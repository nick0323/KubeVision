import React, { useEffect, useState, useCallback, useRef, useMemo } from 'react';
import {
  K8sResource,
  APIResponse,
  ListQueryParams,
  ColumnDef,
  PaginatedResponse,
} from '../types/k8s-resources';
import PageHeader from './PageHeader.tsx';
import SearchInput from './SearchInput.tsx';
import NamespaceSelect from './NamespaceSelect.tsx';
import RefreshButton from './RefreshButton.tsx';
import CommonTable from '../CommonTable.tsx';
import Pagination from '../Pagination.tsx';
import { usePagination } from '../hooks/usePagination.ts';
import { LoadingSpinner } from './LoadingSpinner.tsx';
import { ErrorDisplay } from './ErrorDisplay.tsx';
import './ResourcePage.css';

// ==================== 组件 Props 类型 ====================

/**
 * ResourcePage 组件属性（泛型版本）
 * @template T - K8s 资源类型，必须 extends K8sResource
 */
export interface ResourcePageProps<T extends K8sResource> {
  /** 页面标题 */
  title: string;
  /** API 端点 */
  apiEndpoint: string;
  /** 资源类型（用于日志和错误提示） */
  resourceType: string;
  /** 列定义 */
  columns: ColumnDef<T>[];
  /** 侧边栏是否折叠 */
  collapsed: boolean;
  /** 切换侧边栏折叠状态 */
  onToggleCollapsed: () => void;
  /** 状态映射（可选） */
  statusMap?: Record<string, {
    color: 'success' | 'error' | 'warning' | 'default' | 'processing';
    text?: string;
  }>;
  /** 是否启用命名空间过滤 */
  namespaceFilter?: boolean;
  /** 默认命名空间（可选） */
  defaultNamespace?: string;
  /** 行点击事件（可选） */
  onRowClick?: (record: T) => void;
  /** 自定义渲染函数（可选） */
  customRenderers?: Record<string, (value: any, record: T) => React.ReactNode>;
}

/**
 * 查询参数配置（内部使用）
 */
interface QueryConfig {
  page: number;
  pageSize: number;
  namespace?: string;
  search?: string;
  sortBy?: string;
  sortOrder?: 'asc' | 'desc';
}

// ==================== API 工具函数 ====================

/**
 * 创建分页查询参数
 */
function buildQueryParams(config: QueryConfig): URLSearchParams {
  const params = new URLSearchParams();
  params.set('limit', config.pageSize.toString());
  params.set('offset', ((config.page - 1) * config.pageSize).toString());

  if (config.namespace) params.set('namespace', config.namespace);
  if (config.search) params.set('search', config.search);
  if (config.sortBy) params.set('sortBy', config.sortBy);
  if (config.sortOrder) params.set('sortOrder', config.sortOrder);

  return params;
}

/**
 * 类型安全的 API 请求函数
 */
async function fetchResourceList<T extends K8sResource>(
  endpoint: string,
  config: QueryConfig
): Promise<PaginatedResponse<T>> {
  const params = buildQueryParams(config);
  const response = await fetch(`${endpoint}?${params}`, {
    signal: config.search ? undefined : new AbortController().signal,
  });

  if (!response.ok) {
    throw new Error(`HTTP ${response.status}: ${response.statusText}`);
  }

  const result: APIResponse<T[]> = await response.json();

  if (result.code !== 0) {
    throw new Error(result.message || '请求失败');
  }

  return {
    data: result.data || [],
    page: result.page,
  };
}

// ==================== ResourcePage 组件 ====================

/**
 * 通用资源页面组件（泛型版本）
 *
 * 类型安全特性：
 * - data 状态类型与 columns 定义一致
 * - render 函数参数类型自动推断
 * - API 响应结构类型保护
 *
 * @example
 * ```tsx
 * // 使用 Pod 类型
 * <ResourcePage<Pod>
 *   title="Pods"
 *   apiEndpoint="/api/pods"
 *   resourceType="pod"
 *   columns={[
 *     { title: 'Name', dataIndex: 'metadata.name' },
 *     {
 *       title: 'Status',
 *       dataIndex: 'status.phase',
 *       render: (value, record) => <StatusBadge status={value} />
 *     }
 *   ]}
 * />
 * ```
 */
export function ResourcePage<T extends K8sResource>({
  title,
  apiEndpoint,
  resourceType,
  columns,
  collapsed,
  onToggleCollapsed,
  statusMap = {},
  namespaceFilter = true,
  defaultNamespace = '',
  onRowClick,
  customRenderers = {},
}: ResourcePageProps<T>) {
  // ========== 状态定义（类型安全）==========

  const [data, setData] = useState<T[]>([]);
  const [total, setTotal] = useState<number>(0);
  const [loading, setLoading] = useState<boolean>(false);
  const [error, setError] = useState<string | null>(null);
  const [namespace, setNamespace] = useState<string>(defaultNamespace);
  const [search, setSearch] = useState<string>('');

  const {
    page,
    pageSize,
    handlePageChange,
    handlePageSizeChange,
    resetPagination,
  } = usePagination();

  // ========== Refs ==========

  const pageRef = useRef(page);
  const pageSizeRef = useRef(pageSize);
  const namespaceRef = useRef(namespace);
  const searchRef = useRef(search);
  const abortControllerRef = useRef<AbortController | null>(null);
  const isMountedRef = useRef(true);

  // ========== Effects ==========

  // 更新 refs
  useEffect(() => {
    pageRef.current = page;
    pageSizeRef.current = pageSize;
    namespaceRef.current = namespace;
    searchRef.current = search;
  });

  // 组件挂载状态
  useEffect(() => {
    isMountedRef.current = true;
    return () => {
      isMountedRef.current = false;
      if (abortControllerRef.current) {
        abortControllerRef.current.abort();
      }
    };
  }, []);

  // ========== 数据加载 ==========

  const fetchData = useCallback(async () => {
    // 取消之前的请求
    if (abortControllerRef.current) {
      abortControllerRef.current.abort();
    }

    const controller = new AbortController();
    abortControllerRef.current = controller;

    setLoading(true);
    setError(null);

    try {
      const result = await fetchResourceList<T>(apiEndpoint, {
        page: pageRef.current,
        pageSize: pageSizeRef.current,
        namespace: namespaceRef.current,
        search: searchRef.current,
      });

      if (!isMountedRef.current) return;

      setData(result.data);
      setTotal(result.page?.total ?? result.data.length);
    } catch (err) {
      if (!isMountedRef.current) return;

      if (err instanceof Error) {
        if (err.name !== 'AbortError') {
          setError(err.message);
        }
      } else {
        setError('网络错误');
      }
    } finally {
      if (isMountedRef.current) {
        setLoading(false);
      }
    }
  }, [apiEndpoint]);

  // 监听依赖变化，自动加载数据
  useEffect(() => {
    fetchData();
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

  // ========== 事件处理器 ==========

  const handleSearchChange = useCallback(
    (e: React.ChangeEvent<HTMLInputElement>) => {
      setSearch(e.target.value);
    },
    []
  );

  const handleSearchSubmit = useCallback(() => {
    handlePageChange(1);
  }, [handlePageChange]);

  const handleClearSearch = useCallback(() => {
    setSearch('');
    handlePageChange(1);
  }, [handlePageChange]);

  const handlePageSizeChangeWrapper = useCallback(
    (newPageSize: number) => {
      handlePageSizeChange(newPageSize);
    },
    [handlePageSizeChange]
  );

  const handleRefresh = useCallback(() => {
    fetchData();
  }, [fetchData]);

  // ========== 列处理 ==========

  /**
   * 处理列的 render 函数，支持自定义渲染器
   */
  const processedColumns = useMemo<ColumnDef<T>[]>(() => {
    return columns.map((col) => {
      // 如果有自定义渲染器，优先使用
      if (col.dataIndex && customRenderers[col.dataIndex as string]) {
        return {
          ...col,
          render: (value, record, index) =>
            customRenderers[col.dataIndex as string](value, record),
        };
      }
      return col;
    });
  }, [columns, customRenderers]);

  // ========== 渲染 ==========

  if (loading && data.length === 0) {
    return (
      <div className="resource-page">
        <LoadingSpinner text={`加载${title}...`} size="lg" overlay />
      </div>
    );
  }

  if (error && data.length === 0) {
    return (
      <div className="resource-page">
        <ErrorDisplay
          message={error}
          type="error"
          showRetry
          onRetry={handleRefresh}
        />
      </div>
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
          hasSearchResults={search.length > 0 && data.length > 0}
        />
        <RefreshButton onClick={handleRefresh} loading={loading} />
      </PageHeader>

      <CommonTable<T>
        columns={processedColumns}
        data={data}
        emptyText={`暂无 ${title} 数据`}
        onRowClick={onRowClick}
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
}

export default ResourcePage;
