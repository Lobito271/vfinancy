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
      <div className="flex h-full bg-background">
        <Sidebar />
        <div className="flex flex-1 flex-col overflow-hidden">
          <Topbar />
          <div className="border-b bg-background px-6 py-3">
            <Breadcrumbs path={location.pathname} />
          </div>
          <main className="flex-1 overflow-y-auto">
            <Outlet />
          </main>
        </div>
        <Toaster />
      </div>
    </ErrorBoundary>
  );
}
