import { Shield } from 'lucide-react';
import { ModulePage } from './ModulePage';

export function AdministrationPage() {
  return (
    <ModulePage
      title="Administración"
      subtitle="Configuración avanzada del sistema"
      icon={Shield}
      description="Administre usuarios, roles, permisos y configuración avanzada del sistema."
      phase="Fase 2"
      features={[
        'Gestión de usuarios (crear, editar, desactivar)',
        'Roles y permisos (RBAC)',
        'Auditoría de operaciones (login, logout, CRUD)',
        'Configuración de impuestos',
        'Respaldo y restauración',
        'Parámetros del sistema',
        'Integración con servicios externos',
      ]}
    />
  );
}
