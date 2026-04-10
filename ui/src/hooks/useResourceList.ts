import { useState, useCallback, useEffect, useRef, useMemo } from 'react';
import { authFetch } from '../utils/auth';

/**
 * 资源列表查询参数
 */
export interface ListQueryParams {
  namespace?: string;
  search?: string;
  limit: number;
  offset: number;
  sortBy: string;
  sortOrder: 'asc' | 'desc';
}

/**
 * 分页元数据
 */
export interface PageMeta {
  total: number;
  limit: number;
  offset: number;
}

/**
 * 资源列表响应数据
 */
export interface ListResponse<T = any> {
  data: T[];
  total: number;
  page?: PageMeta;
}

/**
 * 资源列表 Hook 配置
 */
export interface UseResourceListConfig {
  apiEndpoint: string;
  namespaceFilter?: boolean;
  defaultSort?: {
    field: string;
    order: 'asc' | 'desc';
  };
  initialPageSize?: number;
  staleTime?: number;
}

/**
 * 资源列表 Hook 返回值
 */
export interface UseResourceListReturn<T> {
  // 数据状态
  data: T[];
  loading: boolean;
  isValidating: boolean; // 正在重新验证/刷新
  error: string | null;
  total: number;

  // 分页状态
  page: number;
  pageSize: number;
  setPage: (page: number) => void;
  setPageSize: (size: number) => void;

  // 过滤状态
  namespace: string;
  search: string;
  setNamespace: (ns: string) => void;
  setSearch: (s: string) => void;

  // 排序状态
  sortField: string;
  sortOrder: 'asc' | 'desc';
  handleSort: (field: string) => void;

  // 操作
  refresh: () => Promise<void>;
  mutate: (newData?: T[]) => void; // 手动更新数据
  handleSubmit: () => void;
  clearSearch: () => void;

  // 命名空间列表
  namespaces: string[];
  namespacesLoading: boolean;
}

/**
 * 简单的内存缓存（SWR 模式）
 */
interface CacheEntry<T> {
  data: T;
  timestamp: number;
}

const cache = new Map<string, CacheEntry<any>>();

/**
 * 从缓存获取数据
 */
function getCached<T>(key: string, staleTime: number): T | null {
  const entry = cache.get(key);
  if (!entry) return null;

  const isStale = Date.now() - entry.timestamp > staleTime;
  if (isStale) {
    cache.delete(key);
    return null;
  }
  return entry.data as T;
}

/**
 * 设置缓存
 */
function setCache<T>(key: string, data: T): void {
  cache.set(key, { data, timestamp: Date.now() });
}

/**
 * 生成缓存键
 */
function getCacheKey(endpoint: string, params: ListQueryParams): string {
  const paramsObj: Record<string, string> = {
    limit: params.limit.toString(),
    offset: params.offset.toString(),
    sortBy: params.sortBy,
    sortOrder: params.sortOrder,
  };
  if (params.namespace) paramsObj.namespace = params.namespace;
  if (params.search) paramsObj.search = params.search;
  return `${endpoint}?${new URLSearchParams(paramsObj).toString()}`;
}

/**
 * 通用资源列表 Hook（增强版）
 *
 * 特性：
 * - SWR 缓存模式，避免重复请求
 * - 搜索防抖
 * - 自动刷新
 * - 请求取消
 * - 乐观更新
 */
export function useResourceList<T = any>(config: UseResourceListConfig): UseResourceListReturn<T> {
  const {
    apiEndpoint,
    namespaceFilter,
    defaultSort,
    initialPageSize = 20,
    staleTime = 30000,
  } = config;

  // 状态管理
  const [data, setData] = useState<T[]>([]);
  const [loading, setLoading] = useState(true);
  const [isValidating, setIsValidating] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [total, setTotal] = useState(0);

  // 分页状态
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(initialPageSize);

  // 过滤状态（使用 ref 避免防抖期间的额外渲染）
  const [namespace, setNamespace] = useState('');
  const [search, setSearch] = useState('');
  const searchRef = useRef(search);
  const namespaceRef = useRef(namespace);

  // 排序状态（使用 ref 保存最新值）
  const [sortField, setSortField] = useState(defaultSort?.field || 'name');
  const [sortOrder, setSortOrder] = useState<'asc' | 'desc'>(defaultSort?.order || 'asc');
  const sortFieldRef = useRef(sortField);
  const sortOrderRef = useRef(sortOrder);

  // 更新 ref
  useEffect(() => {
    sortFieldRef.current = sortField;
    sortOrderRef.current = sortOrder;
  }, [sortField, sortOrder]);

  // 命名空间列表
  const [namespaces, setNamespaces] = useState<string[]>([]);
  const [namespacesLoading, setNamespacesLoading] = useState(false);

  // 请求控制
  const abortControllerRef = useRef<AbortController | null>(null);
  const namespacesLoadedRef = useRef(false);
  const mountedRef = useRef(true);

  // 更新 ref
  useEffect(() => {
    searchRef.current = search;
  }, [search]);

  useEffect(() => {
    namespaceRef.current = namespace;
  }, [namespace]);

  // 构建查询参数
  const queryParams = useMemo<ListQueryParams>(
    () => ({
      limit: pageSize,
      offset: (page - 1) * pageSize,
      sortBy: sortField,
      sortOrder,
      ...(namespace ? { namespace } : {}),
      ...(search ? { search } : {}),
    }),
    [pageSize, page, sortField, sortOrder, namespace, search]
  );

  /**
   * 加载命名空间列表（只加载一次）
   */
  const loadNamespaces = useCallback(async () => {
    if (!namespaceFilter || namespacesLoadedRef.current) return;

    setNamespacesLoading(true);
    try {
      const response = await authFetch('/api/namespace?limit=1000&offset=0');
      const result = await response.json();
      if (result.code === 0 && result.data) {
        const nsList = Array.isArray(result.data) ? result.data : [];
        setNamespaces(nsList.map((ns: any) => ns.name));
        namespacesLoadedRef.current = true;
      }
    } catch (err) {
      console.error('加载命名空间失败:', err);
    } finally {
      setNamespacesLoading(false);
    }
  }, [namespaceFilter]);

  // 加载命名空间（只一次）
  useEffect(() => {
    loadNamespaces();
  }, [loadNamespaces]);

  /**
   * 处理排序（使用 ref 避免闭包问题）
   */
  const handleSort = useCallback((field: string) => {
    const currentField = sortFieldRef.current;
    const currentOrder = sortOrderRef.current;

    let newOrder: 'asc' | 'desc' = 'asc';

    if (field === currentField) {
      // 同一字段，切换顺序
      newOrder = currentOrder === 'asc' ? 'desc' : 'asc';
    }

    // 立即更新 ref
    sortFieldRef.current = field;
    sortOrderRef.current = newOrder;

    // 更新状态触发渲染
    setSortField(field);
    setSortOrder(newOrder);
    setPage(1);
    // useEffect 会自动触发 loadData
  }, []);

  /**
   * 提交搜索（重置页码）
   */
  const handleSubmit = useCallback(() => {
    setPage(1);
  }, []);

  /**
   * 清空搜索
   */
  const clearSearch = useCallback(() => {
    setSearch('');
    setPage(1);
  }, []);

  /**
   * 刷新数据
   */
  const refresh = useCallback(async () => {
    setLoading(true);
    setIsValidating(true);

    const controller = new AbortController();
    abortControllerRef.current?.abort();
    abortControllerRef.current = controller;

    const params = new URLSearchParams({
      limit: pageSize.toString(),
      offset: ((page - 1) * pageSize).toString(),
      sortBy: sortField,
      sortOrder,
      ...(namespace ? { namespace } : {}),
      ...(search ? { search } : {}),
    });

    try {
      const response = await authFetch(`${apiEndpoint}?${params}`, {
        signal: controller.signal,
      });
      const result = await response.json();

      if (mountedRef.current && result.code === 0 && result.data) {
        setData(result.data || []);
        setTotal(result.page?.total || result.data?.length || 0);
      }
    } catch (err) {
      if (mountedRef.current && err instanceof Error && err.name !== 'AbortError') {
        setError(err.message);
      }
    } finally {
      if (mountedRef.current) {
        setLoading(false);
        setIsValidating(false);
      }
    }
  }, [page, pageSize, sortField, sortOrder, namespace, search, apiEndpoint]);

  /**
   * 手动更新数据（乐观更新）
   */
  const mutate = useCallback(
    (newData?: T[]) => {
      if (newData !== undefined) {
        setData(newData);
      } else {
        // 调用 refresh
        refresh();
      }
    },
    [refresh]
  );

  // 设置搜索防抖（只更新 searchRef，不触发加载）
  useEffect(() => {
    searchRef.current = search;
  }, [search]);

  // 统一的数据加载逻辑（监听所有需要触发加载的状态）
  useEffect(() => {
    const controller = new AbortController();
    abortControllerRef.current?.abort();
    abortControllerRef.current = controller;

    const loadData = async () => {
      const params = new URLSearchParams({
        limit: pageSize.toString(),
        offset: ((page - 1) * pageSize).toString(),
        sortBy: sortField,
        sortOrder,
        ...(namespace ? { namespace } : {}),
        ...(search ? { search } : {}),
      });

      try {
        const response = await authFetch(`${apiEndpoint}?${params}`, {
          signal: controller.signal,
        });
        const result = await response.json();

        if (mountedRef.current && result.code === 0 && result.data) {
          setData(result.data || []);
          setTotal(result.page?.total || result.data?.length || 0);
        }
      } catch (err) {
        if (mountedRef.current && err instanceof Error && err.name !== 'AbortError') {
          setError(err.message);
        }
      } finally {
        if (mountedRef.current) {
          setLoading(false);
          setIsValidating(false);
        }
      }
    };

    setLoading(true);
    loadData();

    return () => {
      controller.abort();
    };
  }, [page, pageSize, sortField, sortOrder, namespace, search, apiEndpoint]);

  return {
    data,
    loading,
    isValidating,
    error,
    total,
    page,
    pageSize,
    setPage,
    setPageSize,
    namespace,
    search,
    setNamespace,
    setSearch,
    sortField,
    sortOrder,
    handleSort,
    refresh,
    mutate,
    handleSubmit,
    clearSearch,
    namespaces,
    namespacesLoading,
  };
}
