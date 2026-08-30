# AGENTS.md

## Project Overview

**vfinancy** is a custom Desktop ERP (Enterprise Resource Planning) system built with Wails v2 (Go + React + TypeScript) backed by PostgreSQL. It replaces manual business processes with a centralized platform for purchasing, sales, inventory, treasury, accounting, and financial reporting.

## Target Stack (mandatory)

- **Desktop framework:** Wails v2
- **Backend:** Go 1.23
- **Frontend:** TypeScript + **React 19** + Vite 5
- **State / data:** Zustand, TanStack Query, React Hook Form, Zod
- **Styling:** plain CSS3 in `src/index.css` (semantic per-component classes + design tokens) — **no Tailwind, no PostCSS**
- **DB:** hybrid — local **SQLite** (primary runtime DB, `modernc.org/sqlite`, pure Go) + cloud **PostgreSQL** mirror (via `github.com/jackc/pgx/v5/stdlib`) synchronized by the built-in sync engine.

## Key Commands

```bash
# Full app (Wails) — must be run on a machine with Wails + Go toolchain installed
wails dev          # Live dev (Go + Vite, hot-reload)
wails build        # Production build → build/bin/

# Frontend-only (run in frontend/)
pnpm install       # install deps
pnpm dev           # Vite dev server (no Go backend)
pnpm build         # Type-check + production build → frontend/dist/
pnpm check         # tsc --noEmit

# Backend CLI (run in repo root or backend/)
go run ./backend/cmd/cli migrate [--postgres]   # apply pending migrations (default: local SQLite via migrations/sqlite)
go run ./backend/cmd/cli status  [--postgres]   # show migration status

# Sync server (cloud PostgreSQL mirror)
DB_DRIVER=postgres DB_MIGRATION_DIR=backend/migrations/postgres SYNC_ENABLED=false \
  go run ./backend/cmd/syncserver

# Go tests
go test ./backend/...
```

Toolchain notes for this dev container:

- **Go IS available** at `/usr/local/go/bin/go` (used for `go build`, `go vet`, `go test`). `pnpm` is on PATH too. Only the `wails` CLI and `psql` are missing.
- `modernc.org/sqlite` is a declared dependency (pure Go, no CGO) — needed to build the desktop app offline.
- The backend assumes a working directory of the repo root: set `DB_MIGRATION_DIR=backend/migrations/sqlite` (or `backend/migrations/postgres`) and `SYNC_ENABLED=false` unless a sync server URL is configured.

## Architecture: feature-based vertical slices (Service + Repository)

```
Desktop Application (Wails)
├── React Frontend  →  Wails Bindings (auto-generated)
└── Go Backend
    ├── Domain          (shared cross-feature primitives: enums, errors, value objects, repo contracts)
    ├── Features        (vertical slices: entity + repository interface + service, one package per module)
    │   └── <feature>/postgres  (concrete SQL repository implementations)
    ├── Shared          (apperrors, logger — cross-feature helpers, no business logic)
    ├── Infrastructure  (DB, config, logger, migrations, persistence)
    └── Interfaces      (Wails-exposed bindings)
```

**Rule:** Frontend never accesses the database directly. All traffic flows React → Wails binding → feature service → Repository → local SQLite (primary runtime DB) or the cloud PostgreSQL mirror (via the sync server).

**Vertical-slice rules:**

- Each feature is a **vertical slice**: `internal/features/<feature>/` holds the entities, the repository interface, and the feature service that owns every business operation of that module. Nothing lives in a separate use-case/workflow layer.
- The **service is the only orchestrator**. Cross-feature operations (e.g. creating a sale touches sales + customer debt) are composed inside the owning feature's service, in a single transaction. There is **no** use case / workflow / application-service layer.
- The **service talks to repositories and other services** — never to SQL. Only the `postgres/` subpackage of the same feature (via `infrastructure/persistence` helpers) knows SQL.
- Cross-feature composition happens through **service-to-service or service-to-repository calls inside `repositories.TransactionManager.WithinTransaction`**. The TxManager joins an already-active transaction in `ctx`, so nested calls stay on one DB transaction (BEGIN → … → COMMIT, ROLLBACK on error).
- Shared application-level error sentinels (`ErrValidation`, `ErrNotFound`, `ErrConflict`, `ErrUnauthorized`, `ErrInternal`), `apperrors.Errorf`, and `apperrors.MapError` live in `internal/shared/apperrors`. Feature-specific typed errors (e.g. `auth.ErrAuthLocked`) live in the feature package.

## Backend Layout (current)

```
backend/
  cmd/
    cli/                 # standalone CLI: `migrate`, `status` (default: local SQLite; `--postgres` for cloud)
    syncserver/          # self-hosted HTTP sync server (cloud PostgreSQL mirror)
  infrastructure/        # NOT under internal/ on purpose (see "Internal packages" below)
    config/              # env-based config loader (DB_DRIVER, DB_PATH, SYNC_*, ...)
    logger/              # structured slog logger (json | text, levels debug..error)
    database/            # *sql.DB wrapper + WithTx(ctx, fn) helper
    postgres/            # Connect(), EnsureDatabase(), DSN helpers
    sqlite/              # embedded SQLite driver (pure Go, modernc.org/sqlite)
    migrations/          # file-based SQL migration runner (dialect-aware bookkeeping)
    persistence/         # Querier, TxManager, dialect switch, scan/decode/builder/error-map helpers
  interfaces/
    bindings/            # Wails-exposed methods (auth, profile, settings, system) + sync worker bootstrap
  internal/
    domain/              # cross-feature domain primitives
      enums/             # 20+ enums (sale_status, journal_type, ...)
      errors/            # derrors: business, validation, notfound, conflict
      valueobjects/      # Money, Percentage, Quantity, Email, SKU, ...
      repositories/      # shared repo interfaces (pagination, transaction, errors)
    shared/
      apperrors/         # shared app-level error sentinels + MapError/Errorf helpers
      logger/            # slog-based app logger
    features/            # one vertical slice per business module
      <feature>/         # entity(ies) + repository interface + service (orchestration lives here)
      <feature>/postgres/# concrete repository implementations
      sync/              # replication engine: device registry, cursors, LWW conflicts, tombstones
  migrations/
    sqlite/              # SQLite schema migrations (up.sql only) — local runtime DB
    postgres/            # PostgreSQL mirror (same version set/order as sqlite)
  pkg/                   # reusable packages
```

Each feature (`auth`, `administration`, `customer`, `supplier`, `product`, `inventory`, `purchasing`, `sales`, `treasury`, `accounting`, `customerpayments`, `reporting`, `notifications`) is a **vertical slice**: it owns its entity definitions, its repository interface, and its business service. When an operation spans multiple features, the owning feature's service composes the other features' services/repositories inside a single transaction (e.g. `sales.SalesService.Create` records the sale AND the customer debt; `auth.AuthenticationService.Login` authenticates, creates the session and records the audit event). Concrete SQL lives only in the feature's `postgres/` subpackage.

The root `main.go` and `app.go` are the Wails entrypoint. `app.go` initializes config + logger, ensures the database exists, opens a connection, and runs pending migrations on startup. It binds two structs: the root `App` (config accessors for the frontend) and `bindings.App` (Wails-exposed methods).

### Internal packages

`backend/internal/...` packages can only be imported by packages whose import path is `vfinancy/backend/...` or a subpath. The root `vfinancy` package (where `main.go`/`app.go` live) cannot import them. So `infrastructure/` and `interfaces/` live directly under `backend/` (not under `internal/`). Moving the root entrypoint into `backend/cmd/server/` would let these be folded back into `internal/`.

## Frontend Layout (current)

```
frontend/
  index.html
  package.json
  tsconfig.json + tsconfig.app.json + tsconfig.node.json
  vite.config.ts
  src/
    main.tsx
    app/                     # App.tsx (routes + lazy), Providers.tsx, ErrorBoundary.tsx
    layouts/                 # AppLayout (sidebar + topbar + breadcrumbs)
    pages/                   # route screens (1 per module + Login)
    features/                # feature-based modules (dashboard/, customers/, sales/)
    components/              # category folders — see "Component organization" below
    services/                # one folder per business domain + queryKeys.ts
    stores/                  # Zustand: theme, sidebar, ui, notification
    hooks/                   # useDebounce
    design-system/           # TS design tokens (colors, spacing, etc.) + IconRegistry
    constants/               # routes, currencies, countries, languages, status, taxes
    utils/                   # format, validators, debounce, clipboard, download, collection, misc, storage
    locales/                 # i18n (es-PE) — t() helper
    types/                   # shared domain types (Customer, Product, Supplier, Sale, InventoryItem, ...)
    lib/                     # nav config
    assets/
    vite-env.d.ts
```

`@/*` resolves to `src/*` (Vite + TS path alias).

### i18n

All UI text is in Spanish (es-PE). Source code (variable names, comments if any) is English.

- Translations live in `src/locales/es.ts`.
- `t('key')` (and `t('key', { n: 5 })`) is the only way to display static UI text. Hardcoded strings in components are not allowed.
- Number / date / currency formatting goes through `formatCurrency`, `formatDate`, `formatNumber`, `formatPercent` in `@/utils/format`. These use `Intl.*` with `es-PE` locale.

### Design system

Un documento canónico:

- **`DESIGN.md`** — reglas de diseño (paleta, tipografía, spacing, componentes, accesibilidad).

Tokens tienen **dos representaciones sincronizadas**:
- CSS variables en `src/index.css` (lo que consume la UI en runtime).
- TypeScript en `src/design-system/` (type safety y valores no-CSS).

Estilos: `src/index.css` es un sistema **plain CSS3** auto-suficiente (sin Tailwind/PostCSS). Define tokens, reset/base y **clases semánticas por componente** (`.btn`, `.card`, `.dialog-overlay`, `.sidebar`, `.datatable`, …) con variantes BEM-style (`--primary`, `--collapsed`, `__header`). Los componentes React reciben estilos propios; para composición puntual hay helpers mínimos (`.stack`, `.hstack`, `.grid-N`). Soporta Radix `data-[state=*]`/`data-[side=*]`, animaciones (`vf-enter`/`vf-exit`) y `prefers-reduced-motion`. Usa `hsl(var(--x) / N)` para opacidades y `--vf-offset`/`--vf-offset-color` para focus rings. La unión condicional de clases se hace con `cx()` de `@/utils/cx`.

Reglas operativas:

- **No hardcoded colors.** Usa tokens CSS (`hsl(var(--primary))`, `hsl(var(--muted-foreground))`, `hsl(var(--destructive))`, …) — definidos en `src/index.css`.
- **No utility classes** (`bg-*`, `text-sm`, `p-4`, `flex items-center gap-2`, …). Usa la clase semántica del componente o los helpers de layout (`.stack`, `.hstack`, `.grid-N`); casos únicos van con `style={{...}}`.
- **Money usa `formatCurrency(value, 'PEN')`**, nunca `toFixed`.
- **Dates usa `formatDate(value)`**, nunca `toLocaleString` ad-hoc.
- **No emojis en la UI** salvo que el usuario lo pida.
- **Acciones destructivas** van en `<AlertDialog variant="destructive">` o `<ConfirmDialog>`.
- **Formularios** usan `<Form>` + zod + componentes de `@/components/form`.
- **Tablas** usan `<DataTable>` (no `<table>` a mano).
- **Iconos** vienen de `@/design-system/icons` (el registry). Nunca importes de `lucide-react` directamente.

### Theme (light / dark / system)

- Stored in `useThemeStore` (Zustand + `persist` to `localStorage`).
- Class strategy: `.dark` variants are defined in `src/index.css`. `themeStore` toggles the `dark` class on `<html>` before React hydrates.
- `system` sigue `prefers-color-scheme`.

### Component organization

`src/components/` está dividido por **categoría** (no por feature). Los **features** específicos van en `src/features/<x>/`.

```
button/      # Button + 7 variants, 5 sizes
input/       # Input, Textarea, Label, SearchInput
select/      # Select (Radix) + SelectContent/Item/Label/Separator
checkbox/    # Checkbox, RadioGroup, Switch
table/       # DataTable (framework), TablePagination
dialog/      # Dialog, AlertDialog (5 variants), ConfirmDialog
card/        # Card, CardHeader/Title/Description/Content/Footer, StatCard
badge/       # Badge, SaleStatusBadge, CustomerStatusBadge
navigation/  # Sidebar, Topbar, Breadcrumbs
feedback/    # Spinner, Skeleton, ProgressBar, EmptyState, ErrorState, Toaster
charts/      # LineChart, BarChart, PieChart (recharts wrappers)
layout/      # PageContainer, PageHeader, Section, Stack, Grid
money/       # MoneyInput, MoneyDisplay
form/        # Form + 16+ field components (TextField, MoneyField, etc.)
auth/        # Can
misc/        # DropdownMenu, Separator, Tooltip
```

Cada carpeta tiene su `index.ts` barrel — importar de `@/components/<categoría>`, nunca del archivo individual.

### State management (3 capas)

| Tipo         | Herramienta          | Ejemplo                                      |
|--------------|----------------------|----------------------------------------------|
| UI / session | Zustand (persist)    | `useThemeStore`, `useSidebarStore`           |
| Server state | TanStack Query       | `useCustomers()`, `useProduct(id)`            |
| Form state   | React Hook Form + Zod | `<Form onSubmit>` con zod schema             |

**Regla:** datos del backend van en TanStack Query, **nunca** se duplican en Zustand.

## Database Rules (non-negotiable)

- **Money types:** always `NUMERIC(18,2)` or `DECIMAL(18,2)`. **Never** `FLOAT`/`REAL` — this is the #1 cause of accounting rounding bugs.
- **PKs:** surrogate UUIDs (use `github.com/google/uuid`).
- **Audit columns** on every important entity: `id`, `created_at`, `updated_at`, `deleted_at` (soft delete), `created_by`, `updated_by`. Feature entities carry these fields directly (e.g. `auth.User`, `customer.Customer`).
- **3NF**, FK constraints, optimized indexes.
- **Audit log** (`audit_logs` table) records every INSERT/UPDATE/DELETE/LOGIN/LOGOUT with `user_id`, `table_name`, `record_id`, `action`, `old_value`, `new_value`, `timestamp`, `ip_address`, `device`.

## Transactional Rules (non-negotiable)

- Every **sale** (and any operation that mutates inventory + receivables + accounting) **must** run inside a DB transaction. Use `repositories.TransactionManager.WithinTransaction(ctx, fn)` (backed by `database.DB.WithTx`) — it begins/commits/rolls back, joins an already-active transaction in `ctx` (so composed service calls share one DB transaction), and ignores `sql.ErrTxDone` on already-rolled-back txns.
- Inside the transaction, use `SELECT ... FOR UPDATE` to lock the inventory row before reading quantity.
- On any error: `ROLLBACK`. No partial writes.
- Pattern: `BEGIN → SELECT FOR UPDATE → UPDATE inventory → INSERT sale → INSERT sale_items → INSERT journal → INSERT receivable → COMMIT`.

## Security Rules

- Passwords: **Argon2id** (not bcrypt, not plain SHA). Tunables live in `config.AuthConfig` (`AUTH_ARGON_*` env vars).
- All SQL parameterized; never string concatenation.
- All errors handled explicitly; structured logging via `slog` everywhere.
- Sensitive config encrypted; account lockout; session expiration.

## Migrations

- Files live in `backend/migrations/` under two dialect directories: `sqlite/` (the local runtime DB) and `postgres/` (the cloud mirror). Both contain the same versions in the same creation order; keep them in sync.
- Only `.up.sql` exists today: this is a pre-release app with no production data, so rollback scripts were intentionally dropped and the runner exposes no rollback command (the `cli` has only `migrate` and `status`).
- Filename must be `VERSION_name.up.sql` (single underscore between version and name; name may contain underscores but no dots). The initial schema is written in FK-safe creation order (companies/branches first, then reference data, then purchasing → inventory → sales). FKs must never reference a table created later in the file: PostgreSQL rejects forward references at CREATE time, while SQLite silently tolerates them.
- Each file is a version. Runner records applied versions in `schema_migrations(version, name, applied_at)`.
- Apply with `go run ./backend/cmd/cli migrate` (default: local SQLite; add `--postgres` for the cloud). Status with `go run ./backend/cmd/cli status`.
- The Wails app also auto-runs pending migrations on `OnStartup`.
- SQLite dialect notes: `gen_random_uuid()` → `lower(hex(randomblob(16)))`, `NOW()` → `(CAST(unixepoch('subsec') * 1000 AS INTEGER))`, `TIMESTAMPTZ`/`JSONB`/`TEXT[]`/`INET` → `TIMESTAMP`/`TEXT`/`TEXT`/`TEXT`. Timestamps are INTEGER ms; `ALTER TABLE ADD CONSTRAINT` is a no-op (FKs are declared inline in `CREATE TABLE`). SQLite requires all columns before table-level `CONSTRAINT`s, and triggers cannot be `BEFORE UPDATE OR DELETE`.

## Sync architecture

- **Local SQLite is the runtime DB.** The background worker in `interfaces/bindings/app.go` (`startSyncWorker`, gated on `SYNC_ENABLED`) pushes/pulls changes to the self-hosted sync server (`backend/cmd/syncserver`, cloud PostgreSQL mirror).
- **Model:** watermark-diff replication with last-writer-wins. Rows travel as "all rows whose time column > per-table cursor" in both directions; hard deletes are captured by the `AFTER DELETE` triggers in `migrations/sqlite/0000_initial_schema` into `sync_tombstones`. The postgres 0000 has no triggers — the server records tombstones explicitly. There are no INSERT/UPDATE triggers (no outbox echo).
- The generic engine lives in `internal/features/sync` (entities, registry of 13 replicated tables, `Repository` interface, `HTTPClient`) and `internal/features/sync/postgres` (dialect-safe implementation). `internal/features/sync/server.go` is the server-side handler.
- Replicated tables are the master-data/auth set: `companies`, `branches`, `roles`, `users`, `user_roles`, `user_profiles`, `user_sessions`, `application_settings`, `taxes`, `currencies`, `countries`. `audit_logs`, `audit_events`, `login_history`, `exchange_rates` are intentionally excluded.
- Config: `SYNC_ENABLED=true`, `SYNC_SERVER_URL`, `SYNC_API_KEY`, `SYNC_POLL_INTERVAL_SEC=30`.

## Notifications (device-local feed)

- `internal/features/notifications` owns the in-app notification feed (`notifications` table: `type`, `title`, `message`, `record_type`, `record_id`, `dedup_key`, `read_at`, soft delete). Unread = `read_at IS NULL`; dedup via `UNIQUE (company_id, type, dedup_key)`.
- The only generator today is **inventory clearance**: `notifications.NotificationsService.Generate` scans `InventoryService.GenerateClearanceCandidates`, inserts one notification per on-clearance batch (`dedup_key = batch id`), and soft-deletes **unread** notifications whose batch left clearance (read ones stay as history).
- The `startNotificationsWorker` goroutine in `interfaces/bindings/app.go` (gated on `NOTIFICATIONS_ENABLED`, default `true`; interval `NOTIFICATIONS_POLL_INTERVAL_SEC`, default `60`) runs `InventoryService.RefreshClearanceFlags` (reconciles the persisted `is_clearance` column, which is otherwise only refreshed on batch writes) and then `Generate`. It skips when no company is active.
- Delivery is **in-app only** (Topbar bell, `services/notifications`); notifications are device-local and intentionally excluded from the sync engine.

## Wails / Build Gotchas

- `frontend/dist/` is gitignored but **must exist** for `//go:embed all:frontend/dist` in `main.go` to compile. Run `pnpm build` in `frontend/` before `go build` or `wails build`. Wails also calls this automatically via `wails.json` `frontend:build`.
- Wails regenerates `frontend/wailsjs/` bindings on every `wails dev`/`wails build`. **Do not** edit those files manually. After adding a new exported method to a Go struct bound via `wails.Run`, regenerate by running Wails. The frontend loads them at runtime through `src/services/bindings.ts` (dynamic `import('../../wailsjs/go/bindings/App')`; it throws a clear error outside the Wails runtime — there is **no** mock fallback, the UI always renders real DB data).
- `wails.json` controls frontend scripts. `frontend:install` and `frontend:build` are run from the project root; `frontend:dev:watcher` runs Vite in the `frontend/` directory.
- The original `go.mod` shipped with a local `replace` directive pointing to a Windows path. It was removed because it was not portable. Re-add it locally if your Wails is installed in a non-default location:
  ```
  // replace github.com/wailsapp/wails/v2 v2.11.0 => /path/to/local/wails
  ```
- This dev container has Go (`/usr/local/go/bin/go`) and Node, but no `wails` CLI or `psql`. `go build`/`go vet`/`go test` run fine here; only Wails-specific steps need a host with the Wails toolchain.

Inventory aging rule: `max_sale_date = arrival_date + 25 days`. Items past that date are **clearance** and appear on dashboards automatically.

## Coding Standards

- SOLID, Clean Architecture, dependency injection, repository pattern.
- Feature-based vertical slices: **service + repository** per module. No use-case / workflow / application-service layer — the feature service is the only orchestrator.
- Business logic independent from UI.
- Small focused functions; document exported funcs; semantic versioning.
- Unit tests per module; integration tests for critical business processes (sales, payments, inventory movements).
- **No comments in code unless explicitly asked.**

## Useful References

- `DESIGN.md` — visual design rules (colors, typography, spacing, components, accessibility).
- `frontend/README.md` — frontend stack and folder conventions.
- `AGENTS.md` (this file) — repo-wide rules for agents.
