/**
 * 容器日志查看器组件
 * 支持实时日志、关键字过滤、自动滚动等功能
 */
import React, { useState, useEffect, useRef, useCallback } from 'react';
import { authFetch } from '../../utils/auth';
import './PodLogViewer.css';

interface PodLogViewerProps {
  podName: string;
  namespace: string;
  containerName?: string;
  onClose: () => void;
}

export const PodLogViewer: React.FC<PodLogViewerProps> = ({
  podName,
  namespace,
  containerName,
  onClose,
}) => {
  const [logs, setLogs] = useState<string>('');
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [isStreaming, setIsStreaming] = useState(false);
  const [autoScroll, setAutoScroll] = useState(true);
  const [searchTerm, setSearchTerm] = useState('');
  const [showTimestamps, setShowTimestamps] = useState(false);
  const [tailLines, setTailLines] = useState(100);
  
  const logEndRef = useRef<HTMLDivElement>(null);
  const logContainerRef = useRef<HTMLDivElement>(null);
  const eventSourceRef = useRef<EventSource | null>(null);

  // 滚动到底部
  const scrollToBottom = useCallback(() => {
    if (autoScroll && logEndRef.current) {
      logEndRef.current.scrollIntoView({ behavior: 'smooth' });
    }
  }, [autoScroll]);

  // 加载日志
  const loadLogs = useCallback(async () => {
    setLoading(true);
    setError(null);

    try {
      const params = new URLSearchParams({
        tailLines: tailLines.toString(),
        timestamps: showTimestamps.toString(),
      });
      
      if (containerName) {
        params.set('container', containerName);
      }

      const response = await authFetch(
        `/api/pods/${namespace}/${podName}/logs?${params}`
      );

      if (!response.ok) {
        throw new Error(`HTTP ${response.status}: ${response.statusText}`);
      }

      const data = await response.json();
      if (data.code === 0) {
        setLogs(data.data || '');
      } else {
        throw new Error(data.message || '加载日志失败');
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : '加载日志失败');
    } finally {
      setLoading(false);
    }
  }, [namespace, podName, containerName, tailLines, showTimestamps]);

  // 流式日志
  const startStreaming = useCallback(() => {
    const params = new URLSearchParams({
      follow: 'true',
      tailLines: '100',
      timestamps: showTimestamps.toString(),
    });
    
    if (containerName) {
      params.set('container', containerName);
    }

    // 使用 EventSource 或 fetch stream
    const url = `/api/pods/${namespace}/${podName}/logs/stream?${params}`;
    
    // 简单实现：定期轮询
    setIsStreaming(true);
    
    const pollLogs = async () => {
      try {
        const response = await authFetch(
          `/api/pods/${namespace}/${podName}/logs?tailLines=100&timestamps=${showTimestamps}`
        );
        const data = await response.json();
        if (data.code === 0) {
          setLogs(prev => {
            if (data.data !== prev) {
              return data.data;
            }
            return prev;
          });
        }
      } catch (err) {
        console.error('流式日志错误:', err);
      }
    };

    // 立即执行一次
    pollLogs();
    
    // 定期轮询
    const interval = setInterval(pollLogs, 2000);
    eventSourceRef.current = { close: () => clearInterval(interval) } as any;
  }, [namespace, podName, containerName, showTimestamps]);

  // 停止流式
  const stopStreaming = useCallback(() => {
    if (eventSourceRef.current) {
      eventSourceRef.current.close();
      eventSourceRef.current = null;
    }
    setIsStreaming(false);
  }, []);

  // 初始加载
  useEffect(() => {
    loadLogs();
    return () => stopStreaming();
  }, [loadLogs, stopStreaming]);

  // 自动滚动
  useEffect(() => {
    scrollToBottom();
  }, [logs, scrollToBottom]);

  // 过滤日志
  const filteredLogs = searchTerm
    ? logs.split('\n').filter(line => 
        line.toLowerCase().includes(searchTerm.toLowerCase())
      ).join('\n')
    : logs;

  // 清空日志
  const handleClear = () => {
    setLogs('');
  };

  // 下载日志
  const handleDownload = () => {
    const blob = new Blob([logs], { type: 'text/plain' });
    const url = URL.createObjectURL(blob);
    const a = document.createElement('a');
    a.href = url;
    a.download = `${podName}-${containerName || 'all'}-logs.txt`;
    a.click();
    URL.revokeObjectURL(url);
  };

  // 复制日志
  const handleCopy = async () => {
    try {
      await navigator.clipboard.writeText(logs);
    } catch (err) {
      console.error('复制失败:', err);
    }
  };

  return (
    <div className="pod-log-viewer">
      {/* 头部工具栏 */}
      <div className="log-toolbar">
        <div className="toolbar-left">
          <h3 className="log-title">
            📜 {podName} {containerName && `/${containerName}`}
          </h3>
        </div>
        
        <div className="toolbar-center">
          {/* 搜索框 */}
          <input
            type="text"
            className="log-search"
            placeholder="搜索日志..."
            value={searchTerm}
            onChange={(e) => setSearchTerm(e.target.value)}
          />
          
          {/* 行数选择 */}
          <select
            className="log-tail-select"
            value={tailLines}
            onChange={(e) => setTailLines(Number(e.target.value))}
            disabled={isStreaming}
          >
            <option value={50}>最近 50 行</option>
            <option value={100}>最近 100 行</option>
            <option value={500}>最近 500 行</option>
            <option value={1000}>最近 1000 行</option>
            <option value={0}>全部</option>
          </select>
        </div>

        <div className="toolbar-right">
          {/* 时间戳开关 */}
          <label className="log-toggle">
            <input
              type="checkbox"
              checked={showTimestamps}
              onChange={(e) => setShowTimestamps(e.target.checked)}
              disabled={isStreaming}
            />
            时间戳
          </label>

          {/* 自动滚动开关 */}
          <label className="log-toggle">
            <input
              type="checkbox"
              checked={autoScroll}
              onChange={(e) => setAutoScroll(e.target.checked)}
            />
            自动滚动
          </label>

          {/* 流式按钮 */}
          <button
            className={`btn btn-stream ${isStreaming ? 'active' : ''}`}
            onClick={() => isStreaming ? stopStreaming() : startStreaming()}
          >
            {isStreaming ? '⏸️ 暂停' : '▶️ 实时'}
          </button>

          {/* 刷新按钮 */}
          <button className="btn btn-refresh" onClick={loadLogs}>
            🔄 刷新
          </button>

          {/* 清空按钮 */}
          <button className="btn btn-clear" onClick={handleClear}>
            🗑️ 清空
          </button>

          {/* 复制按钮 */}
          <button className="btn btn-copy" onClick={handleCopy}>
            📋 复制
          </button>

          {/* 下载按钮 */}
          <button className="btn btn-download" onClick={handleDownload}>
            ⬇️ 下载
          </button>

          {/* 关闭按钮 */}
          <button className="btn btn-close" onClick={onClose}>
            ✕
          </button>
        </div>
      </div>

      {/* 日志内容 */}
      <div className="log-container" ref={logContainerRef}>
        {loading ? (
          <div className="log-loading">
            <span className="loading-spinner">加载中...</span>
          </div>
        ) : error ? (
          <div className="log-error">
            <span className="error-icon">⚠️</span>
            <p>{error}</p>
            <button className="btn btn-retry" onClick={loadLogs}>
              重试
            </button>
          </div>
        ) : (
          <pre className="log-content">
            {filteredLogs || <span className="log-empty">暂无日志</span>}
            <div ref={logEndRef} />
          </pre>
        )}
      </div>

      {/* 状态栏 */}
      <div className="log-statusbar">
        <span className="status-item">
          {logs.split('\n').filter(l => l).length} 行日志
        </span>
        {searchTerm && (
          <span className="status-item">
            过滤："{searchTerm}"
          </span>
        )}
        {isStreaming && (
          <span className="status-item streaming">
            🔴 实时更新中
          </span>
        )}
      </div>
    </div>
  );
};

export default PodLogViewer;
