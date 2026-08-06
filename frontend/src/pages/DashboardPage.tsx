import {
  ShoppingCart,
  Package,
  Wallet,
  AlertTriangle,
  TrendingUp,
  Users,
  Activity,
  Trophy,
  FileText,
} from 'lucide-react';
import { dashboardKpis, salesLast7Days, topProducts, salesByCategory, activity } from '@/data/mock';
import { formatCurrency, formatDateTime, formatPercent } from '@/lib/utils';
import { PageContainer, PageHeader, Section, Grid } from '@/components/layout';
import { Card, CardContent, CardHeader, CardTitle, CardDescription } from '@/components/card';
import { StatCard } from '@/components/card/StatCard';
import { DataTable, type Column } from '@/components/table';
import { LineChart, BarChart, PieChart } from '@/components/charts';
import { Button } from '@/components/button';
import type { ActivityItem, ChartPoint } from '@/data/mock';

const activityTypeMap: Record<ActivityItem['type'], { label: string; icon: typeof Activity }> = {
  sale: { label: 'Venta', icon: ShoppingCart },
  purchase: { label: 'Compra', icon: Package },
  payment: { label: 'Pago', icon: Wallet },
  customer: { label: 'Cliente', icon: Users },
  product: { label: 'Producto', icon: Package },
};

const activityColumns: Column<ActivityItem>[] = [
  {
    id: 'type',
    header: 'Tipo',
    cell: (row) => {
      const cfg = activityTypeMap[row.type];
      const Icon = cfg.icon;
      return (
        <div className="flex items-center gap-2">
          <Icon className="h-4 w-4 text-muted-foreground" aria-hidden="true" />
          <span>{cfg.label}</span>
        </div>
      );
    },
  },
  { id: 'description', header: 'Descripción', cell: (row) => row.description },
  {
    id: 'amount',
    header: 'Monto',
    align: 'right',
    cell: (row) => (row.amount !== undefined ? <span className="tabular-nums">{formatCurrency(row.amount)}</span> : '—'),
  },
  {
    id: 'date',
    header: 'Fecha',
    cell: (row) => <span className="text-muted-foreground">{formatDateTime(row.date)}</span>,
  },
];

const topProductColumns: Column<ChartPoint>[] = [
  { id: 'label', header: 'Producto', cell: (row) => row.label },
  {
    id: 'value',
    header: 'Ventas',
    align: 'right',
    cell: (row) => <span className="font-medium tabular-nums">{formatCurrency(row.value)}</span>,
  },
  {
    id: 'share',
    header: 'Participación',
    align: 'right',
    cell: (row) => {
      const total = topProducts.reduce((s, p) => s + p.value, 0);
      return <span className="text-muted-foreground tabular-nums">{formatPercent(row.value / total)}</span>;
    },
  },
];

export function DashboardPage() {
  return (
    <PageContainer>
      <PageHeader
        title="Inicio"
        subtitle="Resumen general del negocio"
        actions={
          <>
            <Button variant="outline">Exportar</Button>
            <Button>Nuevo</Button>
          </>
        }
      />

      <Grid cols={4} className="lg:grid-cols-4">
        <StatCard
          label="Ventas del Mes"
          value={formatCurrency(dashboardKpis.monthSales)}
          icon={ShoppingCart}
          change={dashboardKpis.monthSalesChange}
          changeLabel="vs. mes anterior"
        />
        <StatCard
          label="Compras del Mes"
          value={formatCurrency(dashboardKpis.monthPurchases)}
          icon={Package}
          change={dashboardKpis.monthPurchasesChange}
          changeLabel="vs. mes anterior"
        />
        <StatCard
          label="Utilidad"
          value={formatCurrency(dashboardKpis.profit)}
          icon={TrendingUp}
          change={dashboardKpis.profitChange}
          changeLabel="vs. mes anterior"
        />
        <StatCard
          label="Valor de Inventario"
          value={formatCurrency(dashboardKpis.inventoryValue)}
          icon={Wallet}
        />
      </Grid>

      <Grid cols={3}>
        <StatCard
          label="Cuentas por Cobrar"
          value={formatCurrency(dashboardKpis.accountsReceivable)}
          icon={Wallet}
        />
        <StatCard
          label="Cuentas por Pagar"
          value={formatCurrency(dashboardKpis.accountsPayable)}
          icon={Wallet}
        />
        <StatCard
          label="Productos en Remate"
          value={String(dashboardKpis.clearanceProducts)}
          icon={AlertTriangle}
        />
        <StatCard
          label="Clientes con Deuda"
          value={String(dashboardKpis.customersWithDebt)}
          icon={Users}
        />
        <StatCard
          label="Stock Bajo"
          value={String(dashboardKpis.lowStock)}
          icon={AlertTriangle}
        />
      </Grid>

      <Grid cols={3} className="lg:grid-cols-3">
        <Card className="lg:col-span-2">
          <CardHeader>
            <CardTitle>Ventas últimos 7 días</CardTitle>
            <CardDescription>Comparación diaria del periodo</CardDescription>
          </CardHeader>
          <CardContent>
            <LineChart data={salesLast7Days} />
          </CardContent>
        </Card>
        <Card>
          <CardHeader>
            <CardTitle>Ventas por categoría</CardTitle>
            <CardDescription>Distribución porcentual</CardDescription>
          </CardHeader>
          <CardContent>
            <PieChart data={salesByCategory} />
          </CardContent>
        </Card>
      </Grid>

      <Grid cols={2} className="lg:grid-cols-3">
        <Card className="lg:col-span-1">
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Trophy className="h-4 w-4" />
              Top productos
            </CardTitle>
            <CardDescription>Más vendidos del mes</CardDescription>
          </CardHeader>
          <CardContent>
            <DataTable
              columns={topProductColumns}
              data={topProducts}
              keyField="label"
            />
          </CardContent>
        </Card>
        <Card className="lg:col-span-2">
          <CardHeader>
            <CardTitle className="flex items-center gap-2">
              <Activity className="h-4 w-4" />
              Actividad reciente
            </CardTitle>
            <CardDescription>Últimas operaciones del sistema</CardDescription>
          </CardHeader>
          <CardContent className="p-0">
            <DataTable
              columns={activityColumns}
              data={activity}
              keyField="id"
            />
          </CardContent>
        </Card>
      </Grid>

      <Section
        title="Operaciones por estado"
        description="Distribución de ventas y compras según su estado actual"
      >
        <Card>
          <CardContent className="pt-6">
            <BarChart
              data={[
                { label: 'Pagado', value: 24 },
                { label: 'Pendiente', value: 6 },
                { label: 'Parcial', value: 3 },
                { label: 'Cancelado', value: 2 },
              ]}
              colors={[
                'hsl(var(--success))',
                'hsl(var(--warning))',
                'hsl(var(--info))',
                'hsl(var(--destructive))',
              ]}
            />
          </CardContent>
        </Card>
      </Section>

      <div className="flex justify-center pt-2 text-xs text-muted-foreground">
        <span className="inline-flex items-center gap-1">
          <FileText className="h-3 w-3" />
          Datos de demostración — no provienen de la base de datos
        </span>
      </div>
    </PageContainer>
  );
}
