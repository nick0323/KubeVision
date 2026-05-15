import React, { useMemo, useCallback, useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { PageConfig } from '../types';
import { StatusBadge } from '../common/StatusBadge';
import { FaInbox } from 'react-icons/fa';
import {
  TableHeaderCell,
  TableRow,
  TableColumn,
  TableSkeleton,
  ErrorState,
} from '../common/Table';
import PageHeader from '../common/PageHeader.tsx';
import SearchInput from '../common/SearchInput.tsx';
import NamespaceSelect from '../common/NamespaceSelect.tsx';
import RefreshButton from '../common/RefreshButton.tsx';
import Pagination from '../common/Pagination.tsx';
import CreateResourceModal from '../common/CreateResourceModal';
import { getTemplateByResourceType } from '../constants/templates';
import { useResourceList } from '../hooks/useResourceList';
import { useConfirm } from '../hooks/useConfirm';
import { ConfirmModal } from '../common/ConfirmModal';
import { logError } from '../utils/errorHandler';
import { formatTime } from '../utils/time';
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

  // Create resource modal state
  const [createModalVisible, setCreateModalVisible] = useState(false);

  // Check if current resource type has a template
  const hasTemplate = useMemo(() => {
    return !!getTemplateByResourceType(config.resourceType);
  }, [config.resourceType]);

  // Use enhanced hook to manage resource list logic
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
          sortable: col.sortable ?? false,
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

  // Handle namespace change
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
        {/* Create button - only shown for resource types with templates */}
        {hasTemplate && (
          <button
            className="create-resource-btn"
            onClick={() => setCreateModalVisible(true)}
            title={`Create ${config.title}`}
          >
            + Create
          </button>
        )}

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

      {/* Data table or empty state */}
      {data.length === 0 ? (
        <div className="resource-empty">
          <div className="resource-empty-icon"><FaInbox /></div>
          <p className="resource-empty-text">No data yet</p>
          {search && <p className="resource-empty-desc">Try adjusting search criteria</p>}
          {search && <button onClick={clearSearch} className="clear-search-btn">Clear search</button>}
        </div>
      ) : (
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
          </tbody>
        </table>
      </div>
      )}

      {/* Pagination */}
      {data.length > 0 && (
        <Pagination
          currentPage={page}
          total={total}
          pageSize={pageSize}
          onPageChange={handlePageChange}
          onPageSizeChange={handlePageSizeChange}
          showQuickJumper
        />
      )}

      <ConfirmModal
        open={confirming && !!confirmConfig}
        title={confirmConfig?.title}
        message={confirmConfig?.message || ''}
        confirmText={confirmConfig?.confirmText}
        cancelText={confirmConfig?.cancelText}
        danger={confirmConfig?.danger}
        onConfirm={onConfirm}
        onCancel={onCancel}
      />

      {/* Create resource modal */}
      <CreateResourceModal
        visible={createModalVisible}
        resourceType={config.resourceType}
        namespace={namespace}
        onClose={() => setCreateModalVisible(false)}
        onSuccess={() => {
          refresh();
          setCreateModalVisible(false);
        }}
      />
    </div>
  );
};

export default ResourceListPage;
