# Estructura de Carpetas

> Qué va en cada carpeta y qué no.
> Si dudas entre dos ubicaciones, mira la sección "Regla de oro" abajo.

```
src/
├── app/                # Entrypoint, providers, error boundary
│   ├── App.tsx         #   <App>: routes + lazy loading
│   ├── Providers.tsx   #   QueryClient, Theme, Router, Tooltip, ErrorBoundary
│   └── ErrorBoundary.tsx
│
├── layouts/            # Chrome de la app (componentes que envuelven pages)
│   └── AppLayout.tsx
│
├── pages/              # Pantallas top-level (lazy, una por ruta)
│   ├── LoginPage.tsx
│   ├── CustomersPage.tsx
│   ├── SuppliersPage.tsx
│   ├── ProductsPage.tsx
│   ├── InventoryPage.tsx
│   ├── PurchasesPage.tsx
│   ├── SalesPage.tsx
│   ├── TreasuryPage.tsx
│   ├── AccountingPage.tsx
│   ├── ReportsPage.tsx
│   ├── SettingsPage.tsx
│   ├── AdministrationPage.tsx
│   ├── ModulePage.tsx  #   wrapper reutilizable para placeholders
│   └── PlaceholderPage.tsx
│
├── features/           # Módulos de negocio (opcional, por feature)
│   ├── dashboard/
│   │   ├── DashboardPage.tsx
│   │   ├── DashboardGrid.tsx
│   │   └── widgets/
│   │       ├── index.tsx          #   registry de widgets
│   │       └── WidgetShell.tsx
│   ├── customers/
│   │   ├── hooks/useCustomers.ts
│   │   ├── schemas/customer.ts    #   zod
│   │   ├── types/index.ts
│   │   └── index.ts               #   barrel
│   └── sales/                     #   scaffold (vacío, se llena en Phase 5)
│
├── components/         # Primitivas reutilizables (por categoría)
│   ├── auth/          #   Can, PermissionGate
│   ├── badge/         #   Badge, SaleStatusBadge, CustomerStatusBadge, StockBadge
│   ├── button/        #   Button (7 variants, 5 sizes)
│   ├── card/          #   Card, CardHeader/Title/Description/Content/Footer, StatCard
│   ├── charts/        #   LineChart, BarChart, PieChart (recharts wrappers)
│   ├── checkbox/      #   Checkbox, RadioGroup, RadioGroupItem, Switch
│   ├── dialog/        #   Dialog, AlertDialog (5 variants), ConfirmDialog
│   ├── feedback/      #   Spinner, Skeleton, ProgressBar, EmptyState, ErrorState, Toaster
│   ├── form/          #   Form + 16+ field components
│   ├── input/         #   Input, Textarea, Field, Label, SearchInput
│   ├── layout/        #   PageContainer, PageHeader, Section, Stack, Grid
│   ├── misc/          #   DropdownMenu, Separator, Tooltip
│   ├── money/         #   MoneyInput, MoneyDisplay
│   ├── navigation/    #   Sidebar, Topbar, Breadcrumbs
│   ├── select/        #   Select (Radix) + SelectContent/Item/Label/Separator
│   └── table/         #   DataTable (framework), TablePagination
│
├── services/           # Acceso a datos (mock ahora, Wails en Phase 1)
│   ├── api/           #   cliente HTTP base (preparado, no usado)
│   ├── auth/          #   login, logout, me
│   ├── customers/     #   CRUD de clientes
│   ├── suppliers/     #   CRUD de proveedores
│   ├── products/      #   CRUD de productos
│   ├── sales/         #   CRUD de ventas
│   ├── inventory/     #   inventario + movimientos
│   ├── treasury/      #   cuentas bancarias + transacciones
│   ├── accounting/    #   plan de cuentas + libro diario
│   ├── reports/       #   generación de reportes
│   ├── administration/    #   usuarios, audit log
│   ├── settings/      #   empresa, preferencias
│   ├── queryKeys.ts   #   cache keys de TanStack Query
│   └── index.ts       #   barrel
│
├── hooks/              # Hooks globales
│   └── usePermission.ts    #   usePermission, useRole, usePermissionContext
│
├── stores/             # Zustand stores
│   ├── theme.ts        #   tema + persist
│   ├── session.ts      #   usuario logueado + persist
│   ├── sidebar.ts      #   colapsado + persist
│   ├── ui.ts           #   global search
│   └── notification.ts #   toast queue
│
├── design-system/      # Design tokens + icon registry
│   ├── colors.ts
│   ├── spacing.ts
│   ├── typography.ts
│   ├── radius.ts
│   ├── shadows.ts
│   ├── animations.ts
│   ├── breakpoints.ts
│   ├── zIndex.ts
│   ├── transitions.ts
│   ├── icons.ts        #   IconRegistry
│   └── index.ts        #   tokens = { ... }
│
├── constants/          # Constantes de dominio
│   ├── routes.ts       #   Routes, ProtectedRoutes, isProtectedRoute
│   ├── permissions.ts  #   Permissions, Roles, RolePermissions
│   ├── currencies.ts   #   Currencies, DefaultCurrency
│   ├── countries.ts    #   Countries, DocumentTypes
│   ├── languages.ts    #   Languages, DefaultLanguage
│   ├── status.ts       #   SaleStatus, CustomerStatus, etc.
│   ├── taxes.ts        #   Taxes (IGV, ISR, etc.) + calculateTax
│   └── index.ts
│
├── utils/              # Helpers puros
│   ├── cn.ts           #   cn() — clsx + tailwind-merge
│   ├── format.ts       #   formatCurrency, formatDate, formatPercent, …
│   ├── validators.ts   #   isEmail, isPhonePE, isDNI, isRUC, …
│   ├── permissions.ts  #   buildContext, hasPermission, hasRole
│   ├── debounce.ts     #   useDebounce, debounce, throttle
│   ├── clipboard.ts    #   copyToClipboard, readFromClipboard
│   ├── download.ts     #   downloadCSV, downloadJSON, downloadText
│   ├── collection.ts   #   unique, groupBy, sum, sortBy, chunk
│   ├── misc.ts         #   generateId, isString, isNumber, isObject, getInitials
│   ├── storage.ts      #   readJSON, writeJSON, persistJSON
│   └── index.ts
│
├── data/               # Mock data determinístico
│   └── mock.ts
│
├── locales/            # i18n (es-PE)
│   ├── es.ts
│   └── index.ts        #   t() helper
│
├── lib/                # Helpers cross-cutting (nav config, etc.)
│   └── nav.ts          #   navRoutes, findRouteLabel
│
├── assets/             # Recursos estáticos
│
├── App.tsx             # Re-exporta app/App.tsx (compat)
├── main.tsx            # createRoot + StrictMode
├── index.css           # Tailwind + tokens CSS + Inter
└── vite-env.d.ts
```

## Regla de oro

| Si el código…                                       | Va en…                          |
|-----------------------------------------------------|----------------------------------|
| Define tokens, íconos, constantes de dominio        | `design-system/`, `constants/`  |
| Es un helper puro (sin React)                       | `utils/`                         |
| Es un store global                                  | `stores/`                        |
| Es un componente UI reutilizable cross-feature      | `components/<category>/`         |
| Es un componente específico de un módulo            | `features/<x>/components/`       |
| Hace CRUD de un dominio (mock o Wails)              | `services/<domain>/`             |
| Es una página top-level                             | `pages/`                         |
| Es una página específica de un feature              | `features/<x>/pages/`            |
| Es la entrada de la app o un provider raíz          | `app/`                           |
| Es un hook global de permisos, etc.                 | `hooks/`                         |
| Es el chrome (sidebar, layout)                      | `layouts/`                       |

## Carpetas vacías (reservadas)

- `src/hooks/` — solo tiene `usePermission.ts` por ahora. Se irá llenando.
- `src/services/<x>/` — todos los servicios existen pero algunos solo con stubs.
- `src/features/<x>/` — solo `dashboard` y `customers` tienen contenido significativo. `sales` es un placeholder.

## Anti-patterns

- ❌ Crear `components/customers/CustomerCard.tsx` → debe ir en `features/customers/components/`.
- ❌ `import { something } from '@/components'` (genérico) → importar de la subcarpeta: `@/components/button`.
- ❌ Hardcodear strings en español en componentes → `t('key')` o en `constants/`.
- ❌ Duplicar tokens en `tailwind.config.js` y en `design-system/` sin documentar.
- ❌ Meter `useState` para datos del servidor → TanStack Query.
- ❌ `console.log` en código de producción (solo en desarrollo).
- ❌ Añadir un campo numérico con `parseFloat` / `toFixed` → `valueAsNumber` en RHF + `formatCurrency` para mostrar.
