import React, { useState, useCallback, useEffect, useRef } from 'react';
import { TerminalTabProps } from '../types';
import { LoadingSpinner } from '../../LoadingSpinner';
import { FaPlug, FaTimes, FaEraser, FaChevronDown } from 'react-icons/fa';
import NamespaceSelect from '../../NamespaceSelect';
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
        selection: 'rgba(0, 0, 0, 0.2)',
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
    fitAddon.fit();

    xtermRef.current = term;
    fitAddonRef.current = fitAddon;

    // Handle terminal input
    term.onData(data => {
      if (wsRef.current && wsRef.current.readyState === WebSocket.OPEN) {
        wsRef.current.send(data);
      }
    });

    // Handle resize
    const handleResize = () => {
      if (xtermRef.current && fitAddonRef.current) {
        fitAddonRef.current.fit();
      }
    };
    window.addEventListener('resize', handleResize);

    return () => {
      console.log('[TerminalTab] Cleanup: disposing xterm');
      if (fitAddonRef.current) {
        fitAddonRef.current.dispose();
        fitAddonRef.current = null;
      }
      if (xtermRef.current) {
        xtermRef.current.dispose();
        xtermRef.current = null;
      }
      window.removeEventListener('resize', handleResize);
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

    // Connect directly to backend WebSocket (port 8080)
    const token = localStorage.getItem('token');
    const wsProtocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    const wsUrl = `${wsProtocol}//${window.location.hostname}:8080/api/ws/exec?namespace=${namespace}&pod=${name}&container=${containerToUse}&command=${shell}&token=${token}`;

    console.log('Connecting to:', wsUrl);

    const ws = new WebSocket(wsUrl);
    wsRef.current = ws;

    ws.onopen = () => {
      console.log('WebSocket connected');
      setConnected(true);
      setSessionStart(new Date());
      xtermRef.current?.writeln('\r\n\x1b[32mConnected!\x1b[0m\r\n');
    };

    ws.onmessage = event => {
      try {
        const data = JSON.parse(event.data);
        console.log('Received:', data);
        if (data.status === 'connected') {
          xtermRef.current?.writeln(`\x1b[32m${data.message}\x1b[0m`);
          xtermRef.current?.writeln('');
        } else if (data.content) {
          xtermRef.current?.write(data.content);
        } else if (data.message) {
          xtermRef.current?.writeln(data.message);
        }
      } catch (e) {
        // Plain text
        xtermRef.current?.write(event.data);
      }
    };

    ws.onerror = error => {
      console.error('WebSocket error:', error);
      setConnected(false);
      xtermRef.current?.writeln('\r\n\x1b[31mConnection failed\x1b[0m\r\n');
    };

    ws.onclose = () => {
      console.log('WebSocket closed');
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

  // Calculate session duration
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
        <div ref={terminalRef} style={{ height: '100%', display: connected ? 'block' : 'none' }} />
      </div>
    </div>
  );
};

export default TerminalTab;
