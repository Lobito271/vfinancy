import { useMemo, useState } from 'react';
import { useQuery } from '@tanstack/react-query';
import { CircleDollarSign, TrendingUp } from 'lucide-react';
import { StatCard } from '@/components/card';
import { Badge } from '@/components/badge';
import { Button } from '@/components/button';
import { ProfitChart, type ProfitChartPoint } from '@/components/charts';
import { EmptyState } from '@/components/feedback';
import { ListRow } from '@/components/misc';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/select';
import { formatCurrency, formatNumber, daysBetween } from '@/utils/format';
import { wailsClient } from '@/services/bindings';
import { queryKeys } from '@/services/queryKeys';
import type { ProfitBreakdownDTO } from '@/services/wails-types';
import { useDashboardData } from '../hooks/useDashboard';
import { WidgetShell } from './WidgetShell';

const YEAR_OPTIONS = 5;
const MONTH_NAMES = [
  'Enero', 'Febrero', 'Marzo', 'Abril', 'Mayo', 'Junio',
  'Julio', 'Agosto', 'Septiembre', 'Octubre', 'Noviembre', 'Diciembre',
];

type PeriodMode = 'month' | 'year';

function useProfitBreakdown(year: number, month: number) {
  return useQuery({
    queryKey: queryKeys.dashboard.profitBreakdown(year, month),
    queryFn: () => wailsClient.listProfitBreakdown(year, month),
  });
}

interface PeriodTotals {
  sales: number;
  base: number;
  extra: number;
  profit: number;
}

// sumPeriod keeps the identity Utilidad = Venta - Base - Extras on the
// totals, so the widget can never show figures that disagree with the
// per-month rows it charts.
function sumPeriod(rows: ProfitBreakdownDTO[]): PeriodTotals {
  const totals = rows.reduce(
    (acc, r) => ({
      sales: acc.sales + r.salesPen,
      base: acc.base + r.baseCostPen,
      extra: acc.extra + r.extraCostPen,
      profit: acc.profit + r.profitPen,
    }),
    { sales: 0, base: 0, extra: 0, profit: 0 },
  );
  return { ...totals, profit: totals.sales - totals.base - totals.extra };
}

function PeriodControls({ mode, year, month, onMode, onYear, onMonth }: {
  mode: PeriodMode;
  year: number;
  month: number;
  onMode: (mode: PeriodMode) => void;
  onYear: (year: number) => void;
  onMonth: (month: number) => void;
}) {
  const currentYear = new Date().getFullYear();
  const years = Array.from({ length: YEAR_OPTIONS }, (_, i) => currentYear - i);
  const yearItems = years.map((y) => ({ value: String(y), label: String(y) }));
  const monthItems = MONTH_NAMES.map((name, index) => ({ value: String(index + 1), label: name }));
  return (
    <div className="hstack hstack--sm widget-controls">
      <div className="hstack hstack--sm">
        {(['month', 'year'] as PeriodMode[]).map((value) => (
          <Button
            key={value}
            variant={mode === value ? 'primary' : 'outline'}
            size="sm"
            aria-pressed={mode === value}
            onClick={() => onMode(value)}
          >
            {value === 'month' ? 'Mensual' : 'Anual'}
          </Button>
        ))}
      </div>
      <Select items={yearItems} value={String(year)} onValueChange={(v) => onYear(Number(v))}>
        <SelectTrigger aria-label="Año" className="select-trigger--compact">
          <SelectValue />
        </SelectTrigger>
        <SelectContent>
          {years.map((y) => (
            <SelectItem key={y} value={String(y)}>
              {y}
            </SelectItem>
          ))}
        </SelectContent>
      </Select>
      {mode === 'month' && (
        <Select items={monthItems} value={String(month)} onValueChange={(v) => onMonth(Number(v))}>
          <SelectTrigger aria-label="Mes" className="select-trigger--compact">
            <SelectValue />
          </SelectTrigger>
          <SelectContent>
            {MONTH_NAMES.map((name, index) => (
              <SelectItem key={name} value={String(index + 1)}>
                {name}
              </SelectItem>
            ))}
          </SelectContent>
        </Select>
      )}
    </div>
  );
}

export function ProfitBreakdownWidget() {
  const today = new Date();
  const [mode, setMode] = useState<PeriodMode>('month');
  const [year, setYear] = useState(today.getFullYear());
  const [month, setMonth] = useState(today.getMonth() + 1);

  const { data, isLoading, isError, error } = useProfitBreakdown(year, mode === 'month' ? month : 0);
  const rows = data ?? [];

  const totals = useMemo(() => sumPeriod(rows), [rows]);
  const points = useMemo<ProfitChartPoint[]>(
    () => rows.map((r) => ({ label: r.label, sales: r.salesPen, base: r.baseCostPen, extra: r.extraCostPen, profit: r.profitPen })),
    [rows],
  );
  const concepts = useMemo(() => {
    const byConcept = new Map<string, { usd: number; pen: number }>();
    for (const r of rows) {
      for (const c of r.extraCosts) {
        const acc = byConcept.get(c.concept) ?? { usd: 0, pen: 0 };
        byConcept.set(c.concept, { usd: acc.usd + c.amountUsd, pen: acc.pen + c.amountPen });
      }
    }
    return [...byConcept.entries()]
      .map(([concept, acc]) => ({ concept, ...acc }))
      .sort((a, b) => b.pen - a.pen);
  }, [rows]);

  const periodLabel = mode === 'month' ? `${MONTH_NAMES[month - 1]} ${year}` : `Año ${year}`;
  const hasActivity = totals.sales > 0 || totals.base > 0 || totals.extra > 0;

  return (
    <WidgetShell
      title="Desglose de utilidad"
      description="Utilidad = precio de venta − costo base − costos extra"
      loading={isLoading}
      error={isError ? (error as Error) : null}
      actions={
        <PeriodControls
          mode={mode}
          year={year}
          month={month}
          onMode={setMode}
          onYear={setYear}
          onMonth={setMonth}
        />
      }
    >
      {!hasActivity ? (
        <EmptyState
          title="Sin datos de utilidad"
          description="Registra compras y ventas para completar el desglose."
        />
      ) : (
        <div className="stack">
          <ProfitChart data={points} single={mode === 'month'} />
          <div className="doc-summary">
            <div className="doc-summary__row">
              <div className="doc-summary__meta">Precio de venta · {periodLabel}</div>
              <div className="doc-summary__amount">{formatCurrency(totals.sales)}</div>
            </div>
            <div className="doc-summary__row">
              <div className="doc-summary__meta">Costo base</div>
              <div className="doc-summary__amount">{formatCurrency(totals.base)}</div>
            </div>
            <div className="doc-summary__row">
              <div className="doc-summary__meta">Costos extra</div>
              <div className="doc-summary__amount">{formatCurrency(totals.extra)}</div>
            </div>
            <div className="doc-summary__row doc-summary__row--total">
              <div className="doc-summary__meta">Utilidad total</div>
              <div className={totals.profit < 0 ? 'doc-summary__amount text-destructive' : 'doc-summary__amount'}>
                {formatCurrency(totals.profit)}
              </div>
            </div>
          </div>
          {concepts.length > 0 && (
            <div className="stack stack--sm">
              <h3 className="section-title">Costos extra por concepto</h3>
              {concepts.map((c) => (
                <ListRow
                  key={c.concept}
                  title={c.concept}
                  meta={formatCurrency(c.usd, 'USD')}
                  trailing={
                    <>
                      <span className="tabular fw-medium">{formatCurrency(c.pen)}</span>
                      <Badge variant="muted">
                        {totals.extra > 0 ? `${Math.round((c.pen / totals.extra) * 100)}%` : '—'}
                      </Badge>
                    </>
                  }
                />
              ))}
            </div>
          )}
        </div>
      )}
    </WidgetShell>
  );
}

export function NetProfitWidget() {
  const today = new Date();
  const { data, isLoading } = useProfitBreakdown(today.getFullYear(), today.getMonth() + 1);
  const current = data?.[0];
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
    .sort((a, b) => b.daysLeft - a.daysLeft);

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
