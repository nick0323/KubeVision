import React, { useState, useCallback, useRef, useEffect } from 'react';
import { LogsTabProps } from '../types';
import { LoadingSpinner } from '../../LoadingSpinner';
import { ErrorDisplay } from '../../ErrorDisplay';
import { authFetch } from '../../../utils/auth';
import './LogsTab.css';

const TIME_OPTIONS = [
  { value: '5m', label: '5 分钟' },
  { value: '15m', label: '15 分钟' },
  { value: '30m', label: '30 分钟' },
  { value: '1h', label: '1 小时' },
  { value: '4h', label: '4 小时' },
  { value: '1d', label: '1 天' },
];

/**
 * Logs Tab - 日志查看
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
  const [searchResults, setSearchResults] = useState<number[]>([]);
  const [currentSearchIndex, setCurrentSearchIndex] = useState(0);
  
  const logsEndRef = useRef<HTMLDivElement>(null);
  const searchInputRef = useRef<HTMLInputElement>(null);

  // 设置默认容器
  useEffect(() => {
    if (containers.length === 1) {
      setSelectedContainer(containers[0].name);
    }
  }, [containers]);

  // 加载日志
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
        
        setLogs(logLines);
      } else {
        setError(result.message || '加载日志失败');
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : '加载失败');
    } finally {
      setLoading(false);
    }
  }, [namespace, name, selectedContainer, since, previous, timestamps]);

  // 初始加载和过滤条件变化时加载日志
  useEffect(() => {
    if (selectedContainer) {
      loadLogs();
    }
  }, [selectedContainer, since, previous, timestamps, loadLogs]);

  // 滚动到底部
  useEffect(() => {
    logsEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  }, [logs]);

  // 搜索日志
  const handleSearch = useCallback(() => {
    if (!searchTerm) {
      setSearchResults([]);
      setCurrentSearchIndex(0);
      return;
    }
    
    const results: number[] = [];
    logs.forEach((line, index) => {
      if (line.toLowerCase().includes(searchTerm.toLowerCase())) {
        results.push(index);
      }
    });
    
    setSearchResults(results);
    setCurrentSearchIndex(0);
    
    // 滚动到第一个匹配项
    if (results.length > 0 && logsEndRef.current) {
      const element = document.getElementById(`log-line-${results[0]}`);
      element?.scrollIntoView({ behavior: 'smooth', block: 'center' });
    }
  }, [searchTerm, logs]);

  // 下一个匹配项
  const handleNext = useCallback(() => {
    if (searchResults.length === 0) return;
    
    const nextIndex = (currentSearchIndex + 1) % searchResults.length;
    setCurrentSearchIndex(nextIndex);
    
    const element = document.getElementById(`log-line-${searchResults[nextIndex]}`);
    element?.scrollIntoView({ behavior: 'smooth', block: 'center' });
  }, [searchResults, currentSearchIndex]);

  // 复制日志
  const handleCopy = useCallback(() => {
    navigator.clipboard.writeText(logs.join('\n'));
    alert('日志已复制');
  }, [logs]);

  // 下载日志
  const handleDownload = useCallback(() => {
    const blob = new Blob([logs.join('\n')], { type: 'text/plain' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = `${name}-${selectedContainer}-logs.txt`;
    a.click();
    URL.revokeObjectURL(url);
  }, [logs, name, selectedContainer]);

  // 清空日志
  const handleClear = useCallback(() => {
    setLogs([]);
  }, []);

  // 高亮搜索词
  const highlightSearch = (line: string, index: number) => {
    if (!searchTerm || !searchResults.includes(index)) {
      return line;
    }
    
    const parts = line.split(new RegExp(`(${searchTerm})`, 'gi'));
    return parts.map((part, i) => 
      part.toLowerCase() === searchTerm.toLowerCase() 
        ? <mark key={i}>{part}</mark> 
        : part
    );
  };

  // 获取日志行样式
  const getLineClass = (line: string, index: number) => {
    const classes = ['log-line'];
    
    if (searchResults.includes(index)) {
      classes.push('highlight');
    }
    
    if (line.toLowerCase().includes('error')) {
      classes.push('error');
    } else if (line.toLowerCase().includes('warn')) {
      classes.push('warn');
    } else if (line.toLowerCase().includes('info')) {
      classes.push('info');
    }
    
    return classes.join(' ');
  };

  return (
    <div className="logs-tab">
      {/* 搜索和过滤栏 */}
      <div className="search-bar">
        <div className="search-input-wrapper">
          <input
            ref={searchInputRef}
            type="text"
            className="search-input"
            placeholder="搜索日志..."
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
            onKeyPress={(e) => e.key === 'Enter' && handleSearch()}
          />
          <button className="search-btn" onClick={handleSearch}>
            搜索
          </button>
          {searchResults.length > 0 && (
            <button className="search-btn" onClick={handleNext}>
              下一个 ({currentSearchIndex + 1}/{searchResults.length})
            </button>
          )}
        </div>
        
        <div className="filter-options">
          {containers.length > 1 && (
            <div className="filter-option">
              <label>容器:</label>
              <select value={selectedContainer} onChange={(e) => setSelectedContainer(e.target.value)}>
                {containers.map((c) => (
                  <option key={c.name} value={c.name}>{c.name}</option>
                ))}
              </select>
            </div>
          )}
          
          <div className="filter-option">
            <label>时间:</label>
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
              时间戳
            </label>
          </div>
          
          <div className="filter-option">
            <label>
              <input
                type="checkbox"
                checked={wrapLines}
                onChange={(e) => setWrapLines(e.target.checked)}
              />
              自动换行
            </label>
          </div>
        </div>
      </div>

      {/* 操作按钮 */}
      <div className="logs-actions">
        <button className="toolbar-btn" onClick={handleCopy}>📋 复制</button>
        <button className="toolbar-btn" onClick={handleDownload}>⬇️ 下载</button>
        <button className="toolbar-btn" onClick={handleClear}>🗑️ 清空</button>
        <span className="logs-count">共 {logs.length} 行</span>
      </div>

      {/* 日志内容 */}
      {loading && logs.length === 0 ? (
        <LoadingSpinner text="加载日志..." size="lg" />
      ) : error ? (
        <ErrorDisplay message={error} type="error" showRetry onRetry={loadLogs} />
      ) : logs.length === 0 ? (
        <div className="empty-state">
          <span className="empty-state-icon">📭</span>
          <span className="empty-state-text">暂无日志</span>
        </div>
      ) : (
        <div className={`log-container ${wrapLines ? 'wrap-lines' : ''}`}>
          {logs.map((line, index) => (
            <div
              key={index}
              id={`log-line-${index}`}
              className={getLineClass(line, index)}
            >
              {highlightSearch(line, index)}
            </div>
          ))}
          <div ref={logsEndRef} />
        </div>
      )}
    </div>
  );
};

export default LogsTab;
