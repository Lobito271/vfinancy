import type { ReactNode } from 'react';
import { usePermissionContext } from '@/hooks/usePermission';
import { hasPermission, hasAny, hasRole, type Permission } from '@/utils/permissions';
import type { Role } from '@/constants/permissions';

interface CanProps {
  permission?: Permission | Permission[];
  anyPermission?: Permission[];
  role?: Role | Role[];
  fallback?: ReactNode;
  children: ReactNode;
}

export function Can({ permission, anyPermission, role, fallback = null, children }: CanProps) {
  const ctx = usePermissionContext();

  let allowed = true;
  if (permission !== undefined) {
    allowed = allowed && hasPermission(ctx, permission);
  }
  if (anyPermission !== undefined) {
    allowed = allowed && hasAny(ctx, anyPermission);
  }
  if (role !== undefined) {
    allowed = allowed && hasRole(ctx, role);
  }

  if (!allowed) return <>{fallback}</>;
  return <>{children}</>;
}

interface PermissionGateProps {
  permission: Permission | Permission[];
  children: ReactNode;
  fallback?: ReactNode;
}

export function PermissionGate({ permission, children, fallback = null }: PermissionGateProps) {
  return (
    <Can permission={permission} fallback={fallback}>
      {children}
    </Can>
  );
}
