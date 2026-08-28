import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { z } from 'zod';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/card';
import { Button } from '@/components/button';
import { EmailField, Form, NumberField, TextField, TextareaField } from '@/components/form';
import { Label } from '@/components/input';
import { Select, SelectContent, SelectItem, SelectTrigger, SelectValue } from '@/components/select';
import { PageContainer, PageHeader, Section } from '@/components/layout';
import { Save } from 'lucide-react';
import { wailsClient } from '@/services/bindings';
import { useThemeStore, type Theme } from '@/stores/theme';

const businessSchema = z.object({
  name: z.string().trim().min(2, 'Ingresa la razón social.'),
  tradeName: z.string().trim(),
  taxId: z.string().trim().min(8, 'Ingresa el número fiscal.'),
  address: z.string().trim(),
  phone: z.string().trim(),
  email: z.string().trim().email('Ingresa un correo válido.'),
});

const preferenceSchema = z.object({
  clearanceDays: z.coerce.number().int().positive(),
  clearanceWarningDays: z.coerce.number().int().nonnegative(),
  saleNumberPrefix: z.string().trim().min(1),
  purchaseNumberPrefix: z.string().trim().min(1),
  journalNumberPrefix: z.string().trim().min(1),
});

type BusinessValues = z.infer<typeof businessSchema>;
type PreferenceValues = z.infer<typeof preferenceSchema>;

export function SettingsPage() {
  const queryClient = useQueryClient();
  const theme = useThemeStore((state) => state.theme);
  const setTheme = useThemeStore((state) => state.setTheme);
  const business = useQuery({ queryKey: ['settings', 'business'], queryFn: () => wailsClient.getBusinessInfo() });
  const preferences = useQuery({ queryKey: ['settings', 'preferences'], queryFn: () => wailsClient.getPreferences() });
  const profile = useQuery({ queryKey: ['settings', 'profile'], queryFn: () => wailsClient.getLocalProfile() });
  const saveBusiness = useMutation({
    mutationFn: (values: BusinessValues) => wailsClient.updateBusinessInfo({ ...values, logo: business.data?.logo ?? '' }),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['settings', 'business'] }),
  });
  const savePreferences = useMutation({
    mutationFn: async (values: PreferenceValues) => {
      await wailsClient.updatePreference('clearance_days', String(values.clearanceDays));
      await wailsClient.updatePreference('clearance_warning_days', String(values.clearanceWarningDays));
      await wailsClient.updatePreference('sale_number_prefix', values.saleNumberPrefix);
      await wailsClient.updatePreference('purchase_number_prefix', values.purchaseNumberPrefix);
      await wailsClient.updatePreference('journal_number_prefix', values.journalNumberPrefix);
    },
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['settings', 'preferences'] }),
  });
  const saveProfile = useMutation({
    mutationFn: (nextTheme: Theme) => wailsClient.updateLocalProfile({
      name: profile.data?.name ?? '', theme: nextTheme, language: profile.data?.language ?? 'es-PE',
      dateFormat: profile.data?.dateFormat ?? 'DD/MM/YYYY', numberFormat: profile.data?.numberFormat ?? 'es-PE',
      decimalPlaces: profile.data?.decimalPlaces ?? 2, timezone: profile.data?.timezone ?? 'America/Lima',
    }),
    onSuccess: () => queryClient.invalidateQueries({ queryKey: ['settings', 'profile'] }),
  });

  if (business.isLoading || preferences.isLoading || profile.isLoading) {
    return <div className="page-loader"><span className="loader-bar" /></div>;
  }

  return (
    <PageContainer>
      <PageHeader title="Configuración" subtitle="Administra la información y las reglas de tu empresa." />
      <Section title="Empresa" description="Estos datos aparecen en tus documentos y reportes.">
        {business.data && <Card>
          <CardHeader><CardTitle>Información fiscal</CardTitle><CardDescription>Datos principales de la empresa activa.</CardDescription></CardHeader>
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
              <div className="form-actions"><Button type="submit" loading={saveBusiness.isPending}><Save /> Guardar empresa</Button></div>
            </Form>
          </CardContent>
        </Card>}
      </Section>
      <Section title="Reglas operativas" description="Valores usados para alertas y numeración.">
        {preferences.data && <Card>
          <CardHeader><CardTitle>Preferencias de operación</CardTitle><CardDescription>Configura el remate de inventario y los prefijos documentales.</CardDescription></CardHeader>
          <CardContent>
            <Form<PreferenceValues> schema={preferenceSchema} defaultValues={{
              clearanceDays: preferences.data.clearanceDays, clearanceWarningDays: preferences.data.clearanceWarningDays,
              saleNumberPrefix: preferences.data.saleNumberPrefix, purchaseNumberPrefix: preferences.data.purchaseNumberPrefix,
              journalNumberPrefix: preferences.data.journalNumberPrefix,
            }} onSubmit={(values) => savePreferences.mutate(values)}>
              <div className="form-grid">
                <NumberField name="clearanceDays" label="Días hasta remate" required min={1} />
                <NumberField name="clearanceWarningDays" label="Aviso anticipado" required min={0} />
                <TextField name="saleNumberPrefix" label="Prefijo de ventas" required />
                <TextField name="purchaseNumberPrefix" label="Prefijo de compras" required />
                <TextField name="journalNumberPrefix" label="Prefijo de asientos" required />
              </div>
              <div className="form-actions"><Button type="submit" loading={savePreferences.isPending}><Save /> Guardar reglas</Button></div>
            </Form>
          </CardContent>
        </Card>}
      </Section>
      <Section title="Perfil local" description="Preferencias de este dispositivo.">
        <Card>
          <CardHeader><CardTitle>Apariencia y región</CardTitle><CardDescription>Solo afecta a tu perfil local.</CardDescription></CardHeader>
          <CardContent>
            <div className="settings-preferences">
              <div className="field"><Label htmlFor="theme">Tema</Label><Select value={theme} onValueChange={(value) => { const next = value as Theme; setTheme(next); saveProfile.mutate(next); }}><SelectTrigger id="theme"><SelectValue /></SelectTrigger><SelectContent><SelectItem value="light">Claro</SelectItem><SelectItem value="dark">Oscuro</SelectItem><SelectItem value="system">Sistema</SelectItem></SelectContent></Select></div>
              <div className="settings-preferences__item"><span className="settings-preferences__label">Perfil</span><strong>{profile.data?.name}</strong></div>
              <div className="settings-preferences__item"><span className="settings-preferences__label">Zona horaria</span><strong>{profile.data?.timezone}</strong></div>
              <div className="settings-preferences__item"><span className="settings-preferences__label">Idioma</span><strong>{profile.data?.language}</strong></div>
            </div>
          </CardContent>
        </Card>
      </Section>
    </PageContainer>
  );
}
