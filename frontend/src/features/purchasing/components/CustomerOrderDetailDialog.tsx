import { useMemo } from 'react';
import { Dialog, DialogContent, DialogFooter, DialogHeader, DialogTitle } from '@/components/dialog';
import { Badge } from '@/components/badge';
import { DataTable, type Column } from '@/components/table';
import { Grid } from '@/components/layout';
import { Button } from '@/components/button';
import { AlertTriangle, Download } from 'lucide-react';
import { useCustomerOrder } from '@/features/purchasing/hooks/usePurchases';
import type { CustomerOrder, CustomerOrderItem, CustomerOrderPayment } from '@/types/domain';
import { formatCurrency, formatDate } from '@/utils/format';

const statusMap: Record<string, { variant: 'success' | 'warning' | 'info' | 'destructive' | 'muted'; label: string }> = {
  pending: { variant: 'warning', label: 'Pendiente' },
  received: { variant: 'info', label: 'Recibida' },
  paid: { variant: 'success', label: 'Pagada' },
  reconciled: { variant: 'success', label: 'Conciliada' },
  cancelled: { variant: 'destructive', label: 'Anulada' },
};

const itemColumns: Column<CustomerOrderItem>[] = [
  { id: 'description', header: 'Producto', cell: (r) => r.description || r.productId },
  { id: 'quantity', header: 'Cant.', align: 'right', cell: (r) => r.quantity },
  {
    id: 'unitPrice',
    header: 'Costo unit.',
    align: 'right',
    cell: (r) => <span className="tabular-nums">{formatCurrency(r.unitPrice, 'USD')}</span>,
  },
  {
    id: 'discountPercent',
    header: 'Dscto %',
    align: 'right',
    cell: (r) => `${r.discountPercent}%`,
  },
  {
    id: 'taxAmount',
    header: 'IGV',
    align: 'right',
    cell: (r) => <span className="tabular-nums">{formatCurrency(r.taxAmount, 'USD')}</span>,
  },
  {
    id: 'lineTotal',
    header: 'Total',
    align: 'right',
    cell: (r) => (
      <span className="tabular-nums font-medium">
        {formatCurrency(r.quantity * r.unitPrice - r.discountAmount + r.taxAmount, 'USD')}
      </span>
    ),
  },
];

const paymentColumns: Column<CustomerOrderPayment>[] = [
  { id: 'number', header: 'N°', cell: (r) => <span className="font-medium tabular-nums">{r.number}</span> },
  { id: 'paymentDate', header: 'Fecha', cell: (r) => <span className="text-muted-foreground">{formatDate(r.paymentDate)}</span> },
  {
    id: 'amount',
    header: 'Monto',
    align: 'right',
    cell: (r) => <span className="tabular-nums">{formatCurrency(r.amount)}</span>,
  },
  {
    id: 'method',
    header: 'Método',
    cell: (r) => {
      const label: Record<string, string> = {
        cash: 'Efectivo',
        bank_transfer: 'Transferencia',
        check: 'Cheque',
        card: 'Tarjeta',
        credit: 'Crédito',
        other: 'Otro',
      };
      return label[r.method] ?? r.method;
    },
  },
  {
    id: 'status',
    header: 'Estado',
    cell: (r) =>
      r.status === 'refunded' ? (
        <Badge variant="destructive">Reembolsado</Badge>
      ) : (
        <Badge variant="success">Activo</Badge>
      ),
  },
];

interface CustomerOrderDetailDialogProps {
  order: CustomerOrder | null;
  onOpenChange: (open: boolean) => void;
  onMarkReceived?: (order: CustomerOrder) => void;
  onMarkFaulty?: (order: CustomerOrder) => void;
}

export function CustomerOrderDetailDialog({ order, onOpenChange, onMarkReceived, onMarkFaulty }: CustomerOrderDetailDialogProps) {
  const open = Boolean(order);
  const { data: detail } = useCustomerOrder(order?.id);
  const current = detail ?? order;
  const cfg = current ? (statusMap[current.status] ?? { variant: 'muted' as const, label: current.status }) : null;

  const items = useMemo(() => current?.items ?? [], [current]);
  const payments = useMemo(() => current?.payments ?? [], [current]);

  const receivable = Boolean(current && current.status === 'pending' && !current.faulty);

  const handleReceived = () => {
    if (!current) return;
    onOpenChange(false);
    onMarkReceived?.(current);
  };

  const handleFaulty = () => {
    if (!current) return;
    onOpenChange(false);
    onMarkFaulty?.(current);
  };

  return (
    <Dialog open={open} onOpenChange={onOpenChange}>
      <DialogContent size="xl">
        {current && (
          <>
            <DialogHeader>
              <DialogTitle>
                Pedido <span className="tabular-nums">{current.number}</span>
              </DialogTitle>
              <div className="flex items-center gap-3 text-sm text-muted-foreground">
                <span>{current.customerName || 'Cliente'}</span>
                <span>·</span>
                <span>{formatDate(current.date)}</span>
                {current.supplierOrderNumber && (
                  <>
                    <span>·</span>
                    <span>Orden Prov: {current.supplierOrderNumber}</span>
                  </>
                )}
                {cfg && <Badge variant={cfg.variant}>{cfg.label}</Badge>}
                {current.faulty && <Badge variant="destructive">Llegó en mal estado</Badge>}
              </div>
            </DialogHeader>

            <div className="space-y-5">
              <Grid cols={4}>
                <div className="rounded-md border bg-muted/30 p-3">
                  <div className="text-xs text-muted-foreground">Costo real (PEN)</div>
                  <div className="text-lg font-semibold tabular-nums">{formatCurrency(current.realCostPEN)}</div>
                </div>
                <div className="rounded-md border bg-muted/30 p-3">
                  <div className="text-xs text-muted-foreground">Precio de venta</div>
                  <div className="text-lg font-semibold tabular-nums">{formatCurrency(current.salePricePEN)}</div>
                </div>
                <div className="rounded-md border bg-muted/30 p-3">
                  <div className="text-xs text-muted-foreground">Anticipo</div>
                  <div className="text-lg font-semibold tabular-nums">{formatCurrency(current.anticipo)}</div>
                </div>
                <div className="rounded-md border bg-muted/30 p-3">
                  <div className="text-xs text-muted-foreground">Por cobrar</div>
                  <div className="text-lg font-semibold tabular-nums">{formatCurrency(current.porCobrar)}</div>
                </div>
              </Grid>

              <div>
                <h3 className="mb-2 text-sm font-medium">Líneas del pedido</h3>
                <DataTable columns={itemColumns} data={items} keyField="id" globalSearch={false} exportable={false} />
              </div>

              <div>
                <h3 className="mb-2 text-sm font-medium">Pagos del cliente</h3>
                {payments.length === 0 ? (
                  <p className="rounded-md border bg-muted/20 p-3 text-sm text-muted-foreground">
                    Aún no se han registrado anticipos para este pedido.
                  </p>
                ) : (
                  <DataTable columns={paymentColumns} data={payments} keyField="id" globalSearch={false} exportable={false} />
                )}
              </div>

              {current.faulty && (
                <div className="rounded-md border border-destructive/30 bg-destructive/10 p-3 text-sm">
                  <span className="font-medium">Motivo del daño:</span>{' '}
                  {current.faultyReason || 'Sin detalle'} · Reembolsado:{' '}
                  <span className="tabular-nums">{formatCurrency(current.refundedAmount)}</span>
                </div>
              )}
            </div>

            <DialogFooter className="border-t pt-4">
              {receivable && (
                <>
                  <Button variant="outline" onClick={handleFaulty}>
                    <AlertTriangle /> Llegó en mal estado
                  </Button>
                  <Button onClick={handleReceived}>
                    <Download /> Llegada y Cobro
                  </Button>
                </>
              )}
            </DialogFooter>
          </>
        )}
      </DialogContent>
    </Dialog>
  );
}
