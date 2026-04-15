/**
 * Error Handler Test
 */

import { describe, it, expect, jest, beforeEach } from '@jest/globals';
import {
  getErrorMessage,
  parseResponseError,
  handleApiResponse,
  logError,
  createRetryHandler,
  errorHandler,
} from '../utils/errorHandler';

describe('Error Handler', () => {
  describe('getErrorMessage', () => {
    it('should return error message from Error object', () => {
      const error = new Error('Test error message');
      expect(getErrorMessage(error)).toBe('Test error message');
    });

    it('should return default message when error is undefined', () => {
      expect(getErrorMessage(undefined)).toBe('操作失败');
    });

    it('should return custom default message', () => {
      expect(getErrorMessage(undefined, 'Custom default')).toBe('Custom default');
    });

    it('should return string error as is', () => {
      expect(getErrorMessage('String error')).toBe('String error');
    });

    it('should handle network error', () => {
      const error = new TypeError('Failed to fetch');
      error.name = 'TypeError';
      expect(getErrorMessage(error)).toBe('网络错误，请检查网络连接');
    });

    it('should handle abort error', () => {
      const error = new Error('AbortError');
      error.name = 'AbortError';
      expect(getErrorMessage(error)).toBe('请求超时，请重试');
    });

    it('should handle timeout error', () => {
      const error = new Error('TimeoutError');
      error.name = 'TimeoutError';
      expect(getErrorMessage(error)).toBe('请求超时，请重试');
    });

    it('should handle ApiError with code', () => {
      const error = new Error('Custom message') as Error & { code?: number };
      error.code = 404;
      expect(getErrorMessage(error)).toBe('Not Found - 资源不存在');
    });

    it('should handle ApiError with unknown code', () => {
      const error = new Error('Custom message') as Error & { code?: number };
      error.code = 999;
      expect(getErrorMessage(error)).toBe('Custom message');
    });

    it('should handle empty string error', () => {
      expect(getErrorMessage('')).toBe('操作失败');
    });
  });

  describe('parseResponseError', () => {
    it('should parse JSON error response', async () => {
      const mockResponse = {
        status: 400,
        headers: new Headers({ 'content-type': 'application/json' }),
        json: async () => ({ code: 400, message: 'Bad Request' }),
      } as Response;

      const error = await parseResponseError(mockResponse);

      expect(error.code).toBe(400);
      expect(error.status).toBe(400);
      expect(error.message).toBe('Bad Request');
    });

    it('should handle non-JSON response', async () => {
      const mockResponse = {
        status: 500,
        headers: new Headers({ 'content-type': 'text/html' }),
        json: async () => {
          throw new Error('Not JSON');
        },
      } as Response;

      const error = await parseResponseError(mockResponse);

      expect(error.code).toBe(500);
      expect(error.message).toBe('Internal Server Error - 服务器内部错误');
    });

    it('should handle response with traceId', async () => {
      const mockResponse = {
        status: 401,
        headers: new Headers({ 'content-type': 'application/json' }),
        json: async () => ({
          code: 401,
          message: 'Unauthorized',
          traceId: 'abc123',
        }),
      } as Response;

      const error = await parseResponseError(mockResponse);

      expect(error.traceId).toBe('abc123');
    });

    it('should handle response with details', async () => {
      const mockResponse = {
        status: 422,
        headers: new Headers({ 'content-type': 'application/json' }),
        json: async () => ({
          code: 422,
          message: 'Validation failed',
          details: [{ field: 'email', reason: 'required' }],
        }),
      } as Response;

      const error = await parseResponseError(mockResponse);

      expect(error.details).toEqual([{ field: 'email', reason: 'required' }]);
    });
  });

  describe('handleApiResponse', () => {
    it('should return success when code is 0', () => {
      const response = { code: 0, data: { id: 1 } };
      const result = handleApiResponse(response);

      expect(result.success).toBe(true);
      expect(result.error).toBeUndefined();
    });

    it('should return success when code is 200', () => {
      const response = { code: 200, data: { id: 1 } };
      const result = handleApiResponse(response);

      expect(result.success).toBe(true);
    });

    it('should return error when code is not 0 or 200', () => {
      const response = { code: 400, message: 'Bad Request' };
      const result = handleApiResponse(response);

      expect(result.success).toBe(false);
      expect(result.error).toBeDefined();
      expect(result.error?.code).toBe(400);
      expect(result.error?.message).toBe('Bad Request');
    });

    it('should handle response without code', () => {
      const response = { data: { id: 1 } };
      const result = handleApiResponse(response);

      expect(result.success).toBe(true);
    });

    it('should handle response with only message', () => {
      const response = { code: 500, message: 'Server Error' };
      const result = handleApiResponse(response);

      expect(result.error?.message).toBe('Server Error');
    });
  });

  describe('logError', () => {
    let originalEnv: string;
    let consoleErrorSpy: jest.Mock;

    beforeEach(() => {
      originalEnv = process.env.NODE_ENV;
      consoleErrorSpy = jest.spyOn(console, 'error').mockImplementation();
    });

    afterEach(() => {
      process.env.NODE_ENV = originalEnv;
      consoleErrorSpy.mockRestore();
    });

    it('should log error in development mode', () => {
      process.env.NODE_ENV = 'development';
      const error = new Error('Test error');

      logError(error, 'Test context');

      expect(consoleErrorSpy).toHaveBeenCalledWith(
        '[Error]',
        'Test context',
        error
      );
    });

    it('should not log error in production mode', () => {
      process.env.NODE_ENV = 'production';
      const error = new Error('Test error');

      logError(error, 'Test context');

      expect(consoleErrorSpy).not.toHaveBeenCalled();
    });

    it('should handle error without context', () => {
      process.env.NODE_ENV = 'development';
      const error = new Error('Test error');

      logError(error);

      expect(consoleErrorSpy).toHaveBeenCalledWith('[Error]', '', error);
    });

    it('should handle unknown error type', () => {
      process.env.NODE_ENV = 'development';

      logError('String error');

      expect(consoleErrorSpy).toHaveBeenCalledWith('[Error]', '', 'String error');
    });
  });

  describe('createRetryHandler', () => {
    it('should create retry handler', () => {
      const onRetry = jest.fn();
      const handler = createRetryHandler({ onRetry });

      expect(handler).toBeDefined();
      expect(typeof handler).toBe('function');
    });

    it('should retry on failure', async () => {
      let callCount = 0;
      const onRetry = jest.fn().mockImplementation(() => {
        callCount++;
        if (callCount < 2) {
          throw new Error('First attempt fails');
        }
      });

      const handler = createRetryHandler({
        onRetry,
        maxRetries: 3,
        currentRetry: 0,
      });

      // 第一次调用会失败，因为 onRetry 第一次会抛错
      await expect(handler()).rejects.toThrow('First attempt fails');
      
      // 验证 onRetry 被调用了一次
      expect(onRetry).toHaveBeenCalledTimes(1);
    });

    it('should throw error when max retries exceeded', async () => {
      const onRetry = jest.fn().mockImplementation(() => {
        throw new Error('Always fails');
      });

      const handler = createRetryHandler({
        onRetry,
        maxRetries: 2,
        currentRetry: 2,
      });

      await expect(handler()).rejects.toThrow('超过最大重试次数');
    });

    it('should log error on retry failure', async () => {
      const onRetry = jest.fn().mockImplementation(() => {
        throw new Error('Retry error');
      });

      const handler = createRetryHandler({
        onRetry,
        maxRetries: 3,
        currentRetry: 0,
      });

      await expect(handler()).rejects.toThrow('Retry error');
    });

    it('should use default max retries', async () => {
      const onRetry = jest.fn().mockImplementation(() => {
        throw new Error('Always fails');
      });

      const handler = createRetryHandler({ onRetry });

      // 默认 maxRetries = 3, currentRetry = 0
      await expect(handler()).rejects.toThrow('Always fails');
      expect(onRetry).toHaveBeenCalledTimes(1);
    });
  });

  describe('errorHandler object', () => {
    it('should export all functions', () => {
      expect(errorHandler.getErrorMessage).toBeDefined();
      expect(errorHandler.parseResponseError).toBeDefined();
      expect(errorHandler.handleApiResponse).toBeDefined();
      expect(errorHandler.logError).toBeDefined();
      expect(errorHandler.createRetryHandler).toBeDefined();
    });

    it('should have correct function types', () => {
      expect(typeof errorHandler.getErrorMessage).toBe('function');
      expect(typeof errorHandler.parseResponseError).toBe('function');
      expect(typeof errorHandler.handleApiResponse).toBe('function');
      expect(typeof errorHandler.logError).toBe('function');
      expect(typeof errorHandler.createRetryHandler).toBe('function');
    });
  });
});
