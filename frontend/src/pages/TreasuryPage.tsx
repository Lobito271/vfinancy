import { useState } from 'react';
import { PageContainer, PageHeader, Grid, Section } from '@/components/layout';
import { Button } from '@/components/button';
import { Card, CardHeader, CardContent } from '@/components/card';
import { EmptyState, Spinner } from '@/components/feedback';
import { ConfirmDialog } from '@/components/dialog';
import { Can } from '@/components/auth';
import { Icons } from '@/design-system/icons';
import { Permissions } from '@/constants/permissions';
import { useCreditCards, useCardProjections, usePayCard } from '@/features/treasury/hooks/useTreasury';
import { formatCurrency, formatDate } from '@/utils/format';
import { useNotificationStore } from '@/stores/notification';

export function TreasuryPage() {
  const [payTarget, setPayTarget] = useState<{ cardId: string; issuer: string; lastFour: string; amount: number } | null>(null);

  const { data: creditCards = [], isLoading: cardsLoading } = useCreditCards();
  const { data: projections = [], isLoading: projectionsLoading } = useCardProjections();
  const payCardMutation = usePayCard();
  const push = useNotificationStore((s) => s.push);

  const loading = cardsLoading || projectionsLoading;

  return (
    <PageContainer>
      <PageHeader
        title="Tesorería"
        subtitle="Tarjetas de crédito, ciclos de facturación y proyección de pagos"
      />

      <Section
        title="Tarjetas de Crédito"
        description="Deuda proyectada por ciclo de facturación. El monto mostrado es lo que debes separar para cancelar al banco en la fecha de pago sin generar intereses."
      >
        {loading ? (
          <div className="page-loader">
            <Spinner />
          </div>
        ) : creditCards.length === 0 ? (
          <EmptyState
            title="No hay tarjetas de crédito"
            description="Registra una tarjeta de crédito para ver la proyección de pagos."
          />
        ) : (
          <Grid cols={2}>
            {creditCards.map((card) => {
              const projection = projections.find((p) => p.cardId === card.id);
              const projectedUSD = projection?.projectedUSD ?? 0;
              return (
                <Card key={card.id}>
                  <CardHeader className="flex-row items-center justify-between">
                    <div>
                      <p className="text-base font-semibold">
                        {card.issuer === 'visa' ? 'Visa' : 'Diners'} •••• {card.lastFour}
                      </p>
                      <p className="text-sm muted">{card.cardHolder}</p>
                    </div>
                    <Can permission={Permissions.Treasury.Create}>
                      <Button
                        variant="outline"
                        size="sm"
                        disabled={projectedUSD <= 0 || payCardMutation.isPending}
                        onClick={() =>
                          setPayTarget({
                            cardId: card.id,
                            issuer: card.issuer,
                            lastFour: card.lastFour,
                            amount: projectedUSD,
                          })
                        }
                      >
                        <Icons.Action.Payment /> Registrar pago
                      </Button>
                    </Can>
                  </CardHeader>
                  <CardContent>
                    <div className="form-grid">
                      <div>
                        <p className="muted">Fecha de corte</p>
                        <p className="font-medium tabular-nums">
                          {projection?.nextCutOffDate ? formatDate(projection.nextCutOffDate) : '—'}
                        </p>
                      </div>
                      <div>
                        <p className="muted">Fecha de pago</p>
                        <p className="font-medium tabular-nums">
                          {projection?.nextPaymentDate ? formatDate(projection.nextPaymentDate) : '—'}
                        </p>
                      </div>
                      <div className="col-span-2 pt-2 border-top">
                        <p className="muted">Deuda proyectada (USD)</p>
                        <p className="text-lg font-bold tabular-nums" style={{ color: 'hsl(var(--primary))' }}>
                          {formatCurrency(projectedUSD, 'USD')}
                        </p>
                      </div>
                      <div className="col-span-2 pt-2 border-top">
                        <div className="hstack" style={{ justifyContent: 'space-between' }}>
                          <span className="text-sm muted">Límite: {formatCurrency(card.creditLimit, 'USD')}</span>
                          <span className="text-sm muted">Disponible: {formatCurrency(card.creditLimit - card.currentBalance, 'USD')}</span>
                        </div>
                      </div>
                    </div>
                  </CardContent>
                </Card>
              );
            })}
          </Grid>
        )}
      </Section>

      <ConfirmDialog
        open={!!payTarget}
        onOpenChange={(open) => {
          if (!open) setPayTarget(null);
        }}
        title="Registrar pago de tarjeta"
        description={`¿Confirmar pago de ${payTarget ? formatCurrency(payTarget.amount, 'USD') : ''} para la tarjeta ${payTarget?.issuer === 'visa' ? 'Visa' : 'Diners'} •••• ${payTarget?.lastFour ?? ''}? Esta acción registrará el pago y reiniciará la proyección del ciclo actual.`}
        confirmLabel="Confirmar pago"
        loading={payCardMutation.isPending}
        onConfirm={() => {
          if (!payTarget) return;
          payCardMutation.mutate(
            { cardId: payTarget.cardId, amount: payTarget.amount },
            {
              onSuccess: () => {
                push({ title: 'Pago registrado', variant: 'success' });
                setPayTarget(null);
              },
              onError: (err: unknown) => {
                push({
                  title: 'No se pudo registrar el pago',
                  description: err instanceof Error ? err.message : undefined,
                  variant: 'destructive',
                });
                setPayTarget(null);
              },
            },
          );
        }}
      />
    </PageContainer>
  );
}
