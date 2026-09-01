import { lazy, Suspense } from 'react';
import type { ChartPoint } from '@/types/domain';

const ReBarChart = lazy(() =>
  import('recharts').then((m) => ({
    default: function LazyBarChart({ data, height, formatY, colors }: BarChartProps) {
      const {
        BarChart: BC,
        Bar,
        XAxis,
        YAxis,
        CartesianGrid,
        Tooltip,
        ResponsiveContainer,
        Cell,
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
                borderRadius: 8,
                fontSize: 12,
              }}
              formatter={formatY ? (v) => [formatY(v as number), 'Valor'] : undefined}
            />
            <Bar dataKey="value" radius={[4, 4, 0, 0]}>
              {data.map((point) => (
                <Cell key={point.label} fill={colors![data.indexOf(point) % colors!.length]} />
              ))}
            </Bar>
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
  colors?: string[];
}

const defaultColors = [
  'var(--color-primary)',
  'var(--color-info)',
  'var(--color-success)',
  'var(--color-warning)',
  'var(--color-destructive)',
];

export function BarChart({ data, height = 280, formatY, colors = defaultColors }: BarChartProps) {
  return (
    <Suspense>
      <ReBarChart data={data} height={height} formatY={formatY} colors={colors} />
    </Suspense>
  );
}
