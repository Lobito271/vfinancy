import {
  LayoutDashboard, ShoppingCart, Package, Receipt, CreditCard, Settings,
} from 'lucide-react';
import type { LucideProps } from 'lucide-react';
import { Routes } from '@/constants/routes';

export interface NavRoute {
  to: string;
  label: string;
  icon: React.ComponentType<LucideProps>;
  end?: boolean;
}

export const navItems: NavRoute[] = [
  { to: Routes.Dashboard, label: 'Inicio', icon: LayoutDashboard, end: true },
  { to: Routes.Purchases, label: 'Compras', icon: ShoppingCart },
  { to: Routes.Inventory, label: 'Inventario', icon: Package },
  { to: Routes.Sales, label: 'Ventas', icon: Receipt },
  { to: Routes.Treasury, label: 'Tesorería', icon: CreditCard },
  { to: Routes.Settings, label: 'Ajustes', icon: Settings },
];
