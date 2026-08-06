export * from './colors';
export * from './spacing';
export * from './typography';
export * from './radius';
export * from './shadows';
export * from './animations';
export * from './breakpoints';
export * from './zIndex';
export * from './transitions';
export * from './icons';

import { ColorTokens, DarkColorTokens } from './colors';
import { Spacing } from './spacing';
import { FontFamily, FontSize, FontWeight, LetterSpacing, TypeRole } from './typography';
import { Radius } from './radius';
import { Shadow, SemanticShadow } from './shadows';
import { Easing, Duration, Keyframes, Transition, Animation } from './animations';
import { Breakpoints, MediaQuery } from './breakpoints';
import { ZIndex } from './zIndex';
import { Transitions, ReducedMotion } from './transitions';

export const tokens = {
  colors: { light: ColorTokens, dark: DarkColorTokens },
  spacing: Spacing,
  font: { family: FontFamily, size: FontSize, weight: FontWeight, tracking: LetterSpacing, role: TypeRole },
  radius: Radius,
  shadow: { ...Shadow, semantic: SemanticShadow },
  animation: { easing: Easing, duration: Duration, keyframes: Keyframes, transition: Transition, definition: Animation },
  breakpoint: { value: Breakpoints, query: MediaQuery },
  zIndex: ZIndex,
  transitions: Transitions,
  reducedMotion: ReducedMotion,
} as const;

export type Tokens = typeof tokens;
