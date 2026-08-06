export const Easing = {
  linear: 'linear',
  in: 'cubic-bezier(0.4, 0, 1, 1)',
  out: 'cubic-bezier(0, 0, 0.2, 1)',
  inOut: 'cubic-bezier(0.4, 0, 0.2, 1)',
  bounce: 'cubic-bezier(0.68, -0.55, 0.265, 1.55)',
} as const;

export const Duration = {
  instant: '0ms',
  fast: '100ms',
  base: '150ms',
  medium: '200ms',
  slow: '300ms',
  slower: '500ms',
} as const;

export const Transition = {
  hover: { duration: Duration.fast, easing: Easing.out },
  button: { duration: Duration.base, easing: Easing.out },
  dialog: { duration: Duration.medium, easing: Easing.inOut },
  page: { duration: Duration.slow, easing: Easing.inOut },
  toast: { duration: Duration.medium, easing: Easing.out },
} as const;

export const Keyframes = {
  'accordion-down': {
    from: { height: '0' },
    to: { height: 'var(--radix-accordion-content-height)' },
  },
  'accordion-up': {
    from: { height: 'var(--radix-accordion-content-height)' },
    to: { height: '0' },
  },
  'fade-in': { from: { opacity: '0' }, to: { opacity: '1' } },
  'fade-out': { from: { opacity: '1' }, to: { opacity: '0' } },
  'slide-in-from-top': { from: { transform: 'translateY(-8px)' }, to: { transform: 'translateY(0)' } },
  'slide-in-from-bottom': { from: { transform: 'translateY(8px)' }, to: { transform: 'translateY(0)' } },
  'slide-in-from-left': { from: { transform: 'translateX(-8px)' }, to: { transform: 'translateX(0)' } },
  'slide-in-from-right': { from: { transform: 'translateX(100%)' }, to: { transform: 'translateX(0)' } },
  'slide-out-to-right': { from: { transform: 'translateX(0)' }, to: { transform: 'translateX(100%)' } },
  'zoom-in-95': { from: { transform: 'scale(0.95)' }, to: { transform: 'scale(1)' } },
  'zoom-out-95': { from: { transform: 'scale(1)' }, to: { transform: 'scale(0.95)' } },
  spin: { to: { transform: 'rotate(360deg)' } },
  pulse: {
    '0%, 100%': { opacity: '1' },
    '50%': { opacity: '0.5' },
  },
} as const;

export const Animation = {
  fadeIn: { name: 'fade-in', duration: Duration.base, easing: Easing.out },
  slideInFromRight: { name: 'slide-in-from-right', duration: Duration.medium, easing: Easing.out },
  slideOutToRight: { name: 'slide-out-to-right', duration: Duration.medium, easing: Easing.in },
  dialogIn: { name: 'fade-in', duration: Duration.medium, easing: Easing.inOut },
  dialogOut: { name: 'fade-out', duration: Duration.base, easing: Easing.in },
  spin: { name: 'spin', duration: '1s', easing: Easing.linear, infinite: true },
  pulse: { name: 'pulse', duration: '2s', easing: Easing.inOut, infinite: true },
} as const;
