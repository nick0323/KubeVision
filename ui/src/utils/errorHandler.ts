/**
 * Error Handler Utilities
 * Unified'sErrorProcess机制
 */

// ==================== ErrorType definitions ====================

import { ApiError as ApiClientError } from './apiClient';

export interface ApiError extends ApiClientError {
  traceId?: string;
}

export interface ErrorResponse {
  code: number;
  message: string;
  details?: Array<{ field?: string; reason?: string }>;
  traceId?: string;
}

// ==================== Error messageMapping ====================

const ERROR_MESSAGES: Record<number, string> = {
  400: 'Bad Request - Invalid request parameters',
  401: 'Unauthorized - Unauthorized, please login again',
  403: 'Forbidden - Access forbidden',
  404: 'Not Found - Resource not found',
  408: 'Request Timeout - Request timeout',
  409: 'Conflict - Resource conflict',
  422: 'Unprocessable Entity - Request validation failed',
  429: 'Too Many Requests - Too many requests',
  500: 'Internal Server Error - Internal server error',
  502: 'Bad Gateway - Gateway error',
  503: 'Service Unavailable - Service unavailable',
  504: 'Gateway Timeout - Gateway timeout',
} as const;

// ==================== ErrorProcessfunction ====================

/**
 * Get友好'sError message，附带 traceId
 */
export function getErrorMessage(error: unknown, defaultMsg = 'Operation failed'): string {
  if (error instanceof Error) {
    // Custom ApiError
    if (isApiError(error)) {
      const traceInfo = (error as ApiError).traceId
        ? ` (traceId: ${(error as ApiError).traceId})`
        : '';
      if (error.code && ERROR_MESSAGES[error.code as keyof typeof ERROR_MESSAGES]) {
        return ERROR_MESSAGES[error.code as keyof typeof ERROR_MESSAGES] + traceInfo;
      }
      return (error.message || defaultMsg) + traceInfo;
    }

    // 网络Error
    if (error.name === 'TypeError' && error.message.includes('fetch')) {
      return 'Network error, please check network connection';
    }

    // 超时Error
    if (error.name === 'AbortError' || error.name === 'TimeoutError') {
      return 'Request timeout, please retry';
    }

    return error.message || defaultMsg;
  }

  if (typeof error === 'string') {
    return error || defaultMsg;
  }

  return defaultMsg;
}

/**
 * determineis否for ApiError
 */
function isApiError(error: Error): error is ApiError {
  return 'code' in error || 'status' in error;
}

/**
 * from响应in解析Error
 */
export async function parseResponseError(response: Response): Promise<ApiError> {
  let errorData: ErrorResponse | null = null;

  try {
    const contentType = response.headers.get('content-type');
    if (contentType?.includes('application/json')) {
      errorData = await response.json();
    }
  } catch {
    // 忽略解析Error
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
 * Process API 响应Error
 */
export function handleApiResponse<T>(response: T): { success: boolean; error?: ApiError } {
  const res = response as { code?: number; message?: string };

  if (res.code !== undefined && res.code !== 0 && res.code !== 200) {
    const error: ApiError = new Error(res.message || 'Operation failed');
    error.code = res.code;
    return { success: false, error };
  }

  return { success: true };
}

/**
 * 日志记录Error（开发环境），包含 traceId
 */
export function logError(error: unknown, context?: string): void {
  if (process.env.NODE_ENV === 'development') {
    const traceId = error instanceof Error && 'traceId' in error
      ? (error as ApiError).traceId
      : undefined;
    const traceInfo = traceId ? ` [traceId: ${traceId}]` : '';
    console.error('[Error]', context || '', error, traceInfo);
  }
}

// ==================== Error边界helper ====================

/**
 * RetryErrorProcess
 */
export interface RetryableError {
  onRetry: () => void | Promise<void>;
  maxRetries?: number;
  currentRetry?: number;
}

/**
 * Create带Retry'sErrorProcess器
 */
export function createRetryHandler(options: RetryableError): () => Promise<void> {
  const { onRetry, maxRetries = 3, currentRetry = 0 } = options;

  return async () => {
    if (currentRetry >= maxRetries) {
      throw new Error('Exceeded max retry count');
    }

    try {
      await onRetry();
    } catch (error) {
      logError(error, 'Retry failed');
      throw error;
    }
  };
}

// ==================== Exportdefault工具object ====================

export const errorHandler = {
  getErrorMessage,
  parseResponseError,
  handleApiResponse,
  logError,
  createRetryHandler,
} as const;

export default errorHandler;
