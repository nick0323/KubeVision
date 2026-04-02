import React, { useState, useCallback, useEffect, useRef } from 'react';
import { YamlTabProps } from '../types';
import { LoadingSpinner } from '../../LoadingSpinner';
import { ErrorDisplay } from '../../ErrorDisplay';
import { authFetch } from '../../../utils/auth';
import { FaCopy, FaEdit, FaExchangeAlt, FaSave, FaRocket, FaTimes, FaEye } from 'react-icons/fa';
import jsyaml from 'js-yaml';
import Prism from 'prismjs';
import 'prismjs/components/prism-yaml';
import ReactDiffViewer from 'react-diff-viewer-continued';
import './YamlTab.css';

/**
 * 始终隐藏的字段（内部使用/已废弃）
 */
const ALWAYS_HIDDEN_FIELDS = [
  'managedFields', // 托管字段（已废弃）
  'resourceVersion', // 资源版本（内部使用，频繁变化）
  'uid', // 唯一标识符（内部使用）
  'selfLink', // 自链接（已废弃）
  'clusterName', // 集群名称（通常为空）
  'generation', // 代次（内部使用）
];

/**
 * 默认隐藏但可切换显示的字段（只读/高级功能）
 */
const DEFAULT_HIDDEN_FIELDS = [
  'status', // 状态信息（只读）
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
    // 始终隐藏的字段
    if (ALWAYS_HIDDEN_FIELDS.includes(key)) continue;

    // 根据选项隐藏字段
    if (!options.showStatus && key === 'status') continue;

    filtered[key] = filterHiddenFields(value, options);
  }
  return filtered;
};

/**
 * YAML Tab - 查看/编辑 Pod YAML
 * 功能：
 * - 语法高亮显示
 * - 行号显示
 * - 复制功能
 * - 编辑模式
 * - Diff 对比
 * - 保存/应用
 * - 自动过滤无用字段
 */
export const YamlTab: React.FC<YamlTabProps> = ({ namespace, name, pod }) => {
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

  // 加载 YAML
  const loadYaml = useCallback(async () => {
    setLoading(true);
    setError(null);

    try {
      const response = await authFetch(`/api/pods/${namespace}/${name}/yaml`);
      const result = await response.json();

      if (result.code === 0 && result.data) {
        const yaml = typeof result.data === 'string' ? result.data : jsyaml.dump(result.data);

        setYamlContent(yaml);
        setOriginalYaml(yaml);
      } else {
        setError(result.message || '加载 YAML 失败');
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : '加载失败');
    } finally {
      setLoading(false);
    }
  }, [namespace, name]);

  // 初始加载
  useEffect(() => {
    if (pod) {
      // 过滤掉不需要的字段
      const filteredPod = filterHiddenFields(pod, displayOptions);

      // 确保包含完整的 TypeMeta 字段
      const podWithMeta = {
        apiVersion: 'v1',
        kind: 'Pod',
        ...filteredPod,
      };
      const yaml = jsyaml.dump(podWithMeta, {
        indent: 2,
        lineWidth: -1, // 不限制行宽
        noRefs: true, // 不使用引用
        quotingType: '"',
        forceQuotes: false,
      });
      setYamlContent(yaml);
      setOriginalYaml(yaml);
    } else {
      loadYaml();
    }
  }, [pod, loadYaml, displayOptions]);

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
    } catch (err) {
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
      const response = await authFetch(`/api/pods/${namespace}/${name}/yaml`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ yaml: parsed }),
      });
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
  }, [namespace, name, yamlContent]);

  // 应用更改
  const handleApply = useCallback(async () => {
    if (!window.confirm('确定要应用此 YAML 配置到集群吗？')) return;

    setLoading(true);
    try {
      const parsed = jsyaml.load(yamlContent);
      const response = await authFetch(`/api/pods/${namespace}/${name}`, {
        method: 'PUT',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(parsed),
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
  }, [namespace, name, yamlContent]);

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
          <span className="yaml-title">Pod YAML</span>
          <span className="yaml-title-separator">|</span>
          <span className="yaml-status-inline">
            Lines: {yamlContent.split('\n').length} | Chars: {yamlContent.length}
          </span>
          {editing && <span className="yaml-editing-badge">Editing</span>}
        </div>
        <div className="yaml-toolbar-actions">
          {/* 显示选项切换 - Toggle Switch */}
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
                  primaryBackgroundColor: '#fff',
                  secondaryBackgroundColor: '#fafafa',
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
