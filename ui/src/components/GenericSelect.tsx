import React, { ChangeEvent } from 'react';
import './NamespaceSelect.css';

interface Option {
  label: string;
  value: string;
}

interface GenericSelectProps {
  value: string;
  onChange: (value: string) => void;
  placeholder?: string;
  options?: Option[];
  disabled?: boolean;
  loading?: boolean;
}

/**
 * 通用选择器组件
 */
export const GenericSelect: React.FC<GenericSelectProps> = ({
  value,
  onChange,
  placeholder = '请选择',
  options = [],
  disabled = false,
  loading = false,
}) => {
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
      {options.map((option) => (
        <option key={option.value} value={option.value}>
          {option.label}
        </option>
      ))}
    </select>
  );
};

export default GenericSelect;
