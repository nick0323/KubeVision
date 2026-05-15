import { useCallback, useState } from 'react';

/**
 * Confirm dialogConfig
 */
export interface ConfirmOptions {
  title?: string;
  message: string;
  confirmText?: string;
  cancelText?: string;
  danger?: boolean;
}

/**
 * confirm结果
 */
export interface ConfirmResult {
  confirmed: boolean;
  cancelled: boolean;
}

/**
 * Actionconfirm Hook
 *
 * use法：
 * const { confirm, confirming, confirmConfig } = useConfirm();
 *
 * const handleDelete = async (record) => {
 *   const result = await confirm({
 *     title: 'Deleteconfirm',
 *     message: `确定wantDelete ${record.name} 吗？`,
 *     danger: true,
 *   });
 *
 *   if (result.confirmed) {
 *     // 执rowDelete
 *   }
 * };
 */
export function useConfirm() {
  const [confirming, setConfirming] = useState(false);
  const [config, setConfig] = useState<ConfirmOptions | null>(null);
  const [resolver, setResolver] = useState<((result: ConfirmResult) => void) | null>(null);

  /**
   * DisplayConfirm dialog
   */
  const confirm = useCallback((options: ConfirmOptions): Promise<ConfirmResult> => {
    return new Promise(resolve => {
      setConfig(options);
      setConfirming(true);
      setResolver(() => resolve);
    });
  }, []);

  /**
   * Confirm Operation
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
   * CancelAction
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
 * permission检查工具
 */
export interface PermissionCheck {
  hasPermission: (permission: string) => boolean;
  checkPermissions: (permissions: string[]) => boolean;
}

/**
 * Createpermission检查器
 *
 * @param userPermissions user拥has'spermissionList
 */
export function createPermissionChecker(userPermissions: string[] = []): PermissionCheck {
  return {
    hasPermission: (permission: string) => {
      return userPermissions.includes(permission) || userPermissions.includes('*');
    },
    checkPermissions: (permissions: string[]) => {
      // 所haspermissionallneed满足（AND logic）
      return permissions.every(p => userPermissions.includes(p) || userPermissions.includes('*'));
    },
  };
}

/**
 * permission检查 Hook
 *
 * @param userPermissions user拥has'spermissionList
 */
export function usePermission(userPermissions: string[] = []) {
  const checker = createPermissionChecker(userPermissions);

  const hasPermission = useCallback(
    (permission: string): boolean => {
      return checker.hasPermission(permission);
    },
    [checker]
  );

  const hasAnyPermission = useCallback(
    (permissions: string[]): boolean => {
      return permissions.some(p => checker.hasPermission(p));
    },
    [checker]
  );

  const hasAllPermissions = useCallback(
    (permissions: string[]): boolean => {
      return checker.checkPermissions(permissions);
    },
    [checker]
  );

  return {
    hasPermission,
    hasAnyPermission,
    hasAllPermissions,
    checker,
  };
}
