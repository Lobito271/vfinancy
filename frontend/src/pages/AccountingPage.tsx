import { BookOpen } from 'lucide-react';
import { ModulePage } from './ModulePage';

export function AccountingPage() {
  return (
    <ModulePage
      title="Contabilidad"
      subtitle="Plan contable, libros y estados financieros"
      icon={BookOpen}
      description="Plan contable, libro diario, libro mayor y estados financieros automáticos."
      phase="Fase 8"
      features={[
        'Plan de cuentas configurable',
        'Libro diario (asientos contables)',
        'Libro mayor',
        'Balance de comprobación',
        'Estado de resultados',
        'Balance general',
        'Flujo de caja',
        'Impuestos (IGV, Renta, otros)',
      ]}
    />
  );
}
