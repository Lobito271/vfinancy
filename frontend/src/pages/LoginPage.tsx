import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { Eye, EyeOff, LogIn } from 'lucide-react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/card';
import { Button } from '@/components/button';
import { Input, Field, Label } from '@/components/input';
import { Checkbox } from '@/components/checkbox';
import { Spinner } from '@/components/feedback';
import { t } from '@/locales';
import { Routes } from '@/constants/routes';
import { useSessionStore } from '@/stores/session';
import { authService } from '@/services/auth';
import { isPresent } from '@/utils/validators';

export function LoginPage() {
  const navigate = useNavigate();
  const { setUser, lastUsername } = useSessionStore();

  const [username, setUsername] = useState(lastUsername);
  const [password, setPassword] = useState('');
  const [remember, setRemember] = useState(true);
  const [showPassword, setShowPassword] = useState(false);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');

  const canSubmit = isPresent(username) && isPresent(password) && !loading;

  const onSubmit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!canSubmit) return;

    setLoading(true);
    setError('');

    try {
      const result = await authService.login({
        username: username.trim(),
        password,
        remember,
      });
      setUser(result.user, result.token, result.expiresAt);
      navigate(Routes.Dashboard);
    } catch {
      setError('Usuario o contrasena incorrectos');
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="grid h-full lg:grid-cols-2">
      <div className="hidden flex-col justify-between bg-primary p-10 text-primary-foreground lg:flex">
        <div className="flex items-center gap-2 text-lg font-semibold">
          <div className="flex h-8 w-8 items-center justify-center rounded-md bg-primary-foreground/10">
            vf
          </div>
          vfinancy
        </div>
        <div className="space-y-3">
          <h1 className="text-3xl font-semibold leading-tight">
            Plataforma ERP para la gestion integral de su negocio
          </h1>
          <p className="max-w-md text-sm text-primary-foreground/70">
            Compras, ventas, inventario, tesoreria y contabilidad en una sola aplicacion.
          </p>
        </div>
        <p className="text-xs text-primary-foreground/50">
          &copy; {new Date().getFullYear()} vfinancy S.A.C. Todos los derechos reservados.
        </p>
      </div>

      <div className="flex items-center justify-center p-6">
        <Card className="w-full max-w-md border-0 shadow-none sm:border sm:shadow-sm">
          <CardHeader>
            <CardTitle>{t('auth.welcomeBack')}</CardTitle>
            <CardDescription>{t('auth.loginSubtitle')}</CardDescription>
          </CardHeader>
          <CardContent>
            <form onSubmit={onSubmit} className="space-y-4">
              <Field>
                <Label htmlFor="username">{t('auth.username')}</Label>
                <Input
                  id="username"
                  value={username}
                  onChange={(e) => {
                    setUsername(e.target.value);
                    setError('');
                  }}
                  autoComplete="username"
                  autoFocus={!lastUsername}
                  required
                  disabled={loading}
                />
              </Field>

              <Field>
                <Label htmlFor="password">{t('auth.password')}</Label>
                <div className="relative">
                  <Input
                    id="password"
                    type={showPassword ? 'text' : 'password'}
                    value={password}
                    onChange={(e) => {
                      setPassword(e.target.value);
                      setError('');
                    }}
                    autoComplete="current-password"
                    className="pr-10"
                    required
                    disabled={loading}
                  />
                  <button
                    type="button"
                    onClick={() => setShowPassword((s) => !s)}
                    className="absolute right-1 top-1/2 -translate-y-1/2 rounded-md p-1.5 text-muted-foreground hover:bg-accent hover:text-foreground"
                    aria-label={showPassword ? 'Ocultar contrasena' : 'Mostrar contrasena'}
                    tabIndex={-1}
                  >
                    {showPassword ? <EyeOff className="h-4 w-4" /> : <Eye className="h-4 w-4" />}
                  </button>
                </div>
              </Field>

              <div className="flex items-center justify-between">
                <label className="inline-flex items-center gap-2 text-sm">
                  <Checkbox
                    checked={remember}
                    onCheckedChange={(v) => setRemember(v === true)}
                    disabled={loading}
                  />
                  {t('auth.rememberMe')}
                </label>
                <a href="#" className="text-sm text-primary hover:underline">
                  {t('auth.forgotPassword')}
                </a>
              </div>

              {error && (
                <p className="text-sm text-destructive" role="alert">
                  {error}
                </p>
              )}

              <Button type="submit" className="w-full" size="lg" disabled={!canSubmit}>
                {loading ? (
                  <Spinner size="sm" className="text-primary-foreground" />
                ) : (
                  <LogIn />
                )}
                {t('auth.login')}
              </Button>
            </form>
          </CardContent>
        </Card>
      </div>
    </div>
  );
}
