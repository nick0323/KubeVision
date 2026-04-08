import React, { useState, useEffect, ChangeEvent, FormEvent } from 'react';
import { LoginPageProps } from '../types';
import { authUtils } from '../utils/auth';
import './LoginPage.css';

/**
 * 登录页面组件
 */
export const LoginPage: React.FC<LoginPageProps> = ({ onLogin }) => {
  const [username, setUsername] = useState<string>('');
  const [password, setPassword] = useState<string>('');
  const [remember, setRemember] = useState<boolean>(false);
  const [loading, setLoading] = useState<boolean>(false);

  const handleSubmit = async (e: FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    setLoading(true);

    // 只保存用户名，不保存密码
    if (remember) {
      localStorage.setItem('remembered_username', username);
    } else {
      localStorage.removeItem('remembered_username');
    }

    try {
      const res = await fetch('/api/login', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ username, password }),
      });

      const data = await res.json();

      if (res.ok && data.code === 0) {
        const token = data.data?.token;

        if (!token) {
          alert('登录响应中没有找到 token');
          return;
        }

        authUtils.setToken(token);

        if (authUtils.isLoggedIn()) {
          onLogin();
        } else {
          alert('Token 验证失败，请重试');
        }
      } else {
        const errMsg = data.message || data.details || '用户名或密码错误';
        alert(errMsg);
      }
    } catch {
      alert('网络错误，请重试');
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    // 只加载保存的用户名
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
        <h2
          style={{
            fontWeight: 700,
            fontSize: 'var(--font-size-lg)',
            marginBottom: 24,
            letterSpacing: 1,
            color: '#2563eb',
          }}
        >
          KubeVision For Kubernetes
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
            style={{ marginRight: 6 }}
            disabled={loading}
          />
          <label
            htmlFor="rememberMe"
            style={{ fontSize: 'var(--font-size-lg)', color: '#666', userSelect: 'none' }}
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
