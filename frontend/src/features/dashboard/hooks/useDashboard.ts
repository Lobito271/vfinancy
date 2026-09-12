import { useQuery } from '@tanstack/react-query';
import { wailsClient } from '@/services/bindings';
import { fetchAllPages } from '@/services/paginate';
import { queryKeys } from '@/services/queryKeys';
import type { ChartPoint } from '@/types/domain';

export interface DashboardData {
  monthCollected: number;
  monthProfit: number;
  profitSeries: ChartPoint[];
  monthPaidCount: number;
  monthPendingCount: number;
  monthCancelledCount: number;
}

const monthLabels = new Intl.DateTimeFormat('es-PE', { month: 'short' });

const dayPattern = /^(\d{4})-(\d{2})-(\d{2})/;

function parseDay(value: string): { y: number; m: number } | null {
  const match = dayPattern.exec(value);
  if (!match) return null;
  return { y: Number(match[1]), m: Number(match[2]) - 1 };
}

function firstOfMonth(date: Date): string {
  return `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, '0')}-01`;
}

export function useDashboardData() {
  return useQuery({
    queryKey: queryKeys.dashboard.overview,
    queryFn: async (): Promise<DashboardData> => {
      const now = new Date();
      const rangeStart = new Date(now.getFullYear(), now.getMonth() - 5, 1);
      const rangeEnd = new Date(now.getFullYear(), now.getMonth() + 1, 1);

      const [collections, sales] = await Promise.all([
        wailsClient.listSaleCollections(firstOfMonth(rangeStart), firstOfMonth(rangeEnd)),
        fetchAllPages((page, pageSize) =>
          wailsClient.listSales({
            page,
            pageSize,
            search: '',
            status: '',
            saleType: '',
            customerId: '',
            from: '',
            to: '',
            onlyUnpaid: false,
          }),
        ),
      ]);

      const saleById = new Map(sales.map((s) => [s.id, s]));

      const series: ChartPoint[] = [];
      const index = new Map<number, number>();
      for (let i = 5; i >= 0; i -= 1) {
        const d = new Date(now.getFullYear(), now.getMonth() - i, 1);
        index.set(d.getFullYear() * 12 + d.getMonth(), series.length);
        series.push({ label: monthLabels.format(d), value: 0 });
      }

      let monthCollected = 0;
      let monthProfit = 0;
      for (const c of collections) {
        const sale = saleById.get(c.saleId);
        if (!sale || sale.status === 'cancelled') continue;
        const day = parseDay(c.paymentDate);
        if (!day) continue;
        const slot = index.get(day.y * 12 + day.m);
        if (slot === undefined) continue;
        const margin = sale.total > 0 ? sale.profit / sale.total : 0;
        const collectedProfit = c.amount * margin;
        series[slot].value += collectedProfit;
        if (slot === series.length - 1) {
          monthCollected += c.amount;
          monthProfit += collectedProfit;
        }
      }

      let monthPaidCount = 0;
      let monthPendingCount = 0;
      let monthCancelledCount = 0;
      for (const sale of sales) {
        const day = parseDay(sale.saleDate);
        if (!day || day.y !== now.getFullYear() || day.m !== now.getMonth()) continue;
        if (sale.status === 'paid') monthPaidCount += 1;
        else if (sale.status === 'pending' || sale.status === 'partial') monthPendingCount += 1;
        else if (sale.status === 'cancelled') monthCancelledCount += 1;
      }

      return {
        monthCollected,
        monthProfit,
        profitSeries: series,
        monthPaidCount,
        monthPendingCount,
        monthCancelledCount,
      };
    },
  });
}
