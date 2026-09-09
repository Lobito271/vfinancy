import { useState } from 'react';
import { useQuery, useQueryClient } from '@tanstack/react-query';
import { useNavigate } from 'react-router-dom';
import { KeyRound, Lock, ShieldCheck } from 'lucide-react';
import { Section } from '@/components/layout';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/card';
import { Button } from '@/components/button';
import { PasswordInput, Label } from '@/components/input';
import { wailsClient } from '@/services/bindings';
import { queryKeys } from '@/services/queryKeys';
import { useNotificationStore } from '@/stores/notification';
import { Routes } from '@/constants/routes';

export function SecuritySection() {
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const push = useNotificationStore((s) => s.push);
  const auth = useQuery({ queryKey: queryKeys.setup, queryFn: () => wailsClient.getLocalAuthState() });

  const [current, setCurrent] = useState('');
  const [next, setNext] = useState('');
  const [busy, setBusy] = useState(false);
  const [recoveryBusy, setRecoveryBusy] = useState(false);
  const [recoveryToken, setRecoveryToken] = useState('');
  const passwordEnabled = auth.data?.passwordEnabled ?? false;

  const valid = next === '' || next.length >= 8;

  async function run(action: () => Promise<void>, okMsg: string) {
    setBusy(true);
    try {
      await action();
      push({ title: okMsg, variant: 'success' });
      setCurrent('');
      setNext('');
      await queryClient.invalidateQueries({ queryKey: queryKeys.setup });
    } catch (cause) {
      push({
        title: 'No se pudo actualizar la contraseña',
        description: cause instanceof Error ? cause.message : undefined,
        variant: 'destructive',
      });
    } finally {
      setBusy(false);
    }
  }

  const savePassword = () =>
    run(
      () => wailsClient.setLocalPassword({ current, next }),
      passwordEnabled ? 'Contraseña actualizada' : 'Contraseña creada',
    );

  const removePassword = () =>
    run(() => wailsClient.removeLocalPassword(current), 'Contraseña eliminada');

  const lockNow = () =>
    run(async () => {
      await wailsClient.lockLocalProfile();
      navigate(Routes.Welcome, { replace: true });
    }, 'Aplicación bloqueada');

  const generateRecovery = async () => {
    setRecoveryBusy(true);
    try {
      const token = await wailsClient.getRecoveryToken();
      setRecoveryToken(token);
    } catch (cause) {
      push({
        title: 'No se pudo generar la clave',
        description: cause instanceof Error ? cause.message : undefined,
        variant: 'destructive',
      });
    } finally {
      setRecoveryBusy(false);
    }
  };

  const downloadRecovery = () => {
    const content = `Clave de recuperación de vfinancy\n\n${recoveryToken}\n\nGuárdala en un lugar seguro. Si pierdes la contraseña,\nesta clave te permite recuperar el acceso a este equipo.\n`;
    const url = URL.createObjectURL(new Blob([content], { type: 'text/plain' }));
    const a = document.createElement('a');
    a.href = url;
    a.download = 'vfinancy-recovery-key.txt';
    a.click();
    URL.revokeObjectURL(url);
  };

  return (
    <Section
      title="Seguridad"
      description="Protege el acceso a la aplicación en este dispositivo."
    >
      <Card>
        <CardHeader>
          <CardTitle>Contraseña local</CardTitle>
          <CardDescription>
            {passwordEnabled
              ? 'La aplicación pide la contraseña al iniciar. Usa al menos 8 caracteres.'
              : 'Aún no hay contraseña configurada. Crea una para proteger tu información.'}
          </CardDescription>
        </CardHeader>
        <CardContent>
          <div className="stack" style={{ maxWidth: '24rem' }}>
            {passwordEnabled && (
              <div className="field">
                <Label htmlFor="security-current">Contraseña actual</Label>
                <PasswordInput
                  id="security-current"
                  value={current}
                  onChange={(e) => setCurrent(e.target.value)}
                  autoComplete="current-password"
                />
              </div>
            )}
            <div className="field">
              <Label htmlFor="security-next">
                {passwordEnabled ? 'Nueva contraseña' : 'Nueva contraseña (opcional)'}
              </Label>
              <PasswordInput
                id="security-next"
                value={next}
                onChange={(e) => setNext(e.target.value)}
                placeholder={passwordEnabled ? 'Déjalo vacío para no cambiarla' : undefined}
                autoComplete="new-password"
                invalid={!valid}
              />
              {!valid && <p className="field-error">Usa al menos 8 caracteres.</p>}
            </div>
            <div className="hstack hstack--sm" style={{ flexWrap: 'wrap' }}>
              <Button onClick={savePassword} loading={busy} disabled={!valid}>
                <KeyRound /> {passwordEnabled ? 'Actualizar contraseña' : 'Crear contraseña'}
              </Button>
              {passwordEnabled && (
                <Button variant="outline" onClick={removePassword} loading={busy} disabled={!current}>
                  Quitar contraseña
                </Button>
              )}
              {passwordEnabled && (
                <Button variant="ghost" onClick={lockNow} loading={busy}>
                  <Lock /> Bloquear ahora
                </Button>
              )}
            </div>

            {!recoveryToken && (
              <div className="hstack hstack--sm">
                <Button variant="outline" onClick={generateRecovery} loading={recoveryBusy}>
                  <ShieldCheck /> Generar clave de recuperación
                </Button>
              </div>
            )}
            {recoveryToken && (
              <div className="dialog-note">
                <p className="fw-medium">Guarda esta clave en un lugar seguro.</p>
                <code className="recovery-token">{recoveryToken}</code>
                <div className="hstack hstack--sm" style={{ marginTop: '0.5rem' }}>
                  <Button size="sm" onClick={downloadRecovery}>
                    Descargar clave
                  </Button>
                  <Button size="sm" variant="ghost" onClick={() => setRecoveryToken('')}>
                    Entendido
                  </Button>
                </div>
              </div>
            )}
          </div>
        </CardContent>
      </Card>
    </Section>
  );
}
