import { Sun, Moon, Monitor, Bell } from 'lucide-react';
import { useThemeStore, type Theme } from '@/stores/theme';
import { useUIStore } from '@/stores/ui';
import { Button } from '@/components/button';
import { SearchInput } from '@/components/input';
import { Badge } from '@/components/badge';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuLabel,
  DropdownMenuRadioGroup,
  DropdownMenuRadioItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/misc';

const themeIcons: Record<Theme, typeof Sun> = {
  light: Sun,
  dark: Moon,
  system: Monitor,
};

export function Topbar() {
  const theme = useThemeStore((s) => s.theme);
  const setTheme = useThemeStore((s) => s.setTheme);
  const search = useUIStore((s) => s.globalSearch);
  const setSearch = useUIStore((s) => s.setGlobalSearch);

  const ThemeIcon = themeIcons[theme];

  return (
    <header className="topbar">
      <div className="topbar__search">
        <SearchInput
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          onClear={() => setSearch('')}
          placeholder="Buscar clientes, productos, ventas…"
          aria-label="Búsqueda global"
        />
      </div>

      <div className="topbar__actions">
        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <Button variant="ghost" size="icon" aria-label="Cambiar tema">
              <ThemeIcon />
            </Button>
          </DropdownMenuTrigger>
          <DropdownMenuContent align="end" style={{ width: '10rem' }}>
            <DropdownMenuLabel>Tema</DropdownMenuLabel>
            <DropdownMenuRadioGroup
              value={theme}
              onValueChange={(v) => setTheme(v as Theme)}
            >
              <DropdownMenuRadioItem value="light">
                <Sun className="menu-item-icon" /> Claro
              </DropdownMenuRadioItem>
              <DropdownMenuRadioItem value="dark">
                <Moon className="menu-item-icon" /> Oscuro
              </DropdownMenuRadioItem>
              <DropdownMenuRadioItem value="system">
                <Monitor className="menu-item-icon" /> Sistema
              </DropdownMenuRadioItem>
            </DropdownMenuRadioGroup>
          </DropdownMenuContent>
        </DropdownMenu>

        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <Button variant="ghost" size="icon" aria-label="Notificaciones">
              <Bell />
              <Badge variant="destructive" className="badge--count">
                3
              </Badge>
            </Button>
          </DropdownMenuTrigger>
          <DropdownMenuContent align="end" style={{ width: '20rem' }}>
            <DropdownMenuLabel>Notificaciones</DropdownMenuLabel>
            <DropdownMenuSeparator />
            <div className="notif-list">
              {[
                { t: 'Stock bajo', d: 'Detergente Nordic kg está bajo el mínimo', time: 'Hace 5 min' },
                { t: 'Pago recibido', d: 'Distribuidora García S.A.C. pagó S/ 3,400.00', time: 'Hace 1 h' },
                { t: 'Producto en remate', d: '5 productos pasaron los 25 días', time: 'Hace 3 h' },
              ].map((n) => (
                <div key={n.t} className="notif-item">
                  <div className="notif-item__head">
                    <p className="notif-item__title">{n.t}</p>
                    <span className="notif-item__time">{n.time}</span>
                  </div>
                  <p className="notif-item__body">{n.d}</p>
                </div>
              ))}
            </div>
          </DropdownMenuContent>
        </DropdownMenu>

        <div className="topbar__divider" aria-hidden="true" />

      </div>
    </header>
  );
}
