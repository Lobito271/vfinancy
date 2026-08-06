import { Link, NavLink } from 'react-router-dom';
import { ChevronsLeft, ChevronsRight } from 'lucide-react';
import { cn } from '@/utils/cn';
import { navRoutes } from '@/lib/nav';
import { useSidebarStore } from '@/stores/sidebar';
import { Button } from '@/components/button';
import { Can } from '@/components/auth';

export function Sidebar() {
  const collapsed = useSidebarStore((s) => s.collapsed);
  const toggle = useSidebarStore((s) => s.toggle);

  return (
    <aside
      className={cn(
        'flex h-full flex-col border-r bg-card transition-[width] duration-200',
        collapsed ? 'w-16' : 'w-64',
      )}
      aria-label="Navegación principal"
    >
      <div className={cn('flex h-14 items-center border-b', collapsed ? 'justify-center px-2' : 'justify-between px-4')}>
        {!collapsed && (
          <Link to="/" className="flex items-center gap-2 font-semibold">
            <div className="flex h-7 w-7 items-center justify-center rounded-md bg-primary text-primary-foreground text-xs">
              vf
            </div>
            <span>vfinancy</span>
          </Link>
        )}
        {collapsed && (
          <div className="flex h-7 w-7 items-center justify-center rounded-md bg-primary text-primary-foreground text-xs">
            vf
          </div>
        )}
      </div>

      <nav className="flex-1 overflow-y-auto p-2 scrollbar-thin">
        <ul className="flex flex-col gap-0.5">
          {navRoutes.map((item) => {
            const Icon = item.icon;
            return (
              <li key={item.to}>
                <Can
                  permission={item.permission}
                  fallback={null}
                >
                  <NavLink
                    to={item.to}
                    end={item.end}
                    className={({ isActive }) =>
                      cn(
                        'group relative flex items-center gap-3 rounded-md px-3 py-2 text-sm transition-colors',
                        collapsed && 'justify-center px-0',
                        isActive
                          ? 'bg-accent text-accent-foreground font-medium'
                          : 'text-muted-foreground hover:bg-accent/50 hover:text-foreground',
                      )
                    }
                  >
                    {({ isActive }) => (
                      <>
                        <Icon className="h-4 w-4 shrink-0" aria-hidden="true" />
                        {!collapsed && <span className="truncate">{item.label}</span>}
                        {collapsed && (
                          <span className="pointer-events-none absolute left-full top-1/2 z-50 ml-2 -translate-y-1/2 whitespace-nowrap rounded-md border bg-popover px-2 py-1 text-xs text-popover-foreground opacity-0 shadow-md transition-opacity group-hover:opacity-100">
                            {item.label}
                          </span>
                        )}
                        {isActive && !collapsed && (
                          <span
                            className="ml-auto h-1.5 w-1.5 rounded-full bg-primary"
                            aria-hidden="true"
                          />
                        )}
                      </>
                    )}
                  </NavLink>
                </Can>
              </li>
            );
          })}
        </ul>
      </nav>

      <div className={cn('border-t p-2', collapsed && 'flex justify-center')}>
        <Button
          variant="ghost"
          size={collapsed ? 'icon-sm' : 'sm'}
          onClick={toggle}
          aria-label={collapsed ? 'Expandir menú' : 'Colapsar menú'}
          className={cn('w-full', !collapsed && 'justify-start')}
        >
          {collapsed ? (
            <ChevronsRight className="h-4 w-4" />
          ) : (
            <>
              <ChevronsLeft className="h-4 w-4" />
              <span>Colapsar</span>
            </>
          )}
        </Button>
      </div>
    </aside>
  );
}
