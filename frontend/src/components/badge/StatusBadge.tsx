import { Badge } from './Badge';
import { t } from '@/locales';
import type { SaleStatus } from '@/types/domain';

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
