import { authUtils } from './auth';
import { ApiResponse, PageMeta } from '../types';

export interface ApiOptions extends RequestInit {
  params?: Record<string, string | number | undefined>;
}

/**
 * API 客户端
 */
export const apiClient = {
  async request<T>(endpoint: string, options: ApiOptions = {}): Promise<ApiResponse<T>> {
    const token = authUtils.getToken();
    const headers: HeadersInit = {
      'Content-Type': 'application/json',
      ...(token && { Authorization: `Bearer ${token}` }),
      ...options.headers,
    };

    // 构建查询参数
    const params = options.params;
    let url = endpoint;
    if (params) {
      const queryString = new URLSearchParams(
        Object.entries(params).filter(([_, v]) => v != null).map(([k, v]) => [k, String(v)])
      ).toString();
      url = queryString ? `${endpoint}?${queryString}` : endpoint;
    }

    const response = await fetch(url, {
      ...options,
      headers,
    });

    if (!response.ok) {
      const error = await response.json().catch(() => ({ message: `HTTP ${response.status}` }));
      throw new Error(error.message || `HTTP ${response.status}`);
    }

    return response.json();
  },

  async get<T>(endpoint: string, params?: Record<string, string | number>): Promise<ApiResponse<T>> {
    return this.request<T>(endpoint, { method: 'GET', params });
  },

  async post<T>(endpoint: string, body?: unknown): Promise<ApiResponse<T>> {
    return this.request<T>(endpoint, {
      method: 'POST',
      body: JSON.stringify(body),
    });
  },
};

/**
 * 分页查询参数
 */
export interface PaginationQueryOptions {
  page: number;
  pageSize: number;
  namespace?: string;
  search?: string;
}

/**
 * 分页查询
 */
export async function createPaginatedQuery<T>(
  endpoint: string,
  options: PaginationQueryOptions
): Promise<{ data: T[]; page: PageMeta }> {
  const { page, pageSize, namespace, search } = options;
  
  const result = await apiClient.get<{ data: T[]; page: PageMeta }>(endpoint, {
    limit: pageSize,
    offset: (page - 1) * pageSize,
    ...(namespace && { namespace }),
    ...(search && { search }),
  });

  return {
    data: result.data || [],
    page: result.page || { total: 0, limit: pageSize, offset: 0 },
  };
}

export default apiClient;
