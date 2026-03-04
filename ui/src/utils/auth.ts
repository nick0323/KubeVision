/**
 * 认证工具
 */

export const authUtils = {
  isLoggedIn: (): boolean => {
    const token = localStorage.getItem('token');
    return !!token && token.split('.').length === 3;
  },

  getToken: (): string | null => {
    return localStorage.getItem('token');
  },

  setToken: (token: string): void => {
    localStorage.setItem('token', token);
  },

  clearToken: (): void => {
    localStorage.removeItem('token');
  },

  setupAuthInterceptor: (onLogout: () => void): void => {
    const token = authUtils.getToken();
    
    if (!token) {
      if (window._originalFetch) {
        window.fetch = window._originalFetch;
      }
      return;
    }

    if (!window._originalFetch) {
      window._originalFetch = window.fetch;
    }

    window.fetch = ((origFetch: typeof fetch) => async (url: string, options: RequestInit = {}) => {
      options.headers = {
        ...options.headers,
        Authorization: `Bearer ${token}`,
      };

      const res = await origFetch(url, options);

      if (res.status === 401) {
        authUtils.clearToken();
        onLogout();
      }

      return res;
    })(window._originalFetch);
  },

  clearAuthInterceptor: (): void => {
    if (window._originalFetch) {
      window.fetch = window._originalFetch;
    }
  },
};
