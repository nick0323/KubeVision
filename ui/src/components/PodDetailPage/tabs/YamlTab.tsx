import React, { useState, useCallback, useMemo } from 'react';
import { YamlTabProps } from '../types';
import { LoadingSpinner } from '../../LoadingSpinner';
import { ErrorDisplay } from '../../ErrorDisplay';
import { authFetch } from '../../../utils/auth';
import jsyaml from 'js-yaml';
import './YamlTab.css';

/**
 * YAML Tab - 查看/编辑 Pod YAML
 */
export const YamlTab: React.FC<YamlTabProps> = ({ namespace, name, pod }) => {
  const [yamlContent, setYamlContent] = useState<string>('');
  const [originalYaml, setOriginalYaml] = useState<string>('');
  const [editing, setEditing] = useState(false);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState<string | null>(null);
  const [showDiff, setShowDiff] = useState(false);

  // 加载 YAML
  const loadYaml = useCallback(async () => {
    setLoading(true);
    setError(null);
    
    try {
      const response = await authFetch(`/api/pods/${namespace}/${name}/yaml`);
      const result = await response.json();
      
      if (result.code === 0 && result.data) {
        const yaml = typeof result.data === 'string' 
          ? result.data 
          : jsyaml.dump(result.data);
        
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
  React.useEffect(() => {
    if (pod) {
      const yaml = jsyaml.dump(pod);
      setYamlContent(yaml);
      setOriginalYaml(yaml);
    } else {
      loadYaml();
    }
  }, [pod, loadYaml]);

  // 复制 YAML
  const handleCopy = useCallback(() => {
    navigator.clipboard.writeText(yamlContent);
    alert('YAML 已复制到剪贴板');
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
      <div className="yaml-toolbar">
        <span className="yaml-title">Pod YAML</span>
        <div className="yaml-toolbar-actions">
          {!editing ? (
            <>
              <button className="toolbar-btn" onClick={handleCopy} title="复制">
                📋 Copy
              </button>
              <button className="toolbar-btn" onClick={handleEdit} title="编辑">
                ✏️ Edit
              </button>
            </>
          ) : (
            <>
              <button className="toolbar-btn" onClick={toggleDiff} title="对比">
                ↔️ Diff
              </button>
              <button className="toolbar-btn" onClick={handleSave} title="保存">
                💾 Save
              </button>
              <button className="toolbar-btn primary" onClick={handleApply} title="应用">
                🚀 Apply
              </button>
              <button className="toolbar-btn danger" onClick={handleCancel} title="取消">
                ✖ Cancel
              </button>
            </>
          )}
        </div>
      </div>

      <div className="yaml-editor-container">
        {editing ? (
          <textarea
            className="yaml-editor"
            value={yamlContent}
            onChange={(e) => setYamlContent(e.target.value)}
            spellCheck={false}
          />
        ) : (
          <pre className="yaml-editor" contentEditable={false}>
            <code>{yamlContent}</code>
          </pre>
        )}
      </div>

      {showDiff && editing && (
        <div className="diff-container">
          <div className="diff-header">YAML Diff</div>
          <div className="diff-view">
            {/* 简化的 Diff 展示 */}
            <pre className="diff-content">{yamlContent}</pre>
          </div>
        </div>
      )}
    </div>
  );
};

export default YamlTab;
