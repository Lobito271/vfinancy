import { useQuery } from '@tanstack/react-query';
import { salesService } from '@/services/sales';
import { purchasingService } from '@/services/purchasing';
import { customersService } from '@/services/customers';
import { suppliersService } from '@/services/suppliers';
import { inventoryService } from '@/services/inventory';
import { wailsClient } from '@/services/bindings';
import type { ActivityItem, ChartPoint } from '@/types/domain';

export interface DashboardKpis {
  monthSales: number;
  monthPurchases: number;
  profit: number;
  inventoryValue: number;
  accountsReceivable: number;
  accountsPayable: number;
  clearanceProducts: number;
  customersWithDebt: number;
  lowStock: number;
  activeProducts: number;
  activeCustomers: number;
}

interface DashboardData {
  kpis: DashboardKpis;
  salesLast7Days: ChartPoint[];
  salesByStatus: ChartPoint[];
  topProducts: ChartPoint[];
  activity: ActivityItem[];
}

function isCurrentMonth(iso: string): boolean {
  const date = new Date(iso);
  const now = new Date();
  return (
    date.getFullYear() === now.getFullYear() &&
    date.getMonth() === now.getMonth()
  );
}

function dayKey(iso: string): string {
  const date = new Date(iso);
  return `${date.getFullYear()}-${date.getMonth()}-${date.getDate()}`;
}

const dayFormatter = new Intl.DateTimeFormat('es-PE', { weekday: 'short' });

function last7Days(): Array<{ key: string; label: string }> {
  const out: Array<{ key: string; label: string }> = [];
  const now = new Date();
  for (let i = 6; i >= 0; i -= 1) {
    const d = new Date(now.getFullYear(), now.getMonth(), now.getDate() - i);
    out.push({ key: dayKey(d.toISOString()), label: dayFormatter.format(d) });
  }
  return out;
}

export function useDashboardData() {
  return useQuery({
    queryKey: ['dashboard', 'overview'],
    queryFn: async (): Promise<DashboardData> => {
      const [sales, purchases, customersRes, suppliersRes, clearance, lowStock, batches] = await Promise.all([
        salesService.list(),
        purchasingService.list(),
        customersService.list({ page: 1, pageSize: 500 }),
        suppliersService.list({ page: 1, pageSize: 500 }),
        inventoryService.getClearance(),
        inventoryService.getLowStock(),
        wailsClient.listInventoryBatches({ onlyClearance: false, page: 1, pageSize: 500 }),
      ]);

      const monthSales = sales.filter((s) => isCurrentMonth(s.date));
      const monthPurchases = purchases.filter((p) => isCurrentMonth(p.date));

      const inventoryValue = batches.items.reduce(
        (sum, b) => sum + Number(b.currentQuantity) * Number(b.unitCost),
        0,
      );

      const customers = customersRes.items;
      const suppliers = suppliersRes.items;

      const kpis: DashboardKpis = {
        monthSales: monthSales.reduce((s, x) => s + x.total, 0),
        monthPurchases: monthPurchases.reduce((s, x) => s + x.total, 0),
        profit: monthSales.reduce((s, x) => s + x.profit, 0),
        inventoryValue,
        accountsReceivable: customers.reduce((s, c) => s + c.currentDebt, 0),
        accountsPayable: suppliers.reduce((s, c) => s + c.currentDebt, 0),
        clearanceProducts: clearance.length,
        customersWithDebt: customers.filter((c) => c.currentDebt > 0).length,
        lowStock: lowStock.length,
        activeProducts: 0,
        activeCustomers: customers.length,
      };

      const days = last7Days();
      const byDay = new Map(days.map((d) => [d.key, 0]));
      for (const sale of sales) {
        const key = dayKey(sale.date);
        if (byDay.has(key)) byDay.set(key, (byDay.get(key) ?? 0) + sale.total);
      }
      const salesLast7Days: ChartPoint[] = days.map((d) => ({ label: d.label, value: byDay.get(d.key) ?? 0 }));

      const statusLabels: Record<string, string> = {
        paid: 'Cobrada',
        pending: 'Pendiente',
        partial: 'Parcial',
        cancelled: 'Anulada',
      };
      const statusCounts = new Map<string, number>();
      for (const sale of sales) {
        statusCounts.set(sale.status, (statusCounts.get(sale.status) ?? 0) + 1);
      }
      const salesByStatus: ChartPoint[] = [];
      for (const [status, count] of statusCounts.entries()) {
        if (count > 0) salesByStatus.push({ label: statusLabels[status] ?? status, value: count });
      }

      const activity: ActivityItem[] = [
        ...sales.map<ActivityItem>((s) => ({
          id: `sale-${s.id}`,
          type: 'sale',
          description: `Venta ${s.number} — ${s.customerName}`,
          amount: s.total,
          date: s.date,
          user: '',
        })),
        ...purchases.map<ActivityItem>((p) => ({
          id: `purchase-${p.id}`,
          type: 'purchase',
          description: `Compra ${p.number} — ${p.supplierName}`,
          amount: p.total,
          date: p.date,
          user: '',
        })),
      ]
        .sort((a, b) => new Date(b.date).getTime() - new Date(a.date).getTime())
        .slice(0, 8);

      return {
        kpis,
        salesLast7Days,
        salesByStatus,
        topProducts: [],
        activity,
      };
    },
  });
}
