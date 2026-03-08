import React from 'react';
import './NamespaceFilter.css';

interface NamespaceFilterProps {
  namespaces: string[];
  value: string;
  onChange: (value: string) => void;
}

/**
 * 命名空间过滤组件
 */
export const NamespaceFilter: React.FC<NamespaceFilterProps> = ({
  namespaces,
  value,
  onChange,
}) => {
  const handleChange = (e: React.ChangeEvent<HTMLSelectElement>) => {
    onChange(e.target.value);
  };

  const handleClear = () => {
    onChange('');
  };

  return (
    <div className="namespace-filter">
      <span className="filter-label">命名空间:</span>
      <select
        className="filter-select"
        value={value}
        onChange={handleChange}
      >
        <option value="">全部</option>
        {namespaces.map((ns) => (
          <option key={ns} value={ns}>
            {ns}
          </option>
        ))}
      </select>
      {value && (
        <button className="clear-btn" onClick={handleClear} type="button">
          ✕
        </button>
      )}
    </div>
  );
};

export default NamespaceFilter;
