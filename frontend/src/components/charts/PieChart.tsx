import {
  PieChart as RePieChart,
  Pie,
  Cell,
  Tooltip as ReTooltip,
  ResponsiveContainer,
  Legend,
} from 'recharts';
import type { ChartPoint } from '@/data/mock';

interface PieChartProps {
  data: ChartPoint[];
  height?: number;
  colors?: string[];
}

const defaultColors = [
  'hsl(var(--primary))',
  'hsl(var(--info))',
  'hsl(var(--success))',
  'hsl(var(--warning))',
  'hsl(var(--destructive))',
];

export function PieChart({ data, height = 280, colors = defaultColors }: PieChartProps) {
  return (
    <ResponsiveContainer width="100%" height={height}>
      <RePieChart>
        <Pie
          data={data}
          dataKey="value"
          nameKey="label"
          innerRadius={60}
          outerRadius={100}
          paddingAngle={2}
        >
          {data.map((_, i) => (
            <Cell key={i} fill={colors[i % colors.length]} />
          ))}
        </Pie>
        <ReTooltip
          contentStyle={{
            backgroundColor: 'hsl(var(--popover))',
            border: '1px solid hsl(var(--border))',
            borderRadius: 8,
            fontSize: 12,
          }}
        />
        <Legend
          verticalAlign="bottom"
          height={36}
          iconType="circle"
          iconSize={8}
          wrapperStyle={{ fontSize: 12 }}
        />
      </RePieChart>
    </ResponsiveContainer>
  );
}
