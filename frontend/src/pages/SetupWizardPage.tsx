import { useState } from 'react';
import { useNavigate } from 'react-router-dom';
import { useQueryClient } from '@tanstack/react-query';
import { z } from 'zod';
import { Button } from '@/components/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/card';
import { Form, TextField, NumberField, PasswordField, SelectField } from '@/components/form';
import { AlertCircle, ArrowLeft, ArrowRight, Plus } from 'lucide-react';
import { wailsClient } from '@/services/bindings';
import type { SetupWorkspaceRequest } from '@/services/wails-types';

const setupSchema = z.object({
  legalName: z.string().trim().min(2, 'Ingresa la razón social.'),
  tradeName: z.string().trim().min(2, 'Ingresa el nombre comercial.'),
  code: z.string().trim().min(2, 'Ingresa un código de empresa.'),
  taxId: z.string().trim().min(8, 'Ingresa el número de identificación fiscal.'),
  address: z.string().trim().min(3, 'Ingresa una dirección.'),
  phone: z.string().trim(),
  email: z.string().trim().email('Ingresa un correo válido.'),
  countryCode: z.string().min(2),
  functionalCurrency: z.string().min(3),
  timezone: z.string().min(1),
  fiscalYearStartMonth: z.number().int().min(1).max(12),
  profileName: z.string().trim().min(2, 'Ingresa tu nombre.'),
  password: z.string().refine((value) => value === '' || value.length >= 8, 'Usa al menos 8 caracteres.'),
});

type SetupValues = z.infer<typeof setupSchema>;

const defaultValues: SetupValues = {
  legalName: '',
  tradeName: '',
  code: '',
  taxId: '',
  address: '',
  phone: '',
  email: '',
  countryCode: 'PE',
  functionalCurrency: 'PEN',
  timezone: 'America/Lima',
  fiscalYearStartMonth: 1,
  profileName: '',
  password: '',
};

const steps = [
  { title: 'Tu empresa', description: 'Identifica el negocio que vas a administrar.' },
  { title: 'Configuración regional', description: 'Define cómo se registrarán tus operaciones.' },
  { title: 'Tu acceso', description: 'Crea el perfil local para entrar a vfinancy.' },
];

export function SetupWizardPage() {
  const navigate = useNavigate();
  const queryClient = useQueryClient();
  const [step, setStep] = useState(0);
  const [error, setError] = useState('');
  const [saving, setSaving] = useState(false);

  async function submit(values: SetupValues) {
    setSaving(true);
    setError('');
    try {
      await wailsClient.setupWorkspace({
        code: values.code,
        legalName: values.legalName,
        tradeName: values.tradeName,
        taxId: values.taxId,
        address: values.address,
        phone: values.phone,
        email: values.email,
        countryCode: values.countryCode,
        functionalCurrency: values.functionalCurrency,
        timezone: values.timezone,
        fiscalYearStartMonth: values.fiscalYearStartMonth,
        profileName: values.profileName,
        password: values.password,
      } satisfies SetupWorkspaceRequest);
      await queryClient.invalidateQueries({ queryKey: ['setup'] });
      navigate('/', { replace: true });
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
          <p>Completa estos datos una sola vez para preparar tu espacio de trabajo.</p>
          <ol className="setup-steps">
            {steps.map((item, index) => (
              <li className={index === step ? 'setup-step setup-step--active' : 'setup-step'} key={item.title}>
                <span className="setup-step__number">{index + 1}</span>
                <span><strong>{item.title}</strong><small>{item.description}</small></span>
              </li>
            ))}
          </ol>
        </aside>
        <main className="setup-page__content">
          <Form<SetupValues> defaultValues={defaultValues} schema={setupSchema} onSubmit={submit}>
            {(form) => (
              <Card className="setup-card">
                <CardHeader>
                  <span className="setup-card__counter">paso {step + 1} de {steps.length}</span>
                  <CardTitle>{steps[step].title}</CardTitle>
                  <CardDescription>{steps[step].description}</CardDescription>
                </CardHeader>
                <CardContent className="setup-card__body">
                  {step === 0 && <div className="setup-form-grid">
                    <TextField name="legalName" label="Razón social" required autoComplete="organization" />
                    <TextField name="tradeName" label="Nombre comercial" required />
                    <TextField name="code" label="Código interno" required description="Ejemplo: ACME" />
                    <TextField name="taxId" label="RUC o identificación fiscal" required description="Solo dígitos." />
                    <TextField name="address" label="Dirección" required className="setup-form-grid__wide" />
                    <TextField name="phone" label="Teléfono" type="tel" />
                    <TextField name="email" label="Correo" type="email" required />
                  </div>}
                  {step === 1 && <div className="setup-form-grid">
                    <SelectField name="countryCode" label="País" required options={[{ value: 'PE', label: 'Perú' }]} clearable={false} />
                    <SelectField name="functionalCurrency" label="Moneda funcional" required options={[{ value: 'PEN', label: 'PEN · Sol peruano' }, { value: 'USD', label: 'USD · Dólar estadounidense' }]} clearable={false} />
                    <SelectField name="timezone" label="Zona horaria" required options={[{ value: 'America/Lima', label: 'America/Lima' }]} clearable={false} />
                    <NumberField name="fiscalYearStartMonth" label="Mes de inicio fiscal" required min={1} max={12} />
                  </div>}
                  {step === 2 && <div className="setup-form-grid">
                    <TextField name="profileName" label="Nombre del perfil" required className="setup-form-grid__wide" autoComplete="name" />
                    <PasswordField name="password" label="Contraseña (opcional)" description="Podrás agregarla después desde Configuración." autoComplete="new-password" className="setup-form-grid__wide" />
                  </div>}
                  {error && <p className="setup-error" role="alert"><AlertCircle />{error}</p>}
                  <div className="setup-card__footer">
                    {step > 0 ? <Button type="button" variant="ghost" onClick={() => setStep((current) => current - 1)}><ArrowLeft /> Atrás</Button> : <span />}
                    {step < steps.length - 1 ? <Button type="button" onClick={async () => {
                      const fields = step === 0 ? ['legalName', 'tradeName', 'code', 'taxId', 'address', 'phone', 'email'] : ['countryCode', 'functionalCurrency', 'timezone', 'fiscalYearStartMonth'];
                      if (await form.trigger(fields as Array<keyof SetupValues>)) setStep((current) => current + 1);
                    }}>Continuar <ArrowRight /></Button> : <Button type="submit" loading={saving}>Crear empresa <Plus /></Button>}
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
