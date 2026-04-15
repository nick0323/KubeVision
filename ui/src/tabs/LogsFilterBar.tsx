import React from 'react';
import { FaDownload, FaEraser, FaCog } from 'react-icons/fa';
import NamespaceSelect from '../common/NamespaceSelect';
import './LogsFilterBar.css';

export interface LogsFilterBarProps {
  namespace: string;
  containers: Array<{ name: string }>;
  selectedContainer: string;
  onContainerChange: (container: string) => void;
  tailLines: string;
  onTailLinesChange: (lines: string) => void;
  timestamps: boolean;
  onTimestampsChange: (value: boolean) => void;
  previous: boolean;
  onPreviousChange: (value: boolean) => void;
  wrapLines: boolean;
  onWrapLinesChange: (value: boolean) => void;
  showSettings: boolean;
  onToggleSettings: () => void;
  searchTerm: string;
  onSearchChange: (value: string) => void;
  onSearch: () => void;
  onClear: () => void;
  onDownload: () => void;
  connected: boolean;
  logCount: number;
  totalLogs: number;
}

const LINES_OPTIONS = [
  { value: '100', label: '100' },
  { value: '200', label: '200' },
  { value: '500', label: '500' },
];

/**
 * LogsFilterBar 日志过滤工具栏
 * 包含：容器选择、Tail Lines、时间戳、设置、搜索、操作按钮
 */
export const LogsFilterBar: React.FC<LogsFilterBarProps> = ({
  namespace,
  containers,
  selectedContainer,
  onContainerChange,
  tailLines,
  onTailLinesChange,
  timestamps,
  onTimestampsChange,
  previous,
  onPreviousChange,
  wrapLines,
  onWrapLinesChange,
  showSettings,
  onToggleSettings,
  searchTerm,
  onSearchChange,
  onSearch,
  onClear,
  onDownload,
  connected,
  logCount,
  totalLogs,
}) => {
  return (
    <div className="logs-filter-bar">
      <div className="logs-filter-bar-left">
        <span className="logs-title">Logs</span>
        <span className="logs-total">
          {logCount} / {totalLogs}
        </span>

        {containers.length > 1 && (
          <NamespaceSelect
            value={selectedContainer}
            onChange={onContainerChange}
            placeholder=""
            options={containers.map(c => c.name)}
            className="logs-container-select"
            width="120px"
          />
        )}
      </div>

      <div className="logs-filter-bar-right">
        <div className="search-input-wrapper">
          <input
            type="text"
            className="search-input"
            placeholder="Search logs..."
            value={searchTerm}
            onChange={e => onSearchChange(e.target.value)}
            onKeyDown={e => e.key === 'Enter' && onSearch()}
          />
          <button className="search-btn" onClick={onSearch}>
            Search
          </button>
          {searchTerm && (
            <button className="search-btn clear-btn" onClick={() => onSearchChange('')}>
              <FaEraser />
            </button>
          )}
        </div>

        <button
          className={`settings-btn ${showSettings ? 'active' : ''}`}
          onClick={onToggleSettings}
          title="Settings"
        >
          <FaCog />
        </button>

        {showSettings && (
          <div className="settings-panel">
            <div className="setting-item">
              <label>Initial Lines</label>
              <select
                value={tailLines}
                onChange={e => onTailLinesChange(e.target.value)}
                className="lines-select"
              >
                {LINES_OPTIONS.map(opt => (
                  <option key={opt.value} value={opt.value}>
                    {opt.label}
                  </option>
                ))}
              </select>
            </div>
            <div className="setting-item">
              <label>Show Timestamps</label>
              <button
                className={`toggle-btn ${timestamps ? 'active' : ''}`}
                onClick={() => onTimestampsChange(!timestamps)}
              />
            </div>
            <div className="setting-item">
              <label>Previous Container</label>
              <button
                className={`toggle-btn ${previous ? 'active' : ''}`}
                onClick={() => onPreviousChange(!previous)}
              />
            </div>
            <div className="setting-item">
              <label>Word Wrap</label>
              <button
                className={`toggle-btn ${wrapLines ? 'active' : ''}`}
                onClick={() => onWrapLinesChange(!wrapLines)}
              />
            </div>
          </div>
        )}

        <div className="filter-actions">
          <button className="action-btn" onClick={onDownload} title="Download logs">
            <FaDownload />
          </button>
          <button className="action-btn" onClick={onClear} title="Clear logs">
            <FaEraser />
          </button>
        </div>

        <div className={`filter-status ${connected ? 'connected' : 'disconnected'}`}>
          <span className="status-dot"></span>
          <span>{connected ? 'Live' : 'Disconnected'}</span>
        </div>
      </div>
    </div>
  );
};
