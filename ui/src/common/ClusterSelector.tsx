import React from 'react';
import { FaCloud } from 'react-icons/fa';
import { ClusterHealth } from '../types';

interface ClusterSelectorProps {
  clusters: string[];
  clusterHealth: Record<string, ClusterHealth>;
  currentCluster: string;
  clusterOpen: boolean;
  onToggle: () => void;
  onChange: (name: string) => void;
  innerRef: React.RefObject<HTMLDivElement | null>;
}

const ClusterSelector: React.FC<ClusterSelectorProps> = ({
  clusters, clusterHealth, currentCluster, clusterOpen, onToggle, onChange, innerRef,
}) => {
  const health = clusterHealth[currentCluster];
  const tooltip = health ? `${health.host} | v${health.version} | ${health.nodeCount} nodes` : '';

  return (
    <div className="cluster-selector" ref={innerRef}>
      <div
        className="cluster-selector-trigger"
        onClick={onToggle}
        title={tooltip}
      >
        <span className="icon"><FaCloud /></span>
        <span className={`cluster-status-dot ${clusterHealth[currentCluster]?.healthy ? 'healthy' : 'unhealthy'}`} />
        <span className="cluster-name">{currentCluster}</span>
        <span className="cluster-arrow">{clusterOpen ? '▲' : '▼'}</span>
      </div>
      {clusterOpen && (
        <div className="cluster-dropdown">
          {clusters.map(name => {
            const h = clusterHealth[name];
            return (
              <div
                key={name}
                className={`cluster-option ${name === currentCluster ? 'active' : ''}`}
                onClick={() => onChange(name)}
              >
                <span className="cluster-check">{name === currentCluster ? '✓' : ''}</span>
                <span className={`cluster-status-dot ${h?.healthy ? 'healthy' : 'unhealthy'}`} />
                {name}
              </div>
            );
          })}
        </div>
      )}
    </div>
  );
};

export default React.memo(ClusterSelector);
