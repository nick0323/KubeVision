import React, { useState, useCallback, useRef, useEffect, useMemo } from 'react';
import { LogsTabProps } from '../resources/types';
import { LoadingSpinner } from '../common/LoadingSpinner';
import { ErrorDisplay } from '../common/ErrorDisplay';
import { FaCog, FaDownload, FaChevronDown, FaEraser } from 'react-icons/fa';
import NamespaceSelect from '../common/NamespaceSelect';
import './LogsTab.css';

const LINES_OPTIONS = [
  { value: '100', label: '100' },
  { value: '200', label: '200' },
  { value: '500', label: '500' },
];

// Performance: max log lines to display
const MAX_LOG_LINES = 1000;
// Virtual scrolling: line height in pixels
const LINE_HEIGHT = 23;
// Virtual scrolling: overscan rows
const OVERSCAN_ROWS = 20;

/**
 * Logs Tab - Log Viewer
 * Performance optimizations:
 * - Virtual scrolling: only render visible rows
 * - useMemo: cache search results
 * - Line limit: max 1000 lines displayed
 */
export const LogsTab: React.FC<LogsTabProps> = ({ namespace, name, containers }) => {
  const [logs, setLogs] = useState<string[]>([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [connected, setConnected] = useState(false);

  // 过滤选项
  const [selectedContainer, setSelectedContainer] = useState<string>(
    containers.length > 0 ? containers[0].name : ''
  );
  const [tailLines, setTailLines] = useState<string>('100');
  const [previous, setPrevious] = useState(false);
  const [timestamps, setTimestamps] = useState(false);
  const [wrapLines, setWrapLines] = useState(true);

  // 搜索
  const [searchTerm, setSearchTerm] = useState('');
  const [currentSearchIndex, setCurrentSearchIndex] = useState(0);

  // WebSocket ref
  const wsRef = useRef<WebSocket | null>(null);
  const reconnectTimeoutRef = useRef<NodeJS.Timeout | null>(null);
  const reconnectCountRef = useRef(0);
  const MAX_RECONNECT_COUNT = 5; // 最多重连 5 次
  const isReconnectingRef = useRef(false); // 防止 StrictMode 下重复重连

  // 设置面板
  const [showSettings, setShowSettings] = useState(false);

  // Tail Lines 下拉
  const [showLinesDropdown, setShowLinesDropdown] = useState(false);
  const linesRef = useRef<HTMLDivElement>(null);

  // 点击外部关闭 Tail Lines 下拉
  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (linesRef.current && !linesRef.current.contains(event.target as Node)) {
        setShowLinesDropdown(false);
      }
    };
    document.addEventListener('mousedown', handleClickOutside);
    return () => document.removeEventListener('mousedown', handleClickOutside);
  }, []);

  const [scrollTop, setScrollTop] = useState(0);
  const [containerHeight, setContainerHeight] = useState(500);

  const logsEndRef = useRef<HTMLDivElement>(null);
  const searchInputRef = useRef<HTMLInputElement>(null);
  const containerRef = useRef<HTMLDivElement>(null);
  const settingsRef = useRef<HTMLDivElement>(null);

  // 点击外部关闭设置面板
  useEffect(() => {
    const handleClickOutside = (event: MouseEvent) => {
      if (settingsRef.current && !settingsRef.current.contains(event.target as Node)) {
        setShowSettings(false);
      }
    };
    document.addEventListener('mousedown', handleClickOutside);
    return () => document.removeEventListener('mousedown', handleClickOutside);
  }, []);

  // Limit log lines
  const limitedLogs = useMemo(() => {
    if (logs.length <= MAX_LOG_LINES) return logs;
    return logs.slice(logs.length - MAX_LOG_LINES);
  }, [logs]);

  const offset = useMemo(() => {
    return logs.length > MAX_LOG_LINES ? logs.length - MAX_LOG_LINES : 0;
  }, [logs]);

  // Load logs via WebSocket
  const loadLogs = useCallback(() => {
    const containerToUse = selectedContainer || (containers.length > 0 ? containers[0].name : '');

    if (!containerToUse) return;

    // Close existing connection
    if (wsRef.current) {
      wsRef.current.close();
    }

    setLoading(true);
    setError(null);
    setLogs([]);

    // Connect to backend WebSocket
    // 后端日志接口：/api/ws/stream?namespace=xxx&pod=xxx&container=xxx&previous=xxx
    // 直接连接后端 8080 端口
    const token = localStorage.getItem('token');
    const wsProtocol = window.location.protocol === 'https:' ? 'wss:' : 'ws:';
    // 直接连接后端 8080 端口（不使用代理）
    const wsUrl = `${wsProtocol}//localhost:8080/api/ws/stream?namespace=${namespace}&pod=${name}&container=${containerToUse}&tailLines=${tailLines}&timestamps=${timestamps}&previous=${previous}&token=${token}`;

    const ws = new WebSocket(wsUrl);
    wsRef.current = ws;

    ws.onopen = () => {
      setConnected(true);
      setLoading(false);
      // 重连成功后重置计数器
      reconnectCountRef.current = 0;
      isReconnectingRef.current = false;
    };

    ws.onmessage = event => {
      try {
        const data = JSON.parse(event.data);

        if (data.type === 'log') {
          const newLines = data.content
            .replace(/\u001b\[[0-9;]*m/g, '')
            .split('\n')
            .filter((line: string) => line.length > 0);

          setLogs(prev => {
            const updated = [...prev, ...newLines];
            if (updated.length > MAX_LOG_LINES) {
              return updated.slice(updated.length - MAX_LOG_LINES);
            }
            return updated;
          });
        } else if (data.type === 'error') {
          setError(data.message);
          setLoading(false);
        } else if (data.type === 'ping') {
          // 响应后端心跳，保持连接
          ws.send(JSON.stringify({ type: 'pong' }));
        }
      } catch {
        // Ignore parse errors
      }
    };

    ws.onerror = error => {
      console.error('WebSocket error:', error);
      setError('WebSocket 连接失败，请检查 Pod 是否存在');
      setLoading(false);
      // 停止重连
      isReconnectingRef.current = true;
    };

    ws.onclose = () => {
      setConnected(false);
      setLoading(false);

      // Pod 不存在时停止重连
      if (error?.includes('not found') || error?.includes('不存在')) {
        setError('Pod 不存在或已被删除');
        isReconnectingRef.current = true;
        return;
      }

      // 防止 StrictMode 下重复重连
      if (isReconnectingRef.current) {
        return;
      }

      // 检查重连次数
      if (reconnectCountRef.current >= MAX_RECONNECT_COUNT) {
        setError('连接失败：已达到最大重连次数');
        return;
      }

      reconnectCountRef.current += 1;
      isReconnectingRef.current = true;

      // 3 秒后自动重连
      reconnectTimeoutRef.current = setTimeout(() => {
        isReconnectingRef.current = false;
        loadLogs();
      }, 3000);
    };
  }, [namespace, name, selectedContainer, containers, tailLines, timestamps]);

  // Load logs when filter options change
  useEffect(() => {
    loadLogs();
  }, [loadLogs]);

  // Cleanup on unmount
  useEffect(() => {
    return () => {
      isReconnectingRef.current = false;

      // 关闭 WebSocket 连接
      if (wsRef.current) {
        wsRef.current.close();
        wsRef.current = null;
      }

      // 清除重连定时器
      if (reconnectTimeoutRef.current) {
        clearTimeout(reconnectTimeoutRef.current);
        reconnectTimeoutRef.current = null;
      }

      reconnectCountRef.current = 0;
    };
  }, []);

  // Auto scroll to bottom when new logs arrive
  useEffect(() => {
    if (containerRef.current) {
      containerRef.current.scrollTop = containerRef.current.scrollHeight;
    }
  }, [logs]);

  // Handle container scroll
  const handleScroll = useCallback(() => {
    if (containerRef.current) {
      setScrollTop(containerRef.current.scrollTop);
      setContainerHeight(containerRef.current.clientHeight);
    }
  }, []);

  // Search logs - memoized (real-time search as you type)
  const searchResults = useMemo(() => {
    if (!searchTerm || searchTerm.trim() === '' || limitedLogs.length === 0) return [];

    const results: number[] = [];
    const searchLower = searchTerm.toLowerCase().trim();
    limitedLogs.forEach((line, index) => {
      if (line.toLowerCase().includes(searchLower)) {
        // 存储实际索引（与 DOM id 一致）
        results.push(index + offset);
      }
    });
    return results;
  }, [searchTerm, limitedLogs, offset]);

  // Reset search index when search results change
  useEffect(() => {
    setCurrentSearchIndex(0);
  }, [searchResults.length]);

  // Next match
  const handleNext = useCallback(() => {
    if (searchResults.length === 0) return;
    const nextIndex = (currentSearchIndex + 1) % searchResults.length;
    setCurrentSearchIndex(nextIndex);

    // 使用 setTimeout 确保状态更新后再滚动
    setTimeout(() => {
      // searchResults 现在存储的是实际索引（与 DOM id 一致）
      const actualIndex = searchResults[nextIndex];
      const element = document.getElementById(`log-line-${actualIndex}`);
      if (element) {
        element.scrollIntoView({ behavior: 'smooth', block: 'center' });
        element.classList.add('highlight-flash');
        setTimeout(() => element.classList.remove('highlight-flash'), 500);
      }
    }, 50);
  }, [searchResults, currentSearchIndex]);

  // Copy logs
  // const handleCopy = useCallback(() => {
  //   navigator.clipboard.writeText(limitedLogs.join('\n'));
  // }, [limitedLogs]);

  // Download logs
  const handleDownload = useCallback(() => {
    const blob = new Blob([limitedLogs.join('\n')], { type: 'text/plain' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = `${name}-${selectedContainer}-logs.txt`;
    a.click();
    URL.revokeObjectURL(url);
  }, [limitedLogs, name, selectedContainer]);

  // Clear logs
  const handleClear = useCallback(() => {
    setLogs([]);
  }, []);

  // Render log lines with virtual scrolling
  const renderedLines = useMemo(() => {
    // Wrap mode: render all lines (virtual scrolling not suitable for variable height)
    if (wrapLines) {
      const lines: React.ReactNode[] = limitedLogs.map((line, i) => {
        const actualIndex = i + offset;
        const isInSearchResults = searchResults.includes(actualIndex);

        let className = 'log-line';
        if (isInSearchResults) className += ' highlight';
        if (line.toLowerCase().includes('error')) className += ' error';
        else if (line.toLowerCase().includes('warn')) className += ' warn';
        else if (line.toLowerCase().includes('info')) className += ' info';

        return (
          <div key={actualIndex} id={`log-line-${actualIndex}`} className={className}>
            {line}
          </div>
        );
      });
      return { lines, visibleStart: 0, visibleEnd: lines.length };
    }

    // Non-wrap mode: use virtual scrolling
    const visibleStart = Math.max(0, Math.floor(scrollTop / LINE_HEIGHT) - OVERSCAN_ROWS);
    const visibleEnd = Math.min(
      limitedLogs.length,
      Math.ceil((scrollTop + containerHeight) / LINE_HEIGHT) + OVERSCAN_ROWS
    );

    const lines: React.ReactNode[] = [];
    for (let i = visibleStart; i < visibleEnd; i++) {
      const line = limitedLogs[i];
      const actualIndex = i + offset;
      const isInSearchResults = searchResults.includes(actualIndex);

      let className = 'log-line';
      if (isInSearchResults) className += ' highlight';
      if (line.toLowerCase().includes('error')) className += ' error';
      else if (line.toLowerCase().includes('warn')) className += ' warn';
      else if (line.toLowerCase().includes('info')) className += ' info';

      lines.push(
        <div
          key={actualIndex}
          id={`log-line-${actualIndex}`}
          className={className}
          style={wrapLines ? {} : { height: `${LINE_HEIGHT}px` }}
        >
          {line}
        </div>
      );
    }

    return { lines, visibleStart, visibleEnd };
  }, [limitedLogs, scrollTop, containerHeight, searchResults, offset, wrapLines]);

  return (
    <div className="logs-tab">
      {/* Search and Filter Bar */}
      <div className="filter-options">
        <div className="filter-options-left">
          <span className="logs-title">
            Logs
            <span className="logs-total">
              {logs.filter(line => line.trim() !== '').length} Lines
            </span>
          </span>
          <div className="search-input-wrapper">
            <input
              ref={searchInputRef}
              type="text"
              className="search-input"
              placeholder="Search logs..."
              value={searchTerm}
              onChange={e => setSearchTerm(e.target.value)}
            />
            {searchTerm && (
              <button
                className="search-btn clear-btn"
                onClick={() => setSearchTerm('')}
                title="Clear search"
              >
                ✕
              </button>
            )}
            {searchResults.length > 0 && (
              <button className="search-btn" onClick={handleNext} title="Next match">
                ↓ ({currentSearchIndex + 1}/{searchResults.length})
              </button>
            )}
          </div>
        </div>

        <div className="filter-options-right">
          {/* Container 选择 */}
          {containers.length > 1 && (
            <NamespaceSelect
              value={selectedContainer}
              onChange={setSelectedContainer}
              placeholder=""
              options={containers.map(c => c.name)}
              className="logs-container-select"
              width="120px"
            />
          )}

          {/* Settings dropdown */}
          <div className="settings-dropdown" ref={settingsRef}>
            <button
              className={`settings-btn ${showSettings ? 'active' : ''}`}
              onClick={() => setShowSettings(!showSettings)}
              title="Settings"
            >
              <FaCog />
            </button>
            {showSettings && (
              <div className="settings-panel">
                <div className="setting-item setting-item-select" ref={linesRef}>
                  <label>Initial Lines</label>
                  <div className="custom-dropdown">
                    <button
                      className={`dropdown-trigger ${showLinesDropdown ? 'active' : ''}`}
                      onClick={() => setShowLinesDropdown(!showLinesDropdown)}
                    >
                      <span className="dropdown-value">{tailLines}</span>
                      <FaChevronDown
                        className={`dropdown-arrow ${showLinesDropdown ? 'rotate' : ''}`}
                      />
                    </button>
                    {showLinesDropdown && (
                      <div className="dropdown-menu">
                        {LINES_OPTIONS.map(opt => (
                          <button
                            key={opt.value}
                            className={`dropdown-option ${tailLines === opt.value ? 'selected' : ''}`}
                            onClick={() => {
                              setTailLines(opt.value);
                              setShowLinesDropdown(false);
                            }}
                          >
                            {opt.label}
                          </button>
                        ))}
                      </div>
                    )}
                  </div>
                </div>
                <div className="setting-item setting-item-toggle">
                  <label>Show Timestamps</label>
                  <button
                    className={`toggle-btn ${timestamps ? 'active' : ''}`}
                    onClick={() => setTimestamps(!timestamps)}
                  />
                </div>
                <div className="setting-item setting-item-toggle">
                  <label>Previous Container</label>
                  <button
                    className={`toggle-btn ${previous ? 'active' : ''}`}
                    onClick={() => setPrevious(!previous)}
                  />
                </div>
                <div className="setting-item setting-item-toggle">
                  <label>Word Wrap</label>
                  <button
                    className={`toggle-btn ${wrapLines ? 'active' : ''}`}
                    onClick={() => setWrapLines(!wrapLines)}
                  />
                </div>
              </div>
            )}
          </div>

          {/* Action buttons */}
          <div className="filter-actions">
            <button className="action-btn" onClick={handleDownload} title="Download logs">
              <FaDownload />
            </button>
            <button className="action-btn" onClick={handleClear} title="Clear logs">
              <FaEraser />
            </button>
          </div>

          {/* Status Display */}
          <div className={`filter-status ${connected ? 'connected' : 'disconnected'}`}>
            <span className="status-dot"></span>
            <span>{connected ? 'Live' : 'Disconnected'}</span>
          </div>
        </div>
      </div>

      {/* Log Content */}
      {loading && logs.length === 0 ? (
        <LoadingSpinner text="Loading logs..." size="lg" />
      ) : error ? (
        <ErrorDisplay message={error} type="error" showRetry onRetry={loadLogs} />
      ) : logs.length === 0 ? (
        <div className="empty-state">
          <span className="empty-state-icon">📭</span>
          <span className="empty-state-text">No logs available</span>
        </div>
      ) : (
        <div
          ref={containerRef}
          className={`log-container ${wrapLines ? 'wrap-lines' : ''}`}
          onScroll={handleScroll}
        >
          {/* Virtual scrolling: top spacer (non-wrap mode only) */}
          {!wrapLines && (
            <div style={{ height: `${renderedLines.visibleStart * LINE_HEIGHT}px` }} />
          )}

          {/* Log lines */}
          {renderedLines.lines}

          {/* Virtual scrolling: bottom spacer (non-wrap mode only) */}
          {!wrapLines && (
            <div
              style={{
                height: `${(limitedLogs.length - renderedLines.visibleEnd) * LINE_HEIGHT}px`,
              }}
            />
          )}

          <div ref={logsEndRef} />
        </div>
      )}
    </div>
  );
};

export default LogsTab;
