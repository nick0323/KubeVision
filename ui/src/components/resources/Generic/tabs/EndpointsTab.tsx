import React, { useState, useCallback, useEffect } from 'react';
import { LoadingSpinner } from '../../../ui/LoadingSpinner';
import { ErrorDisplay } from '../../../ui/ErrorDisplay';
import { authFetch } from '../../../../utils/auth';
import './EndpointsTab.css';

interface EndpointSubset {
  addresses?: { ip: string }[];
  ports?: { port: number; protocol: string }[];
}

interface Endpoints {
  metadata: {
    name: string;
    namespace: string;
  };
  subsets: EndpointSubset[];
}

interface EndpointsTabProps {
  namespace: string;
  serviceName: string;
}

/**
 * Endpoints Tab - 显示 Service 的 Endpoints
 */
export const EndpointsTab: React.FC<EndpointsTabProps> = ({ namespace, serviceName }) => {
  const [endpoints, setEndpoints] = useState<Endpoints | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  // 加载 Endpoints
  const loadEndpoints = useCallback(async () => {
    setLoading(true);
    setError(null);

    try {
      const response = await authFetch(`/api/endpoint/${namespace}/${serviceName}`);
      const result = await response.json();

      if (result.code === 0 && result.data) {
        setEndpoints(result.data);
      } else {
        setError(result.message || '加载失败');
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : '网络错误');
    } finally {
      setLoading(false);
    }
  }, [namespace, serviceName]);

  useEffect(() => {
    loadEndpoints();
  }, [loadEndpoints]);

  if (loading) {
    return <LoadingSpinner text="加载 Endpoints..." size="lg" />;
  }

  if (error) {
    return <ErrorDisplay message={error} type="error" showRetry onRetry={loadEndpoints} />;
  }

  if (!endpoints) {
    return (
      <div className="endpoints-tab">
        <div className="empty-state">
          <span className="empty-state-text">暂无 Endpoints 信息</span>
        </div>
      </div>
    );
  }

  return (
    <div className="endpoints-tab">
      <div className="detail-card">
        <h3 className="detail-card-title">Endpoints</h3>
        <div className="detail-card-body">
          {endpoints.subsets && endpoints.subsets.length > 0 ? (
            endpoints.subsets.map((subset, index) => (
              <div key={index} className="endpoint-subset">
                <h4>Subset {index + 1}</h4>
                {subset.addresses && subset.addresses.length > 0 && (
                  <div className="endpoint-section">
                    <strong>Addresses:</strong>
                    <ul>
                      {subset.addresses.map((addr, i) => (
                        <li key={i}>{addr.ip}</li>
                      ))}
                    </ul>
                  </div>
                )}
                {subset.ports && subset.ports.length > 0 && (
                  <div className="endpoint-section">
                    <strong>Ports:</strong>
                    <ul>
                      {subset.ports.map((port, i) => (
                        <li key={i}>
                          {port.port}/{port.protocol}
                        </li>
                      ))}
                    </ul>
                  </div>
                )}
              </div>
            ))
          ) : (
            <div className="empty-state">
              <span className="empty-state-text">没有配置 Endpoints</span>
            </div>
          )}
        </div>
      </div>
    </div>
  );
};

export default EndpointsTab;
