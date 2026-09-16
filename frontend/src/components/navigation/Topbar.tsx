import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { useNavigate } from 'react-router-dom';
import { Sun, Moon, Monitor, Lock, Menu as MenuIcon } from 'lucide-react';
import { useThemeStore, type Theme } from '@/stores/theme';
import { useSidebarStore } from '@/stores/sidebar';
import { Button } from '@/components/button';
import { queryKeys } from '@/services/queryKeys';
import { wailsClient } from '@/services/bindings';
import { Routes } from '@/constants/routes';

const themeIcons: Record<Theme, typeof Sun> = {
  light: Sun,
  dark: Moon,
  system: Monitor,
};

const themeLabels: Record<Theme, string> = {
  light: 'Claro',
  dark: 'Oscuro',
  system: 'Sistema',
};

export function Topbar() {
  const theme = useThemeStore((s) => s.theme);
  const toggleTheme = useThemeStore((s) => s.toggle);
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
        <Button
          variant="ghost"
          size="icon"
          onClick={toggleTheme}
          aria-label={`Cambiar tema (actual: ${themeLabels[theme]})`}
        >
          <ThemeIcon strokeWidth={2.5} />
        </Button>

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
