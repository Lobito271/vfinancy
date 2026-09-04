import { ShoppingCart, TrendingUp, Wallet, Users, AlertTriangle, type LucideIcon } from 'lucide-react';
import { WidgetShell } from './WidgetShell';
import { StatCard } from '@/components/card';
import { EmptyState } from '@/components/feedback';
import { formatCurrency, formatNumber, formatPercent } from '@/utils/format';
import { useDashboardData, type DashboardKpis } from '../hooks/useDashboard';
import { LineChart, BarChart } from '@/components/charts';
import { DataTable, type Column } from '@/components/table';
import { formatRelative } from '@/utils/format';
import type { ActivityItem, ChartPoint } from '@/types/domain';

interface KpiWidgetProps {
  label: string;
  value: number;
  format: 'currency' | 'number' | 'percent';
  icon: LucideIcon;
  currency?: string;
}

function KpiWidget({ label, value, format, icon, currency }: KpiWidgetProps) {
  const formatted =
    format === 'currency' ? formatCurrency(value, currency) : format === 'percent' ? formatPercent(value) : formatNumber(value);
  return <StatCard label={label} value={formatted} icon={icon} />;
}

function fromKpis(kpis: DashboardKpis | undefined) {
  return {
    monthSales: kpis?.monthSales ?? 0,
    monthPurchases: kpis?.monthPurchases ?? 0,
    profit: kpis?.profit ?? 0,
    inventoryValue: kpis?.inventoryValue ?? 0,
    accountsReceivable: kpis?.accountsReceivable ?? 0,
    accountsPayable: kpis?.accountsPayable ?? 0,
    clearanceProducts: kpis?.clearanceProducts ?? 0,
    customersWithDebt: kpis?.customersWithDebt ?? 0,
    lowStock: kpis?.lowStock ?? 0,
  };
}

export function MonthSalesWidget() {
  const { data } = useDashboardData();
  const k = fromKpis(data?.kpis);
  return <KpiWidget label="Ventas del Mes" value={k.monthSales} format="currency" icon={ShoppingCart} />;
}

export function MonthPurchasesWidget() {
  const { data } = useDashboardData();
  const k = fromKpis(data?.kpis);
  return <KpiWidget label="Compras del Mes" value={k.monthPurchases} format="currency" icon={Wallet} />;
}

export function ProfitWidget() {
  const { data } = useDashboardData();
  const k = fromKpis(data?.kpis);
  return <KpiWidget label="Utilidad" value={k.profit} format="currency" icon={TrendingUp} />;
}

export function AccountsReceivableWidget() {
  const { data } = useDashboardData();
  const k = fromKpis(data?.kpis);
  return <KpiWidget label="Cuentas por Cobrar" value={k.accountsReceivable} format="currency" icon={Wallet} />;
}

export function AccountsPayableWidget() {
  const { data } = useDashboardData();
  const k = fromKpis(data?.kpis);
  return <KpiWidget label="Cuentas por Pagar" value={k.accountsPayable} format="currency" icon={Wallet} />;
}

export function ClearanceWidget() {
  const { data } = useDashboardData();
  const k = fromKpis(data?.kpis);
  return <KpiWidget label="Productos en Remate" value={k.clearanceProducts} format="number" icon={AlertTriangle} />;
}

export function CustomersWithDebtWidget() {
  const { data } = useDashboardData();
  const k = fromKpis(data?.kpis);
  return <KpiWidget label="Clientes con Deuda" value={k.customersWithDebt} format="number" icon={Users} />;
}

export function LowStockWidget() {
  const { data } = useDashboardData();
  const k = fromKpis(data?.kpis);
  return <KpiWidget label="Stock Bajo" value={k.lowStock} format="number" icon={AlertTriangle} />;
}

export function SalesLast7DaysWidget() {
  const { data, isLoading, isError, error } = useDashboardData();
  const points: ChartPoint[] = data?.salesLast7Days ?? [];
  return (
    <WidgetShell title="Ventas últimos 7 días" description="Comparación diaria del periodo" loading={isLoading} error={isError ? (error as Error) : null}>
      {points.length ? (
        <LineChart data={points} />
      ) : (
        <EmptyState title="Sin ventas" description="El gráfico se completará con las ventas del periodo." />
      )}
    </WidgetShell>
  );
}

export function SalesByStatusWidget() {
  const { data, isLoading, isError, error } = useDashboardData();
  const points: ChartPoint[] = data?.salesByStatus ?? [];
  return (
    <WidgetShell title="Ventas por estado" description="Distribución de documentos según su estado" loading={isLoading} error={isError ? (error as Error) : null}>
      {points.length ? (
        <BarChart data={points} />
      ) : (
        <EmptyState title="Sin ventas" description="La distribución se completará con los documentos de venta." />
      )}
    </WidgetShell>
  );
}

export function TopProductsWidget() {
  const { data, isLoading, isError, error } = useDashboardData();
  const rows: ChartPoint[] = data?.topProducts ?? [];
  const columns: Column<ChartPoint>[] = [
    { id: 'label', header: 'Producto', cell: (r) => r.label },
    {
      id: 'value',
      header: 'Ventas',
      align: 'numeric',
      cell: (r) => <span className="fw-medium tabular">{formatCurrency(r.value)}</span>,
    },
  ];
  return (
    <WidgetShell title="Productos más vendidos" description="Top del mes" loading={isLoading} error={isError ? (error as Error) : null}>
      {rows.length ? (
        <DataTable columns={columns} data={rows} keyField="label" />
      ) : (
        <EmptyState title="Sin datos" description="El ranking se completará con las ventas del mes." />
      )}
    </WidgetShell>
  );
}

export function RecentActivityWidget() {
  const { data, isLoading, isError, error } = useDashboardData();
  const rows: ActivityItem[] = data?.activity ?? [];
  const columns: Column<ActivityItem>[] = [
    { id: 'description', header: 'Descripción', cell: (r) => r.description },
    {
      id: 'amount',
      header: 'Monto',
      align: 'numeric',
      cell: (r) => (r.amount != null ? formatCurrency(r.amount) : '—'),
    },
    {
      id: 'date',
      header: 'Fecha',
      cell: (r) => <span className="muted">{formatRelative(r.date)}</span>,
    },
  ];
  return (
    <WidgetShell title="Actividad reciente" description="Últimas operaciones del sistema" loading={isLoading} error={isError ? (error as Error) : null}>
      {rows.length ? (
        <DataTable columns={columns} data={rows} keyField="id" />
      ) : (
        <EmptyState title="Sin actividad" description="Las operaciones registradas aparecerán aquí." />
      )}
    </WidgetShell>
  );
}

