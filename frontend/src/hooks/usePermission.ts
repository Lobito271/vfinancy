import { useMemo } from 'react';
import { useSessionStore } from '@/stores/session';
import {
  buildContext,
  hasPermission,
  hasAny,
  hasRole,
  type Permission,
  type PermissionContext,
} from '@/utils/permissions';
import type { Role } from '@/constants/permissions';

export function usePermissionContext(): PermissionContext {
  const user = useSessionStore((s) => s.user);
  return useMemo(() => buildContext(user), [user]);
}

export function usePermission(permission: Permission | Permission[]): boolean {
  const ctx = usePermissionContext();
  return hasPermission(ctx, permission);
}

export function useAnyPermission(permissions: Permission[]): boolean {
  const ctx = usePermissionContext();
  return hasAny(ctx, permissions);
}

export function useRole(role: Role | Role[]): boolean {
  const ctx = usePermissionContext();
  return hasRole(ctx, role);
}

export function useIsAdmin(): boolean {
  return useRole('admin');
}
