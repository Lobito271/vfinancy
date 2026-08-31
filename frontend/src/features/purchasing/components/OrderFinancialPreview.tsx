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
    <div className="stack">
      <div className="fact-grid">
        <div>
          <div className="fact-grid__label">Costo (USD)</div>
          <div className="fact-grid__value">{formatCurrency(costUSD, 'USD')}</div>
        </div>
        <div className="stack stack--tight">
          <div className="fact-grid__label">T.C. (USD→PEN)</div>
          <NumberField
            name="exchangeRate"
            min={0.01}
            step={0.01}
            description="+ 7% recargo"
          />
        </div>
        <div>
          <div className="fact-grid__label">Costo real (PEN)</div>
          <div className="fact-grid__value">{formatCurrency(realCost)}</div>
        </div>
        {showProfit && (
          <div>
            <div className="fact-grid__label">Utilidad proyectada</div>
            <div className={`fact-grid__value ${profit < 0 ? 'text-destructive' : 'text-success'}`}>
              {formatCurrency(profit)}
            </div>
          </div>
        )}
      </div>
    </div>
  );
}
