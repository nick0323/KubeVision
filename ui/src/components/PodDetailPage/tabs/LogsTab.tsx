import React, { useState, useCallback, useRef, useEffect, useMemo } from 'react';
import { LogsTabProps } from '../types';
import { LoadingSpinner } from '../../LoadingSpinner';
import { ErrorDisplay } from '../../ErrorDisplay';
import { authFetch } from '../../../utils/auth';
import './LogsTab.css';

const TIME_OPTIONS = [
  { value: '5m', label: '5m' },
  { value: '15m', label: '15m' },
  { value: '30m', label: '30m' },
  { value: '1h', label: '1h' },
  { value: '4h', label: '4h' },
  { value: '1d', label: '1d' },
];

// Performance: max log lines to display
const MAX_LOG_LINES = 1000;
// Virtual scrolling: line height in pixels
const LINE_HEIGHT = 23;
// Virtual scrolling: overscan rows
const OVERSCAN_ROWS = 20;

/**
 * Remove ANSI escape sequences from string
 */
const stripAnsiCodes = (str: string): string => {
  return str.replace(/\x1b\[[0-9;]*m/g, '');
};

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

  // 过滤选项
  const [selectedContainer, setSelectedContainer] = useState<string>('');
  const [since, setSince] = useState<string>('30m');
  const [previous, setPrevious] = useState(false);
  const [timestamps, setTimestamps] = useState(false);
  const [wrapLines, setWrapLines] = useState(true);

  // 搜索
  const [searchTerm, setSearchTerm] = useState('');
  const [currentSearchIndex, setCurrentSearchIndex] = useState(0);

  // 虚拟滚动
  const [scrollTop, setScrollTop] = useState(0);
  const [containerHeight, setContainerHeight] = useState(500);

  const logsEndRef = useRef<HTMLDivElement>(null);
  const searchInputRef = useRef<HTMLInputElement>(null);
  const containerRef = useRef<HTMLDivElement>(null);

  // Set default container
  useEffect(() => {
    if (containers.length > 0 && !selectedContainer) {
      setSelectedContainer(containers[0].name);
    }
  }, [containers, selectedContainer]);

  // Limit log lines
  const limitedLogs = useMemo(() => {
    if (logs.length <= MAX_LOG_LINES) return logs;
    return logs.slice(logs.length - MAX_LOG_LINES);
  }, [logs]);

  const offset = useMemo(() => {
    return logs.length > MAX_LOG_LINES ? logs.length - MAX_LOG_LINES : 0;
  }, [logs]);

  // Load logs
  const loadLogs = useCallback(async () => {
    if (!selectedContainer) return;

    setLoading(true);
    setError(null);

    try {
      const params = new URLSearchParams({
        since,
        previous: previous.toString(),
        timestamps: timestamps.toString(),
      });

      const response = await authFetch(
        `/api/pods/${namespace}/${name}/logs?container=${selectedContainer}&${params}`
      );
      const result = await response.json();

      if (result.code === 0 && result.data) {
        const logLines = typeof result.data === 'string'
          ? result.data.split('\n')
          : result.data;

        // Filter ANSI escape sequences
        const cleanedLogs = logLines.map(line => stripAnsiCodes(line));
        setLogs(cleanedLogs);
      } else {
        setError(result.message || 'Failed to load logs');
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Load failed');
    } finally {
      setLoading(false);
    }
  }, [namespace, name, selectedContainer, since, previous, timestamps]);

  // Load logs when filter options change
  useEffect(() => {
    if (selectedContainer) {
      loadLogs();
    }
  }, [selectedContainer, since, previous, timestamps, loadLogs]);

  // Scroll to bottom on initial load only
  useEffect(() => {
    if (logs.length > 0 && containerRef.current) {
      containerRef.current.scrollTop = containerRef.current.scrollHeight;
    }
  }, []);

  // Handle container scroll
  const handleScroll = useCallback(() => {
    if (containerRef.current) {
      setScrollTop(containerRef.current.scrollTop);
      setContainerHeight(containerRef.current.clientHeight);
    }
  }, []);

  // Search logs - memoized (real-time search as you type)
  const searchResults = useMemo(() => {
    if (!searchTerm || limitedLogs.length === 0) return [];

    const results: number[] = [];
    limitedLogs.forEach((line, index) => {
      if (line.toLowerCase().includes(searchTerm.toLowerCase())) {
        results.push(index);
      }
    });
    return results;
  }, [searchTerm, limitedLogs]);

  // Reset search index when search results change
  useEffect(() => {
    setCurrentSearchIndex(0);
  }, [searchResults.length]);

  // Next match
  const handleNext = useCallback(() => {
    if (searchResults.length === 0) return;
    const nextIndex = (currentSearchIndex + 1) % searchResults.length;
    setCurrentSearchIndex(nextIndex);
    const element = document.getElementById(`log-line-${searchResults[nextIndex]}`);
    element?.scrollIntoView({ behavior: 'smooth', block: 'center' });
  }, [searchResults, currentSearchIndex]);

  // Copy logs
  const handleCopy = useCallback(() => {
    navigator.clipboard.writeText(limitedLogs.join('\n'));
    alert('Logs copied to clipboard');
  }, [limitedLogs]);

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
      const searchLower = searchTerm.toLowerCase();
      const lines: React.ReactNode[] = limitedLogs.map((line, i) => {
        const actualIndex = i + offset;
        const isMatch = searchTerm && line.toLowerCase().includes(searchLower);
        const isInSearchResults = searchResults.includes(i);

        let className = 'log-line';
        if (isInSearchResults) className += ' highlight';
        if (line.toLowerCase().includes('error')) className += ' error';
        else if (line.toLowerCase().includes('warn')) className += ' warn';
        else if (line.toLowerCase().includes('info')) className += ' info';

        let content: React.ReactNode = line;
        if (isMatch && searchTerm) {
          const parts = line.split(new RegExp(`(${searchTerm})`, 'gi'));
          content = parts.map((part, j) =>
            part.toLowerCase() === searchLower
              ? <mark key={j}>{part}</mark>
              : part
          );
        }

        return (
          <div
            key={actualIndex}
            id={`log-line-${actualIndex}`}
            className={className}
          >
            {content}
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

    const searchLower = searchTerm.toLowerCase();

    const lines: React.ReactNode[] = [];
    for (let i = visibleStart; i < visibleEnd; i++) {
      const line = limitedLogs[i];
      const actualIndex = i + offset;
      const isMatch = searchTerm && line.toLowerCase().includes(searchLower);
      const isInSearchResults = searchResults.includes(i);

      // Determine class name
      let className = 'log-line';
      if (isInSearchResults) className += ' highlight';
      if (line.toLowerCase().includes('error')) className += ' error';
      else if (line.toLowerCase().includes('warn')) className += ' warn';
      else if (line.toLowerCase().includes('info')) className += ' info';

      // Highlight search term
      let content: React.ReactNode = line;
      if (isMatch && searchTerm) {
        const parts = line.split(new RegExp(`(${searchTerm})`, 'gi'));
        content = parts.map((part, j) =>
          part.toLowerCase() === searchLower
            ? <mark key={j}>{part}</mark>
            : part
        );
      }

      lines.push(
        <div
          key={actualIndex}
          id={`log-line-${actualIndex}`}
          className={className}
          style={wrapLines ? {} : { height: `${LINE_HEIGHT}px` }}
        >
          {content}
        </div>
      );
    }

    return { lines, visibleStart, visibleEnd };
  }, [limitedLogs, scrollTop, containerHeight, searchTerm, searchResults, offset, wrapLines]);

  return (
    <div className="logs-tab">
      {/* Search and Filter Bar */}
      <div className="filter-options">
        <div className="search-input-wrapper">
          <input
            ref={searchInputRef}
            type="text"
            className="search-input"
            placeholder="Search logs..."
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
          />
          {searchTerm && (
            <button className="search-btn clear-btn" onClick={() => setSearchTerm('')}>
              ✕
            </button>
          )}
          {searchResults.length > 0 && (
            <button className="search-btn" onClick={handleNext}>
              Next ({currentSearchIndex + 1}/{searchResults.length})
            </button>
          )}
        </div>

        {containers.length > 1 && (
          <div className="filter-option">
            <label>Container:</label>
            <select value={selectedContainer} onChange={(e) => setSelectedContainer(e.target.value)}>
              {containers.map((c) => (
                <option key={c.name} value={c.name}>{c.name}</option>
              ))}
            </select>
          </div>
        )}

        <div className="filter-option">
          <label>Time:</label>
          <select value={since} onChange={(e) => setSince(e.target.value)}>
            {TIME_OPTIONS.map((opt) => (
              <option key={opt.value} value={opt.value}>{opt.label}</option>
            ))}
          </select>
        </div>

        <div className="filter-option">
          <label>
            <input
              type="checkbox"
              checked={previous}
              onChange={(e) => setPrevious(e.target.checked)}
            />
            Previous
          </label>
        </div>

        <div className="filter-option">
          <label>
            <input
              type="checkbox"
              checked={timestamps}
              onChange={(e) => setTimestamps(e.target.checked)}
            />
            Timestamps
          </label>
        </div>

        <div className="filter-option">
          <label>
            <input
              type="checkbox"
              checked={wrapLines}
              onChange={(e) => setWrapLines(e.target.checked)}
            />
            Wrap Lines
          </label>
        </div>

        <div className="filter-actions">
          <button className="toolbar-btn" onClick={handleDownload}>Download</button>
          <button className="toolbar-btn" onClick={handleClear}>Clear</button>
          <span className="logs-count">
            {logs.length} lines {logs.length > MAX_LOG_LINES && `(last ${MAX_LOG_LINES})`}
          </span>
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
            <div style={{ height: `${(limitedLogs.length - renderedLines.visibleEnd) * LINE_HEIGHT}px` }} />
          )}

          <div ref={logsEndRef} />
        </div>
      )}
    </div>
  );
};

export default LogsTab;
