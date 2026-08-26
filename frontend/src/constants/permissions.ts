export const Permissions = {
  Customers: {
    View: 'customers.view',
    Create: 'customers.create',
    Edit: 'customers.edit',
    Delete: 'customers.delete',
    Export: 'customers.export',
    Import: 'customers.import',
    Print: 'customers.print',
  },
  Suppliers: {
    View: 'suppliers.view',
    Create: 'suppliers.create',
    Edit: 'suppliers.edit',
    Delete: 'suppliers.delete',
    Export: 'suppliers.export',
    Import: 'suppliers.import',
  },
  Products: {
    View: 'products.view',
    Create: 'products.create',
    Edit: 'products.edit',
    Delete: 'products.delete',
    Export: 'products.export',
    Import: 'products.import',
  },
  Inventory: {
    View: 'inventory.view',
    Create: 'inventory.create',
    Edit: 'inventory.edit',
    Delete: 'inventory.delete',
    Transfer: 'inventory.transfer',
    Adjust: 'inventory.adjust',
    Export: 'inventory.export',
  },
  Purchases: {
    View: 'purchases.view',
    Create: 'purchases.create',
    Edit: 'purchases.edit',
    Delete: 'purchases.delete',
    Approve: 'purchases.approve',
    Cancel: 'purchases.cancel',
    Export: 'purchases.export',
  },
  Sales: {
    View: 'sales.view',
    Create: 'sales.create',
    Edit: 'sales.edit',
    Delete: 'sales.delete',
    Approve: 'sales.approve',
    Cancel: 'sales.cancel',
    Export: 'sales.export',
    Print: 'sales.print',
  },
  Treasury: {
    View: 'treasury.view',
    Create: 'treasury.create',
    Edit: 'treasury.edit',
    Delete: 'treasury.delete',
    Conciliate: 'treasury.conciliate',
    Close: 'treasury.close',
    Export: 'treasury.export',
  },
  Settings: {
    View: 'settings.view',
    Edit: 'settings.edit',
  },
  Administration: {
    View: 'administration.view',
    ManageUsers: 'administration.users.manage',
    ManageRoles: 'administration.roles.manage',
    ManagePermissions: 'administration.permissions.manage',
    ViewAudit: 'administration.audit.view',
  },
} as const;

export type PermissionKey =
  | `Customers.${keyof typeof Permissions.Customers}`
  | `Suppliers.${keyof typeof Permissions.Suppliers}`
  | `Products.${keyof typeof Permissions.Products}`
  | `Inventory.${keyof typeof Permissions.Inventory}`
  | `Purchases.${keyof typeof Permissions.Purchases}`
  | `Sales.${keyof typeof Permissions.Sales}`
  | `Treasury.${keyof typeof Permissions.Treasury}`
  | `Settings.${keyof typeof Permissions.Settings}`
  | `Administration.${keyof typeof Permissions.Administration}`;

export const Roles = {
  Admin: 'admin',
  Manager: 'manager',
  Accountant: 'accountant',
  Seller: 'seller',
  WarehouseOperator: 'warehouse',
  Viewer: 'viewer',
} as const;

export type Role = (typeof Roles)[keyof typeof Roles];

export const RolePermissions: Record<Role, string[]> = {
  admin: Object.values(Permissions).flatMap((m) => Object.values(m)),
  manager: [
    Permissions.Customers.View,
    Permissions.Customers.Create,
    Permissions.Customers.Edit,
    Permissions.Customers.Export,
    Permissions.Suppliers.View,
    Permissions.Suppliers.Create,
    Permissions.Suppliers.Edit,
    Permissions.Products.View,
    Permissions.Products.Create,
    Permissions.Products.Edit,
    Permissions.Inventory.View,
    Permissions.Purchases.View,
    Permissions.Purchases.Approve,
    Permissions.Purchases.Cancel,
    Permissions.Sales.View,
    Permissions.Sales.Approve,
    Permissions.Sales.Cancel,
  ],
  accountant: [
    Permissions.Customers.View,
    Permissions.Suppliers.View,
    Permissions.Products.View,
    Permissions.Inventory.View,
    Permissions.Purchases.View,
    Permissions.Sales.View,
    Permissions.Treasury.View,
    Permissions.Treasury.Conciliate,
  ],
  seller: [
    Permissions.Customers.View,
    Permissions.Customers.Create,
    Permissions.Customers.Edit,
    Permissions.Products.View,
    Permissions.Sales.View,
    Permissions.Sales.Create,
    Permissions.Sales.Edit,
  ],
  warehouse: [
    Permissions.Products.View,
    Permissions.Inventory.View,
    Permissions.Inventory.Transfer,
    Permissions.Inventory.Adjust,
  ],
  viewer: [
    Permissions.Customers.View,
    Permissions.Suppliers.View,
    Permissions.Products.View,
    Permissions.Inventory.View,
    Permissions.Purchases.View,
    Permissions.Sales.View,
  ],
};

export function getRolePermissions(role: Role): string[] {
  return RolePermissions[role] ?? [];
}
