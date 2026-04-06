import { useState, useCallback, useEffect, useRef } from 'react';
import { Pod } from '../../../types/k8s-resources';
import { authFetch } from '../../../utils/auth';

export interface UsePodDetailOptions {
  namespace: string;
  name: string;
  autoRefresh?: boolean;
  refreshInterval?: number;
}

export interface UsePodDetailReturn {
  data: Pod | null;
  loading: boolean;
  error: string | null;
  refresh: () => Promise<void>;
  mutate: (data: Pod) => void;
}

/**
 * Pod 详情数据 Hook
 */
export function usePodDetail(options: UsePodDetailOptions): UsePodDetailReturn {
  const { namespace, name, autoRefresh = false, refreshInterval = 30000 } = options;

  const [data, setData] = useState<Pod | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  const abortControllerRef = useRef<AbortController | null>(null);
  const mountedRef = useRef(true);

  /**
   * 加载 Pod 详情
   */
  const loadPod = useCallback(async () => {
    // 取消之前的请求
    if (abortControllerRef.current) {
      abortControllerRef.current.abort();
    }

    const controller = new AbortController();
    abortControllerRef.current = controller;

    setLoading(true);
    setError(null);

    try {
      // 后端只支持单数形式：/api/pod/ns/name
      const response = await authFetch(`/api/pod/${namespace}/${name}`, {
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
  }, [namespace, name]);

  /**
   * 手动刷新
   */
  const refresh = useCallback(async () => {
    await loadPod();
  }, [loadPod]);

  /**
   * 手动更新数据
   */
  const mutate = useCallback((newData: Pod) => {
    setData(newData);
  }, []);

  // 初始加载
  useEffect(() => {
    mountedRef.current = true;
    loadPod();

    return () => {
      mountedRef.current = false;
      if (abortControllerRef.current) {
        abortControllerRef.current.abort();
      }
    };
  }, [namespace, name, loadPod]);

  // 自动刷新
  useEffect(() => {
    if (!autoRefresh || loading) return;

    const interval = setInterval(() => {
      loadPod();
    }, refreshInterval);

    return () => clearInterval(interval);
  }, [autoRefresh, refreshInterval, loading, loadPod]);

  return {
    data,
    loading,
    error,
    refresh,
    mutate,
  };
}

export default usePodDetail;
