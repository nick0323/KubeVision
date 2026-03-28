import React, { useState, useCallback, useEffect, useRef } from 'react';
import { TerminalTabProps } from '../types';
import { LoadingSpinner } from '../../LoadingSpinner';
import './TerminalTab.css';

const SHELL_OPTIONS = ['bash', 'sh', 'zsh', 'powershell', 'cmd'];

/**
 * Terminal Tab - 终端连接
 */
export const TerminalTab: React.FC<TerminalTabProps> = ({ namespace, name, containers }) => {
  const [selectedContainer, setSelectedContainer] = useState<string>('');
  const [shell, setShell] = useState<string>('bash');
  const [connected, setConnected] = useState(false);
  const [connecting, setConnecting] = useState(false);
  const [sessionStart, setSessionStart] = useState<Date | null>(null);
  const terminalRef = useRef<HTMLDivElement>(null);
  const wsRef = useRef<WebSocket | null>(null);

  // 设置默认容器
  useEffect(() => {
    if (containers.length === 1) {
      setSelectedContainer(containers[0].name);
    }
  }, [containers]);

  // 连接终端
  const handleConnect = useCallback(() => {
    if (!selectedContainer) return;
    
    setConnecting(true);
    
    // 模拟 WebSocket 连接（实际项目需要实现 WebSocket）
    setTimeout(() => {
      setConnected(true);
      setConnecting(false);
      setSessionStart(new Date());
      
      if (terminalRef.current) {
        terminalRef.current.innerHTML += `<div class="terminal-line">Connected to ${selectedContainer} (${shell})\n</div>`;
        terminalRef.current.innerHTML += `<div class="terminal-line">$ </div>`;
      }
    }, 1000);
  }, [selectedContainer, shell]);

  // 断开连接
  const handleDisconnect = useCallback(() => {
    wsRef.current?.close();
    setConnected(false);
    setSessionStart(null);
    
    if (terminalRef.current) {
      terminalRef.current.innerHTML += `<div class="terminal-line">\nDisconnected\n</div>`;
    }
  }, []);

  // 清屏
  const handleClear = useCallback(() => {
    if (terminalRef.current) {
      terminalRef.current.innerHTML = '';
    }
  }, []);

  // 计算会话持续时间
  const getSessionDuration = () => {
    if (!sessionStart) return '00:00';
    const now = new Date();
    const diff = Math.floor((now.getTime() - sessionStart.getTime()) / 1000);
    const mins = Math.floor(diff / 60);
    const secs = diff % 60;
    return `${mins.toString().padStart(2, '0')}:${secs.toString().padStart(2, '0')}`;
  };

  const [duration, setDuration] = useState('00:00');
  
  useEffect(() => {
    if (!connected) return;
    const interval = setInterval(() => {
      setDuration(getSessionDuration());
    }, 1000);
    return () => clearInterval(interval);
  }, [connected, sessionStart]);

  return (
    <div className="terminal-tab">
      {/* 工具栏 */}
      <div className="terminal-toolbar">
        {containers.length > 1 && (
          <div className="filter-option">
            <label>容器:</label>
            <select 
              value={selectedContainer} 
              onChange={(e) => setSelectedContainer(e.target.value)}
              disabled={connected}
              className="terminal-select"
            >
              {containers.map((c) => (
                <option key={c.name} value={c.name}>{c.name}</option>
              ))}
            </select>
          </div>
        )}
        
        <div className="filter-option">
          <label>Shell:</label>
          <select 
            value={shell} 
            onChange={(e) => setShell(e.target.value)}
            disabled={connected}
            className="terminal-select"
          >
            {SHELL_OPTIONS.map((s) => (
              <option key={s} value={s}>{s}</option>
            ))}
          </select>
        </div>
        
        <div className="terminal-actions">
          {!connected ? (
            <button 
              className="toolbar-btn primary" 
              onClick={handleConnect}
              disabled={connecting || !selectedContainer}
            >
              {connecting ? '连接中...' : '连接'}
            </button>
          ) : (
            <button 
              className="toolbar-btn danger" 
              onClick={handleDisconnect}
            >
              断开
            </button>
          )}
          <button className="toolbar-btn" onClick={handleClear}>清屏</button>
        </div>
        
        <div className={`terminal-status ${connected ? 'connected' : 'disconnected'}`}>
          <span className="status-dot"></span>
          <span>{connected ? '已连接' : '未连接'}</span>
          {connected && (
            <>
              <span className="separator">|</span>
              <span>会话：{duration}</span>
            </>
          )}
        </div>
      </div>

      {/* 终端区域 */}
      <div className="xterm-wrapper">
        {!connected ? (
          <div className="terminal-placeholder">
            <div className="placeholder-icon">💻</div>
            <div className="placeholder-text">点击"连接"按钮启动终端</div>
            <div className="placeholder-hint">
              支持快捷键：Ctrl+C (中断), Ctrl+L (清屏), Ctrl+D (退出)
            </div>
          </div>
        ) : (
          <div 
            ref={terminalRef} 
            className="terminal-output"
            contentEditable
            suppressContentEditableWarning
          />
        )}
      </div>
    </div>
  );
};

export default TerminalTab;
