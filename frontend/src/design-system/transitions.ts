import { Duration, Easing } from './animations';

export const Transitions = {
  colors: { property: 'color, background-color, border-color, fill, stroke', duration: Duration.base, easing: Easing.out },
  opacity: { property: 'opacity', duration: Duration.base, easing: Easing.out },
  transform: { property: 'transform', duration: Duration.medium, easing: Easing.inOut },
  shadow: { property: 'box-shadow', duration: Duration.base, easing: Easing.out },
  all: { property: 'all', duration: Duration.base, easing: Easing.out },
  none: { property: 'none', duration: Duration.instant, easing: Easing.linear },
} as const;

export const ReducedMotion = {
  duration: Duration.instant,
} as const;
