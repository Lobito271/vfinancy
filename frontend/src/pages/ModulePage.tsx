import type { LucideIcon } from 'lucide-react';
import { PlaceholderPage } from './PlaceholderPage';

interface ModulePageProps {
  title: string;
  subtitle: string;
  description: string;
  phase: string;
  features: string[];
  icon: LucideIcon;
}

export function ModulePage({ title, subtitle, description, phase, features, icon }: ModulePageProps) {
  return (
    <PlaceholderPage
      title={title}
      subtitle={subtitle}
      description={description}
      phase={phase}
      features={features}
      icon={icon}
    />
  );
}
