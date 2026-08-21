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
            <div className="hstack hstack--md">
              <Building2 className="icon-md muted" aria-hidden="true" />
              <div>
                <CardTitle>Datos de la empresa</CardTitle>
                <CardDescription>Información general del negocio</CardDescription>
              </div>
            </div>
          </CardHeader>
          <CardContent>
            <p className="desc-text">
              La edición de datos de la empresa estará disponible en una fase posterior.
            </p>
          </CardContent>
        </Card>
      </Section>

      <Section title="Apariencia" description="Tema y preferencias visuales">
        <Card>
          <CardHeader>
            <div className="hstack hstack--md">
              <Sun className="icon-md muted" aria-hidden="true" />
              <div>
                <CardTitle>Tema</CardTitle>
                <CardDescription>Seleccione el tema visual del sistema</CardDescription>
              </div>
            </div>
          </CardHeader>
          <CardContent className="stack stack--lg">
            <div className="stack stack--sm">
              <Label htmlFor="theme">Tema de la aplicación</Label>
              <Select value={theme} onValueChange={(v) => setTheme(v as Theme)}>
                <SelectTrigger id="theme" style={{ width: "16rem" }}>
                  <SelectValue />
                </SelectTrigger>
                <SelectContent>
                  <SelectItem value="light">
                    <span className="hstack hstack--sm">
                      <Sun className="icon-sm" /> Claro
                    </span>
                  </SelectItem>
                  <SelectItem value="dark">
                    <span className="hstack hstack--sm">
                      <Moon className="icon-sm" /> Oscuro
                    </span>
                  </SelectItem>
                  <SelectItem value="system">
                    <span className="hstack hstack--sm">
                      <Monitor className="icon-sm" /> Sistema
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
            <div className="hstack hstack--md">
              <Globe className="icon-md muted" aria-hidden="true" />
              <div>
                <CardTitle>Idioma</CardTitle>
                <CardDescription>Idioma de la interfaz</CardDescription>
              </div>
            </div>
          </CardHeader>
          <CardContent>
            <p className="desc-text">
              Español (Perú) — es-PE. La selección de otros idiomas se habilitará en versiones futuras.
            </p>
          </CardContent>
        </Card>
      </Section>

      <Section title="Notificaciones" description="Preferencias de alertas">
        <Card>
          <CardHeader>
            <div className="hstack hstack--md">
              <Bell className="icon-md muted" aria-hidden="true" />
              <div>
                <CardTitle>Alertas del sistema</CardTitle>
                <CardDescription>Configure qué alertas desea recibir</CardDescription>
              </div>
            </div>
          </CardHeader>
          <CardContent>
            <p className="desc-text">
              La configuración detallada de notificaciones se habilitará en una fase posterior.
            </p>
          </CardContent>
        </Card>
      </Section>
    </PageContainer>
  );
}
