export const FontFamily = {
  sans: [
    'Inter',
    '-apple-system',
    'BlinkMacSystemFont',
    'Segoe UI',
    'Roboto',
    'sans-serif',
  ],
  mono: [
    'JetBrains Mono',
    'Menlo',
    'Monaco',
    'Consolas',
    'monospace',
  ],
} as const;

export const FontSize = {
  xs: { size: '0.75rem', lineHeight: '1rem' },
  sm: { size: '0.8125rem', lineHeight: '1.25rem' },
  base: { size: '0.875rem', lineHeight: '1.25rem' },
  lg: { size: '1.125rem', lineHeight: '1.75rem' },
  xl: { size: '1.25rem', lineHeight: '1.75rem' },
  '2xl': { size: '1.5rem', lineHeight: '2rem' },
  '3xl': { size: '1.875rem', lineHeight: '2.25rem' },
  '4xl': { size: '2.25rem', lineHeight: '2.5rem' },
} as const;

export const FontWeight = {
  regular: 400,
  medium: 500,
  semibold: 600,
  bold: 700,
} as const;

export const LetterSpacing = {
  tight: '-0.015em',
  tighter: '-0.025em',
  wide: '0.025em',
  wider: '0.05em',
} as const;

export const TypeRole = {
  display: { ...FontSize['3xl'], weight: FontWeight.semibold, tracking: LetterSpacing.tight },
  pageTitle: { ...FontSize['2xl'], weight: FontWeight.semibold, tracking: LetterSpacing.tight },
  sectionTitle: { ...FontSize.xl, weight: FontWeight.semibold, tracking: LetterSpacing.tight },
  cardTitle: { ...FontSize.lg, weight: FontWeight.semibold, tracking: LetterSpacing.tight },
  body: { ...FontSize.base, weight: FontWeight.regular },
  table: { ...FontSize.sm, weight: FontWeight.regular },
  label: { ...FontSize.sm, weight: FontWeight.medium },
  caption: { ...FontSize.xs, weight: FontWeight.medium, tracking: LetterSpacing.wide },
  numeric: { ...FontSize['3xl'], weight: FontWeight.semibold, tabular: true },
} as const;
