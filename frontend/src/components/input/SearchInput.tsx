import * as React from 'react';
import { Search, X } from 'lucide-react';
import { cn } from '@/lib/utils';
import { Input } from './Input';
import { Button } from '@/components/button';

export interface SearchInputProps extends Omit<React.InputHTMLAttributes<HTMLInputElement>, 'type'> {
  onClear?: () => void;
}

export const SearchInput = React.forwardRef<HTMLInputElement, SearchInputProps>(
  ({ className, value, onClear, ...props }, ref) => (
    <div className={cn('relative', className)}>
      <Search
        className="pointer-events-none absolute left-3 top-1/2 size-4 -translate-y-1/2 text-muted-foreground"
        aria-hidden="true"
      />
      <Input
        ref={ref}
        type="search"
        value={value}
        className="pl-9 pr-9"
        aria-label="Buscar"
        {...props}
      />
      {value && (
        <Button
          type="button"
          variant="ghost"
          size="icon-sm"
          onClick={onClear}
          className="absolute right-1 top-1/2 -translate-y-1/2"
          aria-label="Limpiar búsqueda"
        >
          <X />
        </Button>
      )}
    </div>
  ),
);
SearchInput.displayName = 'SearchInput';
