import { getRolePermissions, type Role } from '@/constants/permissions';
import type { SessionUser } from '@/stores/session';

export type Permission = string;

export const SUPERADMIN_PERMISSION = '*.*';

export interface PermissionContext {
  user: SessionUser | null;
  roles: Role[];
  permissions: string[];
}

export function buildContext(user: SessionUser | null): PermissionContext {
  const roles = (user?.roles ?? []) as Role[];
  const permissions = Array.from(new Set(roles.flatMap((r) => getRolePermissions(r))));
  if (roles.includes('admin')) {
    permissions.push(SUPERADMIN_PERMISSION);
  }
  return { user, roles, permissions };
}

function isSuperadmin(ctx: PermissionContext): boolean {
  return ctx.permissions.includes(SUPERADMIN_PERMISSION);
}

export function hasPermission(ctx: PermissionContext, permission: Permission | Permission[]): boolean {
  if (isSuperadmin(ctx)) return true;
  if (Array.isArray(permission)) return permission.every((p) => ctx.permissions.includes(p));
  return ctx.permissions.includes(permission);
}

export function hasAny(ctx: PermissionContext, permissions: Permission[]): boolean {
  if (isSuperadmin(ctx)) return true;
  return permissions.some((p) => ctx.permissions.includes(p));
}

export function hasRole(ctx: PermissionContext, role: Role | Role[]): boolean {
  const list = Array.isArray(role) ? role : [role];
  return list.some((r) => ctx.roles.includes(r));
}
