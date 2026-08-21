import { Outlet, useLocation } from 'react-router-dom';
import { Sidebar } from '@/components/navigation/Sidebar';
import { Topbar } from '@/components/navigation/Topbar';
import { Breadcrumbs } from '@/components/navigation/Breadcrumbs';
import { Toaster } from '@/components/feedback';
import { ErrorBoundary } from '@/app/ErrorBoundary';

export function AppLayout() {
  const location = useLocation();
  return (
    <ErrorBoundary>
      <div className="app-shell">
        <Sidebar />
        <div className="app-shell__main-col">
          <Topbar />
          <div className="app-shell__breadcrumbs">
            <Breadcrumbs path={location.pathname} />
          </div>
          <main className="app-shell__main">
            <Outlet />
          </main>
        </div>
        <Toaster />
      </div>
    </ErrorBoundary>
  );
}
