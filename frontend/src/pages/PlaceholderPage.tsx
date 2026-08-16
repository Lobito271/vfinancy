import { type LucideIcon, Hammer, Clock, Database } from 'lucide-react';
import { PageContainer, PageHeader, Section } from '@/components/layout';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/card';
import { Badge } from '@/components/badge';
import { EmptyState } from '@/components/feedback';

interface PlaceholderPageProps {
  title: string;
  subtitle: string;
  icon: LucideIcon;
  description: string;
  phase: string;
  features: string[];
}

export function PlaceholderPage({ title, subtitle, icon: Icon, description, phase, features }: PlaceholderPageProps) {
  return (
    <PageContainer>
      <PageHeader
        title={title}
        subtitle={subtitle}
        actions={
          <Badge variant="secondary" className="gap-1">
            <Clock className="h-3 w-3" />
            {phase}
          </Badge>
        }
      />

      <Card>
        <CardHeader>
          <div className="flex items-start gap-3">
            <div className="flex h-10 w-10 shrink-0 items-center justify-center rounded-lg bg-primary/10 text-primary">
              <Icon className="h-5 w-5" aria-hidden="true" />
            </div>
            <div className="flex-1">
              <CardTitle>Acerca de este módulo</CardTitle>
              <CardDescription>{description}</CardDescription>
            </div>
          </div>
        </CardHeader>
        <CardContent>
          <p className="text-sm text-muted-foreground">
            Este módulo se implementará en una fase posterior. La interfaz y navegación ya están listas
            para que pueda familiarizarse con el sistema.
          </p>
        </CardContent>
      </Card>

      <Section title="Funcionalidades planificadas" description="Qué podrá hacer en este módulo">
        <Card>
          <CardContent className="pt-6">
            <ul className="grid gap-2 sm:grid-cols-2">
              {features.map((f) => (
                <li key={f} className="flex items-start gap-2 text-sm">
                  <span className="mt-1 h-1.5 w-1.5 shrink-0 rounded-full bg-primary" aria-hidden="true" />
                  {f}
                </li>
              ))}
            </ul>
          </CardContent>
        </Card>
      </Section>

      <Card>
        <CardContent className="pt-6">
          <EmptyState
            icon={Database}
            title="Sin datos aún"
            description="Los registros se mostrarán aquí cuando el módulo esté conectado a la base de datos."
          />
        </CardContent>
      </Card>

      <div className="flex items-center gap-1 rounded-lg border border-dashed p-4 text-xs text-muted-foreground">
        <Hammer className="h-3 w-3" />
        Módulo en construcción
      </div>
    </PageContainer>
  );
}
