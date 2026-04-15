import React, { useState, useCallback, useEffect, useRef } from 'react';
import { YamlTabProps } from '../pages/ResourceDetailPage.types';
import { LoadingSpinner } from '../common/LoadingSpinner';
import { ErrorDisplay } from '../common/ErrorDisplay';
import { authFetch } from '../utils/auth';
import { FaCopy, FaEdit, FaExchangeAlt, FaSave, FaRocket, FaTimes } from 'react-icons/fa';
import jsyaml from 'js-yaml';
import Prism from 'prismjs';
import 'prismjs/components/prism-yaml';
import ReactDiffViewer from 'react-diff-viewer-continued';
import './YamlTab.css';

/**
 * 始终隐藏的字段（内部使用/已废弃）
 */
const ALWAYS_HIDDEN_FIELDS = [
  'managedFields',
  'selfLink',
  'clusterName',
];

/**
 * YAML 显示选项
 */
interface YamlDisplayOptions {
  showStatus: boolean;
}

/**
 * 递归过滤 YAML 对象中的隐藏字段
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
 * 清理 YAML 对象中的无效字段（用于更新）
 */
const cleanYamlForUpdate = (obj: any): any => {
  if (typeof obj !== 'object' || obj === null) return obj;
  if (Array.isArray(obj)) return obj.map(item => cleanYamlForUpdate(item));

  const cleaned: any = {};
  for (const [key, value] of Object.entries(obj)) {
    // 跳过空 uid 的 ownerReferences
    if (key === 'ownerReferences' && Array.isArray(value)) {
      cleaned[key] = value.filter((ref: any) => ref.uid && ref.uid.trim() !== '');
      if (cleaned[key].length > 0) {
        cleaned[key] = cleaned[key].map((ref: any) => cleanYamlForUpdate(ref));
      }
      continue;
    }
    // 保留 resourceVersion 和 uid（K8s 更新必需），但跳过空值
    if (['resourceVersion', 'uid'].includes(key)) {
      if (value === '' || value === null || value === undefined) continue;
      cleaned[key] = value;
      continue;
    }
    // 跳过其他只读字段
    if (['creationTimestamp', 'generation', 'selfLink', 'managedFields'].includes(key)) continue;
    // 跳过空值
    if (value === '' || value === null || value === undefined) continue;
    cleaned[key] = cleanYamlForUpdate(value);
  }
  return cleaned;
};

/**
 * 资源类型映射
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
 * 通用 YAML Tab - 查看/编辑资源 YAML
 */
export const YamlTab: React.FC<YamlTabProps & { pod?: unknown | null }> = ({
  namespace,
  name,
  resourceType = 'pod',
  data,
  pod, // 兼容 PodDetailPage
}) => {
  const [yamlContent, setYamlContent] = useState<string>('');
  const [originalYaml, setOriginalYaml] = useState<string>('');
  const [editing, setEditing] = useState(false);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [showDiff, setShowDiff] = useState(false);
  const [copySuccess, setCopySuccess] = useState(false);

  // YAML 显示选项
  const [displayOptions, setDisplayOptions] = useState<YamlDisplayOptions>({
    showStatus: false,
  });

  const editorRef = useRef<HTMLPreElement>(null);

  // 获取资源类型配置
  const resourceConfig = RESOURCE_TYPE_MAP[resourceType] || {
    apiVersion: 'v1',
    kind: resourceType.charAt(0).toUpperCase() + resourceType.slice(1),
    title: resourceType,
  };

  // 加载 YAML
  const loadYaml = useCallback(async () => {
    setLoading(true);
    setError(null);

    try {
      // 后端使用单数形式：/api/{resourceType}/{namespace}/{name}/yaml
      const response = await authFetch(`/api/${resourceType}/${namespace}/${name}/yaml`);
      const result = await response.json();

      if (result.code === 0 && result.data) {
        // 过滤隐藏字段
        const filteredData =
          typeof result.data === 'string'
            ? jsyaml.load(result.data)
            : filterHiddenFields(result.data, displayOptions);

        // 确保包含完整的 TypeMeta 字段
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
        setError(result.message || '加载 YAML 失败');
      }
    } catch {
      setError('加载失败');
    } finally {
      setLoading(false);
    }
  }, [namespace, name, resourceType, resourceConfig, displayOptions]);

  // 初始加载
  useEffect(() => {
    // 优先使用 data prop，其次使用 pod prop（兼容 PodDetailPage）
    const resourceData = data || pod;

    if (resourceData) {
      // 过滤掉不需要的字段
      const filteredData = filterHiddenFields(resourceData, displayOptions);

      // 确保包含完整的 TypeMeta 字段
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
    // 如果没有数据，从 API 加载
    if (!resourceData) {
      loadYaml();
    }
  }, [data, pod, resourceConfig, displayOptions, loadYaml]);

  // 语法高亮
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
      alert('复制失败');
    }
  }, [yamlContent]);

  // 进入编辑模式
  const handleEdit = useCallback(() => {
    setEditing(true);
    setShowDiff(false);
  }, []);

  // 取消编辑
  const handleCancel = useCallback(() => {
    setEditing(false);
    setYamlContent(originalYaml);
    setShowDiff(false);
  }, [originalYaml]);

  // 保存修改
  const handleSave = useCallback(async () => {
    setLoading(true);
    try {
      const parsed = jsyaml.load(yamlContent);
      // 清理无效字段
      const cleaned = cleanYamlForUpdate(parsed);
      const response = await authFetch(`/api/${resourceType}/${namespace}/${name}/yaml`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ yaml: cleaned }),
      });

      const contentType = response.headers.get('content-type');
      if (!contentType || !contentType.includes('application/json')) {
        const text = await response.text();
        throw new Error(`服务器返回了非 JSON 响应：${text.substring(0, 100)}`);
      }

      const result = await response.json();

      if (result.code === 0) {
        alert('YAML 已保存');
        setEditing(false);
        setOriginalYaml(yamlContent);
      } else {
        alert(`保存失败：${result.message}`);
      }
    } catch (err) {
      alert(`保存失败：${err instanceof Error ? err.message : '未知错误'}`);
    } finally {
      setLoading(false);
    }
  }, [namespace, name, resourceType, yamlContent]);

  // 应用更改
  const handleApply = useCallback(async () => {
    if (!window.confirm('确定要应用此 YAML 配置到集群吗？')) return;

    setLoading(true);
    try {
      const parsed = jsyaml.load(yamlContent);
      // 清理无效字段
      const cleaned = cleanYamlForUpdate(parsed);
      const response = await authFetch(`/api/${resourceType}/${namespace}/${name}/yaml`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(cleaned),
      });
      const result = await response.json();

      if (result.code === 0) {
        alert('配置已应用');
        setEditing(false);
        setOriginalYaml(yamlContent);
      } else {
        alert(`应用失败：${result.message}`);
      }
    } catch (err) {
      alert(`应用失败：${err instanceof Error ? err.message : '未知错误'}`);
    } finally {
      setLoading(false);
    }
  }, [namespace, name, resourceType, yamlContent]);

  // 切换 Diff 视图
  const toggleDiff = useCallback(() => {
    setShowDiff(prev => !prev);
  }, []);

  if (loading && !yamlContent) {
    return <LoadingSpinner text="加载 YAML..." size="lg" />;
  }

  if (error && !yamlContent) {
    return <ErrorDisplay message={error} type="error" showRetry onRetry={loadYaml} />;
  }

  return (
    <div className="yaml-tab">
      {/* 工具栏 */}
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
          {/* 显示选项切换 */}
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

      {/* 编辑器/查看器 */}
      {showDiff && editing ? (
        <div className="diff-container">
          <ReactDiffViewer
            oldValue={originalYaml}
            newValue={yamlContent}
            splitView={true}
            leftTitle="原始版本"
            rightTitle="当前版本"
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
