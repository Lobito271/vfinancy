import { LayoutDashboard, Users, Truck, Package, Warehouse, ShoppingCart, ShoppingBag, Receipt, Landmark, Settings, Tags } from 'lucide-react';
import { Routes } from '@/constants/routes';

interface NavRoute {
  to: string;
  label: string;
  icon: React.ComponentType<{ className?: string }>;
  end?: boolean;
}

export const navRoutes: NavRoute[] = [
  { to: Routes.Dashboard, label: 'Inicio', icon: LayoutDashboard, end: true },
  { to: Routes.Customers, label: 'Clientes', icon: Users },
  { to: Routes.Suppliers, label: 'Proveedores', icon: Truck },
  { to: Routes.Products, label: 'Productos', icon: Package },
  { to: Routes.CatalogSettings, label: 'Catálogo', icon: Tags },
  { to: Routes.Inventory, label: 'Inventario', icon: Warehouse },
  { to: Routes.Purchases, label: 'Compras', icon: ShoppingCart },
  { to: Routes.CustomerOrders, label: 'Pedidos de cliente', icon: ShoppingBag },
  { to: Routes.Sales, label: 'Ventas', icon: Receipt },
  { to: Routes.Treasury, label: 'Tesorería', icon: Landmark },
  { to: Routes.Settings, label: 'Configuración', icon: Settings },
];
