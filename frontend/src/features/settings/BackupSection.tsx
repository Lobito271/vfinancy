import { useMutation } from '@tanstack/react-query';
import { Download, Upload } from 'lucide-react';
import { Section } from '@/components/layout';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/card';
import { Button } from '@/components/button';
import { wailsClient } from '@/services/bindings';
import { useNotificationStore } from '@/stores/notification';

export function BackupSection() {
  const push = useNotificationStore((s) => s.push);

  const exportBackup = useMutation({
    mutationFn: () => wailsClient.exportBackup(),
    onSuccess: (path) => {
      if (!path) return;
      push({ title: 'Respaldo exportado correctamente', variant: 'success' });
    },
    onError: (err: unknown) => {
      push({
        title: 'No se pudo exportar el respaldo',
        description: err instanceof Error ? err.message : undefined,
        variant: 'destructive',
      });
    },
  });

  const importBackup = useMutation({
    mutationFn: () => wailsClient.importBackup(),
    onSuccess: () => {
      push({
        title: 'Respaldo restaurado correctamente',
        description: 'La aplicación se reiniciará para aplicar los cambios.',
        variant: 'success',
      });
      setTimeout(() => window.location.reload(), 2500);
    },
    onError: (err: unknown) => {
      push({
        title: 'No se pudo restaurar el respaldo',
        description: err instanceof Error ? err.message : undefined,
        variant: 'destructive',
      });
    },
  });

  const busy = exportBackup.isPending || importBackup.isPending;

  return (
    <Section
      title="Respaldo y restauración"
      description="Copia de seguridad de la base de datos local."
    >
      <Card>
        <CardHeader>
          <CardTitle>Backup</CardTitle>
          <CardDescription>
            Exporta tu base de datos para transferirla a otro dispositivo o restaura desde un respaldo anterior.
          </CardDescription>
        </CardHeader>
        <CardContent>
          <div className="stack" style={{ maxWidth: '24rem' }}>
            <div className="settings-row">
              <div>
                <span className="settings-row__label">Exportar respaldo</span>
              </div>
              <Button
                onClick={() => exportBackup.mutate()}
                loading={exportBackup.isPending}
                disabled={busy}
              >
                <Download /> Exportar
              </Button>
            </div>
            <div className="settings-row">
              <div>
                <span className="settings-row__label">Importar respaldo</span>
              </div>
              <Button
                variant="outline"
                onClick={() => importBackup.mutate()}
                loading={importBackup.isPending}
                disabled={busy}
              >
                <Upload /> Importar
              </Button>
            </div>
          </div>
        </CardContent>
      </Card>
    </Section>
  );
}
