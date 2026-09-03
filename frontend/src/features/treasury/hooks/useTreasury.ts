import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { treasuryService } from '@/services/treasury';
import { queryKeys } from '@/services/queryKeys';

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

export function useCreateCreditCard() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (input: {
      issuer: string;
      lastFour: string;
      cardHolder: string;
      expirationMonth: number;
      expirationYear: number;
      creditLimit: number;
      cutOffDay: number;
      paymentDueDay: number;
      currencyCode: string;
    }) => treasuryService.createCreditCard(input),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: queryKeys.treasury.creditCards });
      qc.invalidateQueries({ queryKey: queryKeys.treasury.cardProjections });
    },
  });
}

export function useUpdateCreditCard() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: ({
      id,
      ...input
    }: {
      id: string;
      issuer: string;
      lastFour: string;
      cardHolder: string;
      creditLimit: number;
      cutOffDay: number;
      paymentDueDay: number;
      isActive: boolean;
    }) => treasuryService.updateCreditCard(id, input),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: queryKeys.treasury.creditCards });
      qc.invalidateQueries({ queryKey: queryKeys.treasury.cardProjections });
    },
  });
}

export function useDeleteCreditCard() {
  const qc = useQueryClient();
  return useMutation({
    mutationFn: (id: string) => treasuryService.deleteCreditCard(id),
    onSuccess: () => {
      qc.invalidateQueries({ queryKey: queryKeys.treasury.creditCards });
      qc.invalidateQueries({ queryKey: queryKeys.treasury.cardProjections });
    },
  });
}
