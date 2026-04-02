import { useCallback, useState } from 'react';

/**
 * 确认对话框配置
 */
export interface ConfirmOptions {
  title?: string;
  message: string;
  confirmText?: string;
  cancelText?: string;
  danger?: boolean;
}

/**
 * 确认结果
 */
export interface ConfirmResult {
  confirmed: boolean;
  cancelled: boolean;
}

/**
 * 操作确认 Hook
 *
 * 用法：
 * const { confirm, confirming, confirmConfig } = useConfirm();
 *
 * const handleDelete = async (record) => {
 *   const result = await confirm({
 *     title: '删除确认',
 *     message: `确定要删除 ${record.name} 吗？`,
 *     danger: true,
 *   });
 *
 *   if (result.confirmed) {
 *     // 执行删除
 *   }
 * };
 */
export function useConfirm() {
  const [confirming, setConfirming] = useState(false);
  const [config, setConfig] = useState<ConfirmOptions | null>(null);
  const [resolver, setResolver] = useState<((result: ConfirmResult) => void) | null>(null);

  /**
   * 显示确认对话框
   */
  const confirm = useCallback((options: ConfirmOptions): Promise<ConfirmResult> => {
    return new Promise(resolve => {
      setConfig(options);
      setConfirming(true);
      setResolver(() => resolve);
    });
  }, []);

  /**
   * 确认操作
   */
  const handleConfirm = useCallback(() => {
    if (resolver) {
      resolver({ confirmed: true, cancelled: false });
      setResolver(null);
    }
    setConfirming(false);
    setConfig(null);
  }, [resolver]);

  /**
   * 取消操作
   */
  const handleCancel = useCallback(() => {
    if (resolver) {
      resolver({ confirmed: false, cancelled: true });
      setResolver(null);
    }
    setConfirming(false);
    setConfig(null);
  }, [resolver]);

  return {
    confirm,
    confirming,
    config,
    onConfirm: handleConfirm,
    onCancel: handleCancel,
  };
}

/**
 * 简单的浏览器确认对话框（备用方案）
 */
export function browserConfirm(message: string): boolean {
  return window.confirm(message);
}

/**
 * 权限检查工具
 */
export interface PermissionCheck {
  hasPermission: (permission: string) => boolean;
  checkPermissions: (permissions: string[]) => boolean;
}

/**
 * 创建权限检查器
 *
 * @param userPermissions 用户拥有的权限列表
 */
export function createPermissionChecker(userPermissions: string[] = []): PermissionCheck {
  return {
    hasPermission: (permission: string) => {
      return userPermissions.includes(permission) || userPermissions.includes('*');
    },
    checkPermissions: (permissions: string[]) => {
      // 所有权限都需要满足（AND 逻辑）
      return permissions.every(p => userPermissions.includes(p) || userPermissions.includes('*'));
    },
  };
}

/**
 * 权限检查 Hook
 *
 * @param userPermissions 用户拥有的权限列表
 */
export function usePermission(userPermissions: string[] = []) {
  const checker = createPermissionChecker(userPermissions);

  /**
   * 检查是否有权限
   */
  const hasPermission = useCallback(
    (permission: string): boolean => {
      return checker.hasPermission(permission);
    },
    [checker, userPermissions]
  );

  /**
   * 检查是否有任何一个权限
   */
  const hasAnyPermission = useCallback(
    (permissions: string[]): boolean => {
      return permissions.some(p => checker.hasPermission(p));
    },
    [checker, userPermissions]
  );

  /**
   * 检查是否拥有所有权限
   */
  const hasAllPermissions = useCallback(
    (permissions: string[]): boolean => {
      return checker.checkPermissions(permissions);
    },
    [checker, userPermissions]
  );

  return {
    hasPermission,
    hasAnyPermission,
    hasAllPermissions,
    checker,
  };
}
