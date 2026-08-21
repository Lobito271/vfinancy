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
          <Badge variant="secondary" className="badge--with-icon">
            <Clock className="icon-xs" />
            {phase}
          </Badge>
        }
      />

      <Card>
        <CardHeader>
          <div className="hstack hstack--start hstack--md">
            <div className="icon-tile">
              <Icon className="icon-md" aria-hidden="true" />
            </div>
            <div className="grow">
              <CardTitle>Acerca de este módulo</CardTitle>
              <CardDescription>{description}</CardDescription>
            </div>
          </div>
        </CardHeader>
        <CardContent>
          <p className="desc-text">
            Este módulo se implementará en una fase posterior. La interfaz y navegación ya están listas
            para que pueda familiarizarse con el sistema.
          </p>
        </CardContent>
      </Card>

      <Section title="Funcionalidades planificadas" description="Qué podrá hacer en este módulo">
        <Card>
          <CardContent className="card-content--padded">
            <ul className="feature-grid">
              {features.map((f) => (
                <li key={f} className="feature-item">
                  <span className="feature-dot" aria-hidden="true" />
                  {f}
                </li>
              ))}
            </ul>
          </CardContent>
        </Card>
      </Section>

      <Card>
        <CardContent className="card-content--padded">
          <EmptyState
            icon={Database}
            title="Sin datos aún"
            description="Los registros se mostrarán aquí cuando el módulo esté conectado a la base de datos."
          />
        </CardContent>
      </Card>

      <div className="construction-note">
        <Hammer className="icon-xs" />
        Módulo en construcción
      </div>
    </PageContainer>
  );
}
