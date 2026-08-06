import { useCallback } from 'react';
import type { FieldPath, FieldValues } from 'react-hook-form';
import { AsyncSelectField, type SelectOption } from './SelectField';
import { customersService } from '@/services/customers';
import { suppliersService } from '@/services/suppliers';
import { productsService } from '@/services/products';
import { Currencies } from '@/constants/currencies';
import { Taxes } from '@/constants/taxes';
import { DocumentTypes } from '@/constants/countries';

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
    return [
      { value: 'wh-1', label: 'Principal' },
      { value: 'wh-2', label: 'Sucursal Norte' },
      { value: 'wh-3', label: 'Sucursal Sur' },
    ];
  }, []);
  return <AsyncSelectField {...props} loadOptions={load} />;
}

export function CategorySelectField<T extends FieldValues>(props: Omit<Parameters<typeof AsyncSelectField<T>>[0], 'loadOptions'>) {
  const load = useCallback(async (): Promise<SelectOption[]> => {
    return [
      { value: 'abarrotes', label: 'Abarrotes' },
      { value: 'bebidas', label: 'Bebidas' },
      { value: 'limpieza', label: 'Limpieza' },
      { value: 'panaderia', label: 'Panadería' },
      { value: 'lacteos', label: 'Lácteos' },
      { value: 'conservas', label: 'Conservas' },
      { value: 'snacks', label: 'Snacks' },
      { value: 'cuidado-personal', label: 'Cuidado Personal' },
    ];
  }, []);
  return <AsyncSelectField {...props} loadOptions={load} />;
}

export function BrandSelectField<T extends FieldValues>(props: Omit<Parameters<typeof AsyncSelectField<T>>[0], 'loadOptions'>) {
  const load = useCallback(async (): Promise<SelectOption[]> => {
    return ['Acme', 'Genérica', 'ProMax', 'EcoLine', 'Nordic', 'Andina', 'Pacífico', 'Premium'].map((b) => ({ value: b, label: b }));
  }, []);
  return <AsyncSelectField {...props} loadOptions={load} />;
}

export function TaxSelectField<T extends FieldValues>(props: Omit<Parameters<typeof AsyncSelectField<T>>[0], 'loadOptions'>) {
  const load = useCallback(async (): Promise<SelectOption[]> => {
    return Object.values(Taxes).map((t) => ({ value: t.code, label: `${t.shortName} (${(t.rate * 100).toFixed(0)}%)` }));
  }, []);
  return <AsyncSelectField {...props} loadOptions={load} />;
}

export function CurrencySelectField<T extends FieldValues>(props: Omit<Parameters<typeof AsyncSelectField<T>>[0], 'loadOptions'>) {
  const load = useCallback(async (): Promise<SelectOption[]> => {
    return Object.values(Currencies).map((c) => ({ value: c.code, label: `${c.code} — ${c.name}` }));
  }, []);
  return <AsyncSelectField {...props} loadOptions={load} />;
}

export function DocumentTypeSelectField<T extends FieldValues>(props: Omit<Parameters<typeof AsyncSelectField<T>>[0], 'loadOptions'>) {
  const load = useCallback(async (): Promise<SelectOption[]> => {
    return Object.values(DocumentTypes).map((d) => ({ value: d.code, label: d.name }));
  }, []);
  return <AsyncSelectField {...props} loadOptions={load} />;
}
