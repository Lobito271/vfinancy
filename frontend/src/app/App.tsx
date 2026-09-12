import { lazy, Suspense, useEffect } from 'react';
import { useQuery } from '@tanstack/react-query';
import { Routes as RouterRoutes, Route, Navigate } from 'react-router-dom';
import { AppLayout } from '@/components/layout';
import { Spinner } from '@/components/feedback';
import { Providers } from './Providers';
import { useThemeStore } from '@/stores/theme';
import { queryKeys } from '@/services/queryKeys';
import { wailsClient } from '@/services/bindings';
import { Routes } from '@/constants/routes';

// React Router v6 forbids leading slashes in <Route path>, but to:/Navigate
// targets are absolute — strip the prefix for the path attribute.
const rel = (p: string) => p.replace(/^\//, '');

const DashboardPage = lazy(() =>
  import('@/pages/DashboardPage').then((m) => ({ default: m.DashboardPage })),
);
const InventoryPage = lazy(() => import('@/pages/InventoryPage').then((m) => ({ default: m.InventoryPage })));
const PurchasesPage = lazy(() => import('@/pages/PurchasesPage').then((m) => ({ default: m.PurchasesPage })));
const SalesPage = lazy(() => import('@/pages/SalesPage').then((m) => ({ default: m.SalesPage })));
const TreasuryPage = lazy(() => import('@/pages/TreasuryPage').then((m) => ({ default: m.TreasuryPage })));
const SettingsPage = lazy(() => import('@/pages/SettingsPage').then((m) => ({ default: m.SettingsPage })));
const SetupWizardPage = lazy(() => import('@/pages/SetupWizardPage').then((m) => ({ default: m.SetupWizardPage })));
const WelcomePage = lazy(() => import('@/pages/WelcomePage').then((m) => ({ default: m.WelcomePage })));

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
  const state = useQuery({ queryKey: queryKeys.setup, queryFn: () => wailsClient.getLocalAuthState() });
  if (state.isLoading) return <PageLoader />;
  if (state.isError) return <div className="page-loader">No se pudo comprobar la configuración.</div>;
  if (state.data?.configured !== setup)
    return <Navigate to={setup ? Routes.Setup : Routes.Dashboard} replace />;
  if (setup && state.data?.passwordEnabled && !state.data?.unlocked)
    return <Navigate to={Routes.Welcome} replace />;
  return <>{children}</>;
}

function LockedState({ children }: { children: React.ReactNode }) {
  const state = useQuery({ queryKey: queryKeys.setup, queryFn: () => wailsClient.getLocalAuthState() });
  if (state.isLoading) return <PageLoader />;
  if (state.isError || !state.data) return <PageLoader />;
  if (!state.data.configured) return <Navigate to={Routes.Setup} replace />;
  if (!state.data.passwordEnabled || state.data.unlocked) return <Navigate to={Routes.Dashboard} replace />;
  return <>{children}</>;
}

export function App() {
  return (
    <Providers>
      <ThemeInit />
      <Suspense fallback={<PageLoader />}>
        <RouterRoutes>
          <Route path={rel(Routes.Setup)} element={<SetupState setup={false}><SetupWizardPage /></SetupState>} />
          <Route path={rel(Routes.Welcome)} element={<LockedState><WelcomePage /></LockedState>} />
          <Route element={<SetupState setup><AppLayout /></SetupState>}>
            <Route index element={<DashboardPage />} />
            <Route path={rel(Routes.Inventory)} element={<InventoryPage />} />
            <Route path={rel(Routes.Purchases)} element={<PurchasesPage />} />
            <Route path={rel(Routes.Sales)} element={<SalesPage />} />
            <Route path={rel(Routes.Treasury)} element={<TreasuryPage />} />
            <Route path={rel(Routes.Settings)} element={<SettingsPage />} />
            <Route path="*" element={<Navigate to={Routes.Dashboard} replace />} />
          </Route>
        </RouterRoutes>
      </Suspense>
    </Providers>
  );
}
