import { useQuery } from '@tanstack/react-query';
import { wailsClient } from '@/services/bindings';
import { fetchAllPages } from '@/services/paginate';
import { queryKeys } from '@/services/queryKeys';

interface DashboardData {
  monthCollected: number;
  monthPaidCount: number;
  monthPendingCount: number;
  monthCancelledCount: number;
}

const dayPattern = /^(\d{4})-(\d{2})-(\d{2})/;

function parseDay(value: string): { y: number; m: number } | null {
  const match = dayPattern.exec(value);
  if (!match) return null;
  return { y: Number(match[1]), m: Number(match[2]) - 1 };
}

function firstOfMonth(date: Date): string {
  return `${date.getFullYear()}-${String(date.getMonth() + 1).padStart(2, '0')}-01`;
}

// useDashboardData feeds the cash-collection and status widgets. The
// profitability figures live in ListMonthlyProfit, which aggregates
// purchases and sales on the server instead of guessing a margin here.
export function useDashboardData() {
  return useQuery({
    queryKey: queryKeys.dashboard.overview,
    queryFn: async (): Promise<DashboardData> => {
      const now = new Date();
      const rangeStart = new Date(now.getFullYear(), now.getMonth() - 1, 1);
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

      let monthCollected = 0;
      for (const c of collections) {
        const day = parseDay(c.paymentDate);
        if (!day || day.y !== now.getFullYear() || day.m !== now.getMonth()) continue;
        monthCollected += c.amount;
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
        monthPaidCount,
        monthPendingCount,
        monthCancelledCount,
      };
    },
  });
}