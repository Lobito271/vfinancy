import { useQuery, useMutation, useQueryClient } from '@tanstack/react-query';
import { useNavigate } from 'react-router-dom';
import { Sun, Moon, Monitor, Bell, CheckCheck, Lock, Menu as MenuIcon } from 'lucide-react';
import { useThemeStore, type Theme } from '@/stores/theme';
import { useUIStore } from '@/stores/ui';
import { useSidebarStore } from '@/stores/sidebar';
import { Button } from '@/components/button';
import { SearchInput } from '@/components/input';
import { Badge } from '@/components/badge';
import { Spinner } from '@/components/feedback';
import { t } from '@/locales';
import { formatRelative } from '@/utils/format';
import { queryKeys } from '@/services/queryKeys';
import { notificationsService, type AppNotification } from '@/services/notifications';
import { wailsClient } from '@/services/bindings';
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

const themeIcons: Record<Theme, typeof Sun> = {
  light: Sun,
  dark: Moon,
  system: Monitor,
};

function NotificationsBell() {
  const queryClient = useQueryClient();
  const unreadQuery = useQuery({
    queryKey: queryKeys.notifications.unread,
    queryFn: () => notificationsService.unreadCount(),
    refetchInterval: 30_000,
  });
  const listQuery = useQuery({
    queryKey: queryKeys.notifications.list(false),
    queryFn: () => notificationsService.list(false),
    refetchInterval: 30_000,
  });
  const notifications = listQuery.data ?? [];
  const invalidate = () => {
    void queryClient.invalidateQueries({ queryKey: queryKeys.notifications.all });
  };
  const markAll = useMutation({
    mutationFn: () => notificationsService.markAllRead(),
    onSuccess: invalidate,
  });
  const markOne = useMutation({
    mutationFn: (id: string) => notificationsService.markRead([id]),
    onSuccess: invalidate,
  });

  const unread = unreadQuery.data ?? 0;

  return (
    <DropdownMenu>
      <DropdownMenuTrigger asChild>
        <Button variant="ghost" size="icon" aria-label={t('notifications.title')}>
          <Bell />
          {unread > 0 && (
            <Badge variant="destructive" className="badge--count">
              {unread > 99 ? '99+' : unread}
            </Badge>
          )}
        </Button>
      </DropdownMenuTrigger>
      <DropdownMenuContent align="end" style={{ width: '20rem' }}>
        <DropdownMenuLabel>{t('notifications.title')}</DropdownMenuLabel>
        <DropdownMenuSeparator />
        <div className="notif-list">
          {listQuery.isFetching && notifications.length === 0 ? (
            <div className="notif-item">
              <Spinner />
            </div>
          ) : notifications.length === 0 ? (
            <div className="notif-item">
              <p className="notif-item__body">{t('notifications.empty')}</p>
            </div>
          ) : (
            notifications.map((n: AppNotification) => (
              <div
                key={n.id}
                className="notif-item"
                role="button"
                tabIndex={0}
                style={{ opacity: n.isRead ? 0.6 : 1, cursor: 'pointer' }}
                onClick={() => { if (!n.isRead) markOne.mutate(n.id); }}
                onKeyDown={(e) => { if (e.key === 'Enter' && !n.isRead) markOne.mutate(n.id); }}
              >
                <div className="notif-item__head">
                  <p className="notif-item__title">{n.title}</p>
                  <span className="notif-item__time">{formatRelative(n.createdAt)}</span>
                </div>
                <p className="notif-item__body">{n.message}</p>
              </div>
            ))
          )}
        </div>
        {notifications.length > 0 && unread > 0 && (
          <>
            <DropdownMenuSeparator />
            <DropdownMenuItem onSelect={() => markAll.mutate()}>
              <CheckCheck className="menu-item-icon" /> {t('notifications.markAllRead')}
            </DropdownMenuItem>
          </>
        )}
      </DropdownMenuContent>
    </DropdownMenu>
  );
}

export function Topbar() {
  const theme = useThemeStore((s) => s.theme);
  const setTheme = useThemeStore((s) => s.setTheme);
  const search = useUIStore((s) => s.globalSearch);
  const setSearch = useUIStore((s) => s.setGlobalSearch);
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
      navigate('/bienvenida', { replace: true });
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
        <MenuIcon />
      </Button>

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

        <NotificationsBell />

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
