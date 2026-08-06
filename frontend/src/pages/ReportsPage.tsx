import { BarChart3 } from 'lucide-react';
import { ModulePage } from './ModulePage';

export function ReportsPage() {
  return (
    <ModulePage
      title="Reportes"
      subtitle="Reportes financieros y operativos"
      icon={BarChart3}
      description="Genere reportes detallados para análisis y toma de decisiones."
      phase="Fase 10"
      features={[
        'Reporte de ventas por periodo, cliente, producto',
        'Reporte de compras por proveedor',
        'Estado de cuentas por cobrar / pagar',
        'Inventario valorizado',
        'Rentabilidad por producto',
        'Exportación a PDF, Excel y CSV',
        'Programación de reportes automáticos',
        'Envío por correo electrónico',
      ]}
    />
  );
}
