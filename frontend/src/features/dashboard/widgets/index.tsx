import { CircleDollarSign, TrendingUp } from 'lucide-react';
import { StatCard } from '@/components/card';
import { Badge } from '@/components/badge';
import { EmptyState } from '@/components/feedback';
import { BarChart } from '@/components/charts';
import { formatCurrency } from '@/utils/format';
import { useDashboardData } from '../hooks/useDashboard';
import { WidgetShell } from './WidgetShell';

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
  const { data, isLoading, isError, error } = useDashboardData();
  return (
    <WidgetShell
      title="Estado del mes"
      loading={isLoading}
      error={isError ? (error as Error) : null}
    >
      <div className="stack stack--sm">
        <div className="hstack hstack--sm">
          <Badge variant="success">Cobradas: {data?.monthPaidCount ?? 0}</Badge>
          <Badge variant="warning">Pendientes: {data?.monthPendingCount ?? 0}</Badge>
        </div>
        <div className="hstack hstack--sm">
          <Badge variant="muted">Anuladas: {data?.monthCancelledCount ?? 0}</Badge>
        </div>
      </div>
    </WidgetShell>
  );
}
