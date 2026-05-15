import { authFetch } from './auth';
import { APIResponse, PageMeta } from '../types';
import { API_CONFIG, CACHE_CONFIG, STORAGE_KEYS } from '../constants';

export interface ApiOptions extends RequestInit {
  params?: Record<string, string | number | undefined>;
  timeout?: number;
  retry?: number;
  retryDelay?: number;
}

export interface ApiError extends Error {
  code?: number;
  status?: number;
  details?: unknown;
  traceId?: string;
}

function getClusterParam(): Record<string, string> {
  const cluster = localStorage.getItem(STORAGE_KEYS.CURRENT_CLUSTER);
  return cluster && cluster !== 'default' ? { cluster } : {};
}

export const apiClient = {
  /**
   * CommonRequest方法，支持自动重试（指数退避）
   */
  async request<T>(endpoint: string, options: ApiOptions = {}): Promise<APIResponse<T>> {
    const {
      params: rawParams,
      timeout = API_CONFIG.TIMEOUT,
      retry = CACHE_CONFIG.RETRY_COUNT,
      retryDelay = CACHE_CONFIG.RETRY_DELAY,
      ...fetchOptions
    } = options;

    const params = { ...getClusterParam(), ...rawParams };

    // Build URL
    let url = endpoint;
    if (params) {
      const queryString = new URLSearchParams(
        Object.entries(params)
          .filter(([_, v]) => v != null)
          .map(([k, v]) => [k, String(v)])
      ).toString();
      url = queryString ? `${endpoint}?${queryString}` : endpoint;
    }

    let lastError: Error = new Error('Request failed');

    for (let attempt = 0; attempt <= retry; attempt++) {
      // Create带超时'sRequest
      const controller = new AbortController();
      const timeoutId = setTimeout(() => controller.abort(), timeout);

      try {
        const response = await authFetch(url, {
          ...fetchOptions,
          signal: controller.signal,
        });

        clearTimeout(timeoutId);

      if (!response.ok) {
        const errorData = await response.json().catch(() => ({}));
        const error: ApiError = new Error(
          errorData.message || errorData.details || `HTTP ${response.status}`
        );
        error.code = errorData.code;
        error.status = response.status;
        error.details = errorData.details;
        error.traceId = errorData.traceId;
        throw error;
      }

      const result: APIResponse<T> = await response.json();
      return result;
      } catch (error) {
        clearTimeout(timeoutId);
        lastError = error instanceof Error ? error : new Error(String(error));

        // 不重试主动取消的请求
        if (error instanceof Error && error.name === 'AbortError') {
          const timeoutError: ApiError = new Error('Request timeout');
          timeoutError.code = 408;
          throw timeoutError;
        }

        // 不重试 HTTP 错误（只重试网络错误）
        if (error instanceof Error && 'status' in error) {
          throw error;
        }

        // 最后一次尝试，直接抛出
        if (attempt === retry) {
          throw lastError;
        }

        // 指数退避等待
        const delay = retryDelay * Math.pow(2, attempt);
        await new Promise(resolve => setTimeout(resolve, delay));
      }
    }

    throw lastError;
  },

  /**
   * GET Request
   */
  async get<T>(
    endpoint: string,
    params?: Record<string, string | number>
  ): Promise<APIResponse<T>> {
    return this.request<T>(endpoint, { method: 'GET', params });
  },

  /**
   * POST Request
   */
  async post<T>(endpoint: string, body?: unknown): Promise<APIResponse<T>> {
    return this.request<T>(endpoint, {
      method: 'POST',
      body: JSON.stringify(body),
    });
  },

  /**
   * PUT Request
   */
  async put<T>(endpoint: string, body?: unknown): Promise<APIResponse<T>> {
    return this.request<T>(endpoint, {
      method: 'PUT',
      body: JSON.stringify(body),
    });
  },

  /**
   * DELETE Request
   */
  async delete<T>(endpoint: string): Promise<APIResponse<T>> {
    return this.request<T>(endpoint, { method: 'DELETE' });
  },

  async getDetail<T = unknown>(resourceType: string, namespace: string, name: string): Promise<T> {
    const endpoint = `${API_CONFIG.BASE_URL}/${resourceType}/${namespace}/${name}`;
    const result = await this.request<T>(endpoint);
    return result.data;
  },

  async deleteResource<T = unknown>(resourceType: string, namespace: string, name: string): Promise<T> {
    const endpoint = `${API_CONFIG.BASE_URL}/${resourceType}/${namespace}/${name}`;
    const result = await this.request<T>(endpoint, { method: 'DELETE' });
    return result.data;
  },
};

/**
 * Paginationquery参数
 */
export interface PaginationQueryOptions {
  page: number;
  pageSize: number;
  namespace?: string;
  search?: string;
}

/**
 * Paginationquery
 */
type PaginatedResponse<T> = { data: T[]; page: PageMeta } | T[];

export async function createPaginatedQuery<T>(
  endpoint: string,
  options: PaginationQueryOptions
): Promise<{ data: T[]; page: PageMeta }> {
  const { page, pageSize, namespace, search } = options;

  const result = await apiClient.get<PaginatedResponse<T>>(endpoint, {
    limit: pageSize,
    offset: (page - 1) * pageSize,
    ...(namespace && { namespace }),
    ...(search && { search }),
  });

  const responseData = result.data;

  if (Array.isArray(responseData)) {
    return { data: responseData, page: { total: 0, limit: pageSize, offset: 0 } };
  }
  return {
    data: responseData.data || [],
    page: responseData.page || { total: 0, limit: pageSize, offset: 0 },
  };
}

export default apiClient;
