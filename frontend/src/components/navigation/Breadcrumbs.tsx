import { Link } from 'react-router-dom';
import { ChevronRight } from 'lucide-react';
import { cn } from '@/lib/utils';
import { navRoutes, findRouteLabel } from '@/lib/nav';

interface BreadcrumbsProps {
  path: string;
  className?: string;
}

export function Breadcrumbs({ path, className }: BreadcrumbsProps) {
  const segments = path.split('/').filter(Boolean);
  const items: { to: string; label: string }[] = [
    { to: '/', label: 'Inicio' },
  ];

  if (segments.length > 0) {
    let acc = '';
    for (const seg of segments) {
      acc += `/${seg}`;
      items.push({ to: acc, label: findRouteLabel(acc) });
    }
  }

  return (
    <nav aria-label="Breadcrumbs" className={cn('flex items-center text-sm text-muted-foreground', className)}>
      {items.map((it, i) => {
        const last = i === items.length - 1;
        return (
          <span key={it.to} className="flex items-center">
            {i > 0 && <ChevronRight className="mx-1 h-3.5 w-3.5" aria-hidden="true" />}
            {last ? (
              <span className="font-medium text-foreground">{it.label}</span>
            ) : (
              <Link to={it.to} className="hover:text-foreground hover:underline">
                {it.label}
              </Link>
            )}
          </span>
        );
      })}
    </nav>
  );
}

void navRoutes;
