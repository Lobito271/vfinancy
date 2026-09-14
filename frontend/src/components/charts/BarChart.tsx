import { lazy, Suspense } from 'react';
import type { ChartPoint } from '@/types/domain';

const ReBarChart = lazy(() =>
  import('recharts').then((m) => ({
    default: function LazyBarChart({ data, height, formatY }: BarChartProps) {
      const {
        BarChart: BC,
        Bar,
        XAxis,
        YAxis,
        CartesianGrid,
        Tooltip,
        ResponsiveContainer,
      } = m;
      return (
        <ResponsiveContainer width="100%" height={height}>
          <BC data={data} margin={{ top: 5, right: 10, left: 0, bottom: 0 }}>
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
                borderRadius: 0,
                fontSize: 12,
              }}
              formatter={formatY ? (v) => [formatY(v as number), 'Valor'] : undefined}
            />
            <Bar dataKey="value" fill="var(--color-primary)" />
          </BC>
        </ResponsiveContainer>
      );
    },
  }))
);

interface BarChartProps {
  data: ChartPoint[];
  height?: number;
  formatY?: (v: number) => string;
}

export function BarChart({ data, height = 280, formatY }: BarChartProps) {
  return (
    <Suspense>
      <ReBarChart data={data} height={height} formatY={formatY} />
    </Suspense>
  );
}