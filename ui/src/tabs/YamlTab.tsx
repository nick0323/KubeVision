import React, { useState, useCallback, useEffect, useRef } from 'react';
import { YamlTabProps } from '../pages/ResourceDetailPage.types';
import { LoadingSpinner } from '../common/LoadingSpinner';
import { ErrorDisplay } from '../common/ErrorDisplay';
import { notification } from '../common/Notification';
import { authFetch } from '../utils/auth';
import { FaCopy, FaEdit, FaExchangeAlt, FaSave, FaRocket, FaTimes } from 'react-icons/fa';
import jsyaml from 'js-yaml';
import Prism from 'prismjs';
import 'prismjs/components/prism-yaml';
import ReactDiffViewer from 'react-diff-viewer-continued';
import './YamlTab.css';

/**
 * 始终hide's字段（inside部Use/already废弃）
 */
const ALWAYS_HIDDEN_FIELDS = [
  'managedFields',
  'selfLink',
  'clusterName',
];

/**
 * YAML Display options
 */
interface YamlDisplayOptions {
  showStatus: boolean;
}

/**
 * 递归filter YAML object'shide字段
 */
const filterHiddenFields = (obj: any, options: YamlDisplayOptions): any => {
  if (typeof obj !== 'object' || obj === null) return obj;
  if (Array.isArray(obj)) return obj.map(item => filterHiddenFields(item, options));

  const filtered: any = {};
  for (const [key, value] of Object.entries(obj)) {
    if (ALWAYS_HIDDEN_FIELDS.includes(key)) continue;
    if (!options.showStatus && key === 'status') continue;
    filtered[key] = filterHiddenFields(value, options);
  }
  return filtered;
};

/**
 * 清理 YAML object'sno效字段（forUpdate）
 */
const cleanYamlForUpdate = (obj: any): any => {
  if (typeof obj !== 'object' || obj === null) return obj;
  if (Array.isArray(obj)) return obj.map(item => cleanYamlForUpdate(item));

  const cleaned: any = {};
  for (const [key, value] of Object.entries(obj)) {
    // 跳过空 uid 's ownerReferences
    if (key === 'ownerReferences' && Array.isArray(value)) {
      cleaned[key] = value.filter((ref: any) => ref.uid && ref.uid.trim() !== '');
      if (cleaned[key].length > 0) {
        cleaned[key] = cleaned[key].map((ref: any) => cleanYamlForUpdate(ref));
      }
      continue;
    }
    // 保留 resourceVersion and uid（K8s Update必需），but跳过空值
    if (['resourceVersion', 'uid'].includes(key)) {
      if (value === '' || value === null || value === undefined) continue;
      cleaned[key] = value;
      continue;
    }
    // 跳过其他only读字段
    if (['creationTimestamp', 'generation', 'selfLink', 'managedFields'].includes(key)) continue;
    // 跳过空值
    if (value === '' || value === null || value === undefined) continue;
    cleaned[key] = cleanYamlForUpdate(value);
  }
  return cleaned;
};

/**
 * resourceTypeMapping
 */
const RESOURCE_TYPE_MAP: Record<string, { apiVersion: string; kind: string; title: string }> = {
  pod: { apiVersion: 'v1', kind: 'Pod', title: 'Pod' },
  deployment: { apiVersion: 'apps/v1', kind: 'Deployment', title: 'Deployment' },
  statefulset: { apiVersion: 'apps/v1', kind: 'StatefulSet', title: 'StatefulSet' },
  daemonset: { apiVersion: 'apps/v1', kind: 'DaemonSet', title: 'DaemonSet' },
  service: { apiVersion: 'v1', kind: 'Service', title: 'Service' },
  configmap: { apiVersion: 'v1', kind: 'ConfigMap', title: 'ConfigMap' },
  secret: { apiVersion: 'v1', kind: 'Secret', title: 'Secret' },
  ingress: { apiVersion: 'networking.k8s.io/v1', kind: 'Ingress', title: 'Ingress' },
  job: { apiVersion: 'batch/v1', kind: 'Job', title: 'Job' },
  cronjob: { apiVersion: 'batch/v1', kind: 'CronJob', title: 'CronJob' },
  pvc: { apiVersion: 'v1', kind: 'PersistentVolumeClaim', title: 'PersistentVolumeClaim' },
  pv: { apiVersion: 'v1', kind: 'PersistentVolume', title: 'PersistentVolume' },
  storageclass: { apiVersion: 'storage.k8s.io/v1', kind: 'StorageClass', title: 'StorageClass' },
  namespace: { apiVersion: 'v1', kind: 'Namespace', title: 'Namespace' },
  node: { apiVersion: 'v1', kind: 'Node', title: 'Node' },
};

/**
 * Common YAML Tab - 查看/Editresource YAML
 */
export const YamlTab: React.FC<YamlTabProps & { pod?: unknown | null }> = ({
  namespace,
  name,
  resourceType = 'pod',
  data,
  pod, // Compatible with PodDetailPage
}) => {
  const [yamlContent, setYamlContent] = useState<string>('');
  const [originalYaml, setOriginalYaml] = useState<string>('');
  const [editing, setEditing] = useState(false);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [showDiff, setShowDiff] = useState(false);
  const [copySuccess, setCopySuccess] = useState(false);

  // YAML Display options
  const [displayOptions, setDisplayOptions] = useState<YamlDisplayOptions>({
    showStatus: false,
  });

  const editorRef = useRef<HTMLPreElement>(null);

  // GetResource type config
  const resourceConfig = RESOURCE_TYPE_MAP[resourceType] || {
    apiVersion: 'v1',
    kind: resourceType.charAt(0).toUpperCase() + resourceType.slice(1),
    title: resourceType,
  };

  // Load YAML
  const loadYaml = useCallback(async () => {
    setLoading(true);
    setError(null);

    try {
      // backendUse单数形式：/api/{resourceType}/{namespace}/{name}/yaml
      const response = await authFetch(`/api/${resourceType}/${namespace}/${name}/yaml`);
      const result = await response.json();

      if (result.code === 0 && result.data) {
        // filterhide字段
        const filteredData =
          typeof result.data === 'string'
            ? jsyaml.load(result.data)
            : filterHiddenFields(result.data, displayOptions);

        // ensureinclude完整's TypeMeta 字段
        const resourceWithMeta = {
          apiVersion: resourceConfig.apiVersion,
          kind: resourceConfig.kind,
          ...filteredData,
        };

        const yaml = jsyaml.dump(resourceWithMeta, {
          indent: 2,
          lineWidth: -1,
          noRefs: true,
          quotingType: '"',
          forceQuotes: false,
        });

        setYamlContent(yaml);
        setOriginalYaml(yaml);
      } else {
          setError(result.message || 'Failed to load YAML');
      }
    } catch {
      setError('Load Failed');
    } finally {
      setLoading(false);
    }
  }, [namespace, name, resourceType, resourceConfig, displayOptions]);

  // initialLoad
  useEffect(() => {
    // preferUse data prop，其次Use pod prop（Compatible with PodDetailPage）
    const resourceData = data || pod;

    if (resourceData) {
      // filter掉not needed's字段
      const filteredData = filterHiddenFields(resourceData, displayOptions);

      // ensureinclude完整's TypeMeta 字段
      const resourceWithMeta = {
        apiVersion: resourceConfig.apiVersion,
        kind: resourceConfig.kind,
        ...filteredData,
      };
      const yaml = jsyaml.dump(resourceWithMeta, {
        indent: 2,
        lineWidth: -1,
        noRefs: true,
        quotingType: '"',
        forceQuotes: false,
      });
      setYamlContent(yaml);
      setOriginalYaml(yaml);
    }
    // if没hasdata，from API Load
    if (!resourceData) {
      loadYaml();
    }
  }, [data, pod, resourceConfig, displayOptions, loadYaml]);

  // Syntax highlighting
  useEffect(() => {
    if (editorRef.current && !editing) {
      Prism.highlightAllUnder(editorRef.current);
    }
  }, [yamlContent, editing]);

  // 复制 YAML
  const handleCopy = useCallback(async () => {
    try {
      await navigator.clipboard.writeText(yamlContent);
      setCopySuccess(true);
      setTimeout(() => setCopySuccess(false), 2000);
    } catch {
      notification.error('Copy failed');
    }
  }, [yamlContent]);

  // 进入EditMode
  const handleEdit = useCallback(() => {
    setEditing(true);
    setShowDiff(false);
  }, []);

  // CancelEdit
  const handleCancel = useCallback(() => {
    setEditing(false);
    setYamlContent(originalYaml);
    setShowDiff(false);
  }, [originalYaml]);

  // Save修改
  const handleSave = useCallback(async () => {
    setLoading(true);
    try {
      const parsed = jsyaml.load(yamlContent);
      // 清理no效字段
      const cleaned = cleanYamlForUpdate(parsed);
      const response = await authFetch(`/api/${resourceType}/${namespace}/${name}/yaml`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ yaml: cleaned }),
      });

      const contentType = response.headers.get('content-type');
      if (!contentType || !contentType.includes('application/json')) {
        const text = await response.text();
          throw new Error(`Server returned non-JSON response: ${text.substring(0, 100)}`);
      }

      const result = await response.json();

      if (result.code === 0) {
        notification.success('YAML saved');
        setEditing(false);
        setOriginalYaml(yamlContent);
      } else {
        notification.error(`Save failed: ${result.message}`);
      }
    } catch (err) {
      notification.error(`Save failed: ${err instanceof Error ? err.message : 'Unknown error'}`);
    } finally {
      setLoading(false);
    }
  }, [namespace, name, resourceType, yamlContent]);

  // 应use更改
  const handleApply = useCallback(async () => {
    setLoading(true);
    try {
      const parsed = jsyaml.load(yamlContent);
      // 清理no效字段
      const cleaned = cleanYamlForUpdate(parsed);
      const response = await authFetch(`/api/${resourceType}/${namespace}/${name}/yaml`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(cleaned),
      });
      const result = await response.json();

      if (result.code === 0) {
        notification.success('Config applied');
        setEditing(false);
        setOriginalYaml(yamlContent);
      } else {
        notification.error(`Apply failed: ${result.message}`);
      }
    } catch (err) {
      notification.error(`Apply failed: ${err instanceof Error ? err.message : 'Unknown error'}`);
    } finally {
      setLoading(false);
    }
  }, [namespace, name, resourceType, yamlContent]);

  // toggle Diff 视图
  const toggleDiff = useCallback(() => {
    setShowDiff(prev => !prev);
  }, []);

  if (loading && !yamlContent) {
          return <LoadingSpinner text="Loading YAML..." size="lg" />;
  }

  if (error && !yamlContent) {
    return <ErrorDisplay message={error} type="error" showRetry onRetry={loadYaml} />;
  }

  return (
    <div className="yaml-tab">
      {/* Toolbar */}
      <div className="yaml-toolbar">
        <div className="yaml-toolbar-left">
          <span className="yaml-title">{resourceConfig.title} YAML</span>
          <span className="yaml-title-separator">|</span>
          <span className="yaml-status-inline">
            Lines: {yamlContent.split('\n').length} | Chars: {yamlContent.length}
          </span>
          {editing && <span className="yaml-editing-badge">Editing</span>}
        </div>
        <div className="yaml-toolbar-actions">
          {/* Display options toggle */}
          <div className="yaml-display-toggles">
            <label className="toggle-label">
              <span className="toggle-text">Status</span>
              <button
                className={`toggle-switch ${displayOptions.showStatus ? 'active' : ''}`}
                onClick={() =>
                  setDisplayOptions(prev => ({ ...prev, showStatus: !prev.showStatus }))
                }
                title="Toggle Status field"
              >
                <span className="toggle-slider"></span>
              </button>
            </label>
          </div>

          {!editing ? (
            <>
              <button
                className={`toolbar-btn ${copySuccess ? 'success' : ''}`}
                onClick={handleCopy}
                title="Copy to clipboard"
              >
                <FaCopy />
              </button>
              <button className="toolbar-btn" onClick={handleEdit} title="Edit YAML">
                <FaEdit />
              </button>
            </>
          ) : (
            <>
              <button className="toolbar-btn" onClick={toggleDiff} title="Show diff">
                <FaExchangeAlt />
              </button>
              <button className="toolbar-btn" onClick={handleSave} title="Save changes">
                <FaSave />
              </button>
              <button
                className="toolbar-btn primary"
                onClick={handleApply}
                title="Apply to cluster"
              >
                <FaRocket />
              </button>
              <button className="toolbar-btn danger" onClick={handleCancel} title="Cancel editing">
                <FaTimes />
              </button>
            </>
          )}
        </div>
      </div>

      {/* Editor/Viewer */}
      {showDiff && editing ? (
        <div className="diff-container">
          <ReactDiffViewer
            oldValue={originalYaml}
            newValue={yamlContent}
            splitView={true}
            leftTitle="Original Version"
            rightTitle="Current Version"
            useDarkTheme={false}
            styles={{
              variables: {
                light: {
                  diffViewerBackground: '#fff',
                  diffViewerColor: '#333',
                },
              },
              line: {
                '&:hover': {
                  background: 'rgba(0, 0, 0, 0.05)',
                },
              },
              marker: {
                markerBackground: 'rgba(0, 0, 0, 0.1)',
              },
            }}
          />
        </div>
      ) : editing ? (
        <div className="yaml-editor-container">
          <textarea
            className="yaml-editor-textarea"
            value={yamlContent}
            onChange={e => setYamlContent(e.target.value)}
            spellCheck={false}
            autoCapitalize="off"
            autoComplete="off"
          />
        </div>
      ) : (
        <div className="yaml-editor-container">
          <pre className="yaml-editor-pre" ref={editorRef}>
            <code className="language-yaml">{yamlContent}</code>
          </pre>
        </div>
      )}
    </div>
  );
};

export default YamlTab;
