import { useNavigate } from 'react-router-dom';
import { ChevronsUpDown, LogOut, User as UserIcon } from 'lucide-react';
import { useThemeStore, type Theme } from '@/stores/theme';
import { useSessionStore } from '@/stores/session';
import { useUIStore } from '@/stores/ui';
import { Button } from '@/components/button';
import { SearchInput } from '@/components/input';
import { Badge } from '@/components/badge';
import { Icons } from '@/design-system/icons';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuItem,
  DropdownMenuLabel,
  DropdownMenuRadioGroup,
  DropdownMenuRadioItem,
  DropdownMenuSeparator,
  DropdownMenuTrigger,
} from '@/components/misc';

const themeIcons: Record<Theme, typeof Icons.Theme.Sun> = {
  light: Icons.Theme.Sun,
  dark: Icons.Theme.Moon,
  system: Icons.Theme.System,
};

export function Topbar() {
  const theme = useThemeStore((s) => s.theme);
  const setTheme = useThemeStore((s) => s.setTheme);
  const user = useSessionStore((s) => s.user);
  const logout = useSessionStore((s) => s.logout);
  const search = useUIStore((s) => s.globalSearch);
  const setSearch = useUIStore((s) => s.setGlobalSearch);
  const navigate = useNavigate();

  const ThemeIcon = themeIcons[theme];

  return (
    <header className="sticky top-0 z-30 flex h-14 items-center gap-2 border-b bg-background/95 px-4 backdrop-blur supports-[backdrop-filter]:bg-background/80">
      <div className="flex-1 max-w-xl">
        <SearchInput
          value={search}
          onChange={(e) => setSearch(e.target.value)}
          onClear={() => setSearch('')}
          placeholder="Buscar clientes, productos, ventas…"
          aria-label="Búsqueda global"
          className="w-full"
        />
      </div>

      <div className="ml-auto flex items-center gap-1">
        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <Button variant="ghost" size="icon" aria-label="Cambiar tema">
              <ThemeIcon />
            </Button>
          </DropdownMenuTrigger>
          <DropdownMenuContent align="end" className="w-40">
            <DropdownMenuLabel>Tema</DropdownMenuLabel>
            <DropdownMenuRadioGroup
              value={theme}
              onValueChange={(v) => setTheme(v as Theme)}
            >
              <DropdownMenuRadioItem value="light">
                <Icons.Theme.Sun className="mr-2 h-4 w-4" /> Claro
              </DropdownMenuRadioItem>
              <DropdownMenuRadioItem value="dark">
                <Icons.Theme.Moon className="mr-2 h-4 w-4" /> Oscuro
              </DropdownMenuRadioItem>
              <DropdownMenuRadioItem value="system">
                <Icons.Theme.System className="mr-2 h-4 w-4" /> Sistema
              </DropdownMenuRadioItem>
            </DropdownMenuRadioGroup>
          </DropdownMenuContent>
        </DropdownMenu>

        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <Button variant="ghost" size="icon" className="relative" aria-label="Notificaciones">
              <Icons.Bell />
              <Badge
                variant="destructive"
                className="absolute -right-0.5 -top-0.5 h-4 min-w-4 justify-center px-1 text-[10px]"
              >
                3
              </Badge>
            </Button>
          </DropdownMenuTrigger>
          <DropdownMenuContent align="end" className="w-80">
            <DropdownMenuLabel>Notificaciones</DropdownMenuLabel>
            <DropdownMenuSeparator />
            <div className="max-h-80 overflow-y-auto">
              {[
                { t: 'Stock bajo', d: 'Detergente Nordic kg está bajo el mínimo', time: 'Hace 5 min' },
                { t: 'Pago recibido', d: 'Distribuidora García S.A.C. pagó S/ 3,400.00', time: 'Hace 1 h' },
                { t: 'Producto en remate', d: '5 productos pasaron los 25 días', time: 'Hace 3 h' },
              ].map((n, i) => (
                <div key={i} className="flex flex-col gap-0.5 border-b px-3 py-2 last:border-0 hover:bg-accent/50">
                  <div className="flex items-center justify-between">
                    <p className="text-sm font-medium">{n.t}</p>
                    <span className="text-xs text-muted-foreground">{n.time}</span>
                  </div>
                  <p className="text-xs text-muted-foreground">{n.d}</p>
                </div>
              ))}
            </div>
          </DropdownMenuContent>
        </DropdownMenu>

        <div className="mx-2 h-6 w-px bg-border" aria-hidden="true" />

        {user && (
          <DropdownMenu>
            <DropdownMenuTrigger asChild>
              <Button variant="ghost" size="sm" className="gap-2">
                <div className="flex h-7 w-7 items-center justify-center rounded-full bg-primary text-primary-foreground text-xs font-semibold">
                  {user.fullName
                    .split(' ')
                    .map((p) => p[0])
                    .slice(0, 2)
                    .join('')}
                </div>
                <div className="hidden text-left sm:block">
                  <p className="text-sm font-medium leading-tight">{user.fullName}</p>
                  <p className="text-xs text-muted-foreground leading-tight">{user.company}</p>
                </div>
                <ChevronsUpDown className="h-3.5 w-3.5 text-muted-foreground" />
              </Button>
            </DropdownMenuTrigger>
            <DropdownMenuContent align="end" className="w-56">
              <DropdownMenuLabel>Mi cuenta</DropdownMenuLabel>
              <DropdownMenuSeparator />
              <DropdownMenuItem onSelect={() => navigate('/configuracion')}>
                <UserIcon className="mr-2 h-4 w-4" />
                Mi perfil
              </DropdownMenuItem>
              <DropdownMenuItem onSelect={() => navigate('/configuracion')}>
                Configuración
              </DropdownMenuItem>
              <DropdownMenuSeparator />
              <DropdownMenuItem
                onSelect={() => {
                  logout();
                  navigate('/login');
                }}
                className="text-destructive focus:text-destructive"
              >
                <LogOut className="mr-2 h-4 w-4" />
                Cerrar sesión
              </DropdownMenuItem>
            </DropdownMenuContent>
          </DropdownMenu>
        )}
      </div>
    </header>
  );
}
