import React, { useMemo, useCallback } from 'react';
import { useNavigate } from 'react-router-dom';
import { PageConfig } from '../types';
import { StatusBadge } from '../components/ui/StatusBadge';
import {
  TableHeaderCell,
  TableRow,
  TableColumn,
  TableSkeleton,
  EmptyState,
  ErrorState,
} from '../components/ui/Table/index.tsx';
import PageHeader from '../components/common/PageHeader.tsx';
import SearchInput from '../components/common/SearchInput.tsx';
import NamespaceSelect from '../components/common/NamespaceSelect.tsx';
import RefreshButton from '../components/common/RefreshButton.tsx';
import Pagination from '../components/common/Pagination.tsx';
import { useResourceList } from '../hooks/useResourceList';
import { useConfirm } from '../hooks/useConfirm';
import './ResourceListPage.css';

/**
 * 资源列表页面属性
 */
interface ResourceListPageProps {
  config: PageConfig;
  collapsed: boolean;
  onToggleCollapsed: () => void;
  onRowClick?: (record: any) => void;
  actions?: Array<{
    label: string;
    icon?: React.ReactNode;
    onClick: (record: any) => void | Promise<void>;
    confirm?: boolean;
    confirmMessage?: string;
    disabled?: (record: any) => boolean;
    permission?: string;
    danger?: boolean;
  }>;
}

/**
 * 状态列标识
 */
const STATUS_COLUMN_KEYS = ['status', 'Status', 'state', 'Phase'] as const;

/**
 * 通用资源列表页面组件（增强版）
 *
 * 特性：
 * - SWR 缓存模式，避免重复请求
 * - 搜索防抖（300ms）
 * - 骨架屏加载状态
 * - 友好的空状态和错误状态
 * - 支持行点击和操作按钮
 * - 操作确认对话框
 * - 权限校验
 */
export const ResourceListPage: React.FC<ResourceListPageProps> = ({
  config,
  collapsed,
  onToggleCollapsed,
  onRowClick,
  actions = [],
}) => {
  const navigate = useNavigate();

  // 使用增强版 hook 管理所有列表逻辑
  const {
    // 数据状态
    data,
    loading,
    isValidating,
    error,
    total,

    // 分页状态
    page,
    pageSize,
    setPage,
    setPageSize,

    // 过滤状态
    namespace,
    search,
    setNamespace,
    setSearch,

    // 排序状态
    sortField,
    sortOrder,
    handleSort,

    // 操作
    refresh,
    handleSubmit,
    clearSearch,

    // 命名空间列表
    namespaces,
    namespacesLoading,
  } = useResourceList({
    apiEndpoint: config.apiEndpoint,
    namespaceFilter: config.namespaceFilter,
    defaultSort: config.defaultSort,
    initialPageSize: 20,
    staleTime: 30000, // 30 秒缓存
  });

  /**
   * 点击资源名称跳转详情页
   */
  const handleNameClick = useCallback(
    (record: any) => {
      const resourceType = config.resourceType;
      const ns = record.namespace || namespace || 'default';
      const name = record.name;

      if (resourceType && name) {
        navigate(`/${resourceType}/${ns}/${name}`);
      }
    },
    [config.resourceType, namespace, navigate]
  );

  // 确认对话框
  const { confirm, confirming, config: confirmConfig, onConfirm, onCancel } = useConfirm();

  /**
   * 检测状态列
   */
  const statusColumnIndex = useMemo(() => {
    return config.columns.find(col =>
      STATUS_COLUMN_KEYS.includes(col.dataIndex as (typeof STATUS_COLUMN_KEYS)[number])
    )?.dataIndex;
  }, [config.columns]);

  /**
   * 格式化时间
   */
  const formatTime = (timestamp?: string) => {
    if (!timestamp) return '-';
    try {
      const date = new Date(timestamp);
      const year = date.getFullYear();
      const month = String(date.getMonth() + 1).padStart(2, '0');
      const day = String(date.getDate()).padStart(2, '0');
      const hours = String(date.getHours()).padStart(2, '0');
      const minutes = String(date.getMinutes()).padStart(2, '0');
      const seconds = String(date.getSeconds()).padStart(2, '0');
      const formatted = `${year}-${month}-${day} ${hours}:${minutes}:${seconds}`;
      return formatted;
    } catch {
      return timestamp;
    }
  };

  /**
   * 列配置（添加默认值，name 列添加点击事件）
   */
  const columns = useMemo<TableColumn[]>(
    () =>
      config.columns.map(col => {
        const isNameColumn = col.dataIndex === 'name' || col.dataIndex === 'Name';
        const isTimeColumn =
          col.dataIndex === 'lastSeen' ||
          col.dataIndex === 'firstSeen' ||
          col.dataIndex === 'LastSeen' ||
          col.dataIndex === 'FirstSeen';

        return {
          ...col,
          sortable: col.sortable ?? true,
          render:
            col.render ||
            (isNameColumn
              ? (value: any, record: any) => (
                  <span
                    className="resource-name-link"
                    onClick={e => {
                      e.stopPropagation();
                      handleNameClick(record);
                    }}
                  >
                    {value}
                  </span>
                )
              : isTimeColumn
                ? (value: any) => formatTime(value)
                : undefined),
        } as TableColumn;
      }),
    [config.columns, handleNameClick]
  );

  /**
   * 操作列
   */
  const actionColumn = useMemo<TableColumn | null>(() => {
    if (actions.length === 0) return null;

    return {
      title: '操作',
      dataIndex: '__actions__',
      width: '120px',
      sortable: false,
      render: (_, record) => (
        <div className="table-actions">
          {actions.map((action, index) => {
            const disabled = action.disabled?.(record) ?? false;

            return (
              <button
                key={index}
                className={`action-btn ${action.danger ? 'danger' : ''}`}
                onClick={e => {
                  e.stopPropagation();
                  handleActionClick(action, record);
                }}
                disabled={disabled}
                title={action.label}
              >
                {action.icon || action.label}
              </button>
            );
          })}
        </div>
      ),
    };
  }, [actions]);

  /**
   * 所有列（包含操作列）
   */
  const allColumns = useMemo(() => {
    return actionColumn ? [...columns, actionColumn] : columns;
  }, [columns, actionColumn]);

  /**
   * 处理操作点击
   */
  const handleActionClick = useCallback(
    async (action: NonNullable<ResourceListPageProps['actions']>[0], record: any) => {
      if (action.confirm) {
        const result = await confirm({
          message: action.confirmMessage || `确定要执行此操作吗？`,
          danger: action.danger,
        });

        if (!result.confirmed) return;
      }

      try {
        await action.onClick(record);
        // 操作成功后刷新数据
        await refresh();
      } catch (err) {
        console.error('操作失败:', err);
      }
    },
    [confirm, refresh]
  );

  // 事件处理器
  const handleNamespaceChange = useCallback(
    (value: string) => {
      setNamespace(value);
      setPage(1);
    },
    [setNamespace, setPage]
  );

  const handleSearchChange = useCallback(
    (e: React.ChangeEvent<HTMLInputElement>) => {
      setSearch(e.target.value);
    },
    [setSearch]
  );

  const handlePageChange = useCallback(
    (newPage: number) => {
      setPage(newPage);
    },
    [setPage]
  );

  const handlePageSizeChange = useCallback(
    (newSize: number) => {
      setPageSize(newSize);
      setPage(1);
    },
    [setPageSize, setPage]
  );

  const handleRowClick = useCallback(
    (record: any) => {
      onRowClick?.(record);
    },
    [onRowClick]
  );

  // 渲染加载状态（骨架屏）
  if (loading && data.length === 0) {
    return (
      <div className="resource-list-page">
        <PageHeader
          title={config.title}
          collapsed={collapsed}
          onToggleCollapsed={onToggleCollapsed}
        >
          {config.namespaceFilter && (
            <NamespaceSelect
              value={namespace}
              onChange={handleNamespaceChange}
              placeholder="All Namespace"
              options={[]}
              disabled
            />
          )}
          <SearchInput
            placeholder={`搜索 ${config.title}...`}
            value={search}
            onChange={handleSearchChange}
            disabled
          />
          <RefreshButton onClick={refresh} loading />
        </PageHeader>

        <div className="table-container">
          <table className="resource-table">
            <thead>
              <tr>
                {columns.map((column, index) => (
                  <th key={index} style={{ width: column.width }}>
                    <div className="table-header-cell">
                      <span>{column.title}</span>
                    </div>
                  </th>
                ))}
              </tr>
            </thead>
            <tbody>
              <TableSkeleton columns={columns.length} rows={10} />
            </tbody>
          </table>
        </div>
      </div>
    );
  }

  // 渲染错误状态
  if (error && data.length === 0) {
    return (
      <div className="resource-list-page">
        <PageHeader
          title={config.title}
          collapsed={collapsed}
          onToggleCollapsed={onToggleCollapsed}
        >
          {config.namespaceFilter && (
            <NamespaceSelect
              value={namespace}
              onChange={handleNamespaceChange}
              placeholder="All Namespace"
              options={namespaces}
              disabled={namespacesLoading}
            />
          )}
          <SearchInput
            placeholder={`搜索 ${config.title}...`}
            value={search}
            onChange={handleSearchChange}
            onSubmit={handleSubmit}
            onClear={clearSearch}
          />
          <RefreshButton onClick={refresh} loading={loading || isValidating} />
        </PageHeader>

        <div className="table-container">
          <table className="resource-table">
            <tbody>
              <ErrorState
                message={error}
                colSpan={columns.length}
                onRetry={refresh}
                retryText="重新加载"
              />
            </tbody>
          </table>
        </div>
      </div>
    );
  }

  return (
    <div className="resource-list-page">
      {/* 页面头部 */}
      <PageHeader title={config.title} collapsed={collapsed} onToggleCollapsed={onToggleCollapsed}>
        {/* 命名空间选择 */}
        {config.namespaceFilter && (
          <NamespaceSelect
            value={namespace}
            onChange={handleNamespaceChange}
            placeholder="All Namespace"
            options={namespaces}
            disabled={namespacesLoading}
          />
        )}

        {/* 搜索框 */}
        <SearchInput
          placeholder={`搜索 ${config.title}...`}
          value={search}
          onChange={handleSearchChange}
          onSubmit={handleSubmit}
          onClear={clearSearch}
          isSearching={isValidating}
          hasSearchResults={search.length > 0 && data.length > 0}
        />

        {/* 刷新按钮 */}
        <RefreshButton
          onClick={refresh}
          loading={isValidating}
          showLastUpdated={!loading && data.length > 0}
        />
      </PageHeader>

      {/* 数据表格 */}
      <div className="table-container">
        <table className="resource-table">
          <thead>
            <tr>
              {allColumns.map((column, index) => (
                <TableHeaderCell
                  key={index}
                  column={column}
                  sortField={sortField}
                  sortOrder={sortOrder}
                  onSort={handleSort}
                />
              ))}
            </tr>
          </thead>
          <tbody>
            {data.length === 0 ? (
              <EmptyState
                message="暂无数据"
                description={search ? '尝试调整搜索条件' : undefined}
                colSpan={allColumns.length}
                action={
                  search && (
                    <button onClick={clearSearch} className="clear-search-btn">
                      清空搜索
                    </button>
                  )
                }
              />
            ) : (
              <>
                {data.map((record, rowIndex) => (
                  <TableRow
                    key={rowIndex}
                    record={record}
                    rowIndex={rowIndex}
                    columns={allColumns}
                    isStatusColumn={!!statusColumnIndex}
                    statusColumnIndex={statusColumnIndex as string | undefined}
                    StatusBadge={StatusBadge}
                    resourceType={config.resourceType}
                    onClick={handleRowClick}
                  />
                ))}
                {/* 刷新时的骨架屏行 */}
                {isValidating && <TableSkeleton columns={allColumns.length} rows={2} />}
              </>
            )}
          </tbody>
        </table>
      </div>

      {/* 分页 */}
      {data.length > 0 && (
        <Pagination
          currentPage={page}
          total={total}
          pageSize={pageSize}
          onPageChange={handlePageChange}
          onPageSizeChange={handlePageSizeChange}
          pageSizeOptions={[20, 50, 100, 500]}
          showQuickJumper
        />
      )}

      {/* 确认对话框 */}
      {confirming && confirmConfig && (
        <div className="confirm-modal-overlay">
          <div className={`confirm-modal ${confirmConfig.danger ? 'danger' : ''}`}>
            <div className="confirm-modal-header">
              <span className="confirm-icon">{confirmConfig.danger ? '⚠️' : 'ℹ️'}</span>
              <h4>{confirmConfig.title || '确认操作'}</h4>
            </div>
            <div className="confirm-modal-body">{confirmConfig.message}</div>
            <div className="confirm-modal-footer">
              <button className="confirm-btn cancel" onClick={onCancel}>
                {confirmConfig.cancelText || '取消'}
              </button>
              <button
                className={`confirm-btn confirm ${confirmConfig.danger ? 'danger' : ''}`}
                onClick={onConfirm}
              >
                {confirmConfig.confirmText || '确认'}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
};

export default ResourceListPage;
