import { lazy, Suspense } from 'react';
import { Currencies } from '@/constants/currencies';
import { formatCurrency } from '@/utils/format';

// Axis ticks stay short ("S/ 12.3 mil"); the tooltip carries the exact
// amount.
const compact = new Intl.NumberFormat('es-PE', { notation: 'compact', maximumFractionDigits: 1 });
const axisLabel = (value: number) => `${Currencies.PEN.symbol} ${compact.format(value)}`;

export interface ProfitChartPoint {
  label: string;
  sales: number;
  base: number;
  extra: number;
  profit: number;
}

const ReProfitChart = lazy(() =>
  import('recharts').then((m) => ({
    default: function LazyProfitChart({ data, height, single }: ProfitChartProps) {
      const {
        BarChart: BC,
        Bar,
        Line,
        XAxis,
        YAxis,
        CartesianGrid,
        Tooltip,
        Legend,
        ResponsiveContainer,
      } = m;
      const tooltipStyle = {
        backgroundColor: 'var(--color-surface)',
        border: '1px solid var(--color-border)',
        borderRadius: 0,
        fontSize: 12,
      } as const;
      return (
        <ResponsiveContainer width="100%" height={height}>
          <BC data={data} margin={{ top: 5, right: 10, left: 0, bottom: 0 }}>
            <CartesianGrid strokeDasharray="3 3" stroke="var(--color-border)" vertical={false} />
            <XAxis dataKey="label" stroke="var(--text-muted)" fontSize={12} tickLine={false} axisLine={false} />
            <YAxis
              stroke="var(--text-muted)"
              fontSize={12}
              tickLine={false}
              axisLine={false}
              tickFormatter={(v: number) => axisLabel(v)}
              width={72}
            />
            <Tooltip
              contentStyle={tooltipStyle}
              formatter={(value, name) => [formatCurrency(Number(value)), String(name)]}
            />
            {!single && (
              <Legend
                wrapperStyle={{ fontSize: 12 }}
                formatter={(value: string) => <span style={{ color: 'var(--text-secondary)' }}>{value}</span>}
              />
            )}
            <Bar dataKey="sales" name="Precio de venta" fill="var(--color-info)" maxBarSize={single ? 72 : 26} />
            <Bar dataKey="base" name="Costo base" fill="var(--color-primary)" stackId="costo" maxBarSize={single ? 72 : 26} />
            <Bar dataKey="extra" name="Costos extra" fill="var(--color-warning)" stackId="costo" maxBarSize={single ? 72 : 26} />
            <Line
              type="monotone"
              dataKey="profit"
              name="Utilidad"
              stroke="var(--color-success)"
              strokeWidth={2}
              dot={{ r: 2 }}
            />
          </BC>
        </ResponsiveContainer>
      );
    },
  }))
);

interface ProfitChartProps {
  data: ProfitChartPoint[];
  height?: number;
  /** One period: the legend is redundant because the tooltip names every bar. */
  single?: boolean;
}

export function ProfitChart({ data, height = 300, single = false }: ProfitChartProps) {
  return (
    <Suspense>
      <ReProfitChart data={data} height={height} single={single} />
    </Suspense>
  );
}
