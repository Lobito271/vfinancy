import {
  type Permission,
  type PermissionContext,
} from '@/utils/permissions';
import type { Role } from '@/constants/permissions';

export function usePermissionContext(): PermissionContext {
  return { user: null, roles: [], permissions: ['*.*'] };
}

export function usePermission(permission: Permission | Permission[]): boolean {
  void permission;
  return true;
}

export function useAnyPermission(permissions: Permission[]): boolean {
  void permissions;
  return true;
}

export function useRole(role: Role | Role[]): boolean {
  void role;
  return true;
}

export function useIsAdmin(): boolean {
  return useRole('admin');
}
