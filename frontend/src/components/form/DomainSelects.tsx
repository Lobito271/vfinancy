import { useCallback } from 'react';
import type { FieldPath, FieldValues } from 'react-hook-form';
import { AsyncSelectField, type SelectOption } from './SelectField';
import { customersService } from '@/services/customers';
import { suppliersService } from '@/services/suppliers';
import { productsService } from '@/services/products';
import { wailsClient } from '@/services/bindings';

interface CustomerSelectFieldProps<T extends FieldValues> {
  name: FieldPath<T>;
  label?: string;
  description?: string;
  required?: boolean;
  className?: string;
  placeholder?: string;
}

export function CustomerSelectField<T extends FieldValues>(props: CustomerSelectFieldProps<T>) {
  const load = useCallback(async () => customersService.getOptions(), []);
  return <AsyncSelectField {...props} loadOptions={load} />;
}

interface SupplierSelectFieldProps<T extends FieldValues> {
  name: FieldPath<T>;
  label?: string;
  description?: string;
  required?: boolean;
  className?: string;
  placeholder?: string;
}

export function SupplierSelectField<T extends FieldValues>(props: SupplierSelectFieldProps<T>) {
  const load = useCallback(async () => suppliersService.getOptions(), []);
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
  const load = useCallback(async () => productsService.getOptions(), []);
  return <AsyncSelectField {...props} loadOptions={load} />;
}

export function WarehouseSelectField<T extends FieldValues>(props: Omit<Parameters<typeof AsyncSelectField<T>>[0], 'loadOptions'>) {
  const load = useCallback(async (): Promise<SelectOption[]> => {
    const warehouses = await wailsClient.listWarehouses();
    return warehouses.map((w) => ({ value: w.id, label: `${w.code} — ${w.name}` }));
  }, []);
  return <AsyncSelectField {...props} loadOptions={load} />;
}
