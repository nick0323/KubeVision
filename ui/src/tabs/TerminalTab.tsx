import React, { useState, useCallback, useEffect, useRef } from 'react';
import { TerminalTabProps } from '../resources/types';
import { FaPlug, FaTimes, FaEraser, FaChevronDown } from 'react-icons/fa';
import NamespaceSelect from '../common/NamespaceSelect';
import { Terminal } from 'xterm';
import { FitAddon } from 'xterm-addon-fit';
import 'xterm/css/xterm.css';
import './TerminalTab.css';

const SHELL_OPTIONS = ['bash', 'sh', 'zsh'];

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
      if (terminalRef.current.offsetWidth > 0 && terminalRef.current.offsetHeight > 0) {
        fitAddon.fit();
      }
      term.focus();
    });

    // Handle resize using ResizeObserver - 监听整个 tab 容器
    const tabContainer = terminalRef.current.closest('.terminal-tab');
    const resizeObserver = new ResizeObserver(() => {
      if (fitAddonRef.current && xtermRef.current && terminalRef.current) {
        // 防抖处理
        clearTimeout((fitAddonRef.current as any).fitTimeout);
        (fitAddonRef.current as any).fitTimeout = setTimeout(() => {
          // 检查容器是否可见（不为 0）
          if (terminalRef.current.offsetWidth > 0 && terminalRef.current.offsetHeight > 0) {
            fitAddonRef.current?.fit();

            // 发送 resize 消息给后端（只在已连接时发送）
            if (wsRef.current && wsRef.current.readyState === WebSocket.OPEN && connected) {
              wsRef.current.send(
                JSON.stringify({
                  type: 'resize',
                  cols: xtermRef.current.cols,
                  rows: xtermRef.current.rows,
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

    // Connect to backend WebSocket (port 8080)
    // WebSocket 连接 - 直接连接后端 8080 端口
    const token = localStorage.getItem('token');
    const wsProtocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    // 直接连接后端 8080 端口（不使用代理）
    const wsUrl = `${wsProtocol}//localhost:8080/api/ws/exec?namespace=${namespace}&pod=${name}&container=${containerToUse}&command=${shell}&token=${token}`;

    const ws = new WebSocket(wsUrl);
    wsRef.current = ws;

    ws.onopen = () => {
      setConnected(true);
      setSessionStart(new Date());

      // 关键修复：连接成功后重新计算终端尺寸
      // 使用 setTimeout 确保 DOM 已经更新（display: none -> block）
      setTimeout(() => {
        if (fitAddonRef.current && terminalRef.current && xtermRef.current) {
          // 检查容器是否可见
          const width = terminalRef.current.offsetWidth;
          const height = terminalRef.current.offsetHeight;

          if (width > 0 && height > 0) {
            // 强制重新计算尺寸
            fitAddonRef.current.fit();

            // 发送初始尺寸给后端
            wsRef.current?.send(
              JSON.stringify({
                type: 'resize',
                cols: xtermRef.current.cols,
                rows: xtermRef.current.rows,
              })
            );

            // 额外确保：再次调用 fit 确保正确
            setTimeout(() => {
              fitAddonRef.current?.fit();

              // 再次发送尺寸确保后端正确
              wsRef.current?.send(
                JSON.stringify({
                  type: 'resize',
                  cols: xtermRef.current.cols,
                  rows: xtermRef.current.rows,
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
              // Plain text
              xtermRef.current?.write(text);
            }
          }
        };
        reader.readAsText(event.data);
      } else {
        // 普通文本消息
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
      setConnected(false);
      xtermRef.current?.writeln('\r\n\x1b[31mConnection error\x1b[0m\r\n');
    };

    ws.onclose = event => {
      console.log('[Terminal] WebSocket closed:', event.code, event.reason);
      setConnected(false);
      setSessionStart(null);
      xtermRef.current?.writeln('\r\n\x1b[33mDisconnected\x1b[0m\r\n');
    };
  }, [namespace, name, selectedContainer, containers, shell]);

  // Disconnect
  const handleDisconnect = useCallback(() => {
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
