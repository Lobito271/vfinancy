# AGENTS.md

## Project Overview

**vfinancy** is a custom Desktop ERP (Enterprise Resource Planning) system built with Wails v2 (Go + React + TypeScript) backed by PostgreSQL. It replaces manual business processes with a centralized platform for purchasing, sales, inventory, treasury, accounting, and financial reporting.

**Current state: Phase 1.4 (Business Service Layer) — done.** Phase 0 (init), Phase 0.5 (UI/UX), Phase 0.6 (architecture refinement), Phase 1 (database architecture), Phase 1.1 (PostgreSQL schema for Module 1), Phase 1.2 (Domain Layer), Phase 1.3 (Repository Layer) complete. Phase 1.4 produced the complete business service layer under `backend/internal/application/services/`: 8 service packages (customer, supplier, product, inventory, purchasing, sales, customerpayments, treasury, accounting, reporting), shared errors, a TxManager adapter, and a slog-based logger. **No business logic outside the service layer.** 178 unit tests pass across the project.

## Target Stack (mandatory)

- **Desktop framework:** Wails v2
- **Backend:** Go 1.23
- **Frontend:** TypeScript + **React 18** + Vite 5
- **State / data:** Zustand, TanStack Query, React Hook Form, Zod
- **Styling:** TailwindCSS 3 + `tailwindcss-animate` + Shadcn UI patterns
- **DB:** PostgreSQL via `github.com/jackc/pgx/v5/stdlib`

## Key Commands

```bash
# Full app (Wails) — must be run on a machine with Wails + Go toolchain installed
wails dev          # Live dev (Go + Vite, hot-reload)
wails build        # Production build → build/bin/

# Frontend-only (run in frontend/)
npm install        # install deps
npm run dev        # Vite dev server (no Go backend)
npm run build      # Type-check + production build → frontend/dist/
npm run check      # tsc --noEmit

# Backend CLI (run in repo root or backend/)
go run ./backend/cmd/cli migrate    # apply pending SQL migrations
go run ./backend/cmd/cli status     # show migration status

# Go tests
go test ./backend/...
```

In this dev container, only Node/npm are on PATH. Go, the `wails` CLI, and `psql` are not. Frontend changes can be verified with `npm run build` / `npm run check`. Go changes must be verified on a host with the full Wails toolchain.

## Architecture: Clean Architecture + DDD

```
Desktop Application (Wails)
├── React Frontend  →  Wails Bindings (auto-generated)
└── Go Backend
    ├── Domain (entities, value objects, repository interfaces)
    ├── Application (services, use cases)         [Phase 1+]
    ├── Infrastructure (DB, config, logger, migrations)
    ├── Repository (concrete implementations)     [Phase 1+]
    └── Interfaces/Bindings (Wails-exposed methods)
```

**Rule:** Frontend never accesses the database directly. All traffic flows React → Wails binding → Application service → Repository → PostgreSQL.

## Backend Layout (current)

```
backend/
  cmd/
    cli/                 # standalone CLI: `migrate`, `status`
  internal/
    domain/
      entities/          # Base (id + audit), future entities live here
      valueobjects/      # ID, Money
      repositories/      # generic Repository[T] interface
  infrastructure/        # NOT under internal/ on purpose (see "Internal packages" below)
    config/              # env-based config loader
    logger/              # structured slog logger (json | text, levels debug..error)
    database/            # *sql.DB wrapper + WithTx(ctx, fn) helper
    postgres/            # Connect(), EnsureDatabase(), DSN helpers
    migrations/          # file-based SQL migration runner
  interfaces/
    bindings/            # structs exposed to the React frontend via wails.Run
  migrations/            # SQL migration files (0001_xxx.up.sql / .down.sql)
  pkg/                   # reusable packages (reserved; empty in Phase 0)
```

The root `main.go` and `app.go` are the Wails entrypoint. `app.go` initializes config + logger, ensures the database exists, opens a connection, and runs pending migrations on startup. It binds two structs: the root `App` (config accessors for the frontend) and `bindings.App` (Wails-exposed methods).

### Internal packages

`backend/internal/...` packages can only be imported by packages whose import path is `vfinancy/backend/...` or a subpath. The root `vfinancy` package (where `main.go`/`app.go` live) cannot import them. So for Phase 0, `infrastructure/` and `interfaces/` live directly under `backend/` (not under `internal/`). When the root entrypoint is moved into `backend/cmd/server/` (a later phase), these can be folded back into `internal/`.

## Frontend Layout (current)

```
frontend/
  index.html
  package.json
  tsconfig.json + tsconfig.app.json + tsconfig.node.json
  vite.config.ts
  tailwind.config.js
  postcss.config.js
  src/
    main.tsx
    app/                     # App.tsx (routes + lazy), Providers.tsx, ErrorBoundary.tsx
    layouts/                 # AppLayout (sidebar + topbar + breadcrumbs)
    pages/                   # route screens (1 per module + Login)
    features/                # feature-based modules (dashboard/, customers/, sales/)
    components/              # category folders — see "Component organization" below
    services/                # one folder per business domain + queryKeys.ts
    stores/                  # Zustand: theme, session, sidebar, ui, notification
    hooks/                   # usePermission, useDebounce
    design-system/           # TS design tokens (colors, spacing, etc.) + IconRegistry
    constants/               # routes, permissions, currencies, countries, languages, status, taxes
    utils/                   # format, validators, permissions, debounce, clipboard, download, collection, misc, storage
    locales/                 # i18n (es-PE) — t() helper
    data/                    # mock data generators (deterministic seeded RNG)
    lib/                     # nav config
    assets/
    vite-env.d.ts
```

`@/*` resolves to `src/*` (Vite + TS path alias).

### i18n

All UI text is in Spanish (es-PE). Source code (variable names, comments if any) is English.

- Translations live in `src/locales/es.ts`.
- `t('key')` (and `t('key', { n: 5 })`) is the only way to display static UI text. Hardcoded strings in components are not allowed.
- Number / date / currency formatting goes through `formatCurrency`, `formatDate`, `formatDateTime`, `formatLongDate`, `formatNumber`, `formatPercent` in `@/utils/format` (re-exportado como `@/utils`). These use `Intl.*` with `es-PE` locale.

### Design system

Two documentos canónicos:

- **`DESIGN_SYSTEM.md`** — reglas de diseño (paleta, tipografía, spacing, componentes, accesibilidad).
- **`ARCHITECTURE.md`** — reglas arquitectónicas (capas, límites, dependencias, estado, permisos, performance).

Tokens tienen **dos representaciones sincronizadas**:
- CSS variables en `src/index.css` (lo que consume Tailwind).
- TypeScript en `src/design-system/` (type safety y valores no-CSS).

Reglas operativas:

- **No hardcoded colors.** Usa Tailwind tokens (`bg-primary`, `text-muted-foreground`, `border-destructive`, …).
- **Money usa `formatCurrency(value, 'PEN')`**, nunca `toFixed`.
- **Dates usa `formatDate(value)`**, nunca `toLocaleString` ad-hoc.
- **No emojis en la UI** salvo que el usuario lo pida.
- **Acciones destructivas** van en `<AlertDialog variant="destructive">` o `<ConfirmDialog>`.
- **Formularios** usan `<Form>` + zod + componentes de `@/components/form`.
- **Tablas** usan `<DataTable>` (no `<table>` a mano).
- **Iconos** vienen de `@/design-system/icons` (el registry). Nunca importes de `lucide-react` directamente.

### Theme (light / dark / system)

- Stored in `useThemeStore` (Zustand + `persist` to `localStorage`).
- Class strategy in Tailwind (`darkMode: ['class']`). `themeStore` toggles the `dark` class on `<html>` before React hydrates.
- `system` sigue `prefers-color-scheme`.

### Permission system

- Permisos definidos en `src/constants/permissions.ts` con claves `<module>.<verb>`.
- 6 roles predefinidos (admin, manager, accountant, seller, warehouse, viewer) con sus permisos en `RolePermissions`.
- API:
  - Hook: `usePermission(perm)`, `useRole(role)`, `useAnyPermission(perms[])`.
  - Componente: `<Can permission={...} fallback={...}>`, `<PermissionGate permission={...}>`.
  - Util: `hasPermission(ctx, perm)`, `hasAny(ctx, perms)`, `hasRole(ctx, role)`.
- Comportamiento por defecto:
  - **Sidebar** oculta items sin `*.view`.
  - **Botones** se deshabilitan (no se ocultan) cuando falta el permiso específico.

### Component organization

`src/components/` está dividido por **categoría** (no por feature). Los **features** específicos van en `src/features/<x>/`.

```
button/      # Button + 7 variants, 5 sizes
input/       # Input, Textarea, Field, Label, SearchInput
select/      # Select (Radix) + SelectContent/Item/Label/Separator
checkbox/    # Checkbox, RadioGroup, Switch
table/       # DataTable (framework), TablePagination
dialog/      # Dialog, AlertDialog (5 variants), ConfirmDialog
card/        # Card, CardHeader/Title/Description/Content/Footer, StatCard
badge/       # Badge, SaleStatusBadge, CustomerStatusBadge, StockBadge
navigation/  # Sidebar, Topbar, Breadcrumbs
feedback/    # Spinner, Skeleton, ProgressBar, EmptyState, ErrorState, Toaster
charts/      # LineChart, BarChart, PieChart (recharts wrappers)
layout/      # PageContainer, PageHeader, Section, Stack, Grid
money/       # MoneyInput, MoneyDisplay
form/        # Form + 16+ field components (TextField, MoneyField, etc.)
auth/        # Can, PermissionGate
misc/        # DropdownMenu, Separator, Tooltip
```

Cada carpeta tiene su `index.ts` barrel — importar de `@/components/<categoría>`, nunca del archivo individual.

### State management (3 capas)

| Tipo         | Herramienta          | Ejemplo                                      |
|--------------|----------------------|----------------------------------------------|
| UI / session | Zustand (persist)    | `useThemeStore`, `useSessionStore`           |
| Server state | TanStack Query       | `useCustomers()`, `useProduct(id)`            |
| Form state   | React Hook Form + Zod | `<Form onSubmit>` con zod schema             |

**Regla:** datos del backend van en TanStack Query, **nunca** se duplican en Zustand.

## Database Rules (non-negotiable)

- **Money types:** always `NUMERIC(18,2)` or `DECIMAL(18,2)`. **Never** `FLOAT`/`REAL` — this is the #1 cause of accounting rounding bugs.
- **PKs:** surrogate UUIDs (use `github.com/google/uuid`).
- **Audit columns** on every important entity: `id`, `created_at`, `updated_at`, `deleted_at` (soft delete), `created_by`, `updated_by`. See `backend/internal/domain/entities/base.go` for the `Base` struct + `Audit` embedded type to embed in every entity.
- **3NF**, FK constraints, optimized indexes.
- **Audit log** (`audit_logs` table) records every INSERT/UPDATE/DELETE/LOGIN/LOGOUT with `user_id`, `table_name`, `record_id`, `action`, `old_value`, `new_value`, `timestamp`, `ip_address`, `device`.

## Transactional Rules (non-negotiable)

- Every **sale** (and any operation that mutates inventory + receivables + accounting) **must** run inside a DB transaction. Use `database.DB.WithTx(ctx, fn)` — the helper handles begin/commit/rollback semantics and ignores `sql.ErrTxDone` on already-rolled-back txns.
- Inside the transaction, use `SELECT ... FOR UPDATE` to lock the inventory row before reading quantity.
- On any error: `ROLLBACK`. No partial writes.
- Pattern: `BEGIN → SELECT FOR UPDATE → UPDATE inventory → INSERT sale → INSERT sale_items → INSERT journal → INSERT receivable → COMMIT`.

## Security Rules

- Passwords: **Argon2id** (not bcrypt, not plain SHA). Tunables live in `config.AuthConfig` (`AUTH_ARGON_*` env vars).
- RBAC: `users`, `roles`, `permissions`, `user_roles`, `role_permissions` tables.
- All SQL parameterized; never string concatenation.
- All errors handled explicitly; structured logging via `slog` everywhere.
- Sensitive config encrypted; account lockout; session expiration.

## Migrations

- Files live in `backend/migrations/` and follow the pattern `0001_create_users.up.sql` / `0001_create_users.down.sql`.
- Each pair is a version. Runner records applied versions in `schema_migrations(version, name, applied_at)`.
- Filename must be `VERSION_NAME.up.sql` or `VERSION_NAME.down.sql` (single underscore between version and name; name may contain underscores but no dots).
- Apply with `go run ./backend/cmd/cli migrate`. Status with `go run ./backend/cmd/cli status`.
- The Wails app also auto-runs pending migrations on `OnStartup`.

## Wails / Build Gotchas

- `frontend/dist/` is gitignored but **must exist** for `//go:embed all:frontend/dist` in `main.go` to compile. Run `npm run build` in `frontend/` before `go build` or `wails build`. Wails also calls this automatically via `wails.json` `frontend:build`.
- Wails regenerates `frontend/wailsjs/go/main/` bindings on every `wails dev`/`wails build`. **Do not** edit those files manually. After adding a new exported method to a Go struct bound via `wails.Run`, regenerate by running Wails. Until then, a stub `AppBindings` in `src/vite-env.d.ts` keeps the TS build green.
- `wails.json` controls frontend scripts. `frontend:install` and `frontend:build` are run from the project root; `frontend:dev:watcher` runs Vite in the `frontend/` directory.
- The original `go.mod` shipped with a local `replace` directive pointing to a Windows path. It was removed because it was not portable. Re-add it locally if your Wails is installed in a non-default location:
  ```
  // replace github.com/wailsapp/wails/v2 v2.11.0 => /path/to/local/wails
  ```
- This dev container has Node only. `wails` CLI, `go`, and `psql` are not available; verify Go changes on a host with the Wails toolchain.

## Development Phases (from the plan)

Phase 0 = complete: project init, stack swap, env config, logger, DB driver + connection + migrations runner, frontend scaffold with Shadcn primitives.
Phase 0.5 = complete: UI/UX foundation — Spanish i18n, design system, layout shell, theme, navigation, mock data, placeholder pages. **No business logic, no DB connections yet.**
Phase 0.6 = complete: architecture & design system refinement — design tokens, icon registry, enterprise form/table frameworks, dashboard widget system, permission system, providers, services per domain, feature-based modules, lazy loading, error boundaries, full documentation.
Phase 1 = complete: enterprise database architecture — ~65 tables, 9 modules, full ERD, entity/relationship catalogs, naming/index/constraint strategies, no SQL yet.
Phase 1.1 = complete: PostgreSQL schema (Module 1: Authentication) — 10 tables + 12 paired up/down migrations + seed, validated against PostgreSQL 16. Module 1 covers `companies`, `branches`, `permissions`, `roles`, `role_permissions`, `users`, `user_roles`, `login_history`, `audit_logs`.
Phase 1.2 = complete: Domain Layer — typed errors, value objects (Money, Percentage, Quantity, SKU, Barcode, Email, Phone, DocumentNumber, Address, ExchangeRate, CurrencyCode), 15+ enums, validation package, 7 entity packages (identity, masterdata, inventory, purchasing, sales, treasury, accounting) with rich behavior, domain services (TaxCalculator, ProfitCalculator), domain events scaffold, 141 unit tests. **Zero dependencies on infrastructure.**
Phase 1.3 = complete: Repository Layer — 20+ repository interfaces in `internal/domain/repositories/`, `TransactionManager` + `UnitOfWork` + pgx error mapper, fully-implemented `CustomerRepository` with 8 integration tests against an embedded PostgreSQL 16, stubs for the remaining 19 repositories. **No business logic in the persistence layer.**
Phase 1.4 = complete: Service Layer — 8 service packages under `backend/internal/application/services/` (customer, supplier, product, inventory, purchasing, sales, customerpayments, treasury, accounting, reporting), shared errors, TxManager adapter, slog-based logger, 22 unit tests covering customer/inventory/sales/accounting. **No business logic outside the service layer.**
Phase 1.5 = next: Application use case layer (composes services across transactions, e.g. CreateSale = SalesService.Create + InventoryService.Issue + AccountingService.PostSaleEntry).
Phase 2 = after all service + use case layers: Wails bindings + UI integration.
Phase 3 = master data UI on top of the repository layer.
Phase 4 = purchasing, Phase 5 = sales, Phase 6 = inventory (incl. 25-day clearance), Phase 7 = treasury, Phase 8 = accounting, Phase 9 = dashboards, Phase 10 = PDF/Excel/CSV reports, Phase 11 = optimization.

Inventory aging rule: `max_sale_date = arrival_date + 25 days`. Items past that date are **clearance** and appear on dashboards automatically.

## Coding Standards

- SOLID, Clean Architecture, dependency injection, repository pattern.
- Business logic independent from UI.
- Small focused functions; document exported funcs; semantic versioning.
- Unit tests per module; integration tests for critical business processes (sales, payments, inventory movements).
- **No comments in code unless explicitly asked.**

## Useful References

- `DESIGN_SYSTEM.md` — visual design rules (colors, typography, spacing, components, accessibility).
- `ARCHITECTURE.md` — frontend architecture, layers, dependencies, state, permissions, performance.
- `DATABASE_ARCHITECTURE.md` — Phase 1 database design (entities, ERD, indexes, constraints, future scalability).
- `DATABASE_SCHEMA.md` — Phase 1.1+ physical schema reference (per-table columns, types, constraints, indexes).
- `backend/SERVICE_LAYER.md` — Phase 1.4 service layer reference (every service, responsibilities, conventions, anti-patterns).
- `FRONTEND_GUIDE.md` — how to add modules, forms, tables, features, services.
- `COMPONENT_GUIDE.md` — inventory and API of every reusable component.
- `FOLDER_STRUCTURE.md` — what goes where, with anti-patterns.
- `backend/README.md` — backend layout, env vars, migrations format.
- `frontend/README.md` — frontend stack and folder conventions.
- `AGENTS.md` (this file) — repo-wide rules for agents.
