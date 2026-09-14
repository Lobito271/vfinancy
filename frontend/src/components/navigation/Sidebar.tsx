import { Link, NavLink } from 'react-router-dom';
import { ChevronsLeft, ChevronsRight } from 'lucide-react';
import { cx } from '@/utils/cx';
import { navItems } from './nav';
import { useSidebarStore } from '@/stores/sidebar';
import { Button } from '@/components/button';
import { Tooltip, TooltipContent, TooltipTrigger } from '@/components/misc';

export function Sidebar({ mobile = false }: { mobile?: boolean }) {
  const collapsed = useSidebarStore((s) => s.collapsed);
  const toggle = useSidebarStore((s) => s.toggle);

  return (
    <aside
      className={cx(
        'sidebar',
        mobile && 'sidebar--mobile',
        !mobile && collapsed && 'sidebar--collapsed',
      )}
      aria-label="Navegación principal"
    >
      <div className="sidebar__header">
        {!mobile && !collapsed && (
          <Link to="/" className="sidebar__brand">
            vfinancy
          </Link>
        )}
        {!mobile && (
          <Button
            variant="ghost"
            size="icon"
            onClick={toggle}
            aria-label={collapsed ? 'Expandir menú' : 'Colapsar menú'}
            className="sidebar__toggle"
          >
            {collapsed ? (
              <ChevronsRight strokeWidth={2.5} />
            ) : (
              <ChevronsLeft strokeWidth={2.5} />
            )}
          </Button>
        )}
      </div>

      <nav className="sidebar__nav scrollbar-thin">
        <ul className="sidebar__list">
          {navItems.map((item) => {
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
                          <Icon aria-hidden="true" strokeWidth={2.5} />
                          {(mobile || !collapsed) && (
                            <span className="truncate">{item.label}</span>
                          )}
                          {isActive && (mobile || !collapsed) && (
                            <span className="sidebar__dot" aria-hidden="true" />
                          )}
                        </>
                      )}
                    </NavLink>
                  </TooltipTrigger>
                  {!mobile && collapsed && (
                    <TooltipContent side="right">{item.label}</TooltipContent>
                  )}
                </Tooltip>
              </li>
            );
          })}
        </ul>
      </nav>
    </aside>
  );
}
