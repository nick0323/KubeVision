import React, { useCallback, useEffect, useRef, useReducer } from 'react';
import { YamlTabProps } from '../pages/ResourceDetailPage.types';
import { LoadingSpinner } from '../common/LoadingSpinner';
import { ErrorDisplay } from '../common/ErrorDisplay';
import { notification } from '../common/NotificationContext';
import { authFetch, withCluster } from '../utils/auth';
import { isClusterResource } from '../constants/config';
import { capitalize } from '../utils/string';
import { downloadFile } from '../utils/download';
import { filterHiddenFields } from '../utils/filterHiddenFields';
import { FaCopy, FaDownload, FaEdit, FaExchangeAlt, FaSave, FaRocket, FaTimes } from 'react-icons/fa';
import jsyaml from 'js-yaml';
import Prism from 'prismjs';
import 'prismjs/components/prism-yaml';
import ReactDiffViewer from 'react-diff-viewer-continued';
import './YamlTab.css';

interface YamlDisplayOptions {
  showStatus: boolean;
}

interface YamlState {
  yamlContent: string;
  originalYaml: string;
  editing: boolean;
  loading: boolean;
  error: string | null;
  showDiff: boolean;
  copySuccess: boolean;
  displayOptions: YamlDisplayOptions;
}

type YamlAction =
  | { type: 'SET_YAML'; payload: string }
  | { type: 'SET_ORIGINAL'; payload: string }
  | { type: 'SET_BOTH'; payload: string }
  | { type: 'SET_LOADING'; payload: boolean }
  | { type: 'SET_ERROR'; payload: string | null }
  | { type: 'START_EDIT' }
  | { type: 'CANCEL_EDIT' }
  | { type: 'SAVE_SUCCESS' }
  | { type: 'TOGGLE_DIFF' }
  | { type: 'SET_COPY_SUCCESS'; payload: boolean }
  | { type: 'TOGGLE_STATUS' };

function yamlReducer(state: YamlState, action: YamlAction): YamlState {
  switch (action.type) {
    case 'SET_YAML':
      return { ...state, yamlContent: action.payload };
    case 'SET_ORIGINAL':
      return { ...state, originalYaml: action.payload };
    case 'SET_BOTH':
      return { ...state, yamlContent: action.payload, originalYaml: action.payload };
    case 'SET_LOADING':
      return { ...state, loading: action.payload };
    case 'SET_ERROR':
      return { ...state, error: action.payload };
    case 'START_EDIT':
      return { ...state, editing: true, showDiff: false };
    case 'CANCEL_EDIT':
      return { ...state, editing: false, yamlContent: state.originalYaml, showDiff: false };
    case 'SAVE_SUCCESS':
      return { ...state, editing: false, originalYaml: state.yamlContent };
    case 'TOGGLE_DIFF':
      return { ...state, showDiff: !state.showDiff };
    case 'SET_COPY_SUCCESS':
      return { ...state, copySuccess: action.payload };
    case 'TOGGLE_STATUS':
      return { ...state, displayOptions: { ...state.displayOptions, showStatus: !state.displayOptions.showStatus } };
    default:
      return state;
  }
}

const initialState: YamlState = {
  yamlContent: '',
  originalYaml: '',
  editing: false,
  loading: false,
  error: null,
  showDiff: false,
  copySuccess: false,
  displayOptions: { showStatus: false },
};

const cleanYamlForUpdate = (obj: Record<string, unknown>): Record<string, unknown> => {
  if (typeof obj !== 'object' || obj === null) return obj;
  if (Array.isArray(obj)) return obj.map(item => cleanYamlForUpdate(item as Record<string, unknown>)) as unknown as Record<string, unknown>;

  const cleaned: Record<string, unknown> = {};
  for (const [key, value] of Object.entries(obj)) {
    if (key === 'ownerReferences' && Array.isArray(value)) {
      cleaned[key] = value.filter((ref: Record<string, unknown>) => ref.uid && String(ref.uid).trim() !== '');
      if ((cleaned[key] as unknown[]).length > 0) {
        cleaned[key] = (cleaned[key] as unknown[]).map((ref: unknown) => cleanYamlForUpdate(ref as Record<string, unknown>));
      }
      continue;
    }
    if (['resourceVersion', 'uid'].includes(key)) {
      if (value === '' || value === null || value === undefined) continue;
      cleaned[key] = value;
      continue;
    }
    if (['creationTimestamp', 'generation', 'selfLink', 'managedFields'].includes(key)) continue;
    if (value === '' || value === null || value === undefined) continue;
    cleaned[key] = cleanYamlForUpdate(value as Record<string, unknown>);
  }
  return cleaned;
};

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

function yamlToDump(data: Record<string, unknown>, resourceConfig: { apiVersion: string; kind: string }): string {
  return jsyaml.dump(
    { apiVersion: resourceConfig.apiVersion, kind: resourceConfig.kind, ...data },
    { indent: 2, lineWidth: -1, noRefs: true, quotingType: '"', forceQuotes: false },
  );
}

function buildYamlPath(isClusterScope: boolean, resourceType: string, namespace: string | undefined, name: string): string {
  return isClusterScope
    ? `/api/${resourceType}/_cluster_/${name}/yaml`
    : `/api/${resourceType}/${namespace}/${name}/yaml`;
}

async function submitYaml(
  yamlPath: string,
  body: Record<string, unknown>,
  useCluster: boolean,
): Promise<{ success: boolean; message: string }> {
  const url = useCluster ? withCluster(yamlPath) : yamlPath;
  const response = await authFetch(url, {
    method: 'PUT',
    headers: { 'Content-Type': 'application/json' },
    body: JSON.stringify(body),
  });
  const result = await response.json();
  return { success: result.code === 0, message: result.message || 'Unknown error' };
}

export const YamlTab: React.FC<YamlTabProps & { pod?: unknown | null }> = ({
  namespace,
  name,
  resourceType = 'pod',
  data,
  pod,
}) => {
  const [state, dispatch] = useReducer(yamlReducer, initialState);
  const editorRef = useRef<HTMLPreElement>(null);

  const resourceConfig = RESOURCE_TYPE_MAP[resourceType] || {
    apiVersion: 'v1',
    kind: capitalize(resourceType),
    title: resourceType,
  };

  const isClusterScope = isClusterResource(resourceType);

  const loadYaml = useCallback(async () => {
    dispatch({ type: 'SET_LOADING', payload: true });
    dispatch({ type: 'SET_ERROR', payload: null });

    try {
      const yamlPath = buildYamlPath(isClusterScope, resourceType, namespace, name);
      const response = await authFetch(withCluster(yamlPath));
      const result = await response.json();

      if (result.code === 0 && result.data) {
        const filteredData = filterHiddenFields(result.data as Record<string, unknown>, state.displayOptions);
        dispatch({ type: 'SET_BOTH', payload: yamlToDump(filteredData as Record<string, unknown>, resourceConfig) });
      } else {
        dispatch({ type: 'SET_ERROR', payload: result.message || 'Failed to load YAML' });
      }
    } catch {
      dispatch({ type: 'SET_ERROR', payload: 'Load Failed' });
    } finally {
      dispatch({ type: 'SET_LOADING', payload: false });
    }
  }, [namespace, name, resourceType, isClusterScope, resourceConfig, state.displayOptions]);

  useEffect(() => {
    const resourceData = (data || pod) as Record<string, unknown>;
    if (resourceData) {
      const filteredData = filterHiddenFields(resourceData, state.displayOptions);
      dispatch({ type: 'SET_BOTH', payload: yamlToDump(filteredData as Record<string, unknown>, resourceConfig) });
    }
    if (!resourceData) {
      loadYaml();
    }
  }, [data, pod, resourceConfig, state.displayOptions, loadYaml]);

  useEffect(() => {
    if (editorRef.current && !state.editing) {
      Prism.highlightAllUnder(editorRef.current);
    }
  }, [state.yamlContent, state.editing]);

  const handleCopy = useCallback(async () => {
    try {
      await navigator.clipboard.writeText(state.yamlContent);
      dispatch({ type: 'SET_COPY_SUCCESS', payload: true });
      setTimeout(() => dispatch({ type: 'SET_COPY_SUCCESS', payload: false }), 2000);
    } catch {
      notification.error('Copy failed');
    }
  }, [state.yamlContent]);

  const handleDownload = useCallback(() => {
    downloadFile(state.yamlContent, `${name}.yaml`, 'text/yaml;charset=utf-8');
  }, [state.yamlContent, name]);

  const handleSave = useCallback(async () => {
    dispatch({ type: 'SET_LOADING', payload: true });
    try {
      const parsed = jsyaml.load(state.yamlContent) as Record<string, unknown>;
      const cleaned = cleanYamlForUpdate(parsed);
      const yamlPath = buildYamlPath(isClusterScope, resourceType, namespace, name);
      const { success, message } = await submitYaml(yamlPath, { yaml: cleaned }, false);
      if (success) {
        notification.success('YAML saved');
        dispatch({ type: 'SAVE_SUCCESS' });
      } else {
        notification.error(`Save failed: ${message}`);
      }
    } catch (err) {
      notification.error(`Save failed: ${err instanceof Error ? err.message : 'Unknown error'}`);
    } finally {
      dispatch({ type: 'SET_LOADING', payload: false });
    }
  }, [namespace, name, resourceType, isClusterScope, state.yamlContent]);

  const handleApply = useCallback(async () => {
    dispatch({ type: 'SET_LOADING', payload: true });
    try {
      const parsed = jsyaml.load(state.yamlContent) as Record<string, unknown>;
      const cleaned = cleanYamlForUpdate(parsed);
      const yamlPath = buildYamlPath(isClusterScope, resourceType, namespace, name);
      const { success, message } = await submitYaml(yamlPath, cleaned, true);
      if (success) {
        notification.success('Config applied');
        dispatch({ type: 'SAVE_SUCCESS' });
      } else {
        notification.error(`Apply failed: ${message}`);
      }
    } catch (err) {
      notification.error(`Apply failed: ${err instanceof Error ? err.message : 'Unknown error'}`);
    } finally {
      dispatch({ type: 'SET_LOADING', payload: false });
    }
  }, [namespace, name, resourceType, isClusterScope, state.yamlContent]);

  if (state.loading && !state.yamlContent) {
    return <LoadingSpinner text="Loading YAML..." size="lg" />;
  }

  if (state.error && !state.yamlContent) {
    return <ErrorDisplay message={state.error} type="error" showRetry onRetry={loadYaml} />;
  }

  return (
    <div className="yaml-tab">
      <div className="yaml-toolbar">
        <div className="yaml-toolbar-left">
          <span className="yaml-title">{resourceConfig.title} YAML</span>
          <span className="yaml-title-separator">|</span>
          <span className="yaml-status-inline">
            Lines: {state.yamlContent.split('\n').length} | Chars: {state.yamlContent.length}
          </span>
          {state.editing && <span className="yaml-editing-badge">Editing</span>}
        </div>
        <div className="yaml-toolbar-actions">
          <div className="yaml-display-toggles">
            <label className="toggle-label">
              <span className="toggle-text">Status</span>
              <button
                className={`toggle-switch ${state.displayOptions.showStatus ? 'active' : ''}`}
                onClick={() => dispatch({ type: 'TOGGLE_STATUS' })}
                title="Toggle Status field"
              >
                <span className="toggle-slider"></span>
              </button>
            </label>
          </div>

          {!state.editing ? (
            <>
              <button
                className={`toolbar-btn ${state.copySuccess ? 'success' : ''}`}
                onClick={handleCopy}
                title="Copy to clipboard"
              >
                <FaCopy />
              </button>
              <button className="toolbar-btn" onClick={handleDownload} title="Download YAML">
                <FaDownload />
              </button>
              <button className="toolbar-btn" onClick={() => dispatch({ type: 'START_EDIT' })} title="Edit YAML">
                <FaEdit />
              </button>
            </>
          ) : (
            <>
              <button className="toolbar-btn" onClick={() => dispatch({ type: 'TOGGLE_DIFF' })} title="Show diff">
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
              <button className="toolbar-btn danger" onClick={() => dispatch({ type: 'CANCEL_EDIT' })} title="Cancel editing">
                <FaTimes />
              </button>
            </>
          )}
        </div>
      </div>

      {state.showDiff && state.editing ? (
        <div className="diff-container">
          <ReactDiffViewer
            oldValue={state.originalYaml}
            newValue={state.yamlContent}
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
      ) : state.editing ? (
        <div className="yaml-editor-container">
          <textarea
            className="yaml-editor-textarea"
            value={state.yamlContent}
            onChange={e => dispatch({ type: 'SET_YAML', payload: e.target.value })}
            spellCheck={false}
            autoCapitalize="off"
            autoComplete="off"
          />
        </div>
      ) : (
        <div className="yaml-editor-container">
          <pre className="yaml-editor-pre" ref={editorRef}>
            <code className="language-yaml">{state.yamlContent}</code>
          </pre>
        </div>
      )}
    </div>
  );
};

export default YamlTab;
