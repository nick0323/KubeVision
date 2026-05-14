/**
 * Authentication Utilities
 * Token 管理and认证RequestProcess
 */

import { STORAGE_KEYS, CUSTOM_EVENTS, API_CONFIG } from '../constants';

export interface TokenInfo {
  token: string;
  expiry?: number;
  username?: string;
}

export const authUtils = {
  isLoggedIn: (): boolean => {
    const token = localStorage.getItem(STORAGE_KEYS.TOKEN);
    if (!token) return false;

    const parts = token.split('.');
    if (parts.length !== 3) return false;

    try {
      const payload = JSON.parse(atob(parts[1]));
      if (payload.exp && payload.exp < Date.now() / 1000) {
        authUtils.clearTokens();
        return false;
      }
    } catch {
      return false;
    }

    return true;
  },

  getToken: (): string | null => {
    return localStorage.getItem(STORAGE_KEYS.TOKEN);
  },

  getRefreshToken: (): string | null => {
    return localStorage.getItem(STORAGE_KEYS.REFRESH_TOKEN);
  },

  setToken: (token: string): void => {
    localStorage.setItem(STORAGE_KEYS.TOKEN, token);

    try {
      const parts = token.split('.');
      if (parts.length === 3) {
        const payload = JSON.parse(atob(parts[1]));
        if (payload.exp) {
          localStorage.setItem(STORAGE_KEYS.TOKEN_EXPIRY, String(payload.exp * 1000));
        }
        if (payload.username) {
          localStorage.setItem(STORAGE_KEYS.TOKEN_USERNAME, payload.username);
        }
      }
    } catch {
    }
  },

  setTokens: (token: string, refreshToken: string): void => {
    authUtils.setToken(token);
    localStorage.setItem(STORAGE_KEYS.REFRESH_TOKEN, refreshToken);
  },

  clearTokens: (): void => {
    localStorage.removeItem(STORAGE_KEYS.TOKEN);
    localStorage.removeItem(STORAGE_KEYS.TOKEN_EXPIRY);
    localStorage.removeItem(STORAGE_KEYS.TOKEN_USERNAME);
    localStorage.removeItem(STORAGE_KEYS.REFRESH_TOKEN);
  },

  getUsername: (): string | null => {
    return localStorage.getItem(STORAGE_KEYS.TOKEN_USERNAME);
  },

  isTokenExpiringSoon: (minutes = 5): boolean => {
    const expiry = localStorage.getItem(STORAGE_KEYS.TOKEN_EXPIRY);
    if (!expiry) return false;

    const expiryTime = parseInt(expiry, 10);
    const now = Date.now();
    const threshold = minutes * 60 * 1000;

    return expiryTime - now < threshold;
  },

  getTokenExpiry: (): number | null => {
    const expiry = localStorage.getItem(STORAGE_KEYS.TOKEN_EXPIRY);
    return expiry ? parseInt(expiry, 10) : null;
  },
};

let _refreshPromise: Promise<boolean> | null = null;

export function createAuthFetch() {
  return async function authFetch(url: string, options: RequestInit = {}, _isRetry = false): Promise<Response> {
    if (!_isRetry && url !== '/api/v1/refresh' && authUtils.isTokenExpiringSoon(5)) {
      await attemptTokenRefresh();
    }

    const token = authUtils.getToken();

    const headers: HeadersInit = {
      'Content-Type': 'application/json',
      ...(token && { Authorization: `Bearer ${token}` }),
      ...options.headers,
    };

    const response = await fetch(url, {
      ...options,
      headers,
    });

    if (response.status === 401 && !_isRetry && url !== '/api/v1/refresh') {
      const refreshed = await attemptRefresh();
      if (refreshed) {
        const retryResponse = await authFetch(url, options);
        return retryResponse;
      }
      logout();
      throw new Error('Session expired');
    }

    return response;
  };
}

async function attemptTokenRefresh(): Promise<boolean> {
  if (_refreshPromise) return _refreshPromise;

  _refreshPromise = (async () => {
    try {
      const refreshToken = authUtils.getRefreshToken();
      if (!refreshToken) return false;

      const response = await fetch('/api/v1/refresh', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ refreshToken }),
      });

      if (!response.ok) return false;

      const result = await response.json();
      if (result.code === 0 && result.data?.token && result.data?.refreshToken) {
        authUtils.setTokens(result.data.token, result.data.refreshToken);
        return true;
      }
      return false;
    } catch {
      return false;
    } finally {
      _refreshPromise = null;
    }
  })();

  return _refreshPromise;
}

/**
 * Create带认证's WebSocket 连接
 * Token 通过 WebSocket 子协议传递（避免 URL 泄露 token）
 * URL query 参数 (token=) 作为向后兼容的 fallback
 */
export function createAuthWebSocket(url: string, _protocols?: string | string[]): WebSocket {
  const token = authUtils.getToken();

  // 优先使用子协议传递 token（不会出现在服务器日志中）
  if (token) {
    const protocols = _protocols
      ? (Array.isArray(_protocols) ? _protocols : [_protocols])
      : [];
    const authProtocols = ['k8svision.auth', token, ...protocols];
    try {
      return new WebSocket(url, authProtocols);
    } catch {
      // fallback: URL query
    }
  }

  // fallback: URL query 参数
  const separator = (url.includes('?') ? '&' : '?');
  const urlWithAuth = token ? `${url}${separator}token=${encodeURIComponent(token)}` : url;
  return new WebSocket(urlWithAuth);
}

// Createglobal authFetch 实例
export const authFetch = createAuthFetch();

/**
 * Add cluster param to URL if not on default cluster
 */
export function withCluster(url: string): string {
  const cluster = localStorage.getItem(STORAGE_KEYS.CURRENT_CLUSTER);
  if (!cluster || cluster === 'default') return url;
  const separator = url.includes('?') ? '&' : '?';
  return `${url}${separator}cluster=${encodeURIComponent(cluster)}`;
}

// WebSocket URL Config
export const getWsUrl = (endpoint: string): string => {
  const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
  const baseUrl = `${protocol}//${window.location.host}${API_CONFIG.BASE_URL}/ws${endpoint}`;
  const cluster = localStorage.getItem(STORAGE_KEYS.CURRENT_CLUSTER);
  if (!cluster || cluster === 'default') return baseUrl;
  const separator = baseUrl.includes('?') ? '&' : '?';
  return `${baseUrl}${separator}cluster=${encodeURIComponent(cluster)}`;
};
