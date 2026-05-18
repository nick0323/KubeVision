import React from 'react';
import { FaRocket, FaCube, FaNetworkWired, FaFileAlt, FaLock, FaHdd, FaDoorOpen, FaTree, FaCogs, FaBriefcase, FaClock, FaThLarge } from 'react-icons/fa';
import { ResourceTemplate } from '../constants/templates';
import './TemplateSelector.css';

interface TemplateSelectorProps {
  templates: ResourceTemplate[];
  selectedTemplate: string | null;
  onSelect: (template: ResourceTemplate) => void;
}

const ICON_MAP: Record<string, React.ReactNode> = {
  FaRocket: <FaRocket />,
  FaCube: <FaCube />,
  FaNetworkWired: <FaNetworkWired />,
  FaFileAlt: <FaFileAlt />,
  FaLock: <FaLock />,
  FaHdd: <FaHdd />,
  FaDoorOpen: <FaDoorOpen />,
  FaTree: <FaTree />,
  FaCogs: <FaCogs />,
  FaBriefcase: <FaBriefcase />,
  FaClock: <FaClock />,
  FaThLarge: <FaThLarge />,
};

const TemplateSelector: React.FC<TemplateSelectorProps> = ({
  templates,
  selectedTemplate,
  onSelect,
}) => {
  return (
    <div className="template-selector">
      <h3 className="template-selector-title">Select Template</h3>
      <div className="template-grid">
        {templates.map(template => (
          <div
            key={template.name}
            className={`template-card ${selectedTemplate === template.name ? 'selected' : ''}`}
            onClick={() => onSelect(template)}
          >
            <div className="template-icon">{ICON_MAP[template.icon]}</div>
            <div className="template-info">
              <div className="template-label">{template.label}</div>
              <div className="template-description">{template.description}</div>
            </div>
          </div>
        ))}
      </div>
    </div>
  );
};

export default React.memo(TemplateSelector);
