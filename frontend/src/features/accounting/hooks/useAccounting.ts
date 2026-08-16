import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import {
  accountingService,
  type AccountInput,
  type AccountUpdateInput,
  type JournalEntryInput,
} from '@/services/accounting';
import { queryKeys } from '@/services/queryKeys';

export function useChartOfAccounts() {
  return useQuery({
    queryKey: queryKeys.accounting.chart,
    queryFn: () => accountingService.listChartOfAccounts(),
  });
}

export function useCreateChartOfAccount() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (input: AccountInput) => accountingService.createChartOfAccount(input),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: queryKeys.accounting.chart });
    },
  });
}

export function useUpdateChartOfAccount() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({ id, input }: { id: string; input: AccountUpdateInput }) =>
      accountingService.updateChartOfAccount(id, input),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: queryKeys.accounting.chart });
    },
  });
}

export function useDeleteChartOfAccount() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => accountingService.removeChartOfAccount(id),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: queryKeys.accounting.chart });
    },
  });
}

export function useJournalEntries() {
  return useQuery({
    queryKey: queryKeys.accounting.entries,
    queryFn: () => accountingService.listJournalEntries(),
  });
}

export function useCreateJournalEntry() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (input: JournalEntryInput) => accountingService.createJournalEntry(input),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: queryKeys.accounting.entries });
      qc.invalidateQueries({ queryKey: queryKeys.accounting.chart });
    },
  });
}

export function usePostJournalEntry() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => accountingService.postJournalEntry(id),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: queryKeys.accounting.entries });
    },
  });
}

export function useFiscalPeriods() {
  return useQuery({
    queryKey: ['accounting', 'periods'],
    queryFn: () => accountingService.listFiscalPeriods(),
  });
}
