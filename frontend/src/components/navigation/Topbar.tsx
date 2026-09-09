import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { useNavigate } from 'react-router-dom';
import { Sun, Moon, Monitor, Lock, Menu as MenuIcon } from 'lucide-react';
import { useThemeStore, type Theme } from '@/stores/theme';
import { useSidebarStore } from '@/stores/sidebar';
import { Button } from '@/components/button';
import { queryKeys } from '@/services/queryKeys';
import { wailsClient } from '@/services/bindings';
import { Routes } from '@/constants/routes';
import {
  DropdownMenu,
  DropdownMenuContent,
  DropdownMenuLabel,
  DropdownMenuRadioGroup,
  DropdownMenuRadioItem,
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
  const setMobileOpen = useSidebarStore((s) => s.setMobileOpen);
  const navigate = useNavigate();
  const queryClient = useQueryClient();

  const authState = useQuery({
    queryKey: queryKeys.setup,
    queryFn: () => wailsClient.getLocalAuthState(),
    staleTime: 30_000,
  });

  const ThemeIcon = themeIcons[theme];

  const lock = useMutation({
    mutationFn: () => wailsClient.lockLocalProfile(),
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: queryKeys.setup });
      navigate(Routes.Welcome, { replace: true });
    },
  });

  return (
    <header className="topbar">
      <Button
        variant="ghost"
        size="icon"
        className="topbar__hamburger"
        onClick={() => setMobileOpen(true)}
        aria-label="Abrir menú"
      >
        <MenuIcon strokeWidth={2.5} />
      </Button>

      <div className="topbar__actions">
        <DropdownMenu>
          <DropdownMenuTrigger asChild>
            <Button variant="ghost" size="icon" aria-label="Cambiar tema">
              <ThemeIcon strokeWidth={2.5} />
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

        {authState.data?.passwordEnabled && (
          <>
            <div className="topbar__divider" aria-hidden="true" />
            <Button
              variant="ghost"
              size="icon"
              onClick={() => lock.mutate()}
              loading={lock.isPending}
              aria-label="Bloquear aplicación"
            >
              <Lock />
            </Button>
          </>
        )}
      </div>
    </header>
  );
}
