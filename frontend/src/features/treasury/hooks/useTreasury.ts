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
