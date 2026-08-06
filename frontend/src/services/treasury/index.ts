import { sleep } from '@/utils';

export interface AccountBalance {
  id: string;
  bank: string;
  accountNumber: string;
  currency: string;
  balance: number;
  type: 'checking' | 'savings' | 'credit';
}

export interface BankTransaction {
  id: string;
  accountId: string;
  date: string;
  description: string;
  amount: number;
  type: 'credit' | 'debit';
  reconciled: boolean;
}

const MOCK_ACCOUNTS: AccountBalance[] = [
  { id: 'a-1', bank: 'BCP', accountNumber: '193-1234567-0-99', currency: 'PEN', balance: 48230.5, type: 'checking' },
  { id: 'a-2', bank: 'BBVA', accountNumber: '0011-0201-0201-0201', currency: 'PEN', balance: 21500, type: 'savings' },
  { id: 'a-3', bank: 'BCP', accountNumber: '193-9876543-0-55', currency: 'USD', balance: 5200, type: 'checking' },
];

let accounts = [...MOCK_ACCOUNTS];
let transactions: BankTransaction[] = [];

export const treasuryService = {
  async listAccounts(): Promise<AccountBalance[]> {
    await sleep(150);
    return [...accounts];
  },
  async getAccount(id: string): Promise<AccountBalance | null> {
    await sleep(100);
    return accounts.find((a) => a.id === id) ?? null;
  },
  async listTransactions(accountId?: string): Promise<BankTransaction[]> {
    await sleep(150);
    return accountId ? transactions.filter((t) => t.accountId === accountId) : [...transactions];
  },
  async conciliate(id: string): Promise<BankTransaction> {
    await sleep(150);
    const idx = transactions.findIndex((t) => t.id === id);
    if (idx < 0) throw new Error('Transaction not found');
    const updated = { ...transactions[idx], reconciled: true };
    transactions = [...transactions.slice(0, idx), updated, ...transactions.slice(idx + 1)];
    return updated;
  },
  async getExchangeRate(from: string, to: string): Promise<number> {
    await sleep(100);
    if (from === to) return 1;
    if (from === 'USD' && to === 'PEN') return 3.75;
    if (from === 'PEN' && to === 'USD') return 1 / 3.75;
    return 1;
  },
};
