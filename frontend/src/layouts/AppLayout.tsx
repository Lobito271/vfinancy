import { Outlet } from 'react-router-dom';
import { Sidebar } from '@/components/navigation/Sidebar';
import { Topbar } from '@/components/navigation/Topbar';
import { Toaster } from '@/components/feedback';
import { Drawer } from '@/components/misc';
import { useSidebarStore } from '@/stores/sidebar';
import { ErrorBoundary } from '@/app/ErrorBoundary';

export function AppLayout() {
  const mobileOpen = useSidebarStore((s) => s.mobileOpen);
  const setMobileOpen = useSidebarStore((s) => s.setMobileOpen);

  return (
    <ErrorBoundary>
      <div className="app-shell">
        <Sidebar />
        <div className="app-shell__main-col">
          <Topbar />
          <main className="app-shell__main">
            <Outlet />
          </main>
        </div>
        <Drawer open={mobileOpen} onOpenChange={setMobileOpen}>
          <Sidebar mobile />
        </Drawer>
        <Toaster />
      </div>
    </ErrorBoundary>
  );
}
