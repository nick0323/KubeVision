import { authFetch } from './auth';
import { APIResponse, PageMeta } from '../types';

export interface ApiOptions extends RequestInit {
  params?: Record<string, string | number | undefined>;
  timeout?: number;
}

export interface ApiError extends Error {
  code?: number;
  status?: number;
  details?: any;
}

export const apiClient = {
  /**
   * 通用请求方法
   */
  async request<T>(endpoint: string, options: ApiOptions = {}): Promise<APIResponse<T>> {
    const { params, timeout = 30000, ...fetchOptions } = options;

    // 构建 URL
    let url = endpoint;
    if (params) {
      const queryString = new URLSearchParams(
        Object.entries(params)
          .filter(([_, v]) => v != null)
          .map(([k, v]) => [k, String(v)])
      ).toString();
      url = queryString ? `${endpoint}?${queryString}` : endpoint;
    }

    // 创建带超时的请求
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
        throw error;
      }

      const result: APIResponse<T> = await response.json();

      // 后端使用 HTTP 状态码表示状态，code 字段可能为 0 或 HTTP 状态码
      // 只要 HTTP 状态码是 200，就认为是成功
      return result;
    } catch (error) {
      clearTimeout(timeoutId);

      if (error instanceof Error) {
        if (error.name === 'AbortError') {
          const timeoutError: ApiError = new Error('请求超时');
          timeoutError.code = 408;
          throw timeoutError;
        }
      }

      throw error;
    }
  },

  /**
   * GET 请求
   */
  async get<T>(
    endpoint: string,
    params?: Record<string, string | number>
  ): Promise<APIResponse<T>> {
    return this.request<T>(endpoint, { method: 'GET', params });
  },

  /**
   * POST 请求
   */
  async post<T>(endpoint: string, body?: unknown): Promise<APIResponse<T>> {
    return this.request<T>(endpoint, {
      method: 'POST',
      body: JSON.stringify(body),
    });
  },

  /**
   * PUT 请求
   */
  async put<T>(endpoint: string, body?: unknown): Promise<APIResponse<T>> {
    return this.request<T>(endpoint, {
      method: 'PUT',
      body: JSON.stringify(body),
    });
  },

  /**
   * DELETE 请求
   */
  async delete<T>(endpoint: string): Promise<APIResponse<T>> {
    return this.request<T>(endpoint, { method: 'DELETE' });
  },

  async getDetail(resourceType: string, namespace: string, name: string): Promise<any> {
    const endpoint = `/api/${resourceType}/${namespace}/${name}`;
    const result = await this.request<any>(endpoint);
    return result.data;
  },

  async deleteResource(resourceType: string, namespace: string, name: string): Promise<any> {
    const endpoint = `/api/${resourceType}/${namespace}/${name}`;
    const result = await this.request<any>(endpoint, { method: 'DELETE' });
    return result.data;
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

  // 后端返回格式：{ code: 200, message: "xxx", data: { data: [...], page: {...} }, page: {...} }
  // 或者：{ code: 200, message: "xxx", data: [...], page: {...} }
  const responseData = result.data;

  return {
    data: (responseData as any)?.data || responseData || [],
    page: result.page || { total: 0, limit: pageSize, offset: 0 },
  };
}

export default apiClient;
