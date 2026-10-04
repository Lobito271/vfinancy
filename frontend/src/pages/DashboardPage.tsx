import { DashboardGrid, type DashboardGridItem } from '@/features/dashboard/DashboardGrid';
import {
  CollectedMonthWidget,
  NetProfitWidget,
  ProfitBreakdownWidget,
  MonthStatusBadgesWidget,
  ClearanceWidget,
} from '@/features/dashboard/widgets';
import { PageContainer, PageHeader } from '@/components/layout';

const defaultLayout: DashboardGridItem[] = [
  { id: 'monthCollected', size: 'sm', content: <CollectedMonthWidget /> },
  { id: 'monthProfit', size: 'sm', content: <NetProfitWidget /> },
  { id: 'monthStatus', size: 'sm', content: <MonthStatusBadgesWidget /> },
  { id: 'profitBreakdown', size: 'full', content: <ProfitBreakdownWidget /> },
  { id: 'clearance', size: 'full', content: <ClearanceWidget /> },
];

export function DashboardPage() {
  return (
    <PageContainer>
      <PageHeader title="Inicio" />
      <DashboardGrid items={defaultLayout} />
    </PageContainer>
  );
}
