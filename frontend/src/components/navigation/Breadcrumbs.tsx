import { Link } from 'react-router-dom';
import { ChevronRight } from 'lucide-react';
import { cx } from '@/utils/cx';
import { findRouteLabel } from '@/lib/nav';

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
    <nav aria-label="Breadcrumbs" className={cx('breadcrumbs', className)}>
      {items.map((it, i) => {
        const last = i === items.length - 1;
        return (
          <span key={it.to} className="breadcrumbs__item">
            {i > 0 && <ChevronRight className="breadcrumbs__sep" aria-hidden="true" />}
            {last ? (
              <span className="breadcrumbs__current">{it.label}</span>
            ) : (
              <Link to={it.to} className="breadcrumbs__link">
                {it.label}
              </Link>
            )}
          </span>
        );
      })}
    </nav>
  );
}
