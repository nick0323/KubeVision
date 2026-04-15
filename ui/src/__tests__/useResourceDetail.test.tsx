/**
 * useResourceDetail Hook Test
 */

import { describe, it, expect, jest, beforeEach, afterEach } from '@jest/globals';
import { renderHook, waitFor, act } from '@testing-library/react';
import { useResourceDetail } from '../hooks/useResourceDetail';

// Mock authFetch
const mockAuthFetch = jest.fn();
jest.mock('../utils/auth', () => ({
  authFetch: (...args: any[]) => mockAuthFetch(...args),
}));

// Mock AbortController
const mockAbort = jest.fn();
global.AbortController = jest.fn().mockImplementation(() => ({
  signal: { aborted: false },
  abort: mockAbort,
})) as unknown as typeof AbortController;

describe('useResourceDetail', () => {
  beforeEach(() => {
    jest.clearAllMocks();
    mockAuthFetch.mockClear();
    mockAbort.mockClear();
  });

  afterEach(() => {
    jest.useRealTimers();
  });

  const defaultProps = {
    resourceType: 'pod',
    namespace: 'default',
    name: 'test-pod',
  };

  it('should load data successfully', async () => {
    const mockData = { name: 'test-pod', status: 'Running' };
    const mockResponse = {
      ok: true,
      json: async () => ({ code: 0, data: mockData }),
    };
    mockAuthFetch.mockResolvedValue(mockResponse);

    const { result } = renderHook(() => useResourceDetail(defaultProps));

    expect(result.current.loading).toBe(true);
    expect(result.current.error).toBeNull();

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });

    expect(result.current.data).toEqual(mockData);
    expect(result.current.error).toBeNull();
  });

  it('should handle API error', async () => {
    const mockResponse = {
      ok: true,
      json: async () => ({ code: 404, message: 'Not found' }),
    };
    mockAuthFetch.mockResolvedValue(mockResponse);

    const { result } = renderHook(() => useResourceDetail(defaultProps));

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });

    expect(result.current.data).toBeNull();
    expect(result.current.error).toBe('Not found');
  });

  it('should handle network error', async () => {
    mockAuthFetch.mockRejectedValue(new Error('Network error'));

    const { result } = renderHook(() => useResourceDetail(defaultProps));

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });

    expect(result.current.data).toBeNull();
    expect(result.current.error).toBe('Network error');
  });

  it('should handle abort error without setting error', async () => {
    const abortError = new Error('AbortError');
    abortError.name = 'AbortError';
    mockAuthFetch.mockRejectedValue(abortError);

    const { result } = renderHook(() => useResourceDetail(defaultProps));

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });

    // AbortError 不应该设置错误
    expect(result.current.error).toBeNull();
  });

  it('should handle unknown error type', async () => {
    mockAuthFetch.mockRejectedValue('String error');

    const { result } = renderHook(() => useResourceDetail(defaultProps));

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });

    expect(result.current.error).toBe('Network error');
  });

  it('should call refresh function to reload data', async () => {
    const mockData1 = { name: 'test-pod', status: 'Running' };
    const mockData2 = { name: 'test-pod', status: 'Failed' };

    mockAuthFetch
      .mockResolvedValueOnce({ ok: true, json: async () => ({ code: 0, data: mockData1 }) })
      .mockResolvedValueOnce({ ok: true, json: async () => ({ code: 0, data: mockData2 }) });

    const { result } = renderHook(() => useResourceDetail(defaultProps));

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });

    expect(result.current.data).toEqual(mockData1);

    // 调用 refresh
    await act(async () => {
      await result.current.refresh();
    });

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });

    expect(result.current.data).toEqual(mockData2);
  });

  it('should call mutate function to update data directly', async () => {
    const mockData = { name: 'test-pod', status: 'Running' };
    const newData = { name: 'test-pod', status: 'Failed' };

    mockAuthFetch.mockResolvedValue({ ok: true, json: async () => ({ code: 0, data: mockData }) });

    const { result } = renderHook(() => useResourceDetail(defaultProps));

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });

    expect(result.current.data).toEqual(mockData);

    // 调用 mutate
    act(() => {
      result.current.mutate(newData);
    });

    expect(result.current.data).toEqual(newData);
  });

  it('should cancel previous request when params change', async () => {
    const mockData = { name: 'test-pod', status: 'Running' };
    mockAuthFetch.mockResolvedValue({ ok: true, json: async () => ({ code: 0, data: mockData }) });

    const { rerender } = renderHook(
      ({ namespace }) => useResourceDetail({ ...defaultProps, namespace }),
      { initialProps: { namespace: 'default' } }
    );

    // 更改参数
    rerender({ namespace: 'production' });

    // 应该调用 abort 取消之前的请求
    expect(mockAbort).toHaveBeenCalled();
  });

  it('should not update state after unmount', async () => {
    const mockData = { name: 'test-pod', status: 'Running' };
    
    // 模拟延迟响应
    mockAuthFetch.mockImplementation(
      () => new Promise(resolve => 
        setTimeout(() => resolve({ ok: true, json: async () => ({ code: 0, data: mockData }) }), 100)
      )
    );

    const { unmount } = renderHook(() => useResourceDetail(defaultProps));

    // 立即卸载
    unmount();

    // 等待响应完成，不应该报错
    await act(async () => {
      await new Promise(resolve => setTimeout(resolve, 150));
    });
  });

  it('should auto refresh when autoRefresh is enabled', async () => {
    jest.useFakeTimers();
    
    const mockData = { name: 'test-pod', status: 'Running' };
    mockAuthFetch.mockResolvedValue({ ok: true, json: async () => ({ code: 0, data: mockData }) });

    const { result } = renderHook(() => 
      useResourceDetail({ ...defaultProps, autoRefresh: true, refreshInterval: 5000 })
    );

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });

    // 快进时间
    act(() => {
      jest.advanceTimersByTime(5000);
    });

    // 应该再次调用 loadDetail
    await waitFor(() => {
      expect(mockAuthFetch).toHaveBeenCalledTimes(2);
    });
  });

  it('should not auto refresh when loading', async () => {
    jest.useFakeTimers();
    
    // 模拟持续加载
    mockAuthFetch.mockImplementation(
      () => new Promise(() => {}) // 永远不 resolve
    );

    renderHook(() => 
      useResourceDetail({ ...defaultProps, autoRefresh: true, refreshInterval: 5000 })
    );

    // 快进时间
    act(() => {
      jest.advanceTimersByTime(5000);
      jest.advanceTimersByTime(5000);
    });

    // 只应该调用一次（初始加载）
    expect(mockAuthFetch).toHaveBeenCalledTimes(1);
  });

  it('should cleanup on unmount', async () => {
    mockAuthFetch.mockResolvedValue({ ok: true, json: async () => ({ code: 0, data: {} }) });

    const { unmount } = renderHook(() => useResourceDetail(defaultProps));

    unmount();

    // 应该调用 abort 取消请求
    expect(mockAbort).toHaveBeenCalled();
  });

  it('should use default refresh interval', async () => {
    jest.useFakeTimers();
    
    const mockData = { name: 'test-pod', status: 'Running' };
    mockAuthFetch.mockResolvedValue({ ok: true, json: async () => ({ code: 0, data: mockData }) });

    const { result } = renderHook(() => 
      useResourceDetail({ ...defaultProps, autoRefresh: true })
    );

    await waitFor(() => {
      expect(result.current.loading).toBe(false);
    });

    // 默认刷新间隔 30000ms
    act(() => {
      jest.advanceTimersByTime(30000);
    });

    await waitFor(() => {
      expect(mockAuthFetch).toHaveBeenCalledTimes(2);
    });
  });

  it('should return correct initial state', () => {
    mockAuthFetch.mockImplementation(() => new Promise(() => {})); // 永远不 resolve

    const { result } = renderHook(() => useResourceDetail(defaultProps));

    expect(result.current).toEqual({
      data: null,
      loading: true,
      error: null,
      refresh: expect.any(Function),
      mutate: expect.any(Function),
    });
  });
});
