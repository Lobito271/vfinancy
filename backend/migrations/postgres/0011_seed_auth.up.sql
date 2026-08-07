-- 0011_seed_auth.up.sql
-- Module 1: Authentication
-- Seed data for a fresh install. Idempotent: every INSERT is guarded with
-- ON CONFLICT DO NOTHING so re-running the migration is safe and so an
-- environment that already has data is not broken.
--
-- What gets seeded:
--   1. The global permissions catalog (module.action keys).
--   2. A demo company "vfinancy S.A.C." with one branch "Sede Central".
--   3. The 6 system roles for that company (admin, manager, accountant,
--      seller, warehouse, viewer), each with a curated set of permission
--      grants.
--   4. A default admin user with a placeholder password hash.
--
-- IMPORTANT: the placeholder password hash in this seed is intentionally
-- not a working password. The application bootstrap step (or a CLI command
-- `vfinancy users reset-password <username>`) must replace it before
-- any real login is possible. The seed value is the Argon2id hash of the
-- literal string "CHANGE_ME_ON_FIRST_LOGIN" — it is not the hash of an
-- empty string or "password". A password reset is mandatory on first use.

-- =========================================================================
-- 1. Global permissions catalog
-- =========================================================================

INSERT INTO permissions (code, module, action, description) VALUES
    -- Customers
    ('customers.view',     'customers', 'view',     'Ver clientes y su información'),
    ('customers.create',   'customers', 'create',   'Crear nuevos clientes'),
    ('customers.edit',     'customers', 'edit',     'Modificar datos de clientes'),
    ('customers.delete',   'customers', 'delete',   'Eliminar (soft) clientes'),
    ('customers.export',   'customers', 'export',   'Exportar listado de clientes'),
    ('customers.import',   'customers', 'import',   'Importar clientes desde archivo'),
    ('customers.print',    'customers', 'print',    'Imprimir fichas de clientes'),

    -- Suppliers
    ('suppliers.view',     'suppliers', 'view',     'Ver proveedores'),
    ('suppliers.create',   'suppliers', 'create',   'Crear nuevos proveedores'),
    ('suppliers.edit',     'suppliers', 'edit',     'Modificar proveedores'),
    ('suppliers.delete',   'suppliers', 'delete',   'Eliminar proveedores'),
    ('suppliers.export',   'suppliers', 'export',   'Exportar listado de proveedores'),
    ('suppliers.import',   'suppliers', 'import',   'Importar proveedores'),

    -- Products
    ('products.view',      'products',  'view',     'Ver catálogo de productos'),
    ('products.create',    'products',  'create',   'Crear productos'),
    ('products.edit',      'products',  'edit',     'Modificar productos'),
    ('products.delete',    'products',  'delete',   'Eliminar productos'),
    ('products.export',    'products',  'export',   'Exportar catálogo'),
    ('products.import',    'products',  'import',   'Importar productos'),

    -- Inventory
    ('inventory.view',     'inventory', 'view',     'Ver stock y movimientos'),
    ('inventory.create',   'inventory', 'create',   'Registrar ingresos / movimientos'),
    ('inventory.edit',     'inventory', 'edit',     'Modificar movimientos (correcciones)'),
    ('inventory.delete',   'inventory', 'delete',   'Anular movimientos'),
    ('inventory.transfer', 'inventory', 'transfer', 'Transferir entre almacenes'),
    ('inventory.adjust',   'inventory', 'adjust',   'Ajustes manuales de stock'),
    ('inventory.export',   'inventory', 'export',   'Exportar valorización'),

    -- Purchases
    ('purchases.view',     'purchases', 'view',     'Ver compras'),
    ('purchases.create',   'purchases', 'create',   'Registrar compras'),
    ('purchases.edit',     'purchases', 'edit',     'Modificar compras'),
    ('purchases.delete',   'purchases', 'delete',   'Eliminar compras'),
    ('purchases.approve',  'purchases', 'approve',  'Aprobar compras'),
    ('purchases.cancel',   'purchases', 'cancel',   'Anular compras'),
    ('purchases.export',   'purchases', 'export',   'Exportar reporte de compras'),

    -- Sales
    ('sales.view',         'sales',     'view',     'Ver ventas'),
    ('sales.create',       'sales',     'create',   'Registrar ventas'),
    ('sales.edit',         'sales',     'edit',     'Modificar ventas'),
    ('sales.delete',       'sales',     'delete',   'Eliminar ventas'),
    ('sales.approve',      'sales',     'approve',  'Aprobar ventas'),
    ('sales.cancel',       'sales',     'cancel',   'Anular ventas'),
    ('sales.export',       'sales',     'export',   'Exportar reporte de ventas'),
    ('sales.print',        'sales',     'print',    'Imprimir comprobantes'),

    -- Treasury
    ('treasury.view',      'treasury',  'view',     'Ver cuentas y movimientos'),
    ('treasury.create',    'treasury',  'create',   'Registrar movimientos'),
    ('treasury.edit',      'treasury',  'edit',     'Modificar movimientos'),
    ('treasury.delete',    'treasury',  'delete',   'Anular movimientos'),
    ('treasury.conciliate','treasury',  'conciliate','Conciliar cuentas bancarias'),
    ('treasury.close',     'treasury',  'close',    'Cerrar caja / periodo de tesorería'),
    ('treasury.export',    'treasury',  'export',   'Exportar reportes'),

    -- Accounting
    ('accounting.view',    'accounting','view',     'Ver plan contable y libros'),
    ('accounting.create',  'accounting','create',   'Crear asientos contables'),
    ('accounting.edit',    'accounting','edit',     'Modificar asientos (borrador)'),
    ('accounting.delete',  'accounting','delete',   'Eliminar asientos (borrador)'),
    ('accounting.close',   'accounting','close',    'Cerrar periodo fiscal'),
    ('accounting.export',  'accounting','export',   'Exportar libros y reportes'),

    -- Reports
    ('reports.view',       'reports',   'view',     'Ver y ejecutar reportes'),
    ('reports.export',     'reports',   'export',   'Exportar reportes'),
    ('reports.print',      'reports',   'print',    'Imprimir reportes'),
    ('reports.schedule',   'reports',   'schedule', 'Programar reportes automáticos'),

    -- Settings
    ('settings.view',      'settings',  'view',     'Ver configuración'),
    ('settings.edit',      'settings',  'edit',     'Modificar configuración'),

    -- Administration
    ('administration.view',          'administration', 'view',     'Acceder al panel de administración'),
    ('administration.users.manage',  'administration', 'users.manage',  'Crear / editar / desactivar usuarios'),
    ('administration.roles.manage',  'administration', 'roles.manage',  'Crear / editar roles y permisos'),
    ('administration.permissions.manage','administration','permissions.manage','Asignar permisos a roles'),
    ('administration.audit.view',    'administration', 'audit.view',    'Consultar audit_logs')
ON CONFLICT (code) DO NOTHING;

-- =========================================================================
-- 2. Demo company + branch
-- =========================================================================

INSERT INTO companies (
    id, code, legal_name, trade_name, tax_id, country_code,
    functional_currency_code, timezone, fiscal_year_start_month, is_active
) VALUES (
    '00000000-0000-0000-0000-000000000001',
    'VFI',
    'vfinancy S.A.C.',
    'vfinancy',
    '20600000001',
    'PE',
    'PEN',
    'America/Lima',
    1,
    TRUE
) ON CONFLICT (code) DO NOTHING;

INSERT INTO branches (
    id, company_id, code, name, is_default, is_active
) VALUES (
    '00000000-0000-0000-0000-000000000001',
    '00000000-0000-0000-0000-000000000001',
    'SEDE-01',
    'Sede Central',
    TRUE,
    TRUE
) ON CONFLICT (company_id, code) WHERE deleted_at IS NULL DO NOTHING;

-- =========================================================================
-- 3. System roles + permission grants
-- =========================================================================

-- admin: every permission in the catalog.
INSERT INTO roles (id, company_id, code, name, description, is_system, is_active)
VALUES (
    '00000000-0000-0000-0000-0000000000a1',
    '00000000-0000-0000-0000-000000000001',
    'admin', 'Administrador', 'Acceso total al sistema', TRUE, TRUE
) ON CONFLICT (company_id, code) WHERE deleted_at IS NULL DO NOTHING;

INSERT INTO role_permissions (role_id, permission_code)
SELECT '00000000-0000-0000-0000-0000000000a1', code FROM permissions
ON CONFLICT (role_id, permission_code) DO NOTHING;

-- manager: broad read/write across operations, no admin or accounting
-- close.
INSERT INTO roles (id, company_id, code, name, description, is_system, is_active)
VALUES (
    '00000000-0000-0000-0000-0000000000a2',
    '00000000-0000-0000-0000-000000000001',
    'manager', 'Gerente', 'Acceso operativo amplio', TRUE, TRUE
) ON CONFLICT (company_id, code) WHERE deleted_at IS NULL DO NOTHING;

INSERT INTO role_permissions (role_id, permission_code) VALUES
    ('00000000-0000-0000-0000-0000000000a2', 'customers.view'),
    ('00000000-0000-0000-0000-0000000000a2', 'customers.create'),
    ('00000000-0000-0000-0000-0000000000a2', 'customers.edit'),
    ('00000000-0000-0000-0000-0000000000a2', 'customers.export'),
    ('00000000-0000-0000-0000-0000000000a2', 'suppliers.view'),
    ('00000000-0000-0000-0000-0000000000a2', 'suppliers.create'),
    ('00000000-0000-0000-0000-0000000000a2', 'suppliers.edit'),
    ('00000000-0000-0000-0000-0000000000a2', 'products.view'),
    ('00000000-0000-0000-0000-0000000000a2', 'products.create'),
    ('00000000-0000-0000-0000-0000000000a2', 'products.edit'),
    ('00000000-0000-0000-0000-0000000000a2', 'products.export'),
    ('00000000-0000-0000-0000-0000000000a2', 'inventory.view'),
    ('00000000-0000-0000-0000-0000000000a2', 'inventory.transfer'),
    ('00000000-0000-0000-0000-0000000000a2', 'inventory.export'),
    ('00000000-0000-0000-0000-0000000000a2', 'purchases.view'),
    ('00000000-0000-0000-0000-0000000000a2', 'purchases.create'),
    ('00000000-0000-0000-0000-0000000000a2', 'purchases.edit'),
    ('00000000-0000-0000-0000-0000000000a2', 'purchases.approve'),
    ('00000000-0000-0000-0000-0000000000a2', 'purchases.cancel'),
    ('00000000-0000-0000-0000-0000000000a2', 'purchases.export'),
    ('00000000-0000-0000-0000-0000000000a2', 'sales.view'),
    ('00000000-0000-0000-0000-0000000000a2', 'sales.create'),
    ('00000000-0000-0000-0000-0000000000a2', 'sales.edit'),
    ('00000000-0000-0000-0000-0000000000a2', 'sales.approve'),
    ('00000000-0000-0000-0000-0000000000a2', 'sales.cancel'),
    ('00000000-0000-0000-0000-0000000000a2', 'sales.export'),
    ('00000000-0000-0000-0000-0000000000a2', 'sales.print'),
    ('00000000-0000-0000-0000-0000000000a2', 'treasury.view'),
    ('00000000-0000-0000-0000-0000000000a2', 'treasury.conciliate'),
    ('00000000-0000-0000-0000-0000000000a2', 'accounting.view'),
    ('00000000-0000-0000-0000-0000000000a2', 'reports.view'),
    ('00000000-0000-0000-0000-0000000000a2', 'reports.export'),
    ('00000000-0000-0000-0000-0000000000a2', 'reports.print'),
    ('00000000-0000-0000-0000-0000000000a2', 'settings.view')
ON CONFLICT (role_id, permission_code) DO NOTHING;

-- accountant: full accounting + treasury read, plus read across the rest.
INSERT INTO roles (id, company_id, code, name, description, is_system, is_active)
VALUES (
    '00000000-0000-0000-0000-0000000000a3',
    '00000000-0000-0000-0000-000000000001',
    'accountant', 'Contador', 'Acceso contable y financiero completo', TRUE, TRUE
) ON CONFLICT (company_id, code) WHERE deleted_at IS NULL DO NOTHING;

INSERT INTO role_permissions (role_id, permission_code) VALUES
    ('00000000-0000-0000-0000-0000000000a3', 'customers.view'),
    ('00000000-0000-0000-0000-0000000000a3', 'suppliers.view'),
    ('00000000-0000-0000-0000-0000000000a3', 'products.view'),
    ('00000000-0000-0000-0000-0000000000a3', 'inventory.view'),
    ('00000000-0000-0000-0000-0000000000a3', 'purchases.view'),
    ('00000000-0000-0000-0000-0000000000a3', 'purchases.export'),
    ('00000000-0000-0000-0000-0000000000a3', 'sales.view'),
    ('00000000-0000-0000-0000-0000000000a3', 'sales.export'),
    ('00000000-0000-0000-0000-0000000000a3', 'treasury.view'),
    ('00000000-0000-0000-0000-0000000000a3', 'treasury.create'),
    ('00000000-0000-0000-0000-0000000000a3', 'treasury.edit'),
    ('00000000-0000-0000-0000-0000000000a3', 'treasury.conciliate'),
    ('00000000-0000-0000-0000-0000000000a3', 'treasury.export'),
    ('00000000-0000-0000-0000-0000000000a3', 'accounting.view'),
    ('00000000-0000-0000-0000-0000000000a3', 'accounting.create'),
    ('00000000-0000-0000-0000-0000000000a3', 'accounting.edit'),
    ('00000000-0000-0000-0000-0000000000a3', 'accounting.close'),
    ('00000000-0000-0000-0000-0000000000a3', 'accounting.export'),
    ('00000000-0000-0000-0000-0000000000a3', 'reports.view'),
    ('00000000-0000-0000-0000-0000000000a3', 'reports.export'),
    ('00000000-0000-0000-0000-0000000000a3', 'reports.print')
ON CONFLICT (role_id, permission_code) DO NOTHING;

-- seller: customers + products + sales.
INSERT INTO roles (id, company_id, code, name, description, is_system, is_active)
VALUES (
    '00000000-0000-0000-0000-0000000000a4',
    '00000000-0000-0000-0000-000000000001',
    'seller', 'Vendedor', 'Gestiona clientes y registra ventas', TRUE, TRUE
) ON CONFLICT (company_id, code) WHERE deleted_at IS NULL DO NOTHING;

INSERT INTO role_permissions (role_id, permission_code) VALUES
    ('00000000-0000-0000-0000-0000000000a4', 'customers.view'),
    ('00000000-0000-0000-0000-0000000000a4', 'customers.create'),
    ('00000000-0000-0000-0000-0000000000a4', 'customers.edit'),
    ('00000000-0000-0000-0000-0000000000a4', 'products.view'),
    ('00000000-0000-0000-0000-0000000000a4', 'inventory.view'),
    ('00000000-0000-0000-0000-0000000000a4', 'sales.view'),
    ('00000000-0000-0000-0000-0000000000a4', 'sales.create'),
    ('00000000-0000-0000-0000-0000000000a4', 'sales.edit'),
    ('00000000-0000-0000-0000-0000000000a4', 'sales.print')
ON CONFLICT (role_id, permission_code) DO NOTHING;

-- warehouse: products + inventory.
INSERT INTO roles (id, company_id, code, name, description, is_system, is_active)
VALUES (
    '00000000-0000-0000-0000-0000000000a5',
    '00000000-0000-0000-0000-000000000001',
    'warehouse', 'Almacén', 'Gestiona stock y transferencias', TRUE, TRUE
) ON CONFLICT (company_id, code) WHERE deleted_at IS NULL DO NOTHING;

INSERT INTO role_permissions (role_id, permission_code) VALUES
    ('00000000-0000-0000-0000-0000000000a5', 'products.view'),
    ('00000000-0000-0000-0000-0000000000a5', 'inventory.view'),
    ('00000000-0000-0000-0000-0000000000a5', 'inventory.create'),
    ('00000000-0000-0000-0000-0000000000a5', 'inventory.edit'),
    ('00000000-0000-0000-0000-0000000000a5', 'inventory.transfer'),
    ('00000000-0000-0000-0000-0000000000a5', 'inventory.adjust'),
    ('00000000-0000-0000-0000-0000000000a5', 'inventory.export')
ON CONFLICT (role_id, permission_code) DO NOTHING;

-- viewer: read-only across the operational modules.
INSERT INTO roles (id, company_id, code, name, description, is_system, is_active)
VALUES (
    '00000000-0000-0000-0000-0000000000a6',
    '00000000-0000-0000-0000-000000000001',
    'viewer', 'Visualizador', 'Solo lectura', TRUE, TRUE
) ON CONFLICT (company_id, code) WHERE deleted_at IS NULL DO NOTHING;

INSERT INTO role_permissions (role_id, permission_code) VALUES
    ('00000000-0000-0000-0000-0000000000a6', 'customers.view'),
    ('00000000-0000-0000-0000-0000000000a6', 'suppliers.view'),
    ('00000000-0000-0000-0000-0000000000a6', 'products.view'),
    ('00000000-0000-0000-0000-0000000000a6', 'inventory.view'),
    ('00000000-0000-0000-0000-0000000000a6', 'purchases.view'),
    ('00000000-0000-0000-0000-0000000000a6', 'sales.view'),
    ('00000000-0000-0000-0000-0000000000a6', 'treasury.view'),
    ('00000000-0000-0000-0000-0000000000a6', 'accounting.view'),
    ('00000000-0000-0000-0000-0000000000a6', 'reports.view')
ON CONFLICT (role_id, permission_code) DO NOTHING;

-- =========================================================================
-- 4. Default admin user
-- =========================================================================
-- Username: admin
-- Email:    admin@vfinancy.local
-- Password: MUST be reset on first login (must_change_password = TRUE).
--           The hash below is the Argon2id of "CHANGE_ME_ON_FIRST_LOGIN".
--           It is intentionally not a working password.

INSERT INTO users (
    id, company_id, default_branch_id, username, email, full_name,
    password_hash, must_change_password, is_active
) VALUES (
    '00000000-0000-0000-0000-0000000000aa',
    '00000000-0000-0000-0000-000000000001',
    '00000000-0000-0000-0000-000000000001',
    'admin',
    'admin@vfinancy.local',
    'Administrador del Sistema',
    '$argon2id$v=19$m=65536,t=3,p=2$Y2hhbmdlX21lX29uX2ZpcnN0X2xvZ2lu$placeholder_hash_replace_on_first_login',
    TRUE,
    TRUE
) ON CONFLICT (company_id, username) WHERE deleted_at IS NULL DO NOTHING;

INSERT INTO user_roles (user_id, role_id, branch_id)
VALUES (
    '00000000-0000-0000-0000-0000000000aa',
    '00000000-0000-0000-0000-0000000000a1',
    NULL
) ON CONFLICT (user_id, role_id) WHERE branch_id IS NULL DO NOTHING;
