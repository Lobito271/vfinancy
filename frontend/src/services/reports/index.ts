import { sleep } from '@/utils';

export type ReportType =
  | 'sales-by-period'
  | 'sales-by-customer'
  | 'sales-by-product'
  | 'purchases-by-supplier'
  | 'accounts-receivable'
  | 'accounts-payable'
  | 'inventory-valuation'
  | 'profitability';

export interface ReportRunInput {
  type: ReportType;
  from: string;
  to: string;
  format: 'pdf' | 'excel' | 'csv';
  filters?: Record<string, string | number | boolean>;
}

export interface ReportResult {
  id: string;
  type: ReportType;
  generatedAt: string;
  status: 'queued' | 'running' | 'done' | 'error';
  url?: string;
}

export const reportsService = {
  async list(): Promise<ReportType[]> {
    await sleep(100);
    return [
      'sales-by-period',
      'sales-by-customer',
      'sales-by-product',
      'purchases-by-supplier',
      'accounts-receivable',
      'accounts-payable',
      'inventory-valuation',
      'profitability',
    ];
  },
  async run(input: ReportRunInput): Promise<ReportResult> {
    await sleep(800);
    return {
      id: `r-${Date.now()}`,
      type: input.type,
      generatedAt: new Date().toISOString(),
      status: 'done',
      url: '#',
    };
  },
};
