import { Truck } from 'lucide-react';
import { ModulePage } from './ModulePage';
import { suppliers } from '@/data/mock';
import { formatCurrency } from '@/lib/utils';

export function SuppliersPage() {
  const totalDebt = suppliers.reduce((s, x) => s + x.currentDebt, 0);
  return (
    <ModulePage
      title="Proveedores"
      subtitle="Gestión de proveedores y cuentas por pagar"
      icon={Truck}
      description="Administre la información de sus proveedores, contactos y cuentas por pagar."
      phase="Fase 3"
      features={[
        'Registro de proveedores nacionales e internacionales',
        'Múltiples contactos por proveedor',
        'Soporte de múltiples monedas (PEN, USD, EUR)',
        'Historial de compras',
        'Cuentas por pagar',
        'Importaciones y tipo de cambio',
        'Exportación de saldos',
      ]}
      mockStats={[
        { label: 'Total proveedores', value: String(suppliers.length) },
        { label: 'Activos', value: String(suppliers.filter((s) => s.status === 'active').length) },
        { label: 'Deuda total', value: formatCurrency(totalDebt) },
      ]}
    />
  );
}
