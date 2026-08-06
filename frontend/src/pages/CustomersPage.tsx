import { Users } from 'lucide-react';
import { ModulePage } from './ModulePage';
import { customers } from '@/data/mock';
import { formatCurrency } from '@/lib/utils';

export function CustomersPage() {
  const totalDebt = customers.reduce((s, c) => s + c.currentDebt, 0);
  return (
    <ModulePage
      title="Clientes"
      subtitle="Gestión de clientes y cuentas por cobrar"
      icon={Users}
      description="Administre la información de sus clientes, límites de crédito y estado de cuenta."
      phase="Fase 3"
      features={[
        'Registro de clientes (DNI, RUC, CE, Pasaporte)',
        'Búsqueda y filtrado avanzado',
        'Historial de compras',
        'Estado de deuda y cuenta corriente',
        'Bloqueo automático por deuda vencida',
        'Límite de crédito configurable',
        'Exportación a PDF y Excel',
        'Importación masiva desde Excel',
      ]}
      mockStats={[
        { label: 'Total clientes', value: String(customers.length) },
        { label: 'Clientes activos', value: String(customers.filter((c) => c.status === 'active').length) },
        { label: 'Con deuda', value: String(customers.filter((c) => c.currentDebt > 0).length) },
        { label: 'Deuda total', value: formatCurrency(totalDebt) },
      ]}
    />
  );
}
