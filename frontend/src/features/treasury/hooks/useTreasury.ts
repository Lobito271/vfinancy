import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { treasuryService, type CreditCardInput } from '@/services/treasury';
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

function useInvalidateCards() {
  const qc = useQueryClient();
  return () => {
    qc.invalidateQueries({ queryKey: queryKeys.treasury.cardProjections });
    qc.invalidateQueries({ queryKey: queryKeys.treasury.creditCards });
  };
}

export function usePayCard() {
  const invalidate = useInvalidateCards();
  return useMutation({
    mutationFn: ({ cardId, amount }: { cardId: string; amount: number }) =>
      treasuryService.payCard(cardId, amount),
    onSuccess: invalidate,
  });
}

export function useCreateCreditCard() {
  const invalidate = useInvalidateCards();
  return useMutation({
    mutationFn: (input: CreditCardInput) => treasuryService.createCreditCard(input),
    onSuccess: invalidate,
  });
}

export function useUpdateCreditCard() {
  const invalidate = useInvalidateCards();
  return useMutation({
    mutationFn: ({ id, ...input }: { id: string } & CreditCardInput) =>
      treasuryService.updateCreditCard(id, input),
    onSuccess: invalidate,
  });
}

export function useDeleteCreditCard() {
  const invalidate = useInvalidateCards();
  return useMutation({
    mutationFn: (id: string) => treasuryService.deleteCreditCard(id),
    onSuccess: invalidate,
  });
}
