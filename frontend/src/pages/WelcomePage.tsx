import { useState } from 'react';
import { useNavigate, Navigate } from 'react-router-dom';
import { useQuery, useQueryClient } from '@tanstack/react-query';
import { AlertCircle } from 'lucide-react';
import { Card } from '@/components/card';
import { Button } from '@/components/button';
import { PasswordInput } from '@/components/input';
import { Label } from '@/components/input';
import { Spinner } from '@/components/feedback';
import { queryKeys } from '@/services/queryKeys';
import { wailsClient } from '@/services/bindings';

export function WelcomePage() {
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const [password, setPassword] = useState('');
  const [error, setError] = useState('');
  const [submitting, setSubmitting] = useState(false);

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

  if (!state.data.configured) return <Navigate to="/configuracion-inicial" replace />;
  if (!state.data.passwordEnabled || state.data.unlocked) return <Navigate to="/" replace />;

  async function enter() {
    setSubmitting(true);
    setError('');
    try {
      await wailsClient.unlockLocalProfile(password);
      await queryClient.invalidateQueries({ queryKey: queryKeys.setup });
      navigate('/', { replace: true });
    } catch (cause) {
      setError(cause instanceof Error ? cause.message : 'Contraseña incorrecta.');
    } finally {
      setSubmitting(false);
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
      </Card>
    </div>
  );
}
