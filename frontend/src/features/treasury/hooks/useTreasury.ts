import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import {
  treasuryService,
  type BankAccountInput,
  type BankAccountUpdateInput,
  type BankTransactionInput,
} from '@/services/treasury';
import { queryKeys } from '@/services/queryKeys';

export function useBankAccounts() {
  return useQuery({
    queryKey: queryKeys.treasury.accounts,
    queryFn: () => treasuryService.listAccounts(),
  });
}

export function useBankAccount(id: string | undefined) {
  return useQuery({
    queryKey: queryKeys.treasury.accounts,
    queryFn: () => treasuryService.listAccounts(),
    enabled: !!id,
    select: (accounts) => accounts.find((a) => a.id === id),
  });
}

export function useCreditCards() {
  return useQuery({
    queryKey: queryKeys.treasury.creditCards,
    queryFn: () => treasuryService.listCreditCards(),
    staleTime: 60 * 1000,
  });
}

export function useCardProjections() {
  return useQuery({
    queryKey: queryKeys.treasury.cardProjections,
    queryFn: () => treasuryService.getCardProjections(),
    staleTime: 60 * 1000,
  });
}

export function usePayCard() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ cardId, amount }: { cardId: string; amount: number }) =>
      treasuryService.payCard(cardId, amount),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: queryKeys.treasury.cardProjections });
      qc.invalidateQueries({ queryKey: queryKeys.treasury.creditCards });
    },
  });
}

export function useCreateBankAccount() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (input: BankAccountInput) => treasuryService.createAccount(input),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: queryKeys.treasury.accounts });
    },
  });
}

export function useUpdateBankAccount() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, input }: { id: string; input: BankAccountUpdateInput }) =>
      treasuryService.updateAccount(id, input),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: queryKeys.treasury.accounts });
    },
  });
}

export function useDeleteBankAccount() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => treasuryService.removeAccount(id),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: queryKeys.treasury.accounts });
    },
  });
}

export function useBankTransactions(accountId?: string) {
  return useQuery({
    queryKey: queryKeys.treasury.transactions(accountId ?? 'all'),
    queryFn: () => treasuryService.listTransactions(accountId),
  });
}

export function useCreateBankTransaction() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (input: BankTransactionInput) => treasuryService.createTransaction(input),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: queryKeys.treasury.accounts });
      qc.invalidateQueries({ queryKey: ['treasury', 'transactions'] });
    },
  });
}

export function useReconcileBankTransaction() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => treasuryService.conciliate(id),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: ['treasury', 'transactions'] });
    },
  });
}
