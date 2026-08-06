export const Radius = {
  none: '0',
  sm: 'calc(var(--radius) - 4px)',
  md: 'calc(var(--radius) - 2px)',
  lg: 'var(--radius)',
  xl: 'calc(var(--radius) + 4px)',
  '2xl': 'calc(var(--radius) + 8px)',
  full: '9999px',
} as const;

export const CSSVarRadius = {
  base: '0.5rem',
} as const;
