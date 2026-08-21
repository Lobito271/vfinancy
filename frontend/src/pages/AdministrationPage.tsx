import { Shield } from 'lucide-react';
import { ModulePage } from './ModulePage';

export function AdministrationPage() {
  return (
    <ModulePage
      title="Administración"
      subtitle="Configuración avanzada del sistema"
      icon={Shield}
      description="Administre empresas y configuración avanzada del sistema."
      phase="Fase 2"
      features={[
        'Gestión de empresas (crear, editar, desactivar)',
        'Perfil local y contraseña opcional',
        'Auditoría de operaciones empresariales',
        'Configuración de impuestos',
        'Respaldo y restauración',
        'Parámetros del sistema',
        'Integración con servicios externos',
      ]}
    />
  );
}
