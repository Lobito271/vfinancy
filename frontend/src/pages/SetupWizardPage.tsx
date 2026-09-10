import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useQueryClient } from '@tanstack/react-query';
import { z } from 'zod';
import { AlertCircle } from 'lucide-react';
import { Button } from '@/components/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/card';
import { Form, TextField, EmailField, PasswordField } from '@/components/form';
import { wailsClient } from '@/services/bindings';
import { queryKeys } from '@/services/queryKeys';
import { Routes } from '@/constants/routes';

const setupSchema = z.object({
  name: z.string().trim().min(2, 'Ingresa la razón social.'),
  taxId: z
    .string()
    .regex(/^(10|20)\d{9}$/, 'RUC debe tener 11 dígitos e iniciar con 10 o 20.')
    .optional()
    .or(z.literal('')),
  email: z.string().email('Correo inválido.').optional().or(z.literal('')),
  fiscalAddress: z.string().trim().optional().or(z.literal('')),
  password: z.string().refine((value) => value === '' || value.length >= 8, 'Usa al menos 8 caracteres.'),
});

type SetupValues = z.infer<typeof setupSchema>;

export function SetupWizardPage() {
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const [error, setError] = useState('');
  const [saving, setSaving] = useState(false);

  async function submit(values: SetupValues) {
    setSaving(true);
    setError('');
    try {
      await wailsClient.setupWorkspace(values);
      await queryClient.invalidateQueries({ queryKey: queryKeys.setup });
      navigate(Routes.Dashboard, { replace: true });
    } catch (cause) {
      setError(cause instanceof Error ? cause.message : 'No se pudo completar la configuración.');
    } finally {
      setSaving(false);
    }
  }

  return (
    <div className="setup-page">
      <div className="setup-page__brand">vfinancy</div>
      <div className="setup-page__layout">
        <aside className="setup-page__intro">
          <span className="setup-page__eyebrow">configuración inicial</span>
          <h1>Tu operación empieza aquí.</h1>
          <p>Un solo paso para preparar tu espacio de trabajo.</p>
        </aside>
        <main className="setup-page__content">
          <Form<SetupValues> defaultValues={{ taxId: '', email: '', fiscalAddress: '', password: '' }} schema={setupSchema} onSubmit={submit}>
            {() => (
              <Card className="setup-card">
                <CardHeader>
                  <CardTitle>Tu empresa</CardTitle>
                  <CardDescription>
                    Un solo paso para preparar tu espacio de trabajo. Todo es editable después desde Configuración.
                  </CardDescription>
                </CardHeader>
                <CardContent className="setup-card__body">
                  <div className="setup-form-grid">
                    <TextField name="name" label="Razón social" required className="setup-form-grid__wide" autoComplete="organization" />
                    <TextField name="taxId" label="RUC (opcional)" autoComplete="off" />
                    <EmailField name="email" label="Correo electrónico (opcional)" autoComplete="email" />
                    <TextField name="fiscalAddress" label="Dirección fiscal (opcional)" className="setup-form-grid__wide" autoComplete="street-address" />
                    <PasswordField
                      name="password"
                      label="Contraseña (opcional)"
                      description="Mínimo 8 caracteres. Podrás agregarla después desde Configuración."
                      autoComplete="new-password"
                      className="setup-form-grid__wide"
                    />
                  </div>
                  {error && (
                    <p className="setup-error" role="alert">
                      <AlertCircle />
                      {error}
                    </p>
                  )}
                  <div className="setup-card__footer">
                    <Button type="submit" loading={saving}>
                      Comenzar
                    </Button>
                  </div>
                </CardContent>
              </Card>
            )}
          </Form>
        </main>
      </div>
    </div>
  );
}
