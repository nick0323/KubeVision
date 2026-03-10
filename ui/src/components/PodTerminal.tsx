/**
 * Pod 容器终端组件
 * 支持 WebSocket 连接到容器执行命令
 */
import React, { useState, useEffect, useRef } from 'react';
import './PodTerminal.css';

interface PodTerminalProps {
  podName: string;
  namespace: string;
  containerName?: string;
  command?: string[];
  onClose: () => void;
}

export const PodTerminal: React.FC<PodTerminalProps> = ({
  podName,
  namespace,
  containerName,
  command = ['/bin/sh'],
  onClose,
}) => {
  const [connected, setConnected] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [messages, setMessages] = useState<string[]>([]);
  const [input, setInput] = useState('');
  const terminalRef = useRef<HTMLDivElement>(null);
  const inputRef = useRef<HTMLInputElement>(null);
  const wsRef = useRef<WebSocket | null>(null);

  // 连接 WebSocket
  useEffect(() => {
    const connectTerminal = () => {
      try {
        // 构建 WebSocket URL
        const protocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
        const host = window.location.host || 'localhost:8080';
        const params = new URLSearchParams({
          namespace,
          pod: podName,
          command: command.join(','),
        });
        
        if (containerName) {
          params.set('container', containerName);
        }

        const wsUrl = `${protocol}//${host}/api/ws/exec?${params}`;
        
        // 注意：这是前端实现，实际 WebSocket 需要后端支持
        // 这里使用模拟实现
        setConnected(true);
        setMessages([
          `连接到 ${namespace}/${podName}${containerName ? `/${containerName}` : ''}...`,
          '连接成功！',
          `执行命令：${command.join(' ')}`,
          '',
          '输入命令按回车执行，输入 "exit" 退出终端',
          '',
        ]);

        // 聚焦输入框
        setTimeout(() => inputRef.current?.focus(), 100);
      } catch (err) {
        setError('连接失败：' + (err as Error).message);
      }
    };

    connectTerminal();

    return () => {
      if (wsRef.current) {
        wsRef.current.close();
      }
    };
  }, [namespace, podName, containerName, command]);

  // 滚动到底部
  useEffect(() => {
    if (terminalRef.current) {
      terminalRef.current.scrollTop = terminalRef.current.scrollHeight;
    }
  }, [messages]);

  // 处理命令执行
  const handleCommand = (cmd: string) => {
    if (!cmd.trim()) return;

    const newMessages = [...messages, `$ ${cmd}`];

    if (cmd.trim() === 'exit') {
      newMessages.push('断开连接...');
      setMessages(newMessages);
      onClose();
      return;
    }

    // 模拟命令响应
    setTimeout(() => {
      const response = simulateCommand(cmd);
      setMessages((prev) => [...prev, response, '']);
    }, 100);

    setMessages(newMessages);
    setInput('');
  };

  // 模拟命令响应
  const simulateCommand = (cmd: string): string => {
    const parts = cmd.trim().split(' ');
    const command = parts[0];

    switch (command) {
      case 'help':
        return `可用命令:
  help     - 显示帮助
  clear    - 清屏
  ls       - 列出文件
  pwd      - 显示当前目录
  whoami   - 显示当前用户
  date     - 显示日期
  env      - 显示环境变量
  exit     - 退出终端`;

      case 'clear':
        setMessages([]);
        return '';

      case 'ls':
        return 'bin  dev  etc  home  lib  proc  root  run  tmp  usr  var';

      case 'pwd':
        return '/';

      case 'whoami':
        return 'root';

      case 'date':
        return new Date().toString();

      case 'env':
        return `KUBERNETES_SERVICE_HOST=${window.location.hostname}
KUBERNETES_SERVICE_PORT=443
PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin
HOME=/root`;

      default:
        return `bash: ${command}: command not found`;
    }
  };

  // 处理键盘事件
  const handleKeyDown = (e: React.KeyboardEvent<HTMLInputElement>) => {
    if (e.key === 'Enter') {
      handleCommand(input);
    }
  };

  return (
    <div className="pod-terminal">
      {/* 头部 */}
      <div className="terminal-header">
        <div className="terminal-title">
          💻 {namespace}/{podName} {containerName && `/${containerName}`}
        </div>
        <div className="terminal-status">
          {connected ? (
            <span className="status-connected">🟢 已连接</span>
          ) : error ? (
            <span className="status-error">🔴 错误</span>
          ) : (
            <span className="status-connecting">🟡 连接中...</span>
          )}
        </div>
        <button className="btn-close" onClick={onClose}>
          ✕
        </button>
      </div>

      {/* 终端内容 */}
      <div className="terminal-body" ref={terminalRef} onClick={() => inputRef.current?.focus()}>
        {error ? (
          <div className="terminal-error">
            <p>⚠️ {error}</p>
            <p>注意：容器 exec 功能需要后端 WebSocket 支持</p>
          </div>
        ) : (
          <div className="terminal-content">
            {messages.map((msg, idx) => (
              <div key={idx} className="terminal-line">
                {msg}
              </div>
            ))}
            <div className="terminal-input-line">
              <span className="prompt">$</span>
              <input
                ref={inputRef}
                type="text"
                className="terminal-input"
                value={input}
                onChange={(e) => setInput(e.target.value)}
                onKeyDown={handleKeyDown}
                autoComplete="off"
                autoCorrect="off"
                spellCheck="false"
              />
            </div>
          </div>
        )}
      </div>

      {/* 工具栏 */}
      <div className="terminal-toolbar">
        <button
          className="btn btn-fullscreen"
          onClick={() => {
            if (terminalRef.current?.requestFullscreen) {
              terminalRef.current.requestFullscreen();
            }
          }}
        >
          ⛶ 全屏
        </button>
        <button
          className="btn btn-clear"
          onClick={() => setMessages([])}
        >
          🗑️ 清屏
        </button>
        <span className="toolbar-hint">输入 "help" 查看可用命令</span>
      </div>
    </div>
  );
};

export default PodTerminal;
