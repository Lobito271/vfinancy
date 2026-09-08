import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { useEffect, useState } from 'react';
import { Cloud, HardDriveDownload } from 'lucide-react';
import { Section } from '@/components/layout';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/card';
import { Button } from '@/components/button';
import { Input, Label, PasswordInput } from '@/components/input';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/select';
import { wailsClient } from '@/services/bindings';
import { useNotificationStore } from '@/stores/notification';

export function CloudSyncSection() {
  const push = useNotificationStore((s) => s.push);
  const queryClient = useQueryClient();
  const sync = useQuery({ queryKey: ['settings', 'sync'], queryFn: () => wailsClient.getSyncConfig() });

  const [serverUrl, setServerUrl] = useState('');
  const [apiKey, setApiKey] = useState('');
  const [enabled, setEnabled] = useState(false);
  const [pollIntervalSec, setPollIntervalSec] = useState(30);

  const save = useMutation({
    mutationFn: () =>
      wailsClient.saveSyncConfig({
        serverUrl,
        apiKey,
        enabled,
        pollIntervalSec,
      }),
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: ['settings', 'sync'] });
      push({ title: enabled ? 'Sincronización habilitada' : 'Sincronización deshabilitada', variant: 'success' });
    },
    onError: (err: unknown) => {
      push({ title: 'No se pudo guardar la configuración', description: err instanceof Error ? err.message : undefined, variant: 'destructive' });
    },
  });

  const test = useMutation({
    mutationFn: () => wailsClient.testSyncConnection(serverUrl, apiKey),
    onSuccess: () => push({ title: 'Conexión exitosa', description: 'El servidor respondió correctamente.', variant: 'success' }),
    onError: (err: unknown) =>
      push({ title: 'Sin conexión', description: err instanceof Error ? err.message : undefined, variant: 'destructive' }),
  });

  useEffect(() => {
    if (sync.data) {
      setServerUrl(sync.data.serverUrl);
      setApiKey(sync.data.apiKey);
      setEnabled(sync.data.enabled);
      setPollIntervalSec(sync.data.pollIntervalSec);
    }
  }, [sync.data]);

  if (sync.isLoading) return null;

  return (
    <Section title="Sincronización en la nube" description="Conecta este dispositivo a tu espejo PostgreSQL vía el servidor de sincronización.">
      <Card>
        <CardHeader>
          <CardTitle>Servidor de sincronización</CardTitle>
          <CardDescription>
            Todos los equipos que compartan este servidor mantienen la misma información maestra.
          </CardDescription>
        </CardHeader>
        <CardContent>
          <div className="stack" style={{ maxWidth: '24rem' }}>
            <div className="field">
              <Label htmlFor="sync-url">URL del servidor</Label>
              <Input
                id="sync-url"
                value={serverUrl}
                onChange={(e) => setServerUrl(e.target.value)}
                placeholder="https://sync.miempresa.com"
                autoComplete="off"
                spellCheck={false}
              />
            </div>
            <div className="field">
              <Label htmlFor="sync-key">Clave de API</Label>
              <PasswordInput
                id="sync-key"
                value={apiKey}
                onChange={(e) => setApiKey(e.target.value)}
                placeholder="Clave compartida del servidor"
              />
            </div>
            <div className="field">
              <Label htmlFor="sync-interval">Intervalo de sincronización (segundos)</Label>
              <Input
                id="sync-interval"
                type="number"
                min={10}
                value={String(pollIntervalSec)}
                onChange={(e) => setPollIntervalSec(Number(e.target.value) || 30)}
              />
            </div>
            <div className="field">
              <Label htmlFor="sync-enabled">Estado</Label>
              <Select
                items={[
                  { value: '1', label: 'Activado' },
                  { value: '0', label: 'Desactivado' },
                ]}
                value={enabled ? '1' : '0'}
                onValueChange={(v) => setEnabled(v === '1')}
              >
                <SelectTrigger id="sync-enabled">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="1">Activado</SelectItem>
                  <SelectItem value="0">Desactivado</SelectItem>
                </SelectContent>
              </Select>
            </div>
            <div className="hstack hstack--sm">
              <Button onClick={() => test.mutate()} loading={test.isPending} variant="outline">
                Probar conexión
              </Button>
              <Button onClick={() => save.mutate()} loading={save.isPending}>
                <Cloud /> Guardar
              </Button>
            </div>
          </div>
        </CardContent>
      </Card>
    </Section>
  );
}

export function BackupSection() {
  const push = useNotificationStore((s) => s.push);
  const backup = useMutation({
    mutationFn: () => wailsClient.createBackup(),
    onSuccess: (path: string) =>
      push({ title: 'Copia de seguridad creada', description: path, variant: 'success' }),
    onError: (err: unknown) =>
      push({ title: 'No se pudo crear la copia', description: err instanceof Error ? err.message : undefined, variant: 'destructive' }),
  });

  return (
    <Section title="Copia de seguridad" description="Respalda tu base de datos local en una carpeta segura.">
      <Card>
        <CardHeader>
          <CardTitle>Respaldo manual</CardTitle>
          <CardDescription>
            Crea un archivo con la base de datos actual. La carpeta se configura con la preferencia
            <code> backup.folder</code> (por defecto <code>~/.vfinancy/backups</code>).
          </CardDescription>
        </CardHeader>
        <CardContent>
          <Button onClick={() => backup.mutate()} loading={backup.isPending}>
            <HardDriveDownload /> Crear copia ahora
          </Button>
        </CardContent>
      </Card>
    </Section>
  );
}