export const Breakpoints = {
  sm: 640,
  md: 768,
  lg: 1024,
  xl: 1280,
  '2xl': 1536,
} as const;

export type BreakpointKey = keyof typeof Breakpoints;

export const MediaQuery = {
  sm: `(min-width: ${Breakpoints.sm}px)`,
  md: `(min-width: ${Breakpoints.md}px)`,
  lg: `(min-width: ${Breakpoints.lg}px)`,
  xl: `(min-width: ${Breakpoints.xl}px)`,
  '2xl': `(min-width: ${Breakpoints['2xl']}px)`,
  mobile: `(max-width: ${Breakpoints.md - 1}px)`,
  tablet: `(min-width: ${Breakpoints.md}px) and (max-width: ${Breakpoints.lg - 1}px)`,
  desktop: `(min-width: ${Breakpoints.lg}px)`,
  reducedMotion: '(prefers-reduced-motion: reduce)',
  dark: '(prefers-color-scheme: dark)',
  light: '(prefers-color-scheme: light)',
} as const;
