import React, { useEffect, useState, ChangeEvent } from 'react';
import { NamespaceSelectProps } from '../types';
import { apiClient } from '../utils/apiClient';
import './NamespaceSelect.css';

interface Namespace {
  name: string;
  status: string;
}

/**
 * Namespace 选择器组件
 */
export const NamespaceSelect: React.FC<NamespaceSelectProps> = ({
  value,
  onChange,
  placeholder = '选择命名空间',
  disabled = false,
}) => {
  const [namespaces, setNamespaces] = useState<Namespace[]>([]);
  const [loading, setLoading] = useState<boolean>(false);

  useEffect(() => {
    const fetchNamespaces = async () => {
      setLoading(true);
      try {
        const result = await apiClient.get<{ data: Namespace[] }>('/api/namespaces');
        setNamespaces(result.data || []);
      } catch (error) {
        console.error('Failed to fetch namespaces:', error);
      } finally {
        setLoading(false);
      }
    };

    fetchNamespaces();
  }, []);

  const handleChange = (e: ChangeEvent<HTMLSelectElement>) => {
    onChange(e.target.value);
  };

  return (
    <select
      value={value}
      onChange={handleChange}
      disabled={disabled || loading}
      className="namespace-select"
    >
      <option value="">{placeholder}</option>
      {namespaces.map((ns) => (
        <option key={ns.name} value={ns.name}>
          {ns.name}
        </option>
      ))}
    </select>
  );
};

export default NamespaceSelect;
