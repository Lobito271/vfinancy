import type { ReactNode } from 'react';
import { type Permission } from '@/utils/permissions';
import type { Role } from '@/constants/permissions';

interface CanProps {
  permission?: Permission | Permission[];
  anyPermission?: Permission[];
  role?: Role | Role[];
  fallback?: ReactNode;
  children: ReactNode;
}

export function Can({ permission, anyPermission, role, fallback = null, children }: CanProps) {
  void permission;
  void anyPermission;
  void role;
  void fallback;
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
