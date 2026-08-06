export const ColorTokens = {
  brand: {
    primary: 'hsl(222 47% 11%)',
    'primary-foreground': 'hsl(210 40% 98%)',
  },
  neutral: {
    background: 'hsl(0 0% 100%)',
    foreground: 'hsl(222 47% 11%)',
    muted: 'hsl(210 40% 96%)',
    'muted-foreground': 'hsl(215 16% 47%)',
    border: 'hsl(214 32% 91%)',
    input: 'hsl(214 32% 91%)',
    ring: 'hsl(222 47% 11%)',
  },
  surface: {
    card: 'hsl(0 0% 100%)',
    'card-foreground': 'hsl(222 47% 11%)',
    popover: 'hsl(0 0% 100%)',
    'popover-foreground': 'hsl(222 47% 11%)',
    accent: 'hsl(210 40% 96%)',
    'accent-foreground': 'hsl(222 47% 11%)',
    secondary: 'hsl(210 40% 96%)',
    'secondary-foreground': 'hsl(222 47% 11%)',
  },
  semantic: {
    success: 'hsl(142 71% 45%)',
    'success-foreground': 'hsl(0 0% 100%)',
    warning: 'hsl(38 92% 50%)',
    'warning-foreground': 'hsl(0 0% 100%)',
    destructive: 'hsl(0 84% 60%)',
    'destructive-foreground': 'hsl(210 40% 98%)',
    info: 'hsl(199 89% 48%)',
    'info-foreground': 'hsl(0 0% 100%)',
  },
} as const;

export const DarkColorTokens = {
  brand: {
    primary: 'hsl(210 40% 98%)',
    'primary-foreground': 'hsl(222 47% 11%)',
  },
  neutral: {
    background: 'hsl(222 47% 6%)',
    foreground: 'hsl(210 40% 98%)',
    muted: 'hsl(217 33% 11%)',
    'muted-foreground': 'hsl(215 20% 65%)',
    border: 'hsl(217 33% 17%)',
    input: 'hsl(217 33% 17%)',
    ring: 'hsl(213 27% 84%)',
  },
  surface: {
    card: 'hsl(222 47% 8%)',
    'card-foreground': 'hsl(210 40% 98%)',
    popover: 'hsl(222 47% 8%)',
    'popover-foreground': 'hsl(210 40% 98%)',
    accent: 'hsl(217 33% 14%)',
    'accent-foreground': 'hsl(210 40% 98%)',
    secondary: 'hsl(217 33% 14%)',
    'secondary-foreground': 'hsl(210 40% 98%)',
  },
  semantic: {
    success: 'hsl(142 71% 45%)',
    'success-foreground': 'hsl(0 0% 100%)',
    warning: 'hsl(38 92% 50%)',
    'warning-foreground': 'hsl(0 0% 100%)',
    destructive: 'hsl(0 63% 31%)',
    'destructive-foreground': 'hsl(210 40% 98%)',
    info: 'hsl(199 89% 48%)',
    'info-foreground': 'hsl(0 0% 100%)',
  },
} as const;

export type ColorTokenGroup = keyof typeof ColorTokens;
export type ColorTokenPath<C extends ColorTokenGroup = ColorTokenGroup> = {
  [G in C]: keyof (typeof ColorTokens)[G];
}[C];

export function readCssVar(name: string, fallback = ''): string {
  if (typeof window === 'undefined') return fallback;
  const v = getComputedStyle(document.documentElement).getPropertyValue(name).trim();
  return v || fallback;
}
