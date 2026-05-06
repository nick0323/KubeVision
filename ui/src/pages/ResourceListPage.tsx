import React, { useMemo, useCallback } from 'react';
import { useNavigate } from 'react-router-dom';
import { PageConfig } from '../types';
import { StatusBadge } from '../common/StatusBadge';
import {
  TableHeaderCell,
  TableRow,
  TableColumn,
  TableSkeleton,
  EmptyState,
  ErrorState,
} from '../common/Table';
import PageHeader from '../common/PageHeader.tsx';
import SearchInput from '../common/SearchInput.tsx';
import NamespaceSelect from '../common/NamespaceSelect.tsx';
import RefreshButton from '../common/RefreshButton.tsx';
import Pagination from '../common/Pagination.tsx';
import { useResourceList } from '../hooks/useResourceList';
import { useConfirm } from '../hooks/useConfirm';
import { logError } from '../utils/errorHandler';
import './ResourceListPage.css';

/**
 * Resource list page面属性
 */
interface ResourceListPageProps {
  config: PageConfig;
  collapsed: boolean;
  onToggleCollapsed: () => void;
  onRowClick?: (record: Record<string, unknown>) => void;
  actions?: Array<{
    label: string;
    icon?: React.ReactNode;
    onClick: (record: Record<string, unknown>) => void | Promise<void>;
    confirm?: boolean;
    confirmMessage?: string;
    disabled?: (record: Record<string, unknown>) => boolean;
    permission?: string;
    danger?: boolean;
  }>;
}

/**
 * Status列标识
 */
const STATUS_COLUMN_KEYS = ['status', 'Status', 'state', 'Phase', 'type', 'Type'] as const;

export const ResourceListPage: React.FC<ResourceListPageProps> = ({
  config,
  collapsed,
  onToggleCollapsed,
  onRowClick,
  actions = [],
}) => {
  const navigate = useNavigate();

  // Use增强版 hook 管理所hasListlogic
  const {
    // dataStatus
    data,
    loading,
    isValidating,
    error,
    total,

    // PaginationStatus
    page,
    pageSize,
    setPage,
    setPageSize,

    // filterStatus
    namespace,
    search,
    setNamespace,
    setSearch,

    // sortStatus
    sortField,
    sortOrder,
    handleSort,

    // Action
    refresh,
    handleSubmit,
    clearSearch,

    // Namespace list
    namespaces,
    namespacesLoading,
  } = useResourceList({
    apiEndpoint: config.apiEndpoint,
    namespaceFilter: config.namespaceFilter,
    defaultSort: config.defaultSort,
    initialPageSize: 20,
    staleTime: 30000, // 30 seconds cache
  });

  /**
   * ClickResource namejump to详情页
   */
  const handleNameClick = useCallback(
    (record: Record<string, unknown>) => {
      const resourceType = config.resourceType;
      const ns = (record.namespace as string) || namespace || 'default';
      const name = record.name as string;

      if (resourceType && name) {
        navigate(`/${resourceType}/${ns}/${name}`);
      }
    },
    [config.resourceType, namespace, navigate]
  );

  // Confirm dialog
  const { confirm, confirming, config: confirmConfig, onConfirm, onCancel } = useConfirm();

  /**
   * 检测Status列
   */
  const statusColumnIndex = useMemo(() => {
    return config.columns.find(col =>
      STATUS_COLUMN_KEYS.includes(col.dataIndex as (typeof STATUS_COLUMN_KEYS)[number])
    )?.dataIndex;
  }, [config.columns]);

  /**
   * format化time
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
   * 列Config（Adddefault值，name 列AddClick事 component）
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
   * Action列
   */
  const actionColumn = useMemo<TableColumn | null>(() => {
    if (actions.length === 0) return null;

    return {
      title: 'Actions',
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
   * 所has列（includeAction列）
   */
  const allColumns = useMemo(() => {
    return actionColumn ? [...columns, actionColumn] : columns;
  }, [columns, actionColumn]);

  /**
   * ProcessActionClick
   */
  const handleActionClick = useCallback(
    async (action: NonNullable<ResourceListPageProps['actions']>[0], record: Record<string, unknown>) => {
      if (action.confirm) {
        const result = await confirm({
          message: action.confirmMessage || `Are you sure to execute this action?`,
          danger: action.danger,
        });

        if (!result.confirmed) return;
      }

      try {
        await action.onClick(record);
        // Action成功afterRefreshdata
        await refresh();
      } catch (err) {
        logError(err, 'handleActionClick');
      }
    },
    [confirm, refresh]
  );

  // 事 componentProcess器
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

  // RenderLoading...Skeleton screen）
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
            placeholder={`Search ${config.title}...`}
            value={search}
            onChange={handleSearchChange}
            disabled={true}
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

  // RenderError status
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
            placeholder={`Search ${config.title}...`}
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
                retryText="Reload"
              />
            </tbody>
          </table>
        </div>
      </div>
    );
  }

  return (
    <div className="resource-list-page">
      {/* Page header */}
      <PageHeader title={config.title} collapsed={collapsed} onToggleCollapsed={onToggleCollapsed}>
        {/* Namespace selection */}
        {config.namespaceFilter && (
          <NamespaceSelect
            value={namespace}
            onChange={handleNamespaceChange}
            placeholder="All Namespace"
            options={namespaces}
            disabled={namespacesLoading}
          />
        )}

        {/* Search box */}
        <SearchInput
          placeholder={`Search ${config.title}...`}
          value={search}
          onChange={handleSearchChange}
          onSubmit={handleSubmit}
          onClear={clearSearch}
          isSearching={isValidating}
          hasSearchResults={search.length > 0 && data.length > 0}
        />

        {/* Refresh button */}
        <RefreshButton
          onClick={refresh}
          loading={isValidating}
          showLastUpdated={!loading && data.length > 0}
        />
      </PageHeader>

      {/* Data table */}
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
                message="No data yet"
                description={search ? 'Try adjusting search criteria' : undefined}
                colSpan={allColumns.length}
                action={
                  search && (
                    <button onClick={clearSearch} className="clear-search-btn">
                      Clear search
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
                {/* Skeleton rows when refreshing */}
                {isValidating && <TableSkeleton columns={allColumns.length} rows={2} />}
              </>
            )}
          </tbody>
        </table>
      </div>

      {/* Pagination */}
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

      {/* Confirm dialog */}
      {confirming && confirmConfig && (
        <div className="confirm-modal-overlay">
          <div className={`confirm-modal ${confirmConfig.danger ? 'danger' : ''}`}>
            <div className="confirm-modal-header">
              <span className="confirm-icon">{confirmConfig.danger ? '⚠️' : 'ℹ️'}</span>
              <h4>{confirmConfig.title || 'Confirm Operation'}</h4>
            </div>
            <div className="confirm-modal-body">{confirmConfig.message}</div>
            <div className="confirm-modal-footer">
              <button className="confirm-btn cancel" onClick={onCancel}>
                {confirmConfig.cancelText || 'Cancel'}
              </button>
              <button
                className={`confirm-btn confirm ${confirmConfig.danger ? 'danger' : ''}`}
                onClick={onConfirm}
              >
                {confirmConfig.confirmText || 'Confirm'}
              </button>
            </div>
          </div>
        </div>
      )}
    </div>
  );
};

export default ResourceListPage;
