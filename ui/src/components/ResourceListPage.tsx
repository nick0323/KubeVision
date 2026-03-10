import React, { useState, useEffect, useCallback, useRef, useMemo } from 'react';
import { useNavigate } from 'react-router-dom';
import Tippy from '@tippyjs/react';
import 'tippy.js/dist/tippy.css';
import 'tippy.js/themes/light.css';
import { Column, PageConfig } from '../types';
import { StatusBadge } from './StatusBadge';
import Pagination from '../Pagination';
import LoadingSpinner from './LoadingSpinner';
import { authFetch } from '../utils/auth';
import PageHeader from './PageHeader';
import SearchInput from './SearchInput';
import NamespaceSelect from './NamespaceSelect';
import RefreshButton from './RefreshButton';
import './ResourceListPage.css';

// 单元格组件（判断是否显示 Tooltip）
const TableCell: React.FC<{
  content: React.ReactNode;
  text: string;
}> = ({ content, text }) => {
  const [showTooltip, setShowTooltip] = useState(false);
  const spanRef = useRef<HTMLSpanElement>(null);

  useEffect(() => {
    // 判断是否显示省略号（scrollWidth > clientWidth 表示内容溢出）
    if (spanRef.current) {
      setShowTooltip(spanRef.current.scrollWidth > spanRef.current.clientWidth);
    }
  }, [text]);

  // 确保 content 是字符串或 React 元素
  const displayContent = Array.isArray(content) ? content.join(', ') : content;

  const cellElement = (
    <span
      ref={spanRef}
      className="table-cell-text"
    >
      {displayContent}
    </span>
  );

  // 只有长内容才显示 Tooltip
  if (showTooltip) {
    return (
      <Tippy
        content={text}
        theme="light"
        placement="top"
        arrow={true}
        maxWidth={600}
        duration={200}
        allowHTML={false}
        interactive={false}
      >
        {cellElement}
      </Tippy>
    );
  }

  return cellElement;
};

interface ResourceListPageProps {
  config: PageConfig;
  collapsed: boolean;
  onToggleCollapsed: () => void;
}

/**
 * 通用资源列表页面组件 - 独立详情页版本
 */
export const ResourceListPage: React.FC<ResourceListPageProps> = ({ config, collapsed, onToggleCollapsed }) => {
  const navigate = useNavigate();
  const [data, setData] = useState<any[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [total, setTotal] = useState(0);
  const [page, setPage] = useState(1);
  const [pageSize, setPageSize] = useState(20);
  const [search, setSearch] = useState('');
  const [namespace, setNamespace] = useState('');
  const [sortField, setSortField] = useState<string>(config.defaultSort?.field || 'name');
  const [sortOrder, setSortOrder] = useState<'asc' | 'desc'>(config.defaultSort?.order || 'asc');
  const [namespaces, setNamespaces] = useState<string[]>([]);

  // 使用 ref 保存状态
  const pageRef = useRef(page);
  const pageSizeRef = useRef(pageSize);
  const namespaceRef = useRef(namespace);
  const searchRef = useRef(search);
  const sortFieldRef = useRef(sortField);
  const sortOrderRef = useRef(sortOrder);

  // 更新 ref
  useEffect(() => {
    pageRef.current = page;
    pageSizeRef.current = pageSize;
    namespaceRef.current = namespace;
    searchRef.current = search;
    sortFieldRef.current = sortField;
    sortOrderRef.current = sortOrder;
  }, [page, pageSize, namespace, search, sortField, sortOrder]);

  // 加载数据
  const loadData = useCallback(async () => {
    setLoading(true);
    setError(null);

    try {
      const params = new URLSearchParams();
      if (namespaceRef.current) params.set('namespace', namespaceRef.current);
      if (searchRef.current) params.set('search', searchRef.current);
      params.set('limit', pageSizeRef.current.toString());
      params.set('offset', ((pageRef.current - 1) * pageSizeRef.current).toString());
      params.set('sortBy', sortFieldRef.current);
      params.set('sortOrder', sortOrderRef.current);

      const response = await authFetch(`${config.apiEndpoint}?${params}`);
      const result = await response.json();

      if (result.code === 0 && result.data) {
        setData(result.data || []);
        setTotal(result.page?.total || result.data?.length || 0);
      } else {
        setError(result.message || '加载失败');
      }
    } catch (err) {
      setError(err instanceof Error ? err.message : '网络错误');
    } finally {
      setLoading(false);
    }
  }, [config.apiEndpoint]);

  // 加载命名空间列表（只加载一次）
  const loadNamespaces = useCallback(async () => {
    if (!config.namespaceFilter) return;
    if (namespaces.length > 0) return; // 已经加载过就不再加载

    try {
      const response = await authFetch('/api/namespaces?limit=1000&offset=0');
      const result = await response.json();
      if (result.code === 0 && result.data) {
        const nsList = Array.isArray(result.data) ? result.data : [];
        setNamespaces(nsList.map((ns: any) => ns.name));
      }
    } catch (err) {
      console.error('加载命名空间失败:', err);
    }
  }, [config.namespaceFilter, namespaces.length]);

  // 统一的 useEffect - 监听所有需要触发数据加载的变化
  useEffect(() => {
    loadData();
  }, [page, pageSize, namespace, sortField, sortOrder]);

  // 加载命名空间（只一次）
  useEffect(() => {
    loadNamespaces();
  }, [loadNamespaces]);

  // 处理排序
  const handleSort = useCallback((field: string) => {
    const currentField = sortFieldRef.current;
    const currentOrder = sortOrderRef.current;
    
    let newOrder: 'asc' | 'desc' = 'asc';
    
    if (field === currentField) {
      // 同一字段，切换顺序
      newOrder = currentOrder === 'asc' ? 'desc' : 'asc';
    }
    
    // 立即更新 ref
    sortFieldRef.current = field;
    sortOrderRef.current = newOrder;
    
    // 更新状态触发渲染
    setSortField(field);
    setSortOrder(newOrder);
    setPage(1);
    // useEffect 会自动触发 loadData
  }, []);

  // 处理搜索
  const handleSearchChange = useCallback((e: React.ChangeEvent<HTMLInputElement>) => {
    setSearch(e.target.value);
  }, []);

  const handleSearchSubmit = useCallback(() => {
    setPage(1);
    searchRef.current = search; // 更新 ref
    loadData(); // 手动触发加载
  }, [search, loadData]);

  const handleClearSearch = useCallback(() => {
    setSearch('');
    setPage(1);
    searchRef.current = ''; // 清空 ref
    loadData(); // 触发加载
  }, [loadData]);

  // 处理命名空间过滤
  const handleNamespaceChange = useCallback((value: string) => {
    setNamespace(value);
    namespaceRef.current = value; // 更新 ref
    setPage(1);
    // useEffect 会自动触发 loadData
  }, []);

  // 处理分页
  const handlePageChange = useCallback((newPage: number) => {
    setPage(newPage);
    // useEffect 会自动触发 loadData
  }, []);

  // 处理每页数量
  const handlePageSizeChange = useCallback((newSize: number) => {
    setPageSize(newSize);
    // useEffect 会自动触发 loadData
  }, []);

  // 刷新数据
  const handleRefresh = useCallback(() => {
    loadData();
  }, [loadData]);

  // 查看详情（导航到独立详情页）
  const handleViewDetail = useCallback((record: any) => {
    const ns = record.namespace || '';
    navigate(`/${config.resourceType}/${ns}/${record.name}`);
  }, [config.resourceType, navigate]);

  // 列配置
  const columns = useMemo(() => {
    return config.columns.map((col: any) => ({
      ...col,
      sortable: col.sortable !== false,
      render: col.render || ((val: any) => val),
    }));
  }, [config.columns]);

  // 状态字段检测
  const statusColumn = useMemo(() => {
    return columns.find(
      (col: any) =>
        col.dataIndex === 'status' ||
        col.dataIndex === 'Status' ||
        col.dataIndex === 'state'
    );
  }, [columns]);

  // 渲染单元格
  const renderCell = useCallback((record: any, column: Column<any>) => {
    const value = record[column.dataIndex];
    let content: React.ReactNode;
    let text: string;

    // 调试日志
    if (column.dataIndex === 'ports' || column.dataIndex === 'hosts' || column.dataIndex === 'keys' || column.dataIndex === 'role') {
      // 移除调试日志
    }

    // 状态列使用 StatusBadge
    if (statusColumn && column.dataIndex === statusColumn.dataIndex) {
      return <StatusBadge status={value} resourceType={config.resourceType} />;
    }

    // 使用自定义渲染函数
    if (column.render) {
      content = column.render(value, record, 0);
      text = String(value);
      return <TableCell content={content} text={text} />;
    }

    // 数组显示 - 处理多种情况
    if (Array.isArray(value)) {
      text = value.join(', ');
      content = text;
      return <TableCell content={content} text={text} />;
    }

    // 对象显示
    if (typeof value === 'object' && value !== null) {
      text = JSON.stringify(value);
      content = text;
      return <TableCell content={content} text={text} />;
    }

    // 普通文本
    text = String(value);
    content = text;
    return <TableCell content={content} text={text} />;
  }, [statusColumn, config.resourceType]);

  // 渲染表头内容（不包裹 th）
  const renderHeaderContent = useCallback((column: any) => {
    const isSorted = sortField === column.dataIndex;
    const sortIcon = isSorted ? (sortOrder === 'asc' ? ' ↑' : ' ↓') : '';

    return (
      <div
        className={`table-header-cell ${column.sortable ? 'sortable' : ''}`}
        onClick={() => {
          if (column.sortable) {
            handleSort(column.dataIndex);
          }
        }}
      >
        <span>{column.title}</span>
        {column.sortable && <span className="sort-icon">{sortIcon}</span>}
      </div>
    );
  }, [sortField, sortOrder, handleSort]);

  if (loading) {
    return (
      <div className="resource-list-page">
        <LoadingSpinner text={`加载${config.title}...`} size="lg" overlay />
      </div>
    );
  }

  if (error) {
    return (
      <div className="resource-list-page">
        <div className="error-container">
          <h3>加载失败</h3>
          <p>{error}</p>
          <button onClick={loadData} className="retry-btn">
            重试
          </button>
        </div>
      </div>
    );
  }

  return (
    <div className="resource-list-page">
      {/* 页面头部 */}
      <PageHeader
        title={config.title}
        collapsed={collapsed}
        onToggleCollapsed={onToggleCollapsed}
      >
        {/* 命名空间选择 */}
        {config.namespaceFilter && (
          <NamespaceSelect
            value={namespace}
            onChange={handleNamespaceChange}
            placeholder="全部命名空间"
            options={namespaces}
          />
        )}

        {/* 搜索框 */}
        <SearchInput
          placeholder={`搜索 ${config.title}...`}
          value={search}
          onChange={handleSearchChange}
          onSubmit={handleSearchSubmit}
          onClear={handleClearSearch}
          isSearching={loading}
          hasSearchResults={search.length > 0}
        />

        {/* 刷新按钮 */}
        <RefreshButton onClick={handleRefresh} loading={loading} />
      </PageHeader>

      {/* 数据表格 */}
      <div className="table-container">
        <table className="resource-table">
          <thead>
            <tr>
              {columns.map((column: Column<any>, index: number) => (
                <th
                  key={index}
                  style={{ width: column.width }}
                  className={sortField === column.dataIndex ? 'sorted' : ''}
                >
                  {renderHeaderContent(column)}
                </th>
              ))}
            </tr>
          </thead>
          <tbody>
            {data.length === 0 ? (
              <tr className="empty-row">
                <td colSpan={columns.length}>
                  <div className="empty-state">
                    <span className="empty-icon">📭</span>
                    <p>暂无数据</p>
                  </div>
                </td>
              </tr>
            ) : (
              data.map((record: any, rowIndex: number) => (
                <tr
                  key={rowIndex}
                  className="table-row"
                  onClick={() => handleViewDetail(record)}
                  style={{ cursor: 'pointer' }}
                >
                  {columns.map((column: Column<any>, colIndex: number) => (
                    <td key={colIndex} className="table-cell">
                      {renderCell(record, column)}
                    </td>
                  ))}
                </tr>
              ))
            )}
          </tbody>
        </table>
      </div>

      {/* 分页 */}
      <Pagination
        currentPage={page}
        total={total}
        pageSize={pageSize}
        onPageChange={handlePageChange}
        onPageSizeChange={handlePageSizeChange}
        pageSizeOptions={[20, 50, 100, 500]}
      />
    </div>
  );
};

export default ResourceListPage;
