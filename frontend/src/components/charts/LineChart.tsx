import { lazy, Suspense } from 'react';
import type { ChartPoint } from '@/types/domain';
import { formatCurrency } from '@/utils/format';

const ReLineChart = lazy(() =>
  import('recharts').then((m) => ({
    default: function LazyLineChart({ data, height, formatY = formatCurrency }: LineChartProps) {
      const {
        LineChart: LC,
        Line,
        XAxis,
        YAxis,
        CartesianGrid,
        Tooltip,
        ResponsiveContainer,
      } = m;
      return (
        <ResponsiveContainer width="100%" height={height}>
          <LC data={data} margin={{ top: 5, right: 10, left: 0, bottom: 0 }}>
            <CartesianGrid strokeDasharray="3 3" stroke="var(--color-border)" vertical={false} />
            <XAxis
              dataKey="label"
              stroke="var(--color-muted-fg)"
              fontSize={12}
              tickLine={false}
              axisLine={false}
            />
            <YAxis
              stroke="var(--color-muted-fg)"
              fontSize={12}
              tickLine={false}
              axisLine={false}
              tickFormatter={formatY ? (v) => formatY(v as number) : undefined}
              width={70}
            />
            <Tooltip
              contentStyle={{
                backgroundColor: 'var(--color-surface)',
                border: '1px solid var(--color-border)',
                borderRadius: 8,
                fontSize: 12,
              }}
              formatter={formatY ? (v) => [formatY!(v as number), 'Valor'] : undefined}
            />
            <Line
              type="monotone"
              dataKey="value"
              stroke="var(--color-primary)"
              strokeWidth={2}
              dot={{ r: 3, fill: 'var(--color-primary)' }}
              activeDot={{ r: 5 }}
            />
          </LC>
        </ResponsiveContainer>
      );
    },
  }))
);

interface LineChartProps {
  data: ChartPoint[];
  height?: number;
  formatY?: (v: number) => string;
}

export function LineChart({ data, height = 280, formatY = formatCurrency }: LineChartProps) {
  return (
    <Suspense>
      <ReLineChart data={data} height={height} formatY={formatY} />
    </Suspense>
  );
}
