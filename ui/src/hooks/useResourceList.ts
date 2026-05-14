import { useState, useCallback, useEffect, useRef, useMemo } from 'react';
import { authFetch } from '../utils/auth';
import {
  PAGINATION_CONFIG,
  CACHE_CONFIG,
  STORAGE_KEYS,
} from '../constants';
import type { UseListReturn, ListQueryParams, APIResponse } from '../types';
import { logError } from '../utils/errorHandler';

// Constants definition
const API_ENDPOINTS = {
  NAMESPACE_LIST: '/api/namespace',
} as const;

const DEFAULT_LIMIT = 1000;
const DEFAULT_OFFSET = 0;

/**
 * Resource list hook Config
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

interface CacheEntry<T> {
  data: T;
  timestamp: number;
  lastAccess: number;
}

const MAX_CACHE_SIZE = 50;
const cache = new Map<string, CacheEntry<unknown>>();
const accessOrder: string[] = [];

let cleanupTimer: ReturnType<typeof setInterval> | null = null;

function startCleanup(staleTime: number): void {
  if (cleanupTimer) return;
  cleanupTimer = setInterval(() => {
    const now = Date.now();
    for (const [key, entry] of cache) {
      if (now - entry.timestamp > staleTime * 2) {
        cache.delete(key);
        const idx = accessOrder.indexOf(key);
        if (idx !== -1) accessOrder.splice(idx, 1);
      }
    }
    if (cache.size === 0 && cleanupTimer) {
      clearInterval(cleanupTimer);
      cleanupTimer = null;
    }
  }, 60000);
}

function getCached<T>(key: string, staleTime: number): T | null {
  const entry = cache.get(key);
  if (!entry) return null;

  const isStale = Date.now() - entry.timestamp > staleTime;
  if (isStale) {
    cache.delete(key);
    const idx = accessOrder.indexOf(key);
    if (idx !== -1) accessOrder.splice(idx, 1);
    return null;
  }

  entry.lastAccess = Date.now();
  const idx = accessOrder.indexOf(key);
  if (idx !== -1) accessOrder.splice(idx, 1);
  accessOrder.push(key);

  return entry.data as T;
}

function setCache<T>(key: string, data: T): void {
  if (cache.has(key)) {
    const idx = accessOrder.indexOf(key);
    if (idx !== -1) accessOrder.splice(idx, 1);
  }

  while (cache.size >= MAX_CACHE_SIZE && accessOrder.length > 0) {
    const oldest = accessOrder.shift();
    if (oldest) cache.delete(oldest);
  }

  cache.set(key, { data, timestamp: Date.now(), lastAccess: Date.now() });
  accessOrder.push(key);
}

// Ensure cleanup is stopped on HMR
if (import.meta.hot) {
  import.meta.hot.dispose(() => {
    if (cleanupTimer) {
      clearInterval(cleanupTimer);
      cleanupTimer = null;
    }
  });
}

function getCacheKey(endpoint: string, params: ListQueryParams): string {
  const cluster = localStorage.getItem(STORAGE_KEYS.CURRENT_CLUSTER);
  const paramsObj: Record<string, string> = {
    limit: params.limit.toString(),
    offset: params.offset.toString(),
    sortBy: params.sortBy || '',
    sortOrder: params.sortOrder || '',
  };
  if (cluster && cluster !== 'default') paramsObj.cluster = cluster;
  if (params.namespace) paramsObj.namespace = params.namespace;
  if (params.search) paramsObj.search = params.search;
  return `${endpoint}?${new URLSearchParams(paramsObj).toString()}`;
}

/**
 * CommonResource list hook
 *
 * 特性：
 * - SWR 缓存Mode，避免duplicateRequest
 * - Searchdebounce
 * - AutoRefresh
 * - RequestCancel
 * - optimisticUpdate
 */
export function useResourceList<T = unknown>(config: UseResourceListConfig): UseListReturn<T> {
  const {
    apiEndpoint,
    namespaceFilter,
    defaultSort,
    initialPageSize = PAGINATION_CONFIG.DEFAULT_PAGE_SIZE,
    staleTime = CACHE_CONFIG.STALE_TIME,
  } = config;

  // Status管理
  const [data, setData] = useState<T[]>([]);
  const [loading, setLoading] = useState(true);
  const [isValidating, setIsValidating] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [total, setTotal] = useState(0);

  // PaginationStatus
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(initialPageSize);

  // filterStatus（Use ref 避免debounce期间's额outsideRender）
  const [namespace, setNamespace] = useState('');
  const [search, setSearch] = useState('');
  const searchRef = useRef(search);
  const namespaceRef = useRef(namespace);

  // sortStatus（Use ref Save最新值）
  const [sortField, setSortField] = useState(defaultSort?.field || 'name');
  const [sortOrder, setSortOrder] = useState<'asc' | 'desc'>(defaultSort?.order || 'asc');
  const sortFieldRef = useRef(sortField);
  const sortOrderRef = useRef(sortOrder);

  // Update ref
  useEffect(() => {
    sortFieldRef.current = sortField;
    sortOrderRef.current = sortOrder;
  }, [sortField, sortOrder]);

  // Namespace list
  const [namespaces, setNamespaces] = useState<string[]>([]);
  const [namespacesLoading, setNamespacesLoading] = useState(false);

  // Requestcontrol
  const abortControllerRef = useRef<AbortController | null>(null);
  const namespacesLoadedRef = useRef(false);
  const mountedRef = useRef(true);

  // Update ref
  useEffect(() => {
    searchRef.current = search;
  }, [search]);

  useEffect(() => {
    namespaceRef.current = namespace;
  }, [namespace]);

  // Buildquery参数
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
   * Loading...间List（onlyLoading...
   */
  const loadNamespaces = useCallback(async () => {
    if (!namespaceFilter || namespacesLoadedRef.current) return;

    setNamespacesLoading(true);
    try {
      const cluster = localStorage.getItem(STORAGE_KEYS.CURRENT_CLUSTER);
      const clusterParam = cluster && cluster !== 'default' ? `&cluster=${encodeURIComponent(cluster)}` : '';
      const response = await authFetch(`/api/namespace?limit=1000&offset=0${clusterParam}`);
      const result = await response.json();
      if (result.code === 0 && result.data) {
        const nsList = Array.isArray(result.data) ? result.data : [];
        setNamespaces(nsList.map((ns: { name: string }) => ns.name));
        namespacesLoadedRef.current = true;
      }
    } catch (err) {
      logError(err, 'loadNamespaces');
    } finally {
      setNamespacesLoading(false);
    }
  }, [namespaceFilter]);

  // Loading...间（only一次）
  useEffect(() => {
    loadNamespaces();
  }, [loadNamespaces]);

  /**
   * Processsort（Use ref 避免闭包问题）
   */
  const handleSort = useCallback((field: string) => {
    const currentField = sortFieldRef.current;
    const currentOrder = sortOrderRef.current;

    let newOrder: 'asc' | 'desc' = 'asc';

    if (field === currentField) {
      // 同一字段，toggle顺序
      newOrder = currentOrder === 'asc' ? 'desc' : 'asc';
    }

    // 立即Update ref
    sortFieldRef.current = field;
    sortOrderRef.current = newOrder;

    // UpdateStatustriggerRender
    setSortField(field);
    setSortOrder(newOrder);
    setPage(1);
  }, []);

  /**
   * 提交Search（重置页码）
   */
  const handleSubmit = useCallback(() => {
    setPage(1);
  }, []);

  /**
   * Clear search
   */
  const clearSearch = useCallback(() => {
    setSearch('');
    setPage(1);
  }, []);

  /**
   * Refreshdata
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
      force: 'true',
      ...(namespace ? { namespace } : {}),
      ...(search ? { search } : {}),
    });

    const cluster = localStorage.getItem(STORAGE_KEYS.CURRENT_CLUSTER);
    if (cluster && cluster !== 'default') params.set('cluster', cluster);

    try {
      const response = await authFetch(`${apiEndpoint}?${params}`, {
        signal: controller.signal,
      });

      if (!response.ok) {
        const errorData = await response.json().catch(() => ({}));
        throw new Error(errorData.message || `HTTP ${response.status}`);
      }

      const result = (await response.json()) as APIResponse<T[]>;

      if (mountedRef.current && result.code === 0 && result.data) {
        setData(result.data || []);
        setTotal(result.page?.total || result.data?.length || 0);
        setCache(getCacheKey(apiEndpoint, queryParams), result.data || []);
      } else if (mountedRef.current && result.code !== 0) {
        throw new Error(result.message || 'Request failed');
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
  }, [page, pageSize, sortField, sortOrder, namespace, search, apiEndpoint, queryParams]);

  /**
   * manualUpdatedata（optimisticUpdate）
   */
  const mutate = useCallback(
    (newData?: T[]) => {
      if (newData !== undefined) {
        setData(newData);
      } else {
        refresh();
      }
    },
    [refresh]
  );

  // Unified'sdataLoading...监听所hasneedtriggerLoading...）
  useEffect(() => {
    startCleanup(staleTime);

    const cacheKey = getCacheKey(apiEndpoint, queryParams);
    const cached = getCached<T[]>(cacheKey, staleTime);
    if (cached) {
      setData(cached);
      setLoading(false);
    }

    refresh();

    return () => {
      abortControllerRef.current?.abort();
    };
  }, [page, pageSize, sortField, sortOrder, namespace, search, apiEndpoint, queryParams, staleTime, refresh]);

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
