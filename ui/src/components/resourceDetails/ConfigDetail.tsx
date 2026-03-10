/**
 * ConfigMap/Secret 详情页面组件
 * 专为配置数据设计，展示键值对、数据引用等信息
 */
import React, { useState } from 'react';
import { StatusCards } from '../StatusCards';
import { LabelList } from '../LabelList';
import { EventTimeline } from '../EventTimeline';
import { ResourceReferences } from '../ResourceReferences';
import './ResourceDetail.css';

interface ConfigDetailProps {
  data: any;
  isSecret?: boolean;
  events?: any[];
  onResourceClick?: (type: string, namespace: string, name: string) => void;
}

export const ConfigDetail: React.FC<ConfigDetailProps> = ({
  data,
  isSecret = false,
  events = [],
  onResourceClick,
}) => {
  if (!data) return null;

  const { metadata, data: configData, binaryData } = data;
  const [revealedKeys, setRevealedKeys] = useState<Set<string>>(new Set());
  const [copiedKey, setCopiedKey] = useState<string | null>(null);

  // 切换显示/隐藏（用于 Secret）
  const toggleReveal = (key: string) => {
    setRevealedKeys((prev) => {
      const next = new Set(prev);
      if (next.has(key)) {
        next.delete(key);
      } else {
        next.add(key);
      }
      return next;
    });
  };

  // 复制值
  const copyValue = async (key: string, value: string) => {
    try {
      await navigator.clipboard.writeText(value);
      setCopiedKey(key);
      setTimeout(() => setCopiedKey(null), 2000);
    } catch (err) {
      console.error('复制失败:', err);
    }
  };

  // 解码 Secret 数据
  const decodeValue = (value: string) => {
    try {
      return atob(value);
    } catch {
      return value;
    }
  };

  // 获取显示值
  const getDisplayValue = (key: string, value: string) => {
    if (isSecret && !revealedKeys.has(key)) {
      return '••••••••';
    }
    return isSecret ? decodeValue(value) : value;
  };

  // 状态卡片
  const statusCards = [
    {
      title: '类型',
      value: isSecret ? 'Secret' : 'ConfigMap',
      sub: isSecret ? (data.type || 'Opaque') : '-',
    },
    {
      title: '数据项',
      value: Object.keys(configData || {}).length,
      sub: `二进制：${Object.keys(binaryData || {}).length}`,
    },
    {
      title: '命名空间',
      value: metadata.namespace || 'default',
      sub: `创建：${metadata.creationTimestamp ? new Date(metadata.creationTimestamp).toLocaleDateString() : '-'}`,
    },
  ];

  // 渲染数据项
  const renderDataItems = (items: Record<string, string>, title: string) => {
    if (!items || Object.keys(items).length === 0) return null;

    return (
      <div className="data-section">
        <h4 className="subsection-title">{title}</h4>
        <div className="data-list">
          {Object.entries(items).map(([key, value]: [string, any]) => {
            const displayValue = getDisplayValue(key, String(value));
            const isBinary = !isSecret && typeof value !== 'string';
            const isLong = displayValue.length > 100;

            return (
              <div key={key} className="data-item">
                <div className="data-header">
                  <span className="data-key">{key}</span>
                  <div className="data-actions">
                    {isSecret && (
                      <button
                        className="btn-icon"
                        onClick={() => toggleReveal(key)}
                        title={revealedKeys.has(key) ? '隐藏' : '显示'}
                      >
                        {revealedKeys.has(key) ? '🙈' : '👁️'}
                      </button>
                    )}
                    <button
                      className="btn-icon"
                      onClick={() => copyValue(key, displayValue)}
                      title="复制"
                    >
                      {copiedKey === key ? '✅' : '📋'}
                    </button>
                  </div>
                </div>
                <div className={`data-value ${isLong ? 'data-value-long' : ''}`}>
                  {isBinary ? (
                    <span className="binary-indicator">
                      📦 二进制数据 ({value.length} 字节)
                    </span>
                  ) : (
                    <pre className="code-block">
                      {displayValue}
                    </pre>
                  )}
                </div>
              </div>
            );
          })}
        </div>
      </div>
    );
  };

  // 渲染数据引用
  const renderReferences = () => {
    return (
      <ResourceReferences
        name={data.metadata?.name}
        namespace={data.metadata?.namespace}
        type={isSecret ? 'secret' : 'configmap'}
        onResourceClick={onResourceClick}
      />
    );
  };

  return (
    <div className="config-detail">
      {/* 状态概览 */}
      <div className="detail-section">
        <h3 className="section-title">📊 状态概览</h3>
        <StatusCards cards={statusCards} />

        {/* 基本信息 */}
        <div className="info-card" style={{ marginTop: '16px' }}>
          <h4 className="card-title">📋 基本信息</h4>
          <div className="info-grid">
            <div className="info-item">
              <span className="info-label">名称</span>
              <span className="info-value">{metadata.name}</span>
            </div>
            <div className="info-item">
              <span className="info-label">命名空间</span>
              <span className="info-value">{metadata.namespace}</span>
            </div>
            <div className="info-item">
              <span className="info-label">创建时间</span>
              <span className="info-value">
                {metadata.creationTimestamp
                  ? new Date(metadata.creationTimestamp).toLocaleString()
                  : '-'}
              </span>
            </div>
            <div className="info-item">
              <span className="info-label">UID</span>
              <span className="info-value">{metadata.uid}</span>
            </div>
          </div>
        </div>
      </div>

      {/* 数据内容 */}
      <div className="detail-section">
        <h3 className="section-title">
          {isSecret ? '🔐 Secret 数据' : '📄 ConfigMap 数据'}
        </h3>

        {/* 字符串数据 */}
        {configData && Object.keys(configData).length > 0 && (
          <div className="info-card">
            <h4 className="card-title">
              {isSecret ? '📝 键值对' : '📝 数据'}
            </h4>
            {renderDataItems(configData, '')}
          </div>
        )}

        {/* 二进制数据 */}
        {binaryData && Object.keys(binaryData).length > 0 && (
          <div className="info-card">
            <h4 className="card-title">📦 二进制数据</h4>
            {renderDataItems(binaryData, '')}
          </div>
        )}

        {/* 空数据提示 */}
        {(!configData || Object.keys(configData).length === 0) &&
          (!binaryData || Object.keys(binaryData).length === 0) && (
            <div className="empty-state">
              <span className="empty-icon">📭</span>
              <p>没有数据</p>
            </div>
          )}
      </div>

      {/* 数据引用 */}
      <div className="detail-section">
        <h3 className="section-title">🔗 被引用情况</h3>
        {renderReferences()}
      </div>

      {/* 标签和注解 */}
      <div className="detail-section">
        <h3 className="section-title">🏷️ 标签和注解</h3>
        {metadata?.labels && Object.keys(metadata.labels).length > 0 && (
          <div className="labels-subsection">
            <h4 className="subsection-title">标签</h4>
            <LabelList labels={metadata.labels} />
          </div>
        )}
        {metadata?.annotations && Object.keys(metadata.annotations).length > 0 && (
          <div className="labels-subsection">
            <h4 className="subsection-title">注解</h4>
            <LabelList labels={metadata.annotations} />
          </div>
        )}
      </div>

      {/* 事件 */}
      {events.length > 0 && (
        <div className="detail-section">
          <h3 className="section-title">📅 事件</h3>
          <EventTimeline events={events} />
        </div>
      )}
    </div>
  );
};

export default ConfigDetail;
