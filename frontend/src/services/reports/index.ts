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

const reportTypes: ReportType[] = [
  'sales-by-period',
  'sales-by-customer',
  'sales-by-product',
  'purchases-by-supplier',
  'accounts-receivable',
  'accounts-payable',
  'inventory-valuation',
  'profitability',
];

export const reportsService = {
  async list(): Promise<ReportType[]> {
    return [...reportTypes];
  },
  async run(_input: ReportRunInput): Promise<ReportResult> {
    throw new Error('Report generation is not yet available. Please try again later.');
  },
};
