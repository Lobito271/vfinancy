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
  SalesByStatusWidget,
  TopProductsWidget,
  RecentActivityWidget,
} from './widgets';
import { PageContainer, PageHeader } from '@/components/layout';

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
  { id: 'salesByStatus', size: 'md', content: <SalesByStatusWidget /> },
  { id: 'topProducts', size: 'md', content: <TopProductsWidget /> },
  { id: 'recentActivity', size: 'lg', content: <RecentActivityWidget /> },
];

export function DashboardPage() {
  return (
    <PageContainer>
      <PageHeader title="Inicio" subtitle="Resumen general del negocio" />
      <DashboardGrid items={defaultLayout} />
    </PageContainer>
  );
}
