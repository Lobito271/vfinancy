import type {
  AccountDTO,
  CreateChartOfAccountRequest,
  CreateJournalEntryLineRequest,
  FiscalPeriodDTO,
  JournalEntryDTO,
  JournalEntryLineDTO,
  UpdateChartOfAccountRequest,
} from '../wails-types';
import { wailsClient } from '../bindings';

export type AccountType = 'asset' | 'liability' | 'equity' | 'income' | 'expense';

export interface Account {
  id: string;
  code: string;
  name: string;
  type: AccountType;
  parentId: string;
  path: string;
  depth: number;
  isActive: boolean;
  allowsMovement: boolean;
  description: string;
}

export type JournalEntryStatus = 'draft' | 'posted' | 'reversed';
export type JournalEntrySource = 'manual' | 'sale' | 'purchase' | 'payment' | 'receipt' | 'bank';

export interface JournalEntryLine {
  id: string;
  lineNumber: number;
  accountId: string;
  description: string;
  debit: number;
  credit: number;
}

export interface JournalEntry {
  id: string;
  number: string;
  entryDate: string;
  description: string;
  source: JournalEntrySource;
  sourceId: string;
  status: JournalEntryStatus;
  lines: JournalEntryLine[];
  createdAt: string;
}

export interface FiscalPeriod {
  id: string;
  name: string;
  periodStart: string;
  periodEnd: string;
  status: 'open' | 'closing' | 'closed';
}

export interface AccountInput {
  code: string;
  name: string;
  type: AccountType;
  parentId?: string;
  allowsMovement: boolean;
  description?: string;
}

export interface AccountUpdateInput extends AccountInput {
  isActive: boolean;
}

export interface JournalEntryLineInput {
  accountId: string;
  description?: string;
  debit: number;
  credit: number;
}

export interface JournalEntryInput {
  entryDate: string;
  description: string;
  lines: JournalEntryLineInput[];
}

function toAccount(dto: AccountDTO): Account {
  return {
    id: dto.id,
    code: dto.code,
    name: dto.name,
    type: dto.type as AccountType,
    parentId: dto.parentId,
    path: dto.path,
    depth: dto.depth,
    isActive: dto.isActive,
    allowsMovement: dto.allowsMovement,
    description: dto.description,
  };
}

function toLine(dto: JournalEntryLineDTO): JournalEntryLine {
  return {
    id: dto.id,
    lineNumber: dto.lineNumber,
    accountId: dto.accountId,
    description: dto.description,
    debit: Number(dto.debit),
    credit: Number(dto.credit),
  };
}

function toEntry(dto: JournalEntryDTO): JournalEntry {
  return {
    id: dto.id,
    number: dto.number,
    entryDate: dto.entryDate,
    description: dto.description,
    source: dto.source as JournalEntrySource,
    sourceId: dto.sourceId,
    status: dto.status as JournalEntryStatus,
    lines: dto.lines.map(toLine),
    createdAt: dto.createdAt,
  };
}

function toPeriod(dto: FiscalPeriodDTO): FiscalPeriod {
  return {
    id: dto.id,
    name: dto.name,
    periodStart: dto.periodStart,
    periodEnd: dto.periodEnd,
    status: dto.status as FiscalPeriod['status'],
  };
}

function toCreateRequest(input: AccountInput): CreateChartOfAccountRequest {
  return {
    code: input.code,
    name: input.name,
    type: input.type,
    parentId: input.parentId ?? '',
    allowsMovement: input.allowsMovement,
    description: input.description ?? '',
  };
}

function toLineRequest(line: JournalEntryLineInput): CreateJournalEntryLineRequest {
  return {
    accountId: line.accountId,
    description: line.description ?? '',
    debit: line.debit.toFixed(2),
    credit: line.credit.toFixed(2),
  };
}

export const accountingService = {
  async listChartOfAccounts(): Promise<Account[]> {
    return (await wailsClient.listChartOfAccounts()).map(toAccount);
  },
  async createChartOfAccount(input: AccountInput): Promise<Account> {
    return toAccount(await wailsClient.createChartOfAccount(toCreateRequest(input)));
  },
  async updateChartOfAccount(id: string, input: AccountUpdateInput): Promise<Account> {
    const req: UpdateChartOfAccountRequest = {
      ...toCreateRequest(input),
      id,
      isActive: input.isActive,
    };
    return toAccount(await wailsClient.updateChartOfAccount(req));
  },
  async removeChartOfAccount(id: string): Promise<void> {
    await wailsClient.deleteChartOfAccount(id);
  },
  async listJournalEntries(status?: string): Promise<JournalEntry[]> {
    const res = await wailsClient.listJournalEntries({
      status: status ?? '',
      page: 1,
      pageSize: 200,
    });
    return res.items.map(toEntry);
  },
  async createJournalEntry(input: JournalEntryInput): Promise<JournalEntry> {
    return toEntry(
      await wailsClient.createJournalEntry({
        fiscalPeriodId: '',
        entryDate: input.entryDate,
        description: input.description,
        lines: input.lines.map(toLineRequest),
      }),
    );
  },
  async postJournalEntry(id: string): Promise<JournalEntry> {
    return toEntry(await wailsClient.postJournalEntry(id));
  },
  async listFiscalPeriods(): Promise<FiscalPeriod[]> {
    return (await wailsClient.listFiscalPeriods()).map(toPeriod);
  },
};
