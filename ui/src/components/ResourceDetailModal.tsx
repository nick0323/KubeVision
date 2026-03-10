import React, { useState, useEffect } from 'react';
import { FaTimes, FaSync } from 'react-icons/fa';
import './ResourceDetailModal.css';
import { apiClient } from '../utils/apiClient';

interface ResourceDetailModalProps {
  resourceType: string;
  namespace: string;
  name: string;
  visible: boolean;
  onClose: () => void;
}

interface ResourceData {
  apiVersion?: string;
  kind?: string;
  metadata?: {
    name: string;
    namespace: string;
    uid: string;
    creationTimestamp: string;
    labels?: Record<string, string>;
    annotations?: Record<string, string>;
  };
  spec?: any;
  status?: any;
  [key: string]: any;
}

/**
 * 资源详情模态框组件
 */
export const ResourceDetailModal: React.FC<ResourceDetailModalProps> = ({
  resourceType,
  namespace,
  name,
  visible,
  onClose,
}) => {
  const [detailData, setDetailData] = useState<ResourceData | null>(null);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  // 加载详情数据
  const loadDetail = async () => {
    setLoading(true);
    setError(null);
    
    try {
      const data = await apiClient.getDetail(resourceType, namespace, name);
      setDetailData(data);
    } catch (err: any) {
      setError(err.message || '加载失败');
    } finally {
      setLoading(false);
    }
  };

  // 组件挂载或依赖变化时加载数据
  useEffect(() => {
    if (visible) {
      loadDetail();
    }
  }, [visible, resourceType, namespace, name]);

  // 按 ESC 关闭
  useEffect(() => {
    const handleEsc = (e: KeyboardEvent) => {
      if (e.key === 'Escape') {
        onClose();
      }
    };
    
    if (visible) {
      window.addEventListener('keydown', handleEsc);
    }
    
    return () => {
      window.removeEventListener('keydown', handleEsc);
    };
  }, [visible, onClose]);

  if (!visible) {
    return null;
  }

  return (
    <div className="resource-detail-modal" onClick={onClose}>
      <div className="modal-content" onClick={(e) => e.stopPropagation()}>
        {/* 头部 */}
        <div className="modal-header">
          <div className="modal-title-section">
            <h2 className="modal-title">{name}</h2>
            <span className="modal-subtitle">
              {resourceType} {namespace && `/ ${namespace}`}
            </span>
          </div>
          <div className="modal-actions">
            <button
              className="modal-refresh"
              onClick={loadDetail}
              title="刷新"
              disabled={loading}
            >
              <FaSync className={loading ? 'spinning' : ''} />
            </button>
            <button
              className="modal-close"
              onClick={onClose}
              title="关闭"
            >
              <FaTimes />
            </button>
          </div>
        </div>

        {/* 内容区 */}
        <div className="detail-content">
          {loading && (
            <div className="detail-loading">
              <div className="loading-spinner">加载中...</div>
            </div>
          )}

          {error && (
            <div className="detail-error">
              <h3>加载失败</h3>
              <p>{error}</p>
              <button onClick={loadDetail}>重试</button>
            </div>
          )}

          {detailData && (
            <>
              {/* Metadata 卡片 */}
              <div className="detail-card">
                <h3 className="detail-card-title">Metadata</h3>
                <div className="detail-card-content">
                  {detailData.apiVersion && (
                    <div className="detail-item">
                      <span className="detail-label">API Version</span>
                      <span className="detail-value">{detailData.apiVersion}</span>
                    </div>
                  )}
                  {detailData.kind && (
                    <div className="detail-item">
                      <span className="detail-label">Kind</span>
                      <span className="detail-value">{detailData.kind}</span>
                    </div>
                  )}
                  {detailData.metadata && (
                    <>
                      <div className="detail-item">
                        <span className="detail-label">Name</span>
                        <span className="detail-value">{detailData.metadata.name}</span>
                      </div>
                      {detailData.metadata.namespace && (
                        <div className="detail-item">
                          <span className="detail-label">Namespace</span>
                          <span className="detail-value">{detailData.metadata.namespace}</span>
                        </div>
                      )}
                      <div className="detail-item">
                        <span className="detail-label">UID</span>
                        <span className="detail-value">{detailData.metadata.uid}</span>
                      </div>
                      <div className="detail-item">
                        <span className="detail-label">Created</span>
                        <span className="detail-value">
                          {new Date(detailData.metadata.creationTimestamp).toLocaleString()}
                        </span>
                      </div>
                      {detailData.metadata.labels && Object.keys(detailData.metadata.labels).length > 0 && (
                        <div className="detail-item">
                          <span className="detail-label">Labels</span>
                          <div className="detail-value detail-labels">
                            {Object.entries(detailData.metadata.labels).map(([key, value]) => (
                              <div key={key} className="label-item">
                                <span className="label-key">{key}</span>
                                <span className="label-sep">=</span>
                                <span className="label-value">{value}</span>
                              </div>
                            ))}
                          </div>
                        </div>
                      )}
                      {detailData.metadata.annotations && Object.keys(detailData.metadata.annotations).length > 0 && (
                        <div className="detail-item">
                          <span className="detail-label">Annotations</span>
                          <div className="detail-value detail-annotations">
                            {Object.entries(detailData.metadata.annotations).map(([key, value]) => (
                              <div key={key} className="annotation-item">
                                <span className="annotation-key">{key}</span>
                                <span className="annotation-sep">=</span>
                                <span className="annotation-value">{value}</span>
                              </div>
                            ))}
                          </div>
                        </div>
                      )}
                    </>
                  )}
                </div>
              </div>

              {/* Spec 卡片 */}
              {detailData.spec && (
                <div className="detail-card">
                  <h3 className="detail-card-title">Spec</h3>
                  <div className="detail-card-content">
                    <pre className="detail-pre">{JSON.stringify(detailData.spec, null, 2)}</pre>
                  </div>
                </div>
              )}

              {/* Status 卡片 */}
              {detailData.status && (
                <div className="detail-card">
                  <h3 className="detail-card-title">Status</h3>
                  <div className="detail-card-content">
                    <pre className="detail-pre">{JSON.stringify(detailData.status, null, 2)}</pre>
                  </div>
                </div>
              )}

              {/* 其他字段 */}
              {Object.keys(detailData).filter(key => 
                !['apiVersion', 'kind', 'metadata', 'spec', 'status'].includes(key)
              ).map(key => (
                <div key={key} className="detail-card">
                  <h3 className="detail-card-title">{key}</h3>
                  <div className="detail-card-content">
                    <pre className="detail-pre">{JSON.stringify(detailData[key], null, 2)}</pre>
                  </div>
                </div>
              ))}
            </>
          )}
        </div>
      </div>
    </div>
  );
};

export default ResourceDetailModal;
