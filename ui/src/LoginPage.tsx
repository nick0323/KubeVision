import React, { useState, useEffect, ChangeEvent, FormEvent } from 'react';
import { LoginPageProps } from './types';
import './LoginPage.css';

/**
 * 登录页面组件
 * 保持与 LoginPage.jsx 完全一致的功能
 */
export const LoginPage: React.FC<LoginPageProps> = ({ onLogin }) => {
  const [username, setUsername] = useState<string>('');
  const [password, setPassword] = useState<string>('');
  const [error, setError] = useState<string>('');
  const [remember, setRemember] = useState<boolean>(false);

  const handleSubmit = async (e: FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    setError('');

    // 处理"记住我"功能
    if (remember) {
      localStorage.setItem('remembered_username', username);
      localStorage.setItem('remembered_password', password);
    } else {
      // 只有当用户取消勾选"记住我"时才清除保存的信息
      localStorage.removeItem('remembered_username');
      localStorage.removeItem('remembered_password');
    }

    try {
      const res = await fetch('/api/login', {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ username, password })
      });

      if (res.ok) {
        const data = await res.json();

        // 正确提取 token - 根据后端响应格式
        const token = data.data?.token || data.token;

        if (!token) {
          setError('登录响应中没有找到 token');
          return;
        }

        localStorage.setItem('token', token);

        // 调用登录回调，让父组件处理状态更新
        if (onLogin) {
          onLogin();
        }
      } else {
        let errMsg = 'Invalid username or password';
        try {
          const err = await res.json();
          if (err && err.message) errMsg = err.message;
        } catch (e) {
          // Ignore parse error
        }
        setError(errMsg);
      }
    } catch (error) {
      setError('网络错误，请重试');
    }
  };

  useEffect(() => {
    const savedUser = localStorage.getItem('remembered_username') || '';
    const savedPwd = localStorage.getItem('remembered_password') || '';
    if (savedUser && savedPwd) {
      setUsername(savedUser);
      setPassword(savedPwd);
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
        <h2 style={{fontWeight:700, fontSize:'var(--font-size-lg)', marginBottom:24, letterSpacing:1, color:'#2563eb'}}>
          KubeVision For Kubernetes
        </h2>
        <input className="login-input" placeholder="Username" value={username} onChange={handleUsernameChange} autoFocus />
        <input className="login-input" type="password" placeholder="Password" value={password} onChange={handlePasswordChange} />
        <div className="login-remember-row">
          <input id="rememberMe" type="checkbox" checked={remember} onChange={handleRememberChange} style={{marginRight:6}} />
          <label htmlFor="rememberMe" style={{fontSize:'var(--font-size-lg)', color:'#666', userSelect:'none'}}>Remember me</label>
        </div>
        <button className="login-btn" type="submit">Sign in</button>
        {error && <div className="login-error-tip">{error}</div>}
      </form>
    </div>
  );
};

export default LoginPage;
