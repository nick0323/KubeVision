import React, { useState, useEffect } from 'react';
import yaml from 'js-yaml';
import './YamlViewer.css';

interface YamlViewerProps {
  data: any;
  onRefresh?: () => void;
  onEdit?: (yaml: string) => void;
  onCompare?: () => void;
}

export const YamlViewer: React.FC<YamlViewerProps> = ({
  data,
  onRefresh,
  onEdit,
  onCompare,
}) => {
  const [yamlContent, setYamlContent] = useState('');

  useEffect(() => {
    if (data) {
      try {
        const yamlStr = yaml.dump(data, {
          indent: 2,
          lineWidth: -1, // 不限制行宽
          noRefs: true, // 不使用引用
        });
        setYamlContent(yamlStr);
      } catch (err) {
        setYamlContent(JSON.stringify(data, null, 2));
      }
    }
  }, [data]);

  return (
    <div className="yaml-viewer">
      <div className="yaml-actions">
        <div className="action-buttons">
          {onRefresh && (
            <button className="btn btn-default" onClick={onRefresh}>
              🔄 刷新
            </button>
          )}
          {onEdit && (
            <button className="btn btn-primary" onClick={onEdit}>
              ✏️ 编辑
            </button>
          )}
          {onCompare && (
            <button className="btn btn-default" onClick={onCompare}>
              📊 对比差异
            </button>
          )}
        </div>
      </div>

      <div className="yaml-content">
        <pre>{yamlContent}</pre>
      </div>
    </div>
  );
};

export default YamlViewer;
