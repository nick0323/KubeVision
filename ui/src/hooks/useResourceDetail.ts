import { useState, useCallback, useEffect, useRef } from 'react';
import { apiClient } from '../utils/apiClient';
import { TIME_CONFIG } from '../constants';
import type { UseDetailReturn } from '../types';

/**
 * Detail Query Hook Config
 */
export interface UseResourceDetailOptions {
  resourceType: string;
  namespace: string;
  name: string;
  autoRefresh?: boolean;
  refreshInterval?: number;
}

/**
 * Commonresource详情data Hook
 */
export function useResourceDetail<T = unknown>(
  options: UseResourceDetailOptions
): UseDetailReturn<T> {
  const { resourceType, namespace, name, autoRefresh = false, refreshInterval = TIME_CONFIG.REFRESH_INTERVAL } = options;

  const [data, setData] = useState<T | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const abortControllerRef = useRef<AbortController | null>(null);
  const mountedRef = useRef(true);

  /**
   * Loading...情
   */
  const loadDetail = useCallback(async (forceRefresh = false) => {
    // Cancelbefore'sRequest
    if (abortControllerRef.current) {
      abortControllerRef.current.abort();
    }

    const controller = new AbortController();
    abortControllerRef.current = controller;

    setLoading(true);
    setError(null);

    try {
      const result = await apiClient.request<T>(`/api/${resourceType}/${namespace}/${name}`, {
        signal: controller.signal,
        params: forceRefresh ? { force: 'true' } : undefined,
        retry: 0,
      });

      if (!mountedRef.current) return;

      if (result.code === 0 && result.data) {
        setData(result.data);
      } else {
        setError(result.message || 'Failed to load');
      }
    } catch (err) {
      if (!mountedRef.current) return;

      if (err instanceof Error) {
        if (err.name !== 'AbortError') {
          setError(err.message || 'Network error');
        }
      } else {
        setError('Network error');
      }
    } finally {
      if (mountedRef.current) {
        setLoading(false);
      }
    }
  }, [resourceType, namespace, name]);

  /**
   * manualRefresh
   */
  const refresh = useCallback(async () => {
    await loadDetail(true); // 强制刷新，绕过后端缓存
  }, [loadDetail]);

  /**
   * manualUpdatedata
   */
  const mutate = useCallback((newData: T) => {
    setData(newData);
  }, []);

  // initialLoad
  useEffect(() => {
    mountedRef.current = true;
    loadDetail();

    return () => {
      mountedRef.current = false;
      if (abortControllerRef.current) {
        abortControllerRef.current.abort();
      }
    };
  }, [resourceType, namespace, name, loadDetail]);

  // AutoRefresh
  useEffect(() => {
    if (!autoRefresh || loading) return;

    const interval = setInterval(() => {
      loadDetail();
    }, refreshInterval);

    return () => clearInterval(interval);
  }, [autoRefresh, refreshInterval, loading, loadDetail]);

  return {
    data,
    loading,
    error,
    refresh,
    mutate,
  };
}

export default useResourceDetail;
