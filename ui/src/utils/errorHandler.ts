/**
 * Error Handler Utilities
 * 统一的错误处理机制
 */

// ==================== 错误类型定义 ====================

export interface ApiError extends Error {
  code?: number;
  status?: number;
  details?: unknown;
  traceId?: string;
}

export interface ErrorResponse {
  code: number;
  message: string;
  details?: Array<{ field?: string; reason?: string }>;
  traceId?: string;
}

// ==================== 错误消息映射 ====================

const ERROR_MESSAGES: Record<number, string> = {
  400: 'Bad Request - 请求参数错误',
  401: 'Unauthorized - 未授权，请重新登录',
  403: 'Forbidden - 禁止访问',
  404: 'Not Found - 资源不存在',
  408: 'Request Timeout - 请求超时',
  409: 'Conflict - 资源冲突',
  422: 'Unprocessable Entity - 请求参数验证失败',
  429: 'Too Many Requests - 请求过于频繁',
  500: 'Internal Server Error - 服务器内部错误',
  502: 'Bad Gateway - 网关错误',
  503: 'Service Unavailable - 服务不可用',
  504: 'Gateway Timeout - 网关超时',
} as const;

// ==================== 错误处理函数 ====================

/**
 * 获取友好的错误消息
 */
export function getErrorMessage(error: unknown, defaultMsg = '操作失败'): string {
  if (error instanceof Error) {
    // 自定义 ApiError
    if (isApiError(error)) {
      if (error.code && ERROR_MESSAGES[error.code as keyof typeof ERROR_MESSAGES]) {
        return ERROR_MESSAGES[error.code as keyof typeof ERROR_MESSAGES];
      }
      return error.message || defaultMsg;
    }

    // 网络错误
    if (error.name === 'TypeError' && error.message.includes('fetch')) {
      return '网络错误，请检查网络连接';
    }

    // 超时错误
    if (error.name === 'AbortError' || error.name === 'TimeoutError') {
      return '请求超时，请重试';
    }

    return error.message || defaultMsg;
  }

  if (typeof error === 'string') {
    return error || defaultMsg;
  }

  return defaultMsg;
}

/**
 * 判断是否为 ApiError
 */
function isApiError(error: Error): error is ApiError {
  return 'code' in error || 'status' in error;
}

/**
 * 从响应中解析错误
 */
export async function parseResponseError(response: Response): Promise<ApiError> {
  let errorData: ErrorResponse | null = null;

  try {
    const contentType = response.headers.get('content-type');
    if (contentType?.includes('application/json')) {
      errorData = await response.json();
    }
  } catch {
    // 忽略解析错误
  }

  const errorMessage =
    errorData?.message ||
    ERROR_MESSAGES[response.status as keyof typeof ERROR_MESSAGES] ||
    `HTTP ${response.status}`;

  const error: ApiError = new Error(errorMessage);
  error.code = errorData?.code || response.status;
  error.status = response.status;
  error.details = errorData?.details;
  error.traceId = errorData?.traceId;

  return error;
}

/**
 * 处理 API 响应错误
 */
export function handleApiResponse<T>(response: T): { success: boolean; error?: ApiError } {
  const res = response as { code?: number; message?: string };

  if (res.code !== undefined && res.code !== 0 && res.code !== 200) {
    const error: ApiError = new Error(res.message || '操作失败');
    error.code = res.code;
    return { success: false, error };
  }

  return { success: true };
}

/**
 * 日志记录错误（开发环境）
 */
export function logError(error: unknown, context?: string): void {
  if (process.env.NODE_ENV === 'development') {
    console.error('[Error]', context || '', error);
  }
}

// ==================== 错误边界辅助 ====================

/**
 * 重试错误处理
 */
export interface RetryableError {
  onRetry: () => void | Promise<void>;
  maxRetries?: number;
  currentRetry?: number;
}

/**
 * 创建带重试的错误处理器
 */
export function createRetryHandler(options: RetryableError): () => Promise<void> {
  const { onRetry, maxRetries = 3, currentRetry = 0 } = options;

  return async () => {
    if (currentRetry >= maxRetries) {
      throw new Error('超过最大重试次数');
    }

    try {
      await onRetry();
    } catch (error) {
      logError(error, 'Retry failed');
      throw error;
    }
  };
}

// ==================== 导出默认工具对象 ====================

export const errorHandler = {
  getErrorMessage,
  parseResponseError,
  handleApiResponse,
  logError,
  createRetryHandler,
} as const;

export default errorHandler;
