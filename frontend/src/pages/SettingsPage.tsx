import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { z } from 'zod';
import { Save } from 'lucide-react';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/card';
import { Button } from '@/components/button';
import { EmailField, Form, NumberField, TextField, TextareaField } from '@/components/form';
import { Label } from '@/components/input';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/select';
import { PageContainer, PageHeader, Section } from '@/components/layout';
import { SecuritySection } from '@/features/settings/SecuritySection';
import { wailsClient } from '@/services/bindings';
import { queryKeys } from '@/services/queryKeys';
import { useThemeStore, type Theme } from '@/stores/theme';
import { useNotificationStore } from '@/stores/notification';

const businessSchema = z.object({
  name: z.string().trim().min(2, 'Ingresa la razón social.'),
  tradeName: z.string().trim(),
  taxId: z.string().trim().min(8, 'Ingresa el número fiscal.'),
  address: z.string().trim(),
  phone: z.string().trim(),
  email: z.string().trim().email('Ingresa un correo válido.'),
});

const prefixSchema = z.object({
  saleNumberPrefix: z.string().trim().min(1),
  purchaseNumberPrefix: z.string().trim().min(1),
});

const businessVarsSchema = z.object({
  clearanceDaysThreshold: z.number().int().min(1, 'Al menos 1 día.').max(365),
  importCostFactor: z.number().min(0, 'No puede ser negativo.').max(1, 'No puede superar 1.0'),
  fallbackExchangeRate: z.number().positive('Debe ser positivo.').max(100, 'Valor fuera de rango.'),
});

type BusinessValues = z.infer<typeof businessSchema>;
type PrefixValues = z.infer<typeof prefixSchema>;
type BusinessVarsValues = z.infer<typeof businessVarsSchema>;

export function SettingsPage() {
  const queryClient = useQueryClient();
  const push = useNotificationStore((s) => s.push);
  const theme = useThemeStore((state) => state.theme);
  const setTheme = useThemeStore((state) => state.setTheme);
  const business = useQuery({ queryKey: queryKeys.settings.business, queryFn: () => wailsClient.getBusinessInfo() });
  const preferences = useQuery({ queryKey: queryKeys.settings.preferences, queryFn: () => wailsClient.getPreferences() });
  const profile = useQuery({ queryKey: ['settings', 'profile'], queryFn: () => wailsClient.getLocalProfile() });

  const saveBusiness = useMutation({
    mutationFn: (values: BusinessValues) => wailsClient.updateBusinessInfo({ ...values, logo: business.data?.logo ?? '' }),
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: queryKeys.settings.business });
      push({ title: 'Información de empresa guardada', variant: 'success' });
    },
    onError: (err: unknown) => {
      push({ title: 'No se pudo guardar la empresa', description: err instanceof Error ? err.message : undefined, variant: 'destructive' });
    },
  });

  const savePrefixes = useMutation({
    mutationFn: async (values: PrefixValues) => {
      await wailsClient.updatePreference('sale_number_prefix', values.saleNumberPrefix);
      await wailsClient.updatePreference('purchase_number_prefix', values.purchaseNumberPrefix);
    },
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: queryKeys.settings.preferences });
      push({ title: 'Reglas operativas guardadas', variant: 'success' });
    },
    onError: (err: unknown) => {
      push({ title: 'No se pudo guardar las reglas', description: err instanceof Error ? err.message : undefined, variant: 'destructive' });
    },
  });

  const saveBusinessVars = useMutation({
    mutationFn: async (values: BusinessVarsValues) => {
      await wailsClient.updatePreference('clearance_days_threshold', String(values.clearanceDaysThreshold));
      await wailsClient.updatePreference('import_cost_factor', String(values.importCostFactor));
      await wailsClient.updatePreference('fallback_exchange_rate', String(values.fallbackExchangeRate));
    },
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: queryKeys.settings.preferences });
      push({ title: 'Variables de negocio guardadas', variant: 'success' });
    },
    onError: (err: unknown) => {
      push({ title: 'No se pudo guardar las variables', description: err instanceof Error ? err.message : undefined, variant: 'destructive' });
    },
  });

  const saveProfile = useMutation({
    mutationFn: (nextTheme: Theme) =>
      wailsClient.updateLocalProfile({
        name: profile.data?.name ?? '',
        theme: nextTheme,
        language: profile.data?.language ?? 'es-PE',
        dateFormat: profile.data?.dateFormat ?? 'DD/MM/YYYY',
        numberFormat: profile.data?.numberFormat ?? 'es-PE',
        decimalPlaces: profile.data?.decimalPlaces ?? 2,
        timezone: profile.data?.timezone ?? 'America/Lima',
      }),
    onSuccess: async () => {
      await queryClient.invalidateQueries({ queryKey: ['settings', 'profile'] });
      push({ title: 'Perfil guardado', variant: 'success' });
    },
    onError: (err: unknown) => {
      push({ title: 'No se pudo guardar el perfil', description: err instanceof Error ? err.message : undefined, variant: 'destructive' });
    },
  });

  if (business.isLoading || preferences.isLoading || profile.isLoading) {
    return (
      <div className="page-loader">
        <span className="spinner" />
      </div>
    );
  }

  return (
    <PageContainer>
      <PageHeader title="Configuración" subtitle="Administra la información y las reglas de tu empresa." />

      <Section title="Empresa" description="Estos datos aparecen en tus documentos y reportes.">
        {business.data && (
          <Card>
            <CardHeader>
              <CardTitle>Información fiscal</CardTitle>
              <CardDescription>Datos principales de la empresa activa.</CardDescription>
            </CardHeader>
            <CardContent>
              <Form<BusinessValues> schema={businessSchema} defaultValues={business.data} onSubmit={(values) => saveBusiness.mutate(values)}>
                <div className="form-grid">
                  <TextField name="name" label="Razón social" required />
                  <TextField name="tradeName" label="Nombre comercial" />
                  <TextField name="taxId" label="RUC o identificación fiscal" required />
                  <EmailField name="email" label="Correo" required />
                  <TextField name="phone" label="Teléfono" type="tel" />
                  <TextareaField name="address" label="Dirección" className="form-grid__wide" />
                </div>
                <div className="form-actions">
                  <Button type="submit" loading={saveBusiness.isPending}>
                    <Save /> Guardar
                  </Button>
                </div>
              </Form>
            </CardContent>
          </Card>
        )}
      </Section>

      <Section title="Operaciones" description="Numeración de documentos de venta, compra y asientos.">
        {preferences.data && (
          <Card>
            <CardHeader>
              <CardTitle>Prefijos documentales</CardTitle>
              <CardDescription>Se usan al generar números de venta y compra.</CardDescription>
            </CardHeader>
            <CardContent>
              <Form<PrefixValues>
                key={`${preferences.data.saleNumberPrefix}-${preferences.data.purchaseNumberPrefix}`}
                schema={prefixSchema}
                defaultValues={{
                  saleNumberPrefix: preferences.data.saleNumberPrefix,
                  purchaseNumberPrefix: preferences.data.purchaseNumberPrefix,
                }}
                onSubmit={(values) => savePrefixes.mutate(values)}
              >
                <div className="form-grid">
                  <TextField name="saleNumberPrefix" label="Prefijo de ventas" required />
                  <TextField name="purchaseNumberPrefix" label="Prefijo de compras" required />
                </div>
                <div className="form-actions">
                  <Button type="submit" loading={savePrefixes.isPending}>
                    <Save /> Guardar
                  </Button>
                </div>
              </Form>
            </CardContent>
          </Card>
        )}
      </Section>

      <Section title="Variables de negocio" description="Parámetros que afectan el cálculo de costos, remates y tipos de cambio.">
        {preferences.data && (
          <Card>
            <CardHeader>
              <CardTitle>Parámetros operativos</CardTitle>
              <CardDescription>Estos valores se usan en cálculos de importación, remate y proyecciones.</CardDescription>
            </CardHeader>
            <CardContent>
              <Form<BusinessVarsValues>
                key={`${preferences.data.clearanceDaysThreshold}-${preferences.data.importCostFactor}-${preferences.data.fallbackExchangeRate}`}
                schema={businessVarsSchema}
                defaultValues={{
                  clearanceDaysThreshold: preferences.data.clearanceDaysThreshold,
                  importCostFactor: preferences.data.importCostFactor,
                  fallbackExchangeRate: preferences.data.fallbackExchangeRate,
                }}
                onSubmit={(values) => saveBusinessVars.mutate(values)}
              >
                <div className="form-grid">
                  <NumberField
                    name="clearanceDaysThreshold"
                    label="Días para remate"
                    description="Un lote pasa a remate cuando supera estos días desde su ingreso."
                    min={1}
                    max={365}
                    required
                  />
                  <NumberField
                    name="importCostFactor"
                    label="Factor de costo de importación"
                    description="Factor multiplicador para calcular costos de importación (ej: 0.07 = 7%)."
                    min={0}
                    max={1}
                    step={0.01}
                    required
                  />
                  <NumberField
                    name="fallbackExchangeRate"
                    label="Tipo de cambio de respaldo"
                    description="Tipo de cambio USD→PEN a usar cuando no hay conexión a APIs."
                    min={0.01}
                    max={100}
                    step={0.01}
                    required
                  />
                </div>
                <div className="form-actions">
                  <Button type="submit" loading={saveBusinessVars.isPending}>
                    <Save /> Guardar
                  </Button>
                </div>
              </Form>
            </CardContent>
          </Card>
        )}
      </Section>

      <Section title="Apariencia y perfil" description="Preferencias de este dispositivo.">
        <Card>
          <CardHeader>
            <CardTitle>Tema y perfil local</CardTitle>
            <CardDescription>El tema se aplica al instante; guárdalo para recordarlo en tu perfil.</CardDescription>
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
                  onValueChange={(value) => setTheme((value ?? 'system') as Theme)}
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
                <span className="settings-row__label">Perfil</span>
                <strong>{profile.data?.name}</strong>
              </div>
              <div className="settings-row">
                <span className="settings-row__label">Zona horaria</span>
                <strong>{profile.data?.timezone}</strong>
              </div>
              <div className="settings-row">
                <span className="settings-row__label">Idioma</span>
                <strong>{profile.data?.language}</strong>
              </div>
              <div className="form-actions">
                <Button variant="outline" onClick={() => saveProfile.mutate(theme)} loading={saveProfile.isPending}>
                  <Save /> Guardar
                </Button>
              </div>
            </div>
          </CardContent>
        </Card>
      </Section>

      <SecuritySection />
    </PageContainer>
  );
}
