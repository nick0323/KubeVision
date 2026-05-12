import React, { useState } from 'react';
import { FaChevronDown, FaChevronUp } from 'react-icons/fa';
import { useClickOutside } from '../hooks/useClickOutside';
import './NamespaceSelect.css';

interface NamespaceSelectProps {
  value: string;
  onChange: (value: string) => void;
  placeholder?: string;
  disabled?: boolean;
  options?: string[]; // Namespace list passed from parent component
  className?: string; // Custom className
  width?: string; // Custom width
}

/**
 * Custom Namespace SelectorComponent - Support完全Stylecontrol
 */
export const NamespaceSelect: React.FC<NamespaceSelectProps> = ({
  value,
  onChange,
  placeholder = 'Select namespace',
  disabled = false,
  options = [],
  className = '',
  width,
}) => {
  const [isOpen, setIsOpen] = useState(false);
  const containerRef = useClickOutside<HTMLDivElement>(() => setIsOpen(false));

  const handleSelect = (ns: string) => {
    onChange(ns);
    setIsOpen(false);
  };

  const toggleDropdown = () => {
    if (!disabled) {
      setIsOpen(!isOpen);
    }
  };

  const selectedLabel = value || placeholder;

  return (
    <div
      className={`namespace-select-custom ${className || ''}`}
      ref={containerRef}
      style={width ? { width, minWidth: width } : {}}
    >
      {/* Select box body */}
      <div
        className={`namespace-select-value ${isOpen ? 'open' : ''} ${disabled ? 'disabled' : ''}`}
        onClick={toggleDropdown}
      >
        <span className="namespace-select-text">{selectedLabel}</span>
        <span className="namespace-select-arrow">
          {isOpen ? <FaChevronUp /> : <FaChevronDown />}
        </span>
      </div>

      {/* Dropdown options */}
      {isOpen && (
        <div className="namespace-select-dropdown">
          {/* All options - only shown when placeholder has value */}
          {placeholder && (
            <div
              className={`namespace-select-option ${!value ? 'selected' : ''}`}
              onClick={() => handleSelect('')}
            >
              {placeholder}
            </div>
          )}

          {/* Namespace list */}
          {options.length === 0 ? (
            <div className="namespace-select-option disabled">No data yet</div>
          ) : (
            options.map(ns => (
              <div
                key={ns}
                className={`namespace-select-option ${value === ns ? 'selected' : ''}`}
                onClick={() => handleSelect(ns)}
              >
                {ns}
              </div>
            ))
          )}
        </div>
      )}
    </div>
  );
};

export default NamespaceSelect;
