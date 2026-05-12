import React, { useState, useCallback, useEffect } from 'react';
import { FaTimes, FaRocket, FaCode, FaEraser } from 'react-icons/fa';
import { ResourceTemplate, RESOURCE_TEMPLATES, getTemplateByResourceType } from '../constants/templates';
import TemplateSelector from './TemplateSelector';
import { LoadingSpinner } from './LoadingSpinner';
import { notification } from './Notification';
import { authFetch } from '../utils/auth';
import { capitalize } from '../utils/string';
import './CreateResourceModal.css';

interface CreateResourceModalProps {
  visible: boolean;
  resourceType: string;
  namespace: string;
  onClose: () => void;
  onSuccess: () => void;
}

export const CreateResourceModal: React.FC<CreateResourceModalProps> = ({
  visible,
  resourceType,
  namespace,
  onClose,
  onSuccess,
}) => {
  const [mode, setMode] = useState<'template' | 'yaml'>('template');
  const [selectedTemplate, setSelectedTemplate] = useState<ResourceTemplate | null>(null);
  const [yamlContent, setYamlContent] = useState('');
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);

  // 当 resourceType 变化时，自动选择对应模板
  useEffect(() => {
    if (resourceType && visible) {
      const template = getTemplateByResourceType(resourceType);
      if (template) {
        setSelectedTemplate(template);
        setYamlContent(template.yaml);
        setMode('yaml');
      } else {
        setSelectedTemplate(null);
        setYamlContent('');
        setMode('template');
      }
    }
  }, [resourceType, visible]);

  const handleTemplateSelect = useCallback((template: ResourceTemplate) => {
    setSelectedTemplate(template);
    setYamlContent(template.yaml);
    setMode('yaml');
  }, []);

  const handleCreate = useCallback(async () => {
    if (!yamlContent.trim()) {
      setError('YAML content cannot be empty');
      return;
    }

    setLoading(true);
    setError(null);

    try {
      // 解析 YAML 为 JSON
      const jsyaml = await import('js-yaml');
      const yamlObj = jsyaml.load(yamlContent) as Record<string, any>;

      // 确保 namespace 正确
      if (yamlObj.metadata) {
        yamlObj.metadata.namespace = namespace || yamlObj.metadata.namespace || 'default';
      }

      const response = await authFetch(`/api/${resourceType}/yaml`, {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ yaml: yamlObj }),
      });

      const result = await response.json();

      if (result.code === 0) {
        notification.success('Resource created successfully');
        onSuccess();
        onClose();
      } else {
        setError(result.message || 'Failed to create resource');
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : 'Unknown error');
    } finally {
      setLoading(false);
    }
  }, [yamlContent, resourceType, namespace, onSuccess, onClose]);

  if (!visible) return null;

  return (
    <div className="modal-overlay" onClick={onClose}>
      <div className="modal-content create-resource-modal" onClick={e => e.stopPropagation()}>
        {/* Header */}
        <div className="modal-header">
          <h2 className="modal-title">
            Create {resourceType ? capitalize(resourceType) : 'Resource'}
          </h2>
          <button className="modal-close" onClick={onClose}>
            <FaTimes />
          </button>
        </div>

        {/* Body */}
        <div className="modal-body">
          {mode === 'template' ? (
            <TemplateSelector
              templates={RESOURCE_TEMPLATES}
              selectedTemplate={selectedTemplate?.name || null}
              onSelect={handleTemplateSelect}
            />
          ) : (
            <div className="yaml-editor-section">
              <div className="yaml-editor-header">
                <button
                  className="back-btn"
                  onClick={() => setMode('template')}
                >
                  Back to Templates
                </button>
                <div className="yaml-actions">
                  <button
                    className="action-btn clear-btn"
                    onClick={() => {
                      setSelectedTemplate(null);
                      setYamlContent('');
                    }}
                    title="Clear all content"
                  >
                    <FaEraser />
                  </button>
                </div>
              </div>
              <textarea
                className="yaml-textarea"
                value={yamlContent}
                onChange={e => setYamlContent(e.target.value)}
                placeholder="Edit YAML here..."
                spellCheck={false}
              />
            </div>
          )}

          {error && (
            <div className="error-message">
              {error}
            </div>
          )}
        </div>

        {/* Footer */}
        <div className="modal-footer">
          <button className="btn btn-secondary" onClick={onClose} disabled={loading}>
            Cancel
          </button>
          {mode === 'template' ? (
            <button
              className="btn btn-primary"
              onClick={() => selectedTemplate && setMode('yaml')}
              disabled={!selectedTemplate}
            >
              <FaCode /> Edit YAML
            </button>
          ) : (
            <button
              className="btn btn-primary create-btn"
              onClick={handleCreate}
              disabled={loading || !yamlContent.trim()}
            >
              {loading ? <LoadingSpinner size="sm" /> : null}
              {loading ? 'Creating...' : 'Create'}
            </button>
          )}
        </div>
      </div>
    </div>
  );
};

export default CreateResourceModal;
