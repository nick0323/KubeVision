import React, { useState, useCallback, useEffect, useRef } from 'react';
import { TerminalTabProps } from '../pages/ResourceDetailPage.types';
import { FaPlug, FaTimes, FaEraser, FaChevronDown } from 'react-icons/fa';
import NamespaceSelect from '../common/NamespaceSelect';
import { Terminal } from 'xterm';
import { FitAddon } from 'xterm-addon-fit';
import 'xterm/css/xterm.css';
import './TerminalTab.css';
import { createAuthWebSocket, getWsUrl } from '../utils/auth';
import { SHELL_CONFIG } from '../constants';

const SHELL_OPTIONS = SHELL_CONFIG.OPTIONS;

/**
 * Terminal Tab - 终端连接
 */
export const TerminalTab: React.FC<TerminalTabProps> = ({ namespace, name, containers }) => {
  const [selectedContainer, setSelectedContainer] = useState<string>(
    containers.length > 0 ? containers[0].name : ''
  );
  const [shell, setShell] = useState<string>('bash');
  const [connected, setConnected] = useState(false);
  const [sessionStart, setSessionStart] = useState<Date | null>(null);
  const [showShellDropdown, setShowShellDropdown] = useState(false);
  const terminalRef = useRef<HTMLDivElement>(null);
  const wsRef = useRef<WebSocket | null>(null);
  const shellRef = useRef<HTMLDivElement>(null);
  const xtermRef = useRef<Terminal | null>(null);
  const fitAddonRef = useRef<FitAddon | null>(null);
  const heartbeatIntervalRef = useRef<NodeJS.Timeout | null>(null);

  // Sync selected container when containers prop changes
  useEffect(() => {
    if (containers.length > 0 && !selectedContainer) {
      setSelectedContainer(containers[0].name);
    }
  }, [containers, selectedContainer]);

  // Initialize xterm
  useEffect(() => {
    if (!terminalRef.current) return;

    const term = new Terminal({
      cursorBlink: true,
      cursorStyle: 'block',
      fontSize: 14,
      fontFamily: 'Consolas, "Courier New", monospace',
      theme: {
        background: '#ffffff',
        foreground: '#000000',
        cursor: '#000000',
        black: '#000000',
        red: '#c62828',
        green: '#2e7d32',
        yellow: '#f57f17',
        blue: '#1565c0',
        magenta: '#6a1b9a',
        cyan: '#00838f',
        white: '#616161',
        brightBlack: '#424242',
        brightRed: '#d32f2f',
        brightGreen: '#388e3c',
        brightYellow: '#f57c00',
        brightBlue: '#1976d2',
        brightMagenta: '#7b1fa2',
        brightCyan: '#0097a7',
        brightWhite: '#9e9e9e',
      },
      convertEol: true,
    });

    const fitAddon = new FitAddon();
    term.loadAddon(fitAddon);
    term.open(terminalRef.current);

    xtermRef.current = term;
    fitAddonRef.current = fitAddon;

    // Handle terminal input
    term.onData(data => {
      if (wsRef.current && wsRef.current.readyState === WebSocket.OPEN) {
        wsRef.current.send(data);
      }
    });

    // 初次 fit - 使用 requestAnimationFrame 确保在下一帧渲染后调用
    requestAnimationFrame(() => {
      // 检查容器是否可见
      if (terminalRef.current && terminalRef.current.offsetWidth > 0 && terminalRef.current.offsetHeight > 0) {
        fitAddon.fit();
      }
      term.focus();
    });

    // Handle resize using ResizeObserver - 监听整个 tab 容器
    const tabContainer = terminalRef.current?.closest('.terminal-tab');
    const resizeObserver = new ResizeObserver(() => {
      if (fitAddonRef.current && xtermRef.current && terminalRef.current) {
        // 防抖处理
        clearTimeout((fitAddonRef.current as any).fitTimeout);
        (fitAddonRef.current as any).fitTimeout = setTimeout(() => {
          // 检查容器是否可见（不为 0）
          if (terminalRef.current && terminalRef.current.offsetWidth > 0 && terminalRef.current.offsetHeight > 0) {
            fitAddonRef.current?.fit();

            // 发送 resize 消息给后端（只在已连接时发送）
            if (wsRef.current && wsRef.current.readyState === WebSocket.OPEN && connected) {
              wsRef.current.send(
                JSON.stringify({
                  type: 'resize',
                  cols: xtermRef.current?.cols || 0,
                  rows: xtermRef.current?.rows || 0,
                })
              );
            }
          }
        }, 150);
      }
    });

    if (tabContainer) {
      resizeObserver.observe(tabContainer);
    }

    return () => {
      resizeObserver.disconnect();
      // 清理 WebSocket 和心跳
      if (heartbeatIntervalRef.current) {
        clearInterval(heartbeatIntervalRef.current);
        heartbeatIntervalRef.current = null;
      }
      if (wsRef.current) {
        wsRef.current.close();
        wsRef.current = null;
      }
      if (fitAddonRef.current) {
        fitAddonRef.current.dispose();
        fitAddonRef.current = null;
      }
      if (xtermRef.current) {
        xtermRef.current.dispose();
        xtermRef.current = null;
      }
    };
  }, []);

  // Close shell dropdown when clicking outside
  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (shellRef.current && !shellRef.current.contains(event.target as Node)) {
        setShowShellDropdown(false);
      }
    };
    document.addEventListener('mousedown', handleClickOutside);
    return () => document.removeEventListener('mousedown', handleClickOutside);
  }, []);

  // Connect to terminal
  const handleConnect = useCallback(() => {
    const containerToUse = selectedContainer || (containers.length > 0 ? containers[0].name : '');

    if (!containerToUse) {
      alert('Please select a container first');
      return;
    }

    // Clear terminal
    xtermRef.current?.clear();

    // 清理旧的心跳定时器
    if (heartbeatIntervalRef.current) {
      clearInterval(heartbeatIntervalRef.current);
      heartbeatIntervalRef.current = null;
    }

    // 关闭旧连接
    if (wsRef.current) {
      wsRef.current.close();
      wsRef.current = null;
    }

    // Connect to backend WebSocket using auth utility
    const wsUrl = getWsUrl(`/exec?namespace=${namespace}&pod=${name}&container=${containerToUse}&command=${shell}`);

    console.log('[Terminal] Connecting to:', wsUrl);

    const ws = createAuthWebSocket(wsUrl);
    wsRef.current = ws;

    ws.onopen = () => {
      console.log('[Terminal] WebSocket connected.');
      setConnected(true);
      setSessionStart(new Date());

      // 浏览器会自动处理 WebSocket 协议层的心跳 (Ping/Pong)，无需应用层手动发送。
      // 发送应用层数据反而可能导致连接异常。
      
      // 连接成功后重新计算终端尺寸
      setTimeout(() => {
        if (fitAddonRef.current && terminalRef.current && xtermRef.current) {
          const width = terminalRef.current.offsetWidth;
          const height = terminalRef.current.offsetHeight;

          if (width > 0 && height > 0) {
            fitAddonRef.current.fit();
            wsRef.current?.send(
              JSON.stringify({
                type: 'resize',
                cols: xtermRef.current?.cols || 0,
                rows: xtermRef.current?.rows || 0,
              })
            );

            setTimeout(() => {
              fitAddonRef.current?.fit();
              wsRef.current?.send(
                JSON.stringify({
                  type: 'resize',
                  cols: xtermRef.current?.cols || 0,
                  rows: xtermRef.current?.rows || 0,
                })
              );
            }, 50);
          }
        }

        xtermRef.current?.writeln('\r\n\x1b[32mConnected!\x1b[0m\r\n');
        xtermRef.current?.focus();
      }, 100);
    };

    ws.onmessage = event => {
      // 处理 Blob 数据（后端发送的是 BinaryMessage）
      if (event.data instanceof Blob) {
        const reader = new FileReader();
        reader.onload = () => {
          const text = reader.result;
          if (typeof text === 'string') {
            try {
              const data = JSON.parse(text);
              if (data.status === 'connected') {
                xtermRef.current?.writeln(`\x1b[32m${data.message}\x1b[0m`);
                xtermRef.current?.writeln('');
              } else if (data.content) {
                xtermRef.current?.write(data.content);
              } else if (data.message) {
                xtermRef.current?.writeln(data.message);
              }
            } catch {
              xtermRef.current?.write(text);
            }
          }
        };
        reader.readAsText(event.data);
      } else {
        try {
          const data = JSON.parse(event.data);
          if (data.status === 'connected') {
            xtermRef.current?.writeln(`\x1b[32m${data.message}\x1b[0m`);
            xtermRef.current?.writeln('');
          } else if (data.content) {
            xtermRef.current?.write(data.content);
          } else if (data.message) {
            xtermRef.current?.writeln(data.message);
          }
        } catch {
          xtermRef.current?.write(event.data);
        }
      }
    };

    ws.onerror = error => {
      console.error('[Terminal] WebSocket error:', error);
      // 清除心跳
      if (heartbeatIntervalRef.current) {
        clearInterval(heartbeatIntervalRef.current);
        heartbeatIntervalRef.current = null;
      }
      setConnected(false);
      xtermRef.current?.writeln('\r\n\x1b[31mConnection error\x1b[0m\r\n');
    };

    ws.onclose = event => {
      console.log('[Terminal] WebSocket closed. Code:', event.code, 'Reason:', event.reason);
      
      // 提示超时或 Shell 退出问题
      if (event.code === 1005 || event.code === 1006) {
        console.warn('[Terminal] Connection was closed abruptly.');
        // 如果连接时间很短（< 10s），通常是 Shell 启动失败或容器不支持交互
        if (sessionStart && (new Date().getTime() - sessionStart.getTime()) < 10000) {
           xtermRef.current?.writeln('\r\n\x1b[31mConnection closed immediately. Possible reasons:\x1b[0m\r\n');
           xtermRef.current?.writeln('\x1b[33m1. Container does not support TTY/Stdin.\x1b[0m\r\n');
           xtermRef.current?.writeln('\x1b[33m2. Selected shell (bash/sh) does not exist in container.\x1b[0m\r\n');
           xtermRef.current?.writeln('\x1b[33m3. Pod is being terminated.\x1b[0m\r\n');
        } else {
           xtermRef.current?.writeln('\r\n\x1b[33mDisconnected (Proxy timeout or network issue)\x1b[0m\r\n');
        }
      }
      
      // 清除心跳
      if (heartbeatIntervalRef.current) {
        clearInterval(heartbeatIntervalRef.current);
        heartbeatIntervalRef.current = null;
      }
      setConnected(false);
      setSessionStart(null);
    };
  }, [namespace, name, selectedContainer, containers, shell]);

  // Disconnect
  const handleDisconnect = useCallback(() => {
    // 清除心跳
    if (heartbeatIntervalRef.current) {
      clearInterval(heartbeatIntervalRef.current);
      heartbeatIntervalRef.current = null;
    }
    wsRef.current?.close();
    setConnected(false);
    setSessionStart(null);
    xtermRef.current?.writeln('\r\n\x1b[33mDisconnected\x1b[0m\r\n');
  }, []);

  // Clear screen
  const handleClear = useCallback(() => {
    xtermRef.current?.clear();
  }, []);

  const [duration, setDuration] = useState('00:00');

  useEffect(() => {
    if (!connected) return;
    const interval = setInterval(() => {
      if (!sessionStart) return;
      const now = new Date();
      const diff = Math.floor((now.getTime() - sessionStart.getTime()) / 1000);
      const mins = Math.floor(diff / 60);
      const secs = diff % 60;
      setDuration(`${mins.toString().padStart(2, '0')}:${secs.toString().padStart(2, '0')}`);
    }, 1000);
    return () => clearInterval(interval);
  }, [connected, sessionStart]);

  return (
    <div className="terminal-tab">
      {/* Toolbar */}
      <div className="terminal-toolbar">
        <div className="terminal-toolbar-left">
          <span className="terminal-title">Terminal</span>
        </div>

        <div className="terminal-toolbar-right">
          {/* Container Select */}
          {containers.length > 1 && (
            <NamespaceSelect
              value={selectedContainer}
              onChange={setSelectedContainer}
              placeholder=""
              options={containers.map(c => c.name)}
              className="terminal-container-select"
              width="120px"
            />
          )}

          {/* Shell Select */}
          <div className="terminal-shell-select" ref={shellRef}>
            <button
              className={`shell-select ${showShellDropdown ? 'active' : ''}`}
              onClick={() => setShowShellDropdown(!showShellDropdown)}
              disabled={connected}
            >
              <span className="shell-value">{shell}</span>
              <FaChevronDown className={`shell-arrow ${showShellDropdown ? 'rotate' : ''}`} />
            </button>
            {showShellDropdown && (
              <div className="shell-menu">
                {SHELL_OPTIONS.map(s => (
                  <button
                    key={s}
                    className={`shell-option ${shell === s ? 'selected' : ''}`}
                    onClick={() => {
                      setShell(s);
                      setShowShellDropdown(false);
                    }}
                  >
                    {s}
                  </button>
                ))}
              </div>
            )}
          </div>

          {/* Action Buttons */}
          <div className="terminal-actions">
            {!connected ? (
              <button
                className="toolbar-btn primary"
                onClick={handleConnect}
                title="Connect to terminal"
              >
                <FaPlug />
              </button>
            ) : (
              <button className="toolbar-btn danger" onClick={handleDisconnect} title="Disconnect">
                <FaTimes />
              </button>
            )}
            <button
              className="toolbar-btn"
              onClick={handleClear}
              title="Clear screen"
              disabled={!connected}
            >
              <FaEraser />
            </button>
          </div>

          {/* Status Display */}
          <div className={`terminal-status ${connected ? 'connected' : 'disconnected'}`}>
            <span className="status-dot"></span>
            <span>{connected ? 'Connected' : 'Disconnected'}</span>
            {connected && (
              <>
                <span className="separator">|</span>
                <span>{duration}</span>
              </>
            )}
          </div>
        </div>
      </div>

      {/* Terminal Output */}
      <div className="xterm-wrapper">
        {!connected && (
          <div className="terminal-guide">
            <div className="guide-title">Terminal</div>
            <div className="guide-steps">
              <div className="guide-step">
                <span className="step-number">1</span>
                <span className="step-text">Select Container (if multiple)</span>
              </div>
              <div className="guide-step">
                <span className="step-number">2</span>
                <span className="step-text">Choose Shell (bash/sh/zsh)</span>
              </div>
              <div className="guide-step">
                <span className="step-number">3</span>
                <span className="step-text">
                  Click <FaPlug className="inline-icon" /> to connect
                </span>
              </div>
            </div>
            <button className="connect-guide-btn" onClick={handleConnect}>
              <FaPlug /> Connect to Terminal
            </button>
          </div>
        )}
        <div
          ref={terminalRef}
          className="terminal-container"
          style={{ display: connected ? 'block' : 'none' }}
        />
      </div>
    </div>
  );
};

export default TerminalTab;
