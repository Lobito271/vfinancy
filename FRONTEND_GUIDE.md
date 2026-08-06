# Guía del Frontend

> Cómo agregar pantallas, componentes, features y servicios siguiendo las convenciones del proyecto.
> Si una práctica no está documentada aquí, mira `ARCHITECTURE.md` y `FOLDER_STRUCTURE.md`.

## 1. Setup local

```bash
# 1. Instalar dependencias del frontend
cd frontend
npm install

# 2. Verificar tipos
npm run check

# 3. Build de producción (lo que Wails embebe)
npm run build

# 4. (Wails) live dev
# Desde la raíz, con Wails CLI instalado
wails dev
```

Node 18+. Wails CLI, Go y PostgreSQL se necesitan solo para el flujo end-to-end.

## 2. Agregar un nuevo módulo de menú

### 2.1 Crear la ruta en `constants/routes.ts`

```ts
export const Routes = {
  // ...
  NewModule: '/nuevo-modulo',
};
```

### 2.2 Crear el permiso en `constants/permissions.ts`

```ts
export const Permissions = {
  // ...
  NewModule: {
    View: 'newmodule.view',
    Create: 'newmodule.create',
    Edit: 'newmodule.edit',
    Delete: 'newmodule.delete',
  },
};
```

Y agrégalo al array de permisos del rol apropiado (e.g. `manager`) en `RolePermissions`.

### 2.3 Agregar entrada en `lib/nav.ts`

```ts
export const navRoutes: NavRoute[] = [
  // ...
  { to: Routes.NewModule, label: 'Nuevo Módulo', icon: Icons.Navigation.NewModule, permission: Permissions.NewModule.View },
];
```

### 2.4 Crear la page en `pages/`

```tsx
// pages/NewModulePage.tsx
import { ModulePage } from './ModulePage';
import { Icons } from '@/design-system/icons';

export function NewModulePage() {
  return (
    <ModulePage
      title="Nuevo Módulo"
      subtitle="Descripción corta"
      icon={Icons.Navigation.NewModule}
      description="..."
      phase="Fase N"
      features={['Feature 1', 'Feature 2']}
    />
  );
}
```

### 2.5 Registrar la ruta en `app/App.tsx`

```tsx
const NewModulePage = lazy(() => import('@/pages/NewModulePage').then((m) => ({ default: m.NewModulePage })));
// ...
<Route path="nuevo-modulo" element={<NewModulePage />} />
```

## 3. Crear un feature con CRUD

Estructura recomendada:

```
src/features/<feature>/
├── components/     # UI específica
├── hooks/          # use<Feature> (tanstack query)
├── services/       # re-exporta o contiene lógica de negocio
├── schemas/        # zod
├── types/          # tipos
├── utils/          # helpers
└── index.ts        # barrel
```

### 3.1 Schema (zod)

```ts
// features/customers/schemas/customer.ts
import { z } from 'zod';

export const CustomerSchema = z.object({
  id: z.string(),
  documentNumber: z.string().min(8, 'Mínimo 8 caracteres'),
  businessName: z.string().min(1, 'Requerido'),
  email: z.string().email().optional().or(z.literal('')),
  // ...
});

export type CustomerFormValues = z.infer<typeof CustomerSchema>;
```

### 3.2 Hooks (TanStack Query)

```ts
// features/customers/hooks/useCustomers.ts
import { useQuery } from '@tanstack/react-query';
import { customersService } from '@/services/customers';
import { queryKeys } from '@/services/queryKeys';

export function useCustomers(query: CustomerQuery = {}) {
  return useQuery({
    queryKey: queryKeys.customers.list(query),
    queryFn: () => customersService.list(query),
  });
}
```

### 3.3 Página

```tsx
// pages/CustomersPage.tsx
import { PageContainer, PageHeader } from '@/components/layout';
import { useCustomers } from '@/features/customers/hooks/useCustomers';
import { CustomersTable } from '@/features/customers/components/CustomersTable';

export function CustomersPage() {
  const { data, isLoading, error, refetch } = useCustomers();
  return (
    <PageContainer>
      <PageHeader title="Clientes" subtitle="..." />
      <CustomersTable items={data?.items ?? []} loading={isLoading} error={error} onRetry={refetch} />
    </PageContainer>
  );
}
```

### 3.4 Tabla

```tsx
// features/customers/components/CustomersTable.tsx
import { DataTable, type Column } from '@/components/table';
import { CustomerStatusBadge } from '@/components/badge';
import type { Customer } from '@/data/mock';

interface Props {
  items: Customer[];
  loading?: boolean;
  error?: Error | null;
  onRetry?: () => void;
}

export function CustomersTable({ items, loading, error, onRetry }: Props) {
  const columns: Column<Customer>[] = [
    { id: 'documentNumber', header: 'Documento', cell: (r) => r.documentNumber, sortable: true, exportable: true },
    { id: 'businessName', header: 'Razón social', cell: (r) => r.businessName, sortable: true, exportable: true },
    { id: 'currentDebt', header: 'Deuda', align: 'right', cell: (r) => formatCurrency(r.currentDebt), sortable: true },
    { id: 'status', header: 'Estado', cell: (r) => <CustomerStatusBadge status={r.status} /> },
  ];
  return <DataTable columns={columns} data={items} keyField="id" loading={loading} error={error} onRetry={onRetry} preferencesKey="customers" />;
}
```

## 4. Crear un formulario

```tsx
import { z } from 'zod';
import { Form, TextField, EmailField, MoneyField, Button } from '@/components/form';

const schema = z.object({
  name: z.string().min(1, 'Requerido'),
  email: z.string().email('Correo inválido'),
  amount: z.number().positive('Debe ser positivo'),
});

type Values = z.infer<typeof schema>;

export function MyForm({ onSubmit }: { onSubmit: (v: Values) => Promise<void> }) {
  return (
    <Form<Values>
      defaultValues={{ name: '', email: '', amount: 0 }}
      onSubmit={onSubmit}
    >
      {() => (
        <>
          <TextField name="name" label="Nombre" required />
          <EmailField name="email" label="Correo" required />
          <MoneyField name="amount" label="Monto" currency="PEN" required />
          <Button type="submit">Guardar</Button>
        </>
      )}
    </Form>
  );
}
```

> El `Form` actual no aplica `zodResolver` automáticamente. Para validar, agregar en `Form.tsx`:
> ```ts
> const form = useForm<T>({ ...props, resolver: zodResolver(schema) });
> ```
> (Requiere `@hookform/resolvers` como dep.)

## 5. Permisos en componentes

```tsx
import { Can } from '@/components/auth';
import { Permissions } from '@/constants/permissions';
import { usePermission } from '@/hooks/usePermission';

// Mostrar solo si tiene permiso
<Can permission={Permissions.Customers.Delete} fallback={null}>
  <Button variant="destructive">Eliminar</Button>
</Can>

// Lógica condicional
function MyAction() {
  const canExport = usePermission(Permissions.Customers.Export);
  return canExport ? <ExportButton /> : <DisabledExportButton />;
}
```

## 6. Toasts / Notificaciones

```tsx
import { useNotificationStore } from '@/stores/notification';

function MyComponent() {
  const push = useNotificationStore((s) => s.push);
  return (
    <Button
      onClick={() => {
        try {
          // ...acción
          push({ title: 'Guardado', description: 'Cambios aplicados', variant: 'success' });
        } catch (e) {
          push({ title: 'Error', description: 'No se pudo guardar', variant: 'destructive' });
        }
      }}
    >
      Guardar
    </Button>
  );
}
```

## 7. i18n

```tsx
import { t } from '@/locales';

// Simple
<span>{t('common.save')}</span>

// Con variables
<span>{t('validation.minLength', { n: 8 })}</span>  // "Mínimo 8 caracteres"
```

Para agregar una clave nueva, edita `src/locales/es.ts`. El tipado de `t()` se actualiza automáticamente.

## 8. Iconos

```tsx
import { Icons } from '@/design-system/icons';

<Icons.Navigation.Customers />
<Icons.Action.Delete />
<Icons.Status.Warning />
```

No importes íconos de `lucide-react` directamente. Si necesitas uno nuevo, agrégalo al registry en `design-system/icons.ts`.

## 9. Tema

El tema es automático. Si necesitas forzar lectura del tema actual (e.g. en un componente que genera SVG):

```ts
import { useThemeStore } from '@/stores/theme';

const theme = useThemeStore((s) => s.theme);
const resolved = useThemeStore((s) => s.resolved);
```

## 10. Checklist antes de hacer PR

- [ ] `npm run check` pasa.
- [ ] `npm run build` pasa.
- [ ] No hay strings hardcodeados en español.
- [ ] No hay `any` ni `@ts-ignore`.
- [ ] Componentes nuevos siguen el patrón `category/` y tienen `index.ts` barrel.
- [ ] Forms usan `Form` + zod + componentes de `components/form`.
- [ ] Tablas usan `<DataTable>` (no `<table>` a mano).
- [ ] Permisos declarados en `Permissions` y aplicados con `<Can>` o `usePermission`.
- [ ] Si agregaste una página, la registraste en `nav.ts` con su permiso.
- [ ] Si agregaste un servicio, usaste `queryKeys` para cache.
