import { Link, NavLink } from 'react-router-dom';
import { ChevronsLeft, ChevronsRight } from 'lucide-react';
import { cx } from '@/utils/cx';
import { navRoutes } from '@/lib/nav';
import { useSidebarStore } from '@/stores/sidebar';
import { Button } from '@/components/button';
import { Tooltip, TooltipContent, TooltipTrigger } from '@/components/misc';

export function Sidebar() {
  const collapsed = useSidebarStore((s) => s.collapsed);
  const toggle = useSidebarStore((s) => s.toggle);

  return (
    <aside
      className={cx('sidebar', collapsed && 'sidebar--collapsed')}
      aria-label="Navegación principal"
    >
      <div className="sidebar__header">
        {!collapsed && (
          <Link to="/" className="sidebar__brand">
            vfinancy
          </Link>
        )}
      </div>

      <nav className="sidebar__nav scrollbar-thin">
        <ul className="sidebar__list">
          {navRoutes.map((item) => {
            const Icon = item.icon;
            return (
              <li key={item.to}>
                <Tooltip>
                  <TooltipTrigger asChild>
                    <NavLink
                      to={item.to}
                      end={item.end}
                      className={({ isActive }) =>
                        cx('sidebar__link', isActive && 'active')
                      }
                    >
                      {({ isActive }) => (
                        <>
                          <Icon aria-hidden="true" />
                          {!collapsed && <span className="truncate">{item.label}</span>}
                          {isActive && !collapsed && (
                            <span className="sidebar__dot" aria-hidden="true" />
                          )}
                        </>
                      )}
                    </NavLink>
                  </TooltipTrigger>
                  {collapsed && (
                    <TooltipContent side="right">{item.label}</TooltipContent>
                  )}
                </Tooltip>
              </li>
            );
          })}
        </ul>
      </nav>

      <div className="sidebar__footer">
        <Button
          variant="ghost"
          size={collapsed ? 'icon-sm' : 'sm'}
          onClick={toggle}
          aria-label={collapsed ? 'Expandir menú' : 'Colapsar menú'}
          className={cx('sidebar__toggle', !collapsed && 'btn--justify-start')}
        >
          {collapsed ? (
            <ChevronsRight />
          ) : (
            <>
              <ChevronsLeft />
              <span>Colapsar</span>
            </>
          )}
        </Button>
      </div>
    </aside>
  );
}
