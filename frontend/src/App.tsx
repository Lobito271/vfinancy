import { useEffect } from 'react';
import { Routes, Route, Navigate } from 'react-router-dom';
import { AppLayout } from '@/layouts/AppLayout';
import { DashboardPage } from '@/pages/DashboardPage';
import { CustomersPage } from '@/pages/CustomersPage';
import { SuppliersPage } from '@/pages/SuppliersPage';
import { ProductsPage } from '@/pages/ProductsPage';
import { InventoryPage } from '@/pages/InventoryPage';
import { PurchasesPage } from '@/pages/PurchasesPage';
import { SalesPage } from '@/pages/SalesPage';
import { TreasuryPage } from '@/pages/TreasuryPage';
import { AccountingPage } from '@/pages/AccountingPage';
import { ReportsPage } from '@/pages/ReportsPage';
import { SettingsPage } from '@/pages/SettingsPage';
import { AdministrationPage } from '@/pages/AdministrationPage';
import { useThemeStore } from '@/stores/theme';

export default function App() {
  const applyToDocument = useThemeStore((s) => s.applyToDocument);

  useEffect(() => {
    applyToDocument();
  }, [applyToDocument]);

  return (
    <Routes>
      <Route element={<AppLayout />}>
        <Route index element={<DashboardPage />} />
        <Route path="clientes" element={<CustomersPage />} />
        <Route path="proveedores" element={<SuppliersPage />} />
        <Route path="productos" element={<ProductsPage />} />
        <Route path="inventario" element={<InventoryPage />} />
        <Route path="compras" element={<PurchasesPage />} />
        <Route path="ventas" element={<SalesPage />} />
        <Route path="tesoreria" element={<TreasuryPage />} />
        <Route path="contabilidad" element={<AccountingPage />} />
        <Route path="reportes" element={<ReportsPage />} />
        <Route path="configuracion" element={<SettingsPage />} />
        <Route path="administracion" element={<AdministrationPage />} />
        <Route path="*" element={<Navigate to="/" replace />} />
      </Route>
    </Routes>
  );
}
