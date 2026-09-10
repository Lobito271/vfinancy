import { useState } from 'react';
import { useQuery, useQueryClient } from '@tanstack/react-query';
import { z } from 'zod';
import { Briefcase, ShieldCheck, HardDriveDownload, Cloud, Palette, Building2 } from 'lucide-react';
import { PageContainer, PageHeader } from '@/components/layout';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/card';
import { Form, NumberField, TextField, EmailField } from '@/components/form';
import { Button } from '@/components/button';
import { Label } from '@/components/input';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/select';
import { Drawer } from '@/components/misc';
import { SecuritySection } from '@/features/settings/components/SecuritySection';
import { BackupSection, CloudSyncSection } from '@/features/settings/components/SyncAndBackupSection';
import { wailsClient } from '@/services/bindings';
import { queryKeys } from '@/services/queryKeys';
import { useNotificationStore } from '@/stores/notification';
import { useThemeStore, type Theme } from '@/stores/theme';

const tabs = [
  { id: 'business', label: 'Negocio', icon: Briefcase },
  { id: 'company', label: 'Empresa', icon: Building2 },
  { id: 'auth', label: 'Autenticación', icon: ShieldCheck },
  { id: 'backup', label: 'Respaldos', icon: HardDriveDownload },
  { id: 'sync', label: 'Sincronización', icon: Cloud },
  { id: 'appearance', label: 'Apariencia', icon: Palette },
] as const;

type TabId = (typeof tabs)[number]['id'];

const businessSchema = z.object({
  clearanceDays: z.number().int().min(1, 'Entre 1 y 365').max(365),
  importCostFactor: z.number().min(0, 'Debe ser >= 0').max(100),
  fallbackExchangeRate: z.number().min(0.01, 'Entre 0.01 y 100').max(100),
  customsLimitUsd: z.number().min(0, 'Debe ser >= 0').max(1_000_000),
});

type BusinessValues = z.infer<typeof businessSchema>;

function BusinessTab() {
  const queryClient = useQueryClient();
  const push = useNotificationStore((s) => s.push);
  const prefs = useQuery({ queryKey: queryKeys.settings.preferences, queryFn: () => wailsClient.getPreferences() });

  const save = async (values: BusinessValues) => {
    try {
      await wailsClient.updatePreference('clearance_days', String(values.clearanceDays));
      await wailsClient.updatePreference('import_cost_factor', String(values.importCostFactor));
      await wailsClient.updatePreference('fallback_exchange_rate', String(values.fallbackExchangeRate));
      await wailsClient.updatePreference('customs_limit_usd', String(values.customsLimitUsd));
      await queryClient.invalidateQueries({ queryKey: queryKeys.settings.preferences });
      push({ title: 'Parámetros de negocio guardados', variant: 'success' });
    } catch (cause) {
      push({
        title: 'No se pudo guardar la configuración',
        description: cause instanceof Error ? cause.message : undefined,
        variant: 'destructive',
      });
    }
  };

  if (prefs.isLoading) return null;

  const defaults = {
    clearanceDays: prefs.data?.clearanceDays ?? 25,
    importCostFactor: prefs.data?.importCostFactor ?? 0.07,
    fallbackExchangeRate: prefs.data?.fallbackExchangeRate ?? 1,
    customsLimitUsd: prefs.data?.customsLimitUSD ?? 0,
  };

  return (
    <Card>
      <CardHeader>
        <CardTitle>Parámetros de negocio</CardTitle>
        <CardDescription>Controlan el remate, el costo de importación y el tope aduanero.</CardDescription>
      </CardHeader>
      <CardContent>
        <Form<BusinessValues> key={JSON.stringify(defaults)} schema={businessSchema} defaultValues={defaults} onSubmit={save}>
          {({ formState }) => (
            <div className="stack" style={{ maxWidth: '26rem' }}>
              <NumberField
                name="clearanceDays"
                label="Días para remate"
                description="Un lote pasa a remate tras estos días desde su ingreso."
                min={1}
                max={365}
                required
              />
              <NumberField
                name="importCostFactor"
                label="Costo de importación USD"
                description="Factor aplicado sobre el costo en dólares (ej. 0.07 = 7%)."
                min={0}
                step={0.01}
                required
              />
              <NumberField
                name="fallbackExchangeRate"
                label="TC de respaldo"
                description="Tipo de cambio de contingencia cuando no hay conexión."
                min={0.01}
                max={100}
                step={0.01}
                required
              />
              <NumberField
                name="customsLimitUsd"
                label="Tope aduanero USD"
                description="Monto máximo por lote antes de la advertencia."
                min={0}
                max={1_000_000}
                step={1}
                required
              />
              <div>
                <Button type="submit" loading={formState.isSubmitting}>
                  Guardar
                </Button>
              </div>
            </div>
          )}
        </Form>
      </CardContent>
    </Card>
  );
}

const companySchema = z.object({
  name: z.string().trim().min(2, 'Ingresa la razón social.'),
  taxId: z.string().regex(/^(10|20)\d{9}$/, 'RUC debe tener 11 dígitos e iniciar con 10 o 20.').optional().or(z.literal('')),
  email: z.string().email('Correo inválido.').optional().or(z.literal('')),
  fiscalAddress: z.string().trim().optional().or(z.literal('')),
});

type CompanyValues = z.infer<typeof companySchema>;

function CompanyTab() {
  const queryClient = useQueryClient();
  const push = useNotificationStore((s) => s.push);
  const profile = useQuery({ queryKey: ['settings', 'profile'], queryFn: () => wailsClient.getLocalProfile() });

  const save = async (values: CompanyValues) => {
    try {
      await wailsClient.updateLocalProfile(values);
      await queryClient.invalidateQueries({ queryKey: ['settings', 'profile'] });
      push({ title: 'Datos de la empresa guardados', variant: 'success' });
    } catch (cause) {
      push({
        title: 'No se pudo guardar la información de la empresa',
        description: cause instanceof Error ? cause.message : undefined,
        variant: 'destructive',
      });
    }
  };

  if (profile.isLoading) return null;

  const defaults = {
    name: profile.data?.name ?? '',
    taxId: profile.data?.taxId ?? '',
    email: profile.data?.email ?? '',
    fiscalAddress: profile.data?.fiscalAddress ?? '',
  };

  return (
    <Card>
      <CardHeader>
        <CardTitle>Datos de la empresa</CardTitle>
        <CardDescription>Identidad corporativa usada en comprobantes y documentación.</CardDescription>
      </CardHeader>
      <CardContent>
        <Form<CompanyValues> key={JSON.stringify(defaults)} schema={companySchema} defaultValues={defaults} onSubmit={save}>
          {({ formState }) => (
            <div className="stack" style={{ maxWidth: '26rem' }}>
              <TextField name="name" label="Razón social" required />
              <TextField
                name="taxId"
                label="RUC"
                description="11 dígitos, iniciando con 10 o 20."
              />
              <EmailField name="email" label="Correo electrónico" />
              <TextField name="fiscalAddress" label="Dirección fiscal" />
              <div>
                <Button type="submit" loading={formState.isSubmitting}>
                  Guardar
                </Button>
              </div>
            </div>
          )}
        </Form>
      </CardContent>
    </Card>
  );
}

function AuthTab() {
  const [authOpen, setAuthOpen] = useState(false);

  return (
    <>
      <Card>
        <CardHeader>
          <CardTitle>Seguridad</CardTitle>
          <CardDescription>Contraseña local y clave de recuperación del dispositivo.</CardDescription>
        </CardHeader>
        <CardContent>
          <Button onClick={() => setAuthOpen(true)}>
            <ShieldCheck /> Administrar autenticación
          </Button>
        </CardContent>
      </Card>
      <Drawer
        open={authOpen}
        onOpenChange={setAuthOpen}
        title="Autenticación"
        description="Contraseña local y clave de recuperación."
      >
        <SecuritySection />
      </Drawer>
    </>
  );
}

function AppearanceTab() {
  const theme = useThemeStore((state) => state.theme);
  const setTheme = useThemeStore((state) => state.setTheme);
  const push = useNotificationStore((s) => s.push);
  const profile = useQuery({ queryKey: ['settings', 'profile'], queryFn: () => wailsClient.getLocalProfile() });

  return (
    <Card>
      <CardHeader>
        <CardTitle>Apariencia</CardTitle>
        <CardDescription>El tema se aplica al instante en este dispositivo.</CardDescription>
      </CardHeader>
      <CardContent>
        <div className="stack" style={{ maxWidth: '24rem' }}>
          <div className="field">
            <Label htmlFor="settings-theme">Tema</Label>
            <Select
              items={[
                { value: 'light', label: 'Claro' },
                { value: 'dark', label: 'Oscuro' },
                { value: 'system', label: 'Sistema' },
              ]}
              value={theme}
              onValueChange={(value) => {
                setTheme((value ?? 'system') as Theme);
                push({ title: 'Tema actualizado', variant: 'success' });
              }}
            >
              <SelectTrigger id="settings-theme">
                <SelectValue />
              </SelectTrigger>
              <SelectContent>
                <SelectItem value="light">Claro</SelectItem>
                <SelectItem value="dark">Oscuro</SelectItem>
                <SelectItem value="system">Sistema</SelectItem>
              </SelectContent>
            </Select>
          </div>
          <div className="settings-row">
            <span className="settings-row__label">Empresa</span>
            <strong>{profile.data?.name}</strong>
          </div>
        </div>
      </CardContent>
    </Card>
  );
}

export function SettingsPage() {
  const [tab, setTab] = useState<TabId>('business');
  const active = tabs.find((t) => t.id === tab) ?? tabs[0];

  return (
    <PageContainer>
      <PageHeader title="Configuración" subtitle="Administra las preferencias de tu operación." />
      <div className="settings-layout">
        <nav className="settings-nav" aria-label="Secciones de configuración">
          {tabs.map((t) => {
            const Icon = t.icon;
            return (
              <button
                key={t.id}
                type="button"
                className="settings-nav__item"
                data-active={t.id === tab || undefined}
                onClick={() => setTab(t.id)}
              >
                <Icon />
                <span>{t.label}</span>
              </button>
            );
          })}
        </nav>
        <section className="stack" style={{ flex: 1 }}>
          <h2 className="sr-only">{active.label}</h2>
          {tab === 'business' && <BusinessTab />}
          {tab === 'company' && <CompanyTab />}
          {tab === 'auth' && <AuthTab />}
          {tab === 'backup' && <BackupSection />}
          {tab === 'sync' && <CloudSyncSection />}
          {tab === 'appearance' && <AppearanceTab />}
        </section>
      </div>
    </PageContainer>
  );
}
