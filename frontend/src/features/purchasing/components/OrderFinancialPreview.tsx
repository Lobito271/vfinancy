import { useWatch, useFormContext } from 'react-hook-form';
import { NumberField } from '@/components/form';
import { formatCurrency } from '@/utils/format';

const IMPORT_SURCHARGE = 0.07;

function round2(value: number): number {
  return Math.round(value * 100) / 100;
}

interface OrderFinancialPreviewProps {
  showProfit?: boolean;
}

interface WatchedItem {
  unitPrice?: number;
  quantity?: number;
  discountPercent?: number;
}

export function OrderFinancialPreview({ showProfit = true }: OrderFinancialPreviewProps) {
  const { control } = useFormContext();
  const items = useWatch({ control, name: 'items' }) as WatchedItem[] | undefined;
  const salePricePEN = useWatch({ control, name: 'salePricePEN' }) as number | undefined;
  const exchangeRate = (useWatch({ control, name: 'exchangeRate' }) as number | undefined) ?? 0;

  const costUSD = round2(
    (items ?? []).reduce((sum, it) => {
      const base = (it.unitPrice ?? 0) * (it.quantity ?? 0);
      return sum + base - round2(base * ((it.discountPercent ?? 0) / 100));
    }, 0),
  );
  const realCost = round2(costUSD * (exchangeRate + IMPORT_SURCHARGE));
  const profit = showProfit ? round2((salePricePEN ?? 0) - realCost) : 0;

  return (
    <div className="space-y-3">
      <div className="grid grid-cols-1 gap-4 rounded-md border bg-muted/30 p-3 sm:grid-cols-4">
        <div>
          <div className="text-xs text-muted-foreground">Costo (USD)</div>
          <div className="text-base font-semibold tabular-nums">{formatCurrency(costUSD, 'USD')}</div>
        </div>
        <div className="flex flex-col gap-1">
          <div className="text-xs text-muted-foreground">T.C. (USD→PEN)</div>
          <NumberField
            name="exchangeRate"
            min={0.01}
            step={0.01}
            description="+ 7% recargo"
          />
        </div>
        <div>
          <div className="text-xs text-muted-foreground">Costo real (PEN)</div>
          <div className="text-base font-semibold tabular-nums">{formatCurrency(realCost)}</div>
        </div>
        {showProfit && (
          <div>
            <div className="text-xs text-muted-foreground">Utilidad proyectada</div>
            <div className={`text-base font-semibold tabular-nums ${profit < 0 ? 'text-destructive' : 'text-success'}`}>
              {formatCurrency(profit)}
            </div>
          </div>
        )}
      </div>
    </div>
  );
}
