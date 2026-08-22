import * as React from 'react';
import { Search, X } from 'lucide-react';
import { cx } from '@/utils/cx';
import { Input } from './Input';
import { Button } from '@/components/button';

export interface SearchInputProps extends Omit<React.InputHTMLAttributes<HTMLInputElement>, 'type'> {
  onClear?: () => void;
}

export const SearchInput = React.forwardRef<HTMLInputElement, SearchInputProps>(
  ({ className, value, onClear, ...props }, ref) => (
    <div className={cx('search-box', className)}>
      <Search className="search-box__icon" aria-hidden="true" />
      <Input
        ref={ref}
        type="search"
        value={value}
        aria-label="Buscar"
        {...props}
      />
      {value && (
        <Button
          type="button"
          variant="ghost"
          size="icon-sm"
          onClick={onClear}
          className="search-box__clear"
          aria-label="Limpiar búsqueda"
        >
          <X />
        </Button>
      )}
    </div>
  ),
);
SearchInput.displayName = 'SearchInput';
