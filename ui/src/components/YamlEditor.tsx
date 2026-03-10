import React, { useState, useEffect } from 'react';
import jsYaml from 'js-yaml';
import './YamlEditor.css';

interface YamlEditorProps {
  data: any;
  readOnly?: boolean;
  onEdit?: (yaml: string) => void;
  onSave?: (yaml: string) => void;
  onCompare?: () => void;
  onRefresh?: () => void;
}

export const YamlEditor: React.FC<YamlEditorProps> = ({
  data,
  readOnly = false,
  onEdit,
  onSave,
  onCompare,
  onRefresh,
}) => {
  const [yamlString, setYamlString] = useState('');
  const [originalYaml, setOriginalYaml] = useState('');
  const [isEditing, setIsEditing] = useState(false);
  const [hasChanges, setHasChanges] = useState(false);
  const [error, setError] = useState<string | null>(null);

  // 初始化 YAML 字符串
  useEffect(() => {
    if (data) {
      try {
        const yamlStr = jsYaml.dump(data, {
          indent: 2,
          lineWidth: -1,
          noRefs: true,
        });
        setYamlString(yamlStr);
        if (!isEditing) {
          setOriginalYaml(yamlStr);
        }
        setError(null);
      } catch (err) {
        setError('YAML 格式化失败');
      }
    }
  }, [data, isEditing]);

  // 处理编辑模式切换
  const handleEditToggle = () => {
    if (isEditing) {
      // 取消编辑，恢复原始 YAML
      setYamlString(originalYaml);
      setHasChanges(false);
    }
    setIsEditing(!isEditing);
  };

  // 处理 YAML 变化
  const handleYamlChange = (e: React.ChangeEvent<HTMLTextAreaElement>) => {
    setYamlString(e.target.value);
    setHasChanges(e.target.value !== originalYaml);
    if (onEdit) {
      onEdit(e.target.value);
    }
  };

  // 保存修改
  const handleSave = () => {
    try {
      // 验证 YAML 格式
      jsYaml.load(yamlString);
      
      if (onSave) {
        onSave(yamlString);
      }
      
      setOriginalYaml(yamlString);
      setHasChanges(false);
      setIsEditing(false);
    } catch (err) {
      setError('YAML 格式错误，请检查');
    }
  };

  // 对比差异
  const handleCompare = () => {
    if (onCompare) {
      onCompare();
    }
  };

  // 刷新
  const handleRefresh = () => {
    if (onRefresh) {
      onRefresh();
    }
  };

  return (
    <div className="yaml-editor">
      {/* 工具栏 */}
      <div className="yaml-toolbar">
        <div className="yaml-actions">
          {!readOnly && (
            <>
              <button
                className={`btn ${isEditing ? 'btn-primary' : 'btn-default'}`}
                onClick={handleEditToggle}
              >
                {isEditing ? '✕ 取消' : '✏️ 编辑'}
              </button>
              {isEditing && (
                <>
                  <button
                    className={`btn ${hasChanges ? 'btn-success' : 'btn-default'}`}
                    onClick={handleSave}
                    disabled={!hasChanges}
                  >
                    💾 保存
                  </button>
                  <button
                    className="btn btn-default"
                    onClick={handleCompare}
                  >
                    📊 对比
                  </button>
                </>
              )}
            </>
          )}
          <button
            className="btn btn-default"
            onClick={handleRefresh}
          >
            🔄 刷新
          </button>
        </div>
        {hasChanges && isEditing && (
          <div className="yaml-changes-indicator">
            <span className="dot"></span>
            <span>未保存的更改</span>
          </div>
        )}
      </div>

      {/* 错误提示 */}
      {error && (
        <div className="yaml-error">
          <span className="error-icon">⚠️</span>
          <span>{error}</span>
        </div>
      )}

      {/* YAML 内容区 */}
      <div className="yaml-content">
        {isEditing ? (
          <textarea
            className="yaml-textarea"
            value={yamlString}
            onChange={handleYamlChange}
            spellCheck={false}
          />
        ) : (
          <pre className="yaml-pre">
            <code>{yamlString}</code>
          </pre>
        )}
      </div>
    </div>
  );
};

export default YamlEditor;
