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
  /**
   * 检查is否already登录
   */
  isLoggedIn: (): boolean => {
    const token = localStorage.getItem(STORAGE_KEYS.TOKEN);
    if (!token) return false;

    // 验证 token format（JWT has 3 个部分）
    const parts = token.split('.');
    if (parts.length !== 3) return false;

    // 检查is否过期
    try {
      const payload = JSON.parse(atob(parts[1]));
      if (payload.exp && payload.exp < Date.now() / 1000) {
        authUtils.clearToken();
        return false;
      }
    } catch {
      return false;
    }

    return true;
  },

  /**
   * Get Token（every次from localStorage 读取最新值）
   */
  getToken: (): string | null => {
    return localStorage.getItem(STORAGE_KEYS.TOKEN);
  },

  /**
   * settings Token
   */
  setToken: (token: string): void => {
    localStorage.setItem(STORAGE_KEYS.TOKEN, token);

    // 解析 token Get过期time
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
      // 忽略解析Error
    }
  },

  /**
   * 清除 Token
   */
  clearToken: (): void => {
    localStorage.removeItem(STORAGE_KEYS.TOKEN);
    localStorage.removeItem(STORAGE_KEYS.TOKEN_EXPIRY);
    localStorage.removeItem(STORAGE_KEYS.TOKEN_USERNAME);
  },

  /**
   * Getuser名
   */
  getUsername: (): string | null => {
    return localStorage.getItem(STORAGE_KEYS.TOKEN_USERNAME);
  },

  /**
   * 检查 Token is否即will过期（5 分钟inside）
   */
  isTokenExpiringSoon: (minutes = 5): boolean => {
    const expiry = localStorage.getItem(STORAGE_KEYS.TOKEN_EXPIRY);
    if (!expiry) return false;

    const expiryTime = parseInt(expiry, 10);
    const now = Date.now();
    const threshold = minutes * 60 * 1000;

    return expiryTime - now < threshold;
  },

  /**
   * Get Token 过期time
   */
  getTokenExpiry: (): number | null => {
    const expiry = localStorage.getItem(STORAGE_KEYS.TOKEN_EXPIRY);
    return expiry ? parseInt(expiry, 10) : null;
  },
};

/**
 * Create带认证's fetch 包装器
 * every次Requestallfrom localStorage 读取最新 token
 */
export function createAuthFetch() {
  return async function authFetch(url: string, options: RequestInit = {}): Promise<Response> {
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

    // Process 401 not yet授权
    if (response.status === 401) {
      authUtils.clearToken();
      // triggerCustom事 component，Notification其他Component
      window.dispatchEvent(new CustomEvent(CUSTOM_EVENTS.AUTH_UNAUTHORIZED));
    }

    return response;
  };
}

/**
 * Create带认证's WebSocket 连接
 * Token 通过 URL query 参数传递（避免子协议暴露）
 */
export function createAuthWebSocket(url: string, _protocols?: string | string[]): WebSocket {
  const token = authUtils.getToken();

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
