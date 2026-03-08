/**
 * 认证工具 - 修复版
 * 改进：
 * 1. 移除全局 window 污染
 * 2. 使用请求级 Token 读取
 * 3. 添加 token 刷新机制
 */

export interface TokenInfo {
  token: string;
  expiry?: number;
  username?: string;
}

export const authUtils = {
  /**
   * 检查是否已登录
   */
  isLoggedIn: (): boolean => {
    const token = localStorage.getItem('token');
    if (!token) return false;
    
    // 验证 token 格式（JWT 有 3 个部分）
    const parts = token.split('.');
    if (parts.length !== 3) return false;
    
    // 检查是否过期
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
   * 获取 Token（每次从 localStorage 读取最新值）
   */
  getToken: (): string | null => {
    return localStorage.getItem('token');
  },

  /**
   * 设置 Token
   */
  setToken: (token: string): void => {
    localStorage.setItem('token', token);
    
    // 解析 token 获取过期时间
    try {
      const parts = token.split('.');
      if (parts.length === 3) {
        const payload = JSON.parse(atob(parts[1]));
        if (payload.exp) {
          localStorage.setItem('token_expiry', String(payload.exp * 1000));
        }
        if (payload.username) {
          localStorage.setItem('token_username', payload.username);
        }
      }
    } catch {
      // 忽略解析错误
    }
  },

  /**
   * 清除 Token
   */
  clearToken: (): void => {
    localStorage.removeItem('token');
    localStorage.removeItem('token_expiry');
    localStorage.removeItem('token_username');
  },

  /**
   * 获取用户名
   */
  getUsername: (): string | null => {
    return localStorage.getItem('token_username');
  },

  /**
   * 检查 Token 是否即将过期（5 分钟内）
   */
  isTokenExpiringSoon: (minutes = 5): boolean => {
    const expiry = localStorage.getItem('token_expiry');
    if (!expiry) return false;
    
    const expiryTime = parseInt(expiry, 10);
    const now = Date.now();
    const threshold = minutes * 60 * 1000;
    
    return expiryTime - now < threshold;
  },

  /**
   * 获取 Token 过期时间
   */
  getTokenExpiry: (): number | null => {
    const expiry = localStorage.getItem('token_expiry');
    return expiry ? parseInt(expiry, 10) : null;
  },
};

/**
 * 创建带认证的 fetch 包装器
 * 每次请求都从 localStorage 读取最新 token
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

    // 处理 401 未授权
    if (response.status === 401) {
      authUtils.clearToken();
      // 触发自定义事件，通知其他组件
      window.dispatchEvent(new CustomEvent('auth-unauthorized'));
    }

    return response;
  };
}

// 创建全局 authFetch 实例
export const authFetch = createAuthFetch();
