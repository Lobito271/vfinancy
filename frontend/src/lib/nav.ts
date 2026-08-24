import type { LucideIcon } from 'lucide-react';
import { Icons } from '@/design-system/icons';
import { Routes } from '@/constants/routes';
import { Permissions } from '@/constants/permissions';
import type { Permission } from '@/utils/permissions';

export interface NavRoute {
  to: string;
  label: string;
  icon: LucideIcon;
  permission?: Permission;
  end?: boolean;
}

export const navRoutes: NavRoute[] = [
  { to: Routes.Dashboard, label: 'Inicio', icon: Icons.Navigation.Dashboard, end: true },
  { to: Routes.Customers, label: 'Clientes', icon: Icons.Navigation.Customers, permission: Permissions.Customers.View },
  { to: Routes.Suppliers, label: 'Proveedores', icon: Icons.Navigation.Suppliers, permission: Permissions.Suppliers.View },
  { to: Routes.Products, label: 'Productos', icon: Icons.Navigation.Products, permission: Permissions.Products.View },
  { to: Routes.CatalogSettings, label: 'Catálogo', icon: Icons.Navigation.Catalog, permission: Permissions.Products.View },
  { to: Routes.Inventory, label: 'Inventario', icon: Icons.Navigation.Inventory, permission: Permissions.Inventory.View },
  { to: Routes.Purchases, label: 'Compras', icon: Icons.Navigation.Purchases, permission: Permissions.Purchases.View },
  {
    to: Routes.CustomerOrders,
    label: 'Pedidos de cliente',
    icon: Icons.Navigation.CustomerOrders,
    permission: Permissions.Purchases.View,
  },
  { to: Routes.Sales, label: 'Ventas', icon: Icons.Navigation.Sales, permission: Permissions.Sales.View },
  { to: Routes.Treasury, label: 'Tesorería', icon: Icons.Navigation.Treasury, permission: Permissions.Treasury.View },
  { to: Routes.Reports, label: 'Reportes', icon: Icons.Navigation.Reports, permission: Permissions.Reports.View },
  { to: Routes.Settings, label: 'Configuración', icon: Icons.Navigation.Settings, permission: Permissions.Settings.View },
];

export function findRouteLabel(path: string): string {
  const exact = navRoutes.find((r) => r.end && r.to === path);
  if (exact) return exact.label;
  const prefix = navRoutes.find((r) => !r.end && path.startsWith(r.to));
  return prefix?.label ?? 'Inicio';
}
