import type {
  BankAccountDTO,
  BankTransactionDTO,
  CardProjectionDTO,
  CreateBankAccountRequest,
  CreateBankTransactionRequest,
  CreditCardDTO,
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

export interface CreditCard {
  id: string;
  issuer: string;
  lastFour: string;
  cardHolder: string;
  expirationMonth: number;
  expirationYear: number;
  creditLimit: number;
  currentBalance: number;
  cutOffDay: number;
  paymentDueDay: number;
  currencyCode: string;
  isActive: boolean;
}

export interface CardProjection {
  cardId: string;
  issuer: string;
  lastFour: string;
  cardHolder: string;
  projectedUSD: number;
  cycleStart: string;
  nextCutOffDate: string;
  nextPaymentDate: string;
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

function toCreditCard(dto: CreditCardDTO): CreditCard {
  return {
    id: dto.id,
    issuer: dto.issuer,
    lastFour: dto.lastFour,
    cardHolder: dto.cardHolder,
    expirationMonth: dto.expirationMonth,
    expirationYear: dto.expirationYear,
    creditLimit: Number(dto.creditLimit),
    currentBalance: Number(dto.currentBalance),
    cutOffDay: dto.cutOffDay,
    paymentDueDay: dto.paymentDueDay,
    currencyCode: dto.currencyCode,
    isActive: dto.isActive,
  };
}

function toCardProjection(dto: CardProjectionDTO): CardProjection {
  return {
    cardId: dto.cardId,
    issuer: dto.issuer,
    lastFour: dto.lastFour,
    cardHolder: dto.cardHolder,
    projectedUSD: dto.projectedUSD,
    cycleStart: dto.cycleStart,
    nextCutOffDate: dto.nextCutOffDate,
    nextPaymentDate: dto.nextPaymentDate,
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
  async listCreditCards(): Promise<CreditCard[]> {
    return (await wailsClient.listCreditCards()).map(toCreditCard);
  },
  async getCardProjections(): Promise<CardProjection[]> {
    return (await wailsClient.getCardProjections()).map(toCardProjection);
  },
  async payCard(cardId: string, amount: number): Promise<void> {
    await wailsClient.payCreditCard(cardId, amount);
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
