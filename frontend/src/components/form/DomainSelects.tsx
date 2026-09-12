import { useCallback } from 'react';
import type { FieldPath, FieldValues } from 'react-hook-form';
import { AsyncSelectField } from './SelectField';
import { customersService } from '@/services/customers';
import { productsService } from '@/services/products';

interface CustomerSelectFieldProps<T extends FieldValues> {
  name: FieldPath<T>;
  label?: string;
  description?: string;
  required?: boolean;
  className?: string;
  placeholder?: string;
}

export function CustomerSelectField<T extends FieldValues>(props: CustomerSelectFieldProps<T>) {
  const load = useCallback(async () => {
    const options = await customersService.getOptions();
    return options.map((c) => ({ value: c.id, label: c.businessName }));
  }, []);
  return <AsyncSelectField {...props} loadOptions={load} />;
}

interface ProductSelectFieldProps<T extends FieldValues> {
  name: FieldPath<T>;
  label?: string;
  description?: string;
  required?: boolean;
  className?: string;
  placeholder?: string;
}

export function ProductSelectField<T extends FieldValues>(props: ProductSelectFieldProps<T>) {
  const load = useCallback(async () => {
    const options = await productsService.getOptions();
    return options.map((p) => ({ value: p.id, label: `${p.sku} — ${p.description}` }));
  }, []);
  return <AsyncSelectField {...props} loadOptions={load} />;
}
