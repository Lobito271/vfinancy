export const Routes = {
  Login: '/login',
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
  Reports: '/reportes',
  Settings: '/configuracion',
  Administration: '/administracion',
} as const;

export type RouteKey = keyof typeof Routes;

export const ProtectedRoutes = new Set<RouteKey>([
  'Dashboard',
  'Customers',
  'Suppliers',
  'Products',
  'CatalogSettings',
  'Inventory',
  'Purchases',
  'Sales',
  'Treasury',
  'Accounting',
  'Reports',
  'Settings',
  'Administration',
]);

export const PublicRoutes = new Set<RouteKey>(['Login']);

export function isProtectedRoute(path: string): boolean {
  if (path === Routes.Dashboard) return true;
  return Object.values(Routes).some(
    (r) => r !== Routes.Login && r !== Routes.Dashboard && path.startsWith(r),
  );
}
