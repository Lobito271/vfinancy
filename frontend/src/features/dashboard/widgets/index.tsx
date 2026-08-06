import { ShoppingCart, TrendingUp, TrendingDown, Minus, type LucideIcon } from 'lucide-react';
import { WidgetShell } from './WidgetShell';
import { StatCard } from '@/components/card';
import { formatCurrency, formatNumber, formatPercent } from '@/utils/format';
import { cn } from '@/utils/cn';
import { dashboardKpis, salesLast7Days, topProducts, salesByCategory, activity } from '@/data/mock';
import { LineChart, BarChart, PieChart } from '@/components/charts';
import { DataTable, type Column } from '@/components/table';
import { formatDateTime, formatRelative } from '@/utils/format';
import type { ActivityItem, ChartPoint } from '@/data/mock';

export interface WidgetMeta<P = unknown> {
  id: string;
  component: React.ComponentType<P>;
  title: string;
  defaultSize?: 'sm' | 'md' | 'lg' | 'xl' | 'full';
  icon?: LucideIcon;
}

export const WIDGET_REGISTRY = {
  'kpi.monthSales': MonthSalesWidget,
  'kpi.monthPurchases': MonthPurchasesWidget,
  'kpi.profit': ProfitWidget,
  'kpi.accountsReceivable': AccountsReceivableWidget,
  'kpi.accountsPayable': AccountsPayableWidget,
  'kpi.clearance': ClearanceWidget,
  'kpi.customersWithDebt': CustomersWithDebtWidget,
  'kpi.lowStock': LowStockWidget,
  'chart.salesLast7Days': SalesLast7DaysWidget,
  'chart.salesByCategory': SalesByCategoryWidget,
  'list.topProducts': TopProductsWidget,
  'list.recentActivity': RecentActivityWidget,
  'chart.operationsByStatus': OperationsByStatusWidget,
} as const;

export type WidgetId = keyof typeof WIDGET_REGISTRY;

interface BaseKpiProps {
  label: string;
  value: number;
  format: 'currency' | 'number' | 'percent';
  change?: number;
  icon: LucideIcon;
  currency?: string;
}

export function KpiWidget({ label, value, format, change, icon, currency }: BaseKpiProps) {
  const formatted =
    format === 'currency' ? formatCurrency(value, currency) : format === 'percent' ? formatPercent(value) : formatNumber(value);
  return (
    <StatCard
      label={label}
      value={formatted}
      icon={icon}
      change={change}
      changeLabel="vs. mes anterior"
    />
  );
}

void ShoppingCart;
void TrendingUp;
void TrendingDown;
void Minus;
void cn;
void WIDGET_REGISTRY;

export function MonthSalesWidget() {
  return <KpiWidget label="Ventas del Mes" value={dashboardKpis.monthSales} format="currency" change={dashboardKpis.monthSalesChange} icon={ShoppingCart} />;
}

export function MonthPurchasesWidget() {
  return <KpiWidget label="Compras del Mes" value={dashboardKpis.monthPurchases} format="currency" change={dashboardKpis.monthPurchasesChange} icon={ShoppingCart} />;
}

export function ProfitWidget() {
  return <KpiWidget label="Utilidad" value={dashboardKpis.profit} format="currency" change={dashboardKpis.profitChange} icon={TrendingUp} />;
}

export function AccountsReceivableWidget() {
  return <KpiWidget label="Cuentas por Cobrar" value={dashboardKpis.accountsReceivable} format="currency" icon={ShoppingCart} />;
}

export function AccountsPayableWidget() {
  return <KpiWidget label="Cuentas por Pagar" value={dashboardKpis.accountsPayable} format="currency" icon={ShoppingCart} />;
}

export function ClearanceWidget() {
  return <KpiWidget label="Productos en Remate" value={dashboardKpis.clearanceProducts} format="number" icon={TrendingDown} />;
}

export function CustomersWithDebtWidget() {
  return <KpiWidget label="Clientes con Deuda" value={dashboardKpis.customersWithDebt} format="number" icon={Minus} />;
}

export function LowStockWidget() {
  return <KpiWidget label="Stock Bajo" value={dashboardKpis.lowStock} format="number" icon={TrendingDown} />;
}

export function SalesLast7DaysWidget() {
  return (
    <WidgetShell title="Ventas últimos 7 días" description="Comparación diaria del periodo">
      <LineChart data={salesLast7Days} />
    </WidgetShell>
  );
}

export function SalesByCategoryWidget() {
  return (
    <WidgetShell title="Ventas por categoría" description="Distribución porcentual">
      <PieChart data={salesByCategory} />
    </WidgetShell>
  );
}

export function TopProductsWidget() {
  const columns: Column<ChartPoint>[] = [
    { id: 'label', header: 'Producto', cell: (r) => r.label },
    { id: 'value', header: 'Ventas', align: 'right', cell: (r) => <span className="font-medium tabular-nums">{formatCurrency(r.value)}</span> },
  ];
  return (
    <WidgetShell title="Productos más vendidos" description="Top del mes">
      <DataTable columns={columns} data={topProducts} keyField="label" globalSearch={false} exportable={false} />
    </WidgetShell>
  );
}

export function RecentActivityWidget() {
  const columns: Column<ActivityItem>[] = [
    { id: 'description', header: 'Descripción', cell: (r) => r.description },
    { id: 'amount', header: 'Monto', align: 'right', cell: (r) => (r.amount != null ? formatCurrency(r.amount) : '—') },
    { id: 'date', header: 'Fecha', cell: (r) => <span className="text-muted-foreground">{formatRelative(r.date)}</span> },
  ];
  return (
    <WidgetShell title="Actividad reciente" description="Últimas operaciones del sistema">
      <DataTable columns={columns} data={activity} keyField="id" globalSearch={false} exportable={false} />
    </WidgetShell>
  );
}

export function OperationsByStatusWidget() {
  return (
    <WidgetShell title="Operaciones por estado" description="Distribución de ventas y compras">
      <BarChart
        data={[
          { label: 'Pagado', value: 24 },
          { label: 'Pendiente', value: 6 },
          { label: 'Parcial', value: 3 },
          { label: 'Cancelado', value: 2 },
        ]}
        colors={['hsl(var(--success))', 'hsl(var(--warning))', 'hsl(var(--info))', 'hsl(var(--destructive))']}
      />
    </WidgetShell>
  );
}

void formatDateTime;
