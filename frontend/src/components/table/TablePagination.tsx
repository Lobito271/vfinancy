import { useMemo } from 'react';
import { ChevronLeft, ChevronRight } from 'lucide-react';
import { Button } from '@/components/button';
import {
  Select,
  SelectContent,
  SelectItem,
  SelectTrigger,
  SelectValue,
} from '@/components/select';
import { t } from '@/locales';

interface TablePaginationProps {
  page: number;
  pageSize: number;
  total: number;
  onPageChange: (page: number) => void;
  onPageSizeChange: (n: number) => void;
  pageSizeOptions?: number[];
}

export function TablePagination({
  page,
  pageSize,
  total,
  onPageChange,
  onPageSizeChange,
  pageSizeOptions = [10, 25, 50, 100],
}: TablePaginationProps) {
  const totalPages = Math.max(1, Math.ceil(total / pageSize));
  const start = total === 0 ? 0 : (page - 1) * pageSize + 1;
  const end = Math.min(page * pageSize, total);

  const pageNumbers = useMemo(() => {
    const pages: (number | 'ellipsis')[] = [];
    const window = 1;
    for (let i = 1; i <= totalPages; i++) {
      if (i === 1 || i === totalPages || (i >= page - window && i <= page + window)) {
        pages.push(i);
      } else if (pages[pages.length - 1] !== 'ellipsis') {
        pages.push('ellipsis');
      }
    }
    return pages;
  }, [page, totalPages]);

  return (
    <div className="flex flex-col gap-3 px-4 py-3 sm:flex-row sm:items-center sm:justify-between">
      <div className="flex items-center gap-2 text-sm text-muted-foreground">
        <span>{t('common.showing')}</span>
        <span className="font-medium text-foreground">{start}</span>
        <span>–</span>
        <span className="font-medium text-foreground">{end}</span>
        <span>{t('common.of')}</span>
        <span className="font-medium text-foreground">{total}</span>
        <span>{t('common.rows')}</span>
      </div>
      <div className="flex items-center gap-6">
        <div className="flex items-center gap-2">
          <span className="text-sm text-muted-foreground">Por página</span>
          <Select
            value={String(pageSize)}
            onValueChange={(v) => onPageSizeChange(Number(v))}
          >
            <SelectTrigger className="h-8 w-[72px]">
              <SelectValue />
            </SelectTrigger>
            <SelectContent>
              {pageSizeOptions.map((n) => (
                <SelectItem key={n} value={String(n)}>
                  {n}
                </SelectItem>
              ))}
            </SelectContent>
          </Select>
        </div>
        <div className="flex items-center gap-1">
          <Button
            variant="outline"
            size="icon-sm"
            onClick={() => onPageChange(Math.max(1, page - 1))}
            disabled={page <= 1}
            aria-label={t('common.previous')}
          >
            <ChevronLeft />
          </Button>
          {pageNumbers.map((p, i) =>
            p === 'ellipsis' ? (
              <span key={`e-${i}`} className="px-2 text-sm text-muted-foreground">
                …
              </span>
            ) : (
              <Button
                key={p}
                variant={p === page ? 'default' : 'ghost'}
                size="icon-sm"
                onClick={() => onPageChange(p)}
                aria-current={p === page ? 'page' : undefined}
                className="w-8"
              >
                {p}
              </Button>
            ),
          )}
          <Button
            variant="outline"
            size="icon-sm"
            onClick={() => onPageChange(Math.min(totalPages, page + 1))}
            disabled={page >= totalPages}
            aria-label={t('common.next')}
          >
            <ChevronRight />
          </Button>
        </div>
      </div>
    </div>
  );
}
