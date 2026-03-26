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
export interface UseResourceListConfig<T> {
  apiEndpoint: string;
  namespaceFilter?: boolean;
  defaultSort?: {
    field: string;
    order: 'asc' | 'desc';
  };
  initialPageSize?: number;
  staleTime?: number;        // 数据保持新鲜的时长 (ms)，默认 30s
  refreshInterval?: number;  // 自动刷新间隔 (ms)，0 表示不自动刷新
  debounceMs?: number;       // 搜索防抖时间 (ms)，默认 300ms
}

/**
 * 资源列表 Hook 返回值
 */
export interface UseResourceListReturn<T> {
  // 数据状态
  data: T[];
  loading: boolean;
  isValidating: boolean;     // 正在重新验证/刷新
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
  mutate: (newData?: T[]) => void;  // 手动更新数据
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
  return `${endpoint}?${new URLSearchParams(params as Record<string, string>).toString()}`;
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
export function useResourceList<T = any>(
  config: UseResourceListConfig<T>
): UseResourceListReturn<T> {
  const {
    apiEndpoint,
    namespaceFilter,
    defaultSort,
    initialPageSize = 20,
    staleTime = 30000,       // 30 秒
    refreshInterval = 0,     // 默认不自动刷新
    debounceMs = 300,        // 300ms 防抖
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

  // 排序状态
  const [sortField, setSortField] = useState(defaultSort?.field || 'name');
  const [sortOrder, setSortOrder] = useState<'asc' | 'desc'>(
    defaultSort?.order || 'asc'
  );

  // 命名空间列表
  const [namespaces, setNamespaces] = useState<string[]>([]);
  const [namespacesLoading, setNamespacesLoading] = useState(false);

  // 请求控制
  const abortControllerRef = useRef<AbortController | null>(null);
  const namespacesLoadedRef = useRef(false);
  const mountedRef = useRef(true);
  const debounceTimerRef = useRef<NodeJS.Timeout | null>(null);

  // 更新 ref
  useEffect(() => {
    searchRef.current = search;
  }, [search]);

  useEffect(() => {
    namespaceRef.current = namespace;
  }, [namespace]);

  // 构建查询参数
  const queryParams = useMemo<ListQueryParams>(() => ({
    limit: pageSize,
    offset: (page - 1) * pageSize,
    sortBy: sortField,
    sortOrder,
    ...(namespace ? { namespace } : {}),
    ...(search ? { search } : {}),
  }), [pageSize, page, sortField, sortOrder, namespace, search]);

  /**
   * 加载资源列表数据
   */
  const loadData = useCallback(async (isRefresh = false) => {
    const cacheKey = getCacheKey(apiEndpoint, queryParams);
    
    // 检查缓存
    const cachedData = getCached<ListResponse<T>>(cacheKey, staleTime);
    if (cachedData && !isRefresh) {
      setData(cachedData.data || []);
      setTotal(cachedData.page?.total || cachedData.data?.length || 0);
      setLoading(false);
      return;
    }

    // 取消之前的请求
    if (abortControllerRef.current) {
      abortControllerRef.current.abort();
    }

    const controller = new AbortController();
    abortControllerRef.current = controller;

    if (isRefresh) {
      setIsValidating(true);
    } else {
      setLoading(true);
    }
    setError(null);

    try {
      const params = new URLSearchParams({
        limit: queryParams.limit.toString(),
        offset: queryParams.offset.toString(),
        sortBy: queryParams.sortBy,
        sortOrder: queryParams.sortOrder,
      });

      if (queryParams.namespace) params.set('namespace', queryParams.namespace);
      if (queryParams.search) params.set('search', queryParams.search);

      const response = await authFetch(`${apiEndpoint}?${params}`, {
        signal: controller.signal,
      });
      const result = await response.json();

      if (!mountedRef.current) return;

      if (result.code === 0 && result.data) {
        const newData = result.data || [];
        const newTotal = result.page?.total || newData.length || 0;
        
        setData(newData);
        setTotal(newTotal);
        
        // 写入缓存
        setCache(cacheKey, result);
        setError(null);
      } else {
        setError(result.message || '加载失败');
      }
    } catch (err) {
      if (!mountedRef.current) return;
      
      if (err instanceof Error) {
        if (err.name !== 'AbortError') {
          setError(err.message || '网络错误');
        }
      } else {
        setError('网络错误');
      }
    } finally {
      if (mountedRef.current) {
        setLoading(false);
        setIsValidating(false);
      }
    }
  }, [apiEndpoint, queryParams, staleTime]);

  /**
   * 加载命名空间列表（只加载一次）
   */
  const loadNamespaces = useCallback(async () => {
    if (!namespaceFilter || namespacesLoadedRef.current) return;

    setNamespacesLoading(true);
    try {
      // 尝试从缓存获取
      const nsCacheKey = 'namespaces_list';
      const cachedNs = getCached<{ name: string }[]>(nsCacheKey, 60000);
      
      if (cachedNs) {
        setNamespaces(cachedNs.map(ns => ns.name));
        namespacesLoadedRef.current = true;
        return;
      }

      const response = await authFetch('/api/namespaces?limit=1000&offset=0');
      const result = await response.json();
      
      if (result.code === 0 && result.data) {
        const nsList = Array.isArray(result.data) ? result.data : [];
        setNamespaces(nsList.map((ns: any) => ns.name));
        namespacesLoadedRef.current = true;
        
        // 缓存命名空间列表
        setCache(nsCacheKey, nsList);
      }
    } catch (err) {
      console.error('加载命名空间失败:', err);
    } finally {
      setNamespacesLoading(false);
    }
  }, [namespaceFilter]);

  /**
   * 处理排序
   */
  const handleSort = useCallback((field: string) => {
    setSortField((prevField) => {
      if (prevField === field) {
        setSortOrder((prevOrder) => (prevOrder === 'asc' ? 'desc' : 'asc'));
      } else {
        setSortOrder('asc');
      }
      return field;
    });
    setPage(1);
  }, []);

  /**
   * 提交搜索（带防抖）
   */
  const handleSubmit = useCallback(() => {
    if (debounceTimerRef.current) {
      clearTimeout(debounceTimerRef.current);
    }
    debounceTimerRef.current = setTimeout(() => {
      setPage(1);
      loadData();
    }, debounceMs);
  }, [debounceMs, loadData]);

  /**
   * 清空搜索
   */
  const clearSearch = useCallback(() => {
    setSearch('');
    setPage(1);
    if (debounceTimerRef.current) {
      clearTimeout(debounceTimerRef.current);
    }
    debounceTimerRef.current = setTimeout(() => {
      loadData();
    }, debounceMs);
  }, [debounceMs, loadData]);

  /**
   * 刷新数据
   */
  const refresh = useCallback(async () => {
    await loadData(true);
  }, [loadData]);

  /**
   * 手动更新数据（乐观更新）
   */
  const mutate = useCallback((newData?: T[]) => {
    if (newData !== undefined) {
      setData(newData);
    } else {
      // 重新加载
      loadData(true);
    }
  }, [loadData]);

  // 设置搜索防抖
  useEffect(() => {
    if (debounceTimerRef.current) {
      clearTimeout(debounceTimerRef.current);
    }
    
    debounceTimerRef.current = setTimeout(() => {
      loadData();
    }, debounceMs);

    return () => {
      if (debounceTimerRef.current) {
        clearTimeout(debounceTimerRef.current);
      }
    };
  }, [search, namespace, loadData, debounceMs]);

  // 初始加载
  useEffect(() => {
    mountedRef.current = true;
    loadData();
    loadNamespaces();

    return () => {
      mountedRef.current = false;
      if (abortControllerRef.current) {
        abortControllerRef.current.abort();
      }
      if (debounceTimerRef.current) {
        clearTimeout(debounceTimerRef.current);
      }
    };
  }, []);

  // 自动刷新
  useEffect(() => {
    if (refreshInterval <= 0 || loading) return;

    const interval = setInterval(() => {
      loadData(true);
    }, refreshInterval);

    return () => clearInterval(interval);
  }, [refreshInterval, loading, loadData]);

  // 分页、排序变化时加载数据
  useEffect(() => {
    if (page === 1 && namespace === '' && search === '') {
      // 初始状态，跳过
      return;
    }
    loadData();
  }, [page, pageSize, sortField, sortOrder]);

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
