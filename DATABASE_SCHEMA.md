# DATABASE_SCHEMA.md

> Referencia técnica de las tablas, columnas, tipos, constraints e índices ya materializados en SQL.
>
> **Estado actual:** solo Module 1 (Authentication) está implementado en SQL. Los módulos siguientes se agregarán en las próximas fases de Phase 1.1.
>
> **Cómo usar este documento:** cada tabla incluye su definición completa, los constraints de negocio que la protegen, los índices con su propósito y la query típica que los aprovecha. Cuando un módulo aún no esté implementado, se lista como `— pendiente —` con la referencia a `DATABASE_ARCHITECTURE.md`.

---

## Convenciones de la documentación

Cada tabla se documenta con:

- **Propósito** — una línea.
- **Columnas** — nombre, tipo, nullability, default, descripción.
- **PK** y **FKs** explícitas.
- **Constraints** — CHECK, UNIQUE (con o sin parcial), EXCLUDE.
- **Índices** — nombre, columnas, parcialidad, query que sirve.
- **Triggers** — qué hacen y cuándo se disparan.
- **Audit** — estrategia (columnas embebidas, soft delete, etc.).
- **Relaciones inversas** — qué tablas la referencian.

---

# Module 1: Authentication

## Tabla: `companies`

Empresa raíz multi-tenant. Toda tabla de negocio tiene `company_id` apuntando aquí.

| Columna                      | Tipo          | Null     | Default                                | Descripción                                       |
|------------------------------|---------------|----------|----------------------------------------|---------------------------------------------------|
| `id`                         | `UUID`        | NOT NULL | `gen_random_uuid()`                    | PK                                                |
| `code`                       | `VARCHAR(20)` | NOT NULL | —                                      | Identificador corto estable (e.g. `VFI`)          |
| `legal_name`                 | `VARCHAR(200)`| NOT NULL | —                                      | Razón social                                       |
| `trade_name`                 | `VARCHAR(200)`| NULL     | —                                      | Nombre comercial                                  |
| `tax_id`                     | `VARCHAR(30)` | NOT NULL | —                                      | RUC / EIN / NIT                                   |
| `address`                    | `TEXT`        | NULL     | —                                      | Dirección fiscal                                  |
| `phone`                      | `VARCHAR(30)` | NULL     | —                                      | Teléfono                                          |
| `email`                      | `VARCHAR(200)`| NULL     | —                                      | Email de contacto                                 |
| `country_code`               | `CHAR(2)`     | NOT NULL | `'PE'`                                 | ISO 3166-1 alpha-2                                |
| `functional_currency_code`   | `CHAR(3)`     | NOT NULL | `'PEN'`                                | ISO 4217 de la moneda funcional                   |
| `timezone`                   | `VARCHAR(50)` | NOT NULL | `'America/Lima'`                       | IANA timezone                                     |
| `fiscal_year_start_month`    | `SMALLINT`    | NOT NULL | `1`                                    | Mes de inicio del año fiscal (1-12)                |
| `is_active`                  | `BOOLEAN`     | NOT NULL | `TRUE`                                 |                                                   |
| `created_at`                 | `TIMESTAMPTZ` | NOT NULL | `NOW()`                                | Audit                                            |
| `updated_at`                 | `TIMESTAMPTZ` | NOT NULL | `NOW()`                                | Audit (mantenido por trigger)                      |
| `deleted_at`                 | `TIMESTAMPTZ` | NULL     | —                                      | Soft delete                                       |
| `created_by`                 | `UUID`        | NULL     | —                                      | FK → `users(id)` (SET NULL)                       |
| `updated_by`                 | `UUID`        | NULL     | —                                      | FK → `users(id)` (SET NULL)                       |

**PK:** `id`.

**FKs salientes:**
- `created_by`, `updated_by` → `users(id)` ON DELETE SET NULL (agregadas en `0008_add_audit_user_fks.up.sql`).

**Constraints:**
- `ck_companies_fiscal_year_start_month` — `BETWEEN 1 AND 12`.
- `ck_companies_legal_name_nonblank` — `length(trim(legal_name)) > 0`.
- `ck_companies_trade_name_nonblank_or_null` — `trade_name IS NULL OR length(trim(trade_name)) > 0`.

**Índices:**
- `companies_pkey` — PK (`id`).
- `uq_companies_code` — UNIQUE (`code`).
- `uq_companies_tax_id` — UNIQUE (`tax_id`).
- `idx_companies_active` — partial (`is_active`) WHERE `deleted_at IS NULL`.

**Triggers:**
- `trg_companies_set_updated_at` — BEFORE UPDATE → actualiza `updated_at = NOW()`.

**Audit:** columnas embebidas + soft delete vía `deleted_at`.

**Relaciones inversas:** `branches`, `roles`, `users`, `audit_logs`.

---

## Tabla: `branches`

Sucursales físicas de la empresa.

| Columna          | Tipo          | Null     | Default | Descripción                                  |
|------------------|---------------|----------|---------|----------------------------------------------|
| `id`             | `UUID`        | NOT NULL | `gen_random_uuid()` | PK                              |
| `company_id`     | `UUID`        | NOT NULL | —       | FK → `companies(id)`                          |
| `code`           | `VARCHAR(20)` | NOT NULL | —       | Identificador corto de la sucursal            |
| `name`           | `VARCHAR(200)`| NOT NULL | —       | Nombre visible                                |
| `address`        | `TEXT`        | NULL     | —       |                                              |
| `phone`          | `VARCHAR(30)` | NULL     | —       |                                              |
| `email`          | `VARCHAR(200)`| NULL     | —       |                                              |
| `is_default`     | `BOOLEAN`     | NOT NULL | `FALSE` | Una sola branch default por empresa           |
| `is_active`      | `BOOLEAN`     | NOT NULL | `TRUE`  |                                              |
| `created_at`     | `TIMESTAMPTZ` | NOT NULL | `NOW()` | Audit                                        |
| `updated_at`     | `TIMESTAMPTZ` | NOT NULL | `NOW()` | Audit                                        |
| `deleted_at`     | `TIMESTAMPTZ` | NULL     | —       | Soft delete                                   |
| `created_by`     | `UUID`        | NULL     | —       | FK → `users(id)` (SET NULL)                   |
| `updated_by`     | `UUID`        | NULL     | —       | FK → `users(id)` (SET NULL)                   |

**PK:** `id`.

**FKs salientes:**
- `company_id` → `companies(id)` ON DELETE RESTRICT.
- `created_by`, `updated_by` → `users(id)` ON DELETE SET NULL.

**Índices:**
- `branches_pkey` — PK.
- `uq_branches_company_code` — UNIQUE (`company_id`, `code`) WHERE `deleted_at IS NULL`.
- `uq_branches_company_default` — UNIQUE (`company_id`) WHERE `is_default = TRUE AND deleted_at IS NULL`.
- `idx_branches_company_active` — (`company_id`, `is_active`) WHERE `deleted_at IS NULL`.

**Triggers:**
- `trg_branches_set_updated_at` — BEFORE UPDATE.

**Audit:** embebido + soft delete.

**Relaciones inversas:** `users.default_branch_id`, `user_roles.branch_id`.

---

## Tabla: `permissions`

Catálogo global de permisos. **Sin** `company_id` (es universal).

| Columna      | Tipo           | Null     | Default | Descripción                              |
|--------------|----------------|----------|---------|------------------------------------------|
| `code`       | `VARCHAR(100)` | NOT NULL | —       | PK. Formato `module.action`              |
| `module`     | `VARCHAR(50)`  | NOT NULL | —       | `customers`, `sales`, etc.               |
| `action`     | `VARCHAR(50)`  | NOT NULL | —       | `view`, `create`, `delete`, etc.         |
| `description`| `TEXT`         | NULL     | —       | Descripción legible para admin            |
| `created_at` | `TIMESTAMPTZ`  | NOT NULL | `NOW()` |                                          |

**PK:** `code`.

**Sin FKs salientes.** No tiene `updated_at`, `deleted_at`, `is_active`, `is_system` — es un catálogo controlado, no un recurso de negocio por tenant.

**Índices:**
- `permissions_pkey` — PK.
- `idx_permissions_module` — (`module`).

**Sin triggers.**

**Sin soft delete.**

**Relaciones inversas:** `role_permissions.permission_code`.

---

## Tabla: `roles`

Rol de usuario, scoped a una empresa. Cada empresa trae 6 system roles seedeados (admin, manager, accountant, seller, warehouse, viewer).

| Columna      | Tipo           | Null     | Default | Descripción                              |
|--------------|----------------|----------|---------|------------------------------------------|
| `id`         | `UUID`         | NOT NULL | `gen_random_uuid()` | PK                          |
| `company_id` | `UUID`         | NOT NULL | —       | FK → `companies(id)`                      |
| `code`       | `VARCHAR(50)`  | NOT NULL | —       | `admin`, `manager`, etc.                  |
| `name`       | `VARCHAR(100)` | NOT NULL | —       | Nombre visible                            |
| `description`| `TEXT`         | NULL     | —       |                                          |
| `is_system`  | `BOOLEAN`      | NOT NULL | `FALSE` | Roles seedeados; no se eliminan/renombran |
| `is_active`  | `BOOLEAN`      | NOT NULL | `TRUE`  |                                          |
| `created_at` | `TIMESTAMPTZ`  | NOT NULL | `NOW()` |                                          |
| `updated_at` | `TIMESTAMPTZ`  | NOT NULL | `NOW()` |                                          |
| `deleted_at` | `TIMESTAMPTZ`  | NULL     | —       | Soft delete                              |
| `created_by` | `UUID`         | NULL     | —       | FK → `users(id)`                          |
| `updated_by` | `UUID`         | NULL     | —       | FK → `users(id)`                          |

**PK:** `id`.

**FKs salientes:**
- `company_id` → `companies(id)` ON DELETE RESTRICT.
- `created_by`, `updated_by` → `users(id)` ON DELETE SET NULL.

**Constraints:**
- `ck_roles_name_nonblank` — `length(trim(name)) > 0`.

**Índices:**
- `roles_pkey` — PK.
- `uq_roles_company_code` — UNIQUE (`company_id`, `code`) WHERE `deleted_at IS NULL`.
- `idx_roles_company_active` — (`company_id`, `is_active`) WHERE `deleted_at IS NULL`.

**Triggers:**
- `trg_roles_set_updated_at`.

**Audit:** embebido + soft delete.

**Relaciones inversas:** `role_permissions`, `user_roles`.

---

## Tabla: `role_permissions`

Junction role × permission. PK compuesta. Sin `updated_at`/`deleted_at` (es pura linkage).

| Columna          | Tipo           | Null     | Default | Descripción                  |
|------------------|----------------|----------|---------|------------------------------|
| `role_id`        | `UUID`         | NOT NULL | —       | PK + FK → `roles(id)`        |
| `permission_code`| `VARCHAR(100)` | NOT NULL | —       | PK + FK → `permissions(code)`|
| `created_at`     | `TIMESTAMPTZ`  | NOT NULL | `NOW()` |                              |
| `created_by`     | `UUID`         | NULL     | —       | FK → `users(id)` (SET NULL)  |

**PK:** (`role_id`, `permission_code`).

**FKs salientes:**
- `role_id` → `roles(id)` ON DELETE CASCADE.
- `permission_code` → `permissions(code)` ON DELETE CASCADE.
- `created_by` → `users(id)` ON DELETE SET NULL.

**Índices:**
- `role_permissions_pkey` — PK.
- `idx_role_permissions_permission` — (`permission_code`) para resolver "qué roles tienen X permiso" sin escanear la PK.

**Sin triggers.**

**Sin audit / soft delete.**

---

## Tabla: `users`

Cuentas de acceso. Las contraseñas se almacenan como hashes Argon2id.

| Columna                  | Tipo           | Null     | Default                | Descripción                                       |
|--------------------------|----------------|----------|------------------------|---------------------------------------------------|
| `id`                     | `UUID`         | NOT NULL | `gen_random_uuid()`    | PK                                                |
| `company_id`             | `UUID`         | NOT NULL | —                      | FK → `companies(id)`                              |
| `default_branch_id`      | `UUID`         | NULL     | —                      | FK → `branches(id)`                               |
| `username`               | `VARCHAR(100)` | NOT NULL | —                      | Lowercase, alfanumérico + `._-`                   |
| `email`                  | `VARCHAR(200)` | NOT NULL | —                      | Lowercase                                         |
| `full_name`              | `VARCHAR(200)` | NOT NULL | —                      | No blanco                                         |
| `password_hash`          | `TEXT`         | NOT NULL | —                      | Argon2id format `$argon2id$v=19$m=...,t=...,p=...$...` |
| `must_change_password`   | `BOOLEAN`      | NOT NULL | `TRUE`                 | Forzar reset en próximo login                      |
| `failed_login_attempts`  | `INTEGER`      | NOT NULL | `0`                    | Contador de fallos                                 |
| `locked_until`           | `TIMESTAMPTZ`  | NULL     | —                      | NULL = no bloqueado                                 |
| `last_login_at`          | `TIMESTAMPTZ`  | NULL     | —                      |                                                   |
| `last_login_ip`          | `INET`         | NULL     | —                      |                                                   |
| `is_active`              | `BOOLEAN`      | NOT NULL | `TRUE`                 |                                                   |
| `created_at`             | `TIMESTAMPTZ`  | NOT NULL | `NOW()`                |                                                   |
| `updated_at`             | `TIMESTAMPTZ`  | NOT NULL | `NOW()`                |                                                   |
| `deleted_at`             | `TIMESTAMPTZ`  | NULL     | —                      | Soft delete                                       |
| `created_by`             | `UUID`         | NULL     | —                      | FK → `users(id)`                                  |
| `updated_by`             | `UUID`         | NULL     | —                      | FK → `users(id)`                                  |

**PK:** `id`.

**FKs salientes:**
- `company_id` → `companies(id)` ON DELETE RESTRICT.
- `default_branch_id` → `branches(id)` ON DELETE SET NULL.
- `created_by`, `updated_by` → `users(id)` ON DELETE SET NULL.

**Constraints:**
- `ck_users_username_lowercase` — `username = LOWER(username) AND username ~ '^[a-z0-9._-]+$'`.
- `ck_users_email_lowercase` — `email = LOWER(email) AND email LIKE '%@%'`.
- `ck_users_full_name_nonblank` — `length(trim(full_name)) > 0`.
- `ck_users_failed_attempts_cap` — `failed_login_attempts <= 1000`.
- `ck_users_lockout_window` — `locked_until IS NULL OR locked_until > created_at`.

**Índices:**
- `users_pkey` — PK.
- `uq_users_company_username` — UNIQUE (`company_id`, `username`) WHERE `deleted_at IS NULL`.
- `uq_users_company_email` — UNIQUE (`company_id`, `email`) WHERE `deleted_at IS NULL`.
- `idx_users_company_active` — (`company_id`, `is_active`) WHERE `deleted_at IS NULL`.
- `idx_users_locked` — (`locked_until`) WHERE `locked_until IS NOT NULL`.

**Triggers:**
- `trg_users_set_updated_at` — BEFORE UPDATE.

**Audit:** embebido + soft delete.

**Relaciones inversas:** `user_roles`, `login_history`, `audit_logs`, y todas las FKs de `created_by`/`updated_by` en otras tablas.

---

## Tabla: `user_roles`

Junction user × role, con scope opcional por branch. **Surrogate `id` + dos partial unique indexes** para soportar company-wide grants y branch-scoped grants simultáneamente.

| Columna      | Tipo          | Null     | Default             | Descripción                                          |
|--------------|---------------|----------|---------------------|------------------------------------------------------|
| `id`         | `UUID`        | NOT NULL | `gen_random_uuid()` | PK surrogate                                         |
| `user_id`    | `UUID`        | NOT NULL | —                   | FK → `users(id)`                                     |
| `role_id`    | `UUID`        | NOT NULL | —                   | FK → `roles(id)`                                     |
| `branch_id`  | `UUID`        | NULL     | —                   | FK → `branches(id)`. NULL = company-wide             |
| `assigned_at`| `TIMESTAMPTZ` | NOT NULL | `NOW()`             |                                                      |
| `assigned_by`| `UUID`        | NULL     | —                   | FK → `users(id)`                                     |
| `expires_at` | `TIMESTAMPTZ` | NULL     | —                   | Grant con vencimiento                                |

**PK:** `id`.

**FKs salientes:**
- `user_id` → `users(id)` ON DELETE CASCADE.
- `role_id` → `roles(id)` ON DELETE CASCADE.
- `branch_id` → `branches(id)` ON DELETE CASCADE.
- `assigned_by` → `users(id)` ON DELETE SET NULL.

**Constraints:**
- `ck_user_roles_expires_after_assigned` — `expires_at IS NULL OR expires_at > assigned_at`.

**Índices:**
- `user_roles_pkey` — PK (`id`).
- `uq_user_roles_per_branch` — UNIQUE (`user_id`, `role_id`, `branch_id`) WHERE `branch_id IS NOT NULL`. Evita el mismo role dos veces en la misma branch.
- `uq_user_roles_company_wide` — UNIQUE (`user_id`, `role_id`) WHERE `branch_id IS NULL`. Evita el mismo role dos veces como company-wide.
- `idx_user_roles_user_active` — (`user_id`) WHERE `expires_at IS NULL` (grants activos por usuario).
- `idx_user_roles_role` — (`role_id`).

**Sin triggers.**

**Sin soft delete.**

**Nota de diseño:** Se descartaron dos alternativas:
- PK compuesta `(user_id, role_id, branch_id)` con `branch_id` nullable — Postgres trata columnas de PK como `NOT NULL` por lo que esto no funcionaba.
- Columna `scope_key` generada con `COALESCE(branch_id, '<zero-uuid>')` — agrega una columna redundante y un trigger WHEN para mantenerla en sync.

La solución actual (dos partial unique indexes) es la más limpia y aprovecha nativamente los partial indexes de PostgreSQL.

---

## Tabla: `login_history`

Log append-only de cada intento de autenticación.

| Columna              | Tipo           | Null     | Default | Descripción                                |
|----------------------|----------------|----------|---------|--------------------------------------------|
| `id`                 | `UUID`         | NOT NULL | `gen_random_uuid()` | PK                          |
| `user_id`            | `UUID`         | NULL     | —       | FK → `users(id)` (SET NULL)                |
| `username_attempted` | `VARCHAR(100)` | NOT NULL | —       | Username provisto (puede no existir)       |
| `success`            | `BOOLEAN`      | NOT NULL | —       | ¿Login exitoso?                            |
| `failure_reason`     | `VARCHAR(100)` | NULL     | —       | Razón del fallo (solo si `success = FALSE`) |
| `ip_address`         | `INET`         | NULL     | —       |                                            |
| `user_agent`         | `TEXT`         | NULL     | —       |                                            |
| `occurred_at`        | `TIMESTAMPTZ`  | NOT NULL | `NOW()` |                                            |

**PK:** `id`.

**FKs salientes:**
- `user_id` → `users(id)` ON DELETE SET NULL (preserva el log aunque el user sea eliminado).

**Constraints:**
- `ck_login_history_failure_consistency` — `(success = TRUE AND failure_reason IS NULL) OR (success = FALSE AND failure_reason IS NOT NULL)`.

**Índices:**
- `login_history_pkey` — PK.
- `idx_login_history_user_time` — (`user_id`, `occurred_at` DESC) para "actividad reciente por usuario".
- `idx_login_history_ip_time` — (`ip_address`, `occurred_at` DESC) WHERE `ip_address IS NOT NULL` para detección de abuso.
- `idx_login_history_time` — (`occurred_at` DESC) para timeline global.
- `idx_login_history_failures` — (`occurred_at` DESC) WHERE `success = FALSE` para incident review.

**Sin triggers.**

**Sin soft delete.** Append-only.

---

## Tabla: `audit_logs`

Log global inmutable de mutaciones. Diseño compatible con particionamiento futuro por `occurred_at`.

| Columna         | Tipo           | Null     | Default             | Descripción                                       |
|-----------------|----------------|----------|---------------------|---------------------------------------------------|
| `id`            | `UUID`         | NOT NULL | `gen_random_uuid()` | Parte de PK con `occurred_at` (preparado para RANGE partitioning) |
| `occurred_at`   | `TIMESTAMPTZ`  | NOT NULL | `NOW()`             | Parte de PK; futuro partition key                  |
| `company_id`    | `UUID`         | NOT NULL | —                   | FK → `companies(id)`                              |
| `user_id`       | `UUID`         | NULL     | —                   | FK → `users(id)` (SET NULL)                       |
| `table_name`    | `VARCHAR(100)` | NOT NULL | —                   | Tabla afectada                                    |
| `record_id`     | `UUID`         | NULL     | —                   | PK del registro afectado                          |
| `action`        | `VARCHAR(30)`  | NOT NULL | —                   | Ver CHECK abajo                                   |
| `old_value`     | `JSONB`        | NULL     | —                   | Snapshot pre-mutación                              |
| `new_value`     | `JSONB`        | NULL     | —                   | Snapshot post-mutación                             |
| `changed_fields`| `TEXT[]`       | NULL     | —                   | Campos cambiados (opcional, optimización)         |
| `ip_address`    | `INET`         | NULL     | —                   |                                                   |
| `user_agent`    | `TEXT`         | NULL     | —                   |                                                   |
| `device`        | `VARCHAR(200)` | NULL     | —                   | Descripción libre del dispositivo                  |

**PK:** (`id`, `occurred_at`).

**FKs salientes:**
- `company_id` → `companies(id)` ON DELETE RESTRICT.
- `user_id` → `users(id)` ON DELETE SET NULL.

**Constraints:**
- `ck_audit_logs_action` — `action IN ('INSERT', 'UPDATE', 'DELETE', 'HARD_DELETE', 'LOGIN', 'LOGOUT', 'LOGIN_FAILED', 'APPROVE', 'REJECT', 'CANCEL', 'CONCILIATE', 'CLOSE_PERIOD', 'REOPEN_PERIOD', 'EXPORT', 'PRINT', 'SEND')`.

**Índices:**
- `audit_logs_pkey` — PK.
- `idx_audit_logs_company_time` — (`company_id`, `occurred_at` DESC) — query principal.
- `idx_audit_logs_record` — (`table_name`, `record_id`, `occurred_at` DESC) WHERE `record_id IS NOT NULL` — "historial de un registro".
- `idx_audit_logs_user_time` — (`user_id`, `occurred_at` DESC) WHERE `user_id IS NOT NULL` — auditoría por usuario.
- `idx_audit_logs_action_time` — (`action`, `occurred_at` DESC).
- `idx_audit_logs_company_action_time` — (`company_id`, `action`, `occurred_at` DESC).

**Triggers (append-only enforcement):**
- `trg_audit_logs_no_update` — BEFORE UPDATE → `RAISE EXCEPTION 'audit_logs is append-only; UPDATE is not allowed'`.
- `trg_audit_logs_no_delete` — BEFORE DELETE → `RAISE EXCEPTION 'audit_logs is append-only; DELETE is not allowed'`.

**Sin soft delete.**

**Sin updated_at** (no se permite UPDATE).

**Particionamiento futuro:** una migración posterior hará `ALTER TABLE audit_logs ... PARTITION BY RANGE (occurred_at);` y creará particiones mensuales. El PK actual `(id, occurred_at)` ya incluye la partition key, lo cual es requerido para tablas particionadas.

---

# Module 1.1+ (pendiente)

Los siguientes módulos aún no tienen SQL escrito. Ver `DATABASE_ARCHITECTURE.md` § 7 (Entity Catalog) para el diseño previsto.

| Módulo           | Tablas                                                                                  |
|------------------|-----------------------------------------------------------------------------------------|
| Master Data      | `currencies`, `countries`, `taxes`, `tax_rates`, `tax_categories`, `document_types`, `units_of_measure`, `customers`, `suppliers`, `product_categories`, `product_brands`, `products`, `product_images`, `product_barcodes`, `product_suppliers`, `price_lists`, `price_list_items`, `warehouses` |
| Inventory        | `inventory_batches`, `inventory_movements`, `inventory_stock`, `inventory_transfers`, `inventory_adjustments` |
| Purchasing       | `purchases`, `purchase_lines`, `supplier_payments`, `supplier_payment_allocations`, `purchase_returns`, `purchase_return_lines`, `supplier_credits` |
| Sales            | `sales`, `sale_lines`, `customer_payments`, `customer_payment_allocations`, `customer_advances`, `customer_advance_applications`, `sales_returns`, `sale_return_lines`, `customer_credits` |
| Treasury         | `bank_accounts`, `bank_transactions`, `bank_reconciliations`, `credit_cards`, `credit_card_transactions`, `cash_registers`, `cash_movements`, `exchange_rates` |
| Accounting       | `chart_of_accounts`, `fiscal_years`, `fiscal_periods`, `journal_entries`, `journal_entry_lines`, `account_balances` |
| Cross-cutting     | `attachments`, `notifications`, `notification_recipients`, `document_sequences` |

---

## Apéndice A: Catálogo de triggers de Module 1

| Tabla             | Trigger                          | Timing            | Función                       |
|-------------------|----------------------------------|-------------------|--------------------------------|
| `companies`       | `trg_companies_set_updated_at`   | BEFORE UPDATE     | `set_updated_at()`             |
| `branches`        | `trg_branches_set_updated_at`    | BEFORE UPDATE     | `set_updated_at()`             |
| `roles`           | `trg_roles_set_updated_at`       | BEFORE UPDATE     | `set_updated_at()`             |
| `users`           | `trg_users_set_updated_at`       | BEFORE UPDATE     | `set_updated_at()`             |
| `audit_logs`      | `trg_audit_logs_no_update`       | BEFORE UPDATE     | `audit_logs_forbid_mutation()` |
| `audit_logs`      | `trg_audit_logs_no_delete`       | BEFORE DELETE     | `audit_logs_forbid_mutation()` |

`role_permissions`, `user_roles`, `login_history`, `permissions` no tienen triggers (tablas sin `updated_at` o con semántica append-only).

---

## Apéndice B: Secuencia de aplicación

```
0000_init.up.sql
0001_create_companies.up.sql
0002_create_branches.up.sql
0003_create_permissions.up.sql
0004_create_roles.up.sql
0005_create_role_permissions.up.sql
0006_create_users.up.sql
0007_create_user_roles.up.sql
0008_add_audit_user_fks.up.sql          ← agrega FKs de created_by/updated_by
0009_create_login_history.up.sql
0010_create_audit_logs.up.sql
0011_seed_auth.up.sql                  ← seed: 1 empresa, 1 branch, 6 roles, 65 permissions, 1 admin
```

Cada archivo es **inmutable** una vez commiteado. Cambios futuros al esquema requieren un nuevo archivo `NNNN_*.up.sql` con la siguiente versión disponible.
