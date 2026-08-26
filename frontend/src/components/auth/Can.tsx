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


