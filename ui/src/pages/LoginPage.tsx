import React, { useState, useEffect, ChangeEvent, FormEvent } from 'react';
import { LoginPageProps } from '../types';
import { authUtils } from '../utils/auth';
import { apiClient } from '../utils/apiClient';
import { notification } from '../common/NotificationContext';
import { usePageTitle } from '../hooks/usePageTitle';
import k8sLogo from '../assets/kubernetes-logo.svg';
import './LoginPage.css';

/**
 * Login pageComponent
 */
export const LoginPage: React.FC<LoginPageProps> = ({ onLogin }) => {
  usePageTitle('Login');
  const [username, setUsername] = useState<string>('');
  const [password, setPassword] = useState<string>('');
  const [remember, setRemember] = useState<boolean>(false);
  const [loading, setLoading] = useState<boolean>(false);
  const handleSubmit = async (e: FormEvent<HTMLFormElement>) => {
    e.preventDefault();

    // 输入校验
    if (!username || username.trim().length === 0) {
      notification.error('Username cannot be empty');
      return;
    }
    if (username.length < 3 || username.length > 32) {
      notification.error('Username must be between 3 and 32 characters');
      return;
    }
    if (!password || password.length === 0) {
      notification.error('Password cannot be empty');
      return;
    }
    if (password.length < 8 || password.length > 128) {
      notification.error('Password must be between 8 and 128 characters');
      return;
    }

    setLoading(true);

    // onlySaveuser名，notSave密码
    if (remember) {
      localStorage.setItem('remembered_username', username);
    } else {
      localStorage.removeItem('remembered_username');
    }

    try {
      const result = await apiClient.post<{ token: string; refreshToken: string }>('/api/v1/login', { username, password });
      if (result.code === 0) {
        const token = result.data?.token;
        const refreshToken = result.data?.refreshToken;

        if (!token) {
          notification.error('No token found in login response');
          return;
        }

        if (refreshToken) {
          authUtils.setTokens(token, refreshToken);
        } else {
          authUtils.setToken(token);
        }

        if (authUtils.isLoggedIn()) {
          notification.success('Login successful');
          onLogin();
        } else {
          notification.error('Token verification failed, please retry');
        }
      } else {
        const errMsg = result.message || 'Wrong username or password';
        notification.error(errMsg);
      }
    } catch (err) {
      const msg = err instanceof Error ? err.message : 'Network error, please retry';
      notification.error(msg);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    // onlyLoading...user名
    const savedUser = localStorage.getItem('remembered_username') || '';
    if (savedUser) {
      setUsername(savedUser);
      setRemember(true);
    }
  }, []);

  const handleUsernameChange = (e: ChangeEvent<HTMLInputElement>) => {
    setUsername(e.target.value);
  };

  const handlePasswordChange = (e: ChangeEvent<HTMLInputElement>) => {
    setPassword(e.target.value);
  };

  const handleRememberChange = (e: ChangeEvent<HTMLInputElement>) => {
    setRemember(e.target.checked);
  };

  return (
    <div className="login-bg">
      <form onSubmit={handleSubmit} className="login-form-card">
        <h2>
          <img src={k8sLogo} alt="Kubernetes" className="login-logo-icon" />
          KubeVision
        </h2>

        <input
          className="login-input"
          placeholder="Username"
          value={username}
          onChange={handleUsernameChange}
          autoFocus
          disabled={loading}
        />

        <input
          className="login-input"
          type="password"
          placeholder="Password"
          value={password}
          onChange={handlePasswordChange}
          disabled={loading}
        />

        <div className="login-remember-row">
          <input
            id="rememberMe"
            type="checkbox"
            checked={remember}
            onChange={handleRememberChange}
            disabled={loading}
          />
          <label
            htmlFor="rememberMe"
          >
            Remember Username
          </label>
        </div>

        <button
          className="login-btn"
          type="submit"
          disabled={loading}
          style={{ opacity: loading ? 0.7 : 1, cursor: loading ? 'not-allowed' : 'pointer' }}
        >
          {loading ? 'Signing in...' : 'Sign in'}
        </button>
      </form>
    </div>
  );
};

export default LoginPage;
