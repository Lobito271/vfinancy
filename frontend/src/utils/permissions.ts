import { getRolePermissions, type Role } from '@/constants/permissions';
import type { SessionUser } from '@/stores/session';

export type Permission = string;

export interface PermissionContext {
  user: SessionUser | null;
  roles: Role[];
  permissions: string[];
}

export function buildContext(user: SessionUser | null): PermissionContext {
  const roles = (user?.roles ?? []) as Role[];
  const permissions = Array.from(new Set(roles.flatMap((r) => getRolePermissions(r))));
  return { user, roles, permissions };
}

export function hasPermission(ctx: PermissionContext, permission: Permission | Permission[]): boolean {
  if (Array.isArray(permission)) return permission.every((p) => ctx.permissions.includes(p));
  return ctx.permissions.includes(permission);
}

export function hasAny(ctx: PermissionContext, permissions: Permission[]): boolean {
  return permissions.some((p) => ctx.permissions.includes(p));
}

export function hasRole(ctx: PermissionContext, role: Role | Role[]): boolean {
  const list = Array.isArray(role) ? role : [role];
  return list.some((r) => ctx.roles.includes(r));
}
