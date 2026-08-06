# PROJECT_PLAN.md — vfinancy ERP

> Master development plan document. Defines vision, architecture, phases, deliverables, and acceptance criteria. Updated as each phase closes.

## Vision

Replace the business's manual processes with a centralized **desktop ERP**: purchasing, sales, inventory, treasury, accounting, and financial reports in a single, multi-role, auditable system.

**Product:** cross-platform desktop application (Wails) with a Spanish (es-PE) UI, Go backend in Clean Architecture + DDD, and PostgreSQL as the single source of truth. No offline mode: the app depends on a configured PostgreSQL instance.

## Product principles

1. **Single source of truth:** PostgreSQL. The UI never touches the DB directly — everything flows React → Wails binding → service → repository → SQL.
2. **Business rules in the backend:** zero business logic in the frontend or the repository.
3. **Exact money:** `NUMERIC(18,2)` in the DB, `Money` (rational) in Go. `FLOAT`/`REAL`/`toFixed` forbidden.
4. **Cross-domain operations in transactions:** sales = inventory + receivables + accounting, all in a single `WithTx`.
5. **Auditability:** `audit_logs` records INSERT/UPDATE/DELETE/LOGIN/LOGOUT with `user_id`, table, record, action, before/after, timestamp, IP, and device.
6. **RBAC:** 6 roles (admin, manager, accountant, seller, warehouse, viewer) with `<module>.<verb>` permissions.

## Target architecture

```
Desktop Application (Wails)
├── React Frontend  →  Wails Bindings (auto-generated)
└── Go Backend
    ├── Domain        (entities, value objects, repository interfaces)
    ├── Features      (one package per module: entity + repository + service)
    ├── Infrastructure(config, logger, database, migrations, persistence)
    └── Interfaces    (bindings exposed to the UI)
```

**Layering rule:** `internal/features/<feature>/` is the core; each feature's `postgres/` is the only layer that knows SQL; the feature service is the only orchestrator (cross-feature composition inside a transaction, **without** a use case / workflow layer); `interfaces/bindings` is the only gateway to the frontend.

## Module map

| Module        | Route            | Permission `*.view` | Phase |
|---------------|------------------|---------------------|-------|
| Home          | `/`              | —                   | 0.5 / 9 |
| Login         | `/login`         | public              | 2    |
| Customers     | `/clientes`      | customers.view      | 3    |
| Suppliers     | `/proveedores`   | suppliers.view      | 3    |
| Products      | `/productos`     | products.view       | 3    |
| Inventory     | `/inventario`    | inventory.view      | 6    |
| Purchasing    | `/compras`       | purchases.view      | 4    |
| Sales         | `/ventas`        | sales.view          | 5    |
| Treasury      | `/tesoreria`     | treasury.view       | 7    |
| Accounting    | `/contabilidad`  | accounting.view     | 8    |
| Reports       | `/reportes`      | reports.view        | 10   |
| Settings      | `/configuracion` | settings.view       | 2    |
| Administration| `/administracion`| administration.view | 2    |

## Phase schedule

| Phase | Name | Status | Deliverable |
|-------|------|--------|-------------|
| 0     | Initialization | ✅ | Wails scaffold, stack, config, logger, driver, migrations, Radix UI scaffold |
| 0.5   | UI/UX foundations | ✅ | es-PE i18n, design system, layout, theme, navigation, mock data, placeholders |
| 0.6   | Architecture refinement | ✅ | Design tokens, icon registry, form/table frameworks, widgets, permissions, providers, docs |
| 1     | DB architecture | ✅ | ~65 tables, 9 modules, ERD, catalogs, strategies |
| 1.1   | PostgreSQL schema | ✅ | 20 Module 1 migrations (auth + administration) |
| 1.2   | Domain layer | ✅ | Value objects, 20+ enums, validation, rich entities |
| 1.3   | Repository layer | ✅ | Per-feature interfaces + PostgreSQL implementations |
| 1.4   | Service layer | ✅ | Per-feature business services |
| 1.5   | Service orchestration | ✅ | Cross-feature composition in the feature service (no use case layer) |
| 2     | Bindings + UI integration | 🔄 | Wails methods + frontend consumption |
| 3     | Master data UI | ⏳ | Customers/Suppliers/Products CRUD |
| 4     | Purchasing | ⏳ | Purchase orders, receiving, supplier payments |
| 5     | Sales | ⏳ | Sales, collections, advances, returns |
| 6     | Inventory | ⏳ | Movements, batches, 25-day clearance |
| 7     | Treasury | ⏳ | Bank accounts, transactions, cards |
| 8     | Accounting | ⏳ | Chart of accounts, journals, entries, periods |
| 9     | Dashboards | ⏳ | KPIs, charts, clearance alerts |
| 10    | Reports | ⏳ | PDF / Excel / CSV |
| 11    | Optimization | ⏳ | Performance, packaging, hardening |

`✅ complete · 🔄 in progress · ⏳ pending`

---

## Detailed phases

### Phase 0 — Initialization `[✅]`
- Wails v2 scaffold (Go + React + TS + Vite).
- Config via env vars (`config.Config`), `slog` logger (json|text).
- PostgreSQL connection (`pgx/v5/stdlib`), `EnsureDatabase`, migration runner.
- Frontend scaffold with Radix primitives.

### Phase 0.5 — UI/UX foundations `[✅]`
- es-PE i18n (`src/locales/es.ts`, `t()` helper).
- Initial design system, layout shell (sidebar + topbar + breadcrumbs), light/dark/system theme.
- Deterministic mock data, placeholder page per module.

### Phase 0.6 — Architecture refinement `[✅]`
- Design tokens in two representations: CSS vars (`src/index.css`) + TS (`src/design-system/`).
- Self-contained **plain CSS3** styling system (no Tailwind/PostCSS): tokens, utilities, responsive variants (`sm:`/`lg:`/`xl:`), Radix states, `vf-enter`/`vf-exit` animations, `prefers-reduced-motion`, `hsl(var(--x) / N)` opacities, `--vf-offset` focus rings.
- Icon registry, DataTable and Form frameworks, dashboard widgets, permissions, providers, lazy loading, error boundaries.
- Documentation: `DESIGN.md` (design rules), frontend guides.

### Phase 1 — DB architecture `[✅]`
- ~65 tables, 9 modules; full ERD; entity and relationship catalogs.
- Naming, index, and constraint strategies; **3NF**; FKs with indexes; UUIDs; `NUMERIC(18,2)`.

### Phase 1.1 — PostgreSQL schema `[✅]`
- 20 `up/down` pairs (0000–0019) validated on PostgreSQL 16.
- Module 1 (auth): `companies`, `branches`, `permissions`, `roles`, `role_permissions`, `users`, `user_roles`, `login_history`, `audit_logs`, `user_sessions`, `user_profiles`.
- Module 1.5 (administration): `application_settings`, `currencies`, `countries`, `taxes`, `exchange_rates`, `audit_events`.
- `set_updated_at()` trigger, `schema_migrations` control table, auth seed.

### Phase 1.2 — Domain layer `[✅]`
- `internal/domain/`: `enums` (20+), `errors` (`derrors`: business/validation/notfound/conflict), `valueobjects` (Money, Percentage, Quantity, SKU, Barcode, Email, Phone, DocumentNumber, Address, ExchangeRate, CurrencyCode, ChartCode, Reference), `repositories` (pagination, transaction, errors).
- Rich per-feature entities in `internal/features/<feature>/` (e.g. `auth.User`, `customer.Customer`, `sales.Sale`).
- **Zero infrastructure dependencies.**

### Phase 1.3 — Repository layer `[✅]`
- Per-feature interfaces in `internal/features/<feature>/` (e.g. `UserRepository`, `CustomerRepository`, `SalesRepository`).
- PostgreSQL implementations in `features/<feature>/postgres/` using `infrastructure/persistence/` helpers (`Querier`, `TxManager`, scan/decode/builder/error-map).
- **No business logic in persistence; 100% parameterized SQL.**

### Phase 1.4 — Service layer `[✅]`
- Per-feature services (`customer_service.go`, `sales_service.go`, `inventory_service.go`, `auth` (authentication/profile/session), `administration` (settings/audit), `treasury`, `purchasing`, `accounting`, `reporting`, `customerpayments`).
- Shared errors in `internal/shared/apperrors` (sentinels + `MapError`/`Errorf`); `shared/logger` logger.
- **No business logic outside the service layer.**

### Phase 1.5 — Service orchestration `[✅]`
- **No use case / workflow layer.** Each feature is a vertical slice (entity + repository + service); the feature service is the only orchestrator.
- Cross-feature composition within a single transaction via `repositories.TransactionManager.WithinTransaction` (joins active tx in `ctx`):
  - `auth.AuthenticationService.Login` = authentication + session + audit.
  - `sales.SalesService.Create` = sale + customer debt.
- Pattern: `BEGIN → SELECT … FOR UPDATE → UPDATE inventory → INSERT sale → INSERT sale_items → INSERT journal → INSERT receivable → COMMIT`; `ROLLBACK` on error.

### Phase 2 — Bindings + UI integration `[🔄]`
- `backend/interfaces/bindings/`: `App` (auth: login/logout/changePassword/validateSession; profile; settings: businessInfo/preferences/currencies/taxes; system: connection config, app settings, modules).
- Frontend: `src/services/bindings.ts` (`wailsClient`) loads generated bindings via `import('../../wailsjs/go/bindings/App')` with a mock fallback when there is no Wails runtime (`window.go`).
- Per-domain services (`services/auth`, `services/settings`, …) with TanStack Query query keys.
- Next: Login/Settings/Administration pages connected to real bindings.

### Phase 3 — Master data UI `[⏳]`
- Customers, Suppliers, Products CRUD (categories, brands, units, warehouses).
- Lists with `DataTable` + pagination, search/filters, `Form` + zod forms.
- Create, edit, deactivate (soft delete), block on overdue debt, credit limit.

### Phase 4 — Purchasing `[⏳]`
- Purchase orders, receiving, purchase statuses, supplier payments, returns.
- Impact: inventory + accounts payable + accounting (transactional workflow).

### Phase 5 — Sales `[⏳]`
- Sales, `sale_items`, customer payments, advances, sale statuses.
- Full `sales.SalesService.Create`: inventory + receivables + accounting in a transaction.
- Business rule: automatic blocking of customers with overdue debt.

### Phase 6 — Inventory `[⏳]`
- Movements (in/out), batches (`batch_status`), adjustments.
- **Clearance rule:** `max_sale_date = arrival_date + 25 days`. Expired items = clearance, `Remate` badge in the UI, and automatically visible on dashboards.

### Phase 7 — Treasury `[⏳]`
- Bank accounts, transactions, credit cards, exchange rate (`exchange_rates`).

### Phase 8 — Accounting `[⏳]`
- Chart of accounts (`chart_of_account`), entries (`journal_entry` + `journal_entry_line`), periods, journal statuses.
- Integration with sales/purchases/treasury for automatic entries.

### Phase 9 — Dashboards `[⏳]`
- Per-module KPIs (Phase 0.6 widget system), recharts charts (line/bar/pie).
- Automatic clearance alerts; trends vs. previous period.

### Phase 10 — Reports `[⏳]`
- PDF / Excel / CSV export from all tables.
- Financial reports: statement of account, accounts receivable/payable, inventory, general journal.

### Phase 11 — Optimization `[⏳]`
- Performance (indexes, pagination, lazy loading), Wails packaging, security hardening, backups.

---

## Cross-cutting business rules (non-negotiable)

1. **Money:** `NUMERIC(18,2)` / `DECIMAL(18,2)`. Never `FLOAT`/`REAL`. UI with `formatCurrency(value, 'PEN')`.
2. **IDs:** surrogate UUIDs (`github.com/google/uuid`).
3. **Audit:** `id`, `created_at`, `updated_at`, `deleted_at`, `created_by`, `updated_by` on every important entity; `audit_logs` records every change.
4. **Transactions:** any operation that mutates inventory + receivables/payables + accounting **must** go in `WithTx` with `SELECT … FOR UPDATE`.
5. **Parameterized SQL:** never string concatenation.
6. **Security:** Argon2id for passwords (never bcrypt/plain SHA); account lockout; session expiration.
7. **i18n:** UI texts in es-PE via `t()`; code names in English.
8. **Permissions:** buttons are **disabled** (not hidden) without the specific permission; sidebar hides without `*.view`.

## Phase close criteria

- Code compiles: `go build ./...` and `npm run build` + `npm run check` without errors.
- Tests (where applicable): `go test ./backend/...`.
- Documentation updated (`AGENTS.md`, `PROJECT_PLAN.md`, `DESIGN.md`).
- Bindings regenerated if the Go surface changed.
- No code comments unless explicitly requested.

## Risks and mitigations

| Risk | Impact | Mitigation |
|------|--------|------------|
| Money rounding errors | High | `NUMERIC(18,2)` + rational `Money` + `formatCurrency` |
| Partial multi-table writes | High | Service composition with `WithTx` + `SELECT … FOR UPDATE` + `ROLLBACK` |
| Drift between CSS vars and TS tokens | Medium | Two documented representations; double update on token changes |
| Drift between Go bindings and TS types | Medium | `wailsjs/` regenerated by Wails; `bindings.ts` as single access point |
| Mutable migrations | High | Immutable migrations; changes = new sequential version |
| Business logic in the wrong layer | Medium | Review by layering rules; service tests |
