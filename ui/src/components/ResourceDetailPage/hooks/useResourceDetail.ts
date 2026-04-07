import { useState, useCallback, useEffect, useRef } from 'react';
import { authFetch } from '../../../utils/auth';

export interface UseResourceDetailOptions {
  resourceType: string;
  namespace: string;
  name: string;
  autoRefresh?: boolean;
  refreshInterval?: number;
}

export interface UseResourceDetailReturn<T = any> {
  data: T | null;
  loading: boolean;
  error: string | null;
  refresh: () => Promise<void>;
  mutate: (data: T) => void;
}

/**
 * 通用资源详情数据 Hook
 */
export function useResourceDetail<T = any>(options: UseResourceDetailOptions): UseResourceDetailReturn<T> {
  const { resourceType, namespace, name, autoRefresh = false, refreshInterval = 30000 } = options;

  const [data, setData] = useState<T | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const abortControllerRef = useRef<AbortController | null>(null);
  const mountedRef = useRef(true);

  /**
   * 加载资源详情
   */
  const loadDetail = useCallback(async () => {
    // 取消之前的请求
    if (abortControllerRef.current) {
      abortControllerRef.current.abort();
    }

    const controller = new AbortController();
    abortControllerRef.current = controller;

    setLoading(true);
    setError(null);

    try {
      const response = await authFetch(`/api/${resourceType}/${namespace}/${name}`, {
        signal: controller.signal,
      });
      const result = await response.json();

      if (!mountedRef.current) return;

      if (result.code === 0 && result.data) {
        setData(result.data);
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
      }
    }
  }, [resourceType, namespace, name]);

  /**
   * 手动刷新
   */
  const refresh = useCallback(async () => {
    await loadDetail();
  }, [loadDetail]);

  /**
   * 手动更新数据
   */
  const mutate = useCallback((newData: T) => {
    setData(newData);
  }, []);

  // 初始加载
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

  // 自动刷新
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
