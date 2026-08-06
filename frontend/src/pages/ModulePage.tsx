import { Users, type LucideIcon } from 'lucide-react';
import { PlaceholderPage } from './PlaceholderPage';

interface ModulePageProps {
  title: string;
  subtitle: string;
  description: string;
  phase: string;
  features: string[];
  icon: LucideIcon;
  mockStats?: { label: string; value: string }[];
}

export function ModulePage({ title, subtitle, description, phase, features, icon, mockStats }: ModulePageProps) {
  return (
    <PlaceholderPage
      title={title}
      subtitle={subtitle}
      description={description}
      phase={phase}
      features={features}
      icon={icon}
      mockStats={mockStats}
    />
  );
}

void Users;
