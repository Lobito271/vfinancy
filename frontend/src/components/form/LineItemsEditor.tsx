import { useFieldArray, useFormContext, type Path } from 'react-hook-form';
import { Button } from '@/components/button';
import { Grid } from '@/components/layout';
import { Trash2, Plus } from 'lucide-react';
import { DefaultCurrency, type CurrencyCode } from '@/constants/currencies';
import { formatCurrency } from '@/utils/format';
import { NumberField, MoneyField, PercentageField, SelectField } from './';

export interface LineItemFormValues {
  productId: string;
  quantity: number;
  unitPrice: number;
  discountPercent: number;
  discountAmount: number;
  taxRate: number;
  taxAmount: number;
  description: string;
}

export interface SaleLineItemFormValues extends LineItemFormValues {
  costSnapshot: number;
}

export interface ProductLineOption {
  value: string;
  label: string;
  unitCost: number;
  salePrice: number;
  taxRate: number;
}

type LineForm = { items: LineItemFormValues[] };

interface LineItemsEditorProps {
  products: ProductLineOption[];
  isSale?: boolean;
  currency?: CurrencyCode;
}

export function LineItemsEditor({ products, isSale = false, currency = DefaultCurrency }: LineItemsEditorProps) {
  const { control, watch, setValue } = useFormContext<LineForm>();
  const { fields, append, remove } = useFieldArray<LineForm, 'items', 'id'>({ control, name: 'items' });
  const rows = watch('items');

  const handleProductChange = (index: number, productId: string) => {
    const product = products.find((p) => p.value === productId);
    if (!product) return;
    if (!isSale) {
      setValue(`items.${index}.unitPrice` as Path<LineForm>, product.unitCost as never);
    }
    setValue(`items.${index}.taxRate` as Path<LineForm>, product.taxRate as never);
  };

  const newRow = (): LineItemFormValues => ({
    productId: '',
    quantity: 1,
    unitPrice: 0,
    discountPercent: 0,
    discountAmount: 0,
    taxRate: 0,
    taxAmount: 0,
    description: '',
  });

  const subtotal = rows?.reduce((s, r) => s + (r.unitPrice ?? 0) * (r.quantity ?? 0), 0) ?? 0;
  const discount = rows?.reduce((s, r) => s + (r.unitPrice ?? 0) * (r.quantity ?? 0) * ((r.discountPercent ?? 0) / 100), 0) ?? 0;
  const tax = rows?.reduce((s, r) => {
    const base = (r.unitPrice ?? 0) * (r.quantity ?? 0) * (1 - (r.discountPercent ?? 0) / 100);
    return s + base * ((r.taxRate ?? 0) / 100);
  }, 0) ?? 0;
  const total = subtotal - discount + tax;

  return (
    <div className="line-items">
      <div className="line-items">
        {fields.map((field, index) => (
          <div key={field.id} className="line-item">
            <div className="line-item__head">
              <div className="line-item__field">
                <SelectField
                  name={`items.${index}.productId` as Path<LineForm>}
                  label={`Producto ${index + 1}`}
                  required
                  options={products}
                  clearable={false}
                  onChange={(v) => handleProductChange(index, v)}
                />
              </div>
              <Button
                type="button"
                variant="ghost"
                size="icon-sm"
                className="line-item__remove"
                aria-label={`Quitar línea ${index + 1}`}
                onClick={() => remove(index)}
              >
                <Trash2 />
              </Button>
            </div>
            <Grid cols={4}>
              <NumberField name={`items.${index}.quantity` as Path<LineForm>} label="Cant." required min={0} step={0.01} />
              <MoneyField name={`items.${index}.unitPrice` as Path<LineForm>} label="P. unitario" currency={currency} />
              <PercentageField name={`items.${index}.discountPercent` as Path<LineForm>} label="Dscto %" />
              <PercentageField name={`items.${index}.taxRate` as Path<LineForm>} label="IGV %" />
            </Grid>
          </div>
        ))}
      </div>

      <Button type="button" variant="outline" size="sm" onClick={() => append(newRow())}>
        <Plus /> Agregar línea
      </Button>

      <div className="line-items-totals">
        <div className="totals-row">
          <span className="muted">Subtotal</span>
          <span className="tabular">{formatCurrency(subtotal, currency)}</span>
        </div>
        <div className="totals-row">
          <span className="muted">Descuento</span>
          <span className="tabular">{formatCurrency(discount, currency)}</span>
        </div>
        <div className="totals-row">
          <span className="muted">IGV</span>
          <span className="tabular">{formatCurrency(tax, currency)}</span>
        </div>
        <div className="totals-row totals-row--total">
          <span>Total</span>
          <span className="tabular">{formatCurrency(total, currency)}</span>
        </div>
      </div>
    </div>
  );
}
