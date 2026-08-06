import { ShoppingCart } from 'lucide-react';
import { ModulePage } from './ModulePage';
import { purchases } from '@/data/mock';
import { formatCurrency } from '@/lib/utils';

export function PurchasesPage() {
  return (
    <ModulePage
      title="Compras"
      subtitle="Órdenes de compra y cuentas por pagar"
      icon={ShoppingCart}
      description="Registre órdenes de compra, gestione pagos a proveedores y conciliaciones."
      phase="Fase 4"
      features={[
        'Órdenes de compra (Pendiente, Pagado, Conciliado, Cancelado)',
        'Compras al contado y al crédito',
        'Pagos a proveedores',
        'Compras internacionales con tipo de cambio',
        'Conciliación con facturas',
        'Devoluciones',
        'Reportes por proveedor y periodo',
      ]}
      mockStats={[
        { label: 'Órdenes registradas', value: String(purchases.length) },
        { label: 'Pendientes', value: String(purchases.filter((p) => p.status === 'pending').length) },
        { label: 'Total comprado', value: formatCurrency(purchases.reduce((s, p) => s + p.total, 0)) },
      ]}
    />
  );
}
