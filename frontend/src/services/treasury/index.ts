import type {
  BankAccountDTO,
  BankTransactionDTO,
  CreateBankAccountRequest,
  CreateBankTransactionRequest,
  UpdateBankAccountRequest,
} from '../wails-types';
import { wailsClient } from '../bindings';

export type BankAccountType = 'checking' | 'savings';

export interface BankAccount {
  id: string;
  bank: string;
  accountNumber: string;
  accountType: BankAccountType;
  currency: string;
  balance: number;
  isDefault: boolean;
  isActive: boolean;
}

export type BankTxType = 'deposit' | 'withdrawal' | 'fee' | 'interest' | 'transfer' | 'other';

export interface BankTransaction {
  id: string;
  accountId: string;
  date: string;
  description: string;
  amount: number;
  type: BankTxType;
  balanceAfter: number;
  reference: string;
  isReconciled: boolean;
}

export interface BankAccountInput {
  bank: string;
  accountNumber: string;
  accountType: BankAccountType;
  currency: string;
  isDefault: boolean;
}

export interface BankAccountUpdateInput extends BankAccountInput {
  isActive: boolean;
}

export interface BankTransactionInput {
  accountId: string;
  date: string;
  description: string;
  amount: number;
  type: BankTxType;
  reference?: string;
}

function toAccount(dto: BankAccountDTO): BankAccount {
  return {
    id: dto.id,
    bank: dto.bankName,
    accountNumber: dto.accountNumber,
    accountType: dto.accountType as BankAccountType,
    currency: dto.currencyCode,
    balance: Number(dto.currentBalance),
    isDefault: dto.isDefault,
    isActive: dto.isActive,
  };
}

function toTransaction(dto: BankTransactionDTO): BankTransaction {
  return {
    id: dto.id,
    accountId: dto.accountId,
    date: dto.date,
    description: dto.description,
    amount: Number(dto.amount),
    type: dto.type as BankTxType,
    balanceAfter: Number(dto.balanceAfter),
    reference: dto.reference,
    isReconciled: dto.isReconciled,
  };
}

function toCreateRequest(input: BankAccountInput): CreateBankAccountRequest {
  return {
    bankName: input.bank,
    accountNumber: input.accountNumber,
    accountType: input.accountType,
    currencyCode: input.currency,
    isDefault: input.isDefault,
  };
}

export const treasuryService = {
  async listAccounts(): Promise<BankAccount[]> {
    return (await wailsClient.listBankAccounts()).map(toAccount);
  },
  async getAccount(id: string): Promise<BankAccount | null> {
    try {
      return toAccount(await wailsClient.getBankAccount(id));
    } catch {
      return null;
    }
  },
  async createAccount(input: BankAccountInput): Promise<BankAccount> {
    return toAccount(await wailsClient.createBankAccount(toCreateRequest(input)));
  },
  async updateAccount(id: string, input: BankAccountUpdateInput): Promise<BankAccount> {
    const req: UpdateBankAccountRequest = {
      ...toCreateRequest(input),
      id,
      isActive: input.isActive,
    };
    return toAccount(await wailsClient.updateBankAccount(req));
  },
  async removeAccount(id: string): Promise<void> {
    await wailsClient.deleteBankAccount(id);
  },
  async listTransactions(accountId?: string): Promise<BankTransaction[]> {
    const res = await wailsClient.listBankTransactions({
      accountId: accountId ?? '',
      reconciled: null,
      page: 1,
      pageSize: 200,
    });
    return res.items.map(toTransaction);
  },
  async createTransaction(input: BankTransactionInput): Promise<BankTransaction> {
    const req: CreateBankTransactionRequest = {
      accountId: input.accountId,
      date: input.date,
      description: input.description,
      amount: input.amount.toFixed(2),
      type: input.type,
      reference: input.reference ?? '',
    };
    return toTransaction(await wailsClient.createBankTransaction(req));
  },
  async conciliate(id: string): Promise<BankTransaction> {
    return toTransaction(await wailsClient.reconcileBankTransaction(id));
  },
  async getExchangeRate(from: string, to: string): Promise<number> {
    if (from === to) return 1;
    return Number(await wailsClient.latestExchangeRate(from, to));
  },
};
