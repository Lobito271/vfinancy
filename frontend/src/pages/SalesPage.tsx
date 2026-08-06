import { Receipt } from 'lucide-react';
import { ModulePage } from './ModulePage';
import { sales } from '@/data/mock';
import { formatCurrency } from '@/lib/utils';

export function SalesPage() {
  return (
    <ModulePage
      title="Ventas"
      subtitle="Registro de ventas y cuentas por cobrar"
      icon={Receipt}
      description="Registre ventas, calcule márgenes y gestione los cobros a clientes."
      phase="Fase 5"
      features={[
        'Registro de ventas con múltiples items',
        'Estados (Pendiente, Parcial, Pagado, Cancelado)',
        'Cálculo automático de margen y ganancia',
        'Impuestos (IGV) y descuentos',
        'Cuentas por cobrar',
        'Anticipos de clientes',
        'Pagos parciales y aplicación',
        'Notas de crédito / devoluciones',
      ]}
      mockStats={[
        { label: 'Ventas registradas', value: String(sales.length) },
        { label: 'Pagadas', value: String(sales.filter((s) => s.status === 'paid').length) },
        { label: 'Pendientes', value: String(sales.filter((s) => s.status === 'pending').length) },
        { label: 'Total vendido', value: formatCurrency(sales.reduce((s, x) => s + x.total, 0)) },
      ]}
    />
  );
}
