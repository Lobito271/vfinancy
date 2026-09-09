import { useState, useEffect } from 'react';
import { Link, NavLink, useLocation } from 'react-router-dom';
import { ChevronsLeft, ChevronsRight, ChevronDown } from 'lucide-react';
import { cx } from '@/utils/cx';
import { navItems, isNavGroup } from './nav';
import type { NavItem } from './nav';
import { useSidebarStore } from '@/stores/sidebar';
import { Button } from '@/components/button';
import { Tooltip, TooltipContent, TooltipTrigger } from '@/components/misc';

const STORAGE_KEY = 'vfinancy.sidebar.expanded';

function loadExpanded(): number[] {
  try {
    const raw = localStorage.getItem(STORAGE_KEY);
    return raw ? JSON.parse(raw) : [];
  } catch {
    return [];
  }
}

function saveExpanded(indices: number[]) {
  localStorage.setItem(STORAGE_KEY, JSON.stringify(indices));
}

function findGroupIndexForPath(items: NavItem[], pathname: string): number | null {
  for (let i = 0; i < items.length; i++) {
    const item = items[i];
    if (isNavGroup(item) && item.children.some((c) => pathname.startsWith(c.to) && c.to !== '/')) {
      return i;
    }
  }
  return null;
}

export function Sidebar({ mobile = false }: { mobile?: boolean }) {
  const collapsed = useSidebarStore((s) => s.collapsed);
  const toggle = useSidebarStore((s) => s.toggle);
  const location = useLocation();
  const [expanded, setExpanded] = useState<Set<number>>(() => new Set(loadExpanded()));

  useEffect(() => {
    const idx = findGroupIndexForPath(navItems, location.pathname);
    if (idx !== null) {
      setExpanded((prev) => {
        if (prev.has(idx)) return prev;
        const next = new Set(prev);
        next.add(idx);
        return next;
      });
    }
  }, [location.pathname]);

  useEffect(() => {
    saveExpanded(Array.from(expanded));
  }, [expanded]);

  const toggleGroup = (index: number) => {
    setExpanded((prev) => {
      const next = new Set(prev);
      if (next.has(index)) next.delete(index);
      else next.add(index);
      return next;
    });
  };

  const isGroupActive = (item: NavItem) => {
    if (!isNavGroup(item)) return false;
    return item.children.some((c) =>
      c.end ? location.pathname === c.to : location.pathname.startsWith(c.to),
    );
  };

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
      </div>

      <nav className="sidebar__nav scrollbar-thin">
        <ul className="sidebar__list">
          {navItems.map((item, index) => {
            if (!isNavGroup(item)) {
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
            }

            const Icon = item.icon;
            const open = expanded.has(index);
            const active = isGroupActive(item);

            if (mobile || !collapsed) {
              return (
                <li key={item.label} className="sidebar__group">
                  <button
                    type="button"
                    onClick={() => toggleGroup(index)}
                    className={cx('sidebar__group-header', active && 'active')}
                    aria-expanded={open}
                  >
                    <Icon aria-hidden="true" strokeWidth={2.5} />
                    <span className="truncate">{item.label}</span>
                    <ChevronDown
                      className={cx('sidebar__group-chevron', open && 'open')}
                      aria-hidden="true"
                      strokeWidth={2}
                    />
                  </button>
                  {open && (
                    <ul className="sidebar__sublist">
                      {item.children.map((child) => {
                        const ChildIcon = child.icon;
                        return (
                          <li key={child.to}>
                            <NavLink
                              to={child.to}
                              end={child.end}
                              className={({ isActive }) =>
                                cx('sidebar__sublink', isActive && 'active')
                              }
                            >
                              {({ isActive }) => (
                                <>
                                  <ChildIcon aria-hidden="true" strokeWidth={2} />
                                  <span className="truncate">{child.label}</span>
                                  {isActive && (
                                    <span className="sidebar__dot" aria-hidden="true" />
                                  )}
                                </>
                              )}
                            </NavLink>
                          </li>
                        );
                      })}
                    </ul>
                  )}
                </li>
              );
            }

            return (
              <li key={item.label}>
                <Tooltip>
                  <TooltipTrigger asChild>
                    <button
                      type="button"
                      className={cx('sidebar__link', active && 'active')}
                      onClick={() => {
                        toggleGroup(index);
                        if (collapsed) toggle();
                      }}
                    >
                      <Icon aria-hidden="true" strokeWidth={2.5} />
                    </button>
                  </TooltipTrigger>
                  <TooltipContent side="right">{item.label}</TooltipContent>
                </Tooltip>
              </li>
            );
          })}
        </ul>
      </nav>

      {!mobile && (
        <div className="sidebar__footer">
          <Button
            variant="ghost"
            size={collapsed ? 'icon-sm' : 'sm'}
            onClick={toggle}
            aria-label={collapsed ? 'Expandir menú' : 'Colapsar menú'}
            className={cx('sidebar__toggle', !collapsed && 'btn--justify-start')}
          >
            {collapsed ? (
              <ChevronsRight strokeWidth={2.5} />
            ) : (
              <>
                <ChevronsLeft strokeWidth={2.5} />
                <span>Colapsar</span>
              </>
            )}
          </Button>
        </div>
      )}
    </aside>
  );
}
