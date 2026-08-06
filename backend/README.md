# vfinancy Backend

Clean Architecture + DDD backend for the vfinancy ERP desktop application.

## Layout

```
backend/
  cmd/
    cli/                          # standalone CLI (migrate, status)
    server/                       # Wails entrypoint (planned for Phase 0+)
  internal/
    domain/
      entities/                   # core entities (Base, User, Customer, ...)
      valueobjects/               # ID, Money, ...
      repositories/               # repository interfaces
    application/
      services/                   # application services
      usecases/                   # use case orchestrators
    infrastructure/
      config/                     # env-based config loader
      logger/                     # structured logging (slog)
      database/                   # *sql.DB wrapper + WithTx helper
      postgres/                   # DSN helpers, EnsureDatabase
      migrations/                 # file-based SQL migration runner
    interfaces/
      bindings/                   # structs exposed to the React frontend
  pkg/                            # reusable packages (no business logic)
  migrations/                     # SQL migration files (0001_xxx.up.sql / .down.sql)
```

## Modules Are Independent

- `domain/` must not import `infrastructure/`, `application/`, or `interfaces/`.
- `application/` may import `domain/`.
- `infrastructure/` implements `domain/repositories` interfaces and provides config, logger, DB.
- `interfaces/bindings/` is the only package the Wails runtime imports; it wires the application services.

## Environment

All config is loaded from env vars. See `internal/infrastructure/config/config.go` for the full list. Examples:

```bash
export DB_HOST=localhost
export DB_PORT=5432
export DB_USER=postgres
export DB_PASSWORD=casa123
export DB_NAME=vfinancy
export LOG_LEVEL=info            # debug | info | warn | error
export LOG_FORMAT=json           # json | text
```

## CLI

```bash
go run ./cmd/cli migrate         # apply pending migrations
go run ./cmd/cli status          # show migration status
```

## Migrations

Migration files live in `backend/migrations/` and follow the pattern:

```
0000_init.up.sql
0000_init.down.sql
0001_create_companies.up.sql
0001_create_companies.down.sql
0002_create_branches.up.sql
0002_create_branches.down.sql
...
0011_seed_auth.up.sql
0011_seed_auth.down.sql
```

Each pair must have a matching version. The runner is transactional and records applied versions in `schema_migrations(version, name, applied_at)`.

**Phase 1.1 / Module 1 (Authentication)** is implemented: `companies`, `branches`, `permissions`, `roles`, `role_permissions`, `users`, `user_roles`, `login_history`, `audit_logs`, plus a `set_updated_at()` trigger and the `schema_migrations` bookkeeping table.

Reference docs:
- `DATABASE_ARCHITECTURE.md` — the approved high-level design (entities, ERD, strategies).
- `DATABASE_SCHEMA.md` — column-level reference for every implemented table.

Migrations are **immutable** once committed. Any schema change requires a new file with the next sequential version number.

## Repository Layer

The repository abstraction lives in `internal/domain/repositories/`. Every repository
exposes only operations the application layer needs; the implementation
lives in `internal/infrastructure/persistence/postgres/`.

```
internal/domain/repositories/
  pagination.go        # Page[T], PageRequest, Sort, TimeRange
  errors.go            # ErrNotFound, ErrDuplicate, ErrForeignKey, ...
  transaction.go       # TransactionManager
  unit_of_work.go      # UnitOfWork
  user_repository.go, role_repository.go, ...   # one interface per aggregate

internal/infrastructure/persistence/postgres/
  errors.go            # *pgconn.PgError → domain error mapping
  tx.go                # Querier interface, dbBox, txBox
  unit_of_work.go      # UnitOfWork implementation
  repositories.go      # Repositories facade
  common.go            # limitOffset, joinClauses, helpers
  decode.go            # DB string → typed domain values
  customer_repository.go       # full implementation + 8 integration tests
  stubs.go             # placeholders for the remaining 19 repos
```

**Repository rules** (enforced by code review):
- Never contain business logic. No totals, taxes, profit, inventory adjustments, accounting.
- Only persistence: read rows into entities, write entities into rows.
- Use `Querier` (an interface satisfied by both `*sql.DB` and `*sql.Tx`) so the same code runs against the connection pool (auto-commit) or a transaction.
- Always go through `Translate()` for pgx error mapping; never return raw `*pgconn.PgError`.
- Use placeholder parameters (`$1`, `$2`, …). Never string-concat user input into SQL.

**Tests** are integration tests against an embedded PostgreSQL 16. The setup boots the postgres binary, creates the test database, applies all migrations, and creates the Module 2 tables on demand. Run with `go test -count=1 ./backend/internal/infrastructure/persistence/postgres/...`.

## Money

All monetary fields use `NUMERIC(18,2)` in PostgreSQL. Go-side calculations use `valueobject.Money` (rational number) to avoid float rounding.

## Transactions

`database.DB.WithTx(ctx, func(tx *Tx) error { ... })` wraps the unit of work. Every sale / purchase / payment / inventory adjustment that touches multiple tables must run inside `WithTx` and use `SELECT ... FOR UPDATE` to lock inventory rows before reading quantities.
