import { useState } from 'react';
import { CreditCard, Plus, Pencil, Trash2 } from 'lucide-react';
import { PageContainer, PageHeader, Grid, Section } from '@/components/layout';
import { Button } from '@/components/button';
import { Card, CardHeader, CardContent } from '@/components/card';
import { EmptyState, Spinner } from '@/components/feedback';
import { ConfirmDialog } from '@/components/dialog';
import { RowActions } from '@/components/misc';
import { useCreditCards, useCardProjections, usePayCard, useDeleteCreditCard } from '@/features/treasury/hooks/useTreasury';
import { CreditCardFormDialog } from '@/features/treasury/components/CreditCardFormDialog';
import { formatCurrency, formatDate } from '@/utils/format';
import { useNotificationStore } from '@/stores/notification';

interface CardTarget {
  id: string;
  issuer: string;
  lastFour: string;
  cardHolder: string;
  creditLimit: number;
  cutOffDay: number;
  paymentDueDay: number;
  currencyCode: string;
  isActive: boolean;
}

export function TreasuryPage() {
  const [payTarget, setPayTarget] = useState<{ cardId: string; issuer: string; lastFour: string; amount: number } | null>(null);
  const [formOpen, setFormOpen] = useState(false);
  const [editCard, setEditCard] = useState<CardTarget | null>(null);
  const [deleteTarget, setDeleteTarget] = useState<CardTarget | null>(null);

  const { data: creditCards = [], isLoading: cardsLoading } = useCreditCards();
  const { data: projections = [], isLoading: projectionsLoading } = useCardProjections();
  const payCardMutation = usePayCard();
  const deleteCardMutation = useDeleteCreditCard();
  const push = useNotificationStore((s) => s.push);

  const loading = cardsLoading || projectionsLoading;

  function openCreate() {
    setEditCard(null);
    setFormOpen(true);
  }

  function openEdit(card: CardTarget) {
    setEditCard(card);
    setFormOpen(true);
  }

  return (
    <PageContainer>
      <PageHeader
        title="Tesorería"
        subtitle="Tarjetas de crédito, ciclos de facturación y proyección de pagos"
        actions={
          <Button onClick={openCreate}>
            <Plus /> Nueva tarjeta
          </Button>
        }
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
                  <CardHeader className="card-header--row">
                    <div className="vstack">
                      <p className="card-title">{card.issuer === 'visa' ? 'Visa' : card.issuer === 'diners' ? 'Diners' : card.issuer} •••• {card.lastFour}</p>
                      <p className="muted">{card.cardHolder}</p>
                    </div>
                    <div className="hstack" style={{ gap: '0.5rem' }}>
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
                        <CreditCard /> Registrar pago
                      </Button>
                      <RowActions
                        label={`Acciones de ${card.issuer} •••• ${card.lastFour}`}
                        actions={[
                          {
                            label: 'Editar',
                            icon: Pencil,
                            onSelect: () =>
                              openEdit({
                                id: card.id,
                                issuer: card.issuer,
                                lastFour: card.lastFour,
                                cardHolder: card.cardHolder,
                                creditLimit: card.creditLimit,
                                cutOffDay: card.cutOffDay,
                                paymentDueDay: card.paymentDueDay,
                                currencyCode: card.currencyCode,
                                isActive: card.isActive,
                              }),
                          },
                          {
                            label: 'Eliminar',
                            icon: Trash2,
                            danger: true,
                            disabled: card.currentBalance > 0,
                            onSelect: () =>
                              setDeleteTarget({
                                id: card.id,
                                issuer: card.issuer,
                                lastFour: card.lastFour,
                                cardHolder: card.cardHolder,
                                creditLimit: card.creditLimit,
                                cutOffDay: card.cutOffDay,
                                paymentDueDay: card.paymentDueDay,
                                currencyCode: card.currencyCode,
                                isActive: card.isActive,
                              }),
                          },
                        ]}
                      />
                    </div>
                  </CardHeader>
                  <CardContent>
                    <div className="treasury-grid">
                      <div>
                        <p className="muted">Fecha de corte</p>
                        <p className="fw-medium tabular">
                          {projection?.nextCutOffDate ? formatDate(projection.nextCutOffDate) : '—'}
                        </p>
                      </div>
                      <div>
                        <p className="muted">Fecha de pago</p>
                        <p className="fw-medium tabular">
                          {projection?.nextPaymentDate ? formatDate(projection.nextPaymentDate) : '—'}
                        </p>
                      </div>
                      <div className="treasury-grid__divider">
                        <p className="muted">Deuda proyectada (USD)</p>
                        <p className="treasury-grid__amount tabular">{formatCurrency(projectedUSD, 'USD')}</p>
                      </div>
                      <div className="treasury-grid__divider">
                        <div className="hstack" style={{ justifyContent: 'space-between' }}>
                          <span className="muted">Límite: {formatCurrency(card.creditLimit, 'USD')}</span>
                          <span className="muted">Disponible: {formatCurrency(card.creditLimit - card.currentBalance, 'USD')}</span>
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

      <CreditCardFormDialog open={formOpen} onOpenChange={setFormOpen} editCard={editCard} />

      <ConfirmDialog
        open={!!deleteTarget}
        onOpenChange={(open) => {
          if (!open) setDeleteTarget(null);
        }}
        title="Eliminar tarjeta"
        description={
          deleteTarget
            ? `¿Eliminar la tarjeta ${deleteTarget.issuer} •••• ${deleteTarget.lastFour}? La tarjeta se desactivará y ya no estará disponible para nuevas compras.`
            : undefined
        }
        confirmLabel="Eliminar"
        loading={deleteCardMutation.isPending}
        onConfirm={() => {
          if (!deleteTarget) return;
          deleteCardMutation.mutate(deleteTarget.id, {
            onSuccess: () => {
              push({ title: 'Tarjeta eliminada', variant: 'success' });
              setDeleteTarget(null);
            },
            onError: (err: unknown) => {
              push({
                title: 'No se pudo eliminar',
                description: err instanceof Error ? err.message : undefined,
                variant: 'destructive',
              });
              setDeleteTarget(null);
            },
          });
        }}
      />

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
