import { lazy, Suspense } from 'react';
import { Routes, Route, Navigate } from 'react-router-dom';
import { AppLayout } from '@/layouts/AppLayout';
import { Spinner } from '@/components/feedback';
import { Providers } from './Providers';
import { LoginPage } from '@/pages/LoginPage';
import { useThemeStore } from '@/stores/theme';
import { useEffect } from 'react';

const DashboardPage = lazy(() =>
  import('@/features/dashboard/DashboardPage').then((m) => ({ default: m.DashboardPage })),
);
const CustomersPage = lazy(() => import('@/pages/CustomersPage').then((m) => ({ default: m.CustomersPage })));
const SuppliersPage = lazy(() => import('@/pages/SuppliersPage').then((m) => ({ default: m.SuppliersPage })));
const ProductsPage = lazy(() => import('@/pages/ProductsPage').then((m) => ({ default: m.ProductsPage })));
const InventoryPage = lazy(() => import('@/pages/InventoryPage').then((m) => ({ default: m.InventoryPage })));
const PurchasesPage = lazy(() => import('@/pages/PurchasesPage').then((m) => ({ default: m.PurchasesPage })));
const SalesPage = lazy(() => import('@/pages/SalesPage').then((m) => ({ default: m.SalesPage })));
const TreasuryPage = lazy(() => import('@/pages/TreasuryPage').then((m) => ({ default: m.TreasuryPage })));
const AccountingPage = lazy(() => import('@/pages/AccountingPage').then((m) => ({ default: m.AccountingPage })));
const ReportsPage = lazy(() => import('@/pages/ReportsPage').then((m) => ({ default: m.ReportsPage })));
const SettingsPage = lazy(() => import('@/pages/SettingsPage').then((m) => ({ default: m.SettingsPage })));
const AdministrationPage = lazy(() => import('@/pages/AdministrationPage').then((m) => ({ default: m.AdministrationPage })));

function PageLoader() {
  return (
    <div className="flex h-full items-center justify-center p-16">
      <Spinner size="lg" />
    </div>
  );
}

function ThemeInit() {
  const applyToDocument = useThemeStore((s) => s.applyToDocument);
  useEffect(() => {
    applyToDocument();
  }, [applyToDocument]);
  return null;
}

export function App() {
  return (
    <Providers>
      <ThemeInit />
      <Suspense fallback={<PageLoader />}>
        <Routes>
          <Route path="/login" element={<LoginPage />} />
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
      </Suspense>
    </Providers>
  );
}

export default App;
