import { useQuery } from '@tanstack/react-query';
import { CircleDollarSign, TrendingUp } from 'lucide-react';
import { StatCard } from '@/components/card';
import { Badge } from '@/components/badge';
import { EmptyState } from '@/components/feedback';
import { formatCurrency, formatNumber, daysBetween } from '@/utils/format';
import { wailsClient } from '@/services/bindings';
import { queryKeys } from '@/services/queryKeys';
import type { MonthlyProfitDTO } from '@/services/wails-types';
import { useDashboardData } from '../hooks/useDashboard';
import { WidgetShell } from './WidgetShell';
import { ListRow } from '@/components/misc';

const PROFIT_MONTHS = 12;

function useMonthlyProfit() {
  return useQuery({
    queryKey: queryKeys.dashboard.monthlyProfit(PROFIT_MONTHS),
    queryFn: () => wailsClient.listMonthlyProfit(PROFIT_MONTHS),
  });
}

/** Last row of the series, which is the running month. */
function currentMonth(rows: MonthlyProfitDTO[]): MonthlyProfitDTO | undefined {
  return rows[rows.length - 1];
}

export function MonthlyProfitWidget() {
  const { data, isLoading, isError, error } = useMonthlyProfit();
  const rows = data ?? [];

  return (
    <WidgetShell
      title="Utilidad mensual"
      description="Precio de venta menos el costo total (base + extras), mes a mes"
      loading={isLoading}
      error={isError ? (error as Error) : null}
      actions={<Badge variant="primary">Últimos {PROFIT_MONTHS} meses</Badge>}
    >
      {rows.length === 0 ? (
        <EmptyState
          title="Sin datos de utilidad"
          description="Registra compras y ventas para completar el desglose."
        />
      ) : (
        <div className="stack stack--sm">
          {rows.map((r) => (
            <ListRow
              key={`${r.year}-${r.month}`}
              title={r.label}
              meta={
                <>
                  {`Base ${formatCurrency(r.baseCostUsd, 'USD')}`}
                  {` · Extras ${formatCurrency(r.extraCostUsd, 'USD')}`}
                  {` · Total ${formatCurrency(r.totalCostUsd, 'USD')}`}
                  {` · Ventas ${formatCurrency(r.salesPen)}`}
                </>
              }
              trailing={
                <>
                  <span className="list-row__meta">Utilidad</span>
                  <span
                    className={
                      r.profitPen < 0 ? 'tabular text-destructive' : 'tabular fw-medium'
                    }
                  >
                    {formatCurrency(r.profitPen)}
                  </span>
                </>
              }
            />
          ))}
        </div>
      )}
    </WidgetShell>
  );
}

export function NetProfitWidget() {
  const { data, isLoading } = useMonthlyProfit();
  const current = currentMonth(data ?? []);
  return (
    <StatCard
      label="Utilidad del mes"
      value={isLoading ? '—' : formatCurrency(current?.profitPen ?? 0)}
      icon={TrendingUp}
    />
  );
}

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