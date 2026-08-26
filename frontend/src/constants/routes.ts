export const Routes = {
  Dashboard: '/',
  Customers: '/clientes',
  Suppliers: '/proveedores',
  Products: '/productos',
  CatalogSettings: '/configuracion-catalogo',
  Inventory: '/inventario',
  Purchases: '/compras',
  CustomerOrders: '/pedidos-cliente',
  Sales: '/ventas',
  Treasury: '/tesoreria',
  Settings: '/configuracion',
} as const;

export type RouteKey = keyof typeof Routes;
