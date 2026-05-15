import { useState, useCallback, useEffect, useRef } from 'react';
import { apiClient } from '../utils/apiClient';
import { PAGINATION_CONFIG } from '../constants';
import type { UseListReturn } from '../types';
import { logError } from '../utils/errorHandler';

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
export interface UseResourceListConfig {
  apiEndpoint: string;
  namespaceFilter?: boolean;
  defaultSort?: {
    field: string;
    order: 'asc' | 'desc';
  };
  initialPageSize?: number;
}

export function useResourceList<T = unknown>(config: UseResourceListConfig): UseListReturn<T> {
  const {
    apiEndpoint,
    namespaceFilter,
    defaultSort,
    initialPageSize = PAGINATION_CONFIG.DEFAULT_PAGE_SIZE,
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

  /**
   * Loading...间List（onlyLoading...
   */
  const loadNamespaces = useCallback(async () => {
    if (!namespaceFilter || namespacesLoadedRef.current) return;

    setNamespacesLoading(true);
    try {
      const result = await apiClient.request<{ name: string }[]>('/api/namespace', {
        params: { limit: 1000, offset: 0 },
        retry: 0,
      });
      if (result.code === 0 && result.data) {
        const nsList = Array.isArray(result.data) ? result.data : [];
        setNamespaces(nsList.map(ns => ns.name));
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

    try {
      const result = await apiClient.request<T[]>(apiEndpoint, {
        signal: controller.signal,
        params: {
          limit: pageSize,
          offset: (page - 1) * pageSize,
          sortBy: sortField,
          sortOrder,
          force: 'true',
          ...(namespace ? { namespace } : undefined),
          ...(search ? { search } : undefined),
        },
        retry: 0,
      });

      if (mountedRef.current && result.code === 0 && result.data) {
        setData(result.data || []);
        setTotal(result.page?.total || result.data?.length || 0);
      } else if (mountedRef.current && result.code !== 0) {
        setError(result.message || 'Request failed');
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
    refresh();

    return () => {
      abortControllerRef.current?.abort();
    };
  }, [page, pageSize, sortField, sortOrder, namespace, search, apiEndpoint, refresh]);

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
