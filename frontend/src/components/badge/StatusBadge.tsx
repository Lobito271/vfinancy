import { Badge } from './Badge';
import { t } from '@/locales';
import type { SaleStatus } from '@/data/mock';

const statusMap: Record<SaleStatus, { variant: 'success' | 'warning' | 'info' | 'destructive' | 'muted'; label: string }> = {
  paid: { variant: 'success', label: t('status.paid') },
  pending: { variant: 'warning', label: t('status.pending') },
  partial: { variant: 'info', label: t('status.partial') },
  cancelled: { variant: 'destructive', label: t('status.cancelled') },
};

export function SaleStatusBadge({ status }: { status: SaleStatus }) {
  const cfg = statusMap[status];
  return <Badge variant={cfg.variant}>{cfg.label}</Badge>;
}

const customerStatusMap = {
  active: { variant: 'success' as const, label: t('status.active') },
  inactive: { variant: 'muted' as const, label: t('status.inactive') },
  blocked: { variant: 'destructive' as const, label: t('status.blocked') },
};

export function CustomerStatusBadge({ status }: { status: 'active' | 'inactive' | 'blocked' }) {
  const cfg = customerStatusMap[status];
  return <Badge variant={cfg.variant}>{cfg.label}</Badge>;
}

const stockMap = {
  inStock: { variant: 'success' as const, label: 'En Stock' },
  lowStock: { variant: 'warning' as const, label: t('status.lowStock') },
  outOfStock: { variant: 'destructive' as const, label: t('status.outOfStock') },
  clearance: { variant: 'destructive' as const, label: t('status.clearance') },
};

export function StockBadge({ stock, minStock, isClearance }: { stock: number; minStock: number; isClearance?: boolean }) {
  if (isClearance) return <Badge variant={stockMap.clearance.variant}>{stockMap.clearance.label}</Badge>;
  if (stock === 0) return <Badge variant={stockMap.outOfStock.variant}>{stockMap.outOfStock.label}</Badge>;
  if (stock < minStock) return <Badge variant={stockMap.lowStock.variant}>{stockMap.lowStock.label}</Badge>;
  return <Badge variant={stockMap.inStock.variant}>{stockMap.inStock.label}</Badge>;
}
