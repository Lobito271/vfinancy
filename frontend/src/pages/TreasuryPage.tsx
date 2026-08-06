import { Landmark } from 'lucide-react';
import { ModulePage } from './ModulePage';

export function TreasuryPage() {
  return (
    <ModulePage
      title="Tesorería"
      subtitle="Cuentas bancarias, tarjetas y conciliaciones"
      icon={Landmark}
      description="Gestione cuentas bancarias, tarjetas de crédito, conciliaciones y flujo de caja."
      phase="Fase 7"
      features={[
        'Cuentas bancarias (múltiples monedas)',
        'Tarjetas de crédito',
        'Transacciones bancarias y de tarjeta',
        'Conciliación bancaria automática',
        'Devoluciones internacionales',
        'Tipo de cambio diario',
        'Flujo de caja proyectado',
        'Reportes de movimientos',
      ]}
    />
  );
}
