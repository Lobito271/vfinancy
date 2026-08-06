import { sleep } from '@/utils';

export interface Account {
  code: string;
  name: string;
  type: 'asset' | 'liability' | 'equity' | 'income' | 'expense';
  parentCode?: string;
}

export interface JournalEntry {
  id: string;
  date: string;
  description: string;
  reference?: string;
  lines: JournalEntryLine[];
  status: 'draft' | 'posted';
}

export interface JournalEntryLine {
  accountCode: string;
  debit: number;
  credit: number;
}

const MOCK_ACCOUNTS: Account[] = [
  { code: '1', name: 'ACTIVO', type: 'asset' },
  { code: '1.1', name: 'Activo Corriente', type: 'asset', parentCode: '1' },
  { code: '1.1.01', name: 'Caja y Bancos', type: 'asset', parentCode: '1.1' },
  { code: '1.1.02', name: 'Cuentas por Cobrar', type: 'asset', parentCode: '1.1' },
  { code: '1.1.03', name: 'Inventario', type: 'asset', parentCode: '1.1' },
  { code: '2', name: 'PASIVO', type: 'liability' },
  { code: '2.1', name: 'Pasivo Corriente', type: 'liability', parentCode: '2' },
  { code: '2.1.01', name: 'Cuentas por Pagar', type: 'liability', parentCode: '2.1' },
  { code: '4', name: 'INGRESOS', type: 'income' },
  { code: '4.1', name: 'Ventas', type: 'income', parentCode: '4' },
  { code: '5', name: 'GASTOS', type: 'expense' },
  { code: '5.1', name: 'Costo de Ventas', type: 'expense', parentCode: '5' },
];

let accounts = [...MOCK_ACCOUNTS];
let entries: JournalEntry[] = [];

export const accountingService = {
  async getChartOfAccounts(): Promise<Account[]> {
    await sleep(150);
    return [...accounts];
  },
  async listEntries(): Promise<JournalEntry[]> {
    await sleep(150);
    return [...entries];
  },
  async postEntry(input: Omit<JournalEntry, 'id' | 'status'>): Promise<JournalEntry> {
    await sleep(200);
    const debit = input.lines.reduce((s, l) => s + l.debit, 0);
    const credit = input.lines.reduce((s, l) => s + l.credit, 0);
    if (Math.abs(debit - credit) > 0.01) throw new Error('Asiento descuadrado');
    const entry: JournalEntry = {
      id: `je-${Date.now()}`,
      status: 'posted',
      ...input,
    };
    entries = [entry, ...entries];
    return entry;
  },
};
