import { useQuery } from '@tanstack/react-query';
import { wailsClient } from '@/services/bindings';
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

function aggregate(sales: Awaited<ReturnType<typeof wailsClient.listSales>>['items'], now: Date) {
  const series: ChartPoint[] = [];
  const index = new Map<number, number>();
  for (let i = 5; i >= 0; i -= 1) {
    const d = new Date(now.getFullYear(), now.getMonth() - i, 1);
    index.set(d.getFullYear() * 12 + d.getMonth(), series.length);
    series.push({ label: monthLabels.format(d), value: 0 });
  }
  let monthProfit = 0;
  let monthPaidCount = 0;
  let monthPendingCount = 0;
  let monthCancelledCount = 0;
  for (const sale of sales) {
    const day = parseDay(sale.saleDate);
    if (!day) continue;
    const slot = index.get(day.y * 12 + day.m);
    if (slot === undefined) continue;
    if (sale.status === 'paid') {
      series[slot].value += sale.profit;
      if (slot === series.length - 1) monthProfit += sale.profit;
      monthPaidCount += 1;
    } else if (sale.status === 'pending' || sale.status === 'partial') {
      monthPendingCount += 1;
    } else if (sale.status === 'cancelled') {
      monthCancelledCount += 1;
    }
  }
  return { series, monthProfit, monthPaidCount, monthPendingCount, monthCancelledCount } as const;
}

export function useDashboardData() {
  return useQuery({
    queryKey: queryKeys.dashboard.overview,
    queryFn: async (): Promise<DashboardData> => {
      const [payments, sales] = await Promise.all([
        wailsClient.listSalePayments({ page: 1, pageSize: 1000 }, '', ''),
        wailsClient.listSales({
          page: 1,
          pageSize: 1000,
          search: '',
          status: '',
          saleType: '',
          customerId: '',
          from: '',
          to: '',
          onlyUnpaid: false,
        }),
      ]);
      const now = new Date();
      const agg = aggregate(sales.items, now);
      const currentY = now.getFullYear();
      const currentM = now.getMonth();
      const monthCollected = payments.items.reduce((sum, p) => {
        const day = parseDay(p.paymentDate);
        return day && day.y === currentY && day.m === currentM ? sum + p.amount : sum;
      }, 0);
      return {
        monthCollected,
        profitSeries: agg.series,
        monthProfit: agg.monthProfit,
        monthPaidCount: agg.monthPaidCount,
        monthPendingCount: agg.monthPendingCount,
        monthCancelledCount: agg.monthCancelledCount,
      };
    },
  });
}
