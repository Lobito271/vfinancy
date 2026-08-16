import { lazy, Suspense, useEffect } from 'react';
import { Routes, Route, Navigate, useLocation } from 'react-router-dom';
import { AppLayout } from '@/layouts/AppLayout';
import { Spinner } from '@/components/feedback';
import { Providers } from './Providers';
import { LoginPage } from '@/pages/LoginPage';
import { useThemeStore } from '@/stores/theme';
import { useSessionStore } from '@/stores/session';
import { Routes as RoutePaths } from '@/constants/routes';

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

function isSessionExpired(expiresAt: string | null): boolean {
  if (!expiresAt) return true;
  return new Date(expiresAt).getTime() <= Date.now();
}

function hasValidSession(
  isAuthenticated: boolean,
  token: string | null,
  expiresAt: string | null,
): boolean {
  return isAuthenticated && Boolean(token) && !isSessionExpired(expiresAt);
}

function RequireAuth({ children }: { children: React.ReactNode }) {
  const isAuthenticated = useSessionStore((s) => s.isAuthenticated);
  const token = useSessionStore((s) => s.token);
  const expiresAt = useSessionStore((s) => s.expiresAt);
  const location = useLocation();

  if (!hasValidSession(isAuthenticated, token, expiresAt)) {
    return <Navigate to={RoutePaths.Login} replace state={{ from: location }} />;
  }
  return <>{children}</>;
}

function PublicOnly({ children }: { children: React.ReactNode }) {
  const isAuthenticated = useSessionStore((s) => s.isAuthenticated);
  const token = useSessionStore((s) => s.token);
  const expiresAt = useSessionStore((s) => s.expiresAt);

  if (hasValidSession(isAuthenticated, token, expiresAt)) {
    return <Navigate to={RoutePaths.Dashboard} replace />;
  }
  return <>{children}</>;
}

export function App() {
  return (
    <Providers>
      <ThemeInit />
      <Suspense fallback={<PageLoader />}>
        <Routes>
          <Route
            path="/login"
            element={
              <PublicOnly>
                <LoginPage />
              </PublicOnly>
            }
          />
          <Route
            element={
              <RequireAuth>
                <AppLayout />
              </RequireAuth>
            }
          >
            <Route index element={<DashboardPage />} />
            <Route path="clientes" element={<CustomersPage />} />
            <Route path="proveedores" element={<SuppliersPage />} />
            <Route path="productos" element={<ProductsPage />} />
            <Route path="configuracion-catalogo" element={<CatalogSettingsPage />} />
            <Route path="inventario" element={<InventoryPage />} />
            <Route path="inventory" element={<InventoryPage />} />
            <Route path="compras" element={<PurchasesPage />} />
            <Route path="ventas" element={<SalesPage />} />
            <Route path="sales" element={<SalesPage />} />
            <Route path="tesoreria" element={<TreasuryPage />} />
            <Route path="treasury" element={<TreasuryPage />} />
            <Route path="contabilidad" element={<AccountingPage />} />
            <Route path="accounting" element={<AccountingPage />} />
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
