import React from 'react';

interface ContainerDetailsProps {
  container: any;
}

export const ContainerDetails: React.FC<ContainerDetailsProps> = ({ container }) => {
  return (
    <div className="container-card-body">
      <div className="sub-module">
        <div className="sub-module-title">Ports</div>
        <div className="info-grid">
          {container?.ports?.map((port: any, idx: number) => (
            <div key={idx} className="info-item">
              <span className="info-value">
                {port.name ? `${port.name}: ` : ''}
                {port.containerPort}/{port.protocol || 'TCP'}
              </span>
            </div>
          )) || (
            <div className="empty-state">No ports defined</div>
          )}
        </div>
      </div>

      <div className="sub-module">
        <div className="sub-module-title">Environment Variables</div>
        <div className="info-grid">
          {container?.env?.map((env: any, idx: number) => (
            <div key={idx} className="info-item">
              <span className="info-value">
                <span className="env-key">{env.name}</span>
                <span className="env-separator">: </span>
                <span className="env-value">
                  {env.value || (env.valueFrom ? '(From ConfigMap/Secret)' : '-')}
                </span>
              </span>
            </div>
          )) || (
            <div className="empty-state">No environment variables</div>
          )}
        </div>
      </div>

      <div className="sub-module">
        <div className="sub-module-title">Resources</div>
        <div className="info-grid">
          <div className="info-item">
            <span className="info-label">Requests (CPU)</span>
            <span className="info-value">{container?.resources?.requests?.cpu || '-'}</span>
          </div>
          <div className="info-item">
            <span className="info-label">Requests (Memory)</span>
            <span className="info-value">{container?.resources?.requests?.memory || '-'}</span>
          </div>
          <div className="info-item">
            <span className="info-label">Limits (CPU)</span>
            <span className="info-value">{container?.resources?.limits?.cpu || '-'}</span>
          </div>
          <div className="info-item">
            <span className="info-label">Limits (Memory)</span>
            <span className="info-value">{container?.resources?.limits?.memory || '-'}</span>
          </div>
        </div>
      </div>

      <div className="sub-module">
        <div className="sub-module-title">Health Checks</div>
        <div className="info-grid">
          <div className="info-item">
            <span className="info-label">Liveness Probe</span>
            <span className="info-value">
              {container?.livenessProbe ? 'Configured' : 'Not Configured'}
            </span>
          </div>
          <div className="info-item">
            <span className="info-label">Readiness Probe</span>
            <span className="info-value">
              {container?.readinessProbe ? 'Configured' : 'Not Configured'}
            </span>
          </div>
        </div>
      </div>

      <div className="sub-module">
        <div className="sub-module-title">Volumes</div>
        <div className="info-grid">
          {container?.volumeMounts?.map((mount: any, idx: number) => (
            <div key={idx} className="info-item">
              <span className="info-label">{mount.name}</span>
              <span className="info-value">
                {mount.mountPath} {mount.readOnly ? '(RO)' : '(RW)'}
              </span>
            </div>
          )) || (
            <div className="empty-state">No volumes</div>
          )}
        </div>
      </div>
    </div>
  );
};
