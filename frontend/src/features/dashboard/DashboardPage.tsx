import { DashboardGrid, type DashboardGridItem } from './DashboardGrid';
import {
  MonthSalesWidget,
  MonthPurchasesWidget,
  ProfitWidget,
  AccountsReceivableWidget,
  AccountsPayableWidget,
  ClearanceWidget,
  CustomersWithDebtWidget,
  LowStockWidget,
  SalesLast7DaysWidget,
  SalesByCategoryWidget,
  TopProductsWidget,
  RecentActivityWidget,
  OperationsByStatusWidget,
} from './widgets';
import { PageContainer, PageHeader } from '@/components/layout';
import { FileText } from 'lucide-react';

const defaultLayout: DashboardGridItem[] = [
  { id: 'monthSales', size: 'sm', content: <MonthSalesWidget /> },
  { id: 'monthPurchases', size: 'sm', content: <MonthPurchasesWidget /> },
  { id: 'profit', size: 'sm', content: <ProfitWidget /> },
  { id: 'receivable', size: 'sm', content: <AccountsReceivableWidget /> },
  { id: 'payable', size: 'sm', content: <AccountsPayableWidget /> },
  { id: 'clearance', size: 'sm', content: <ClearanceWidget /> },
  { id: 'debt', size: 'sm', content: <CustomersWithDebtWidget /> },
  { id: 'lowStock', size: 'sm', content: <LowStockWidget /> },
  { id: 'sales7', size: 'lg', content: <SalesLast7DaysWidget /> },
  { id: 'salesByCategory', size: 'md', content: <SalesByCategoryWidget /> },
  { id: 'topProducts', size: 'md', content: <TopProductsWidget /> },
  { id: 'recentActivity', size: 'lg', content: <RecentActivityWidget /> },
  { id: 'operationsByStatus', size: 'full', content: <OperationsByStatusWidget /> },
];

export function DashboardPage() {
  return (
    <PageContainer>
      <PageHeader title="Inicio" subtitle="Resumen general del negocio" />
      <DashboardGrid items={defaultLayout} />
      <div className="flex justify-center pt-2 text-xs text-muted-foreground">
        <span className="inline-flex items-center gap-1">
          <FileText className="h-3 w-3" />
          Datos de demostración — no provienen de la base de datos
        </span>
      </div>
    </PageContainer>
  );
}
