import { lazy, Suspense, useEffect } from 'react';
import { useQuery } from '@tanstack/react-query';
import { Routes, Route, Navigate } from 'react-router-dom';
import { AppLayout } from '@/layouts/AppLayout';
import { Spinner } from '@/components/feedback';
import { Providers } from './Providers';
import { useThemeStore } from '@/stores/theme';
import { wailsClient } from '@/services/bindings';

const DashboardPage = lazy(() =>
  import('@/features/dashboard/DashboardPage').then((m) => ({ default: m.DashboardPage })),
);
const CustomersPage = lazy(() => import('@/pages/CustomersPage').then((m) => ({ default: m.CustomersPage })));
const SuppliersPage = lazy(() => import('@/pages/SuppliersPage').then((m) => ({ default: m.SuppliersPage })));
const ProductsPage = lazy(() => import('@/pages/ProductsPage').then((m) => ({ default: m.ProductsPage })));
const CatalogSettingsPage = lazy(() => import('@/pages/CatalogSettingsPage').then((m) => ({ default: m.CatalogSettingsPage })));
const InventoryPage = lazy(() => import('@/pages/InventoryPage').then((m) => ({ default: m.InventoryPage })));
const PurchasesPage = lazy(() => import('@/pages/PurchasesPage').then((m) => ({ default: m.PurchasesPage })));
const SalesPage = lazy(() => import('@/pages/SalesPage').then((m) => ({ default: m.SalesPage })));
const TreasuryPage = lazy(() => import('@/pages/TreasuryPage').then((m) => ({ default: m.TreasuryPage })));
const AccountingPage = lazy(() => import('@/pages/AccountingPage').then((m) => ({ default: m.AccountingPage })));
const SettingsPage = lazy(() => import('@/pages/SettingsPage').then((m) => ({ default: m.SettingsPage })));
const SetupWizardPage = lazy(() => import('@/pages/SetupWizardPage').then((m) => ({ default: m.SetupWizardPage })));

function PageLoader() {
  return (
    <div className="page-loader">
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

function SetupState({ children, setup }: { children: React.ReactNode; setup: boolean }) {
  const state = useQuery({ queryKey: ['setup'], queryFn: () => wailsClient.getLocalAuthState() });
  if (state.isLoading) return <PageLoader />;
  if (state.isError) return <div className="page-loader">No se pudo comprobar la configuración.</div>;
  if (state.data?.configured !== setup)
    return <Navigate to={setup ? '/configuracion-inicial' : '/'} replace />;
  return <>{children}</>;
}

export function App() {
  return (
    <Providers>
      <ThemeInit />
      <Suspense fallback={<PageLoader />}>
        <Routes>
          <Route path="configuracion-inicial" element={<SetupState setup={false}><SetupWizardPage /></SetupState>} />
          <Route element={<SetupState setup><AppLayout /></SetupState>}>
            <Route index element={<DashboardPage />} />
            <Route path="clientes" element={<CustomersPage />} />
            <Route path="proveedores" element={<SuppliersPage />} />
            <Route path="productos" element={<ProductsPage />} />
            <Route path="configuracion-catalogo" element={<CatalogSettingsPage />} />
            <Route path="inventario" element={<InventoryPage />} />
            <Route path="compras" element={<PurchasesPage />} />
            <Route path="ventas" element={<SalesPage />} />
            <Route path="tesoreria" element={<TreasuryPage />} />
            <Route path="contabilidad" element={<AccountingPage />} />
            <Route path="configuracion" element={<SettingsPage />} />
            <Route path="*" element={<Navigate to="/" replace />} />
          </Route>
        </Routes>
      </Suspense>
    </Providers>
  );
}

export default App;
