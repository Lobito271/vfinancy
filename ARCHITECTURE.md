# Arquitectura del Frontend — vfinancy

> Decisiones arquitectónicas, capas, límites y reglas de dependencia.
> Toda pantalla, módulo o servicio nuevo debe respetar este documento.

## Vista General

```
┌────────────────────────────────────────────────────────────────┐
│  Browser / Wails WebView                                        │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │  React App (src/main.tsx)                                │  │
│  │   └─ App (src/app/App.tsx)                               │  │
│  │       └─ Providers (ErrorBoundary, QueryClient, Router)  │  │
│  │           └─ <Routes> lazy (Suspense)                    │  │
│  │               └─ AppLayout (Sidebar + Topbar + Breadcr.) │  │
│  │                   └─ <Outlet> → Page / Feature / Widget   │  │
│  └──────────────────────────────────────────────────────────┘  │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │  State                                                    │  │
│  │   • Zustand stores (session, theme, sidebar, ui, notif.) │  │
│  │   • TanStack Query cache (server state)                  │  │
│  │   • React Hook Form (form state + validation)            │  │
│  └──────────────────────────────────────────────────────────┘  │
│  ┌──────────────────────────────────────────────────────────┐  │
│  │  Backend bridge (Phase 1+)                               │  │
│  │   • Wails bindings (auto-generated in src/wailsjs/)      │  │
│  │   • Domain services call bindings (services/<domain>/)  │  │
│  └──────────────────────────────────────────────────────────┘  │
│  Go backend (Wails + pgx + PostgreSQL)                         │
└────────────────────────────────────────────────────────────────┘
```

## Capas y Reglas de Dependencia

Las dependencias **solo pueden fluir hacia abajo** (las capas inferiores no conocen las superiores).

```
┌──────────────────────────────────────────────────────────────┐
│  pages/          ← pantallas de ruta (orquestan, no lógica) │
├──────────────────────────────────────────────────────────────┤
│  features/<x>/   ← módulos de negocio (customers, sales…)   │
│  ├── pages/      ←   screens específicos del feature        │
│  ├── components/ ←   UI específica del feature              │
│  ├── hooks/      ←   hooks de datos (tanstack query)        │
│  ├── services/   ←   consume services globales o reexporta   │
│  ├── schemas/    ←   zod schemas del feature                 │
│  ├── types/      ←   tipos específicos                       │
│  └── utils/      ←   helpers específicos                    │
├──────────────────────────────────────────────────────────────┤
│  components/     ← UI reutilizable cross-feature (por tipo)  │
│  ├── button/  input/  select/  checkbox/  table/  dialog/   │
│  ├── card/    badge/  navigation/  feedback/  charts/        │
│  ├── layout/  money/  date/  form/  auth/  misc/            │
├──────────────────────────────────────────────────────────────┤
│  services/       ← acceso a datos (Wails bindings en Phase 1)│
│  ├── api/        ←   cliente HTTP base (preparado, no usado)│
│  ├── auth/ customers/ suppliers/ products/ sales/ inventory/ │
│  ├── treasury/ accounting/ reports/ administration/ settings/│
│  ├── queryKeys.ts ← cache keys de TanStack Query            │
│  └── index.ts    ←   barrel                                  │
├──────────────────────────────────────────────────────────────┤
│  hooks/          ← hooks globales (usePermission, etc.)      │
├──────────────────────────────────────────────────────────────┤
│  design-system/  ← tokens de diseño (colors, spacing, etc.)  │
│  constants/      ← constantes de dominio (rutas, permisos…)  │
│  utils/          ← helpers puros (format, validators…)       │
│  stores/         ← Zustand stores                            │
│  data/           ← mock data determinístico                  │
│  locales/        ← i18n                                      │
├──────────────────────────────────────────────────────────────┤
│  app/            ← entrypoint, providers, error boundary     │
│  layouts/        ← chrome de la app (AppLayout, etc.)        │
└──────────────────────────────────────────────────────────────┘
```

### Reglas

1. **Pages no contienen lógica de negocio.** Llaman hooks/services y pasan datos a componentes.
2. **Componentes en `components/` no importan de `features/`.** Son agnósticos al dominio.
3. **Services no importan React ni componentes.** Son funciones puras (idealmente) que devuelven promesas o datos.
4. **Stores no importan services directamente.** Coordinan UI; los datos viven en TanStack Query.
5. **`features/<x>/` puede importar de `components/`, `services/`, `hooks/`, `design-system/`, `constants/`, `utils/`, `stores/`, otros `features/`.** No al revés.
6. **`utils/` no importa de `components/`, `features/`, `services/`, `stores/`.** Es la capa más baja.
7. **`design-system/` y `constants/` no importan nada de la app.** Son hojas.

## Tipos de Organización: Categoría vs. Feature

El proyecto usa **dos patrones complementarios**:

| Patrón             | Carpeta                | Contiene                                                | Ejemplo                |
|--------------------|------------------------|---------------------------------------------------------|------------------------|
| Por categoría (UI) | `components/<kind>/`   | Primitivas reutilizables cross-feature                  | `<Button>`, `<Card>`   |
| Por feature (dominio) | `features/<name>/`   | Todo lo específico de un módulo de negocio              | `customers/`, `sales/` |

**Regla de oro:** Si el componente se usa en ≥ 2 features, va a `components/`. Si solo se usa en un feature, va a `features/<x>/components/`.

## Estado: tres almacenes, una fuente de verdad

| Tipo de estado    | Herramienta             | Ejemplo                                         |
|-------------------|-------------------------|--------------------------------------------------|
| **UI / session**  | Zustand (con `persist`) | `useThemeStore`, `useSessionStore`, `useUIStore`|
| **Server state**  | TanStack Query          | `useCustomers()`, `useProduct(id)`               |
| **Form state**    | React Hook Form + Zod   | `<Form onSubmit>` con `zodResolver`              |

**Reglas:**
- Datos que vienen del backend → TanStack Query. **Nunca** duplicarlos en Zustand.
- Configuración del usuario (tema, idioma, sidebar colapsado) → Zustand con `persist`.
- Estado de formulario → React Hook Form. **No** usar `useState` para cada campo.
- Toast queue → Zustand `useNotificationStore` (sin persist).

## Sistema de Permisos (5 niveles)

```
Módulo  →  Página  →  Acción  →  Campo  →  Botón/Menú
```

Las claves de permiso siguen el formato `<module>.<verb>` (ver `src/constants/permissions.ts`):

```ts
Permissions.Customers.View      // 'customers.view'
Permissions.Customers.Delete    // 'customers.delete'
Permissions.Sales.Approve       // 'sales.approve'
```

Roles predefinidos (`admin`, `manager`, `accountant`, `seller`, `warehouse`, `viewer`) reciben un conjunto de permisos vía `RolePermissions`.

### API

```tsx
// Hook
const canDelete = usePermission(Permissions.Customers.Delete);

// Componente
<Can permission={Permissions.Customers.Delete} fallback={<DisabledButton />}>
  <Button variant="destructive">Eliminar</Button>
</Can>

// Hard check
if (hasPermission(ctx, Permissions.Sales.Approve)) { ... }
```

### Comportamiento por defecto

- **Sidebar:** oculta items si el usuario no tiene `*.view`.
- **Botones:** se deshabilitan (no se ocultan) cuando el usuario no tiene el permiso específico, salvo en acciones destructivas donde se muestra el botón pero el handler aborta.
- **Formularios:** el `Field` soporta `permission` opcional para ocultar inputs a nivel campo.
- **Rutas:** la protección de rutas (redirect a `/login`) se hace en el `<Route>` o en un layout `ProtectedRoute` (pendiente de fase 2).

## Ruteo

- **Públicas:** `/login`
- **Protegidas:** todas las demás (ver `Routes` en `constants/routes.ts`)
- **404:** `*` → redirect a `/`
- **Lazy loading:** cada page se importa con `React.lazy()` y se envuelve en `<Suspense>`. El fallback es `<Spinner />` por defecto.

## Formularios

Stack obligatorio:
- `react-hook-form` para estado
- `zod` para schema
- `@hookform/resolvers/zod` para integrar (instalar cuando se agregue)
- `components/form/*` para UI (TextField, MoneyField, SelectField, etc.)

```tsx
const schema = z.object({
  name: z.string().min(1),
  email: z.string().email(),
  amount: z.number().positive(),
});

<Form<z.infer<typeof schema>>
  schema={schema}
  defaultValues={{ name: '', email: '', amount: 0 }}
  onSubmit={async (v) => customersService.create(v)}
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
```

> Nota: el `Form` actual no inyecta `zodResolver` automáticamente. En la primera fase con formularios reales, agregar `resolver: zodResolver(schema)` en `useForm` del componente `Form`.

## Tablas (DataTable Framework)

`<DataTable>` en `components/table/DataTable.tsx` es **el** componente de tabla de la app. Cualquier tabla nueva debe usarlo.

Capacidades:
- Sorting por columna (`sortable: true`)
- Paginación (10/25/50/100)
- Búsqueda global
- Visibilidad de columnas (con persistencia en localStorage vía `preferencesKey`)
- Selección múltiple + bulk actions
- Sticky header
- Sticky first column
- Estados: loading (skeleton) / empty / error
- Exportación a CSV
- Acciones por fila en columna fija
- Toolbar inyectable (`toolbarLeft`, `toolbarRight`)
- Responsive (overflow horizontal)

Pendientes arquitectónicos (no implementados aún):
- Reordering de columnas (drag & drop)
- Resizing de columnas
- Filtros por columna (popover)
- Saved filter presets
- Virtual scrolling (preparado con `virtualized` prop stub, requiere `@tanstack/react-virtual`)
- Exportación a Excel
- Keyboard navigation (flechas arriba/abajo)

## Dashboard (Widget System)

El dashboard no es un componente monolítico. Es un grid que compone **widgets** independientes:

- `src/features/dashboard/DashboardPage.tsx` define el layout (`DashboardGrid` con items).
- Cada widget es un componente autocontenido (e.g. `MonthSalesWidget`).
- `WidgetShell` provee header/loading/error común.
- Futuros widgets se agregan a `widgets/index.tsx` y se incluyen en el layout.

Pluggable: el layout se puede serializar (futuro) para drag & drop.

## i18n

- **Fuente única:** `src/locales/es.ts` (es-PE).
- **Helper:** `t('path.to.key')` y `t('key', { n: 5 })` (con variables `{n}`).
- **Prohibido** hardcodear strings en componentes.
- Excepción: notificaciones mock en Topbar (estos vienen de un módulo futuro).

## Temas

- `useThemeStore` (Zustand + `persist`) con 3 valores: `light` | `dark` | `system`.
- Aplicado al `<html>` antes del primer render (ver `src/app/Providers.tsx` y `stores/theme.ts`).
- Clase `dark` en `<html>`; Tailwind con `darkMode: ['class']`.
- Tokens de color en CSS variables (`src/index.css`) **alineados** con los tokens de `src/design-system/colors.ts`. Si cambias uno, cambia el otro.

## Servicios (capa de datos)

Cada dominio (`customers`, `suppliers`, `products`, `sales`, `inventory`, `treasury`, `accounting`, `reports`, `administration`, `settings`, `auth`) expone:

1. **Funciones CRUD** síncronas (Phase 0.5 mock) o asíncronas (Phase 1+: Wails bindings).
2. **Query keys** centralizadas en `services/queryKeys.ts`.
3. **Hooks** de TanStack Query en `features/<x>/hooks/use<X>.ts` que envuelven los servicios y exponen `useQuery` / `useMutation`.

Patrón actual (mock):

```ts
// services/customers/index.ts
export const customersService = {
  async list(q: CustomerQuery) { await sleep(150); return { items, total }; },
  async get(id: string) { ... },
  async create(input: CustomerCreateInput) { ... },
  // ...
};
```

Reemplazo futuro (Phase 1, con Wails):

```ts
export const customersService = {
  async list(q: CustomerQuery) { return wails.customers.List(q); },
  // ...
};
```

Las **firmas no cambian**, así los hooks y componentes no necesitan refactor.

## Performance

- **Code splitting** por ruta con `React.lazy()`.
- **Suspense** con `Spinner` como fallback (no skeletons aún).
- **Memoización** selectiva: `React.memo` para widgets del dashboard, `useMemo` para filtros pesados.
- **Virtual scrolling** stub: `virtualized: true` prop. Instalar `@tanstack/react-virtual` cuando se necesite.
- **Persistencia de estado UI** en localStorage (no en cada render).

## Testing

- **Type check:** `npm run check` (tsc --noEmit) — corre en CI.
- **Build:** `npm run build` — corre en CI.
- **Unit tests:** pendiente (Vitest no instalado aún).
- **Manual:** `wails dev` para el flujo end-to-end.

## Convenciones de Código

- **No comentarios** salvo que el usuario lo pida explícitamente.
- **TypeScript strict mode** activo.
- **No `any`**, **no `@ts-ignore`**. Si no se puede evitar, usar `unknown` y narrowing.
- **Naming:** componentes PascalCase, hooks `useX`, utils camelCase, constantes UPPER_SNAKE_CASE en `constants/`, kebab-case en archivos.
- **Imports:** `import type { X }` para tipos. `@/foo` para todo lo interno.
- **Sin emojis en UI.**
- **Textos en español (es-PE).** Código en inglés.

## Próximas Fases (impacto arquitectónico)

| Fase | Impacto                                                              |
|------|----------------------------------------------------------------------|
| 1    | DB schema + migraciones + Wails bindings reemplazan mocks en services|
| 2    | Auth real, login, sesión persistente, `ProtectedRoute`               |
| 3    | CRUD real de customers/suppliers/products                             |
| 5    | Sales con transacciones → requiere `database.WithTx` en services     |
| 9    | Drag & drop de widgets + persistencia del layout                     |

---

## 14. Domain Layer (Phase 1.2)

`backend/internal/domain/` is the heart of the ERP. It contains every business entity, value object, enum, and business rule. It **must not** import PostgreSQL, Wails, React, or any other infrastructure — the domain is the most stable, most-tested, and most-isolated layer.

### 14.1 Structure

```
domain/
  errors/        typed domain errors with stable codes (INSUFFICIENT_STOCK, etc.)
  enums/         strongly-typed enumerations (SaleStatus, CustomerStatus, ...)
  valueobjects/  Money, Percentage, Quantity, SKU, Barcode, Email, Phone,
                 DocumentNumber, Address, ExchangeRate, CurrencyCode, ...
  validation/    low-level validators (RequiredString, InRange, etc.)
  events/        domain event types (scaffold; event bus lives in app layer)
  services/      stateless cross-entity services (TaxCalculator, ProfitCalculator)
  entities/      domain entities, grouped by bounded context:
    identity/     Company, Branch, User, Role, Permission
    masterdata/   Customer, Supplier, Product, ProductCategory, ProductBrand,
                  UnitOfMeasure, Warehouse, Currency, Country, Tax
    inventory/    InventoryBatch, InventoryMovement (25-day clearance rule)
    purchasing/   PurchaseOrder, PurchaseOrderItem, SupplierPayment
    sales/        Sale, SaleItem, CustomerPayment, CustomerAdvance
    treasury/     BankAccount, CreditCard
    accounting/   ChartOfAccount, JournalEntry, JournalEntryLine
```

### 14.2 Money

`valueobjects.Money` wraps `shopspring/decimal` with **two-decimal precision** (mirroring PostgreSQL `NUMERIC(18,2)`). It supports `+`, `-`, `*int`, `*percentage`, comparison, and string serialization. **No `float64` anywhere in the financial path.**

### 14.3 Entity Conventions

Every entity:

- Exposes a `New<Name>(...Options)` constructor that returns `(*Entity, error)`.
- Validates inputs at construction. No setters on private fields.
- Exposes meaningful behavior methods (`Activate`, `Cancel`, `ApplyPayment`, `Recalculate`, `Post`, `Reverse`).
- Owns its state transitions. State changes are gated by methods that return errors on invalid transitions.
- Carries `CreatedAt`, `UpdatedAt`, `DeletedAt`, `CreatedBy`, `UpdatedBy` for audit (matching the SQL schema in `DATABASE_SCHEMA.md`).
- Is reconstructable from a database row by an unexported / persistence-layer factory. The persistence layer (Phase 2) is responsible for that.

### 14.4 The 25-Day Clearance Rule

`InventoryBatch.MaximumSaleDate()` returns `ArrivalDate + 25 days`. `IsClearance(today)` returns true when the batch is past that date and still has quantity. `NeedsClearanceSoon(today)` returns true when the batch has 3 or fewer days left and still has quantity. These methods are pure: they do not mutate the batch.

### 14.5 Domain Errors

`derrors` (the `errors` package) defines:

- **Sentinel errors** with stable codes (`ErrInsufficientStock`, `ErrSaleAlreadyPaid`, `ErrUnbalancedJournalEntry`, ...) usable with `errors.IsCode(err, "INSUFFICIENT_STOCK")`.
- A `DomainError` interface that the application layer maps to HTTP status codes / UI messages.
- `New`, `Wrap`, `IsCode`, `IsAnyCode` helpers for context-tagging.

### 14.6 Aggregate Roots

The `Sale`, `PurchaseOrder`, `JournalEntry` types are aggregate roots. They own their child collections (`Items`, `Lines`), enforce invariants (no duplicate products, balanced debits/credits, payment ≤ outstanding), and expose state transitions. The application layer (Phase 2) and repository (Phase 2) treat these as the unit of consistency.

### 14.7 What's NOT in the Domain

- No SQL, no `pgx`, no `database/sql`.
- No HTTP, no Wails bindings, no JSON marshaling (except for `Money.MarshalJSON` for legacy JSON consumers).
- No clocks — `time.Time` is passed in (now := time.Now().UTC()).
- No I/O, no goroutines, no randomness.

The domain is deterministic and pure (modulo the `time.Time` inputs it accepts).

### 14.8 Tests

141 unit tests, all passing, with the following distribution:

- `valueobjects/` — Money, Quantity, Percentage, Email, Phone, SKU, Barcode, DocumentNumber, ExchangeRate, CurrencyCode, Address, FullName, ShortCode, ChartOfAccountsCode
- `entities/identity/` — Company, Branch, User (lockout, status, role assignment), Role (system vs custom)
- `entities/masterdata/` — Customer (status, credit limit, debt, payment), Supplier, Product (margin, cost, price, stock limits)
- `entities/inventory/` — InventoryBatch (25-day clearance, consume, receive, write-off), InventoryMovement (sign validation, total cost)
- `entities/sales/` — Sale (line totals, profit, payments, state transitions, duplicates), SaleItem (subtotal, total, profit)
- `entities/purchasing/` — PurchaseOrder (totals, payments, approve, cancel, reconcile), SupplierPayment (allocations)
- `entities/treasury/` — BankAccount, CreditCard (charge, pay, available credit, validations)
- `entities/accounting/` — JournalEntry (add line, remove line, balance, post, immutability, reversal), ChartOfAccount (normal balance)
- `services/` — TaxCalculator (exclusive, inclusive, with discount), ProfitCalculator

Run with: `go test ./backend/internal/domain/...`

---

## 15. Repository Layer (Phase 1.3)

`backend/internal/domain/repositories/` declares the persistence abstraction. `backend/internal/infrastructure/persistence/postgres/` provides the PostgreSQL implementation.

### 15.1 Layer rules

- The domain defines interfaces; the infrastructure implements them. The application layer depends on the domain interfaces only.
- Repositories never contain business logic. They translate between the domain's typed entities and the database's rows. A repository method like `GetOutstandingBalance(ctx, id) (string, error)` is a query, not a calculation.
- Repositories never update inventory automatically, never compute taxes, never post journal entries. Those are application-layer use cases.

### 15.2 Common abstractions

`internal/domain/repositories/`:
- `Page[T]`, `PageRequest`, `Sort`, `TimeRange` — pagination and filtering primitives.
- `TransactionManager` — `WithinTransaction(ctx, fn)` runs `fn` inside a transaction and rolls back on error.
- `UnitOfWork` — one method per aggregate; each call returns a repository bound to the current transaction. Inside a transaction, the unit-of-work is stored in the `context.Context` via `ContextWithUnitOfWork` / `UnitOfWorkFromContext`.
- `ErrNotFound`, `ErrDuplicate`, `ErrForeignKey`, `ErrCheckConstraint`, `ErrConnection`, `ErrTx` — sentinel errors with stable codes. The application layer matches by code via `derrors.IsCode(err, "DUPLICATE")`.

### 15.3 Error mapping

`internal/infrastructure/persistence/postgres/errors.go` translates `*pgconn.PgError` to the domain-friendly sentinels:
- `23505 unique_violation` → `ErrDuplicate`
- `23503 foreign_key_violation` → `ErrForeignKey`
- `23514 check_violation` / `23502 not_null_violation` → `ErrCheckConstraint`
- `40001` / `40P01` → `ErrTx`

`sql.ErrNoRows` is translated to `ErrNotFound`. Raw SQL errors never reach the upper layers.

### 15.4 Implementation status

- **CustomerRepository** — fully implemented in `customer_repository.go` with 8 integration tests (`customer_repository_test.go`). Covers Create, Update, Delete (soft), GetByID, GetByDocument, List (paginated + filtered + searched), GetOutstandingBalance, and the transaction rollback path.
- **19 other repositories** — stubs in `stubs.go` that satisfy the interface and return a watermark error if any unstubbed method is called. Each is replaced by a real implementation in its own file as the corresponding migration lands.
- **No business logic** anywhere in the persistence layer.

### 15.5 Testing pattern

The `postgres` package's test setup (`setup_test.go`):
- starts an embedded PostgreSQL 16 process (downloaded from Maven Central).
- creates the test database and applies all 12 Module 1 migrations.
- creates the Module 2 tables (`customers`, etc.) on demand so the test runs even before those migrations land.
- exposes `getDB(t)` to every test.
- truncates the customers table between tests for isolation.

Tests use real SQL through the pgx driver — no mocking, no in-memory adapter. The test pattern is reusable for every future repository.

---

## 16. Service Layer (Phase 1.4)

`backend/internal/application/services/` is the **business service layer**. It is the only place in the system that is allowed to make business decisions. Repositories are pure persistence; the application use case layer (Phase 1.5) will compose services across transactions; the UI is a thin shell over use cases.

The full reference lives in `backend/SERVICE_LAYER.md`. Key points:

- One service package per bounded context (`customer`, `supplier`, `product`, `inventory`, `purchasing`, `sales`, `customerpayments`, `treasury`, `accounting`, `reporting`).
- Every method is a verb (`Create`, `Update`, `Cancel`, `Post`, `Reverse`, `Approve`, etc.) backed by an `XxxInput` struct.
- Multi-step work runs inside `txm.WithinTransaction(ctx, fn)` with the UoW stamped on the context. The UoW hands out transaction-bound repositories.
- No cross-service calls inside a single service — the use case layer composes them.
- Sentinel errors with stable codes (`IsCode` matches) live in `services/errors.go`; the service wraps them around the domain errors.
- Structured logging for every business outcome; no PII.

The full layout, every method per service, anti-patterns, and the test pattern are in `backend/SERVICE_LAYER.md`.
