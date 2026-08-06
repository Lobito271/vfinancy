import { useThemeStore } from '@/stores/theme';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/card';
import { Label } from '@/components/input';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/select';
import { PageContainer, PageHeader, Section } from '@/components/layout';
import { Sun, Moon, Monitor, Building2, Globe, Bell } from 'lucide-react';
import type { Theme } from '@/stores/theme';

export function SettingsPage() {
  const theme = useThemeStore((s) => s.theme);
  const setTheme = useThemeStore((s) => s.setTheme);

  return (
    <PageContainer>
      <PageHeader
        title="Configuración"
        subtitle="Preferencias generales del sistema"
      />

      <Section title="Empresa" description="Información de la empresa actual">
        <Card>
          <CardHeader>
            <div className="flex items-center gap-3">
              <Building2 className="h-5 w-5 text-muted-foreground" aria-hidden="true" />
              <div>
                <CardTitle>Datos de la empresa</CardTitle>
                <CardDescription>Información general del negocio</CardDescription>
              </div>
            </div>
          </CardHeader>
          <CardContent>
            <p className="text-sm text-muted-foreground">
              La edición de datos de la empresa estará disponible en una fase posterior.
            </p>
          </CardContent>
        </Card>
      </Section>

      <Section title="Apariencia" description="Tema y preferencias visuales">
        <Card>
          <CardHeader>
            <div className="flex items-center gap-3">
              <Sun className="h-5 w-5 text-muted-foreground" aria-hidden="true" />
              <div>
                <CardTitle>Tema</CardTitle>
                <CardDescription>Seleccione el tema visual del sistema</CardDescription>
              </div>
            </div>
          </CardHeader>
          <CardContent className="space-y-4">
            <div className="space-y-2">
              <Label htmlFor="theme">Tema de la aplicación</Label>
              <Select value={theme} onValueChange={(v) => setTheme(v as Theme)}>
                <SelectTrigger id="theme" className="w-full sm:w-64">
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="light">
                    <span className="flex items-center gap-2">
                      <Sun className="h-4 w-4" /> Claro
                    </span>
                  </SelectItem>
                  <SelectItem value="dark">
                    <span className="flex items-center gap-2">
                      <Moon className="h-4 w-4" /> Oscuro
                    </span>
                  </SelectItem>
                  <SelectItem value="system">
                    <span className="flex items-center gap-2">
                      <Monitor className="h-4 w-4" /> Sistema
                    </span>
                  </SelectItem>
                </SelectContent>
              </Select>
            </div>
          </CardContent>
        </Card>
      </Section>

      <Section title="Idioma y región" description="Configuración regional">
        <Card>
          <CardHeader>
            <div className="flex items-center gap-3">
              <Globe className="h-5 w-5 text-muted-foreground" aria-hidden="true" />
              <div>
                <CardTitle>Idioma</CardTitle>
                <CardDescription>Idioma de la interfaz</CardDescription>
              </div>
            </div>
          </CardHeader>
          <CardContent>
            <p className="text-sm text-muted-foreground">
              Español (Perú) — es-PE. La selección de otros idiomas se habilitará en versiones futuras.
            </p>
          </CardContent>
        </Card>
      </Section>

      <Section title="Notificaciones" description="Preferencias de alertas">
        <Card>
          <CardHeader>
            <div className="flex items-center gap-3">
              <Bell className="h-5 w-5 text-muted-foreground" aria-hidden="true" />
              <div>
                <CardTitle>Alertas del sistema</CardTitle>
                <CardDescription>Configure qué alertas desea recibir</CardDescription>
              </div>
            </div>
          </CardHeader>
          <CardContent>
            <p className="text-sm text-muted-foreground">
              La configuración detallada de notificaciones se habilitará en una fase posterior.
            </p>
          </CardContent>
        </Card>
      </Section>
    </PageContainer>
  );
}
