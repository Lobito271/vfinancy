import { useState } from 'react';
import { useNavigate, Navigate } from 'react-router-dom';
import { useQuery, useQueryClient } from '@tanstack/react-query';
import { AlertCircle, ChevronDown, KeyRound } from 'lucide-react';
import { Card } from '@/components/card';
import { Button } from '@/components/button';
import { PasswordInput } from '@/components/input';
import { Input, Label } from '@/components/input';
import { Spinner } from '@/components/feedback';
import { queryKeys } from '@/services/queryKeys';
import { wailsClient } from '@/services/bindings';
import { Routes } from '@/constants/routes';

export function WelcomePage() {
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const [password, setPassword] = useState('');
  const [error, setError] = useState('');
  const [submitting, setSubmitting] = useState(false);
  const [recoveryOpen, setRecoveryOpen] = useState(false);
  const [recoveryToken, setRecoveryToken] = useState('');
  const [recoveryPassword, setRecoveryPassword] = useState('');
  const [recoveryError, setRecoveryError] = useState('');
  const [recovering, setRecovering] = useState(false);

  const state = useQuery({ queryKey: queryKeys.setup, queryFn: () => wailsClient.getLocalAuthState() });

  if (state.isLoading) {
    return (
      <div className="welcome">
        <Spinner size="lg" />
      </div>
    );
  }

  if (state.isError || !state.data) {
    return (
      <div className="welcome">
        <p>No se pudo comprobar el estado de la aplicación.</p>
      </div>
    );
  }

  if (!state.data.configured) return <Navigate to={Routes.Setup} replace />;
  if (!state.data.passwordEnabled || state.data.unlocked) return <Navigate to={Routes.Dashboard} replace />;

  async function enter() {
    setSubmitting(true);
    setError('');
    try {
      await wailsClient.unlockLocalProfile(password);
      await queryClient.invalidateQueries({ queryKey: queryKeys.setup });
      navigate(Routes.Dashboard, { replace: true });
    } catch (cause) {
      setError(cause instanceof Error ? cause.message : 'Contraseña incorrecta.');
    } finally {
      setSubmitting(false);
    }
  }

  async function recover() {
    setRecovering(true);
    setRecoveryError('');
    try {
      await wailsClient.recoverWithToken({ token: recoveryToken, newPassword: recoveryPassword });
      await queryClient.invalidateQueries({ queryKey: queryKeys.setup });
      navigate(Routes.Dashboard, { replace: true });
    } catch (cause) {
      setRecoveryError(cause instanceof Error ? cause.message : 'No se pudo recuperar el acceso.');
    } finally {
      setRecovering(false);
    }
  }

  return (
    <div className="welcome">
      <Card className="welcome__card">
        <div className="welcome__logo">vfinancy</div>
        <form
          className="welcome__form"
          onSubmit={(e) => {
            e.preventDefault();
            void enter();
          }}
        >
          <Label htmlFor="welcome-password">Contraseña</Label>
          <PasswordInput
            id="welcome-password"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            autoComplete="current-password"
            autoFocus
            required
          />
          {error && (
            <p className="welcome__error" role="alert">
              <AlertCircle />
              {error}
            </p>
          )}
          <Button type="submit" loading={submitting}>
            Entrar
          </Button>
        </form>

        <div className="welcome__recovery">
          <Button
            variant="ghost"
            size="sm"
            aria-expanded={recoveryOpen}
            onClick={() => setRecoveryOpen((open) => !open)}
          >
            <KeyRound /> Usar token de recuperación <ChevronDown />
          </Button>
          {recoveryOpen && (
            <form
              className="welcome__form"
              onSubmit={(e) => {
                e.preventDefault();
                void recover();
              }}
            >
              <Label htmlFor="welcome-recovery-token">Token de recuperación</Label>
              <Input
                id="welcome-recovery-token"
                value={recoveryToken}
                onChange={(e) => setRecoveryToken(e.target.value)}
                autoComplete="off"
                spellCheck={false}
                required
              />
              <Label htmlFor="welcome-recovery-password">Nueva contraseña</Label>
              <PasswordInput
                id="welcome-recovery-password"
                value={recoveryPassword}
                onChange={(e) => setRecoveryPassword(e.target.value)}
                autoComplete="new-password"
                required
              />
              {recoveryError && (
                <p className="welcome__error" role="alert">
                  <AlertCircle />
                  {recoveryError}
                </p>
              )}
              <Button type="submit" variant="outline" loading={recovering}>
                Recuperar acceso
              </Button>
            </form>
          )}
        </div>
      </Card>
    </div>
  );
}
