import { useState } from 'react';
import { useMutation, useQuery } from '@tanstack/react-query';
import { Lock, Unlock } from 'lucide-react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/card';
import { Button } from '@/components/button';
import { Input, Label } from '@/components/input';
import { Section } from '@/components/layout';
import { wailsClient } from '@/services/bindings';
import { useNotificationStore } from '@/stores/notification';

export function SecuritySection() {
  const push = useNotificationStore((s) => s.push);
  const authState = useQuery({
    queryKey: ['auth', 'state'],
    queryFn: () => wailsClient.getLocalAuthState(),
  });

  const [currentPassword, setCurrentPassword] = useState('');
  const [newPassword, setNewPassword] = useState('');
  const [confirmPassword, setConfirmPassword] = useState('');

  const setPassword = useMutation({
    mutationFn: () => {
      if (newPassword !== confirmPassword) {
        throw new Error('Las contraseñas no coinciden');
      }
      if (newPassword.length < 6) {
        throw new Error('La contraseña debe tener al menos 6 caracteres');
      }
      if (authState.data?.passwordEnabled) {
        return wailsClient.setLocalPassword(currentPassword, newPassword);
      }
      return wailsClient.setLocalPassword('', newPassword);
    },
    onSuccess: () => {
      push({ title: 'Contraseña actualizada correctamente', variant: 'success' });
      setCurrentPassword('');
      setNewPassword('');
      setConfirmPassword('');
      authState.refetch();
    },
    onError: (err: unknown) => {
      push({
        title: 'No se pudo actualizar la contraseña',
        description: err instanceof Error ? err.message : undefined,
        variant: 'destructive',
      });
    },
  });

  const removePassword = useMutation({
    mutationFn: () => {
      if (!authState.data?.passwordEnabled) {
        throw new Error('No hay contraseña configurada');
      }
      return wailsClient.removeLocalPassword(currentPassword);
    },
    onSuccess: () => {
      push({ title: 'Contraseña eliminada correctamente', variant: 'success' });
      setCurrentPassword('');
      authState.refetch();
    },
    onError: (err: unknown) => {
      push({
        title: 'No se pudo eliminar la contraseña',
        description: err instanceof Error ? err.message : undefined,
        variant: 'destructive',
      });
    },
  });

  const passwordEnabled = authState.data?.passwordEnabled ?? false;

  return (
    <Section title="Seguridad" description="Configura la protección de acceso local.">
      <Card>
        <CardHeader>
          <CardTitle>Autenticación local</CardTitle>
          <CardDescription>
            {passwordEnabled
              ? 'La aplicación está protegida con contraseña. Puedes cambiarla o eliminarla.'
              : 'La aplicación no tiene contraseña. Configura una para proteger el acceso local.'}
          </CardDescription>
        </CardHeader>
        <CardContent>
          <div className="stack" style={{ maxWidth: '24rem' }}>
            {passwordEnabled && (
              <div className="field">
                <Label htmlFor="current-password">Contraseña actual</Label>
                <Input
                  id="current-password"
                  type="password"
                  value={currentPassword}
                  onChange={(e) => setCurrentPassword(e.target.value)}
                />
              </div>
            )}
            <div className="field">
              <Label htmlFor="new-password">
                {passwordEnabled ? 'Nueva contraseña' : 'Contraseña'}
              </Label>
              <Input
                id="new-password"
                type="password"
                value={newPassword}
                onChange={(e) => setNewPassword(e.target.value)}
              />
            </div>
            <div className="field">
              <Label htmlFor="confirm-password">Confirmar contraseña</Label>
              <Input
                id="confirm-password"
                type="password"
                value={confirmPassword}
                onChange={(e) => setConfirmPassword(e.target.value)}
              />
            </div>
            <div className="hstack" style={{ gap: '0.5rem', marginTop: '0.5rem' }}>
              <Button
                onClick={() => setPassword.mutate()}
                loading={setPassword.isPending}
                disabled={!newPassword || !confirmPassword}
              >
                <Lock /> {passwordEnabled ? 'Cambiar contraseña' : 'Establecer contraseña'}
              </Button>
              {passwordEnabled && (
                <Button
                  variant="outline"
                  onClick={() => removePassword.mutate()}
                  loading={removePassword.isPending}
                  disabled={!currentPassword}
                >
                  <Unlock /> Eliminar contraseña
                </Button>
              )}
            </div>
          </div>
        </CardContent>
      </Card>
    </Section>
  );
}
