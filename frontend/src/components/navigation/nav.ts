import {
  LayoutDashboard, Warehouse, ShoppingCart, Receipt, Landmark, Settings,
  ClipboardList, Boxes, Package, Tag, CreditCard, Repeat,
  Shield, Database, RefreshCw, Palette,
} from 'lucide-react';
import type { LucideProps } from 'lucide-react';
import { Routes } from '@/constants/routes';

export interface NavRoute {
  to: string;
  label: string;
  icon: React.ComponentType<LucideProps>;
  end?: boolean;
}

export interface NavGroup {
  label: string;
  icon: React.ComponentType<LucideProps>;
  children: NavRoute[];
}

export type NavItem = NavRoute | NavGroup;

export function isNavGroup(item: NavItem): item is NavGroup {
  return 'children' in item;
}

export const navItems: NavItem[] = [
  { to: Routes.Dashboard, label: 'Inicio', icon: LayoutDashboard, end: true },

  {
    label: 'Compras e Importaciones',
    icon: ShoppingCart,
    children: [
      { to: Routes.Purchases, label: 'Pedidos', icon: ClipboardList },
      { to: Routes.PurchasesLots, label: 'Lotes de importación', icon: Boxes },
    ],
  },

  {
    label: 'Inventario y Almacén',
    icon: Warehouse,
    children: [
      { to: Routes.Inventory, label: 'Stock', icon: Package },
      { to: Routes.InventoryClearance, label: 'Remates', icon: Tag },
    ],
  },

  {
    label: 'Ventas y Facturación',
    icon: Receipt,
    children: [
      { to: Routes.Sales, label: 'Ventas', icon: Receipt },
      { to: Routes.SalesPayments, label: 'Cobros', icon: ClipboardList },
    ],
  },

  {
    label: 'Tesorería y Pasivos',
    icon: Landmark,
    children: [
      { to: Routes.Treasury, label: 'Tarjetas', icon: CreditCard },
      { to: Routes.TreasuryCycles, label: 'Ciclos', icon: Repeat },
    ],
  },

  {
    label: 'Configuración y Sistema',
    icon: Settings,
    children: [
      { to: Routes.Settings, label: 'Negocio', icon: Settings },
      { to: Routes.Settings, label: 'Autenticación', icon: Shield },
      { to: Routes.Settings, label: 'Respaldos', icon: Database },
      { to: Routes.Settings, label: 'Sincronización', icon: RefreshCw },
      { to: Routes.Settings, label: 'Apariencia', icon: Palette },
    ],
  },
];
