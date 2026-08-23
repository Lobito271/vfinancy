export const Routes = {
  Dashboard: '/',
  Customers: '/clientes',
  Suppliers: '/proveedores',
  Products: '/productos',
  CatalogSettings: '/configuracion-catalogo',
  Inventory: '/inventario',
  Purchases: '/compras',
  Sales: '/ventas',
  Treasury: '/tesoreria',
  Accounting: '/contabilidad',
  Settings: '/configuracion',
} as const;

export type RouteKey = keyof typeof Routes;

export function isProtectedRoute(path: string): boolean {
  if (path === Routes.Dashboard) return true;
  return Object.values(Routes).some(
    (r) => r !== Routes.Dashboard && path.startsWith(r),
  );
}
