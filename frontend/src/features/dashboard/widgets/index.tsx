import { CircleDollarSign, TrendingUp } from 'lucide-react';
import { useQuery } from '@tanstack/react-query';
import { StatCard } from '@/components/card';
import { Badge } from '@/components/badge';
import { EmptyState } from '@/components/feedback';
import { BarChart } from '@/components/charts';
import { formatCurrency, formatNumber, daysBetween } from '@/utils/format';
import { wailsClient } from '@/services/bindings';
import { queryKeys } from '@/services/queryKeys';
import { useDashboardData } from '../hooks/useDashboard';
import { WidgetShell } from './WidgetShell';
import { ListRow } from '@/components/misc';

export function CollectedMonthWidget() {
  const { data } = useDashboardData();
  return (
    <StatCard
      label="Ventas cobradas del mes"
      value={formatCurrency(data?.monthCollected ?? 0)}
      icon={CircleDollarSign}
    />
  );
}

export function NetProfitWidget() {
  const { data } = useDashboardData();
  return (
    <StatCard
      label="Ganancia neta del mes"
      value={formatCurrency(data?.monthProfit ?? 0)}
      icon={TrendingUp}
    />
  );
}

export function MonthlyNetProfitChart() {
  const { data, isLoading, isError, error } = useDashboardData();
  const points = data?.profitSeries ?? [];
  return (
    <WidgetShell
      title="Ganancia neta mensual"
      description="Utilidad consolidada de los últimos 6 meses"
      loading={isLoading}
      error={isError ? (error as Error) : null}
      actions={<Badge variant="primary">Últimos 6 meses</Badge>}
    >
      {points.some((p) => p.value !== 0) ? (
        <BarChart data={points} formatY={(v) => formatCurrency(v)} />
      ) : (
        <EmptyState
          title="Sin ganancias registradas"
          description="Las ventas cobradas completarán el gráfico."
        />
      )}
    </WidgetShell>
  );
}

export function MonthStatusBadgesWidget() {
  const { data } = useDashboardData();
  return (
    <div className="stat-card">
      <p className="stat-card__label">Estado del mes</p>
      <div className="stat-card__badges">
        <Badge variant="success">Cobradas: {data?.monthPaidCount ?? 0}</Badge>
        <Badge variant="warning">Pendientes: {data?.monthPendingCount ?? 0}</Badge>
        <Badge variant="muted">Anuladas: {data?.monthCancelledCount ?? 0}</Badge>
      </div>
    </div>
  );
}

export function ClearanceWidget() {
  const { data, isLoading, isError, error } = useQuery({
    queryKey: queryKeys.inventory.clearance,
    queryFn: () => wailsClient.listClearanceProducts(),
  });

  const today = new Date();
  const items = (data ?? [])
    .filter((b) => b.quantity > 0)
    .map((b) => ({ ...b, daysLeft: b.maxSaleDate ? daysBetween(today, b.maxSaleDate) : 0 }))
    .sort((a, b) => a.daysLeft - b.daysLeft);

  return (
    <WidgetShell
      title="Productos en Remate"
      description="Lotes en liquidación o próximos a vencer"
      loading={isLoading}
      error={isError ? (error as Error) : null}
    >
      {items.length === 0 ? (
        <EmptyState
          title="Sin productos en remate"
          description="Los lotes que superen la fecha límite de venta aparecerán aquí."
        />
      ) : (
        <div className="stack stack--sm">
          {items.map((b) => (
            <ListRow
              key={b.id}
              title={b.productDescription}
              trailing={
                <>
                  <span className="tabular">{formatNumber(b.quantity)}</span>
                  <Badge variant={b.daysLeft <= 0 ? 'destructive' : 'warning'}>
                    {b.daysLeft <= 0 ? 'En remate' : `Vence en ${b.daysLeft} días`}
                  </Badge>
                </>
              }
            />
          ))}
        </div>
      )}
    </WidgetShell>
  );
}
