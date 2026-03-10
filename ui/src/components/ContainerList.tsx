import React from 'react';
import './ContainerList.css';

interface Container {
  name: string;
  image: string;
  ports?: Array<{ containerPort: number; protocol?: string }>;
  resources?: {
    requests?: { cpu?: string; memory?: string };
    limits?: { cpu?: string; memory?: string };
  };
  livenessProbe?: any;
  readinessProbe?: any;
}

interface ContainerListProps {
  containers: Container[];
  onViewLog?: (containerName: string) => void;
  onExec?: (containerName: string) => void;
}

export const ContainerList: React.FC<ContainerListProps> = ({
  containers,
  onViewLog,
  onExec,
}) => {
  return (
    <div className="container-list">
      {containers.map((container, index) => (
        <div key={index} className="container-card">
          <div className="container-header">
            <div>
              <div className="container-name">
                <span className="status-dot running">🟢</span>
                {container.name}
                <span className="container-image">{container.image}</span>
              </div>
            </div>
            <div className="action-buttons">
              {onViewLog && (
                <button className="btn btn-default" onClick={() => onViewLog(container.name)}>
                  查看日志
                </button>
              )}
              {onExec && (
                <button className="btn btn-default" onClick={() => onExec(container.name)}>
                  终端
                </button>
              )}
            </div>
          </div>

          <div className="container-grid">
            <div className="container-stat">
              <div className="container-stat-label">端口</div>
              <div className="container-stat-value">
                {container.ports?.length ? (
                  container.ports.map((port, i) => (
                    <span key={i}>
                      {port.containerPort}/{port.protocol || 'TCP'}
                      {i < container.ports.length - 1 ? ', ' : ''}
                    </span>
                  ))
                ) : (
                  '-'
                )}
              </div>
            </div>
            <div className="container-stat">
              <div className="container-stat-label">CPU 请求</div>
              <div className="container-stat-value">
                {container.resources?.requests?.cpu || '-'}
              </div>
            </div>
            <div className="container-stat">
              <div className="container-stat-label">CPU 限制</div>
              <div className="container-stat-value">
                {container.resources?.limits?.cpu || '-'}
              </div>
            </div>
            <div className="container-stat">
              <div className="container-stat-label">内存请求</div>
              <div className="container-stat-value">
                {container.resources?.requests?.memory || '-'}
              </div>
            </div>
            <div className="container-stat">
              <div className="container-stat-label">内存限制</div>
              <div className="container-stat-value">
                {container.resources?.limits?.memory || '-'}
              </div>
            </div>
            <div className="container-stat">
              <div className="container-stat-label">探针</div>
              <div className="container-stat-value">
                {container.livenessProbe ? '🟢 存活' : ''}{' '}
                {container.readinessProbe ? '就绪' : ''}
              </div>
            </div>
          </div>
        </div>
      ))}
    </div>
  );
};

export default ContainerList;
